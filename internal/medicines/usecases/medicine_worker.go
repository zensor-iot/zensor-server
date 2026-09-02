package usecases

import (
	"context"
	"log/slog"
	"sync"
	"time"
	"zensor-server/internal/infra/async"

	medicinesDomain "zensor-server/internal/medicines/domain"
	shareddomain "zensor-server/internal/shared_kernel/domain"
	sharedUsecases "zensor-server/internal/shared_kernel/usecases"
)

const (
	_dosesTopic   = "medicine_doses"
	_doseDueEvent = "dose_due"

	_defaultTimezone = "UTC"

	// How far ahead doses are materialised. A horizon rather than a fixed
	// count, so an hourly treatment is covered as well as a daily one and the
	// UI can show tomorrow's plan.
	_materialiseHorizon = 48 * time.Hour
	// How far back materialisation reaches. After an outage the doses that fell
	// due while the server was down still belong in the record, as overdue rows
	// the user can mark skipped, rather than as silent holes.
	_materialiseBackfill = 24 * time.Hour

	// _reminderLead must stay strictly greater than the worker's ticker
	// interval. Every dose has to fall inside at least one tick's window, or
	// its reminder is silently dropped; the overlap between consecutive windows
	// is deduplicated by reminder_sent_at.
	_reminderLead = 3 * time.Minute
	// Past this much delay a dose no longer generates a push. It is already
	// visibly overdue in the UI, and this stops a burst of stale notifications
	// when the server comes back from an outage.
	_reminderGrace      = 30 * time.Minute
	_reminderBatchLimit = 200
)

func NewMedicineWorker(
	ticker *time.Ticker,
	treatmentRepository TreatmentRepository,
	doseRepository DoseRepository,
	tenantService sharedUsecases.TenantService,
	tenantConfigurationService sharedUsecases.TenantConfigurationService,
	broker async.InternalBroker,
) *MedicineWorker {
	return &MedicineWorker{
		ticker:                     ticker,
		treatmentRepository:        treatmentRepository,
		doseRepository:             doseRepository,
		tenantService:              tenantService,
		tenantConfigurationService: tenantConfigurationService,
		broker:                     broker,
	}
}

var _ async.Worker = &MedicineWorker{}

// MedicineWorker materialises the doses of the active treatments and emits a
// reminder when one falls due.
type MedicineWorker struct {
	ticker                     *time.Ticker
	treatmentRepository        TreatmentRepository
	doseRepository             DoseRepository
	tenantService              sharedUsecases.TenantService
	tenantConfigurationService sharedUsecases.TenantConfigurationService
	broker                     async.InternalBroker
}

func (w *MedicineWorker) Run(ctx context.Context, done func()) {
	slog.Info("medicine worker started")
	defer done()
	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			slog.Info("medicine worker cancelled")
			wg.Wait()
			return
		case <-w.ticker.C:
			tickCtx := context.Background()
			w.MaterialiseDoses(tickCtx)
			w.EmitReminders(tickCtx)
		}
	}
}

func (w *MedicineWorker) Shutdown() {
	slog.Warn("medicine worker shutdown")
}

// MaterialiseDoses creates the doses of every active treatment that fall inside
// the materialisation window and do not exist yet.
//
// It is idempotent: sequence numbers are derived arithmetically from the
// schedule, so two ticks compute identical indices for the overlapping part of
// the window and the existing rows filter the new ones out.
func (w *MedicineWorker) MaterialiseDoses(ctx context.Context) {
	treatments, err := w.treatmentRepository.FindAllActive(ctx)
	if err != nil {
		slog.Error("finding active treatments", slog.Any("error", err))
		return
	}

	now := time.Now()
	from := now.Add(-_materialiseBackfill)
	to := now.Add(_materialiseHorizon)

	for _, treatment := range treatments {
		w.materialiseTreatment(ctx, treatment, from, to)
	}
}

func (w *MedicineWorker) materialiseTreatment(ctx context.Context, treatment medicinesDomain.Treatment, from, to time.Time) {
	occurrences, err := treatment.DueOccurrences(from, to)
	if err != nil {
		slog.Error("resolving due doses",
			slog.String("treatment_id", treatment.ID.String()),
			slog.Any("error", err))
		return
	}
	if len(occurrences) == 0 {
		return
	}

	existing, err := w.doseRepository.FindByTreatmentInWindow(ctx, treatment.ID, from, to)
	if err != nil {
		slog.Error("finding existing doses",
			slog.String("treatment_id", treatment.ID.String()),
			slog.Any("error", err))
		return
	}

	materialised := make(map[int]struct{}, len(existing))
	for _, dose := range existing {
		materialised[dose.SequenceNumber] = struct{}{}
	}

	missing := make([]medicinesDomain.Dose, 0, len(occurrences))
	for _, occurrence := range occurrences {
		if _, ok := materialised[occurrence.SequenceNumber]; ok {
			continue
		}

		dose, buildErr := medicinesDomain.NewDoseBuilder().
			WithTreatmentID(treatment.ID.String()).
			WithSequenceNumber(occurrence.SequenceNumber).
			WithScheduledAt(occurrence.ScheduledAt).
			WithDose(treatment.Quantity, treatment.Unit).
			Build()
		if buildErr != nil {
			slog.Error("building dose",
				slog.String("treatment_id", treatment.ID.String()),
				slog.Int("sequence_number", occurrence.SequenceNumber),
				slog.Any("error", buildErr))
			continue
		}

		missing = append(missing, dose)
	}

	if len(missing) == 0 {
		return
	}

	if err := w.doseRepository.CreateBatch(ctx, missing); err != nil {
		slog.Error("creating doses",
			slog.String("treatment_id", treatment.ID.String()),
			slog.Any("error", err))
		return
	}

	slog.Info("doses materialised",
		slog.String("treatment_id", treatment.ID.String()),
		slog.Int("count", len(missing)))
}

// EmitReminders publishes one dose_due event per dose falling due, marking each
// one before publishing so a dose is reminded about at most once.
func (w *MedicineWorker) EmitReminders(ctx context.Context) {
	now := time.Now()
	entries, err := w.doseRepository.FindDueForReminder(
		ctx,
		now.Add(-_reminderGrace),
		now.Add(_reminderLead),
		_reminderBatchLimit,
	)
	if err != nil {
		slog.Error("finding doses due for reminder", slog.Any("error", err))
		return
	}

	// One timezone lookup per tenant rather than one per dose.
	locations := make(map[shareddomain.ID]*time.Location)

	for _, entry := range entries {
		entry.Dose.MarkReminderSent()
		if err := w.doseRepository.Update(ctx, entry.Dose); err != nil {
			slog.Error("marking reminder as sent",
				slog.String("dose_id", entry.Dose.ID.String()),
				slog.Any("error", err))
			continue
		}

		w.publishDoseDue(ctx, entry, w.tenantLocation(ctx, entry.Treatment.TenantID, locations))
	}
}

func (w *MedicineWorker) publishDoseDue(ctx context.Context, entry DoseWithContext, location *time.Location) {
	// Every value is a flat string the push templating can interpolate with
	// {{key}}. Anything a person reads is rendered here, because that engine is
	// a plain fmt.Sprintf("%v") and would print a raw timestamp verbatim.
	message := async.BrokerMessage{
		Event: _doseDueEvent,
		Value: map[string]any{
			"dose_id":        entry.Dose.ID.String(),
			"treatment_id":   entry.Treatment.ID.String(),
			"patient_id":     entry.Patient.ID.String(),
			"tenant_id":      entry.Treatment.TenantID.String(),
			"patient_name":   string(entry.Patient.Name),
			"medicine_name":  string(entry.Treatment.MedicineName),
			"dose_text":      entry.Treatment.DoseText(),
			"scheduled_time": entry.Dose.ScheduledAt.In(location).Format("15:04"),
		},
	}

	if err := w.broker.Publish(ctx, async.BrokerTopicName(_dosesTopic), message); err != nil {
		slog.Error("publishing dose due event",
			slog.String("dose_id", entry.Dose.ID.String()),
			slog.Any("error", err))
		return
	}

	slog.Info("dose due event published",
		slog.String("dose_id", entry.Dose.ID.String()),
		slog.String("patient", string(entry.Patient.Name)))
}

func (w *MedicineWorker) tenantLocation(
	ctx context.Context,
	tenantID shareddomain.ID,
	cache map[shareddomain.ID]*time.Location,
) *time.Location {
	if location, ok := cache[tenantID]; ok {
		return location
	}

	location := time.UTC
	defer func() { cache[tenantID] = location }()

	tenant, err := w.tenantService.GetTenant(ctx, tenantID)
	if err != nil {
		slog.Warn("getting tenant for dose reminder, falling back to UTC",
			slog.String("tenant_id", tenantID.String()),
			slog.Any("error", err))
		return location
	}

	configuration, err := w.tenantConfigurationService.GetOrCreateTenantConfiguration(ctx, tenant, _defaultTimezone)
	if err != nil {
		slog.Warn("getting tenant configuration, falling back to UTC",
			slog.String("tenant_id", tenantID.String()),
			slog.Any("error", err))
		return location
	}

	loaded, err := time.LoadLocation(configuration.Timezone)
	if err != nil {
		slog.Warn("loading tenant timezone, falling back to UTC",
			slog.String("timezone", configuration.Timezone),
			slog.Any("error", err))
		return location
	}

	location = loaded

	return location
}

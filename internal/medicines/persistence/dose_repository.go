package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/medicines/persistence/internal"
	"zensor-server/internal/medicines/usecases"

	medicinesDomain "zensor-server/internal/medicines/domain"

	shareddomain "zensor-server/internal/shared_kernel/domain"
)

// Every time bound is normalised to UTC before it reaches the driver: doses are
// stored as UTC instants, and SQLite compares timestamps as text, so passing a
// zoned local time would silently shift the window.
const (
	_treatmentJoin = "JOIN medicine_treatments ON medicine_doses.treatment_id = medicine_treatments.id"
	_patientJoin   = "JOIN medicine_patients ON medicine_treatments.patient_id = medicine_patients.id"
)

func NewDoseRepository(orm sql.ORM) (*SimpleDoseRepository, error) {
	if err := orm.AutoMigrate(&internal.Dose{}); err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
	}

	return &SimpleDoseRepository{orm: orm}, nil
}

var _ usecases.DoseRepository = (*SimpleDoseRepository)(nil)

type SimpleDoseRepository struct {
	orm sql.ORM
}

func (r *SimpleDoseRepository) CreateBatch(ctx context.Context, doses []medicinesDomain.Dose) error {
	if len(doses) == 0 {
		return nil
	}

	entities := make([]internal.Dose, len(doses))
	for i, dose := range doses {
		entities[i] = internal.FromDose(dose)
	}

	if err := r.orm.WithContext(ctx).Create(&entities).Error(); err != nil {
		return fmt.Errorf("creating doses in database: %w", err)
	}

	return nil
}

func (r *SimpleDoseRepository) GetByID(ctx context.Context, id shareddomain.ID) (medicinesDomain.Dose, error) {
	var entity internal.Dose
	err := r.orm.WithContext(ctx).First(&entity, "id = ?", id.String()).Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return medicinesDomain.Dose{}, usecases.ErrDoseNotFound
	}
	if err != nil {
		return medicinesDomain.Dose{}, fmt.Errorf("database query: %w", err)
	}

	return entity.ToDomain(), nil
}

func (r *SimpleDoseRepository) FindAllByTreatment(
	ctx context.Context,
	treatmentID shareddomain.ID,
	pagination usecases.Pagination,
) ([]medicinesDomain.Dose, int, error) {
	entities, total, err := paginateByFilter(ctx, r.orm, internal.Dose{},
		"treatment_id = ? AND deleted_at IS NULL", pagination, treatmentID.String())
	if err != nil {
		return nil, 0, err
	}

	return toDomainDoses(entities), total, nil
}

func (r *SimpleDoseRepository) FindByTreatmentInWindow(
	ctx context.Context,
	treatmentID shareddomain.ID,
	from, to time.Time,
) ([]medicinesDomain.Dose, error) {
	var entities []internal.Dose
	err := r.orm.
		WithContext(ctx).
		Where("treatment_id = ? AND deleted_at IS NULL AND scheduled_at BETWEEN ? AND ?",
			treatmentID.String(), from.UTC(), to.UTC()).
		Find(&entities).
		Error()
	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	return toDomainDoses(entities), nil
}

// FindDueForReminder returns the pending doses falling in the window that have
// not been reminded about yet, together with the treatment and patient they
// belong to. The window is applied in SQL rather than in Go, so a tick reads
// only the rows it is about to act on.
func (r *SimpleDoseRepository) FindDueForReminder(
	ctx context.Context,
	from, to time.Time,
	limit int,
) ([]usecases.DoseWithContext, error) {
	var entities []internal.Dose
	err := r.orm.
		WithContext(ctx).
		Model(&internal.Dose{}).
		Joins(_treatmentJoin).
		Joins(_patientJoin).
		Where(`medicine_doses.deleted_at IS NULL
			AND medicine_doses.status = ?
			AND medicine_doses.reminder_sent_at IS NULL
			AND medicine_doses.scheduled_at BETWEEN ? AND ?
			AND medicine_treatments.deleted_at IS NULL
			AND medicine_treatments.is_active = ?
			AND medicine_patients.deleted_at IS NULL`,
			string(medicinesDomain.DoseStatusPending), from.UTC(), to.UTC(), true).
		Order("medicine_doses.scheduled_at").
		Limit(limit).
		Find(&entities).
		Error()
	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	return r.withContext(ctx, entities)
}

// FindAgenda returns the doses of a tenant in the window, whatever their
// status, so the UI can show what has been given and what is still due.
func (r *SimpleDoseRepository) FindAgenda(
	ctx context.Context,
	tenantID shareddomain.ID,
	from, to time.Time,
	limit int,
) ([]usecases.DoseWithContext, error) {
	var entities []internal.Dose
	err := r.orm.
		WithContext(ctx).
		Model(&internal.Dose{}).
		Joins(_treatmentJoin).
		Joins(_patientJoin).
		Where(`medicine_doses.deleted_at IS NULL
			AND medicine_doses.scheduled_at BETWEEN ? AND ?
			AND medicine_treatments.tenant_id = ?
			AND medicine_treatments.deleted_at IS NULL
			AND medicine_patients.deleted_at IS NULL`,
			from.UTC(), to.UTC(), tenantID.String()).
		Order("medicine_doses.scheduled_at").
		Limit(limit).
		Find(&entities).
		Error()
	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	return r.withContext(ctx, entities)
}

func (r *SimpleDoseRepository) Update(ctx context.Context, dose medicinesDomain.Dose) error {
	entity := internal.FromDose(dose)

	if err := r.orm.WithContext(ctx).Save(&entity).Error(); err != nil {
		return fmt.Errorf("updating dose in database: %w", err)
	}

	return nil
}

// DeletePendingFrom removes the not yet resolved doses scheduled after the given
// instant. Resolved doses are the record of what actually happened and are never
// touched.
//
// The removal is physical rather than a soft delete on purpose. A pending dose
// nobody acted on carries no history, and the unique index on (treatment_id,
// sequence_number) does not exclude soft deleted rows: leaving tombstones behind
// would permanently occupy those sequence numbers, so after a schedule change
// the worker could never rebuild the doses and would fail on every tick.
func (r *SimpleDoseRepository) DeletePendingFrom(ctx context.Context, treatmentID shareddomain.ID, from time.Time) error {
	err := r.orm.
		WithContext(ctx).
		Unscoped().
		Where("treatment_id = ? AND status = ? AND scheduled_at > ?",
			treatmentID.String(), string(medicinesDomain.DoseStatusPending), from.UTC()).
		Delete(&internal.Dose{}).
		Error()
	if err != nil {
		return fmt.Errorf("deleting pending doses: %w", err)
	}

	return nil
}

// withContext hydrates the treatment and patient of each dose with two batched
// queries rather than one per row.
func (r *SimpleDoseRepository) withContext(ctx context.Context, entities []internal.Dose) ([]usecases.DoseWithContext, error) {
	if len(entities) == 0 {
		return nil, nil
	}

	treatmentIDs := make([]string, 0, len(entities))
	seenTreatments := make(map[string]struct{}, len(entities))
	for _, entity := range entities {
		if _, ok := seenTreatments[entity.TreatmentID]; ok {
			continue
		}
		seenTreatments[entity.TreatmentID] = struct{}{}
		treatmentIDs = append(treatmentIDs, entity.TreatmentID)
	}

	var treatmentEntities []internal.Treatment
	if err := r.orm.WithContext(ctx).Where("id IN ?", treatmentIDs).Find(&treatmentEntities).Error(); err != nil {
		return nil, fmt.Errorf("fetching treatments: %w", err)
	}

	treatments := make(map[string]medicinesDomain.Treatment, len(treatmentEntities))
	patientIDs := make([]string, 0, len(treatmentEntities))
	seenPatients := make(map[string]struct{}, len(treatmentEntities))
	for _, entity := range treatmentEntities {
		treatments[entity.ID] = entity.ToDomain()
		if _, ok := seenPatients[entity.PatientID]; ok {
			continue
		}
		seenPatients[entity.PatientID] = struct{}{}
		patientIDs = append(patientIDs, entity.PatientID)
	}

	var patientEntities []internal.Patient
	if err := r.orm.WithContext(ctx).Where("id IN ?", patientIDs).Find(&patientEntities).Error(); err != nil {
		return nil, fmt.Errorf("fetching patients: %w", err)
	}

	patients := make(map[string]medicinesDomain.Patient, len(patientEntities))
	for _, entity := range patientEntities {
		patients[entity.ID] = entity.ToDomain()
	}

	result := make([]usecases.DoseWithContext, 0, len(entities))
	for _, entity := range entities {
		treatment, ok := treatments[entity.TreatmentID]
		if !ok {
			continue
		}
		result = append(result, usecases.DoseWithContext{
			Dose:      entity.ToDomain(),
			Treatment: treatment,
			Patient:   patients[treatment.PatientID.String()],
		})
	}

	return result, nil
}

func toDomainDoses(entities []internal.Dose) []medicinesDomain.Dose {
	result := make([]medicinesDomain.Dose, len(entities))
	for i, entity := range entities {
		result[i] = entity.ToDomain()
	}

	return result
}

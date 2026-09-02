package domain

import (
	"fmt"
	"time"
	"zensor-server/internal/infra/utils"

	shareddomain "zensor-server/internal/shared_kernel/domain"
)

// DoseOccurrence is a dose the schedule says is due, before it is persisted.
type DoseOccurrence struct {
	SequenceNumber int
	ScheduledAt    time.Time
}

// Treatment is the prescription for a patient: which medicine, how much of it
// per dose, and how often.
//
// A treatment ends at whichever of its two optional limits comes first: after
// TotalDoses doses, or after EndAt. With neither set it runs indefinitely,
// which is the normal case for chronic medication.
type Treatment struct {
	ID           shareddomain.ID          `json:"id"`
	Version      shareddomain.Version     `json:"version"`
	TenantID     shareddomain.ID          `json:"tenant_id"`
	PatientID    shareddomain.ID          `json:"patient_id"`
	MedicineName MedicineName             `json:"medicine_name"`
	Quantity     DoseQuantity             `json:"quantity"`
	Unit         DoseUnit                 `json:"unit"`
	Schedule     MedicineSchedule         `json:"schedule"`
	EndAt        *time.Time               `json:"end_at,omitempty"`
	TotalDoses   *int                     `json:"total_doses,omitempty"`
	Notes        shareddomain.Description `json:"notes"`
	IsActive     bool                     `json:"is_active"`
	CreatedAt    utils.Time               `json:"created_at"`
	UpdatedAt    utils.Time               `json:"updated_at"`
	DeletedAt    *utils.Time              `json:"deleted_at,omitempty"`
}

func (t *Treatment) IsDeleted() bool {
	return t.DeletedAt != nil
}

func (t *Treatment) SoftDelete() {
	now := utils.Time{Time: time.Now()}
	t.DeletedAt = &now
	t.UpdatedAt = now
}

func (t *Treatment) Activate() {
	t.IsActive = true
	t.UpdatedAt = utils.Time{Time: time.Now()}
}

func (t *Treatment) Deactivate() {
	t.IsActive = false
	t.UpdatedAt = utils.Time{Time: time.Now()}
}

// DoseText renders the dose the way a person reads it, such as "15 gotas".
//
// It lives in the domain because the push notification templating engine only
// interpolates strings, so anything human readable has to be rendered before
// it reaches the payload.
func (t *Treatment) DoseText() string {
	return fmt.Sprintf("%s %s", t.Quantity.String(), t.Unit.Label(t.Quantity))
}

// EffectiveEndAt returns the instant after which no dose is due, whichever of
// the two limits comes first, or nil when the treatment is indefinite.
func (t *Treatment) EffectiveEndAt() *time.Time {
	var byCount *time.Time
	if t.TotalDoses != nil && *t.TotalDoses > 0 {
		last := t.Schedule.OccurrenceAt(*t.TotalDoses - 1)
		byCount = &last
	}

	switch {
	case byCount == nil:
		return t.EndAt
	case t.EndAt == nil:
		return byCount
	case t.EndAt.Before(*byCount):
		return t.EndAt
	default:
		return byCount
	}
}

// DueOccurrences returns the doses falling in the closed interval [from, to],
// bounded by both end conditions. An inactive or deleted treatment produces
// none.
func (t *Treatment) DueOccurrences(from, to time.Time) ([]DoseOccurrence, error) {
	if !t.IsActive || t.IsDeleted() {
		return nil, nil
	}

	if t.EndAt != nil && t.EndAt.Before(to) {
		to = *t.EndAt
	}

	indices, err := t.Schedule.OccurrenceIndicesBetween(from, to)
	if err != nil {
		return nil, fmt.Errorf("resolving occurrences: %w", err)
	}

	occurrences := make([]DoseOccurrence, 0, len(indices))
	for _, n := range indices {
		if t.TotalDoses != nil && n >= *t.TotalDoses {
			break
		}
		occurrences = append(occurrences, DoseOccurrence{
			SequenceNumber: n,
			ScheduledAt:    t.Schedule.OccurrenceAt(n),
		})
	}

	return occurrences, nil
}

type treatmentBuilder struct {
	actions []func(*Treatment) error
}

func NewTreatmentBuilder() *treatmentBuilder {
	return &treatmentBuilder{}
}

func (b *treatmentBuilder) WithTenantID(value string) *treatmentBuilder {
	b.actions = append(b.actions, func(t *Treatment) error {
		t.TenantID = shareddomain.ID(value)
		return nil
	})
	return b
}

func (b *treatmentBuilder) WithPatientID(value string) *treatmentBuilder {
	b.actions = append(b.actions, func(t *Treatment) error {
		t.PatientID = shareddomain.ID(value)
		return nil
	})
	return b
}

func (b *treatmentBuilder) WithMedicineName(value string) *treatmentBuilder {
	b.actions = append(b.actions, func(t *Treatment) error {
		t.MedicineName = MedicineName(value)
		return nil
	})
	return b
}

func (b *treatmentBuilder) WithDose(quantity float64, unit DoseUnit) *treatmentBuilder {
	b.actions = append(b.actions, func(t *Treatment) error {
		t.Quantity = DoseQuantity(quantity)
		t.Unit = unit
		return nil
	})
	return b
}

func (b *treatmentBuilder) WithSchedule(value MedicineSchedule) *treatmentBuilder {
	b.actions = append(b.actions, func(t *Treatment) error {
		t.Schedule = value
		return nil
	})
	return b
}

func (b *treatmentBuilder) WithEndAt(value *time.Time) *treatmentBuilder {
	b.actions = append(b.actions, func(t *Treatment) error {
		t.EndAt = value
		return nil
	})
	return b
}

func (b *treatmentBuilder) WithTotalDoses(value *int) *treatmentBuilder {
	b.actions = append(b.actions, func(t *Treatment) error {
		t.TotalDoses = value
		return nil
	})
	return b
}

func (b *treatmentBuilder) WithNotes(value string) *treatmentBuilder {
	b.actions = append(b.actions, func(t *Treatment) error {
		t.Notes = shareddomain.Description(value)
		return nil
	})
	return b
}

func (b *treatmentBuilder) Build() (Treatment, error) {
	now := utils.Time{Time: time.Now()}
	result := Treatment{
		ID:        shareddomain.ID(utils.GenerateUUID()),
		Version:   1,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	for _, action := range b.actions {
		if err := action(&result); err != nil {
			return Treatment{}, err
		}
	}

	if err := result.validate(); err != nil {
		return Treatment{}, err
	}

	return result, nil
}

func (t *Treatment) validate() error {
	if t.TenantID == "" {
		return ErrTenantIDRequired
	}
	if t.PatientID == "" {
		return ErrPatientIDRequired
	}
	if t.MedicineName == "" {
		return ErrMedicineNameRequired
	}
	if t.Quantity <= 0 {
		return ErrDoseQuantityRequired
	}
	if err := t.Unit.Validate(); err != nil {
		return err
	}
	if err := t.Schedule.Validate(); err != nil {
		return err
	}
	if t.TotalDoses != nil && *t.TotalDoses <= 0 {
		return ErrTotalDosesInvalid
	}
	if t.EndAt != nil && !t.EndAt.After(t.Schedule.StartAt) {
		return ErrEndAtBeforeStartAt
	}

	return nil
}

// Package domain holds the medicines module aggregates: the family members
// that receive medication, their treatments, and the individual doses.
package domain

import (
	"time"
	"zensor-server/internal/infra/utils"

	shareddomain "zensor-server/internal/shared_kernel/domain"
)

// Dose is one scheduled administration of a treatment, materialised ahead of
// time so it can be reminded about and then confirmed.
//
// SequenceNumber is the treatment relative index of the dose, counting from
// zero, and is what makes materialisation idempotent. It is derived
// arithmetically from the schedule, so re-running the materialiser produces
// the same indices and the existing rows filter the new ones out. Deduplicating
// on the scheduled instant instead would be unreliable: Go keeps nanoseconds
// and Postgres timestamps keep microseconds, so a value read back from the
// database no longer equals the one that was written.
type Dose struct {
	ID             shareddomain.ID      `json:"id"`
	Version        shareddomain.Version `json:"version"`
	TreatmentID    shareddomain.ID      `json:"treatment_id"`
	SequenceNumber int                  `json:"sequence_number"`
	ScheduledAt    utils.Time           `json:"scheduled_at"`
	Status         DoseStatus           `json:"status"`
	Quantity       DoseQuantity         `json:"quantity"`
	Unit           DoseUnit             `json:"unit"`
	ResolvedAt     *utils.Time          `json:"resolved_at,omitempty"`
	ResolvedBy     *ResolvedBy          `json:"resolved_by,omitempty"`
	Notes          *DoseNotes           `json:"notes,omitempty"`
	ReminderSentAt *utils.Time          `json:"reminder_sent_at,omitempty"`
	CreatedAt      utils.Time           `json:"created_at"`
	UpdatedAt      utils.Time           `json:"updated_at"`
	DeletedAt      *utils.Time          `json:"deleted_at,omitempty"`
}

func (d *Dose) IsDeleted() bool {
	return d.DeletedAt != nil
}

func (d *Dose) SoftDelete() {
	now := utils.Time{Time: time.Now()}
	d.DeletedAt = &now
	d.UpdatedAt = now
}

// IsResolved reports whether the dose has already been administered or skipped.
func (d *Dose) IsResolved() bool {
	return d.Status != DoseStatusPending
}

// IsOverdue reports whether the dose is still pending past its scheduled time.
func (d *Dose) IsOverdue(now time.Time) bool {
	return !d.IsResolved() && now.After(d.ScheduledAt.Time)
}

func (d *Dose) MarkAdministered(by ResolvedBy, notes *DoseNotes) error {
	return d.resolve(DoseStatusAdministered, by, notes)
}

func (d *Dose) MarkSkipped(by ResolvedBy, notes *DoseNotes) error {
	return d.resolve(DoseStatusSkipped, by, notes)
}

func (d *Dose) resolve(status DoseStatus, by ResolvedBy, notes *DoseNotes) error {
	if d.IsResolved() {
		return ErrDoseAlreadyResolved
	}

	now := utils.Time{Time: time.Now()}
	d.Status = status
	d.ResolvedAt = &now
	d.ResolvedBy = &by
	d.Notes = notes
	d.UpdatedAt = now
	d.Version++

	return nil
}

func (d *Dose) HasReminderBeenSent() bool {
	return d.ReminderSentAt != nil
}

func (d *Dose) MarkReminderSent() {
	now := utils.Time{Time: time.Now()}
	d.ReminderSentAt = &now
	d.UpdatedAt = now
}

type doseBuilder struct {
	actions []func(*Dose) error
}

func NewDoseBuilder() *doseBuilder {
	return &doseBuilder{}
}

func (b *doseBuilder) WithTreatmentID(value string) *doseBuilder {
	b.actions = append(b.actions, func(d *Dose) error {
		d.TreatmentID = shareddomain.ID(value)
		return nil
	})
	return b
}

func (b *doseBuilder) WithSequenceNumber(value int) *doseBuilder {
	b.actions = append(b.actions, func(d *Dose) error {
		d.SequenceNumber = value
		return nil
	})
	return b
}

func (b *doseBuilder) WithScheduledAt(value time.Time) *doseBuilder {
	b.actions = append(b.actions, func(d *Dose) error {
		d.ScheduledAt = utils.Time{Time: value}
		return nil
	})
	return b
}

func (b *doseBuilder) WithDose(quantity DoseQuantity, unit DoseUnit) *doseBuilder {
	b.actions = append(b.actions, func(d *Dose) error {
		d.Quantity = quantity
		d.Unit = unit
		return nil
	})
	return b
}

func (b *doseBuilder) Build() (Dose, error) {
	now := utils.Time{Time: time.Now()}
	result := Dose{
		ID:        shareddomain.ID(utils.GenerateUUID()),
		Version:   1,
		Status:    DoseStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	for _, action := range b.actions {
		if err := action(&result); err != nil {
			return Dose{}, err
		}
	}

	if result.TreatmentID == "" {
		return Dose{}, ErrTreatmentIDRequired
	}
	if result.ScheduledAt.IsZero() {
		return Dose{}, ErrScheduledAtRequired
	}
	if result.SequenceNumber < 0 {
		return Dose{}, ErrSequenceNumberInvalid
	}
	if err := result.Unit.Validate(); err != nil {
		return Dose{}, err
	}

	return result, nil
}

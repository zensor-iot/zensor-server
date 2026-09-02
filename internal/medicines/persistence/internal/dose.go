package internal

import (
	"zensor-server/internal/infra/utils"

	medicinesDomain "zensor-server/internal/medicines/domain"
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

// Dose is a materialised dose row.
//
// The composite unique index on (treatment_id, sequence_number) is what makes
// materialisation idempotent: re-running it recomputes the same sequence
// numbers, and the index is the backstop for the read then write race.
type Dose struct {
	ID             string      `json:"id" gorm:"primaryKey"`
	Version        int         `json:"version"`
	TreatmentID    string      `json:"treatment_id" gorm:"index;not null;uniqueIndex:idx_medicine_doses_treatment_sequence"`
	SequenceNumber int         `json:"sequence_number" gorm:"not null;uniqueIndex:idx_medicine_doses_treatment_sequence"`
	ScheduledAt    utils.Time  `json:"scheduled_at" gorm:"index;not null"`
	Status         string      `json:"status" gorm:"index;not null"`
	Quantity       float64     `json:"quantity"`
	Unit           string      `json:"unit"`
	ResolvedAt     *utils.Time `json:"resolved_at,omitempty"`
	ResolvedBy     *string     `json:"resolved_by,omitempty"`
	Notes          *string     `json:"notes,omitempty"`
	ReminderSentAt *utils.Time `json:"reminder_sent_at,omitempty"`
	CreatedAt      utils.Time  `json:"created_at"`
	UpdatedAt      utils.Time  `json:"updated_at"`
	DeletedAt      *utils.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func (Dose) TableName() string {
	return "medicine_doses"
}

func (m Dose) ToDomain() medicinesDomain.Dose {
	result := medicinesDomain.Dose{
		ID:             shareddomain.ID(m.ID),
		Version:        shareddomain.Version(m.Version),
		TreatmentID:    shareddomain.ID(m.TreatmentID),
		SequenceNumber: m.SequenceNumber,
		ScheduledAt:    m.ScheduledAt,
		Status:         medicinesDomain.DoseStatus(m.Status),
		Quantity:       medicinesDomain.DoseQuantity(m.Quantity),
		Unit:           medicinesDomain.DoseUnit(m.Unit),
		ResolvedAt:     m.ResolvedAt,
		ReminderSentAt: m.ReminderSentAt,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		DeletedAt:      m.DeletedAt,
	}

	if m.ResolvedBy != nil {
		resolvedBy := medicinesDomain.ResolvedBy(*m.ResolvedBy)
		result.ResolvedBy = &resolvedBy
	}

	if m.Notes != nil {
		notes := medicinesDomain.DoseNotes(*m.Notes)
		result.Notes = &notes
	}

	return result
}

func FromDose(value medicinesDomain.Dose) Dose {
	// Doses are stored as UTC instants so range queries compare like with like.
	// Callers render them in the tenant timezone when they display them.
	scheduledAt := utils.Time{Time: value.ScheduledAt.UTC()}

	result := Dose{
		ID:             value.ID.String(),
		Version:        int(value.Version),
		TreatmentID:    value.TreatmentID.String(),
		SequenceNumber: value.SequenceNumber,
		ScheduledAt:    scheduledAt,
		Status:         string(value.Status),
		Quantity:       float64(value.Quantity),
		Unit:           string(value.Unit),
		ResolvedAt:     value.ResolvedAt,
		ReminderSentAt: value.ReminderSentAt,
		CreatedAt:      value.CreatedAt,
		UpdatedAt:      value.UpdatedAt,
		DeletedAt:      value.DeletedAt,
	}

	if value.ResolvedBy != nil {
		resolvedBy := string(*value.ResolvedBy)
		result.ResolvedBy = &resolvedBy
	}

	if value.Notes != nil {
		notes := string(*value.Notes)
		result.Notes = &notes
	}

	return result
}

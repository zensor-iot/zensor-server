package internal

import (
	"time"

	medicinesDomain "zensor-server/internal/medicines/domain"
	medicinesUsecases "zensor-server/internal/medicines/usecases"
)

type DoseResponse struct {
	ID             string     `json:"id"`
	Version        int        `json:"version"`
	TreatmentID    string     `json:"treatment_id"`
	SequenceNumber int        `json:"sequence_number"`
	ScheduledAt    time.Time  `json:"scheduled_at"`
	Status         string     `json:"status"`
	Quantity       float64    `json:"quantity"`
	Unit           string     `json:"unit"`
	IsOverdue      bool       `json:"is_overdue"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy     *string    `json:"resolved_by,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// DoseResolveRequest deliberately carries no actor: who resolved the dose comes
// from the authenticated request headers, not from the body, so the record is
// auditable.
type DoseResolveRequest struct {
	Notes *string `json:"notes,omitempty"`
}

type AgendaEntryResponse struct {
	Dose         DoseResponse `json:"dose"`
	TreatmentID  string       `json:"treatment_id"`
	MedicineName string       `json:"medicine_name"`
	DoseText     string       `json:"dose_text"`
	PatientID    string       `json:"patient_id"`
	PatientName  string       `json:"patient_name"`
	PatientKind  string       `json:"patient_kind"`
}

type AgendaResponse struct {
	Data []AgendaEntryResponse `json:"data"`
}

func ToDoseResponse(dose medicinesDomain.Dose) DoseResponse {
	response := DoseResponse{
		ID:             dose.ID.String(),
		Version:        int(dose.Version),
		TreatmentID:    dose.TreatmentID.String(),
		SequenceNumber: dose.SequenceNumber,
		ScheduledAt:    dose.ScheduledAt.Time,
		Status:         string(dose.Status),
		Quantity:       float64(dose.Quantity),
		Unit:           string(dose.Unit),
		IsOverdue:      dose.IsOverdue(time.Now()),
		CreatedAt:      dose.CreatedAt.Time,
		UpdatedAt:      dose.UpdatedAt.Time,
	}

	if dose.ResolvedAt != nil {
		response.ResolvedAt = &dose.ResolvedAt.Time
	}

	if dose.ResolvedBy != nil {
		resolvedBy := string(*dose.ResolvedBy)
		response.ResolvedBy = &resolvedBy
	}

	if dose.Notes != nil {
		notes := string(*dose.Notes)
		response.Notes = &notes
	}

	return response
}

func ToAgendaEntryResponse(entry medicinesUsecases.DoseWithContext) AgendaEntryResponse {
	return AgendaEntryResponse{
		Dose:         ToDoseResponse(entry.Dose),
		TreatmentID:  entry.Treatment.ID.String(),
		MedicineName: string(entry.Treatment.MedicineName),
		DoseText:     entry.Treatment.DoseText(),
		PatientID:    entry.Patient.ID.String(),
		PatientName:  string(entry.Patient.Name),
		PatientKind:  string(entry.Patient.Kind),
	}
}

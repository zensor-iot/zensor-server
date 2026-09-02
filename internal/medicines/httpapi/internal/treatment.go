package internal

import (
	"time"

	medicinesDomain "zensor-server/internal/medicines/domain"
)

type MedicineScheduleRequest struct {
	StartAt time.Time `json:"start_at"`
	Every   int       `json:"every"`
	Unit    string    `json:"unit"`
}

type MedicineScheduleResponse struct {
	StartAt time.Time `json:"start_at"`
	Every   int       `json:"every"`
	Unit    string    `json:"unit"`
}

type TreatmentResponse struct {
	ID           string                   `json:"id"`
	Version      int                      `json:"version"`
	TenantID     string                   `json:"tenant_id"`
	PatientID    string                   `json:"patient_id"`
	MedicineName string                   `json:"medicine_name"`
	Quantity     float64                  `json:"quantity"`
	Unit         string                   `json:"unit"`
	DoseText     string                   `json:"dose_text"`
	Schedule     MedicineScheduleResponse `json:"schedule"`
	EndAt        *time.Time               `json:"end_at,omitempty"`
	TotalDoses   *int                     `json:"total_doses,omitempty"`
	EndsAt       *time.Time               `json:"ends_at,omitempty"`
	Notes        string                   `json:"notes"`
	IsActive     bool                     `json:"is_active"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
}

type TreatmentCreateRequest struct {
	TenantID     string                  `json:"tenant_id"`
	PatientID    string                  `json:"patient_id"`
	MedicineName string                  `json:"medicine_name"`
	Quantity     float64                 `json:"quantity"`
	Unit         string                  `json:"unit"`
	Schedule     MedicineScheduleRequest `json:"schedule"`
	EndAt        *time.Time              `json:"end_at,omitempty"`
	TotalDoses   *int                    `json:"total_doses,omitempty"`
	Notes        string                  `json:"notes"`
}

type TreatmentUpdateRequest struct {
	MedicineName *string                  `json:"medicine_name,omitempty"`
	Quantity     *float64                 `json:"quantity,omitempty"`
	Unit         *string                  `json:"unit,omitempty"`
	Schedule     *MedicineScheduleRequest `json:"schedule,omitempty"`
	EndAt        *time.Time               `json:"end_at,omitempty"`
	TotalDoses   *int                     `json:"total_doses,omitempty"`
	Notes        *string                  `json:"notes,omitempty"`
}

func ToTreatmentResponse(treatment medicinesDomain.Treatment) TreatmentResponse {
	return TreatmentResponse{
		ID:           treatment.ID.String(),
		Version:      int(treatment.Version),
		TenantID:     treatment.TenantID.String(),
		PatientID:    treatment.PatientID.String(),
		MedicineName: string(treatment.MedicineName),
		Quantity:     float64(treatment.Quantity),
		Unit:         string(treatment.Unit),
		DoseText:     treatment.DoseText(),
		Schedule: MedicineScheduleResponse{
			StartAt: treatment.Schedule.StartAt,
			Every:   treatment.Schedule.Every,
			Unit:    string(treatment.Schedule.Unit),
		},
		EndAt:      treatment.EndAt,
		TotalDoses: treatment.TotalDoses,
		EndsAt:     treatment.EffectiveEndAt(),
		Notes:      string(treatment.Notes),
		IsActive:   treatment.IsActive,
		CreatedAt:  treatment.CreatedAt.Time,
		UpdatedAt:  treatment.UpdatedAt.Time,
	}
}

func ToMedicineSchedule(request MedicineScheduleRequest) medicinesDomain.MedicineSchedule {
	return medicinesDomain.MedicineSchedule{
		StartAt: request.StartAt,
		Every:   request.Every,
		Unit:    medicinesDomain.IntervalUnit(request.Unit),
	}
}

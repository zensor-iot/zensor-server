package internal

import (
	"encoding/json"
	"log/slog"
	"time"
	"zensor-server/internal/infra/utils"

	medicinesDomain "zensor-server/internal/medicines/domain"
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

type Treatment struct {
	ID           string     `json:"id" gorm:"primaryKey"`
	Version      int        `json:"version"`
	TenantID     string     `json:"tenant_id" gorm:"index;not null"`
	PatientID    string     `json:"patient_id" gorm:"index;not null"`
	MedicineName string     `json:"medicine_name" gorm:"not null"`
	Quantity     float64    `json:"quantity" gorm:"not null"`
	Unit         string     `json:"unit" gorm:"not null"`
	Schedule     string     `json:"schedule" gorm:"not null"`
	EndAt        *time.Time `json:"end_at,omitempty"`
	TotalDoses   *int       `json:"total_doses,omitempty"`
	Notes        string     `json:"notes"`
	// IsActive carries no gorm default: GORM omits a false value when the
	// column has one, which would silently revive a treatment created paused.
	IsActive  bool        `json:"is_active" gorm:"index"`
	CreatedAt utils.Time  `json:"created_at"`
	UpdatedAt utils.Time  `json:"updated_at"`
	DeletedAt *utils.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func (Treatment) TableName() string {
	return "medicine_treatments"
}

func (m Treatment) ToDomain() medicinesDomain.Treatment {
	result := medicinesDomain.Treatment{
		ID:           shareddomain.ID(m.ID),
		Version:      shareddomain.Version(m.Version),
		TenantID:     shareddomain.ID(m.TenantID),
		PatientID:    shareddomain.ID(m.PatientID),
		MedicineName: medicinesDomain.MedicineName(m.MedicineName),
		Quantity:     medicinesDomain.DoseQuantity(m.Quantity),
		Unit:         medicinesDomain.DoseUnit(m.Unit),
		EndAt:        m.EndAt,
		TotalDoses:   m.TotalDoses,
		Notes:        shareddomain.Description(m.Notes),
		IsActive:     m.IsActive,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		DeletedAt:    m.DeletedAt,
	}

	var schedule medicinesDomain.MedicineSchedule
	if err := json.Unmarshal([]byte(m.Schedule), &schedule); err != nil {
		slog.Error("unmarshalling stored treatment schedule",
			slog.String("treatment_id", m.ID),
			slog.String("error", err.Error()))
		return result
	}

	if err := schedule.Validate(); err != nil {
		slog.Error("stored treatment schedule is invalid",
			slog.String("treatment_id", m.ID),
			slog.String("error", err.Error()))
	}
	result.Schedule = schedule

	return result
}

func FromTreatment(value medicinesDomain.Treatment) Treatment {
	result := Treatment{
		ID:           value.ID.String(),
		Version:      int(value.Version),
		TenantID:     value.TenantID.String(),
		PatientID:    value.PatientID.String(),
		MedicineName: string(value.MedicineName),
		Quantity:     float64(value.Quantity),
		Unit:         string(value.Unit),
		EndAt:        value.EndAt,
		TotalDoses:   value.TotalDoses,
		Notes:        string(value.Notes),
		IsActive:     value.IsActive,
		CreatedAt:    value.CreatedAt,
		UpdatedAt:    value.UpdatedAt,
		DeletedAt:    value.DeletedAt,
	}

	encoded, err := json.Marshal(value.Schedule)
	if err != nil {
		slog.Error("marshalling treatment schedule",
			slog.String("treatment_id", result.ID),
			slog.String("error", err.Error()))
		result.Schedule = "{}"
		return result
	}
	result.Schedule = string(encoded)

	return result
}

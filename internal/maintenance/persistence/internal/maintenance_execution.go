package internal

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"zensor-server/internal/infra/utils"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

type MaintenanceExecution struct {
	ID            string      `json:"id" gorm:"primaryKey"`
	Version       int         `json:"version"`
	ActivityID    string      `json:"activity_id" gorm:"index;not null"`
	ScheduledDate utils.Time  `json:"scheduled_date" gorm:"not null"`
	CompletedAt   *utils.Time `json:"completed_at,omitempty"`
	CompletedBy   *string     `json:"completed_by,omitempty"`
	OverdueDays   int         `json:"overdue_days" gorm:"default:0"`
	FieldValues   FieldValues `json:"field_values"`
	CreatedAt     utils.Time  `json:"created_at"`
	UpdatedAt     utils.Time  `json:"updated_at"`
	DeletedAt     *utils.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func (MaintenanceExecution) TableName() string {
	return "maintenance_executions"
}

type FieldValues map[string]any

func (fv FieldValues) Value() (driver.Value, error) {
	if len(fv) == 0 {
		return "{}", nil
	}
	return json.Marshal(fv)
}

func (fv *FieldValues) Scan(src any) error {
	var data []byte

	switch val := src.(type) {
	case string:
		if val == "" {
			*fv = make(map[string]any)
			return nil
		}
		data = []byte(val)
	case []byte:
		if len(val) == 0 {
			*fv = make(map[string]any)
			return nil
		}
		data = val
	case nil:
		*fv = make(map[string]any)
		return nil
	default:
		return errors.New("invalid type for field_values")
	}

	return json.Unmarshal(data, fv)
}

func (m MaintenanceExecution) ToDomain() maintenanceDomain.MaintenanceExecution {
	result := maintenanceDomain.MaintenanceExecution{
		ID:            shareddomain.ID(m.ID),
		Version:       shareddomain.Version(m.Version),
		ActivityID:    shareddomain.ID(m.ActivityID),
		ScheduledDate: m.ScheduledDate,
		OverdueDays:   maintenanceDomain.OverdueDays(m.OverdueDays),
		FieldValues:   map[string]any(m.FieldValues),
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}

	if m.CompletedAt != nil {
		result.CompletedAt = m.CompletedAt
	}

	if m.CompletedBy != nil {
		completedBy := maintenanceDomain.CompletedBy(*m.CompletedBy)
		result.CompletedBy = &completedBy
	}

	if m.DeletedAt != nil {
		result.DeletedAt = m.DeletedAt
	}

	return result
}

func FromMaintenanceExecution(value maintenanceDomain.MaintenanceExecution) MaintenanceExecution {
	result := MaintenanceExecution{
		ID:            value.ID.String(),
		Version:       int(value.Version),
		ActivityID:    value.ActivityID.String(),
		ScheduledDate: value.ScheduledDate,
		OverdueDays:   int(value.OverdueDays),
		FieldValues:   FieldValues(value.FieldValues),
		CreatedAt:     value.CreatedAt,
		UpdatedAt:     value.UpdatedAt,
	}

	if value.CompletedAt != nil {
		result.CompletedAt = value.CompletedAt
	}

	if value.CompletedBy != nil {
		completedBy := string(*value.CompletedBy)
		result.CompletedBy = &completedBy
	}

	if value.DeletedAt != nil {
		result.DeletedAt = value.DeletedAt
	}

	return result
}

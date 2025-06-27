package internal

import (
	"encoding/json"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/infra/utils"
)

type ScheduledTask struct {
	ID               string      `json:"id" gorm:"primaryKey"`
	Version          uint        `json:"version"`
	TenantID         string      `json:"tenant_id"`
	DeviceID         string      `json:"device_id"`
	CommandTemplates string      `json:"command_templates"` // JSON array of command templates
	Schedule         string      `json:"schedule"`
	IsActive         bool        `json:"is_active"`
	CreatedAt        utils.Time  `json:"created_at"`
	UpdatedAt        utils.Time  `json:"updated_at"`
	LastExecutedAt   *utils.Time `json:"last_executed_at"`
}

func (ScheduledTask) TableName() string {
	return "scheduled_tasks_final"
}

func FromScheduledTask(value domain.ScheduledTask) ScheduledTask {
	commandTemplatesJSON, _ := json.Marshal(value.CommandTemplates)

	return ScheduledTask{
		ID:               value.ID.String(),
		Version:          uint(value.Version),
		TenantID:         value.Tenant.ID.String(),
		DeviceID:         value.Device.ID.String(),
		CommandTemplates: string(commandTemplatesJSON),
		Schedule:         value.Schedule,
		IsActive:         value.IsActive,
		CreatedAt:        value.CreatedAt,
		UpdatedAt:        value.UpdatedAt,
		LastExecutedAt:   value.LastExecutedAt,
	}
}

func (s ScheduledTask) ToDomain() domain.ScheduledTask {
	var commandTemplates []domain.CommandTemplate
	json.Unmarshal([]byte(s.CommandTemplates), &commandTemplates)

	return domain.ScheduledTask{
		ID:               domain.ID(s.ID),
		Version:          domain.Version(s.Version),
		Tenant:           domain.Tenant{ID: domain.ID(s.TenantID)},
		Device:           domain.Device{ID: domain.ID(s.DeviceID)},
		CommandTemplates: commandTemplates,
		Schedule:         s.Schedule,
		IsActive:         s.IsActive,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
		LastExecutedAt:   s.LastExecutedAt,
	}
}

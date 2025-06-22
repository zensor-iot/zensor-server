package internal

import (
	"time"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/infra/utils"
)

type ScheduledTask struct {
	ID        string     `json:"id" gorm:"primaryKey"`
	Version   uint       `json:"version"`
	TenantID  string     `json:"tenant_id"`
	DeviceID  string     `json:"device_id"`
	TaskID    string     `json:"task_id"`
	Schedule  string     `json:"schedule"`
	IsActive  bool       `json:"is_active"`
	CreatedAt utils.Time `json:"created_at"`
	UpdatedAt utils.Time `json:"updated_at"`
}

func (ScheduledTask) TableName() string {
	return "scheduled_tasks_final"
}

func FromScheduledTask(value domain.ScheduledTask) ScheduledTask {
	return ScheduledTask{
		ID:        value.ID.String(),
		Version:   uint(value.Version),
		TenantID:  value.Tenant.ID.String(),
		DeviceID:  value.Device.ID.String(),
		TaskID:    value.Task.ID.String(),
		Schedule:  value.Schedule,
		IsActive:  value.IsActive,
		CreatedAt: utils.Time{Time: time.Now()},
		UpdatedAt: utils.Time{Time: time.Now()},
	}
}

func (s ScheduledTask) ToDomain() domain.ScheduledTask {
	return domain.ScheduledTask{
		ID:       domain.ID(s.ID),
		Version:  domain.Version(s.Version),
		Tenant:   domain.Tenant{ID: domain.ID(s.TenantID)},
		Device:   domain.Device{ID: domain.ID(s.DeviceID)},
		Task:     domain.Task{ID: domain.ID(s.TaskID)},
		Schedule: s.Schedule,
		IsActive: s.IsActive,
	}
}

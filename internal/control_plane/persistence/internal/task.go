package internal

import (
	"time"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/infra/utils"
)

type Task struct {
	ID              string     `json:"id" gorm:"primaryKey"`
	DeviceID        string     `json:"device_id" gorm:"foreignKey:device_id"`
	ScheduledTaskID string     `json:"scheduled_task_id,omitempty"` // UUID of the scheduled task
	Version         uint       `json:"version"`
	CreatedAt       utils.Time `json:"created_at"`
	UpdatedAt       utils.Time `json:"updated_at"`
}

func (Task) TableName() string {
	return "tasks"
}

func FromTask(value domain.Task) Task {
	var scheduledTaskID string
	if value.ScheduledTask != nil {
		scheduledTaskID = value.ScheduledTask.ID.String()
	}

	return Task{
		ID:              value.ID.String(),
		DeviceID:        value.Device.ID.String(),
		ScheduledTaskID: scheduledTaskID,
		Version:         uint(value.Version),
		CreatedAt:       value.CreatedAt,
		UpdatedAt:       utils.Time{Time: time.Now()},
	}
}

func (t Task) ToDomain() domain.Task {
	return domain.Task{
		ID:        domain.ID(t.ID),
		Device:    domain.Device{ID: domain.ID(t.DeviceID)},
		Version:   domain.Version(t.Version),
		CreatedAt: t.CreatedAt,
		ScheduledTask: &domain.ScheduledTask{
			ID: domain.ID(t.ScheduledTaskID),
		},
	}
}

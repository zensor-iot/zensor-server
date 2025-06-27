package internal

import (
	"encoding/json"
	"time"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/infra/utils"
)

type Task struct {
	ID            string     `json:"id" gorm:"primaryKey"`
	DeviceID      string     `json:"device_id" gorm:"foreignKey:device_id"`
	ScheduledTask string     `json:"scheduled_task,omitempty"` // JSON representation of ScheduledTask
	Version       uint       `json:"version"`
	CreatedAt     utils.Time `json:"created_at"`
	UpdatedAt     utils.Time `json:"updated_at"`
}

func (Task) TableName() string {
	return "tasks_final"
}

func FromTask(value domain.Task) Task {
	var scheduledTaskJSON string
	if value.ScheduledTask != nil {
		scheduledTaskData := FromScheduledTask(*value.ScheduledTask)
		jsonData, _ := json.Marshal(scheduledTaskData)
		scheduledTaskJSON = string(jsonData)
	}

	return Task{
		ID:            value.ID.String(),
		DeviceID:      value.Device.ID.String(),
		ScheduledTask: scheduledTaskJSON,
		Version:       uint(value.Version),
		CreatedAt:     utils.Time{Time: time.Now()},
		UpdatedAt:     utils.Time{Time: time.Now()},
	}
}

func (t Task) ToDomain() domain.Task {
	var scheduledTask *domain.ScheduledTask
	if t.ScheduledTask != "" {
		var scheduledTaskData ScheduledTask
		json.Unmarshal([]byte(t.ScheduledTask), &scheduledTaskData)
		domainScheduledTask := scheduledTaskData.ToDomain()
		scheduledTask = &domainScheduledTask
	}

	return domain.Task{
		ID:            domain.ID(t.ID),
		Device:        domain.Device{ID: domain.ID(t.DeviceID)},
		ScheduledTask: scheduledTask,
		Version:       domain.Version(t.Version),
	}
}

package internal

import (
	"time"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/infra/utils"
)

type Task struct {
	ID         string     `json:"id" gorm:"primaryKey"`
	DeviceID   string     `json:"device_id" gorm:"foreignKey:device_id"`
	Version    uint       `json:"version"`
	CreatedAt  utils.Time `json:"created_at"`
	UpdaatedAt utils.Time `json:"updated_at"`
}

func FromTask(value domain.Task) Task {
	return Task{
		ID:         value.ID.String(),
		DeviceID:   value.Device.ID.String(),
		Version:    1,
		CreatedAt:  utils.Time{Time: time.Now()},
		UpdaatedAt: utils.Time{Time: time.Now()},
	}
}

package internal

import (
	"time"
	"zensor-server/internal/control_plane/domain"
)

type Device struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	Version    int       `json:"version"`
	Name       string    `json:"name"`
	AppEUI     string    `json:"app_eui"`
	DevEUI     string    `json:"dev_eui"`
	AppKey     string    `json:"app_key"`
	CreatedAt  time.Time `json:"created_at"`
	UpdaatedAt time.Time `json:"updated_at"`
}

func (Device) TableName() string {
	return "devices_final"
}

func (s Device) ToDomain() domain.Device {
	return domain.Device{
		ID:     s.ID,
		Name:   s.Name,
		AppEUI: s.AppEUI,
		DevEUI: s.DevEUI,
		AppKey: s.AppKey,
	}
}

func FromDevice(value domain.Device) Device {
	return Device{
		ID:         value.ID,
		Version:    1,
		Name:       value.Name,
		AppEUI:     value.AppEUI,
		DevEUI:     value.DevEUI,
		AppKey:     value.AppKey,
		CreatedAt:  time.Now(),
		UpdaatedAt: time.Now(),
	}
}

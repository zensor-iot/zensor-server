package internal

import (
	"time"
	"zensor-server/internal/control_plane/domain"
)

type Device struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Version     int       `json:"version"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	AppEUI      string    `json:"app_eui" gorm:"column:app_eui"`
	DevEUI      string    `json:"dev_eui" gorm:"column:dev_eui"`
	AppKey      string    `json:"app_key"`
	TenantID    *string   `json:"tenant_id,omitempty" gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdaatedAt  time.Time `json:"updated_at"`
}

func (Device) TableName() string {
	return "devices_final"
}

func (s Device) ToDomain() domain.Device {
	device := domain.Device{
		ID:          domain.ID(s.ID),
		Name:        s.Name,
		DisplayName: s.DisplayName,
		AppEUI:      s.AppEUI,
		DevEUI:      s.DevEUI,
		AppKey:      s.AppKey,
	}

	if s.TenantID != nil {
		tenantID := domain.ID(*s.TenantID)
		device.TenantID = &tenantID
	}

	return device
}

func FromDevice(value domain.Device) Device {
	device := Device{
		ID:          value.ID.String(),
		Version:     1,
		Name:        value.Name,
		DisplayName: value.DisplayName,
		AppEUI:      value.AppEUI,
		DevEUI:      value.DevEUI,
		AppKey:      value.AppKey,
		CreatedAt:   time.Now(),
		UpdaatedAt:  time.Now(),
	}

	if value.TenantID != nil {
		tenantIDStr := value.TenantID.String()
		device.TenantID = &tenantIDStr
	}

	return device
}

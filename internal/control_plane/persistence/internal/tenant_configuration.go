package internal

import (
	"time"
	"zensor-server/internal/shared_kernel/domain"
)

type TenantConfiguration struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	TenantID  string    `json:"tenant_id" gorm:"uniqueIndex;not null"`
	Timezone  string    `json:"timezone" gorm:"not null"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (TenantConfiguration) TableName() string {
	return "tenant_configurations"
}

func (tc TenantConfiguration) ToDomain() domain.TenantConfiguration {
	return domain.TenantConfiguration{
		ID:        domain.ID(tc.ID),
		TenantID:  domain.ID(tc.TenantID),
		Timezone:  tc.Timezone,
		Version:   tc.Version,
		CreatedAt: tc.CreatedAt,
		UpdatedAt: tc.UpdatedAt,
	}
}

func FromTenantConfiguration(value domain.TenantConfiguration) TenantConfiguration {
	return TenantConfiguration{
		ID:        value.ID.String(),
		TenantID:  value.TenantID.String(),
		Timezone:  value.Timezone,
		Version:   value.Version,
		CreatedAt: value.CreatedAt,
		UpdatedAt: value.UpdatedAt,
	}
}

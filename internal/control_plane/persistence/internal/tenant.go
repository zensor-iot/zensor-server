package internal

import (
	"time"
	"zensor-server/internal/control_plane/domain"
)

type Tenant struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	Version     int        `json:"version"`
	Name        string     `json:"name" gorm:"uniqueIndex;not null"`
	Email       string     `json:"email" gorm:"not null"`
	Description string     `json:"description"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func (Tenant) TableName() string {
	return "tenants"
}

func (t Tenant) ToDomain() domain.Tenant {
	return domain.Tenant{
		ID:          domain.ID(t.ID),
		Name:        t.Name,
		Email:       t.Email,
		Description: t.Description,
		IsActive:    t.IsActive,
		Version:     t.Version,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
		DeletedAt:   t.DeletedAt,
	}
}

func FromTenant(value domain.Tenant) Tenant {
	return Tenant{
		ID:          value.ID.String(),
		Version:     value.Version,
		Name:        value.Name,
		Email:       value.Email,
		Description: value.Description,
		IsActive:    value.IsActive,
		CreatedAt:   value.CreatedAt,
		UpdatedAt:   value.UpdatedAt,
		DeletedAt:   value.DeletedAt,
	}
}

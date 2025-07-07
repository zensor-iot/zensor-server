package internal

import (
	"time"
	"zensor-server/internal/shared_kernel/domain"
)

// Request models
type TenantCreateRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Email       string `json:"email" validate:"required,email"`
	Description string `json:"description" validate:"max=500"`
}

type TenantUpdateRequest struct {
	Name        string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Email       string `json:"email,omitempty" validate:"omitempty,email"`
	Description string `json:"description,omitempty" validate:"omitempty,max=500"`
	Version     int    `json:"version,omitempty"`
}

type TenantAdoptDeviceRequest struct {
	DeviceID string `json:"device_id" validate:"required,uuid"`
}

// Response models
type TenantResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Description string     `json:"description"`
	IsActive    bool       `json:"is_active"`
	Version     int        `json:"version"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type TenantListResponse struct {
	Tenants []TenantResponse `json:"tenants"`
	Total   int              `json:"total"`
}

// Conversion functions
func ToTenantResponse(tenant domain.Tenant) TenantResponse {
	return TenantResponse{
		ID:          tenant.ID.String(),
		Name:        tenant.Name,
		Email:       tenant.Email,
		Description: tenant.Description,
		IsActive:    tenant.IsActive,
		Version:     tenant.Version,
		CreatedAt:   tenant.CreatedAt,
		UpdatedAt:   tenant.UpdatedAt,
		DeletedAt:   tenant.DeletedAt,
	}
}

func ToTenantListResponse(tenants []domain.Tenant) TenantListResponse {
	responses := make([]TenantResponse, len(tenants))
	for i, tenant := range tenants {
		responses[i] = ToTenantResponse(tenant)
	}

	return TenantListResponse{
		Tenants: responses,
		Total:   len(tenants),
	}
}

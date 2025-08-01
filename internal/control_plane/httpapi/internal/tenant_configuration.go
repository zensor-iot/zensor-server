package internal

import (
	"time"
	"zensor-server/internal/shared_kernel/domain"
)

// TenantConfigurationResponse represents the response for tenant configuration operations
type TenantConfigurationResponse struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Timezone  string    `json:"timezone"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TenantConfigurationCreateRequest represents the request for creating a tenant configuration
type TenantConfigurationCreateRequest struct {
	Timezone string `json:"timezone" validate:"required"`
}

// TenantConfigurationUpdateRequest represents the request for updating a tenant configuration
type TenantConfigurationUpdateRequest struct {
	Timezone string `json:"timezone" validate:"required"`
}

// ToTenantConfigurationResponse converts a domain.TenantConfiguration to TenantConfigurationResponse
func ToTenantConfigurationResponse(config domain.TenantConfiguration) TenantConfigurationResponse {
	return TenantConfigurationResponse{
		ID:        config.ID.String(),
		TenantID:  config.TenantID.String(),
		Timezone:  config.Timezone,
		Version:   config.Version,
		CreatedAt: config.CreatedAt,
		UpdatedAt: config.UpdatedAt,
	}
}

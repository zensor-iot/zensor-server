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

type DeviceResponse struct {
	ID                    string     `json:"id"`
	Name                  string     `json:"name"`
	DisplayName           string     `json:"display_name"`
	AppEUI                string     `json:"app_eui"`
	DevEUI                string     `json:"dev_eui"`
	AppKey                string     `json:"app_key"`
	TenantID              *string    `json:"tenant_id,omitempty"`
	Status                string     `json:"status"`
	LastMessageReceivedAt *time.Time `json:"last_message_received_at,omitempty"`
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

func ToDeviceResponse(device domain.Device) DeviceResponse {
	response := DeviceResponse{
		ID:          device.ID.String(),
		Name:        device.Name,
		DisplayName: device.DisplayName,
		AppEUI:      device.AppEUI,
		DevEUI:      device.DevEUI,
		AppKey:      device.AppKey,
		Status:      device.GetStatus(),
	}

	if !device.LastMessageReceivedAt.IsZero() {
		response.LastMessageReceivedAt = &device.LastMessageReceivedAt.Time
	}

	if device.TenantID != nil {
		tenantIDStr := device.TenantID.String()
		response.TenantID = &tenantIDStr
	}

	return response
}

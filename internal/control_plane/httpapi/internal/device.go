package internal

import (
	"time"
	"zensor-server/internal/shared_kernel/domain"
)

type DeviceListResponse struct {
	Data []DeviceResponse `json:"data"`
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

type DeviceCreateRequest struct {
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	AppEUI      *string `json:"app_eui,omitempty"`
	DevEUI      *string `json:"dev_eui,omitempty"`
	AppKey      *string `json:"app_key,omitempty"`
}

type DeviceUpdateRequest struct {
	DisplayName string `json:"display_name" validate:"required,min=1,max=100"`
}

// ToDeviceResponse converts a domain Device to DeviceResponse
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

	// Convert utils.Time to *time.Time
	if !device.LastMessageReceivedAt.IsZero() {
		response.LastMessageReceivedAt = &device.LastMessageReceivedAt.Time
	}

	if device.TenantID != nil {
		tenantIDStr := device.TenantID.String()
		response.TenantID = &tenantIDStr
	}

	return response
}

// ToDeviceListResponse converts a slice of domain Devices to DeviceListResponse
func ToDeviceListResponse(devices []domain.Device) DeviceListResponse {
	responses := make([]DeviceResponse, len(devices))
	for i, device := range devices {
		responses[i] = ToDeviceResponse(device)
	}

	return DeviceListResponse{
		Data: responses,
	}
}

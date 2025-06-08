package internal

import "zensor-server/internal/control_plane/domain"

type DeviceListResponse struct {
	Data []DeviceResponse `json:"data"`
}

type DeviceResponse struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	AppEUI   string  `json:"app_eui"`
	DevEUI   string  `json:"dev_eui"`
	AppKey   string  `json:"app_key"`
	TenantID *string `json:"tenant_id,omitempty"`
}

type DeviceCreateRequest struct {
	Name string `json:"name"`
}

// ToDeviceResponse converts a domain Device to DeviceResponse
func ToDeviceResponse(device domain.Device) DeviceResponse {
	response := DeviceResponse{
		ID:     device.ID.String(),
		Name:   device.Name,
		AppEUI: device.AppEUI,
		DevEUI: device.DevEUI,
		AppKey: device.AppKey,
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

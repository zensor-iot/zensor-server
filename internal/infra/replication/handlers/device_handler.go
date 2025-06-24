package handlers

import (
	"context"
	"fmt"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
)

// DeviceHandler handles replication of device data
type DeviceHandler struct {
	orm sql.ORM
}

// NewDeviceHandler creates a new device handler
func NewDeviceHandler(orm sql.ORM) *DeviceHandler {
	return &DeviceHandler{
		orm: orm,
	}
}

// TopicName returns the devices topic
func (h *DeviceHandler) TopicName() pubsub.Topic {
	return "devices"
}

// Create handles creating a new device record
func (h *DeviceHandler) Create(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	device, ok := message.(domain.Device)
	if !ok {
		return fmt.Errorf("invalid message type for device: %T", message)
	}

	// Convert domain device to internal representation for persistence
	internalDevice := h.toInternalDevice(device)

	err := h.orm.WithContext(ctx).Create(&internalDevice).Error()
	if err != nil {
		return fmt.Errorf("creating device: %w", err)
	}

	return nil
}

// GetByID retrieves a device by its ID
func (h *DeviceHandler) GetByID(ctx context.Context, id string) (pubsub.Message, error) {
	var internalDevice struct {
		ID                    string `gorm:"primaryKey"`
		Version               int
		Name                  string
		DisplayName           string
		AppEUI                string `gorm:"column:app_eui"`
		DevEUI                string `gorm:"column:dev_eui"`
		AppKey                string
		TenantID              *string `gorm:"index"`
		LastMessageReceivedAt string
		CreatedAt             string
		UpdatedAt             string
	}

	err := h.orm.WithContext(ctx).First(&internalDevice, "id = ?", id).Error()
	if err != nil {
		return nil, fmt.Errorf("getting device: %w", err)
	}

	// Convert to domain model
	device := h.toDomainDevice(internalDevice)
	return device, nil
}

// Update handles updating an existing device record
func (h *DeviceHandler) Update(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	device, ok := message.(domain.Device)
	if !ok {
		return fmt.Errorf("invalid message type for device: %T", message)
	}

	// Convert domain device to internal representation for persistence
	internalDevice := h.toInternalDevice(device)

	err := h.orm.WithContext(ctx).Save(&internalDevice).Error()
	if err != nil {
		return fmt.Errorf("updating device: %w", err)
	}

	return nil
}

// toInternalDevice converts domain device to internal representation
func (h *DeviceHandler) toInternalDevice(device domain.Device) map[string]any {
	result := map[string]any{
		"id":           device.ID.String(),
		"version":      1,
		"name":         device.Name,
		"display_name": device.DisplayName,
		"app_eui":      device.AppEUI,
		"dev_eui":      device.DevEUI,
		"app_key":      device.AppKey,
	}

	if device.TenantID != nil {
		result["tenant_id"] = device.TenantID.String()
	}

	return result
}

// toDomainDevice converts internal device to domain representation
func (h *DeviceHandler) toDomainDevice(internalDevice any) domain.Device {
	// This is a simplified conversion - in practice you'd want more robust mapping
	deviceMap := internalDevice.(map[string]any)

	device := domain.Device{
		ID:          domain.ID(deviceMap["id"].(string)),
		Name:        deviceMap["name"].(string),
		DisplayName: deviceMap["display_name"].(string),
		AppEUI:      deviceMap["app_eui"].(string),
		DevEUI:      deviceMap["dev_eui"].(string),
		AppKey:      deviceMap["app_key"].(string),
	}

	if tenantID, ok := deviceMap["tenant_id"].(string); ok {
		tenantIDDomain := domain.ID(tenantID)
		device.TenantID = &tenantIDDomain
	}

	return device
}

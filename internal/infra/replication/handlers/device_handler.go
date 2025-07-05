package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/shared_kernel/avro"
)

type DeviceData struct {
	ID                    string     `json:"id" gorm:"primaryKey"`
	Version               int        `json:"version"`
	Name                  string     `json:"name"`
	DisplayName           string     `json:"display_name"`
	AppEUI                string     `json:"app_eui" gorm:"column:app_eui"`
	DevEUI                string     `json:"dev_eui" gorm:"column:dev_eui"`
	AppKey                string     `json:"app_key"`
	TenantID              *string    `json:"tenant_id" gorm:"index"`
	LastMessageReceivedAt *time.Time `json:"last_message_received_at"`
	CreatedAt             time.Time  `json:"created_at"`
}

func (DeviceData) TableName() string {
	return "devices_final"
}

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
	internalDevice := h.extractDeviceFields(message)

	err := h.orm.WithContext(ctx).Create(&internalDevice).Error()
	if err != nil {
		return fmt.Errorf("creating device: %w", err)
	}

	return nil
}

// GetByID retrieves a device by its ID
func (h *DeviceHandler) GetByID(ctx context.Context, id string) (pubsub.Message, error) {
	var internalDevice DeviceData

	err := h.orm.WithContext(ctx).First(&internalDevice, "id = ?", id).Error()
	if err != nil {
		return nil, fmt.Errorf("getting device: %w", err)
	}

	device := h.toDomainDevice(internalDevice)
	return device, nil
}

// Update handles updating an existing device record
func (h *DeviceHandler) Update(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	internalDevice := h.extractDeviceFields(message)

	err := h.orm.WithContext(ctx).Save(&internalDevice).Error()
	if err != nil {
		return fmt.Errorf("updating device: %w", err)
	}

	return nil
}

func (h *DeviceHandler) extractDeviceFields(message pubsub.Message) DeviceData {
	avroDevice, ok := message.(*avro.AvroDevice)
	if !ok {
		slog.Error("message is not *avro.AvroDevice", "message", message)
		return DeviceData{}
	}

	result := DeviceData{
		ID:                    avroDevice.ID,
		Version:               int(avroDevice.Version),
		Name:                  avroDevice.Name,
		DisplayName:           avroDevice.DisplayName,
		AppEUI:                avroDevice.AppEUI,
		DevEUI:                avroDevice.DevEUI,
		AppKey:                avroDevice.AppKey,
		TenantID:              avroDevice.TenantID,
		LastMessageReceivedAt: avroDevice.LastMessageReceivedAt,
		CreatedAt:             avroDevice.CreatedAt,
	}

	if avroDevice.TenantID != nil {
		result.TenantID = avroDevice.TenantID
	}

	if avroDevice.LastMessageReceivedAt != nil {
		result.LastMessageReceivedAt = avroDevice.LastMessageReceivedAt
	}

	return result
}

func (h *DeviceHandler) toDomainDevice(internalDevice DeviceData) map[string]any {
	result := map[string]any{
		"id":           internalDevice.ID,
		"name":         internalDevice.Name,
		"display_name": internalDevice.DisplayName,
		"app_eui":      internalDevice.AppEUI,
		"dev_eui":      internalDevice.DevEUI,
		"app_key":      internalDevice.AppKey,
	}

	if internalDevice.TenantID != nil {
		result["tenant_id"] = *internalDevice.TenantID
	}

	if internalDevice.LastMessageReceivedAt != nil {
		result["last_message_received_at"] = internalDevice.LastMessageReceivedAt
	}

	return result
}

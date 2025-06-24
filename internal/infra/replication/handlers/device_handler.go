package handlers

import (
	"context"
	"fmt"
	"reflect"
	"time"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
)

// DeviceData represents the device table structure for GORM operations
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
	// UpdatedAt             time.Time  `json:"updated_at"`
}

func (DeviceData) TableName() string {
	return "devices_final"
}

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

// extractDeviceFields uses reflection to extract device fields from any message type
func (h *DeviceHandler) extractDeviceFields(message pubsub.Message) DeviceData {
	val := reflect.ValueOf(message)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	result := DeviceData{
		Version: 1,
	}

	// Extract ID field
	if idField := val.FieldByName("ID"); idField.IsValid() {
		if idField.Type().String() == "domain.ID" {
			// Handle domain.ID type by calling String() method
			if stringMethod := idField.MethodByName("String"); stringMethod.IsValid() {
				results := stringMethod.Call(nil)
				if len(results) > 0 {
					result.ID = results[0].String()
				}
			}
		} else {
			result.ID = idField.Interface().(string)
		}
	}

	// Extract Name field
	if nameField := val.FieldByName("Name"); nameField.IsValid() {
		result.Name = nameField.Interface().(string)
	}

	// Extract DisplayName field
	if displayNameField := val.FieldByName("DisplayName"); displayNameField.IsValid() {
		result.DisplayName = displayNameField.Interface().(string)
	}

	// Extract AppEUI field
	if appEUIField := val.FieldByName("AppEUI"); appEUIField.IsValid() {
		result.AppEUI = appEUIField.Interface().(string)
	}

	// Extract DevEUI field
	if devEUIField := val.FieldByName("DevEUI"); devEUIField.IsValid() {
		result.DevEUI = devEUIField.Interface().(string)
	}

	// Extract AppKey field
	if appKeyField := val.FieldByName("AppKey"); appKeyField.IsValid() {
		result.AppKey = appKeyField.Interface().(string)
	}

	// Extract TenantID field
	if tenantIDField := val.FieldByName("TenantID"); tenantIDField.IsValid() {
		if tenantIDField.IsNil() {
			result.TenantID = nil
		} else {
			tenantID := tenantIDField.Interface().(*string)
			result.TenantID = tenantID
		}
	}

	// Extract LastMessageReceivedAt field
	if lastMessageField := val.FieldByName("LastMessageReceivedAt"); lastMessageField.IsValid() {
		if lastMessageField.IsZero() {
			result.LastMessageReceivedAt = nil
		} else {
			lastMessage := lastMessageField.Interface().(time.Time)
			result.LastMessageReceivedAt = &lastMessage
		}
	}

	// Extract CreatedAt field
	if createdAtField := val.FieldByName("CreatedAt"); createdAtField.IsValid() {
		result.CreatedAt = createdAtField.Interface().(time.Time)
	}

	// Extract UpdatedAt field
	// if updatedAtField := val.FieldByName("UpdatedAt"); updatedAtField.IsValid() {
	// 	result.UpdatedAt = updatedAtField.Interface().(time.Time)
	// }

	return result
}

// toDomainDevice converts internal device to domain representation
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

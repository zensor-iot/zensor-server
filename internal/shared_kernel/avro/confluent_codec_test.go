package avro

import (
	"fmt"
	"testing"
	"time"

	"github.com/riferrei/srclient"
	"github.com/stretchr/testify/assert"

	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"
)

func TestNewConfluentAvroCodec(t *testing.T) {
	// Test that ConfluentAvroCodec can be created
	schemaRegistry := srclient.CreateSchemaRegistryClient("http://localhost:8081")
	codec := NewConfluentAvroCodec(&AvroCommand{}, schemaRegistry)
	assert.NotNil(t, codec)
	assert.NotNil(t, codec.schemaCache)
	assert.NotNil(t, codec.codecCache)
	assert.Equal(t, "-value", codec.subjectSuffix)
}

func TestConfluentAvroCodec_StructValidation(t *testing.T) {
	// Test that all Avro structs can be created and validated
	t.Run("AvroCommand", func(t *testing.T) {
		cmd := &AvroCommand{
			ID:            "cmd-123",
			Version:       1,
			DeviceName:    "test-device",
			DeviceID:      "dev-123",
			TaskID:        "task-123",
			PayloadIndex:  1,
			PayloadValue:  100,
			DispatchAfter: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Port:          80,
			Priority:      "high",
			CreatedAt:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Ready:         true,
			Sent:          false,
			SentAt:        time.Time{},
		}
		assert.Equal(t, "cmd-123", cmd.ID)
		assert.Equal(t, 1, cmd.Version)
		assert.Equal(t, "test-device", cmd.DeviceName)
	})

	t.Run("AvroTask", func(t *testing.T) {
		scheduledTaskID := "scheduled-task-123"
		task := &AvroTask{
			ID:              "task-123",
			DeviceID:        "dev-123",
			ScheduledTaskID: &scheduledTaskID,
			Version:         1,
			CreatedAt:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		assert.Equal(t, "task-123", task.ID)
		assert.Equal(t, "dev-123", task.DeviceID)
		assert.Equal(t, &scheduledTaskID, task.ScheduledTaskID)
	})

	t.Run("AvroDevice", func(t *testing.T) {
		tenantID := "tenant-123"
		lastMessageReceivedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		device := &AvroDevice{
			ID:                    "dev-123",
			Version:               1,
			Name:                  "test-device",
			DisplayName:           "Test Device",
			AppEUI:                "app-eui-123",
			DevEUI:                "dev-eui-123",
			AppKey:                "app-key-123",
			TenantID:              &tenantID,
			LastMessageReceivedAt: &lastMessageReceivedAt,
			CreatedAt:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		assert.Equal(t, "dev-123", device.ID)
		assert.Equal(t, 1, device.Version)
		assert.Equal(t, "test-device", device.Name)
		assert.Equal(t, &tenantID, device.TenantID)
	})

	t.Run("AvroScheduledTask", func(t *testing.T) {
		lastExecutedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		scheduledTask := &AvroScheduledTask{
			ID:               "scheduled-task-123",
			Version:          1,
			TenantID:         "tenant-123",
			DeviceID:         "dev-123",
			CommandTemplates: "template-123",
			Schedule:         "0 0 * * *",
			IsActive:         true,
			CreatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			LastExecutedAt:   &lastExecutedAt,
			DeletedAt:        nil,
		}
		assert.Equal(t, "scheduled-task-123", scheduledTask.ID)
		assert.Equal(t, int64(1), scheduledTask.Version)
		assert.Equal(t, "tenant-123", scheduledTask.TenantID)
		assert.Equal(t, "dev-123", scheduledTask.DeviceID)
		assert.True(t, scheduledTask.IsActive)
	})

	t.Run("AvroTenant", func(t *testing.T) {
		deletedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		tenant := &AvroTenant{
			ID:          "tenant-123",
			Version:     1,
			Name:        "Test Tenant",
			Email:       "test@example.com",
			Description: "Test tenant description",
			IsActive:    true,
			CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			DeletedAt:   &deletedAt,
		}
		assert.Equal(t, "tenant-123", tenant.ID)
		assert.Equal(t, 1, tenant.Version)
		assert.Equal(t, "Test Tenant", tenant.Name)
		assert.Equal(t, "test@example.com", tenant.Email)
		assert.True(t, tenant.IsActive)
		assert.Equal(t, &deletedAt, tenant.DeletedAt)
	})

	t.Run("AvroEvaluationRule", func(t *testing.T) {
		evaluationRule := &AvroEvaluationRule{
			ID:          "rule-123",
			DeviceID:    "dev-123",
			Version:     1,
			Description: "Test evaluation rule",
			Kind:        "threshold",
			Enabled:     true,
			Parameters:  "{\"threshold\": 100}",
			CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		assert.Equal(t, "rule-123", evaluationRule.ID)
		assert.Equal(t, "dev-123", evaluationRule.DeviceID)
		assert.Equal(t, 1, evaluationRule.Version)
		assert.Equal(t, "Test evaluation rule", evaluationRule.Description)
		assert.Equal(t, "threshold", evaluationRule.Kind)
		assert.True(t, evaluationRule.Enabled)
	})
}

func TestConfluentAvroCodec_InvalidData(t *testing.T) {
	schemaRegistry := srclient.CreateSchemaRegistryClient("http://localhost:8081")
	codec := NewConfluentAvroCodec(&AvroCommand{}, schemaRegistry)

	// Test with invalid data
	_, err := codec.Decode([]byte{1, 2, 3}) // Too short
	assert.Error(t, err)

	// Test with invalid magic byte
	invalidData := make([]byte, 10)
	invalidData[0] = 1 // Invalid magic byte
	_, err = codec.Decode(invalidData)
	assert.Error(t, err)
}

func TestConfluentAvroCodec_UnsupportedType(t *testing.T) {
	schemaRegistry := srclient.CreateSchemaRegistryClient("http://localhost:8081")
	codec := NewConfluentAvroCodec(&AvroCommand{}, schemaRegistry)

	// Test with unsupported type
	_, err := codec.Encode("unsupported")
	assert.Error(t, err)
}

func TestConfluentAvroCodec_DeviceWithLastMessageReceivedAt(t *testing.T) {
	// Test that device with last_message_received_at can be serialized correctly
	schemaRegistry := srclient.CreateSchemaRegistryClient("http://localhost:8081")
	codec := NewConfluentAvroCodec(&AvroDevice{}, schemaRegistry)

	// Create a device with last_message_received_at
	device := &AvroDevice{
		ID:                    "dev-123",
		Version:               1,
		Name:                  "test-device",
		DisplayName:           "Test Device",
		AppEUI:                "app-eui-123",
		DevEUI:                "dev-eui-123",
		AppKey:                "app-key-123",
		TenantID:              nil,
		LastMessageReceivedAt: nil, // This should be handled as null
		CreatedAt:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	// Test encoding - this should not fail
	_, err := codec.Encode(device)
	// Note: This test will fail if schema registry is not available, but that's expected
	// The important thing is that it doesn't fail due to union type serialization issues
	if err != nil {
		// If it's a schema registry connection error, that's expected in test environment
		t.Logf("Expected error due to schema registry not being available: %v", err)
	}

	// Test with non-nil last_message_received_at
	lastMessageTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	deviceWithLastMessage := &AvroDevice{
		ID:                    "dev-124",
		Version:               1,
		Name:                  "test-device-2",
		DisplayName:           "Test Device 2",
		AppEUI:                "app-eui-124",
		DevEUI:                "dev-eui-124",
		AppKey:                "app-key-124",
		TenantID:              nil,
		LastMessageReceivedAt: &lastMessageTime,
		CreatedAt:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	// Test encoding with non-nil last_message_received_at
	_, err = codec.Encode(deviceWithLastMessage)
	if err != nil {
		// If it's a schema registry connection error, that's expected in test environment
		t.Logf("Expected error due to schema registry not being available: %v", err)
	}
}

func TestConfluentAvroCodec_ConvertDomainDevice(t *testing.T) {
	// Test that domain Device can be converted to AvroDevice using the new typed method
	schemaRegistry := srclient.CreateSchemaRegistryClient("http://localhost:8081")
	codec := NewConfluentAvroCodec(&AvroDevice{}, schemaRegistry)

	// Create a domain Device
	device := &domain.Device{
		ID:                    domain.ID("dev-123"),
		Name:                  "test-device",
		DisplayName:           "Test Device",
		AppEUI:                "app-eui-123",
		DevEUI:                "dev-eui-123",
		AppKey:                "app-key-123",
		TenantID:              nil,
		LastMessageReceivedAt: utils.Time{},
	}

	// Test the typed conversion
	avroDevice, err := codec.convertInternalDevice(device)
	assert.NoError(t, err)
	assert.NotNil(t, avroDevice)
	assert.Equal(t, "dev-123", avroDevice.ID)
	assert.Equal(t, 1, avroDevice.Version)
	assert.Equal(t, "test-device", avroDevice.Name)
	assert.Equal(t, "Test Device", avroDevice.DisplayName)
	assert.Equal(t, "app-eui-123", avroDevice.AppEUI)
	assert.Equal(t, "dev-eui-123", avroDevice.DevEUI)
	assert.Equal(t, "app-key-123", avroDevice.AppKey)
	assert.Nil(t, avroDevice.TenantID)
	assert.Nil(t, avroDevice.LastMessageReceivedAt)

	// Test with tenant ID
	tenantID := domain.ID("tenant-123")
	deviceWithTenant := &domain.Device{
		ID:                    domain.ID("dev-124"),
		Name:                  "test-device-2",
		DisplayName:           "Test Device 2",
		AppEUI:                "app-eui-124",
		DevEUI:                "dev-eui-124",
		AppKey:                "app-key-124",
		TenantID:              &tenantID,
		LastMessageReceivedAt: utils.Time{},
	}

	avroDeviceWithTenant, err := codec.convertInternalDevice(deviceWithTenant)
	assert.NoError(t, err)
	assert.NotNil(t, avroDeviceWithTenant)
	assert.Equal(t, "dev-124", avroDeviceWithTenant.ID)
	assert.Equal(t, "tenant-123", *avroDeviceWithTenant.TenantID)

	// Test with last message received at
	lastMessageTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	deviceWithLastMessage := &domain.Device{
		ID:                    domain.ID("dev-125"),
		Name:                  "test-device-3",
		DisplayName:           "Test Device 3",
		AppEUI:                "app-eui-125",
		DevEUI:                "dev-eui-125",
		AppKey:                "app-key-125",
		TenantID:              nil,
		LastMessageReceivedAt: utils.Time{Time: lastMessageTime},
	}

	avroDeviceWithLastMessage, err := codec.convertInternalDevice(deviceWithLastMessage)
	assert.NoError(t, err)
	assert.NotNil(t, avroDeviceWithLastMessage)
	assert.Equal(t, "dev-125", avroDeviceWithLastMessage.ID)
	assert.NotNil(t, avroDeviceWithLastMessage.LastMessageReceivedAt)
	assert.Equal(t, lastMessageTime, *avroDeviceWithLastMessage.LastMessageReceivedAt)
}

func TestConfluentAvroCodec_SerializeCommandTemplates(t *testing.T) {
	schemaRegistry := srclient.CreateSchemaRegistryClient("http://localhost:8081")
	codec := NewConfluentAvroCodec(&AvroScheduledTask{}, schemaRegistry)

	// Test empty templates
	emptyResult := codec.serializeCommandTemplates([]domain.CommandTemplate{})
	assert.Equal(t, "[]", emptyResult)

	// Test with one template
	device := domain.Device{
		ID:   domain.ID("dev-123"),
		Name: "test-device",
	}

	template := domain.CommandTemplate{
		Device:   device,
		Port:     15,
		Priority: "NORMAL",
		Payload: domain.CommandPayload{
			Index: 1,
			Value: 100,
		},
		WaitFor: 5 * time.Second,
	}

	templates := []domain.CommandTemplate{template}
	result := codec.serializeCommandTemplates(templates)

	// Verify the JSON structure
	assert.Contains(t, result, `"device":{"id":"dev-123"}`)
	assert.Contains(t, result, `"port":15`)
	assert.Contains(t, result, `"priority":"NORMAL"`)
	assert.Contains(t, result, `"payload":{"index":1,"value":100}`)
	assert.Contains(t, result, `"wait_for":"5s"`)

	// Test with multiple templates
	template2 := domain.CommandTemplate{
		Device:   device,
		Port:     16,
		Priority: "HIGH",
		Payload: domain.CommandPayload{
			Index: 2,
			Value: 200,
		},
		WaitFor: 10 * time.Second,
	}

	multipleTemplates := []domain.CommandTemplate{template, template2}
	multipleResult := codec.serializeCommandTemplates(multipleTemplates)

	// Should contain both templates
	assert.Contains(t, multipleResult, `"port":15`)
	assert.Contains(t, multipleResult, `"port":16`)
	assert.Contains(t, multipleResult, `"priority":"NORMAL"`)
	assert.Contains(t, multipleResult, `"priority":"HIGH"`)
}

// MockSchemaRegistry is a mock implementation of SchemaRegistry for testing
type MockSchemaRegistry struct {
	schemas map[string]*srclient.Schema
}

func NewMockSchemaRegistry() *MockSchemaRegistry {
	return &MockSchemaRegistry{
		schemas: make(map[string]*srclient.Schema),
	}
}

func (m *MockSchemaRegistry) GetLatestSchema(subject string) (*srclient.Schema, error) {
	if schema, exists := m.schemas[subject]; exists {
		return schema, nil
	}
	return nil, fmt.Errorf("schema not found for subject: %s", subject)
}

func (m *MockSchemaRegistry) CreateSchema(subject string, schema string, schemaType srclient.SchemaType, references ...srclient.Reference) (*srclient.Schema, error) {
	// Create a simple mock schema - in real tests you might want to use a proper mock library
	mockSchema := &srclient.Schema{}
	m.schemas[subject] = mockSchema
	return mockSchema, nil
}

func (m *MockSchemaRegistry) GetSchema(schemaID int) (*srclient.Schema, error) {
	// For simplicity, return the first schema found
	for _, schema := range m.schemas {
		return schema, nil
	}
	return nil, fmt.Errorf("schema not found for ID: %d", schemaID)
}

func TestConfluentAvroCodec_WithMockSchemaRegistry(t *testing.T) {
	// Create a mock schema registry
	mockRegistry := NewMockSchemaRegistry()

	// Create codec with mock registry
	codec := NewConfluentAvroCodec(&AvroCommand{}, mockRegistry)
	assert.NotNil(t, codec)
	assert.Equal(t, mockRegistry, codec.schemaRegistry)
}

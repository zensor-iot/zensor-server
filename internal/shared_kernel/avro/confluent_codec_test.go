package avro

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewConfluentAvroCodec(t *testing.T) {
	// Test that ConfluentAvroCodec can be created
	codec, err := NewConfluentAvroCodec(&AvroCommand{}, "http://localhost:8081")
	assert.NoError(t, err)
	assert.NotNil(t, codec)
	assert.NotNil(t, codec.schemas)
	assert.NotNil(t, codec.codecs)
	assert.NotNil(t, codec.subjectToID)
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
			Payload:       AvroCommandPayload{Index: 1, Value: 100},
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
	codec, err := NewConfluentAvroCodec(&AvroCommand{}, "http://localhost:8081")
	assert.NoError(t, err)

	// Test with invalid data
	_, err = codec.Decode([]byte{1, 2, 3}) // Too short
	assert.Error(t, err)

	// Test with invalid magic byte
	invalidData := make([]byte, 10)
	invalidData[0] = 1 // Invalid magic byte
	_, err = codec.Decode(invalidData)
	assert.Error(t, err)
}

func TestConfluentAvroCodec_UnsupportedType(t *testing.T) {
	codec, err := NewConfluentAvroCodec(&AvroCommand{}, "http://localhost:8081")
	assert.NoError(t, err)

	// Test with unsupported type
	_, err = codec.Encode("unsupported")
	assert.Error(t, err)
}

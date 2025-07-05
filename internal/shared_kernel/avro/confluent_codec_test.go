package avro

import (
	"testing"
	"time"

	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfluentAvroCodec(t *testing.T) {
	// Test with a valid prototype
	codec, err := NewConfluentAvroCodec(&shared_kernel.Command{}, "http://localhost:8081")
	require.NoError(t, err)
	assert.NotNil(t, codec)
	assert.NotEmpty(t, codec.schemas)
	assert.NotEmpty(t, codec.codecs)
	assert.NotEmpty(t, codec.subjectToID)
}

func TestConfluentAvroCodec_EncodeDecode_Command(t *testing.T) {
	codec, err := NewConfluentAvroCodec(&shared_kernel.Command{}, "http://localhost:8081")
	require.NoError(t, err)

	original := &shared_kernel.Command{
		ID:            "cmd-123",
		Version:       1,
		DeviceName:    "test-device",
		DeviceID:      "dev-123",
		TaskID:        "task-123",
		Payload:       shared_kernel.CommandPayload{Index: 1, Value: 100},
		DispatchAfter: utils.Time{Time: time.Now()},
		Port:          80,
		Priority:      "high",
		CreatedAt:     utils.Time{Time: time.Now()},
		Ready:         true,
		Sent:          false,
		SentAt:        utils.Time{},
	}

	// Encode
	encoded, err := codec.Encode(original)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)

	// Decode
	decoded, err := codec.Decode(encoded)
	require.NoError(t, err)
	assert.NotNil(t, decoded)

	// Verify
	result, ok := decoded.(*shared_kernel.Command)
	assert.True(t, ok)
	assert.Equal(t, original.ID, result.ID)
	assert.Equal(t, original.Version, result.Version)
	assert.Equal(t, original.DeviceName, result.DeviceName)
	assert.Equal(t, original.DeviceID, result.DeviceID)
	assert.Equal(t, original.TaskID, result.TaskID)
	assert.Equal(t, original.Payload, result.Payload)
	assert.Equal(t, original.DispatchAfter, result.DispatchAfter)
	assert.Equal(t, original.Port, result.Port)
	assert.Equal(t, original.Priority, result.Priority)
	assert.Equal(t, original.CreatedAt, result.CreatedAt)
	assert.Equal(t, original.Ready, result.Ready)
	assert.Equal(t, original.Sent, result.Sent)
	assert.Equal(t, original.SentAt, result.SentAt)
}

func TestConfluentAvroCodec_EncodeDecode_Task(t *testing.T) {
	codec, err := NewConfluentAvroCodec(&AvroTask{}, "http://localhost:8081")
	require.NoError(t, err)

	scheduledTaskID := "scheduled-task-123"
	original := &AvroTask{
		ID:              "task-123",
		DeviceID:        "dev-123",
		ScheduledTaskID: &scheduledTaskID,
		Version:         1,
		CreatedAt:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	// Encode
	encoded, err := codec.Encode(original)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)

	// Decode
	decoded, err := codec.Decode(encoded)
	require.NoError(t, err)
	assert.NotNil(t, decoded)

	// Verify
	result, ok := decoded.(*AvroTask)
	assert.True(t, ok)
	assert.Equal(t, original.ID, result.ID)
	assert.Equal(t, original.DeviceID, result.DeviceID)
	assert.Equal(t, original.ScheduledTaskID, result.ScheduledTaskID)
	assert.Equal(t, original.Version, result.Version)
	assert.Equal(t, original.CreatedAt, result.CreatedAt)
	assert.Equal(t, original.UpdatedAt, result.UpdatedAt)
}

func TestConfluentAvroCodec_EncodeDecode_Device(t *testing.T) {
	codec, err := NewConfluentAvroCodec(&AvroDevice{}, "http://localhost:8081")
	require.NoError(t, err)

	tenantID := "tenant-123"
	lastMessageReceivedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	original := &AvroDevice{
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

	// Encode
	encoded, err := codec.Encode(original)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)

	// Decode
	decoded, err := codec.Decode(encoded)
	require.NoError(t, err)
	assert.NotNil(t, decoded)

	// Verify
	result, ok := decoded.(*AvroDevice)
	assert.True(t, ok)
	assert.Equal(t, original.ID, result.ID)
	assert.Equal(t, original.Version, result.Version)
	assert.Equal(t, original.Name, result.Name)
	assert.Equal(t, original.DisplayName, result.DisplayName)
	assert.Equal(t, original.AppEUI, result.AppEUI)
	assert.Equal(t, original.DevEUI, result.DevEUI)
	assert.Equal(t, original.AppKey, result.AppKey)
	assert.Equal(t, original.TenantID, result.TenantID)
	assert.Equal(t, original.LastMessageReceivedAt, result.LastMessageReceivedAt)
	assert.Equal(t, original.CreatedAt, result.CreatedAt)
	assert.Equal(t, original.UpdatedAt, result.UpdatedAt)
}

func TestConfluentAvroCodec_EncodeDecode_ScheduledTask(t *testing.T) {
	codec, err := NewConfluentAvroCodec(&AvroScheduledTask{}, "http://localhost:8081")
	require.NoError(t, err)

	lastExecutedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	original := &AvroScheduledTask{
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

	// Encode
	encoded, err := codec.Encode(original)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)

	// Decode
	decoded, err := codec.Decode(encoded)
	require.NoError(t, err)
	assert.NotNil(t, decoded)

	// Verify
	result, ok := decoded.(*AvroScheduledTask)
	assert.True(t, ok)
	assert.Equal(t, original.ID, result.ID)
	assert.Equal(t, original.Version, result.Version)
	assert.Equal(t, original.TenantID, result.TenantID)
	assert.Equal(t, original.DeviceID, result.DeviceID)
	assert.Equal(t, original.CommandTemplates, result.CommandTemplates)
	assert.Equal(t, original.Schedule, result.Schedule)
	assert.Equal(t, original.IsActive, result.IsActive)
	assert.Equal(t, original.CreatedAt, result.CreatedAt)
	assert.Equal(t, original.UpdatedAt, result.UpdatedAt)
	assert.Equal(t, original.LastExecutedAt, result.LastExecutedAt)
	assert.Equal(t, original.DeletedAt, result.DeletedAt)
}

func TestConfluentAvroCodec_EncodeDecode_Tenant(t *testing.T) {
	codec, err := NewConfluentAvroCodec(&AvroTenant{}, "http://localhost:8081")
	require.NoError(t, err)

	deletedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	original := &AvroTenant{
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

	// Encode
	encoded, err := codec.Encode(original)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)

	// Decode
	decoded, err := codec.Decode(encoded)
	require.NoError(t, err)
	assert.NotNil(t, decoded)

	// Verify
	result, ok := decoded.(*AvroTenant)
	assert.True(t, ok)
	assert.Equal(t, original.ID, result.ID)
	assert.Equal(t, original.Version, result.Version)
	assert.Equal(t, original.Name, result.Name)
	assert.Equal(t, original.Email, result.Email)
	assert.Equal(t, original.Description, result.Description)
	assert.Equal(t, original.IsActive, result.IsActive)
	assert.Equal(t, original.CreatedAt, result.CreatedAt)
	assert.Equal(t, original.UpdatedAt, result.UpdatedAt)
	assert.Equal(t, original.DeletedAt, result.DeletedAt)
}

func TestConfluentAvroCodec_EncodeDecode_EvaluationRule(t *testing.T) {
	codec, err := NewConfluentAvroCodec(&AvroEvaluationRule{}, "http://localhost:8081")
	require.NoError(t, err)

	original := &AvroEvaluationRule{
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

	// Encode
	encoded, err := codec.Encode(original)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)

	// Decode
	decoded, err := codec.Decode(encoded)
	require.NoError(t, err)
	assert.NotNil(t, decoded)

	// Verify
	result, ok := decoded.(*AvroEvaluationRule)
	assert.True(t, ok)
	assert.Equal(t, original.ID, result.ID)
	assert.Equal(t, original.DeviceID, result.DeviceID)
	assert.Equal(t, original.Version, result.Version)
	assert.Equal(t, original.Description, result.Description)
	assert.Equal(t, original.Kind, result.Kind)
	assert.Equal(t, original.Enabled, result.Enabled)
	assert.Equal(t, original.Parameters, result.Parameters)
	assert.Equal(t, original.CreatedAt, result.CreatedAt)
	assert.Equal(t, original.UpdatedAt, result.UpdatedAt)
}

func TestConfluentAvroCodec_EncodeDecode_AvroStructs(t *testing.T) {
	codec, err := NewConfluentAvroCodec(&AvroCommand{}, "http://localhost:8081")
	require.NoError(t, err)

	original := &AvroCommand{
		ID:            "cmd-123",
		Version:       1,
		DeviceName:    "test-device",
		DeviceID:      "dev-123",
		TaskID:        "task-123",
		Payload:       AvroCommandPayload{Index: 1, Value: 100},
		DispatchAfter: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Port:          8080,
		Priority:      "high",
		CreatedAt:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Ready:         true,
		Sent:          false,
		SentAt:        time.Time{},
	}

	// Encode
	encoded, err := codec.Encode(original)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)

	// Decode
	decoded, err := codec.Decode(encoded)
	require.NoError(t, err)
	assert.NotNil(t, decoded)

	// Verify
	result, ok := decoded.(*AvroCommand)
	assert.True(t, ok)
	assert.Equal(t, original.ID, result.ID)
	assert.Equal(t, original.Version, result.Version)
	assert.Equal(t, original.DeviceName, result.DeviceName)
	assert.Equal(t, original.DeviceID, result.DeviceID)
	assert.Equal(t, original.TaskID, result.TaskID)
	assert.Equal(t, original.Payload, result.Payload)
	assert.Equal(t, original.DispatchAfter, result.DispatchAfter)
	assert.Equal(t, original.Port, result.Port)
	assert.Equal(t, original.Priority, result.Priority)
	assert.Equal(t, original.CreatedAt, result.CreatedAt)
	assert.Equal(t, original.Ready, result.Ready)
	assert.Equal(t, original.Sent, result.Sent)
	assert.Equal(t, original.SentAt, result.SentAt)
}

func TestConfluentAvroCodec_InvalidData(t *testing.T) {
	codec, err := NewConfluentAvroCodec(&shared_kernel.Command{}, "http://localhost:8081")
	require.NoError(t, err)

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
	codec, err := NewConfluentAvroCodec(&shared_kernel.Command{}, "http://localhost:8081")
	require.NoError(t, err)

	// Test with unsupported type
	_, err = codec.Encode("unsupported")
	assert.Error(t, err)
}

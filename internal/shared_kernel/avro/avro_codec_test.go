package avro

import (
	"testing"
	"time"
)

func TestAvroCodec_Command(t *testing.T) {
	// Create a test command using the actual AvroCommand struct
	testCommand := AvroCommand{
		ID:         "test-command-123",
		Version:    1,
		DeviceName: "test-device",
		DeviceID:   "device-123",
		TaskID:     "task-123",
		Payload: AvroCommandPayload{
			Index: 1,
			Value: 100,
		},
		DispatchAfter: time.Now(),
		Port:          15,
		Priority:      "NORMAL",
		CreatedAt:     time.Now(),
		Ready:         true,
		Sent:          false,
		SentAt:        time.Time{},
	}

	// Create Avro codec with schema registry URL
	codec := NewAvroCodec(testCommand, "http://localhost:8081")

	// Encode the command
	encoded, err := codec.Encode(testCommand)
	if err != nil {
		t.Fatalf("Failed to encode command: %v", err)
	}

	// Decode the command
	decoded, err := codec.Decode(encoded)
	if err != nil {
		t.Fatalf("Failed to decode command: %v", err)
	}

	// Verify the decoded data is not nil
	if decoded == nil {
		t.Fatal("Decoded command is nil")
	}

	t.Logf("Successfully encoded and decoded command: %+v", decoded)
}

func TestAvroCodec_Task(t *testing.T) {
	// Create a test task using the actual AvroTask struct
	scheduledTaskID := "scheduled-task-123"
	testTask := AvroTask{
		ID:              "test-task-123",
		DeviceID:        "device-123",
		ScheduledTaskID: &scheduledTaskID,
		Version:         1,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Create Avro codec with schema registry URL
	codec := NewAvroCodec(testTask, "http://localhost:8081")

	// Encode the task
	encoded, err := codec.Encode(testTask)
	if err != nil {
		t.Fatalf("Failed to encode task: %v", err)
	}

	// Decode the task
	decoded, err := codec.Decode(encoded)
	if err != nil {
		t.Fatalf("Failed to decode task: %v", err)
	}

	// Verify the decoded data is not nil
	if decoded == nil {
		t.Fatal("Decoded task is nil")
	}

	t.Logf("Successfully encoded and decoded task: %+v", decoded)
}

func TestAvroCodec_Device(t *testing.T) {
	// Create a test device using the actual AvroDevice struct
	testDevice := AvroDevice{
		ID:                    "test-device-123",
		Version:               1,
		Name:                  "test-device",
		DisplayName:           "Test Device",
		AppEUI:                "app-eui-123",
		DevEUI:                "dev-eui-123",
		AppKey:                "app-key-123",
		TenantID:              nil,
		LastMessageReceivedAt: nil,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	// Create Avro codec with schema registry URL
	codec := NewAvroCodec(testDevice, "http://localhost:8081")

	// Encode the device
	encoded, err := codec.Encode(testDevice)
	if err != nil {
		t.Fatalf("Failed to encode device: %v", err)
	}

	// Decode the device
	decoded, err := codec.Decode(encoded)
	if err != nil {
		t.Fatalf("Failed to decode device: %v", err)
	}

	// Verify the decoded data is not nil
	if decoded == nil {
		t.Fatal("Decoded device is nil")
	}

	t.Logf("Successfully encoded and decoded device: %+v", decoded)
}

func TestAvroCodec_ScheduledTask(t *testing.T) {
	// Create a test scheduled task using the actual AvroScheduledTask struct
	testScheduledTask := AvroScheduledTask{
		ID:               "test-scheduled-task-123",
		Version:          1,
		TenantID:         "tenant-123",
		DeviceID:         "device-123",
		CommandTemplates: `[{"port":15,"priority":"NORMAL","payload":{"index":1,"value":100},"wait_for":"5s"}]`,
		Schedule:         "* * * * *",
		IsActive:         true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		LastExecutedAt:   nil,
		DeletedAt:        nil,
	}

	// Create Avro codec with schema registry URL
	codec := NewAvroCodec(testScheduledTask, "http://localhost:8081")

	// Encode the scheduled task
	encoded, err := codec.Encode(testScheduledTask)
	if err != nil {
		t.Fatalf("Failed to encode scheduled task: %v", err)
	}

	// Decode the scheduled task
	decoded, err := codec.Decode(encoded)
	if err != nil {
		t.Fatalf("Failed to decode scheduled task: %v", err)
	}

	// Verify the decoded data is not nil
	if decoded == nil {
		t.Fatal("Decoded scheduled task is nil")
	}

	t.Logf("Successfully encoded and decoded scheduled task: %+v", decoded)
}

func TestAvroCodec_Tenant(t *testing.T) {
	// Create a test tenant using the actual AvroTenant struct
	testTenant := AvroTenant{
		ID:          "test-tenant-123",
		Version:     1,
		Name:        "Test Tenant",
		Email:       "test@example.com",
		Description: "Test tenant description",
		IsActive:    true,
		CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		DeletedAt:   nil,
	}

	// Create Avro codec with schema registry URL
	codec := NewAvroCodec(testTenant, "http://localhost:8081")

	// Encode the tenant
	encoded, err := codec.Encode(testTenant)
	if err != nil {
		t.Fatalf("Failed to encode tenant: %v", err)
	}

	// Decode the tenant
	decoded, err := codec.Decode(encoded)
	if err != nil {
		t.Fatalf("Failed to decode tenant: %v", err)
	}

	// Verify the decoded data is not nil
	if decoded == nil {
		t.Fatal("Decoded tenant is nil")
	}

	t.Logf("Successfully encoded and decoded tenant: %+v", decoded)
}

func TestAvroCodec_EvaluationRule(t *testing.T) {
	// Create a test evaluation rule using the actual AvroEvaluationRule struct
	testEvaluationRule := AvroEvaluationRule{
		ID:          "test-evaluation-rule-123",
		DeviceID:    "device-123",
		Version:     1,
		Description: "Test evaluation rule",
		Kind:        "threshold",
		Enabled:     true,
		Parameters:  `{"metric":"temperature","lower_threshold":20.0,"upper_threshold":30.0}`,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create Avro codec with schema registry URL
	codec := NewAvroCodec(testEvaluationRule, "http://localhost:8081")

	// Encode the evaluation rule
	encoded, err := codec.Encode(testEvaluationRule)
	if err != nil {
		t.Fatalf("Failed to encode evaluation rule: %v", err)
	}

	// Decode the evaluation rule
	decoded, err := codec.Decode(encoded)
	if err != nil {
		t.Fatalf("Failed to decode evaluation rule: %v", err)
	}

	// Verify the decoded data is not nil
	if decoded == nil {
		t.Fatal("Decoded evaluation rule is nil")
	}

	t.Logf("Successfully encoded and decoded evaluation rule: %+v", decoded)
}

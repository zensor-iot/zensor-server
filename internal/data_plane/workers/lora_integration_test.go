package workers

import (
	"testing"
	"time"

	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/avro"
	"zensor-server/internal/shared_kernel/device"
)

func TestLoraIntegrationWorker_ConvertToSharedCommand(t *testing.T) {
	worker := &LoraIntegrationWorker{}

	// Test with AvroCommand
	avroCmd := &avro.AvroCommand{
		ID:            "test-command-id",
		Version:       2,
		DeviceID:      "test-device-id",
		DeviceName:    "test-device",
		TaskID:        "test-task-id",
		PayloadIndex:  5,
		PayloadValue:  123,
		DispatchAfter: time.Now(),
		Port:          16,
		Priority:      "HIGH",
		CreatedAt:     time.Now(),
		Ready:         true,
		Sent:          false,
		SentAt:        time.Time{},
	}

	command, err := worker.convertToSharedCommand(avroCmd)
	if err != nil {
		t.Fatalf("Failed to convert AvroCommand: %v", err)
	}

	// Verify that all fields are preserved correctly
	if command.ID != "test-command-id" {
		t.Errorf("Expected ID 'test-command-id', got %s", command.ID)
	}
	if command.DeviceID != "test-device-id" {
		t.Errorf("Expected DeviceID 'test-device-id', got %s", command.DeviceID)
	}
	if command.DeviceName != "test-device" {
		t.Errorf("Expected DeviceName 'test-device', got %s", command.DeviceName)
	}
	if command.TaskID != "test-task-id" {
		t.Errorf("Expected TaskID 'test-task-id', got %s", command.TaskID)
	}
	if command.Payload.Index != 5 {
		t.Errorf("Expected Payload.Index 5, got %d", command.Payload.Index)
	}
	if command.Payload.Value != 123 {
		t.Errorf("Expected Payload.Value 123, got %d", command.Payload.Value)
	}
	if command.Port != 16 {
		t.Errorf("Expected Port 16, got %d", command.Port)
	}
	if command.Priority != "HIGH" {
		t.Errorf("Expected Priority 'HIGH', got %s", command.Priority)
	}
	if command.Ready != true {
		t.Errorf("Expected Ready true, got %t", command.Ready)
	}
	if command.Sent != false {
		t.Errorf("Expected Sent false, got %t", command.Sent)
	}
}

func TestLoraIntegrationWorker_ConvertToSharedCommand_WithStructMessage(t *testing.T) {
	worker := &LoraIntegrationWorker{}

	// Test with device.Command (fallback case)
	structMessage := device.Command{
		ID:         "test-456",
		Version:    2,
		DeviceID:   "device-456",
		DeviceName: "test-device-2",
		TaskID:     "task-456",
		Payload: device.CommandPayload{
			Index: 2,
			Value: 200,
		},
		DispatchAfter: utils.Time{Time: time.Now()},
		Port:          16,
		Priority:      "HIGH",
		CreatedAt:     utils.Time{Time: time.Now()},
		Ready:         false,
		Sent:          true,
		SentAt:        utils.Time{Time: time.Now()},
	}

	command, err := worker.convertToSharedCommand(structMessage)
	if err != nil {
		t.Fatalf("Failed to convert struct message: %v", err)
	}

	if command.ID != "test-456" {
		t.Errorf("Expected ID 'test-456', got %s", command.ID)
	}
	if command.DeviceName != "test-device-2" {
		t.Errorf("Expected DeviceName 'test-device-2', got %s", command.DeviceName)
	}
	if command.Payload.Index != 2 {
		t.Errorf("Expected Payload.Index 2, got %d", command.Payload.Index)
	}
	if command.Payload.Value != 200 {
		t.Errorf("Expected Payload.Value 200, got %d", command.Payload.Value)
	}
	if command.Port != 16 {
		t.Errorf("Expected Port 16, got %d", command.Port)
	}
}

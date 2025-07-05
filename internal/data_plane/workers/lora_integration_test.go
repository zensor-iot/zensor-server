package workers

import (
	"testing"
	"time"

	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel"
)

func TestLoraIntegrationWorker_ConvertToSharedCommand(t *testing.T) {
	worker := &LoraIntegrationWorker{}

	// Test with a map (old format or schema payload)
	mapMessage := map[string]any{
		"id":          "test-123",
		"version":     1,
		"device_id":   "device-123",
		"device_name": "test-device",
		"task_id":     "task-123",
		"payload": map[string]any{
			"index": 1,
			"value": 100,
		},
		"dispatch_after": "2023-01-01T00:00:00Z",
		"port":           15,
		"priority":       "NORMAL",
		"created_at":     "2023-01-01T00:00:00Z",
		"ready":          true,
		"sent":           false,
		"sent_at":        "2023-01-01T00:00:00Z",
	}

	command, err := worker.convertToSharedCommand(mapMessage)
	if err != nil {
		t.Fatalf("Failed to convert map message: %v", err)
	}

	if command.ID != "test-123" {
		t.Errorf("Expected ID 'test-123', got %s", command.ID)
	}
	if command.DeviceName != "test-device" {
		t.Errorf("Expected DeviceName 'test-device', got %s", command.DeviceName)
	}
	if command.Payload.Index != 1 {
		t.Errorf("Expected Payload.Index 1, got %d", command.Payload.Index)
	}
	if command.Payload.Value != 100 {
		t.Errorf("Expected Payload.Value 100, got %d", command.Payload.Value)
	}

	// Test with a pointer to map
	mapPtrMessage := &mapMessage
	command, err = worker.convertToSharedCommand(mapPtrMessage)
	if err != nil {
		t.Fatalf("Failed to convert map pointer message: %v", err)
	}

	if command.ID != "test-123" {
		t.Errorf("Expected ID 'test-123', got %s", command.ID)
	}

	// Test with a struct (new format)
	structMessage := shared_kernel.Command{
		ID:         "test-456",
		Version:    2,
		DeviceID:   "device-456",
		DeviceName: "test-device-2",
		TaskID:     "task-456",
		Payload: shared_kernel.CommandPayload{
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

	command, err = worker.convertToSharedCommand(structMessage)
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
}

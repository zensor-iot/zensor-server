package communication

import (
	"context"
	"testing"
	"time"

	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/avro"
)

func TestCommandPublisher_Dispatch(t *testing.T) {
	// Create a mock publisher factory
	mockFactory := &mockPublisherFactory{
		publisher: &mockPublisher{},
	}

	// Create the command publisher
	publisher, err := NewCommandPublisher(mockFactory)
	if err != nil {
		t.Fatalf("Failed to create command publisher: %v", err)
	}

	// Create a test domain command
	cmd := domain.Command{
		ID:       domain.ID("test-command-id"),
		Version:  domain.Version(1),
		Device:   domain.Device{ID: domain.ID("test-device-id"), Name: "test-device"},
		Task:     domain.Task{ID: domain.ID("test-task-id")},
		Port:     domain.Port(1),
		Priority: domain.CommandPriority("NORMAL"),
		Payload: domain.CommandPayload{
			Index: domain.Index(0),
			Value: domain.CommandValue(100),
		},
		DispatchAfter: utils.Time{Time: time.Now()},
		Ready:         true,
		Sent:          false,
		SentAt:        utils.Time{},
	}

	// Dispatch the command
	err = publisher.Dispatch(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Failed to dispatch command: %v", err)
	}

	// Verify that the mock publisher was called with the correct data
	mockPub := mockFactory.publisher.(*mockPublisher)
	if mockPub.publishedKey != pubsub.Key(cmd.ID) {
		t.Errorf("Expected key %s, got %s", cmd.ID, mockPub.publishedKey)
	}

	// Verify that the published value is an AvroCommand
	avroCmd, ok := mockPub.publishedValue.(*avro.AvroCommand)
	if !ok {
		t.Fatalf("Expected published value to be *avro.AvroCommand, got %T", mockPub.publishedValue)
	}

	// Verify the AvroCommand fields match the domain command
	if avroCmd.ID != string(cmd.ID) {
		t.Errorf("Expected ID %s, got %s", cmd.ID, avroCmd.ID)
	}
	if avroCmd.Version != int(cmd.Version)+1 { // Version should be incremented
		t.Errorf("Expected version %d, got %d", int(cmd.Version)+1, avroCmd.Version)
	}
	if avroCmd.DeviceID != string(cmd.Device.ID) {
		t.Errorf("Expected device ID %s, got %s", cmd.Device.ID, avroCmd.DeviceID)
	}
	if avroCmd.DeviceName != cmd.Device.Name {
		t.Errorf("Expected device name %s, got %s", cmd.Device.Name, avroCmd.DeviceName)
	}
	if avroCmd.TaskID != string(cmd.Task.ID) {
		t.Errorf("Expected task ID %s, got %s", cmd.Task.ID, avroCmd.TaskID)
	}
	if avroCmd.PayloadIndex != int(cmd.Payload.Index) {
		t.Errorf("Expected payload index %d, got %d", cmd.Payload.Index, avroCmd.PayloadIndex)
	}
	if avroCmd.PayloadValue != int(cmd.Payload.Value) {
		t.Errorf("Expected payload value %d, got %d", cmd.Payload.Value, avroCmd.PayloadValue)
	}
	if avroCmd.Port != int(cmd.Port) {
		t.Errorf("Expected port %d, got %d", cmd.Port, avroCmd.Port)
	}
	if avroCmd.Priority != string(cmd.Priority) {
		t.Errorf("Expected priority %s, got %s", cmd.Priority, avroCmd.Priority)
	}
	if avroCmd.Ready != cmd.Ready {
		t.Errorf("Expected ready %t, got %t", cmd.Ready, avroCmd.Ready)
	}
	if avroCmd.Sent != cmd.Sent {
		t.Errorf("Expected sent %t, got %t", cmd.Sent, avroCmd.Sent)
	}
}

// Mock implementations for testing

type mockPublisherFactory struct {
	publisher pubsub.Publisher
}

func (f *mockPublisherFactory) New(topic pubsub.Topic, prototype pubsub.Message) (pubsub.Publisher, error) {
	return f.publisher, nil
}

type mockPublisher struct {
	publishedKey   pubsub.Key
	publishedValue pubsub.Message
}

func (p *mockPublisher) Publish(ctx context.Context, key pubsub.Key, value pubsub.Message) error {
	p.publishedKey = key
	p.publishedValue = value
	return nil
}

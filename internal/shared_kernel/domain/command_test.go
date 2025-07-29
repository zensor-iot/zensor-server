package domain

import (
	"testing"
	"time"
)

func TestCommandBuilder_SetsCreatedAt(t *testing.T) {
	// Create a command using the builder
	cmd, err := NewCommandBuilder().
		WithDevice(Device{ID: "test-device", Name: "Test Device"}).
		WithPayload(CommandPayload{Index: 1, Value: 100}).
		Build()

	if err != nil {
		t.Fatalf("Failed to build command: %v", err)
	}

	// Verify that CreatedAt is set and is recent
	if cmd.CreatedAt.Time.IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	// Verify that CreatedAt is within the last second
	now := time.Now()
	if cmd.CreatedAt.Time.After(now) {
		t.Error("CreatedAt should not be in the future")
	}

	if now.Sub(cmd.CreatedAt.Time) > time.Second {
		t.Error("CreatedAt should be set to a recent time")
	}

	// Verify other fields are set correctly
	if cmd.ID == "" {
		t.Error("ID should be set")
	}
	if cmd.Version != 1 {
		t.Error("Version should be 1")
	}
	if cmd.Device.ID != "test-device" {
		t.Error("Device ID should be set correctly")
	}
	// Task field is not set by the builder, so we don't check it
	if cmd.Payload.Index != 1 {
		t.Error("Payload Index should be set correctly")
	}
	if cmd.Payload.Value != 100 {
		t.Error("Payload Value should be set correctly")
	}
}

func TestCommand_UpdateStatus_SetsQueuedAt(t *testing.T) {
	// Create a command
	cmd, err := NewCommandBuilder().
		WithDevice(Device{ID: "test-device", Name: "Test Device"}).
		WithPayload(CommandPayload{Index: 1, Value: 100}).
		Build()

	if err != nil {
		t.Fatalf("Failed to build command: %v", err)
	}

	// Initially, QueuedAt should be nil
	if cmd.QueuedAt != nil {
		t.Error("QueuedAt should be nil initially")
	}

	// Update status to queued
	cmd.UpdateStatus(CommandStatusQueued, nil)

	// Verify that QueuedAt is now set
	if cmd.QueuedAt == nil {
		t.Error("QueuedAt should be set after updating status to queued")
	}

	// Verify that QueuedAt is recent
	now := time.Now()
	if cmd.QueuedAt.Time.After(now) {
		t.Error("QueuedAt should not be in the future")
	}

	if now.Sub(cmd.QueuedAt.Time) > time.Second {
		t.Error("QueuedAt should be set to a recent time")
	}

	// Verify that status is updated
	if cmd.Status != CommandStatusQueued {
		t.Errorf("Status should be CommandStatusQueued, got %s", cmd.Status)
	}

	// Verify that version is incremented
	if cmd.Version != 2 {
		t.Errorf("Version should be incremented to 2, got %d", cmd.Version)
	}
}

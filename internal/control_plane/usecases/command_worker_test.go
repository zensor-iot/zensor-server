package usecases

import (
	"context"
	"testing"
	"time"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCommandWorker_HandleCommandStatusUpdate_Queued(t *testing.T) {
	// Create a mock command repository
	mockRepo := &MockCommandRepository{}

	// Create a command worker
	worker := &CommandWorker{
		commandRepository: mockRepo,
	}

	// Create a test command
	testCommand := domain.Command{
		ID:       "test-command-123",
		Version:  1,
		Device:   domain.Device{ID: "device-1", Name: "Test Device"},
		Task:     domain.Task{ID: "task-1"},
		Port:     15,
		Priority: domain.CommandPriority("NORMAL"),
		Payload: domain.CommandPayload{
			Index: 1,
			Value: 100,
		},
		DispatchAfter: utils.Time{Time: time.Now()},
		CreatedAt:     utils.Time{Time: time.Now()},
		Ready:         true,
		Sent:          false,
		Status:        domain.CommandStatusPending,
	}

	// Set up mock expectations
	mockRepo.On("GetByID", "test-command-123").Return(testCommand, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("domain.Command")).Return(nil)

	// Create a status update
	statusUpdate := domain.CommandStatusUpdate{
		CommandID:  "test-command-123",
		DeviceName: "Test Device",
		Status:     domain.CommandStatusQueued,
		Timestamp:  time.Now(),
	}

	// Process the status update
	ctx := context.Background()
	worker.handleCommandStatusUpdateStruct(ctx, statusUpdate)

	// Verify that the repository was called correctly
	mockRepo.AssertExpectations(t)

	// Get the updated command that was passed to Update
	updateCalls := mockRepo.Calls
	require.Greater(t, len(updateCalls), 0, "Update should have been called")

	// Find the Update call
	var updateCall *mock.Call
	for _, call := range updateCalls {
		if call.Method == "Update" {
			updateCall = &call
			break
		}
	}
	require.NotNil(t, updateCall, "Update call should be found")

	// Extract the command from the Update call (second argument)
	updatedCmd := updateCall.Arguments[1].(domain.Command)

	// Verify that the status was updated correctly
	assert.Equal(t, domain.CommandStatusQueued, updatedCmd.Status)
	assert.NotNil(t, updatedCmd.QueuedAt, "QueuedAt should be set")
	assert.Equal(t, int(domain.Version(2)), int(updatedCmd.Version), "Version should be incremented")
}

// MockCommandRepository is a mock implementation of CommandRepository for testing
type MockCommandRepository struct {
	mock.Mock
}

func (m *MockCommandRepository) Create(ctx context.Context, cmd domain.Command) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockCommandRepository) Update(ctx context.Context, cmd domain.Command) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockCommandRepository) GetByID(ctx context.Context, id domain.ID) (domain.Command, error) {
	args := m.Called(id.String())
	if args.Get(0) == nil {
		return domain.Command{}, args.Error(1)
	}
	return args.Get(0).(domain.Command), args.Error(1)
}

func (m *MockCommandRepository) FindAllPending(ctx context.Context) ([]domain.Command, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Command), args.Error(1)
}

func (m *MockCommandRepository) FindPendingByDevice(ctx context.Context, deviceID domain.ID) ([]domain.Command, error) {
	args := m.Called(ctx, deviceID)
	return args.Get(0).([]domain.Command), args.Error(1)
}

func (m *MockCommandRepository) FindByTaskID(ctx context.Context, taskID domain.ID) ([]domain.Command, error) {
	args := m.Called(ctx, taskID)
	return args.Get(0).([]domain.Command), args.Error(1)
}

package handlers

import (
	"context"
	"testing"
	"time"

	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestScheduledTaskHandler_TopicName(t *testing.T) {
	orm := &MockORM{}
	handler := NewScheduledTaskHandler(orm)

	topic := handler.TopicName()
	assert.Equal(t, pubsub.Topic("scheduled_tasks"), topic)
}

func TestScheduledTaskHandler_Create(t *testing.T) {
	orm := &MockORM{}
	handler := NewScheduledTaskHandler(orm)

	// Mock the ORM Create method
	orm.On("WithContext", mock.Anything).Return(orm)
	orm.On("Create", mock.AnythingOfType("*handlers.ScheduledTaskData")).Return(orm)
	orm.On("Error").Return(nil)

	// Create test data as struct
	testData := struct {
		ID               string
		Version          int
		TenantID         string
		DeviceID         string
		CommandTemplates string
		Schedule         string
		IsActive         bool
		CreatedAt        utils.Time
		UpdatedAt        utils.Time
	}{
		ID:               "test-id",
		Version:          1,
		TenantID:         "tenant-1",
		DeviceID:         "device-1",
		CommandTemplates: `[{"device":{"id":"device-1"},"port":15,"priority":"NORMAL","payload":{"index":1,"value":100},"wait_for":"5s"}]`,
		Schedule:         "* * * * *",
		IsActive:         true,
		CreatedAt:        utils.Time{Time: time.Now()},
		UpdatedAt:        utils.Time{Time: time.Now()},
	}

	ctx := context.Background()
	key := pubsub.Key("test-id")

	err := handler.Create(ctx, key, testData)
	require.NoError(t, err)

	orm.AssertExpectations(t)
}

func TestScheduledTaskHandler_GetByID(t *testing.T) {
	orm := &MockORM{}
	handler := NewScheduledTaskHandler(orm)

	// Mock the ORM First method
	orm.On("WithContext", mock.Anything).Return(orm)
	orm.On("First", mock.Anything, mock.Anything).Return(orm)
	orm.On("Error").Return(nil)

	ctx := context.Background()
	_, err := handler.GetByID(ctx, "test-id")
	require.NoError(t, err)

	orm.AssertExpectations(t)
}

func TestScheduledTaskHandler_Update(t *testing.T) {
	orm := &MockORM{}
	handler := NewScheduledTaskHandler(orm)

	// Mock the ORM methods for update
	orm.On("WithContext", mock.Anything).Return(orm)
	orm.On("First", mock.Anything, mock.Anything).Return(orm)
	orm.On("Error").Return(nil).Once() // First call for First
	orm.On("Save", mock.AnythingOfType("*handlers.ScheduledTaskData")).Return(orm)
	orm.On("Error").Return(nil).Once() // Second call for Save

	// Create test data as struct
	testData := struct {
		ID               string
		Version          int
		TenantID         string
		DeviceID         string
		CommandTemplates string
		Schedule         string
		IsActive         bool
		CreatedAt        utils.Time
		UpdatedAt        utils.Time
	}{
		ID:               "test-id",
		Version:          2,
		TenantID:         "tenant-1",
		DeviceID:         "device-1",
		CommandTemplates: `[{"device":{"id":"device-1"},"port":15,"priority":"HIGH","payload":{"index":2,"value":200},"wait_for":"10s"}]`,
		Schedule:         "*/5 * * * *",
		IsActive:         false,
		CreatedAt:        utils.Time{Time: time.Now()},
		UpdatedAt:        utils.Time{Time: time.Now()},
	}

	ctx := context.Background()
	key := pubsub.Key("test-id")

	err := handler.Update(ctx, key, testData)
	require.NoError(t, err)

	orm.AssertExpectations(t)
}

func TestScheduledTaskHandler_ExtractScheduledTaskFields(t *testing.T) {
	orm := &MockORM{}
	handler := NewScheduledTaskHandler(orm)

	// Create test data with CommandTemplate structure as struct
	testData := struct {
		ID               string
		Version          int
		TenantID         string
		DeviceID         string
		CommandTemplates string
		Schedule         string
		IsActive         bool
		CreatedAt        utils.Time
		UpdatedAt        utils.Time
	}{
		ID:               "test-id",
		Version:          1,
		TenantID:         "tenant-1",
		DeviceID:         "device-1",
		CommandTemplates: `[{"device":{"id":"device-1"},"port":15,"priority":"NORMAL","payload":{"index":1,"value":100},"wait_for":"5s"}]`,
		Schedule:         "* * * * *",
		IsActive:         true,
		CreatedAt:        utils.Time{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		UpdatedAt:        utils.Time{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
	}

	result := handler.extractScheduledTaskFields(testData)

	assert.Equal(t, "test-id", result.ID)
	assert.Equal(t, 1, result.Version)
	assert.Equal(t, "tenant-1", result.TenantID)
	assert.Equal(t, "device-1", result.DeviceID)
	assert.Equal(t, `[{"device":{"id":"device-1"},"port":15,"priority":"NORMAL","payload":{"index":1,"value":100},"wait_for":"5s"}]`, result.CommandTemplates)
	assert.Equal(t, "* * * * *", result.Schedule)
	assert.True(t, result.IsActive)
	assert.Equal(t, utils.Time{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)}, result.CreatedAt)
	assert.Equal(t, utils.Time{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)}, result.UpdatedAt)
}

func TestScheduledTaskHandler_ToDomainScheduledTask(t *testing.T) {
	orm := &MockORM{}
	handler := NewScheduledTaskHandler(orm)

	internalData := ScheduledTaskData{
		ID:               "test-id",
		Version:          1,
		TenantID:         "tenant-1",
		DeviceID:         "device-1",
		CommandTemplates: `[{"device":{"id":"device-1"},"port":15,"priority":"NORMAL","payload":{"index":1,"value":100},"wait_for":"5s"}]`,
		Schedule:         "* * * * *",
		IsActive:         true,
		CreatedAt:        utils.Time{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		UpdatedAt:        utils.Time{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		LastExecutedAt:   nil,
		DeletedAt:        nil,
	}

	result := handler.toDomainScheduledTask(internalData)

	expected := map[string]any{
		"id":                "test-id",
		"version":           1,
		"tenant_id":         "tenant-1",
		"device_id":         "device-1",
		"command_templates": `[{"device":{"id":"device-1"},"port":15,"priority":"NORMAL","payload":{"index":1,"value":100},"wait_for":"5s"}]`,
		"schedule":          "* * * * *",
		"is_active":         true,
		"created_at":        utils.Time{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		"updated_at":        utils.Time{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		"last_executed_at":  (*utils.Time)(nil),
		"deleted_at":        (*utils.Time)(nil),
	}

	assert.Equal(t, expected, result)
}

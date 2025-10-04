package internal_test

import (
	"encoding/json"
	"testing"
	"time"
	"zensor-server/internal/control_plane/httpapi/internal"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchedulingConfigurationRequest_ToSchedulingConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		request  internal.SchedulingConfigurationRequest
		expected domain.SchedulingConfiguration
	}{
		{
			name: "interval scheduling configuration",
			request: internal.SchedulingConfigurationRequest{
				Type:        "interval",
				InitialDay:  timePtr(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)),
				DayInterval: intPtr(2),
				ExecutionTime: stringPtr("02:00"),
			},
			expected: domain.SchedulingConfiguration{
				Type:          domain.SchedulingTypeInterval,
				InitialDay:    &utils.Time{Time: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
				DayInterval:   intPtr(2),
				ExecutionTime: stringPtr("02:00"),
			},
		},
		{
			name: "cron scheduling configuration",
			request: internal.SchedulingConfigurationRequest{
				Type:     "cron",
				Schedule: stringPtr("0 0 * * *"),
			},
			expected: domain.SchedulingConfiguration{
				Type: domain.SchedulingTypeCron,
			},
		},
		{
			name: "empty configuration",
			request: internal.SchedulingConfigurationRequest{
				Type: "",
			},
			expected: domain.SchedulingConfiguration{
				Type: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.ToSchedulingConfiguration()
			assert.Equal(t, tt.expected.Type, result.Type)
			if tt.expected.InitialDay != nil {
				assert.NotNil(t, result.InitialDay)
				assert.Equal(t, tt.expected.InitialDay.Time, result.InitialDay.Time)
			} else {
				assert.Nil(t, result.InitialDay)
			}
			assert.Equal(t, tt.expected.DayInterval, result.DayInterval)
			assert.Equal(t, tt.expected.ExecutionTime, result.ExecutionTime)
		})
	}
}

func TestFromSchedulingConfiguration(t *testing.T) {
	initialDay := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	nextExecution := time.Date(2024, 1, 17, 2, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		config         domain.SchedulingConfiguration
		nextExecution  *time.Time
		expectedFields map[string]interface{}
	}{
		{
			name: "interval scheduling with next execution",
			config: domain.SchedulingConfiguration{
				Type:          domain.SchedulingTypeInterval,
				InitialDay:    &utils.Time{Time: initialDay},
				DayInterval:   intPtr(2),
				ExecutionTime: stringPtr("02:00"),
			},
			nextExecution: &nextExecution,
			expectedFields: map[string]interface{}{
				"type":           "interval",
				"initial_day":    &initialDay,
				"day_interval":   2,
				"execution_time": "02:00",
				"next_execution": &nextExecution,
			},
		},
		{
			name: "cron scheduling without next execution",
			config: domain.SchedulingConfiguration{
				Type: domain.SchedulingTypeCron,
			},
			nextExecution: nil,
			expectedFields: map[string]interface{}{
				"type": "cron",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := internal.FromSchedulingConfiguration(tt.config, tt.nextExecution)
			require.NotNil(t, result)
			
			assert.Equal(t, tt.expectedFields["type"], result.Type)
			
			if expectedInitialDay, ok := tt.expectedFields["initial_day"].(*time.Time); ok {
				assert.NotNil(t, result.InitialDay)
				assert.Equal(t, *expectedInitialDay, *result.InitialDay)
			} else {
				assert.Nil(t, result.InitialDay)
			}
			
			if expectedDayInterval, ok := tt.expectedFields["day_interval"].(int); ok {
				assert.NotNil(t, result.DayInterval)
				assert.Equal(t, expectedDayInterval, *result.DayInterval)
			} else {
				assert.Nil(t, result.DayInterval)
			}
			
			if expectedExecutionTime, ok := tt.expectedFields["execution_time"].(string); ok {
				assert.NotNil(t, result.ExecutionTime)
				assert.Equal(t, expectedExecutionTime, *result.ExecutionTime)
			} else {
				assert.Nil(t, result.ExecutionTime)
			}
			
			if expectedNextExecution, ok := tt.expectedFields["next_execution"].(*time.Time); ok {
				assert.NotNil(t, result.NextExecution)
				assert.Equal(t, *expectedNextExecution, *result.NextExecution)
			} else {
				assert.Nil(t, result.NextExecution)
			}
		})
	}
}

func TestScheduledTaskCreateRequest_JSONSerialization(t *testing.T) {
	initialDay := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	
	request := internal.ScheduledTaskCreateRequest{
		Commands: []internal.CommandSendPayloadRequest{
			{
				Index:    1,
				Value:    100,
				Priority: "NORMAL",
				WaitFor:  utils.Duration(0),
			},
		},
		Scheduling: &internal.SchedulingConfigurationRequest{
			Type:        "interval",
			InitialDay:  &initialDay,
			DayInterval: intPtr(2),
			ExecutionTime: stringPtr("02:00"),
		},
		IsActive: true,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(request)
	require.NoError(t, err)
	
	var unmarshaled internal.ScheduledTaskCreateRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)
	
	assert.Equal(t, request.Commands, unmarshaled.Commands)
	assert.Equal(t, request.IsActive, unmarshaled.IsActive)
	require.NotNil(t, unmarshaled.Scheduling)
	assert.Equal(t, request.Scheduling.Type, unmarshaled.Scheduling.Type)
	assert.Equal(t, request.Scheduling.InitialDay, unmarshaled.Scheduling.InitialDay)
	assert.Equal(t, request.Scheduling.DayInterval, unmarshaled.Scheduling.DayInterval)
	assert.Equal(t, request.Scheduling.ExecutionTime, unmarshaled.Scheduling.ExecutionTime)
}

func TestScheduledTaskResponse_JSONSerialization(t *testing.T) {
	initialDay := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	nextExecution := time.Date(2024, 1, 17, 2, 0, 0, 0, time.UTC)
	
	response := internal.ScheduledTaskResponse{
		ID:       "test-id",
		DeviceID: "test-device-id",
		Commands: []internal.CommandSendPayloadRequest{
			{
				Index:    1,
				Value:    100,
				Priority: "NORMAL",
				WaitFor:  utils.Duration(0),
			},
		},
		Scheduling: &internal.SchedulingConfigurationResponse{
			Type:        "interval",
			InitialDay:  &initialDay,
			DayInterval: intPtr(2),
			ExecutionTime: stringPtr("02:00"),
			NextExecution: &nextExecution,
		},
		IsActive: true,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	require.NoError(t, err)
	
	var unmarshaled internal.ScheduledTaskResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)
	
	assert.Equal(t, response.ID, unmarshaled.ID)
	assert.Equal(t, response.DeviceID, unmarshaled.DeviceID)
	assert.Equal(t, response.Commands, unmarshaled.Commands)
	assert.Equal(t, response.IsActive, unmarshaled.IsActive)
	require.NotNil(t, unmarshaled.Scheduling)
	assert.Equal(t, response.Scheduling.Type, unmarshaled.Scheduling.Type)
	assert.Equal(t, response.Scheduling.InitialDay, unmarshaled.Scheduling.InitialDay)
	assert.Equal(t, response.Scheduling.DayInterval, unmarshaled.Scheduling.DayInterval)
	assert.Equal(t, response.Scheduling.ExecutionTime, unmarshaled.Scheduling.ExecutionTime)
	assert.Equal(t, response.Scheduling.NextExecution, unmarshaled.Scheduling.NextExecution)
}

// Helper functions
func timePtr(t time.Time) *time.Time {
	return &t
}

func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

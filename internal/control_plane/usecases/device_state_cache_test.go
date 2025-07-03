package usecases

import (
	"context"
	"fmt"
	"testing"
	"zensor-server/internal/data_plane/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDeviceStateCacheService(t *testing.T) {
	cache := NewDeviceStateCacheService()
	assert.NotNil(t, cache)
}

func TestSimpleDeviceStateCacheService_SetState(t *testing.T) {
	cache := NewDeviceStateCacheService()
	ctx := context.Background()

	deviceID := "test-device-1"
	sensorData := map[string][]dto.SensorData{
		"temperature": {
			{Index: 0, Value: 25.5},
			{Index: 1, Value: 26.2},
		},
		"humidity": {
			{Index: 0, Value: 60.0},
		},
	}

	err := cache.SetState(ctx, deviceID, sensorData)
	require.NoError(t, err)

	// Verify the state was cached
	state, exists := cache.GetState(ctx, deviceID)
	assert.True(t, exists)
	assert.Equal(t, deviceID, state.DeviceID)
	assert.Len(t, state.Data, 2)
	assert.Len(t, state.Data["temperature"], 2)
	assert.Len(t, state.Data["humidity"], 1)
	assert.Equal(t, 25.5, state.Data["temperature"][0].Value)
	assert.Equal(t, 26.2, state.Data["temperature"][1].Value)
	assert.Equal(t, 60.0, state.Data["humidity"][0].Value)
}

func TestSimpleDeviceStateCacheService_GetState_NotExists(t *testing.T) {
	cache := NewDeviceStateCacheService()
	ctx := context.Background()

	state, exists := cache.GetState(ctx, "non-existent-device")
	assert.False(t, exists)
	assert.Empty(t, state.DeviceID)
}

func TestSimpleDeviceStateCacheService_GetAllDeviceIDs(t *testing.T) {
	cache := NewDeviceStateCacheService()
	ctx := context.Background()

	// Add multiple device states
	device1Data := map[string][]dto.SensorData{
		"temperature": {{Index: 0, Value: 25.0}},
	}
	device2Data := map[string][]dto.SensorData{
		"humidity": {{Index: 0, Value: 70.0}},
	}

	err := cache.SetState(ctx, "device-1", device1Data)
	require.NoError(t, err)
	err = cache.SetState(ctx, "device-2", device2Data)
	require.NoError(t, err)

	deviceIDs := cache.GetAllDeviceIDs(ctx)
	assert.Len(t, deviceIDs, 2)
	assert.Contains(t, deviceIDs, "device-1")
	assert.Contains(t, deviceIDs, "device-2")
}

func TestSimpleDeviceStateCacheService_SetState_Overwrites(t *testing.T) {
	cache := NewDeviceStateCacheService()
	ctx := context.Background()

	deviceID := "test-device"
	initialData := map[string][]dto.SensorData{
		"temperature": {{Index: 0, Value: 25.0}},
	}
	updatedData := map[string][]dto.SensorData{
		"temperature": {{Index: 0, Value: 30.0}},
		"humidity":    {{Index: 0, Value: 65.0}},
	}

	// Add initial state
	err := cache.SetState(ctx, deviceID, initialData)
	require.NoError(t, err)

	// Update with new data
	err = cache.SetState(ctx, deviceID, updatedData)
	require.NoError(t, err)

	// Verify the state was updated
	state, exists := cache.GetState(ctx, deviceID)
	assert.True(t, exists)
	assert.Len(t, state.Data, 2)
	assert.Equal(t, 30.0, state.Data["temperature"][0].Value)
	assert.Equal(t, 65.0, state.Data["humidity"][0].Value)
}

func TestSimpleDeviceStateCacheService_ConcurrentAccess(t *testing.T) {
	cache := NewDeviceStateCacheService()
	ctx := context.Background()

	// Test concurrent updates
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			deviceID := fmt.Sprintf("device-%d", id)
			sensorData := map[string][]dto.SensorData{
				"temperature": {{Index: 0, Value: float64(id)}},
			}
			cache.SetState(ctx, deviceID, sensorData)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all states were cached
	deviceIDs := cache.GetAllDeviceIDs(ctx)
	assert.Len(t, deviceIDs, 10)
}

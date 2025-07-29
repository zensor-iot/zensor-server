package usecases

import (
	"context"
	"time"
	"zensor-server/internal/data_plane/dto"
)

//go:generate mockgen -source=device_state_cache.go -destination=../../../test/unit/doubles/control_plane/usecases/device_state_cache_mock.go -package=usecases -mock_names=DeviceStateCacheService=MockDeviceStateCacheService

// DeviceState represents the cached state of a device
type DeviceState struct {
	DeviceID  string
	Timestamp time.Time
	Data      map[string][]SensorData
}

// SensorData represents a single sensor reading
type SensorData struct {
	Index int
	Value float64
}

// DeviceStateCacheService defines the interface for caching device states
type DeviceStateCacheService interface {
	// SetState stores the current state of a device in the cache
	SetState(ctx context.Context, deviceID string, data map[string][]dto.SensorData) error

	// GetState retrieves the cached state of a device
	GetState(ctx context.Context, deviceID string) (DeviceState, bool)

	// GetAllDeviceIDs returns all device IDs that have cached states
	GetAllDeviceIDs(ctx context.Context) []string
}

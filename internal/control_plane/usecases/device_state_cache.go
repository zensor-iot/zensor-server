package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
	"zensor-server/internal/data_plane/dto"
)

// DeviceState represents the last known state of a device
type DeviceState struct {
	DeviceID  string                  `json:"device_id"`
	Timestamp time.Time               `json:"timestamp"`
	Data      map[string][]SensorData `json:"data"`
}

// SensorData represents a single sensor reading
type SensorData struct {
	Index uint    `json:"index"`
	Value float64 `json:"value"`
}

// DeviceStateCacheService manages the in-memory cache of device states
// Only SetState and GetState are provided for simplicity
type DeviceStateCacheService interface {
	// SetState sets the cached state for a device
	SetState(ctx context.Context, deviceID string, data map[string][]dto.SensorData) error

	// GetState retrieves the cached state for a device
	GetState(ctx context.Context, deviceID string) (DeviceState, bool)

	// GetAllDeviceIDs returns all device IDs that have cached states
	GetAllDeviceIDs(ctx context.Context) []string
}

// SimpleDeviceStateCacheService implements DeviceStateCacheService
type SimpleDeviceStateCacheService struct {
	states map[string]DeviceState
	mutex  sync.RWMutex
}

// NewDeviceStateCacheService creates a new device state cache service
func NewDeviceStateCacheService() DeviceStateCacheService {
	return &SimpleDeviceStateCacheService{
		states: make(map[string]DeviceState),
	}
}

// SetState sets the cached state for a device
func (s *SimpleDeviceStateCacheService) SetState(ctx context.Context, deviceID string, data map[string][]dto.SensorData) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Convert dto.SensorData to our internal SensorData
	convertedData := make(map[string][]SensorData)
	for sensorType, sensorData := range data {
		convertedData[sensorType] = make([]SensorData, len(sensorData))
		for i, sd := range sensorData {
			convertedData[sensorType][i] = SensorData{
				Index: sd.Index,
				Value: sd.Value,
			}
		}
	}

	s.states[deviceID] = DeviceState{
		DeviceID:  deviceID,
		Timestamp: time.Now(),
		Data:      convertedData,
	}

	slog.Info("device state set in cache",
		slog.String("device_id", deviceID),
		slog.Int("sensor_types", len(convertedData)),
		slog.Int("total_cached_states", len(s.states)),
		slog.String("cache_instance", fmt.Sprintf("%p", s)))

	return nil
}

// GetState retrieves the cached state for a device
func (s *SimpleDeviceStateCacheService) GetState(ctx context.Context, deviceID string) (DeviceState, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	state, exists := s.states[deviceID]
	return state, exists
}

// GetAllDeviceIDs returns all device IDs that have cached states
func (s *SimpleDeviceStateCacheService) GetAllDeviceIDs(ctx context.Context) []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	deviceIDs := make([]string, 0, len(s.states))
	for deviceID := range s.states {
		deviceIDs = append(deviceIDs, deviceID)
	}

	slog.Info("GetAllDeviceIDs called",
		slog.Int("total_states", len(s.states)),
		slog.Int("returned_ids", len(deviceIDs)),
		slog.String("cache_instance", fmt.Sprintf("%p", s)))
	return deviceIDs
}

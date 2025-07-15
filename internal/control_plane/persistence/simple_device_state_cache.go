package persistence

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/data_plane/dto"
)

// SimpleDeviceStateCacheService implements usecases.DeviceStateCacheService
// In-memory implementation for development/testing
// Not suitable for distributed/multi-instance deployments
type SimpleDeviceStateCacheService struct {
	states map[string]usecases.DeviceState
	mutex  sync.RWMutex
}

// NewSimpleDeviceStateCacheService creates a new in-memory device state cache service
func NewSimpleDeviceStateCacheService() usecases.DeviceStateCacheService {
	return &SimpleDeviceStateCacheService{
		states: make(map[string]usecases.DeviceState),
	}
}

func (s *SimpleDeviceStateCacheService) SetState(ctx context.Context, deviceID string, data map[string][]dto.SensorData) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	convertedData := make(map[string][]usecases.SensorData)
	for sensorType, sensorData := range data {
		convertedData[sensorType] = make([]usecases.SensorData, len(sensorData))
		for i, sd := range sensorData {
			convertedData[sensorType][i] = usecases.SensorData{
				Index: int(sd.Index),
				Value: sd.Value,
			}
		}
	}
	s.states[deviceID] = usecases.DeviceState{
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

func (s *SimpleDeviceStateCacheService) GetState(ctx context.Context, deviceID string) (usecases.DeviceState, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	state, exists := s.states[deviceID]
	return state, exists
}

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

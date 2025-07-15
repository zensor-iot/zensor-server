package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/data_plane/dto"
	"zensor-server/internal/infra/cache"
)

// RedisDeviceStateCacheService implements usecases.DeviceStateCacheService using Redis
type RedisDeviceStateCacheService struct {
	cache      cache.Cache
	keyPrefix  string
	defaultTTL time.Duration
}

// RedisDeviceStateCacheConfig holds configuration for the Redis device state cache
type RedisDeviceStateCacheConfig struct {
	Cache      cache.Cache
	KeyPrefix  string
	DefaultTTL time.Duration
}

func DefaultRedisDeviceStateCacheConfig() *RedisDeviceStateCacheConfig {
	return &RedisDeviceStateCacheConfig{
		KeyPrefix:  "device_state:",
		DefaultTTL: 24 * time.Hour,
	}
}

func NewRedisDeviceStateCacheService(config *RedisDeviceStateCacheConfig) (usecases.DeviceStateCacheService, error) {
	if config == nil {
		config = DefaultRedisDeviceStateCacheConfig()
	}
	if config.Cache == nil {
		return nil, fmt.Errorf("cache instance is required")
	}
	service := &RedisDeviceStateCacheService{
		cache:      config.Cache,
		keyPrefix:  config.KeyPrefix,
		defaultTTL: config.DefaultTTL,
	}
	slog.Info("Redis device state cache service initialized",
		slog.String("key_prefix", config.KeyPrefix),
		slog.Duration("default_ttl", config.DefaultTTL))
	return service, nil
}

func (s *RedisDeviceStateCacheService) SetState(ctx context.Context, deviceID string, data map[string][]dto.SensorData) error {
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
	deviceState := usecases.DeviceState{
		DeviceID:  deviceID,
		Timestamp: time.Now(),
		Data:      convertedData,
	}
	key := s.makeDeviceStateKey(deviceID)
	success := s.cache.Set(ctx, key, deviceState, s.defaultTTL)
	if !success {
		return fmt.Errorf("failed to set device state in cache for device %s", deviceID)
	}
	slog.Info("device state set in Redis cache",
		slog.String("device_id", deviceID),
		slog.Int("sensor_types", len(convertedData)),
		slog.String("cache_key", key))
	return nil
}

func (s *RedisDeviceStateCacheService) GetState(ctx context.Context, deviceID string) (usecases.DeviceState, bool) {
	key := s.makeDeviceStateKey(deviceID)
	value, found := s.cache.Get(ctx, key)
	if !found {
		return usecases.DeviceState{}, false
	}
	var deviceState usecases.DeviceState
	switch v := value.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &deviceState); err != nil {
			slog.Error("failed to unmarshal device state from cache",
				slog.String("device_id", deviceID),
				slog.String("error", err.Error()))
			return usecases.DeviceState{}, false
		}
	case map[string]any:
		data, err := json.Marshal(v)
		if err != nil {
			slog.Error("failed to marshal device state for conversion",
				slog.String("device_id", deviceID),
				slog.String("error", err.Error()))
			return usecases.DeviceState{}, false
		}
		if err := json.Unmarshal(data, &deviceState); err != nil {
			slog.Error("failed to unmarshal device state from cache",
				slog.String("device_id", deviceID),
				slog.String("error", err.Error()))
			return usecases.DeviceState{}, false
		}
	default:
		slog.Error("unexpected value type in cache",
			slog.String("device_id", deviceID),
			slog.String("type", fmt.Sprintf("%T", value)))
		return usecases.DeviceState{}, false
	}
	return deviceState, true
}

func (s *RedisDeviceStateCacheService) GetAllDeviceIDs(ctx context.Context) []string {
	pattern := s.keyPrefix + "*"
	keys, err := s.cache.Keys(ctx, pattern)
	if err != nil {
		slog.Error("failed to get device state keys from cache",
			slog.String("pattern", pattern),
			slog.String("error", err.Error()))
		return []string{}
	}
	deviceIDs := make([]string, 0, len(keys))
	for _, key := range keys {
		deviceID := s.extractDeviceIDFromKey(key)
		if deviceID != "" {
			deviceIDs = append(deviceIDs, deviceID)
		}
	}
	slog.Info("GetAllDeviceIDs called",
		slog.Int("total_keys", len(keys)),
		slog.Int("returned_ids", len(deviceIDs)))
	return deviceIDs
}

func (s *RedisDeviceStateCacheService) makeDeviceStateKey(deviceID string) string {
	return s.keyPrefix + deviceID
}

func (s *RedisDeviceStateCacheService) extractDeviceIDFromKey(key string) string {
	if !strings.HasPrefix(key, s.keyPrefix) {
		return ""
	}
	return strings.TrimPrefix(key, s.keyPrefix)
}

# Redis Device State Cache Migration

## Overview

The application has been updated to use `RedisDeviceStateCacheService` instead of `SimpleDeviceStateCacheService` for better scalability and persistence.

## Changes Made

### 1. Wire Configuration Update

The dependency injection configuration in `cmd/api/wire/control_plane.go` has been updated to use Redis:

```go
func provideDeviceStateCacheService() usecases.DeviceStateCacheService {
    deviceStateCacheOnce.Do(func() {
        appConfig := provideAppConfig()
        
        // Create Redis cache instance
        redisCache, err := cache.NewRedisCache(&cache.RedisConfig{
            Addr:     appConfig.Redis.Addr,
            Password: appConfig.Redis.Password,
            DB:       appConfig.Redis.DB,
        })
        if err != nil {
            slog.Error("failed to create Redis cache", slog.String("error", err.Error()))
            // Fallback to simple cache if Redis is not available
            deviceStateCacheService = persistence.NewSimpleDeviceStateCacheService()
            slog.Info("falling back to simple device state cache service")
            return
        }
        
        // Create Redis device state cache service
        deviceStateCacheService, err = persistence.NewRedisDeviceStateCacheService(&persistence.RedisDeviceStateCacheConfig{
            Cache:      redisCache,
            KeyPrefix:  "device_state:",
            DefaultTTL: 24 * time.Hour,
        })
        if err != nil {
            slog.Error("failed to create Redis device state cache service", slog.String("error", err.Error()))
            // Fallback to simple cache if Redis service creation fails
            deviceStateCacheService = persistence.NewSimpleDeviceStateCacheService()
            slog.Info("falling back to simple device state cache service")
            return
        }
        
        slog.Info("Redis device state cache service singleton created")
    })
    return deviceStateCacheService
}
```

### 2. Configuration

Redis configuration is already available in `config/server.yaml`:

```yaml
redis:
  addr: "localhost:6379"
  password: ""
  db: 0
```

### 3. Key Format

Device state is stored in Redis using the following key format:
```
device_state:<device_id>
```

Example: `device_state:abc123`

## Benefits

1. **Persistence**: Device state survives application restarts
2. **Scalability**: Can be shared across multiple application instances
3. **TTL Support**: Automatic expiration of old device states (24 hours by default)
4. **Fallback**: Graceful fallback to in-memory cache if Redis is unavailable

## Fallback Behavior

If Redis is not available or fails to initialize, the application will automatically fall back to the simple in-memory cache (`SimpleDeviceStateCacheService`). This ensures the application continues to function even without Redis.

## Testing

The switch has been tested with:
- Unit tests for both simple and Redis cache implementations
- Application build verification
- Redis connection tests

## Migration Notes

- No data migration is required as device state is transient
- The application will automatically use Redis when available
- Existing functionality remains unchanged
- All existing tests continue to pass 
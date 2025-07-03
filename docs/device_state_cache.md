# Device State Cache

## Overview

The Device State Cache is an in-memory caching system that stores the last known state of each device. This provides quick access to device sensor data without requiring database queries, enabling immediate response to WebSocket connections.

## Architecture

### Components

1. **DeviceStateCacheService** - Interface and implementation for managing cached device states
2. **LoRa Integration Worker** - Updates the cache when new sensor data arrives
3. **WebSocket Controller** - Sends cached states to new connections

### Data Flow

```
Sensor Data → LoRa Worker → Device State Cache → WebSocket Clients
```

## Features

### In-Memory Storage
- Fast access to device states
- Thread-safe operations with read/write mutex
- Simple interface with only essential methods

### Real-Time Updates
- Cache is updated immediately when new sensor data arrives
- WebSocket clients receive live updates
- Historical data is preserved until overwritten

### WebSocket Integration
- New connections receive all cached device states immediately
- Real-time updates for ongoing connections
- Efficient message delivery with non-blocking channels

## API

### DeviceStateCacheService Interface

```go
type DeviceStateCacheService interface {
    // SetState sets the cached state for a device
    SetState(ctx context.Context, deviceID string, data map[string][]dto.SensorData) error
    
    // GetState retrieves the cached state for a device
    GetState(ctx context.Context, deviceID string) (DeviceState, bool)
    
    // GetAllDeviceIDs returns all device IDs that have cached states
    GetAllDeviceIDs(ctx context.Context) []string
}
```

### DeviceState Structure

```go
type DeviceState struct {
    DeviceID  string                 `json:"device_id"`
    Timestamp time.Time              `json:"timestamp"`
    Data      map[string][]SensorData `json:"data"`
}

type SensorData struct {
    Index uint    `json:"index"`
    Value float64 `json:"value"`
}
```

## WebSocket Messages

### Device State Message
When a new WebSocket client connects, it receives cached device states in this format:

```json
{
    "type": "device_state",
    "device_id": "device-123",
    "timestamp": "2024-01-01T12:00:00Z",
    "data": {
        "temperature": [
            {"index": 0, "value": 25.5},
            {"index": 1, "value": 26.2}
        ],
        "humidity": [
            {"index": 0, "value": 60.0}
        ]
    }
}
```

### Real-Time Device Message
Ongoing sensor updates are sent as:

```json
{
    "type": "device_state",
    "device_id": "device-123",
    "timestamp": "2024-01-01T12:00:00Z",
    "data": {
        "temperature": [{"index": 0, "value": 25.5}],
        "humidity": [{"index": 0, "value": 60.0}]
    }
}
```

## Configuration

The device state cache is automatically initialized with the application and requires no additional configuration. It uses the existing dependency injection system through Wire.

## Testing

### Unit Tests
Run the device state cache unit tests:

```bash
go test ./internal/control_plane/usecases/device_state_cache_test.go ./internal/control_plane/usecases/device_state_cache.go -v
```

### Functional Tests
Run the WebSocket integration tests:

```bash
just functional @device_state_cache
```

## Performance Considerations

- **Memory Usage**: Cache grows with the number of active devices
- **Concurrency**: Thread-safe operations with minimal lock contention
- **Scalability**: In-memory storage limits horizontal scaling (consider Redis for distributed deployments)

## Future Enhancements

1. **Persistence**: Add option to persist cache to database
2. **TTL**: Implement time-to-live for cached states
3. **Compression**: Compress cached data for memory efficiency
4. **Distributed Cache**: Support Redis or similar for multi-instance deployments
5. **Metrics**: Add cache hit/miss metrics and monitoring 
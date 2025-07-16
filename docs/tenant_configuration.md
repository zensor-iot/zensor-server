# Tenant Configuration

## Overview

The Tenant Configuration entity allows storing configuration settings specific to each tenant. The first configuration option implemented is timezone support, which enables tenants to set their preferred timezone for date/time operations.

## Architecture

### Domain Model

The `TenantConfiguration` domain model is located in `internal/shared_kernel/domain/tenant_configuration.go`:

```go
type TenantConfiguration struct {
    ID        ID
    TenantID  ID
    Timezone  string
    Version   int
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Key Features

- **One-to-One Relationship**: Each tenant can have exactly one configuration
- **Timezone Support**: Stores the tenant's preferred timezone (e.g., "America/New_York", "UTC")
- **Optimistic Locking**: Uses version field for concurrent update protection
- **Builder Pattern**: Uses the established builder pattern for object creation

### API Endpoints

The Tenant Configuration is exposed as a subresource of Tenant:

- `GET /v1/tenants/{id}/configuration` - Get tenant configuration
- `POST /v1/tenants/{id}/configuration` - Create tenant configuration
- `PUT /v1/tenants/{id}/configuration` - Update tenant configuration

### Request/Response Models

#### Create Request
```json
{
  "timezone": "America/New_York"
}
```

#### Update Request
```json
{
  "timezone": "Europe/London",
  "version": 1
}
```

#### Response
```json
{
  "id": "config-uuid",
  "tenant_id": "tenant-uuid",
  "timezone": "America/New_York",
  "version": 1,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

## Implementation Details

### Persistence Layer

- **Internal Model**: `internal/control_plane/persistence/internal/tenant_configuration.go`
- **Repository**: `internal/control_plane/persistence/tenant_configuration_repository.go`
- **Database Table**: `tenant_configurations`
- **Unique Constraint**: One configuration per tenant (enforced by unique index on `tenant_id`)

### Service Layer

- **Service Interface**: `usecases.TenantConfigurationService`
- **Implementation**: `usecases.SimpleTenantConfigurationService`
- **Key Methods**:
  - `CreateTenantConfiguration` - Creates new configuration
  - `GetTenantConfiguration` - Retrieves configuration by tenant ID
  - `UpdateTenantConfiguration` - Updates existing configuration with optimistic locking
  - `GetOrCreateTenantConfiguration` - Gets existing or creates default configuration

### Event Streaming

- **Avro Schema**: `schemas/tenant_configuration.avsc`
- **Topic**: `tenant_configurations`
- **Kafka Connect**: `kafka-connect/postgres-tenant_configurations-sink.json`
- **Message Type**: `AvroTenantConfiguration`

### Error Handling

- `ErrTenantConfigurationNotFound` - When configuration doesn't exist for a tenant
- Proper HTTP status codes (404 for not found, 400 for bad request, etc.)
- Optimistic locking conflicts handled gracefully

## Usage Examples

### Creating a Tenant Configuration

```bash
curl -X POST http://localhost:8080/v1/tenants/tenant-uuid/configuration \
  -H "Content-Type: application/json" \
  -d '{"timezone": "America/New_York"}'
```

### Getting a Tenant Configuration

```bash
curl http://localhost:8080/v1/tenants/tenant-uuid/configuration
```

### Updating a Tenant Configuration

```bash
curl -X PUT http://localhost:8080/v1/tenants/tenant-uuid/configuration \
  -H "Content-Type: application/json" \
  -d '{"timezone": "Europe/London", "version": 1}'
```

## Business Logic

### Timezone Validation

The service accepts any string as a timezone value. In a production environment, you might want to add validation to ensure the timezone is valid (e.g., using Go's `time.LoadLocation`).

### Default Configuration

The `GetOrCreateTenantConfiguration` method allows creating a default configuration if one doesn't exist:

```go
config, err := service.GetOrCreateTenantConfiguration(ctx, tenant, "UTC")
```

### Optimistic Locking

Updates require the current version to prevent concurrent modification conflicts:

```go
// This will fail if the configuration was updated by another request
err := service.UpdateTenantConfiguration(ctx, config)
```

## Testing

### Unit Tests

- Domain model tests in `internal/control_plane/persistence/tenant_configuration_repository_test.go`
- Builder pattern tests
- Internal model conversion tests

### Integration Tests

Integration tests should be added to verify:
- HTTP endpoint behavior
- Database persistence
- Event streaming
- Error scenarios

## Future Enhancements

### Additional Configuration Options

The TenantConfiguration entity is designed to be extensible. Future configuration options could include:

- **Date Format**: Preferred date display format
- **Time Format**: 12-hour vs 24-hour time format
- **Language**: Preferred language for UI
- **Currency**: Default currency for financial data
- **Units**: Metric vs Imperial units
- **Notifications**: Email/SMS notification preferences

### Validation

- Add timezone validation using `time.LoadLocation`
- Add JSON schema validation for request bodies
- Add business rule validation (e.g., allowed timezone list)

### Caching

- Add Redis caching for frequently accessed configurations
- Implement cache invalidation on updates

## Database Schema

```sql
CREATE TABLE tenant_configurations (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL UNIQUE,
    timezone VARCHAR(100) NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_tenant_configurations_tenant_id ON tenant_configurations(tenant_id);
```

## Dependencies

- **Domain**: `internal/shared_kernel/domain`
- **Persistence**: `internal/control_plane/persistence`
- **Use Cases**: `internal/control_plane/usecases`
- **HTTP API**: `internal/control_plane/httpapi`
- **Avro**: `internal/shared_kernel/avro`
- **Infrastructure**: `internal/infra/pubsub`, `internal/infra/sql` 
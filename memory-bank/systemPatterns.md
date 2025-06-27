# System Patterns

## Architecture Overview

Zensor Server follows a clean architecture pattern with clear separation of concerns:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP API      │    │   Domain        │    │   Persistence   │
│   Controllers   │◄──►│   Services      │◄──►│   Repositories  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Event Bus     │    │   Workers       │    │   Database      │
│   (Kafka/MQTT)  │    │   (Async)       │    │   (Materialize) │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Domain Model Patterns

### Entity Relationships
- **Tenant** → **Devices** (1:N)
- **Device** → **Tasks** (1:N)
- **Device** → **EvaluationRules** (1:N)
- **Task** → **Commands** (1:N)
- **ScheduledTask** → **Commands** (1:N) - *Template commands*
- **ScheduledTask** → **Tasks** (1:N) - *Generated tasks with ScheduledTaskID*

### Scheduled Task Pattern
Scheduled tasks now follow a template-execution pattern:

1. **ScheduledTask** contains a set of **Commands** as a template
2. When scheduled task executes, it creates a new **Task** with:
   - Fresh command instances (new IDs, current timestamps)
   - Reference to the original **ScheduledTask** via `ScheduledTask` field
3. This allows tracking which scheduled task generated which task executions

```
ScheduledTask (Template)
├── Commands: [Command1, Command2, Command3]
└── Schedule: "0 */6 * * *"

When executed → Creates Task (Execution)
├── Commands: [Command1_new, Command2_new, Command3_new]
├── ScheduledTask: {reference to original ScheduledTask}
└── CreatedAt: "2024-01-15T10:00:00Z"
```

## Data Flow Patterns

### Command Processing Flow
1. **Task Creation** → Commands published to Kafka
2. **Command Worker** → Consumes commands, sends to devices
3. **Device Response** → Events published back to Kafka
4. **Event Processing** → Updates command status, triggers evaluation rules

### Scheduled Task Execution Flow
1. **ScheduledTask Worker** → Evaluates cron schedules
2. **Task Creation** → Creates new Task from ScheduledTask commands
3. **Command Processing** → Normal command flow begins
4. **Tracking** → Task linked back to ScheduledTask via ScheduledTaskID

### Replication Pattern (Local Development)
1. **Event Publishing** → Domain events published to Kafka topics
2. **Replication Service** → Consumes events and persists to database
3. **Materialize Views** → Provides real-time queryable data
4. **API Queries** → Read from Materialize views

## Repository Pattern

### Generic Repository Interface
All repositories follow a consistent pattern:
- `Create(ctx, entity) error`
- `GetByID(ctx, id) (Entity, error)`
- `FindAll(ctx, filters) ([]Entity, error)`
- `Update(ctx, entity) error`

### Specialized Query Methods
- `FindAllByDevice(ctx, device, pagination) ([]Task, int, error)`
- `FindAllByScheduledTask(ctx, scheduledTaskID, pagination) ([]Task, int, error)`
- `FindAllByTenant(ctx, tenantID) ([]ScheduledTask, error)`

## Service Layer Patterns

### Validation Pattern
Services implement business logic validation:
- Command overlap detection
- Device ownership verification
- Tenant isolation enforcement

### Transaction Pattern
- Domain operations wrapped in transactions
- Event publishing on successful commits
- Rollback on validation failures

## Worker Patterns

### Async Worker Interface
All workers implement:
```go
type Worker interface {
    Run(ctx context.Context, done func())
    Shutdown()
}
```

### Scheduled Task Worker
- Uses cron parser for schedule evaluation
- Creates tasks from scheduled task templates
- Maintains execution history via ScheduledTaskID

### Command Worker
- Consumes commands from Kafka
- Handles device communication
- Updates command status

## API Patterns

### RESTful Endpoints
- Resource-based URL structure
- Consistent HTTP status codes
- JSON request/response format

### Pagination Pattern
```go
type Pagination struct {
    Limit  int
    Offset int
}
```

### Error Handling
- Structured error responses
- Consistent error message format
- Proper HTTP status codes

## Event-Driven Patterns

### Event Publishing
- Domain events published to Kafka topics
- Event schema versioning
- Dead letter queue for failed events

### Event Consumption
- Idempotent event processing
- Event ordering preservation
- Retry mechanisms with exponential backoff

## Configuration Patterns

### Environment-Based Configuration
- Development vs production settings
- Feature flags for experimental features
- Conditional service initialization

### Dependency Injection
- Wire-based dependency injection
- Interface-based abstractions
- Testable service boundaries

---

> _This document captures the key architectural patterns and design decisions used throughout the system._ 
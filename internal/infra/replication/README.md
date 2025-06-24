# Replication Module

The replication module provides a bridge between the in-memory pub/sub system and the in-memory database for local development. It ensures that data published to Kafka topics gets persisted to the database.

## Overview

The replication module consists of:

- **Replicator**: The main component that coordinates replication between pub/sub and database
- **TopicHandler**: Interface for handling specific topic replication operations
- **Service**: High-level service that manages the replication process
- **Handlers**: Concrete implementations for different entity types (devices, tenants, etc.)

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Pub/Sub       │───▶│   Replicator     │───▶│   Database      │
│   (Kafka/       │    │                  │    │   (Materialize/ │
│    Memory)      │    │  ┌─────────────┐ │    │    PostgreSQL)  │
└─────────────────┘    │  │TopicHandlers│ │    └─────────────────┘
                       │  └─────────────┘ │
                       └──────────────────┘
```

## Usage

### Basic Setup

The replication module is automatically initialized when the application runs in local mode (`ENV=local`).

### Creating Topic Handlers

To create a new topic handler for a specific entity:

```go
type MyEntityHandler struct {
    orm sql.ORM
}

func NewMyEntityHandler(orm sql.ORM) *MyEntityHandler {
    return &MyEntityHandler{orm: orm}
}

func (h *MyEntityHandler) TopicName() pubsub.Topic {
    return "my_entities"
}

func (h *MyEntityHandler) Create(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
    // Convert message to internal representation
    // Save to database
    return h.orm.WithContext(ctx).Create(&internalEntity).Error()
}

func (h *MyEntityHandler) GetByID(ctx context.Context, id string) (pubsub.Message, error) {
    // Retrieve from database
    // Convert to domain model
    return domainEntity, nil
}

func (h *MyEntityHandler) Update(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
    // Convert message to internal representation
    // Update in database
    return h.orm.WithContext(ctx).Save(&internalEntity).Error()
}
```

## Configuration

The replication module is only active when:
- `ENV=local` environment variable is set
- In-memory pub/sub and database are being used

## Topics Supported

Currently supported topics:
- `devices` - Device entity replication
- `tenants` - Tenant entity replication

## Error Handling

The replication module includes comprehensive error handling:
- Logs all replication operations
- Continues processing even if individual messages fail
- Provides detailed error messages for debugging 
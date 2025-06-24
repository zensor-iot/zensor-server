package replication

import (
	"context"
	"zensor-server/internal/infra/pubsub"
)

// TopicHandler defines the interface for handling topic replication operations
type TopicHandler interface {
	// TopicName returns the name of the topic this handler manages
	TopicName() pubsub.Topic

	// Create handles creating a new record in the database
	Create(ctx context.Context, key pubsub.Key, message pubsub.Message) error

	// GetByID retrieves a record from the database by its ID
	GetByID(ctx context.Context, id string) (pubsub.Message, error)

	// Update handles updating an existing record in the database
	Update(ctx context.Context, key pubsub.Key, message pubsub.Message) error
}

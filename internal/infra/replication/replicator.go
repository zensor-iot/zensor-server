package replication

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
)

// Replicator handles replication of data from pub/sub to database
type Replicator struct {
	consumerFactory pubsub.ConsumerFactory
	orm             sql.ORM
	handlers        map[pubsub.Topic]TopicHandler
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
}

// NewReplicator creates a new replicator instance
func NewReplicator(
	consumerFactory pubsub.ConsumerFactory,
	orm sql.ORM,
) *Replicator {
	ctx, cancel := context.WithCancel(context.Background())

	return &Replicator{
		consumerFactory: consumerFactory,
		orm:             orm,
		handlers:        make(map[pubsub.Topic]TopicHandler),
		ctx:             ctx,
		cancel:          cancel,
	}
}

// RegisterHandler registers a topic handler for replication
func (r *Replicator) RegisterHandler(handler TopicHandler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	topic := handler.TopicName()
	if _, exists := r.handlers[topic]; exists {
		return fmt.Errorf("handler already registered for topic: %s", topic)
	}

	r.handlers[topic] = handler
	slog.Debug("registered topic handler", slog.String("topic", string(topic)))

	return nil
}

// Start begins the replication process for all registered topics
func (r *Replicator) Start() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.handlers) == 0 {
		slog.Warn("no topic handlers registered, replication not started")
		return nil
	}

	slog.Info("starting replication", slog.Int("topics", len(r.handlers)))

	for topic, handler := range r.handlers {
		r.wg.Add(1)
		go r.replicateTopic(topic, handler)
	}

	return nil
}

// Stop gracefully stops the replication process
func (r *Replicator) Stop() {
	slog.Info("stopping replication")
	r.cancel()
	r.wg.Wait()
	slog.Info("replication stopped")
}

// replicateTopic handles replication for a specific topic
func (r *Replicator) replicateTopic(topic pubsub.Topic, handler TopicHandler) {
	defer r.wg.Done()

	consumer := r.consumerFactory.New()
	messageHandler := func(ctx context.Context, key pubsub.Key, msg pubsub.Prototype) error {
		return r.handleMessage(ctx, topic, handler, key, msg.(pubsub.Message))
	}

	// Use the topic name as the consumer group to ensure proper partitioning
	consumerGroup := fmt.Sprintf("replicator-%s", topic)

	slog.Debug("starting topic replication",
		slog.String("topic", string(topic)),
		slog.String("consumer_group", consumerGroup))

	// Start consuming messages
	err := consumer.Consume(topic, messageHandler, nil)
	if err != nil {
		slog.Error("error consuming topic",
			slog.String("topic", string(topic)),
			slog.Any("error", err))
	}
}

// handleMessage processes a single message for replication
func (r *Replicator) handleMessage(ctx context.Context, topic pubsub.Topic, handler TopicHandler, key pubsub.Key, msg pubsub.Message) error {
	slog.Debug("replicating message",
		slog.String("topic", string(topic)),
		slog.String("key", string(key)))

	// Check if record exists
	_, err := handler.GetByID(ctx, string(key))
	if err != nil {
		// Record doesn't exist, create new one
		err = handler.Create(ctx, key, msg)
		if err != nil {
			slog.Error("failed to create record",
				slog.String("topic", string(topic)),
				slog.String("key", string(key)),
				slog.Any("error", err))
			return fmt.Errorf("creating record: %w", err)
		}

		slog.Debug("created new record",
			slog.String("topic", string(topic)),
			slog.String("key", string(key)))
	} else {
		// Record exists, update it
		err = handler.Update(ctx, key, msg)
		if err != nil {
			slog.Error("failed to update record",
				slog.String("topic", string(topic)),
				slog.String("key", string(key)),
				slog.Any("error", err))
			return fmt.Errorf("updating record: %w", err)
		}

		slog.Debug("updated existing record",
			slog.String("topic", string(topic)),
			slog.String("key", string(key)))
	}

	return nil
}

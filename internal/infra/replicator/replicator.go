package replicator

import (
	"context"
	"fmt"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
)

type CreationFunc func(ctx context.Context, orm sql.ORM, msg pubsub.Prototype) error
type GetFunc func(ctx context.Context, orm sql.ORM, msg pubsub.Prototype) (any, error)
type UpdateFunc func(ctx context.Context, orm sql.ORM, msg pubsub.Prototype) error

type TopicHandler struct {
	Topic        pubsub.Topic
	CreationFunc CreationFunc
	GetFunc      GetFunc
	UpdateFunc   UpdateFunc
	Proto        pubsub.Prototype
}

type Replicator struct {
	orm     sql.ORM
	factory *pubsub.Factory
	entries []TopicHandler
}

func NewReplicator(orm sql.ORM, factory *pubsub.Factory, entries []TopicHandler) *Replicator {
	return &Replicator{
		orm:     orm,
		factory: factory,
		entries: entries,
	}
}

func (r *Replicator) Start(ctx context.Context) error {
	consumerFactory := r.factory.GetConsumerFactory()
	consumer := consumerFactory.New()

	for _, entry := range r.entries {
		topic := entry.Topic
		create := entry.CreationFunc
		get := entry.GetFunc
		update := entry.UpdateFunc
		proto := entry.Proto

		err := consumer.Consume(topic, func(msg pubsub.Prototype) error {
			found, err := get(ctx, r.orm, msg)
			if err != nil {
				return fmt.Errorf("get failed: %w", err)
			}
			if found != nil {
				return update(ctx, r.orm, msg)
			}
			return create(ctx, r.orm, msg)
		}, proto)
		if err != nil {
			return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
		}
	}
	return nil
}

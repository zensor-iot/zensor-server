package pubsub

import "context"

//go:generate mockgen -source=pubsub.go -destination=../../../test/unit/doubles/infra/pubsub/pubsub_mock.go -package=pubsub -mock_names=ConsumerFactory=MockConsumerFactory,Consumer=MockConsumer,PublisherFactory=MockPublisherFactory,Publisher=MockPublisher

type PublisherFactory interface {
	New(Topic, Message) (Publisher, error)
}

type Publisher interface {
	Publish(context.Context, Key, Message) error
}

type Key string
type Message any

type ConsumerFactory interface {
	New() Consumer
}

type Consumer interface {
	Consume(Topic, MessageHandler, Prototype) error
}

type Topic string
type MessageHandler func(context.Context, Key, Prototype) error
type ConsumedMessage struct {
	Ctx   context.Context
	Key   Key
	Value Prototype
}
type Prototype any

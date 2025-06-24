package pubsub

import "context"

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
type MessageHandler func(Key, Prototype) error
type Prototype any

package async

import (
	"context"
	"errors"
	"slices"
	"sync"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type BrokerTopicName string

type BrokerMessage struct {
	Event string
	Value any
	Span  trace.Span
	Error error
}

type InternalBrokerSubscriptor interface {
	AddSubscription(b InternalBroker)
}

type InternalBroker interface {
	Subscribe(topic BrokerTopicName) (Subscription, error)
	Unsubscribe(topic BrokerTopicName, subscription Subscription) error
	Publish(ctx context.Context, topic BrokerTopicName, msg BrokerMessage) error
	Stop()
}

type MessageHandler func(msg BrokerMessage)

var _ InternalBroker = (*LocalBroker)(nil)

var ErrTopicNotFound = errors.New("topic not found")
var ErrSubscriptorNotFound = errors.New("subscriptor not found")

func NewLocalBroker() *LocalBroker {
	return &LocalBroker{
		subscriptors: sync.Map{},
	}
}

type LocalBroker struct {
	subscriptors sync.Map
}

type subscriptor struct {
	once         sync.Once
	active       bool
	subscription Subscription
}

type Subscription struct {
	ID       string
	Receiver chan BrokerMessage
}

func (b *LocalBroker) Subscribe(topic BrokerTopicName) (Subscription, error) {
	var subscriptors []*subscriptor
	value, ok := b.subscriptors.Load(topic)
	if !ok {
		subscriptors = make([]*subscriptor, 0)
	} else {
		subscriptors = value.([]*subscriptor)
	}
	id := uuid.NewString()
	receiver := make(chan BrokerMessage)
	subscription := Subscription{ID: id, Receiver: receiver}
	subscriptors = append(subscriptors, &subscriptor{subscription: subscription, active: true})
	b.subscriptors.Store(topic, subscriptors)
	return subscription, nil
}

func (b *LocalBroker) Unsubscribe(topic BrokerTopicName, subscription Subscription) error {
	value, ok := b.subscriptors.Load(topic)
	if !ok {
		return ErrTopicNotFound
	}

	subscriptors := value.([]*subscriptor)
	index := slices.IndexFunc(subscriptors, func(s *subscriptor) bool { return s.subscription.ID == subscription.ID })
	if index < 0 {
		return ErrSubscriptorNotFound
	}

	subscriptors[index].safeClose()

	return nil
}

func (b *LocalBroker) Publish(ctx context.Context, topic BrokerTopicName, msg BrokerMessage) error {
	msg.Span = trace.SpanFromContext(ctx)
	topicSubscriptors, ok := b.subscriptors.Load(topic)
	if !ok {
		return ErrTopicNotFound
	}

	go b.publish(topicSubscriptors.([]*subscriptor), msg)

	return nil
}

func (b *LocalBroker) publish(topicSubscriptors []*subscriptor, msg BrokerMessage) {
	for _, s := range topicSubscriptors {
		if s.active {
			s.subscription.Receiver <- msg
		}
	}
}

func (b *LocalBroker) Stop() {
	b.subscriptors.Range(func(key, value any) bool {
		for _, s := range value.([]*subscriptor) {
			s.safeClose()
		}
		return true
	})
}

func (s *subscriptor) safeClose() {
	s.once.Do(func() {
		s.active = false
		close(s.subscription.Receiver)
	})
}

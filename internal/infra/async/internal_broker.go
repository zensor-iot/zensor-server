// Package async provides an in-process event broker and worker abstractions.
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

//go:generate mockgen -source=internal_broker.go -destination=../../../test/unit/doubles/infra/async/internal_broker_mock.go -package=async -mock_names=InternalBroker=MockInternalBroker
type InternalBroker interface {
	Subscribe(topic BrokerTopicName) (Subscription, error)
	Unsubscribe(topic BrokerTopicName, subscription Subscription) error
	Publish(ctx context.Context, topic BrokerTopicName, msg BrokerMessage) error
	Stop()
}

type MessageHandler func(msg BrokerMessage)

var _ InternalBroker = (*LocalBroker)(nil)

var (
	ErrTopicNotFound       = errors.New("topic not found")
	ErrSubscriptorNotFound = errors.New("subscriptor not found")
)

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
	value, ok := b.subscriptors.Load(topic)
	var subscriptors []*subscriptor
	if !ok {
		subscriptors = make([]*subscriptor, 0, 1)
	} else {
		subscriptors, ok = value.([]*subscriptor)
		if !ok {
			return Subscription{}, ErrTopicNotFound
		}
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

	subscriptors, ok := value.([]*subscriptor)
	if !ok {
		return ErrTopicNotFound
	}
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
	subscriptors, ok := topicSubscriptors.([]*subscriptor)
	if !ok {
		return ErrTopicNotFound
	}

	go b.publishToSubscriptors(subscriptors, msg)

	return nil
}

func (b *LocalBroker) publishToSubscriptors(topicSubscriptors []*subscriptor, msg BrokerMessage) {
	for _, s := range topicSubscriptors {
		if s.active {
			s.subscription.Receiver <- msg
		}
	}
}

func (b *LocalBroker) Stop() {
	b.subscriptors.Range(func(key, value any) bool {
		if subscriptors, ok := value.([]*subscriptor); ok {
			for _, s := range subscriptors {
				s.safeClose()
			}
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

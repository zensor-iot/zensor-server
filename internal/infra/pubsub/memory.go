package pubsub

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// In-memory implementation for testing
type MemoryPublisherFactory struct {
	broker *MemoryBroker
}

func NewMemoryPublisherFactory() *MemoryPublisherFactory {
	return &MemoryPublisherFactory{
		broker: GetMemoryBroker(),
	}
}

func (f *MemoryPublisherFactory) New(topic Topic, prototype Message) (Publisher, error) {
	return &MemoryPublisher{
		broker:    f.broker,
		topic:     topic,
		prototype: prototype,
	}, nil
}

type MemoryPublisher struct {
	broker    *MemoryBroker
	topic     Topic
	prototype Message
}

func (p *MemoryPublisher) Publish(ctx context.Context, key Key, message Message) error {
	span := trace.SpanFromContext(ctx)
	slog.Debug("publishing message",
		slog.String("key", string(key)),
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()),
	)

	// For memory implementation, we don't wrap messages since there are no headers
	// but we still log trace information for debugging
	return p.broker.Publish(p.topic, key, message)
}

type MemoryConsumerFactory struct {
	broker *MemoryBroker
	group  string
}

func NewMemoryConsumerFactory(group string) *MemoryConsumerFactory {
	return &MemoryConsumerFactory{
		broker: GetMemoryBroker(),
		group:  group,
	}
}

func (f *MemoryConsumerFactory) New() Consumer {
	return &MemoryConsumer{
		broker: f.broker,
		group:  f.group,
	}
}

type MemoryConsumer struct {
	broker *MemoryBroker
	group  string
}

func (c *MemoryConsumer) Consume(topic Topic, handler MessageHandler, prototype Prototype) error {
	return c.broker.Subscribe(topic, c.group, handler, prototype)
}

// MemoryBroker is a singleton that manages all in-memory pubsub operations
type MemoryBroker struct {
	topics    map[Topic]*TopicChannel
	consumers map[string]*ConsumerInfo
	mu        sync.RWMutex
}

type TopicChannel struct {
	messages    chan MessageEvent
	subscribers map[string][]*ConsumerInfo
	mu          sync.RWMutex
}

type MessageEvent struct {
	Key     Key
	Message Message
	Topic   Topic
}

type ConsumerInfo struct {
	Group     string
	Handler   MessageHandler
	Prototype Prototype
	Topic     Topic
	Active    bool
}

var (
	memoryBroker     *MemoryBroker
	memoryBrokerOnce sync.Once
)

func GetMemoryBroker() *MemoryBroker {
	memoryBrokerOnce.Do(func() {
		memoryBroker = &MemoryBroker{
			topics:    make(map[Topic]*TopicChannel),
			consumers: make(map[string]*ConsumerInfo),
		}
	})
	return memoryBroker
}

func (b *MemoryBroker) Publish(topic Topic, key Key, message Message) error {
	slog.Debug("publish")
	b.mu.Lock()
	defer b.mu.Unlock()

	topicChan, exists := b.topics[topic]
	if !exists {
		topicChan = &TopicChannel{
			messages:    make(chan MessageEvent, 100), // Buffer for async processing
			subscribers: make(map[string][]*ConsumerInfo),
		}
		b.topics[topic] = topicChan
	}

	event := MessageEvent{
		Key:     key,
		Message: message,
		Topic:   topic,
	}

	// Send message to topic channel
	select {
	case topicChan.messages <- event:
		// Process subscribers
		go b.processSubscribers(topicChan, event)
	default:
		return fmt.Errorf("topic channel buffer full")
	}

	return nil
}

func (b *MemoryBroker) Subscribe(topic Topic, group string, handler MessageHandler, prototype Prototype) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	consumerID := fmt.Sprintf("%s-%s", group, string(topic))

	consumerInfo := &ConsumerInfo{
		Group:     group,
		Handler:   handler,
		Prototype: prototype,
		Topic:     topic,
		Active:    true,
	}

	b.consumers[consumerID] = consumerInfo

	topicChan, exists := b.topics[topic]
	if !exists {
		topicChan = &TopicChannel{
			messages:    make(chan MessageEvent, 100),
			subscribers: make(map[string][]*ConsumerInfo),
		}
		b.topics[topic] = topicChan
	}

	// Use the topic channel's mutex for subscriber operations
	topicChan.mu.Lock()
	topicChan.subscribers[group] = append(topicChan.subscribers[group], consumerInfo)
	topicChan.mu.Unlock()

	return nil
}

func (b *MemoryBroker) processSubscribers(topicChan *TopicChannel, event MessageEvent) {
	// Create a copy of subscribers to avoid holding the lock during processing
	topicChan.mu.RLock()
	subscribersCopy := make(map[string][]*ConsumerInfo)
	for group, consumers := range topicChan.subscribers {
		subscribersCopy[group] = make([]*ConsumerInfo, len(consumers))
		copy(subscribersCopy[group], consumers)
	}
	topicChan.mu.RUnlock()

	for _, consumers := range subscribersCopy {
		// In a real pubsub system, only one consumer per group would receive the message
		// For simplicity, we'll send to all consumers in the group
		for _, consumer := range consumers {
			if consumer.Active {
				go func(c *ConsumerInfo) {
					defer func() {
						if r := recover(); r != nil {
							// Log panic but don't crash
							fmt.Printf("panic in message handler: %v\n", r)
						}
					}()

					// Create a new context with a span for message processing
					ctx, span := CreateChildSpan(context.Background(), "memory.message.process",
						trace.WithAttributes(
							attribute.String("memory.topic", string(event.Topic)),
							attribute.String("memory.key", string(event.Key)),
						),
					)
					defer span.End()

					// Call the handler with context, key and message
					if err := c.Handler(ctx, event.Key, event.Message); err != nil {
						span.RecordError(err)
						fmt.Printf("error in message handler: %v\n", err)
					}
				}(consumer)
			}
		}
	}
}

// Reset clears all topics and consumers (useful for testing)
func (b *MemoryBroker) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.topics = make(map[Topic]*TopicChannel)
	b.consumers = make(map[string]*ConsumerInfo)
}

// GetMessageCount returns the number of messages in a topic's buffer
func (b *MemoryBroker) GetMessageCount(topic Topic) int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	topicChan, exists := b.topics[topic]
	if !exists {
		return 0
	}

	return len(topicChan.messages)
}

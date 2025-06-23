package pubsub

import (
	"context"
	"fmt"
	"sync"
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

	topicChan.mu.Lock()
	topicChan.subscribers[group] = append(topicChan.subscribers[group], consumerInfo)
	topicChan.mu.Unlock()

	return nil
}

func (b *MemoryBroker) processSubscribers(topicChan *TopicChannel, event MessageEvent) {
	topicChan.mu.RLock()
	subscribers := topicChan.subscribers
	topicChan.mu.RUnlock()

	for _, consumers := range subscribers {
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

					// Call the handler with the actual message
					if err := c.Handler(event.Message); err != nil {
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

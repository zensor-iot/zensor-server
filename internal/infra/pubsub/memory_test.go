package pubsub

import (
	"context"
	"testing"
	"time"
)

func TestMemoryPubSub(t *testing.T) {
	// Reset the broker for clean test state
	broker := GetMemoryBroker()
	broker.Reset()

	// Create factories
	publisherFactory := NewMemoryPublisherFactory()
	consumerFactory := NewMemoryConsumerFactory("test-group")

	// Create a publisher
	publisher, err := publisherFactory.New("test-topic", "test-message")
	if err != nil {
		t.Fatalf("Failed to create publisher: %v", err)
	}

	// Create a consumer
	consumer := consumerFactory.New()

	// Channel to receive messages
	messageReceived := make(chan bool, 1)
	var receivedMessage any

	// Message handler
	handler := func(_ context.Context, key Key, prototype Prototype) error {
		receivedMessage = prototype
		messageReceived <- true
		return nil
	}

	// Start consuming
	err = consumer.Consume("test-topic", handler, "test-message")
	if err != nil {
		t.Fatalf("Failed to start consumer: %v", err)
	}

	// Give some time for consumer to register
	time.Sleep(10 * time.Millisecond)

	// Publish a message
	testMessage := "hello world"
	err = publisher.Publish(context.Background(), "test-key", testMessage)
	if err != nil {
		t.Fatalf("Failed to publish message: %v", err)
	}

	// Wait for message to be received
	select {
	case <-messageReceived:
		// Message received successfully
		if receivedMessage != testMessage {
			t.Errorf("Expected message %v, got %v", testMessage, receivedMessage)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

func TestMemoryPubSubFactory(t *testing.T) {
	factory := NewFactory(FactoryOptions{
		Environment:       "local",
		KafkaBrokers:      []string{"localhost:9092"},
		ConsumerGroup:     "test-group",
		SchemaRegistryURL: "http://localhost:8081",
	})

	// Verify we get memory implementations
	publisherFactory := factory.GetPublisherFactory()
	consumerFactory := factory.GetConsumerFactory()

	_, ok := publisherFactory.(*MemoryPublisherFactory)
	if !ok {
		t.Error("Expected MemoryPublisherFactory when Environment=local")
	}

	_, ok = consumerFactory.(*MemoryConsumerFactory)
	if !ok {
		t.Error("Expected MemoryConsumerFactory when Environment=local")
	}
}

func TestMemoryPubSubNonLocal(t *testing.T) {
	factory := NewFactory(FactoryOptions{
		Environment:       "production",
		KafkaBrokers:      []string{"localhost:9092"},
		ConsumerGroup:     "test-group",
		SchemaRegistryURL: "http://localhost:8081",
	})

	// Verify we get Kafka implementations
	publisherFactory := factory.GetPublisherFactory()
	consumerFactory := factory.GetConsumerFactory()

	_, ok := publisherFactory.(*KafkaPublisherFactory)
	if !ok {
		t.Error("Expected KafkaPublisherFactory when Environment!=local")
	}

	_, ok = consumerFactory.(*KafkaConsumerFactory)
	if !ok {
		t.Error("Expected KafkaConsumerFactory when Environment!=local")
	}
}

package pubsub

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// Example of how to use the pubsub factory in tests
func ExampleFactory_usage() {
	// Create factory with in-memory implementation
	factory := NewFactory(FactoryOptions{
		Environment:   "local",
		KafkaBrokers:  []string{"localhost:9092"}, // Not used when Environment=local
		ConsumerGroup: "test-group",
	})

	// Get publisher factory
	publisherFactory := factory.GetPublisherFactory()

	// Create a publisher
	publisher, err := publisherFactory.New("test-topic", "test-message")
	if err != nil {
		fmt.Printf("Error creating publisher: %v\n", err)
		return
	}

	// Get consumer factory
	consumerFactory := factory.GetConsumerFactory()

	// Create a consumer
	consumer := consumerFactory.New()

	// Message received flag
	messageReceived := make(chan bool, 1)

	// Message handler
	handler := func(prototype Prototype) error {
		fmt.Printf("Received message: %v\n", prototype)
		messageReceived <- true
		return nil
	}

	// Start consuming
	err = consumer.Consume("test-topic", handler, "test-message")
	if err != nil {
		fmt.Printf("Error starting consumer: %v\n", err)
		return
	}

	// Give some time for consumer to register
	time.Sleep(10 * time.Millisecond)

	// Publish a message
	testMessage := "hello from example"
	err = publisher.Publish(context.Background(), "test-key", testMessage)
	if err != nil {
		fmt.Printf("Error publishing message: %v\n", err)
		return
	}

	// Wait for message to be received
	select {
	case <-messageReceived:
		fmt.Println("Message received successfully!")
	case <-time.After(1 * time.Second):
		fmt.Println("Timeout waiting for message")
	}

	// Output:
	// Received message: hello from example
	// Message received successfully!
}

// Example of how to use the factory in production (with Kafka)
func ExampleFactory_production() {
	// Create factory with Kafka implementation
	factory := NewFactory(FactoryOptions{
		Environment:   "production",
		KafkaBrokers:  []string{"kafka1:9092", "kafka2:9092"},
		ConsumerGroup: "production-group",
	})

	// Get publisher factory (will be Kafka)
	publisherFactory := factory.GetPublisherFactory()

	// Verify it's Kafka implementation
	if _, ok := publisherFactory.(*KafkaPublisherFactory); ok {
		fmt.Println("Using Kafka publisher factory")
	}

	// Get consumer factory (will be Kafka)
	consumerFactory := factory.GetConsumerFactory()

	// Verify it's Kafka implementation
	if _, ok := consumerFactory.(*KafkaConsumerFactory); ok {
		fmt.Println("Using Kafka consumer factory")
	}

	// Output:
	// Using Kafka publisher factory
	// Using Kafka consumer factory
}

// Test demonstrating the factory behavior
func TestFactoryBehavior(t *testing.T) {
	// Test local environment
	factory := NewFactory(FactoryOptions{
		Environment:   "local",
		KafkaBrokers:  []string{"localhost:9092"},
		ConsumerGroup: "test-group",
	})

	// Should be memory implementations
	if _, ok := factory.GetPublisherFactory().(*MemoryPublisherFactory); !ok {
		t.Error("Expected MemoryPublisherFactory when Environment=local")
	}

	if _, ok := factory.GetConsumerFactory().(*MemoryConsumerFactory); !ok {
		t.Error("Expected MemoryConsumerFactory when Environment=local")
	}

	// Test production environment
	factory = NewFactory(FactoryOptions{
		Environment:   "production",
		KafkaBrokers:  []string{"localhost:9092"},
		ConsumerGroup: "test-group",
	})

	// Should be Kafka implementations
	if _, ok := factory.GetPublisherFactory().(*KafkaPublisherFactory); !ok {
		t.Error("Expected KafkaPublisherFactory when Environment=production")
	}

	if _, ok := factory.GetConsumerFactory().(*KafkaConsumerFactory); !ok {
		t.Error("Expected KafkaConsumerFactory when Environment=production")
	}
}

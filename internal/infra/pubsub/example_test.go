package pubsub

import (
	"context"
	"fmt"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// Example of how to use the pubsub factory in tests
func ExampleFactory_usage() {
	// Create factory with in-memory implementation
	factory := NewFactory(FactoryOptions{
		Environment:       "local",
		KafkaBrokers:      []string{"localhost:9092"}, // Not used when Environment=local
		ConsumerGroup:     "test-group",
		SchemaRegistryURL: "http://localhost:8081", // Not used when Environment=local
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
	handler := func(_ context.Context, key Key, prototype Prototype) error {
		fmt.Printf("Received message: %v (key: %v)\n", prototype, key)
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
	// Received message: hello from example (key: test-key)
	// Message received successfully!
}

// Example of how to use the factory in production (with Kafka)
func ExampleFactory_production() {
	// Create factory with Kafka implementation
	factory := NewFactory(FactoryOptions{
		Environment:       "production",
		KafkaBrokers:      []string{"kafka1:9092", "kafka2:9092"},
		ConsumerGroup:     "production-group",
		SchemaRegistryURL: "http://schema-registry:8081",
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

var _ = ginkgo.Describe("Factory", func() {
	ginkgo.Context("FactoryBehavior", func() {
		var factory *Factory

		ginkgo.When("creating factory with local environment", func() {
			ginkgo.BeforeEach(func() {
				factory = NewFactory(FactoryOptions{
					Environment:       "local",
					KafkaBrokers:      []string{"localhost:9092"},
					ConsumerGroup:     "test-group",
					SchemaRegistryURL: "http://localhost:8081",
				})
			})

			ginkgo.It("should return memory implementations", func() {
				// Should be memory implementations
				_, ok := factory.GetPublisherFactory().(*MemoryPublisherFactory)
				gomega.Expect(ok).To(gomega.BeTrue())

				_, ok = factory.GetConsumerFactory().(*MemoryConsumerFactory)
				gomega.Expect(ok).To(gomega.BeTrue())
			})
		})

		ginkgo.When("creating factory with production environment", func() {
			ginkgo.BeforeEach(func() {
				factory = NewFactory(FactoryOptions{
					Environment:       "production",
					KafkaBrokers:      []string{"localhost:9092"},
					ConsumerGroup:     "test-group",
					SchemaRegistryURL: "http://localhost:8081",
				})
			})

			ginkgo.It("should return Kafka implementations", func() {
				// Should be Kafka implementations
				_, ok := factory.GetPublisherFactory().(*KafkaPublisherFactory)
				gomega.Expect(ok).To(gomega.BeTrue())

				_, ok = factory.GetConsumerFactory().(*KafkaConsumerFactory)
				gomega.Expect(ok).To(gomega.BeTrue())
			})
		})
	})
})

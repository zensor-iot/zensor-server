package pubsub_test

import (
	"context"
	"time"
	"zensor-server/internal/infra/pubsub"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("MemoryPubSub", func() {
	ginkgo.Context("MemoryPubSub", func() {
		var (
			broker           *pubsub.MemoryBroker
			publisherFactory pubsub.PublisherFactory
			consumerFactory  pubsub.ConsumerFactory
			publisher        pubsub.Publisher
			consumer         pubsub.Consumer
			messageReceived  chan bool
			receivedMessage  any
			testMessage      string
		)

		ginkgo.When("publishing and consuming messages", func() {
			ginkgo.BeforeEach(func() {
				// Reset the broker for clean test state
				broker = pubsub.GetMemoryBroker()
				broker.Reset()

				// Create factories
				publisherFactory = pubsub.NewMemoryPublisherFactory()
				consumerFactory = pubsub.NewMemoryConsumerFactory("test-group")

				// Create a publisher
				var err error
				publisher, err = publisherFactory.New("test-topic", "test-message")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Create a consumer
				consumer = consumerFactory.New()

				// Channel to receive messages
				messageReceived = make(chan bool, 1)
				testMessage = "hello world"
			})

			ginkgo.It("should successfully publish and consume messages", func() {
				// Message handler
				handler := func(_ context.Context, key pubsub.Key, prototype pubsub.Prototype) error {
					receivedMessage = prototype
					messageReceived <- true
					return nil
				}

				// Start consuming
				err := consumer.Consume("test-topic", handler, "test-message")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Give some time for consumer to register
				time.Sleep(10 * time.Millisecond)

				// Publish a message
				err = publisher.Publish(context.Background(), "test-key", testMessage)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Wait for message to be received
				select {
				case <-messageReceived:
					// Message received successfully
					gomega.Expect(receivedMessage).To(gomega.Equal(testMessage))
				case <-time.After(1 * time.Second):
					ginkgo.Fail("Timeout waiting for message")
				}
			})
		})
	})

	ginkgo.Context("MemoryPubSubFactory", func() {
		var factory *pubsub.Factory

		ginkgo.When("creating factory with local environment", func() {
			ginkgo.BeforeEach(func() {
				factory = pubsub.NewFactory(pubsub.FactoryOptions{
					Environment:       "local",
					KafkaBrokers:      []string{"localhost:9092"},
					ConsumerGroup:     "test-group",
					SchemaRegistryURL: "http://localhost:8081",
				})
			})

			ginkgo.It("should return memory implementations", func() {
				// Verify we get memory implementations
				publisherFactory := factory.GetPublisherFactory()
				consumerFactory := factory.GetConsumerFactory()

				_, ok := publisherFactory.(*pubsub.MemoryPublisherFactory)
				gomega.Expect(ok).To(gomega.BeTrue())

				_, ok = consumerFactory.(*pubsub.MemoryConsumerFactory)
				gomega.Expect(ok).To(gomega.BeTrue())
			})
		})
	})

	ginkgo.Context("MemoryPubSubNonLocal", func() {
		var factory *pubsub.Factory

		ginkgo.When("creating factory with non-local environment", func() {
			ginkgo.BeforeEach(func() {
				factory = pubsub.NewFactory(pubsub.FactoryOptions{
					Environment:       "production",
					KafkaBrokers:      []string{"localhost:9092"},
					ConsumerGroup:     "test-group",
					SchemaRegistryURL: "http://localhost:8081",
				})
			})

			ginkgo.It("should return Kafka implementations", func() {
				// Verify we get Kafka implementations
				publisherFactory := factory.GetPublisherFactory()
				consumerFactory := factory.GetConsumerFactory()

				_, ok := publisherFactory.(*pubsub.KafkaPublisherFactory)
				gomega.Expect(ok).To(gomega.BeTrue())

				_, ok = consumerFactory.(*pubsub.KafkaConsumerFactory)
				gomega.Expect(ok).To(gomega.BeTrue())
			})
		})
	})
})

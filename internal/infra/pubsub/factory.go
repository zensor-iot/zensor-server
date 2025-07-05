package pubsub

import "log/slog"

// Factory creates the appropriate pubsub implementation based on environment
type Factory struct {
	publisherFactory PublisherFactory
	consumerFactory  ConsumerFactory
}

// NewFactory creates a new factory with the appropriate implementation
func NewFactory(opts FactoryOptions) *Factory {
	if opts.Environment == "local" {
		slog.Debug("memory publisher initializing...")
		return &Factory{
			publisherFactory: NewMemoryPublisherFactory(),
			consumerFactory:  NewMemoryConsumerFactory(opts.ConsumerGroup),
		}
	}

	// Use Kafka for all other environments
	slog.Debug("kafka publisher initializing...")
	return &Factory{
		publisherFactory: NewKafkaPublisherFactory(KafkaPublisherFactoryOptions{
			Brokers:           opts.KafkaBrokers,
			SchemaRegistryURL: opts.SchemaRegistryURL,
		}),
		consumerFactory: NewKafkaConsumerFactory(opts.KafkaBrokers, opts.ConsumerGroup, opts.SchemaRegistryURL),
	}
}

type FactoryOptions struct {
	Environment       string
	KafkaBrokers      []string
	ConsumerGroup     string
	SchemaRegistryURL string
}

// GetPublisherFactory returns the configured publisher factory
func (f *Factory) GetPublisherFactory() PublisherFactory {
	return f.publisherFactory
}

// GetConsumerFactory returns the configured consumer factory
func (f *Factory) GetConsumerFactory() ConsumerFactory {
	return f.consumerFactory
}

// NewPublisher creates a new publisher for the given topic and prototype
func (f *Factory) NewPublisher(topic Topic, prototype Message) (Publisher, error) {
	return f.publisherFactory.New(topic, prototype)
}

// NewConsumer creates a new consumer
func (f *Factory) NewConsumer() Consumer {
	return f.consumerFactory.New()
}

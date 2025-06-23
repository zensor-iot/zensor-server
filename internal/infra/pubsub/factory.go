package pubsub

// Factory creates the appropriate pubsub implementation based on environment
type Factory struct {
	publisherFactory PublisherFactory
	consumerFactory  ConsumerFactory
}

// NewFactory creates a new factory with the appropriate implementation
func NewFactory(opts FactoryOptions) *Factory {
	if opts.Environment == "local" {
		return &Factory{
			publisherFactory: NewMemoryPublisherFactory(),
			consumerFactory:  NewMemoryConsumerFactory(opts.ConsumerGroup),
		}
	}

	// Use Kafka for all other environments
	return &Factory{
		publisherFactory: NewKafkaPublisherFactory(KafkaPublisherFactoryOptions{
			Brokers: opts.KafkaBrokers,
		}),
		consumerFactory: NewKafkaConsumerFactory(opts.KafkaBrokers, opts.ConsumerGroup),
	}
}

type FactoryOptions struct {
	Environment   string
	KafkaBrokers  []string
	ConsumerGroup string
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

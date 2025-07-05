package pubsub

import "fmt"

var _ PublisherFactory = (*KafkaPublisherFactory)(nil)

type KafkaPublisherFactoryOptions struct {
	Brokers           []string
	SchemaRegistryURL string
}

func NewKafkaPublisherFactory(opts KafkaPublisherFactoryOptions) *KafkaPublisherFactory {
	return &KafkaPublisherFactory{
		brokers:           opts.Brokers,
		schemaRegistryURL: opts.SchemaRegistryURL,
	}
}

type KafkaPublisherFactory struct {
	brokers           []string
	schemaRegistryURL string
}

func (f *KafkaPublisherFactory) New(topic Topic, prototype Message) (Publisher, error) {
	publisher, err := NewKafkaPublisher(f.brokers, string(topic), prototype, f.schemaRegistryURL)
	if err != nil {
		return nil, fmt.Errorf("creating publisher: %w", err)
	}

	return publisher, nil
}

var _ ConsumerFactory = (*KafkaConsumerFactory)(nil)

func NewKafkaConsumerFactory(brokers []string, group string, schemaRegistryURL string) *KafkaConsumerFactory {
	return &KafkaConsumerFactory{
		brokers:           brokers,
		group:             group,
		schemaRegistryURL: schemaRegistryURL,
	}
}

var _ ConsumerFactory = (*KafkaConsumerFactory)(nil)

type KafkaConsumerFactory struct {
	brokers           []string
	group             string
	schemaRegistryURL string
}

func (f *KafkaConsumerFactory) New() Consumer {
	consumer := NewKafkaConsumer(f.brokers, f.group, f.schemaRegistryURL)

	return consumer
}

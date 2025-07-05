package pubsub

import (
	"fmt"
	"zensor-server/internal/shared_kernel/avro"

	"github.com/riferrei/srclient"
)

var _ PublisherFactory = (*KafkaPublisherFactory)(nil)

type KafkaPublisherFactoryOptions struct {
	Brokers           []string
	SchemaRegistryURL string
}

func NewKafkaPublisherFactory(opts KafkaPublisherFactoryOptions) *KafkaPublisherFactory {
	schemaRegistry := srclient.CreateSchemaRegistryClient(opts.SchemaRegistryURL)
	return &KafkaPublisherFactory{
		brokers:        opts.Brokers,
		schemaRegistry: schemaRegistry,
	}
}

type KafkaPublisherFactory struct {
	brokers        []string
	schemaRegistry avro.SchemaRegistry
}

func (f *KafkaPublisherFactory) New(topic Topic, prototype Message) (Publisher, error) {
	publisher, err := NewKafkaPublisher(f.brokers, string(topic), prototype, f.schemaRegistry)
	if err != nil {
		return nil, fmt.Errorf("creating publisher: %w", err)
	}

	return publisher, nil
}

var _ ConsumerFactory = (*KafkaConsumerFactory)(nil)

func NewKafkaConsumerFactory(brokers []string, group string, schemaRegistryURL string) *KafkaConsumerFactory {
	schemaRegistry := srclient.CreateSchemaRegistryClient(schemaRegistryURL)
	return &KafkaConsumerFactory{
		brokers:        brokers,
		group:          group,
		schemaRegistry: schemaRegistry,
	}
}

var _ ConsumerFactory = (*KafkaConsumerFactory)(nil)

type KafkaConsumerFactory struct {
	brokers        []string
	group          string
	schemaRegistry avro.SchemaRegistry
}

func (f *KafkaConsumerFactory) New() Consumer {
	consumer := NewKafkaConsumer(f.brokers, f.group, f.schemaRegistry)

	return consumer
}

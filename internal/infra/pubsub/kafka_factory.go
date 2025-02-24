package pubsub

import "fmt"

var _ PublisherFactory = (*KafkaPublisherFactory)(nil)

type KafkaPublisherFactoryOptions struct {
	Brokers []string
}

func NewKafkaPublisherFactory(opts KafkaPublisherFactoryOptions) *KafkaPublisherFactory {
	return &KafkaPublisherFactory{
		brokers: opts.Brokers,
	}
}

type KafkaPublisherFactory struct {
	brokers []string
}

func (f *KafkaPublisherFactory) New(topic Topic, prototype Message) (Publisher, error) {
	publisher, err := NewKafkaPublisher(f.brokers, string(topic), prototype)
	if err != nil {
		return nil, fmt.Errorf("creating publisher: %w", err)
	}

	return publisher, nil
}

var _ ConsumerFactory = (*KafkaConsumerFactory)(nil)

func NewKafkaConsumerFactory(brokers []string, group string) *KafkaConsumerFactory {
	return &KafkaConsumerFactory{
		brokers,
		group,
	}
}

var _ ConsumerFactory = (*KafkaConsumerFactory)(nil)

type KafkaConsumerFactory struct {
	brokers []string
	group   string
}

func (f *KafkaConsumerFactory) New() Consumer {
	consumer := NewKafkaConsumer(f.brokers, f.group)

	return consumer
}

package kafka

import "fmt"

func NewKafkaPublisherFactory(brokers []string) *KafkaPublisherFactory {
	return &KafkaPublisherFactory{
		brokers,
	}
}

type KafkaPublisherFactory struct {
	brokers []string
}

func (f *KafkaPublisherFactory) CreatePublisher(topic string, prototype any) (KafkaPublisher, error) {
	publisher, err := NewKafkaPublisher(f.brokers, topic, prototype)
	if err != nil {
		return nil, fmt.Errorf("creating publisher: %w", err)
	}

	return publisher, nil
}

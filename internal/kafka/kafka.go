package kafka

import (
	"context"
	"fmt"
	"time"

	"zensor-server/internal/logger"

	"github.com/lovoo/goka"
	"github.com/lovoo/goka/codec"
)

const (
	maxRetries int = 10
)

type KafkaPublisher interface {
	Publish(string, string) error
}

type kafkaPublisherWrapper struct {
	emitter *goka.Emitter
}

type KafkaConsumer interface {
	Consume(func(string))
}

type kafkaConsumerWrapper struct {
	brokers []string
	topic   string
	group   string
}

func (p *kafkaPublisherWrapper) Publish(key, message string) error {
	err := p.emitter.EmitSync(key, message)
	if err != nil {
		logger.Info("error emitting message", "err", err)
		return err
	} else {
		logger.Info("emit sync done")
		return nil
	}
}

func (c *kafkaConsumerWrapper) Consume(fn func(string)) {
	cb := func(ctx goka.Context, msg interface{}) {
		logger.Info("message received", "msg", msg)
		fn(fmt.Sprintf("%v", msg))
	}

	group := goka.Group(c.group)
	topic := goka.Stream(c.topic)
	g := goka.DefineGroup(group,
		goka.Input(topic, new(codec.String), cb),
	)

	p, err := goka.NewProcessor(c.brokers, g)
	if err != nil {
		logger.Info("error creating processor", "err", err)
	}
	p.Run(context.Background())
}

func NewKafkaPublisher(brokers []string, topic string) (KafkaPublisher, error) {
	for try := 0; try < maxRetries; try++ {
		e, err := goka.NewEmitter(brokers, goka.Stream(topic), new(codec.String))

		if err != nil {
			time.Sleep(5 * time.Second)
		} else {
			return &kafkaPublisherWrapper{e}, nil
		}
	}

	return nil, fmt.Errorf("🤦‍♂️ imposible to connect to kafka brokers after %d retries", maxRetries)
}

func NewKafkaConsumer(brokers []string, topic string) KafkaConsumer {
	group := "zensor_server"
	return &kafkaConsumerWrapper{brokers, topic, group}
}

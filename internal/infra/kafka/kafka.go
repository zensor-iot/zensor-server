package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lovoo/goka"
	"github.com/lovoo/goka/codec"
)

const (
	maxRetries int = 10
)

type KafkaPublisher interface {
	Publish(context.Context, string, any) error
}

type SimpleKafkaPublisher struct {
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

func (p *SimpleKafkaPublisher) Publish(_ context.Context, key string, message any) error {
	err := p.emitter.EmitSync(key, message)
	if err != nil {
		slog.Info("error emitting message", "err", err)
		return err
	} else {
		slog.Info("emit sync done")
		return nil
	}
}

func (c *kafkaConsumerWrapper) Consume(fn func(string)) {
	cb := func(ctx goka.Context, msg interface{}) {
		slog.Info("message received", "msg", msg)
		fn(fmt.Sprintf("%v", msg))
	}

	group := goka.Group(c.group)
	topic := goka.Stream(c.topic)
	g := goka.DefineGroup(group,
		goka.Input(topic, new(codec.String), cb),
	)

	p, err := goka.NewProcessor(c.brokers, g)
	if err != nil {
		slog.Info("error creating processor", "err", err)
	}
	p.Run(context.Background())
}

func NewKafkaPublisher(brokers []string, topic string, prototype any) (*SimpleKafkaPublisher, error) {
	for try := 0; try < maxRetries; try++ {
		e, err := goka.NewEmitter(brokers, goka.Stream(topic), newJSONCodec(prototype))

		if err != nil {
			time.Sleep(5 * time.Second)
		} else {
			return &SimpleKafkaPublisher{e}, nil
		}
	}

	return nil, fmt.Errorf("🤦‍♂️ imposible to connect to kafka brokers after %d retries", maxRetries)
}

func NewKafkaConsumer(brokers []string, topic string) KafkaConsumer {
	group := "zensor_server"
	return &kafkaConsumerWrapper{brokers, topic, group}
}

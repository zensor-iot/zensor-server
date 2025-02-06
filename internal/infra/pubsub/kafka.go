package pubsub

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lovoo/goka"
)

const (
	maxRetries int = 10
)

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

type SimpleKafkaPublisher struct {
	emitter *goka.Emitter
}

func (p *SimpleKafkaPublisher) Publish(_ context.Context, key Key, message Message) error {
	err := p.emitter.EmitSync(string(key), message)
	if err != nil {
		slog.Error("emitting message", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func NewKafkaConsumer(brokers []string, group string) *SimpleKafkaConsumer {
	gokaGroup := goka.Group(group)
	return &SimpleKafkaConsumer{
		brokers: brokers,
		group:   gokaGroup,
	}
}

var _ Consumer = (*SimpleKafkaConsumer)(nil)

type SimpleKafkaConsumer struct {
	brokers []string
	group   goka.Group
}

func (c *SimpleKafkaConsumer) Consume(topic Topic, handler MessageHandler, prototype Prototype) error {
	cb := func(ctx goka.Context, msg any) {
		slog.Debug("message received", slog.Any("msg", msg))
		handler(msg)
	}
	stream := goka.Stream(topic)
	gg := goka.DefineGroup(
		c.group,
		goka.Input(stream, newJSONCodec(prototype), cb),
	)
	p, err := goka.NewProcessor(c.brokers, gg)
	if err != nil {
		return nil
	}

	p.Run(context.Background())
	return nil
}

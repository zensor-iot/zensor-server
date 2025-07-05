package pubsub

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"zensor-server/internal/shared_kernel/avro"

	"github.com/lovoo/goka"
)

const (
	maxRetries int = 10
)

// publisherKey represents a unique key for a publisher instance
type publisherKey struct {
	brokers           string
	topic             string
	prototypeType     string
	schemaRegistryURL string
}

// publisherInstance holds a publisher and its initialization state
type publisherInstance struct {
	publisher *SimpleKafkaPublisher
	once      sync.Once
	err       error
}

// publishersMap stores singleton instances of publishers
var (
	publishersMap   = make(map[publisherKey]*publisherInstance)
	publishersMutex sync.RWMutex
)

func NewKafkaPublisher(brokers []string, topic string, prototype any, schemaRegistryURL string) (*SimpleKafkaPublisher, error) {
	// Create a unique key for this publisher configuration
	key := publisherKey{
		brokers:           strings.Join(brokers, ","),
		topic:             topic,
		prototypeType:     fmt.Sprintf("%T", prototype),
		schemaRegistryURL: schemaRegistryURL,
	}

	// Get or create the publisher instance
	publishersMutex.Lock()
	instance, exists := publishersMap[key]
	if !exists {
		instance = &publisherInstance{}
		publishersMap[key] = instance
	}
	publishersMutex.Unlock()

	// Initialize the publisher exactly once
	instance.once.Do(func() {
		slog.Debug("creating kafka publisher",
			slog.String("schemaRegistryURL", schemaRegistryURL),
			slog.String("topic", topic),
			slog.String("prototypeType", key.prototypeType))

		for try := 0; try < maxRetries; try++ {
			slog.Debug("connecting to kafka brokers", slog.String("brokers", key.brokers))
			codec, err := avro.NewConfluentAvroCodec(prototype, schemaRegistryURL)
			if err != nil {
				instance.err = fmt.Errorf("failed to create Avro codec: %w", err)
				return
			}
			e, err := goka.NewEmitter(brokers, goka.Stream(topic), codec)

			if err != nil {
				time.Sleep(5 * time.Second)
			} else {
				instance.publisher = &SimpleKafkaPublisher{e}
				return
			}
		}

		instance.err = fmt.Errorf("🤦‍♂️ imposible to connect to kafka brokers after %d retries", maxRetries)
	})

	if instance.err != nil {
		return nil, instance.err
	}

	return instance.publisher, nil
}

type SimpleKafkaPublisher struct {
	emitter *goka.Emitter
}

func (p *SimpleKafkaPublisher) Publish(_ context.Context, key Key, message Message) error {
	slog.Debug("publishing message", slog.String("key", string(key)))
	err := p.emitter.EmitSync(string(key), message)
	if err != nil {
		slog.Error("emitting message", slog.String("error", err.Error()))
		return err
	}

	return nil
}

// consumerKey represents a unique key for a consumer instance
type consumerKey struct {
	brokers           string
	group             string
	schemaRegistryURL string
}

// consumerInstance holds a consumer and its initialization state
type consumerInstance struct {
	consumer *SimpleKafkaConsumer
	once     sync.Once
	err      error
}

// consumersMap stores singleton instances of consumers
var (
	consumersMap   = make(map[consumerKey]*consumerInstance)
	consumersMutex sync.RWMutex
)

func NewKafkaConsumer(brokers []string, group string, schemaRegistryURL string) *SimpleKafkaConsumer {
	// Create a unique key for this consumer configuration
	key := consumerKey{
		brokers:           strings.Join(brokers, ","),
		group:             group,
		schemaRegistryURL: schemaRegistryURL,
	}

	// Get or create the consumer instance
	consumersMutex.Lock()
	instance, exists := consumersMap[key]
	if !exists {
		instance = &consumerInstance{}
		consumersMap[key] = instance
	}
	consumersMutex.Unlock()

	// Initialize the consumer exactly once
	instance.once.Do(func() {
		slog.Debug("creating kafka consumer",
			slog.String("schemaRegistryURL", schemaRegistryURL),
			slog.String("group", group),
			slog.String("brokers", key.brokers))

		gokaGroup := goka.Group(group)
		instance.consumer = &SimpleKafkaConsumer{
			brokers:           brokers,
			group:             gokaGroup,
			schemaRegistryURL: schemaRegistryURL,
		}
	})

	return instance.consumer
}

var _ Consumer = (*SimpleKafkaConsumer)(nil)

type SimpleKafkaConsumer struct {
	brokers           []string
	group             goka.Group
	schemaRegistryURL string
}

func (c *SimpleKafkaConsumer) Consume(topic Topic, handler MessageHandler, prototype Prototype) error {
	cb := func(ctx goka.Context, msg any) {
		slog.Debug("message received", slog.Any("msg", msg))
		key := Key(ctx.Key())
		handler(key, msg)
	}
	stream := goka.Stream(topic)
	codec, err := avro.NewConfluentAvroCodec(prototype, c.schemaRegistryURL)
	if err != nil {
		return fmt.Errorf("failed to create Avro codec: %w", err)
	}
	gg := goka.DefineGroup(
		c.group,
		goka.Input(stream, codec, cb),
	)
	p, err := goka.NewProcessor(c.brokers, gg)
	if err != nil {
		return nil
	}

	p.Run(context.Background())
	return nil
}

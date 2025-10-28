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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	maxRetries int = 10
)

type publisherKey struct {
	brokers        string
	topic          string
	prototypeType  string
	schemaRegistry avro.SchemaRegistry
}

type publisherInstance struct {
	publisher *SimpleKafkaPublisher
	once      sync.Once
	err       error
}

var (
	publishersMap   = make(map[publisherKey]*publisherInstance)
	publishersMutex sync.RWMutex
)

func NewKafkaPublisher(brokers []string, topic string, prototype any, schemaRegistry avro.SchemaRegistry) (*SimpleKafkaPublisher, error) {
	key := publisherKey{
		brokers:        strings.Join(brokers, ","),
		topic:          topic,
		prototypeType:  fmt.Sprintf("%T", prototype),
		schemaRegistry: schemaRegistry,
	}

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
			slog.String("topic", topic),
			slog.String("prototypeType", key.prototypeType))

		for try := 0; try < maxRetries; try++ {
			slog.Debug("connecting to kafka brokers", slog.String("brokers", key.brokers))

			codec := avro.NewConfluentAvroCodec(prototype, schemaRegistry)

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

func (p *SimpleKafkaPublisher) Publish(ctx context.Context, key Key, message Message) error {
	tracer := otel.Tracer("zensor-server")
	ctx, span := tracer.Start(ctx, "kafka.publish",
		trace.WithAttributes(
			attribute.String("span.kind", "client"),
			attribute.String("component", "kafka-publisher"),
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination", "kafka-topic"),
			attribute.String("messaging.destination_kind", "topic"),
			attribute.String("messaging.message_key", string(key)),
		),
	)
	defer span.End()

	slog.Debug("publishing message",
		slog.String("key", string(key)),
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()),
	)

	traceHeaders := ExtractTraceFromContext(ctx)
	kafkaHeaders := SerializeTraceHeaders(traceHeaders)

	err := p.emitter.EmitSyncWithHeaders(string(key), message, kafkaHeaders)
	if err != nil {
		span.RecordError(err)
		slog.Error("emitting message",
			slog.String("error", err.Error()),
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
		return err
	}

	return nil
}

type consumerKey struct {
	brokers        string
	group          string
	schemaRegistry avro.SchemaRegistry
}

type consumerInstance struct {
	consumer *SimpleKafkaConsumer
	once     sync.Once
}

var (
	consumersMap   = make(map[consumerKey]*consumerInstance)
	consumersMutex sync.RWMutex
)

func NewKafkaConsumer(brokers []string, group string, schemaRegistry avro.SchemaRegistry) *SimpleKafkaConsumer {
	key := consumerKey{
		brokers:        strings.Join(brokers, ","),
		group:          group,
		schemaRegistry: schemaRegistry,
	}

	consumersMutex.Lock()
	instance, exists := consumersMap[key]
	if !exists {
		instance = &consumerInstance{}
		consumersMap[key] = instance
	}
	consumersMutex.Unlock()

	instance.once.Do(func() {
		slog.Debug("creating kafka consumer",
			slog.String("group", group),
			slog.String("brokers", key.brokers))

		gokaGroup := goka.Group(group)
		instance.consumer = &SimpleKafkaConsumer{
			brokers:        brokers,
			group:          gokaGroup,
			schemaRegistry: schemaRegistry,
		}
	})

	return instance.consumer
}

var _ Consumer = (*SimpleKafkaConsumer)(nil)

type SimpleKafkaConsumer struct {
	brokers        []string
	group          goka.Group
	schemaRegistry avro.SchemaRegistry
}

func (c *SimpleKafkaConsumer) Consume(topic Topic, handler MessageHandler, prototype Prototype) error {
	cb := func(ctx goka.Context, msg any) {
		key := Key(ctx.Key())

		traceCtx := ExtractTraceFromKafkaHeaders(context.Background(), ctx.Headers())

		spanCtx, span := CreateChildSpan(traceCtx, "kafka.message.process",
			trace.WithAttributes(
				attribute.String("kafka.topic", string(topic)),
				attribute.String("kafka.key", string(key)),
			),
		)
		defer span.End()

		if err := handler(spanCtx, key, msg); err != nil {
			span.RecordError(err)
			slog.Error("error handling message",
				slog.String("error", err.Error()),
				slog.String("key", string(key)),
				slog.String("trace_id", span.SpanContext().TraceID().String()),
				slog.String("span_id", span.SpanContext().SpanID().String()),
			)
		}
	}
	stream := goka.Stream(topic)

	codec := avro.NewConfluentAvroCodec(prototype, c.schemaRegistry)

	gg := goka.DefineGroup(
		c.group,
		goka.Input(stream, codec, cb),
	)
	p, err := goka.NewProcessor(c.brokers, gg)
	if err != nil {
		return nil
	}

	err = p.Run(context.Background())
	if err != nil {
		return fmt.Errorf("run kafka consumer: %w", err)
	}
	return nil
}

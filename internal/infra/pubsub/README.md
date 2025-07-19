# PubSub Package

This package provides a unified interface for publish-subscribe messaging with support for both Kafka and in-memory implementations.

## Features

- **Unified Interface**: Common interfaces for publishers and consumers
- **Multiple Backends**: Support for Kafka (via Goka) and in-memory messaging
- **OpenTelemetry Integration**: Automatic trace propagation through Kafka headers
- **Schema Support**: Avro schema registry integration for Kafka
- **Factory Pattern**: Easy creation of publishers and consumers

## OpenTelemetry Trace Propagation

The package automatically propagates OpenTelemetry trace context through Kafka messages using Kafka headers. This enables distributed tracing across services that communicate via Kafka.

### How it Works

1. **Publishing**: When a message is published, the current trace context is extracted and serialized into Kafka headers
2. **Consuming**: When a message is consumed, the trace context is extracted from Kafka headers and injected into the consumer context
3. **Child Spans**: Each message processing creates a child span for observability

### Usage

#### Publishing with Trace Context

```go
// Create a context with a span
ctx, span := tracer.Start(context.Background(), "publish.message")
defer span.End()

// Publish message - trace context is automatically included in Kafka headers
err := publisher.Publish(ctx, "message-key", message)
```

#### Consuming with Trace Context

```go
// The handler receives a context with trace context already injected
handler := func(ctx context.Context, key pubsub.Key, message pubsub.Prototype) error {
    // Create a child span for message processing
    spanCtx, span := pubsub.CreateChildSpan(ctx, "process.message")
    defer span.End()
    
    // Add custom attributes
    span.SetAttributes(attribute.String("message.key", string(key)))
    
    // Process the message
    return processMessage(spanCtx, message)
}

// Start consuming - trace context is automatically extracted from headers
err := consumer.Consume("topic-name", handler, &MyMessage{})
```

### Implementation Details

#### Kafka Implementation

- Uses `EmitSyncWithHeaders` to include trace context in Kafka headers
- Extracts trace context from `Context.Headers()` in consumer callbacks
- Headers include: `trace_id`, `span_id`, `trace_flags`

#### Memory Implementation

- Since memory doesn't support headers, trace context is logged for debugging
- Each message processing creates a new span for observability
- No trace propagation between memory publishers/consumers

### Trace Headers Format

```go
type TraceHeaders struct {
    TraceID    string `json:"trace_id"`    // OpenTelemetry trace ID
    SpanID     string `json:"span_id"`     // OpenTelemetry span ID  
    TraceFlags string `json:"trace_flags"` // OpenTelemetry trace flags
}
```

### Error Handling

- Invalid trace headers are gracefully ignored
- Missing trace context doesn't prevent message processing
- Errors during message processing are recorded on the span

## Basic Usage

### Creating Publishers

```go
// Kafka publisher
kafkaFactory := pubsub.NewKafkaPublisherFactory(brokers, schemaRegistry)
publisher, err := kafkaFactory.New("my-topic", &MyMessage{})

// Memory publisher (for testing)
memoryFactory := pubsub.NewMemoryPublisherFactory()
publisher, err := memoryFactory.New("my-topic", &MyMessage{})
```

### Creating Consumers

```go
// Kafka consumer
kafkaFactory := pubsub.NewKafkaConsumerFactory(brokers, "my-group", schemaRegistry)
consumer := kafkaFactory.New()

// Memory consumer (for testing)
memoryFactory := pubsub.NewMemoryConsumerFactory("my-group")
consumer := memoryFactory.New()
```

### Message Handler

```go
func handleMessage(ctx context.Context, key pubsub.Key, message pubsub.Prototype) error {
    // Cast to your message type
    myMsg := message.(*MyMessage)
    
    // Process the message
    return processMyMessage(ctx, myMsg)
}
```

## Testing

The package includes comprehensive tests for both Kafka and memory implementations:

```bash
# Run all tests
go test ./internal/infra/pubsub -v

# Run specific test
go test ./internal/infra/pubsub -run TestTracePropagation
```

## Configuration

### Kafka Configuration

- **Brokers**: List of Kafka broker addresses
- **Schema Registry**: Avro schema registry for message serialization
- **Consumer Groups**: For load balancing and fault tolerance

### Memory Configuration

- **Groups**: For simulating consumer group behavior
- **Buffers**: Configurable message buffers for testing

## Best Practices

1. **Always use context**: Pass context through your application for proper trace propagation
2. **Create child spans**: Use `CreateChildSpan` for message processing to maintain trace hierarchy
3. **Add attributes**: Include relevant business data as span attributes for better observability
4. **Handle errors**: Record errors on spans using `span.RecordError(err)`
5. **Test with memory**: Use memory implementation for unit tests to avoid Kafka dependencies 
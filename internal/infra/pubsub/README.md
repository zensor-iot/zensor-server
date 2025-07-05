# PubSub Infrastructure

This package provides a pubsub infrastructure with support for both Kafka and in-memory implementations.

## Overview

The pubsub package includes:
- **Kafka implementation**: For production use with real Kafka brokers
- **In-memory implementation**: For testing and local development without external dependencies
- **Factory pattern**: Switches between implementations based on environment parameter

## Usage

### Basic Usage

```go
import "zensor-server/internal/infra/pubsub"

// Create a factory
factory := pubsub.NewFactory(pubsub.FactoryOptions{
    Environment:       "local", // or "production"
    KafkaBrokers:      []string{"localhost:9092"},
    ConsumerGroup:     "my-app",
    SchemaRegistryURL: "http://localhost:8081", // Required for Kafka environments
})

// Get publisher factory
publisherFactory := factory.GetPublisherFactory()

// Create a publisher
publisher, err := publisherFactory.New("my-topic", "message-type")
if err != nil {
    // handle error
}

// Publish a message
err = publisher.Publish(context.Background(), "key", "message")
if err != nil {
    // handle error
}

// Get consumer factory
consumerFactory := factory.GetConsumerFactory()

// Create a consumer
consumer := consumerFactory.New()

// Define message handler
handler := func(prototype pubsub.Prototype) error {
    // Process the message
    fmt.Printf("Received: %v\n", prototype)
    return nil
}

// Start consuming
err = consumer.Consume("my-topic", handler, "message-type")
if err != nil {
    // handle error
}
```

### Environment-Based Implementation Selection

The factory selects the appropriate implementation based on the `Environment` parameter:

- **`"local"`**: Uses in-memory implementation (no external dependencies)
- **Any other value**: Uses Kafka implementation

### Testing with In-Memory Implementation

The in-memory implementation is perfect for testing as it:
- Has no external dependencies
- Provides immediate message delivery
- Supports multiple consumers per topic
- Includes helper methods for testing

```go
func TestMyPubSub(t *testing.T) {
    factory := pubsub.NewFactory(pubsub.FactoryOptions{
        Environment:       "local",
        KafkaBrokers:      []string{"localhost:9092"}, // Not used when Environment=local
        ConsumerGroup:     "test-group",
        SchemaRegistryURL: "http://localhost:8081", // Not used when Environment=local
    })
    
    // Use factory as normal - it will use in-memory implementation
    publisherFactory := factory.GetPublisherFactory()
    consumerFactory := factory.GetConsumerFactory()
    
    // ... rest of test
}
```

### In-Memory Implementation Features

The in-memory implementation provides several features useful for testing:

#### Message Buffering
Messages are buffered in channels with a capacity of 100 messages per topic.

#### Concurrent Processing
Messages are processed concurrently using goroutines.

#### Consumer Groups
Supports consumer groups similar to Kafka (though all consumers in a group receive messages for simplicity).

#### Testing Helpers

```go
// Get the memory broker for testing
broker := pubsub.GetMemoryBroker()

// Reset all topics and consumers
broker.Reset()

// Get message count for a topic
count := broker.GetMessageCount("my-topic")
```

### Wire Integration

To use the factory with Google Wire dependency injection:

```go
// In your wire configuration
func providePubSubFactory(config config.AppConfig) *pubsub.Factory {
    env := os.Getenv("ENV")
    if env == "" {
        env = "production" // default to production if not set
    }
    
    return pubsub.NewFactory(pubsub.FactoryOptions{
        Environment:      env,
        KafkaBrokers:     config.Kafka.Brokers,
        ConsumerGroup:    "zensor-server",
        SchemaRegistryURL: config.Kafka.SchemaRegistry, // <-- added
    })
}

func providePublisherFactory(factory *pubsub.Factory) pubsub.PublisherFactory {
    return factory.GetPublisherFactory()
}

// In your wire.Build
wire.Build(
    providePubSubFactory,
    providePublisherFactory,
    // ... other dependencies
)
```

## Architecture

### Interfaces

- `PublisherFactory`: Creates publishers for specific topics
- `Publisher`: Publishes messages to topics
- `ConsumerFactory`: Creates consumers
- `Consumer`: Consumes messages from topics

### Implementations

#### Kafka Implementation
- Uses the Goka library for Kafka integration
- Supports retry logic for connection failures
- Provides JSON encoding/decoding

#### In-Memory Implementation
- Uses Go channels for message passing
- Singleton broker pattern for shared state
- Thread-safe operations with mutexes
- Async message processing

### Message Flow

1. **Publisher**: Creates a publisher for a specific topic and message type
2. **Publish**: Sends a message with a key to the topic
3. **Broker**: Routes the message to all subscribers of the topic
4. **Consumer**: Receives the message and calls the handler function
5. **Handler**: Processes the message according to business logic

## Error Handling

The in-memory implementation includes robust error handling:
- Panic recovery in message handlers
- Buffer overflow protection
- Graceful shutdown support

## Performance Considerations

- In-memory implementation is suitable for testing and development
- For production, use Kafka implementation for scalability and persistence
- Message handlers should be non-blocking for best performance
- Consider using timeouts for long-running operations 
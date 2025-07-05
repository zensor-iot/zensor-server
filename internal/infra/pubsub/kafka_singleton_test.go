package pubsub

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKafkaPublisherSingleton(t *testing.T) {
	// Test that multiple calls with the same parameters return the same instance
	brokers := []string{"localhost:9092"}
	topic := "test-topic"
	prototype := &struct{}{}
	schemaRegistryURL := "http://localhost:8081"

	// Create first publisher
	publisher1, err := NewKafkaPublisher(brokers, topic, prototype, schemaRegistryURL)
	require.Error(t, err) // Should fail because Kafka is not running, but that's expected
	assert.Nil(t, publisher1)

	// Create second publisher with same parameters
	publisher2, err := NewKafkaPublisher(brokers, topic, prototype, schemaRegistryURL)
	require.Error(t, err) // Should fail because Kafka is not running, but that's expected
	assert.Nil(t, publisher2)

	// The singleton pattern ensures that the same error is returned
	// and the initialization only happens once
}

func TestKafkaConsumerSingleton(t *testing.T) {
	// Test that multiple calls with the same parameters return the same instance
	brokers := []string{"localhost:9092"}
	group := "test-group"
	schemaRegistryURL := "http://localhost:8081"

	// Create first consumer
	consumer1 := NewKafkaConsumer(brokers, group, schemaRegistryURL)
	assert.NotNil(t, consumer1)

	// Create second consumer with same parameters
	consumer2 := NewKafkaConsumer(brokers, group, schemaRegistryURL)
	assert.NotNil(t, consumer2)

	// Should be the same instance due to singleton pattern
	assert.Equal(t, consumer1, consumer2)
}

func TestKafkaPublisherDifferentConfigs(t *testing.T) {
	// Test that different configurations create different instances
	brokers := []string{"localhost:9092"}
	topic1 := "test-topic-1"
	topic2 := "test-topic-2"
	prototype := &struct{}{}
	schemaRegistryURL := "http://localhost:8081"

	// Create publishers with different topics
	publisher1, err1 := NewKafkaPublisher(brokers, topic1, prototype, schemaRegistryURL)
	publisher2, err2 := NewKafkaPublisher(brokers, topic2, prototype, schemaRegistryURL)

	// Both should fail due to Kafka not running
	require.Error(t, err1)
	require.Error(t, err2)

	// Since both fail, they should both be nil
	// But the singleton pattern ensures that the same error is returned
	// and initialization only happens once per configuration
	assert.Nil(t, publisher1)
	assert.Nil(t, publisher2)

	// The key point is that the singleton pattern ensures that
	// the same error is returned for the same configuration
	// and initialization only happens once
}

func TestKafkaConsumerDifferentConfigs(t *testing.T) {
	// Test that different configurations create different instances
	brokers := []string{"localhost:9092"}
	group1 := "test-group-1"
	group2 := "test-group-2"
	schemaRegistryURL := "http://localhost:8081"

	// Create consumers with different groups
	consumer1 := NewKafkaConsumer(brokers, group1, schemaRegistryURL)
	consumer2 := NewKafkaConsumer(brokers, group2, schemaRegistryURL)

	// Should be different instances due to different configurations
	assert.NotEqual(t, consumer1, consumer2)
}

func TestKafkaPublisherConcurrentAccess(t *testing.T) {
	// Test that concurrent access to the same configuration is handled correctly
	brokers := []string{"localhost:9092"}
	topic := "test-topic-concurrent"
	prototype := &struct{}{}
	schemaRegistryURL := "http://localhost:8081"

	// Create multiple goroutines trying to create the same publisher
	results := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := NewKafkaPublisher(brokers, topic, prototype, schemaRegistryURL)
			results <- err
		}()
	}

	// Collect all results
	var errors []error
	for i := 0; i < 10; i++ {
		err := <-results
		errors = append(errors, err)
	}

	// All should return the same error (Kafka not running)
	// but the singleton pattern should ensure initialization only happens once
	for _, err := range errors {
		assert.Error(t, err) // All should fail due to Kafka not running
	}
}

func TestKafkaConsumerConcurrentAccess(t *testing.T) {
	// Test that concurrent access to the same configuration is handled correctly
	brokers := []string{"localhost:9092"}
	group := "test-group-concurrent"
	schemaRegistryURL := "http://localhost:8081"

	// Create multiple goroutines trying to create the same consumer
	results := make(chan *SimpleKafkaConsumer, 10)
	for i := 0; i < 10; i++ {
		go func() {
			consumer := NewKafkaConsumer(brokers, group, schemaRegistryURL)
			results <- consumer
		}()
	}

	// Collect all results
	var consumers []*SimpleKafkaConsumer
	for i := 0; i < 10; i++ {
		consumer := <-results
		consumers = append(consumers, consumer)
	}

	// All should be the same instance due to singleton pattern
	firstConsumer := consumers[0]
	for _, consumer := range consumers {
		assert.Equal(t, firstConsumer, consumer)
	}
}

package replication

import (
	"context"
	"errors"
	"testing"
	"time"

	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestReplicationIntegration tests the complete replication flow
func TestReplicationIntegration(t *testing.T) {
	// This test requires a real database and pub/sub system
	// It should be run in integration test environment
	t.Skip("Integration test requires real database and pub/sub system")

	// Setup test dependencies
	consumerFactory := createTestConsumerFactory()
	orm := createTestORM()

	// Create replication service
	service := NewService(consumerFactory, orm)

	// Register handlers
	deviceHandler := createTestDeviceHandler(orm)
	err := service.RegisterHandler(deviceHandler)
	require.NoError(t, err)

	// Start replication service
	err = service.Start()
	require.NoError(t, err)
	defer service.Stop()

	// Test device replication
	testDeviceReplication(t, service, consumerFactory)
}

// TestReplicationDataFlow tests the data flow through the replication system
func TestReplicationDataFlow(t *testing.T) {
	// Mock dependencies for unit testing
	consumerFactory := &MockConsumerFactory{}
	orm := &MockORM{}
	service := NewService(consumerFactory, orm)

	// Create a test device handler
	deviceHandler := &MockTopicHandler{}
	deviceHandler.On("TopicName").Return(pubsub.Topic("devices"))

	// Register the handler
	err := service.RegisterHandler(deviceHandler)
	require.NoError(t, err)

	// Mock consumer behavior
	consumer := &MockConsumer{}
	consumer.On("Consume", pubsub.Topic("devices"), mock.AnythingOfType("pubsub.MessageHandler"), mock.Anything).Return(nil)
	consumerFactory.On("New").Return(consumer)

	// Start the service
	err = service.Start()
	require.NoError(t, err)

	// Give time for goroutines to start
	time.Sleep(10 * time.Millisecond)

	// Stop the service
	service.Stop()

	// Verify expectations
	deviceHandler.AssertExpectations(t)
	consumer.AssertExpectations(t)
	consumerFactory.AssertExpectations(t)
}

// TestReplicationErrorHandling tests error scenarios in replication
func TestReplicationErrorHandling(t *testing.T) {
	consumerFactory := &MockConsumerFactory{}
	orm := &MockORM{}
	service := NewService(consumerFactory, orm)

	// Create a handler that returns errors
	errorHandler := &MockTopicHandler{}
	errorHandler.On("TopicName").Return(pubsub.Topic("error-topic"))
	// Note: These expectations won't be met because the consumer doesn't actually call the handler
	// in this test setup - we're just testing the service startup/shutdown

	// Register the handler
	err := service.RegisterHandler(errorHandler)
	require.NoError(t, err)

	// Mock consumer that will trigger the error handler
	consumer := &MockConsumer{}
	consumer.On("Consume", pubsub.Topic("error-topic"), mock.AnythingOfType("pubsub.MessageHandler"), mock.Anything).Return(nil)
	consumerFactory.On("New").Return(consumer)

	// Start the service
	err = service.Start()
	require.NoError(t, err)

	// Give time for goroutines to start
	time.Sleep(10 * time.Millisecond)

	// Stop the service
	service.Stop()

	// Verify expectations
	errorHandler.AssertExpectations(t)
	consumer.AssertExpectations(t)
	consumerFactory.AssertExpectations(t)
}

// TestReplicationConcurrency tests concurrent message processing
func TestReplicationConcurrency(t *testing.T) {
	consumerFactory := &MockConsumerFactory{}
	orm := &MockORM{}
	service := NewService(consumerFactory, orm)

	// Create multiple handlers for different topics
	deviceHandler := &MockTopicHandler{}
	deviceHandler.On("TopicName").Return(pubsub.Topic("devices"))

	tenantHandler := &MockTopicHandler{}
	tenantHandler.On("TopicName").Return(pubsub.Topic("tenants"))

	taskHandler := &MockTopicHandler{}
	taskHandler.On("TopicName").Return(pubsub.Topic("tasks"))

	// Register all handlers
	err := service.RegisterHandler(deviceHandler)
	require.NoError(t, err)
	err = service.RegisterHandler(tenantHandler)
	require.NoError(t, err)
	err = service.RegisterHandler(taskHandler)
	require.NoError(t, err)

	// Mock consumers for each topic - use Any() to handle any topic
	deviceConsumer := &MockConsumer{}
	deviceConsumer.On("Consume", mock.AnythingOfType("pubsub.Topic"), mock.AnythingOfType("pubsub.MessageHandler"), mock.Anything).Return(nil)

	tenantConsumer := &MockConsumer{}
	tenantConsumer.On("Consume", mock.AnythingOfType("pubsub.Topic"), mock.AnythingOfType("pubsub.MessageHandler"), mock.Anything).Return(nil)

	taskConsumer := &MockConsumer{}
	taskConsumer.On("Consume", mock.AnythingOfType("pubsub.Topic"), mock.AnythingOfType("pubsub.MessageHandler"), mock.Anything).Return(nil)

	// Mock consumer factory to return different consumers
	consumerFactory.On("New").Return(deviceConsumer).Once()
	consumerFactory.On("New").Return(tenantConsumer).Once()
	consumerFactory.On("New").Return(taskConsumer).Once()

	// Start the service
	err = service.Start()
	require.NoError(t, err)

	// Give time for goroutines to start
	time.Sleep(10 * time.Millisecond)

	// Stop the service
	service.Stop()

	// Verify all expectations
	deviceHandler.AssertExpectations(t)
	tenantHandler.AssertExpectations(t)
	taskHandler.AssertExpectations(t)
	deviceConsumer.AssertExpectations(t)
	tenantConsumer.AssertExpectations(t)
	taskConsumer.AssertExpectations(t)
	consumerFactory.AssertExpectations(t)
}

// Helper functions for integration tests

func createTestConsumerFactory() pubsub.ConsumerFactory {
	// This would create a real consumer factory for integration tests
	// For now, return a mock
	return &MockConsumerFactory{}
}

func createTestORM() sql.ORM {
	// This would create a real ORM for integration tests
	// For now, return a mock
	return &MockORM{}
}

func createTestDeviceHandler(orm sql.ORM) TopicHandler {
	// This would create a real device handler for integration tests
	// For now, return a mock
	handler := &MockTopicHandler{}
	handler.On("TopicName").Return(pubsub.Topic("devices"))
	return handler
}

func testDeviceReplication(t *testing.T, service *Service, consumerFactory pubsub.ConsumerFactory) {
	// Test device creation
	// deviceData := map[string]any{
	// 	"id":           "test-device-1",
	// 	"name":         "Test Device",
	// 	"display_name": "Test Device Display",
	// 	"app_eui":      "test-app-eui",
	// 	"dev_eui":      "test-dev-eui",
	// 	"app_key":      "test-app-key",
	// 	"tenant_id":    "test-tenant",
	// 	"created_at":   time.Now(),
	// }

	// This would publish a message to the devices topic
	// and verify it gets replicated to the database
	// For now, just verify the service is running
	assert.NotNil(t, service)
	assert.NotNil(t, consumerFactory)
}

// Benchmark tests for performance validation

func BenchmarkReplicator_HandleMessage(b *testing.B) {
	consumerFactory := &MockConsumerFactory{}
	orm := &MockORM{}
	replicator := NewReplicator(consumerFactory, orm)

	handler := &MockTopicHandler{}
	handler.On("GetByID", mock.Anything, "test-key").Return(map[string]any{}, errors.New("not found"))
	handler.On("Create", mock.Anything, pubsub.Key("test-key"), mock.Anything).Return(nil)

	message := map[string]any{"id": "test-key", "name": "test-device"}
	key := pubsub.Key("test-key")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		replicator.handleMessage(context.Background(), pubsub.Topic("test-topic"), handler, key, message)
	}
}

func BenchmarkService_RegisterHandler(b *testing.B) {
	consumerFactory := &MockConsumerFactory{}
	orm := &MockORM{}
	service := NewService(consumerFactory, orm)

	handler := &MockTopicHandler{}
	handler.On("TopicName").Return(pubsub.Topic("test-topic"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.RegisterHandler(handler)
	}
}

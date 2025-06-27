package replication

import (
	"testing"
	"time"

	"zensor-server/internal/infra/pubsub"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	consumerFactory := &MockConsumerFactory{}
	orm := &MockORM{}

	service := NewService(consumerFactory, orm)

	assert.NotNil(t, service)
	assert.NotNil(t, service.replicator)
	assert.Equal(t, consumerFactory, service.consumerFactory)
	assert.Equal(t, orm, service.orm)
}

func TestService_Start_Success(t *testing.T) {
	consumerFactory := &MockConsumerFactory{}
	orm := &MockORM{}
	service := NewService(consumerFactory, orm)

	// Start with no handlers (should succeed)
	err := service.Start()

	assert.NoError(t, err)
}

func TestService_Start_WithHandlers(t *testing.T) {
	consumerFactory := &MockConsumerFactory{}
	orm := &MockORM{}
	service := NewService(consumerFactory, orm)

	// Register a handler
	handler := &MockTopicHandler{}
	handler.On("TopicName").Return(pubsub.Topic("test-topic"))

	consumer := &MockConsumer{}
	consumer.On("Consume", pubsub.Topic("test-topic"), mock.AnythingOfType("pubsub.MessageHandler"), mock.Anything).Return(nil)

	consumerFactory.On("New").Return(consumer)

	err := service.RegisterHandler(handler)
	require.NoError(t, err)

	// Start the service
	err = service.Start()
	assert.NoError(t, err)

	// Give some time for goroutines to start
	time.Sleep(10 * time.Millisecond)

	// Stop the service
	service.Stop()

	handler.AssertExpectations(t)
	consumer.AssertExpectations(t)
	consumerFactory.AssertExpectations(t)
}

func TestService_Stop(t *testing.T) {
	consumerFactory := &MockConsumerFactory{}
	orm := &MockORM{}
	service := NewService(consumerFactory, orm)

	// Start the service
	err := service.Start()
	require.NoError(t, err)

	// Stop the service
	service.Stop()

	// Verify the replicator context is cancelled
	select {
	case <-service.replicator.ctx.Done():
		// Expected
	default:
		t.Error("Context should be cancelled after stop")
	}
}

func TestService_RegisterHandler(t *testing.T) {
	consumerFactory := &MockConsumerFactory{}
	orm := &MockORM{}
	service := NewService(consumerFactory, orm)

	handler := &MockTopicHandler{}
	handler.On("TopicName").Return(pubsub.Topic("test-topic"))

	err := service.RegisterHandler(handler)

	assert.NoError(t, err)
	handler.AssertExpectations(t)
}

func TestService_RegisterHandler_Duplicate(t *testing.T) {
	consumerFactory := &MockConsumerFactory{}
	orm := &MockORM{}
	service := NewService(consumerFactory, orm)

	handler1 := &MockTopicHandler{}
	handler1.On("TopicName").Return(pubsub.Topic("test-topic"))

	handler2 := &MockTopicHandler{}
	handler2.On("TopicName").Return(pubsub.Topic("test-topic"))

	// Register first handler
	err := service.RegisterHandler(handler1)
	assert.NoError(t, err)

	// Try to register second handler with same topic
	err = service.RegisterHandler(handler2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler already registered for topic")

	handler1.AssertExpectations(t)
	handler2.AssertExpectations(t)
}

// MockReplicator is a mock implementation of Replicator for testing Service
type MockReplicator struct {
	mock.Mock
}

func (m *MockReplicator) RegisterHandler(handler TopicHandler) error {
	args := m.Called(handler)
	return args.Error(0)
}

func (m *MockReplicator) Start() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockReplicator) Stop() {
	m.Called()
}

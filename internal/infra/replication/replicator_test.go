package replication

import (
	"context"
	dbsql "database/sql"
	"errors"
	"testing"
	"time"

	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockConsumerFactory is a mock implementation of pubsub.ConsumerFactory
type MockConsumerFactory struct {
	mock.Mock
}

func (m *MockConsumerFactory) New() pubsub.Consumer {
	args := m.Called()
	return args.Get(0).(pubsub.Consumer)
}

// MockConsumer is a mock implementation of pubsub.Consumer
type MockConsumer struct {
	mock.Mock
}

func (m *MockConsumer) Consume(topic pubsub.Topic, handler pubsub.MessageHandler, prototype pubsub.Prototype) error {
	args := m.Called(topic, handler, prototype)
	return args.Error(0)
}

// MockORM is a mock implementation of sql.ORM
type MockORM struct {
	mock.Mock
}

func (m *MockORM) AutoMigrate(dst ...any) error {
	args := m.Called(dst)
	return args.Error(0)
}

func (m *MockORM) Count(count *int64) sql.ORM {
	args := m.Called(count)
	return args.Get(0).(sql.ORM)
}

func (m *MockORM) Create(value any) sql.ORM {
	args := m.Called(value)
	return args.Get(0).(sql.ORM)
}

func (m *MockORM) Delete(value any, conds ...any) sql.ORM {
	args := m.Called(value, conds)
	return args.Get(0).(sql.ORM)
}

func (m *MockORM) Find(dest any, conds ...any) sql.ORM {
	args := m.Called(dest, conds)
	return args.Get(0).(sql.ORM)
}

func (m *MockORM) First(dest any, conds ...any) sql.ORM {
	args := m.Called(dest, conds)
	return args.Get(0).(sql.ORM)
}

func (m *MockORM) Limit(limit int) sql.ORM {
	args := m.Called(limit)
	return args.Get(0).(sql.ORM)
}

func (m *MockORM) Model(value any) sql.ORM {
	args := m.Called(value)
	return args.Get(0).(sql.ORM)
}

func (m *MockORM) Offset(offset int) sql.ORM {
	args := m.Called(offset)
	return args.Get(0).(sql.ORM)
}

func (m *MockORM) Preload(query string, args ...any) sql.ORM {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(sql.ORM)
}

func (m *MockORM) Save(value any) sql.ORM {
	args := m.Called(value)
	return args.Get(0).(sql.ORM)
}

func (m *MockORM) Transaction(fc func(tx sql.ORM) error, opts ...*dbsql.TxOptions) error {
	args := m.Called(fc, opts)
	return args.Error(0)
}

func (m *MockORM) Unscoped() sql.ORM {
	args := m.Called()
	return args.Get(0).(sql.ORM)
}

func (m *MockORM) Where(query any, args ...any) sql.ORM {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(sql.ORM)
}

func (m *MockORM) WithContext(ctx context.Context) sql.ORM {
	args := m.Called(ctx)
	return args.Get(0).(sql.ORM)
}

func (m *MockORM) Joins(value string, args ...any) sql.ORM {
	mockArgs := m.Called(value, args)
	return mockArgs.Get(0).(sql.ORM)
}

func (m *MockORM) InnerJoins(value string, args ...any) sql.ORM {
	mockArgs := m.Called(value, args)
	return mockArgs.Get(0).(sql.ORM)
}

func (m *MockORM) Error() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockORM) Order(value any) sql.ORM {
	args := m.Called(value)
	return args.Get(0).(sql.ORM)
}

// MockTopicHandler is a mock implementation of TopicHandler
type MockTopicHandler struct {
	mock.Mock
}

func (m *MockTopicHandler) TopicName() pubsub.Topic {
	args := m.Called()
	return args.Get(0).(pubsub.Topic)
}

func (m *MockTopicHandler) Create(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	args := m.Called(ctx, key, message)
	return args.Error(0)
}

func (m *MockTopicHandler) GetByID(ctx context.Context, id string) (pubsub.Message, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(pubsub.Message), args.Error(1)
}

func (m *MockTopicHandler) Update(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	args := m.Called(ctx, key, message)
	return args.Error(0)
}

func TestNewReplicator(t *testing.T) {
	consumerFactory := &MockConsumerFactory{}
	orm := &MockORM{}

	replicator := NewReplicator(consumerFactory, orm)

	assert.NotNil(t, replicator)
	assert.Equal(t, consumerFactory, replicator.consumerFactory)
	assert.Equal(t, orm, replicator.orm)
	assert.NotNil(t, replicator.handlers)
	assert.NotNil(t, replicator.ctx)
	assert.NotNil(t, replicator.cancel)
}

func TestReplicator_RegisterHandler(t *testing.T) {
	consumerFactory := &MockConsumerFactory{}
	orm := &MockORM{}
	replicator := NewReplicator(consumerFactory, orm)

	handler := &MockTopicHandler{}
	handler.On("TopicName").Return(pubsub.Topic("test-topic"))

	err := replicator.RegisterHandler(handler)

	assert.NoError(t, err)
	assert.Len(t, replicator.handlers, 1)
	assert.Equal(t, handler, replicator.handlers["test-topic"])
	handler.AssertExpectations(t)
}

func TestReplicator_RegisterHandler_Duplicate(t *testing.T) {
	consumerFactory := &MockConsumerFactory{}
	orm := &MockORM{}
	replicator := NewReplicator(consumerFactory, orm)

	handler1 := &MockTopicHandler{}
	handler1.On("TopicName").Return(pubsub.Topic("test-topic"))

	handler2 := &MockTopicHandler{}
	handler2.On("TopicName").Return(pubsub.Topic("test-topic"))

	// Register first handler
	err := replicator.RegisterHandler(handler1)
	assert.NoError(t, err)

	// Try to register second handler with same topic
	err = replicator.RegisterHandler(handler2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler already registered for topic")

	handler1.AssertExpectations(t)
	handler2.AssertExpectations(t)
}

func TestReplicator_Start_NoHandlers(t *testing.T) {
	consumerFactory := &MockConsumerFactory{}
	orm := &MockORM{}
	replicator := NewReplicator(consumerFactory, orm)

	err := replicator.Start()

	assert.NoError(t, err)
	// Should not start any consumers when no handlers are registered
}

func TestReplicator_Start_WithHandlers(t *testing.T) {
	consumerFactory := &MockConsumerFactory{}
	orm := &MockORM{}
	replicator := NewReplicator(consumerFactory, orm)

	handler := &MockTopicHandler{}
	handler.On("TopicName").Return(pubsub.Topic("test-topic"))

	consumer := &MockConsumer{}
	consumer.On("Consume", pubsub.Topic("test-topic"), mock.AnythingOfType("pubsub.MessageHandler"), mock.Anything).Return(nil)

	consumerFactory.On("New").Return(consumer)

	err := replicator.RegisterHandler(handler)
	require.NoError(t, err)

	err = replicator.Start()
	assert.NoError(t, err)

	// Give some time for goroutines to start
	time.Sleep(10 * time.Millisecond)

	handler.AssertExpectations(t)
	consumer.AssertExpectations(t)
	consumerFactory.AssertExpectations(t)
}

func TestReplicator_Stop(t *testing.T) {
	consumerFactory := &MockConsumerFactory{}
	orm := &MockORM{}
	replicator := NewReplicator(consumerFactory, orm)

	// Start replication
	handler := &MockTopicHandler{}
	handler.On("TopicName").Return(pubsub.Topic("test-topic"))

	consumer := &MockConsumer{}
	consumer.On("Consume", pubsub.Topic("test-topic"), mock.AnythingOfType("pubsub.MessageHandler"), mock.Anything).Return(nil)

	consumerFactory.On("New").Return(consumer)

	err := replicator.RegisterHandler(handler)
	require.NoError(t, err)

	err = replicator.Start()
	require.NoError(t, err)

	// Stop replication
	replicator.Stop()

	// Context should be cancelled
	select {
	case <-replicator.ctx.Done():
		// Expected
	default:
		t.Error("Context should be cancelled after stop")
	}

	handler.AssertExpectations(t)
	consumer.AssertExpectations(t)
	consumerFactory.AssertExpectations(t)
}

func TestReplicator_handleMessage_CreateNewRecord(t *testing.T) {
	consumerFactory := &MockConsumerFactory{}
	orm := &MockORM{}
	replicator := NewReplicator(consumerFactory, orm)

	handler := &MockTopicHandler{}
	handler.On("GetByID", mock.Anything, "test-key").Return(map[string]any{}, errors.New("not found"))
	handler.On("Create", mock.Anything, pubsub.Key("test-key"), mock.Anything).Return(nil)

	message := map[string]any{"id": "test-key", "name": "test-device"}
	key := pubsub.Key("test-key")

	err := replicator.handleMessage(pubsub.Topic("test-topic"), handler, key, message)

	assert.NoError(t, err)
	handler.AssertExpectations(t)
}

func TestReplicator_handleMessage_UpdateExistingRecord(t *testing.T) {
	consumerFactory := &MockConsumerFactory{}
	orm := &MockORM{}
	replicator := NewReplicator(consumerFactory, orm)

	handler := &MockTopicHandler{}
	handler.On("GetByID", mock.Anything, "test-key").Return(map[string]any{"id": "test-key"}, nil)
	handler.On("Update", mock.Anything, pubsub.Key("test-key"), mock.Anything).Return(nil)

	message := map[string]any{"id": "test-key", "name": "test-device"}
	key := pubsub.Key("test-key")

	err := replicator.handleMessage(pubsub.Topic("test-topic"), handler, key, message)

	assert.NoError(t, err)
	handler.AssertExpectations(t)
}

func TestReplicator_handleMessage_CreateError(t *testing.T) {
	consumerFactory := &MockConsumerFactory{}
	orm := &MockORM{}
	replicator := NewReplicator(consumerFactory, orm)

	handler := &MockTopicHandler{}
	handler.On("GetByID", mock.Anything, "test-key").Return(map[string]any{}, errors.New("not found"))
	handler.On("Create", mock.Anything, pubsub.Key("test-key"), mock.Anything).Return(errors.New("create failed"))

	message := map[string]any{"id": "test-key", "name": "test-device"}
	key := pubsub.Key("test-key")

	err := replicator.handleMessage(pubsub.Topic("test-topic"), handler, key, message)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "creating record")
	handler.AssertExpectations(t)
}

func TestReplicator_handleMessage_UpdateError(t *testing.T) {
	consumerFactory := &MockConsumerFactory{}
	orm := &MockORM{}
	replicator := NewReplicator(consumerFactory, orm)

	handler := &MockTopicHandler{}
	handler.On("GetByID", mock.Anything, "test-key").Return(map[string]any{"id": "test-key"}, nil)
	handler.On("Update", mock.Anything, pubsub.Key("test-key"), mock.Anything).Return(errors.New("update failed"))

	message := map[string]any{"id": "test-key", "name": "test-device"}
	key := pubsub.Key("test-key")

	err := replicator.handleMessage(pubsub.Topic("test-topic"), handler, key, message)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "updating record")
	handler.AssertExpectations(t)
}

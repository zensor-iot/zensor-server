package workers

import (
	"context"
	"testing"
	"time"

	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/data_plane/dto"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/mqtt"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/device"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing
type MockDeviceService struct {
	mock.Mock
}

func (m *MockDeviceService) CreateDevice(ctx context.Context, device domain.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *MockDeviceService) GetDevice(ctx context.Context, id domain.ID) (domain.Device, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Device), args.Error(1)
}

func (m *MockDeviceService) AllDevices(ctx context.Context) ([]domain.Device, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Device), args.Error(1)
}

func (m *MockDeviceService) DevicesByTenant(ctx context.Context, tenantID domain.ID, pagination usecases.Pagination) ([]domain.Device, int, error) {
	args := m.Called(ctx, tenantID, pagination)
	return args.Get(0).([]domain.Device), args.Int(1), args.Error(2)
}

func (m *MockDeviceService) UpdateDeviceDisplayName(ctx context.Context, id domain.ID, displayName string) error {
	args := m.Called(ctx, id, displayName)
	return args.Error(0)
}

func (m *MockDeviceService) QueueCommand(ctx context.Context, command domain.Command) error {
	args := m.Called(ctx, command)
	return args.Error(0)
}

func (m *MockDeviceService) QueueCommandSequence(ctx context.Context, sequence domain.CommandSequence) error {
	args := m.Called(ctx, sequence)
	return args.Error(0)
}

func (m *MockDeviceService) AdoptDeviceToTenant(ctx context.Context, deviceID domain.ID, tenantID domain.ID) error {
	args := m.Called(ctx, deviceID, tenantID)
	return args.Error(0)
}

func (m *MockDeviceService) UpdateLastMessageReceivedAt(ctx context.Context, deviceName string) error {
	args := m.Called(ctx, deviceName)
	return args.Error(0)
}

type MockDeviceStateCacheService struct {
	mock.Mock
}

func (m *MockDeviceStateCacheService) SetState(ctx context.Context, deviceName string, state map[string][]dto.SensorData) error {
	args := m.Called(ctx, deviceName, state)
	return args.Error(0)
}

func (m *MockDeviceStateCacheService) GetState(ctx context.Context, deviceID string) (usecases.DeviceState, bool) {
	args := m.Called(ctx, deviceID)
	return args.Get(0).(usecases.DeviceState), args.Bool(1)
}

func (m *MockDeviceStateCacheService) GetAllDeviceIDs(ctx context.Context) []string {
	args := m.Called(ctx)
	return args.Get(0).([]string)
}

type MockMQTTClient struct {
	mock.Mock
}

func (m *MockMQTTClient) Publish(topic string, payload interface{}) error {
	args := m.Called(topic, payload)
	return args.Error(0)
}

func (m *MockMQTTClient) Subscribe(topic string, qos byte, handler mqtt.MessageHandler) error {
	args := m.Called(topic, qos, handler)
	return args.Error(0)
}

func (m *MockMQTTClient) Disconnect() {
	m.Called()
}

type MockInternalBroker struct {
	mock.Mock
}

func (m *MockInternalBroker) Publish(ctx context.Context, topic async.BrokerTopicName, message async.BrokerMessage) error {
	args := m.Called(ctx, topic, message)
	return args.Error(0)
}

func (m *MockInternalBroker) Subscribe(topic async.BrokerTopicName) (async.Subscription, error) {
	args := m.Called(topic)
	return args.Get(0).(async.Subscription), args.Error(1)
}

func (m *MockInternalBroker) Unsubscribe(topic async.BrokerTopicName, subscription async.Subscription) error {
	args := m.Called(topic, subscription)
	return args.Error(0)
}

func (m *MockInternalBroker) Stop() {
	m.Called()
}

type MockPubSubConsumerFactory struct {
	mock.Mock
}

func (m *MockPubSubConsumerFactory) New(topic pubsub.Topic, prototype pubsub.Prototype) (pubsub.Consumer, error) {
	args := m.Called(topic, prototype)
	return args.Get(0).(pubsub.Consumer), args.Error(1)
}

type MockPubSubConsumer struct {
	mock.Mock
}

func (m *MockPubSubConsumer) Consume(topic pubsub.Topic, handler pubsub.MessageHandler, prototype pubsub.Prototype) error {
	args := m.Called(topic, handler, prototype)
	return args.Error(0)
}

func TestLoraIntegrationWorker_CorrelationIDInTTNMessage(t *testing.T) {
	// Create mock services
	mockDeviceService := &MockDeviceService{}
	mockStateCache := &MockDeviceStateCacheService{}
	mockMQTTClient := &MockMQTTClient{}
	mockBroker := &MockInternalBroker{}
	mockConsumer := &MockPubSubConsumer{}

	// Create a test command
	testCommand := &device.Command{
		ID:         "test-command-123",
		Version:    1,
		DeviceID:   "test-device-id",
		DeviceName: "test-device",
		TaskID:     "test-task-id",
		Payload: device.CommandPayload{
			Index: 1,
			Value: 100,
		},
		DispatchAfter: utils.Time{Time: time.Now()},
		Port:          15,
		Priority:      "NORMAL",
		CreatedAt:     utils.Time{Time: time.Now()},
		Ready:         true,
		Sent:          false,
	}

	// Set up mock expectations
	mockMQTTClient.On("Publish", mock.AnythingOfType("string"), mock.AnythingOfType("dto.TTNMessage")).Return(nil)
	mockBroker.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Create the worker
	worker := &LoraIntegrationWorker{
		service:        mockDeviceService,
		stateCache:     mockStateCache,
		mqttClient:     mockMQTTClient,
		broker:         mockBroker,
		pubsubConsumer: mockConsumer,
	}

	// Test the deviceCommandHandler method
	ctx := context.Background()
	msg := pubsub.ConsumedMessage{
		Ctx:   ctx,
		Key:   "test-key",
		Value: testCommand,
	}

	// Call the handler
	worker.deviceCommandHandler(ctx, msg, func() {})

	// Verify that Publish was called with a TTN message containing correlation IDs
	mockMQTTClient.AssertExpectations(t)

	// Get the actual TTN message that was published
	calls := mockMQTTClient.Calls
	assert.Greater(t, len(calls), 0, "Publish should have been called")

	// Extract the TTN message from the call
	ttnMsg := calls[0].Arguments[1].(dto.TTNMessage)
	assert.Equal(t, 1, len(ttnMsg.Downlinks), "Should have one downlink")
	assert.Equal(t, []string{"zensor:test-command-123"}, ttnMsg.Downlinks[0].CorrelationIDs, "Correlation IDs should contain the command ID with zensor prefix")
}

func TestLoraIntegrationWorker_UpdateCommandStatusWithCorrelationID(t *testing.T) {
	// Create mock services
	mockDeviceService := &MockDeviceService{}
	mockStateCache := &MockDeviceStateCacheService{}
	mockMQTTClient := &MockMQTTClient{}
	mockBroker := &MockInternalBroker{}
	mockConsumer := &MockPubSubConsumer{}

	// Create the worker
	worker := &LoraIntegrationWorker{
		service:        mockDeviceService,
		stateCache:     mockStateCache,
		mqttClient:     mockMQTTClient,
		broker:         mockBroker,
		pubsubConsumer: mockConsumer,
	}

	// Create a test envelope with correlation IDs
	envelop := dto.Envelop{
		EndDeviceIDs: dto.EndDeviceIDs{
			DeviceID: "test-device",
		},
		CorrelationIDs: []string{"zensor:test-command-123"},
	}

	// Set up mock expectations
	mockBroker.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Test the updateCommandStatus method
	ctx := context.Background()
	status := domain.CommandStatusQueued
	var errorMessage *string

	worker.updateCommandStatus(ctx, envelop, status, errorMessage)

	// Verify that Publish was called with the correct status update
	mockBroker.AssertExpectations(t)

	// Get the actual status update that was published
	calls := mockBroker.Calls
	assert.Greater(t, len(calls), 0, "Publish should have been called")

	// Extract the broker message from the call
	brokerMsg := calls[0].Arguments[2].(async.BrokerMessage)
	assert.Equal(t, "command_status_update", brokerMsg.Event, "Event should be command_status_update")

	// Verify the status update contains the command ID
	statusUpdate := brokerMsg.Value.(domain.CommandStatusUpdate)
	assert.Equal(t, "test-command-123", statusUpdate.CommandID, "Command ID should be extracted from correlation IDs")
	assert.Equal(t, "test-device", statusUpdate.DeviceName, "Device name should be preserved")
	assert.Equal(t, status, statusUpdate.Status, "Status should be preserved")
}

func TestLoraIntegrationWorker_UpdateCommandStatusWithoutCorrelationID(t *testing.T) {
	// Create mock services
	mockDeviceService := &MockDeviceService{}
	mockStateCache := &MockDeviceStateCacheService{}
	mockMQTTClient := &MockMQTTClient{}
	mockBroker := &MockInternalBroker{}
	mockConsumer := &MockPubSubConsumer{}

	// Create the worker
	worker := &LoraIntegrationWorker{
		service:        mockDeviceService,
		stateCache:     mockStateCache,
		mqttClient:     mockMQTTClient,
		broker:         mockBroker,
		pubsubConsumer: mockConsumer,
	}

	// Create a test envelope without correlation IDs (backward compatibility)
	envelop := dto.Envelop{
		EndDeviceIDs: dto.EndDeviceIDs{
			DeviceID: "test-device",
		},
		CorrelationIDs: []string{}, // Empty correlation IDs
	}

	// Test the updateCommandStatus method
	ctx := context.Background()
	status := domain.CommandStatusQueued
	var errorMessage *string

	worker.updateCommandStatus(ctx, envelop, status, errorMessage)

	// Verify that Publish was NOT called when there are no correlation IDs
	mockBroker.AssertNotCalled(t, "Publish")
}

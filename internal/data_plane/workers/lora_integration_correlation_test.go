package workers

import (
	"context"
	"time"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/data_plane/dto"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/mqtt"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/device"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
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

var _ = ginkgo.Describe("LoraIntegrationWorker", func() {
	ginkgo.Context("CorrelationID", func() {
		var (
			mockDeviceService *MockDeviceService
			mockStateCache    *MockDeviceStateCacheService
			mockMQTTClient    *MockMQTTClient
			mockBroker        *MockInternalBroker
			mockConsumer      *MockPubSubConsumer
			worker            *LoraIntegrationWorker
		)

		ginkgo.BeforeEach(func() {
			// Create mock services
			mockDeviceService = &MockDeviceService{}
			mockStateCache = &MockDeviceStateCacheService{}
			mockMQTTClient = &MockMQTTClient{}
			mockBroker = &MockInternalBroker{}
			mockConsumer = &MockPubSubConsumer{}

			// Create the worker
			worker = &LoraIntegrationWorker{
				service:        mockDeviceService,
				stateCache:     mockStateCache,
				mqttClient:     mockMQTTClient,
				broker:         mockBroker,
				pubsubConsumer: mockConsumer,
			}
		})

		ginkgo.When("handling TTN message with correlation ID", func() {
			ginkgo.It("should include correlation ID in TTN message", func() {
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
				mockMQTTClient.AssertExpectations(ginkgo.GinkgoT())

				// Get the actual TTN message that was published
				calls := mockMQTTClient.Calls
				gomega.Expect(len(calls)).To(gomega.BeNumerically(">", 0))

				// Extract the TTN message from the call
				ttnMsg := calls[0].Arguments[1].(dto.TTNMessage)
				gomega.Expect(len(ttnMsg.Downlinks)).To(gomega.Equal(1))
				gomega.Expect(ttnMsg.Downlinks[0].CorrelationIDs).To(gomega.Equal([]string{"zensor:test-command-123"}))
			})
		})

		ginkgo.Context("UpdateCommandStatus", func() {
			ginkgo.When("updating command status with correlation ID", func() {
				ginkgo.It("should extract command ID from correlation IDs", func() {
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
					mockBroker.AssertExpectations(ginkgo.GinkgoT())

					// Get the actual status update that was published
					calls := mockBroker.Calls
					gomega.Expect(len(calls)).To(gomega.BeNumerically(">", 0))

					// Extract the broker message from the call
					brokerMsg := calls[0].Arguments[2].(async.BrokerMessage)
					gomega.Expect(brokerMsg.Event).To(gomega.Equal("command_status_update"))

					// Verify the status update contains the command ID
					statusUpdate := brokerMsg.Value.(domain.CommandStatusUpdate)
					gomega.Expect(statusUpdate.CommandID).To(gomega.Equal("test-command-123"))
					gomega.Expect(statusUpdate.DeviceName).To(gomega.Equal("test-device"))
					gomega.Expect(statusUpdate.Status).To(gomega.Equal(status))
				})
			})

			ginkgo.When("updating command status without correlation ID", func() {
				ginkgo.It("should not publish when no correlation IDs are present", func() {
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
					mockBroker.AssertNotCalled(ginkgo.GinkgoT(), "Publish")
				})
			})
		})
	})
})

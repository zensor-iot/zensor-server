package usecases

import (
	"context"
	"testing"
	"time"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing
type MockScheduledTaskRepository struct {
	mock.Mock
}

func (m *MockScheduledTaskRepository) Create(ctx context.Context, scheduledTask domain.ScheduledTask) error {
	args := m.Called(ctx, scheduledTask)
	return args.Error(0)
}

func (m *MockScheduledTaskRepository) GetByID(ctx context.Context, id domain.ID) (domain.ScheduledTask, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.ScheduledTask), args.Error(1)
}

func (m *MockScheduledTaskRepository) FindAllByTenant(ctx context.Context, tenantID domain.ID) ([]domain.ScheduledTask, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]domain.ScheduledTask), args.Error(1)
}

func (m *MockScheduledTaskRepository) FindAllByTenantAndDevice(ctx context.Context, tenantID domain.ID, deviceID domain.ID, pagination Pagination) ([]domain.ScheduledTask, int, error) {
	args := m.Called(ctx, tenantID, deviceID, pagination)
	return args.Get(0).([]domain.ScheduledTask), args.Int(1), args.Error(2)
}

func (m *MockScheduledTaskRepository) Update(ctx context.Context, scheduledTask domain.ScheduledTask) error {
	args := m.Called(ctx, scheduledTask)
	return args.Error(0)
}

func (m *MockScheduledTaskRepository) Delete(ctx context.Context, id domain.ID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockScheduledTaskRepository) FindAllActive(ctx context.Context) ([]domain.ScheduledTask, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.ScheduledTask), args.Error(1)
}

type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) Create(ctx context.Context, task domain.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskService) Run(ctx context.Context, task domain.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskService) FindAllByDevice(ctx context.Context, deviceID domain.ID, pagination Pagination) ([]domain.Task, int, error) {
	args := m.Called(ctx, deviceID, pagination)
	return args.Get(0).([]domain.Task), args.Int(1), args.Error(2)
}

func (m *MockTaskService) FindAllByScheduledTask(ctx context.Context, scheduledTaskID domain.ID, pagination Pagination) ([]domain.Task, int, error) {
	args := m.Called(ctx, scheduledTaskID, pagination)
	return args.Get(0).([]domain.Task), args.Int(1), args.Error(2)
}

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

func (m *MockDeviceService) DevicesByTenant(ctx context.Context, tenantID domain.ID, pagination Pagination) ([]domain.Device, int, error) {
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

type MockTenantConfigurationService struct {
	mock.Mock
}

func (m *MockTenantConfigurationService) CreateTenantConfiguration(ctx context.Context, config domain.TenantConfiguration) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockTenantConfigurationService) GetTenantConfiguration(ctx context.Context, tenantID domain.ID) (domain.TenantConfiguration, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).(domain.TenantConfiguration), args.Error(1)
}

func (m *MockTenantConfigurationService) UpdateTenantConfiguration(ctx context.Context, config domain.TenantConfiguration) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockTenantConfigurationService) GetOrCreateTenantConfiguration(ctx context.Context, tenantID domain.ID, defaultTimezone string) (domain.TenantConfiguration, error) {
	args := m.Called(ctx, tenantID, defaultTimezone)
	return args.Get(0).(domain.TenantConfiguration), args.Error(1)
}

type MockInternalBroker struct {
	mock.Mock
}

func (m *MockInternalBroker) Subscribe(topic async.BrokerTopicName) (async.Subscription, error) {
	args := m.Called(topic)
	return args.Get(0).(async.Subscription), args.Error(1)
}

func (m *MockInternalBroker) Unsubscribe(topic async.BrokerTopicName, subscription async.Subscription) error {
	args := m.Called(topic, subscription)
	return args.Error(0)
}

func (m *MockInternalBroker) Publish(ctx context.Context, topic async.BrokerTopicName, msg async.BrokerMessage) error {
	args := m.Called(ctx, topic, msg)
	return args.Error(0)
}

func (m *MockInternalBroker) Stop() {
	m.Called()
}

func TestScheduledTaskWorker_ShouldExecuteSchedule_WithTimezone(t *testing.T) {
	tests := []struct {
		name           string
		schedule       string
		lastExecuted   time.Time
		tenantTimezone string
		expectedResult bool
		description    string
	}{
		{
			name:           "UTC timezone - should execute",
			schedule:       "00 04 * * *", // 4:00 AM daily
			lastExecuted:   time.Date(2024, 1, 1, 3, 0, 0, 0, time.UTC),
			tenantTimezone: "UTC",
			expectedResult: true,
			description:    "Should execute when current time is past the scheduled time in UTC",
		},
		{
			name:           "America/New_York timezone - should execute",
			schedule:       "00 04 * * *", // 4:00 AM daily
			lastExecuted:   time.Date(2024, 1, 1, 3, 0, 0, 0, time.UTC),
			tenantTimezone: "America/New_York",
			expectedResult: true,
			description:    "Should execute when current time is past the scheduled time in EST",
		},
		{
			name:           "Europe/London timezone - should execute",
			schedule:       "00 04 * * *", // 4:00 AM daily
			lastExecuted:   time.Date(2024, 1, 1, 3, 0, 0, 0, time.UTC),
			tenantTimezone: "Europe/London",
			expectedResult: true,
			description:    "Should execute when current time is past the scheduled time in GMT",
		},
		{
			name:           "Invalid timezone - fallback to UTC",
			schedule:       "00 04 * * *", // 4:00 AM daily
			lastExecuted:   time.Date(2024, 1, 1, 3, 0, 0, 0, time.UTC),
			tenantTimezone: "Invalid/Timezone",
			expectedResult: true,
			description:    "Should fallback to UTC when timezone is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockTenantConfigService := &MockTenantConfigurationService{}
			mockScheduledTaskRepo := &MockScheduledTaskRepository{}
			mockTaskService := &MockTaskService{}
			mockDeviceService := &MockDeviceService{}
			mockBroker := &MockInternalBroker{}

			// Create tenant configuration
			tenantConfig, _ := domain.NewTenantConfigurationBuilder().
				WithTenantID(domain.ID("test-tenant")).
				WithTimezone(tt.tenantTimezone).
				Build()

			// Setup mock expectations
			mockTenantConfigService.On("GetOrCreateTenantConfiguration", mock.Anything, domain.ID("test-tenant"), "UTC").
				Return(tenantConfig, nil)

			// Create worker
			worker := &ScheduledTaskWorker{
				tenantConfigurationService: mockTenantConfigService,
				scheduledTaskRepository:    mockScheduledTaskRepo,
				taskService:                mockTaskService,
				deviceService:              mockDeviceService,
				broker:                     mockBroker,
				cronParser:                 cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
			}

			// Test the method
			result, err := worker.shouldExecuteSchedule(tt.schedule, tt.lastExecuted, domain.ID("test-tenant"))

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result, tt.description)

			// Verify mock expectations
			mockTenantConfigService.AssertExpectations(t)
		})
	}
}

func TestScheduledTaskWorker_ShouldExecuteSchedule_WithTenantConfigurationError(t *testing.T) {
	// Create mocks
	mockTenantConfigService := &MockTenantConfigurationService{}
	mockScheduledTaskRepo := &MockScheduledTaskRepository{}
	mockTaskService := &MockTaskService{}
	mockDeviceService := &MockDeviceService{}
	mockBroker := &MockInternalBroker{}

	// Setup mock to return error
	mockTenantConfigService.On("GetOrCreateTenantConfiguration", mock.Anything, domain.ID("test-tenant"), "UTC").
		Return(domain.TenantConfiguration{}, assert.AnError)

	// Create worker
	worker := &ScheduledTaskWorker{
		tenantConfigurationService: mockTenantConfigService,
		scheduledTaskRepository:    mockScheduledTaskRepo,
		taskService:                mockTaskService,
		deviceService:              mockDeviceService,
		broker:                     mockBroker,
		cronParser:                 cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	}

	// Test the method - should fallback to UTC
	result, err := worker.shouldExecuteSchedule("00 04 * * *", time.Date(2024, 1, 1, 3, 0, 0, 0, time.UTC), domain.ID("test-tenant"))

	// Assertions
	assert.NoError(t, err)
	assert.True(t, result, "Should fallback to UTC when tenant configuration fails")

	// Verify mock expectations
	mockTenantConfigService.AssertExpectations(t)
}

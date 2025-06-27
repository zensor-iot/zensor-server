package handlers

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
)

// MockORM is a mock implementation of sql.ORM for testing handlers
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

func TestNewDeviceHandler(t *testing.T) {
	orm := &MockORM{}
	handler := NewDeviceHandler(orm)

	assert.NotNil(t, handler)
	assert.Equal(t, orm, handler.orm)
}

func TestDeviceHandler_TopicName(t *testing.T) {
	orm := &MockORM{}
	handler := NewDeviceHandler(orm)

	topic := handler.TopicName()

	assert.Equal(t, pubsub.Topic("devices"), topic)
}

func TestDeviceHandler_Create_Success(t *testing.T) {
	orm := &MockORM{}
	handler := NewDeviceHandler(orm)

	// Mock ORM calls
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
	mockOrm.On("Create", mock.Anything).Return(mockOrm)
	mockOrm.On("Error").Return(nil)

	handler.orm = mockOrm

	device := struct {
		ID          string
		Name        string
		DisplayName string
		AppEUI      string
		DevEUI      string
		AppKey      string
		TenantID    *string
		CreatedAt   time.Time
	}{
		ID:          "test-device-1",
		Name:        "Test Device",
		DisplayName: "Test Device Display",
		AppEUI:      "test-app-eui",
		DevEUI:      "test-dev-eui",
		AppKey:      "test-app-key",
		TenantID:    nil,
		CreatedAt:   time.Now(),
	}

	err := handler.Create(context.Background(), "test-device-1", device)

	assert.NoError(t, err)
	mockOrm.AssertExpectations(t)
}

func TestDeviceHandler_Create_Error(t *testing.T) {
	orm := &MockORM{}
	handler := NewDeviceHandler(orm)

	// Mock ORM calls with error
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
	mockOrm.On("Create", mock.Anything).Return(mockOrm)
	mockOrm.On("Error").Return(errors.New("database error"))

	handler.orm = mockOrm

	device := struct {
		ID   string
		Name string
	}{
		ID:   "test-device-1",
		Name: "Test Device",
	}

	err := handler.Create(context.Background(), "test-device-1", device)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "creating device")
	mockOrm.AssertExpectations(t)
}

func TestDeviceHandler_GetByID_Success(t *testing.T) {
	orm := &MockORM{}
	handler := NewDeviceHandler(orm)

	// Mock ORM calls
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)

	// Mock the First method to populate the device data
	var deviceData DeviceData
	mockOrm.On("First", &deviceData, mock.Anything).Run(func(args mock.Arguments) {
		// Populate the device data
		dest := args.Get(0).(*DeviceData)
		*dest = DeviceData{
			ID:          "test-device-1",
			Name:        "Test Device",
			DisplayName: "Test Device Display",
			AppEUI:      "test-app-eui",
			DevEUI:      "test-dev-eui",
			AppKey:      "test-app-key",
			TenantID:    nil,
			CreatedAt:   time.Now(),
		}
	}).Return(mockOrm)

	mockOrm.On("Error").Return(nil)

	handler.orm = mockOrm

	result, err := handler.GetByID(context.Background(), "test-device-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the result is a map with expected fields
	resultMap, ok := result.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "test-device-1", resultMap["id"])
	assert.Equal(t, "Test Device", resultMap["name"])
	assert.Equal(t, "Test Device Display", resultMap["display_name"])

	mockOrm.AssertExpectations(t)
}

func TestDeviceHandler_GetByID_NotFound(t *testing.T) {
	orm := &MockORM{}
	handler := NewDeviceHandler(orm)

	// Mock ORM calls with not found error
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
	mockOrm.On("First", mock.Anything, mock.Anything).Return(mockOrm)
	mockOrm.On("Error").Return(sql.ErrRecordNotFound)

	handler.orm = mockOrm

	result, err := handler.GetByID(context.Background(), "test-device-1")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "getting device")
	mockOrm.AssertExpectations(t)
}

func TestDeviceHandler_Update_Success(t *testing.T) {
	orm := &MockORM{}
	handler := NewDeviceHandler(orm)

	// Mock ORM calls
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
	mockOrm.On("Save", mock.Anything).Return(mockOrm)
	mockOrm.On("Error").Return(nil)

	handler.orm = mockOrm

	device := struct {
		ID          string
		Name        string
		DisplayName string
	}{
		ID:          "test-device-1",
		Name:        "Updated Device",
		DisplayName: "Updated Device Display",
	}

	err := handler.Update(context.Background(), "test-device-1", device)

	assert.NoError(t, err)
	mockOrm.AssertExpectations(t)
}

func TestDeviceHandler_Update_Error(t *testing.T) {
	orm := &MockORM{}
	handler := NewDeviceHandler(orm)

	// Mock ORM calls with error
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
	mockOrm.On("Save", mock.Anything).Return(mockOrm)
	mockOrm.On("Error").Return(errors.New("database error"))

	handler.orm = mockOrm

	device := struct {
		ID   string
		Name string
	}{
		ID:   "test-device-1",
		Name: "Updated Device",
	}

	err := handler.Update(context.Background(), "test-device-1", device)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "updating device")
	mockOrm.AssertExpectations(t)
}

func TestDeviceHandler_extractDeviceFields(t *testing.T) {
	orm := &MockORM{}
	handler := NewDeviceHandler(orm)

	tenantID := "test-tenant"
	lastMessageTime := time.Now()
	createdAt := time.Now()

	device := struct {
		ID                    string
		Name                  string
		DisplayName           string
		AppEUI                string
		DevEUI                string
		AppKey                string
		TenantID              *string
		LastMessageReceivedAt time.Time
		CreatedAt             time.Time
	}{
		ID:                    "test-device-1",
		Name:                  "Test Device",
		DisplayName:           "Test Device Display",
		AppEUI:                "test-app-eui",
		DevEUI:                "test-dev-eui",
		AppKey:                "test-app-key",
		TenantID:              &tenantID,
		LastMessageReceivedAt: lastMessageTime,
		CreatedAt:             createdAt,
	}

	result := handler.extractDeviceFields(device)

	assert.Equal(t, "test-device-1", result.ID)
	assert.Equal(t, "Test Device", result.Name)
	assert.Equal(t, "Test Device Display", result.DisplayName)
	assert.Equal(t, "test-app-eui", result.AppEUI)
	assert.Equal(t, "test-dev-eui", result.DevEUI)
	assert.Equal(t, "test-app-key", result.AppKey)
	assert.Equal(t, &tenantID, result.TenantID)
	assert.Equal(t, &lastMessageTime, result.LastMessageReceivedAt)
	assert.Equal(t, createdAt, result.CreatedAt)
	assert.Equal(t, 1, result.Version)
}

func TestDeviceHandler_toDomainDevice(t *testing.T) {
	orm := &MockORM{}
	handler := NewDeviceHandler(orm)

	tenantID := "test-tenant"
	lastMessageTime := time.Now()

	deviceData := DeviceData{
		ID:                    "test-device-1",
		Name:                  "Test Device",
		DisplayName:           "Test Device Display",
		AppEUI:                "test-app-eui",
		DevEUI:                "test-dev-eui",
		AppKey:                "test-app-key",
		TenantID:              &tenantID,
		LastMessageReceivedAt: &lastMessageTime,
		CreatedAt:             time.Now(),
	}

	result := handler.toDomainDevice(deviceData)

	assert.Equal(t, "test-device-1", result["id"])
	assert.Equal(t, "Test Device", result["name"])
	assert.Equal(t, "Test Device Display", result["display_name"])
	assert.Equal(t, "test-app-eui", result["app_eui"])
	assert.Equal(t, "test-dev-eui", result["dev_eui"])
	assert.Equal(t, "test-app-key", result["app_key"])
	assert.Equal(t, tenantID, result["tenant_id"])
	assert.Equal(t, &lastMessageTime, result["last_message_received_at"])
}

func TestDeviceHandler_toDomainDevice_NilFields(t *testing.T) {
	orm := &MockORM{}
	handler := NewDeviceHandler(orm)

	deviceData := DeviceData{
		ID:          "test-device-1",
		Name:        "Test Device",
		DisplayName: "Test Device Display",
		AppEUI:      "test-app-eui",
		DevEUI:      "test-dev-eui",
		AppKey:      "test-app-key",
		TenantID:    nil,
		CreatedAt:   time.Now(),
	}

	result := handler.toDomainDevice(deviceData)

	assert.Equal(t, "test-device-1", result["id"])
	assert.Equal(t, "Test Device", result["name"])
	assert.Equal(t, "Test Device Display", result["display_name"])
	assert.Equal(t, "test-app-eui", result["app_eui"])
	assert.Equal(t, "test-dev-eui", result["dev_eui"])
	assert.Equal(t, "test-app-key", result["app_key"])

	// Should not contain tenant_id or last_message_received_at when nil
	_, hasTenantID := result["tenant_id"]
	assert.False(t, hasTenantID)

	_, hasLastMessage := result["last_message_received_at"]
	assert.False(t, hasLastMessage)
}

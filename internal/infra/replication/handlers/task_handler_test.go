package handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"zensor-server/internal/infra/sql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTaskHandler_Create_Success(t *testing.T) {
	orm := &MockORM{}
	handler := NewTaskHandler(orm)
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
	mockOrm.On("Create", mock.Anything).Return(mockOrm)
	mockOrm.On("Error").Return(nil)
	handler.orm = mockOrm
	task := struct {
		ID       string
		DeviceID string
	}{ID: "task-1", DeviceID: "device-1"}
	err := handler.Create(context.Background(), "task-1", task)
	assert.NoError(t, err)
	mockOrm.AssertExpectations(t)
}

func TestTaskHandler_Create_Error(t *testing.T) {
	orm := &MockORM{}
	handler := NewTaskHandler(orm)
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
	mockOrm.On("Create", mock.Anything).Return(mockOrm)
	mockOrm.On("Error").Return(errors.New("db error"))
	handler.orm = mockOrm
	task := struct {
		ID       string
		DeviceID string
	}{ID: "task-1", DeviceID: "device-1"}
	err := handler.Create(context.Background(), "task-1", task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "creating task")
	mockOrm.AssertExpectations(t)
}

func TestTaskHandler_GetByID_Success(t *testing.T) {
	orm := &MockORM{}
	handler := NewTaskHandler(orm)
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
	mockOrm.On("First", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*TaskData)
		*dest = TaskData{ID: "task-1", DeviceID: "device-1", Version: 1}
	}).Return(mockOrm)
	mockOrm.On("Error").Return(nil)
	handler.orm = mockOrm

	result, err := handler.GetByID(context.Background(), "task-1")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	resultMap, ok := result.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "task-1", resultMap["id"])
	assert.Equal(t, "device-1", resultMap["device_id"])
	mockOrm.AssertExpectations(t)
}

func TestTaskHandler_GetByID_NotFound(t *testing.T) {
	orm := &MockORM{}
	handler := NewTaskHandler(orm)
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
	mockOrm.On("First", mock.Anything, mock.Anything).Return(mockOrm)
	mockOrm.On("Error").Return(sql.ErrRecordNotFound)
	handler.orm = mockOrm
	result, err := handler.GetByID(context.Background(), "task-1")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "getting task")
	mockOrm.AssertExpectations(t)
}

func TestTaskHandler_Update_Success(t *testing.T) {
	orm := &MockORM{}
	handler := NewTaskHandler(orm)
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
	mockOrm.On("Save", mock.Anything).Return(mockOrm)
	mockOrm.On("Error").Return(nil)
	handler.orm = mockOrm
	task := struct {
		ID       string
		DeviceID string
	}{ID: "task-1", DeviceID: "device-1"}
	err := handler.Update(context.Background(), "task-1", task)
	assert.NoError(t, err)
	mockOrm.AssertExpectations(t)
}

func TestTaskHandler_Update_Error(t *testing.T) {
	orm := &MockORM{}
	handler := NewTaskHandler(orm)
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
	mockOrm.On("Save", mock.Anything).Return(mockOrm)
	mockOrm.On("Error").Return(errors.New("db error"))
	handler.orm = mockOrm
	task := struct {
		ID       string
		DeviceID string
	}{ID: "task-1", DeviceID: "device-1"}
	err := handler.Update(context.Background(), "task-1", task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "updating task")
	mockOrm.AssertExpectations(t)
}

func TestTaskHandler_extractTaskFields(t *testing.T) {
	orm := &MockORM{}
	handler := NewTaskHandler(orm)
	timeNow := time.Now()
	task := struct {
		ID        string
		DeviceID  string
		CreatedAt time.Time
		UpdatedAt time.Time
	}{
		ID:        "task-1",
		DeviceID:  "device-1",
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	result := handler.extractTaskFields(task)
	assert.Equal(t, "task-1", result.ID)
	assert.Equal(t, "device-1", result.DeviceID)
	assert.Equal(t, timeNow, result.CreatedAt)
	assert.Equal(t, timeNow, result.UpdatedAt)
}

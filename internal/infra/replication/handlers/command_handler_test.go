package handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"zensor-server/internal/infra/sql"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/avro"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCommandHandler_Create_Success(t *testing.T) {
	orm := &MockORM{}
	handler := NewCommandHandler(orm)
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
	mockOrm.On("Create", mock.Anything).Return(mockOrm)
	mockOrm.On("Error").Return(nil)
	handler.orm = mockOrm
	cmd := struct {
		ID         string
		DeviceName string
		DeviceID   string
		TaskID     string
	}{ID: "cmd-1", DeviceName: "dev", DeviceID: "dev-1", TaskID: "task-1"}
	err := handler.Create(context.Background(), "cmd-1", cmd)
	assert.NoError(t, err)
	mockOrm.AssertExpectations(t)
}

func TestCommandHandler_Create_Error(t *testing.T) {
	orm := &MockORM{}
	handler := NewCommandHandler(orm)
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
	mockOrm.On("Create", mock.Anything).Return(mockOrm)
	mockOrm.On("Error").Return(errors.New("db error"))
	handler.orm = mockOrm
	cmd := struct {
		ID         string
		DeviceName string
		DeviceID   string
		TaskID     string
	}{ID: "cmd-1", DeviceName: "dev", DeviceID: "dev-1", TaskID: "task-1"}
	err := handler.Create(context.Background(), "cmd-1", cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "creating command")
	mockOrm.AssertExpectations(t)
}

func TestCommandHandler_GetByID_Success(t *testing.T) {
	orm := &MockORM{}
	handler := NewCommandHandler(orm)
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
	mockOrm.On("First", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*CommandData)
		*dest = CommandData{
			ID:         "cmd-1",
			DeviceName: "dev",
			DeviceID:   "device-1",
			TaskID:     "task-1",
			Payload:    CommandPayload{Index: 1, Data: 100},
		}
	}).Return(mockOrm)
	mockOrm.On("Error").Return(nil)
	handler.orm = mockOrm

	result, err := handler.GetByID(context.Background(), "cmd-1")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	resultMap, ok := result.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "cmd-1", resultMap["id"])
	assert.Equal(t, "dev", resultMap["device_name"])
	mockOrm.AssertExpectations(t)
}

func TestCommandHandler_GetByID_NotFound(t *testing.T) {
	orm := &MockORM{}
	handler := NewCommandHandler(orm)
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
	mockOrm.On("First", mock.Anything, mock.Anything).Return(mockOrm)
	mockOrm.On("Error").Return(sql.ErrRecordNotFound)
	handler.orm = mockOrm
	result, err := handler.GetByID(context.Background(), "cmd-1")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "getting command")
	mockOrm.AssertExpectations(t)
}

func TestCommandHandler_Update_Success(t *testing.T) {
	orm := &MockORM{}
	handler := NewCommandHandler(orm)
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
	mockOrm.On("Save", mock.Anything).Return(mockOrm)
	mockOrm.On("Error").Return(nil)
	handler.orm = mockOrm
	cmd := struct {
		ID         string
		DeviceName string
		DeviceID   string
		TaskID     string
	}{ID: "cmd-1", DeviceName: "dev", DeviceID: "dev-1", TaskID: "task-1"}
	err := handler.Update(context.Background(), "cmd-1", cmd)
	assert.NoError(t, err)
	mockOrm.AssertExpectations(t)
}

func TestCommandHandler_Update_Error(t *testing.T) {
	orm := &MockORM{}
	handler := NewCommandHandler(orm)
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
	mockOrm.On("Save", mock.Anything).Return(mockOrm)
	mockOrm.On("Error").Return(errors.New("db error"))
	handler.orm = mockOrm
	cmd := struct {
		ID         string
		DeviceName string
		DeviceID   string
		TaskID     string
	}{ID: "cmd-1", DeviceName: "dev", DeviceID: "dev-1", TaskID: "task-1"}
	err := handler.Update(context.Background(), "cmd-1", cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "updating command")
	mockOrm.AssertExpectations(t)
}

func TestCommandHandler_extractCommandFields(t *testing.T) {
	orm := &MockORM{}
	handler := NewCommandHandler(orm)
	timeNow := time.Now()

	cmd := &avro.AvroCommand{
		ID:            "cmd-1",
		DeviceName:    "dev",
		DeviceID:      "dev-1",
		TaskID:        "task-1",
		PayloadIndex:  1,
		PayloadValue:  2,
		DispatchAfter: timeNow,
		Port:          3,
		Priority:      "high",
		CreatedAt:     timeNow,
		Ready:         true,
		Sent:          false,
		SentAt:        timeNow,
	}

	result := handler.extractCommandFields(cmd)
	assert.Equal(t, "cmd-1", result.ID)
	assert.Equal(t, "dev", result.DeviceName)
	assert.Equal(t, "dev-1", result.DeviceID)
	assert.Equal(t, "task-1", result.TaskID)
	assert.Equal(t, uint8(1), result.Payload.Index)
	assert.Equal(t, uint8(2), result.Payload.Data)
	assert.Equal(t, utils.Time{Time: timeNow}, result.DispatchAfter)
	assert.Equal(t, uint8(3), result.Port)
	assert.Equal(t, "high", result.Priority)
	assert.Equal(t, utils.Time{Time: timeNow}, result.CreatedAt)
	assert.Equal(t, true, result.Ready)
	assert.Equal(t, false, result.Sent)
	assert.Equal(t, utils.Time{Time: timeNow}, result.SentAt)
}

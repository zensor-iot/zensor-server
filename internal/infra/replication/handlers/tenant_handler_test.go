package handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"zensor-server/internal/infra/sql"
	"zensor-server/internal/shared_kernel/avro"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTenantHandler_Create_Success(t *testing.T) {
	orm := &MockORM{}
	handler := NewTenantHandler(orm)
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
	mockOrm.On("Create", mock.Anything).Return(mockOrm)
	mockOrm.On("Error").Return(nil)
	handler.orm = mockOrm

	tenant := struct {
		ID    string
		Name  string
		Email string
	}{ID: "tenant-1", Name: "Tenant", Email: "tenant@example.com"}
	err := handler.Create(context.Background(), "tenant-1", tenant)
	assert.NoError(t, err)
	mockOrm.AssertExpectations(t)
}

func TestTenantHandler_Create_Error(t *testing.T) {
	orm := &MockORM{}
	handler := NewTenantHandler(orm)
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
	mockOrm.On("Create", mock.Anything).Return(mockOrm)
	mockOrm.On("Error").Return(errors.New("db error"))
	handler.orm = mockOrm

	tenant := struct {
		ID    string
		Name  string
		Email string
	}{ID: "tenant-1", Name: "Tenant", Email: "tenant@example.com"}
	err := handler.Create(context.Background(), "tenant-1", tenant)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "creating tenant")
	mockOrm.AssertExpectations(t)
}

func TestTenantHandler_GetByID_Success(t *testing.T) {
	orm := &MockORM{}
	handler := NewTenantHandler(orm)
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
	mockOrm.On("First", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*TenantData)
		*dest = TenantData{ID: "tenant-1", Name: "Tenant", Email: "tenant@example.com"}
	}).Return(mockOrm)
	mockOrm.On("Error").Return(nil)
	handler.orm = mockOrm

	result, err := handler.GetByID(context.Background(), "tenant-1")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	resultMap, ok := result.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "tenant-1", resultMap["id"])
	assert.Equal(t, "Tenant", resultMap["name"])
	mockOrm.AssertExpectations(t)
}

func TestTenantHandler_GetByID_NotFound(t *testing.T) {
	orm := &MockORM{}
	handler := NewTenantHandler(orm)
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
	mockOrm.On("First", mock.Anything, mock.Anything).Return(mockOrm)
	mockOrm.On("Error").Return(sql.ErrRecordNotFound)
	handler.orm = mockOrm
	result, err := handler.GetByID(context.Background(), "tenant-1")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "getting tenant")
	mockOrm.AssertExpectations(t)
}

func TestTenantHandler_Update_Success(t *testing.T) {
	orm := &MockORM{}
	handler := NewTenantHandler(orm)
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
	mockOrm.On("First", mock.Anything, mock.Anything).Return(mockOrm)
	mockOrm.On("Save", mock.Anything).Return(mockOrm)
	mockOrm.On("Error").Return(nil)
	handler.orm = mockOrm
	tenant := struct {
		ID    string
		Name  string
		Email string
	}{ID: "tenant-1", Name: "Tenant Updated", Email: "tenant2@example.com"}
	err := handler.Update(context.Background(), "tenant-1", tenant)
	assert.NoError(t, err)
	mockOrm.AssertExpectations(t)
}

func TestTenantHandler_Update_Error(t *testing.T) {
	orm := &MockORM{}
	handler := NewTenantHandler(orm)
	mockOrm := &MockORM{}
	mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
	mockOrm.On("First", mock.Anything, mock.Anything).Return(mockOrm)
	mockOrm.On("Error").Return(errors.New("db error"))
	handler.orm = mockOrm
	tenant := struct {
		ID    string
		Name  string
		Email string
	}{ID: "tenant-1", Name: "Tenant Updated", Email: "tenant2@example.com"}
	err := handler.Update(context.Background(), "tenant-1", tenant)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fetching existing tenant")
	mockOrm.AssertExpectations(t)
}

func TestTenantHandler_extractTenantFields(t *testing.T) {
	orm := &MockORM{}
	handler := NewTenantHandler(orm)

	deletedAt := time.Date(2025, time.July, 5, 2, 5, 57, 34788000, time.Local)

	avroTenant := &avro.AvroTenant{
		ID:          "tenant-1",
		Version:     2,
		Name:        "Tenant",
		Email:       "tenant@example.com",
		Description: "desc",
		IsActive:    true,
		CreatedAt:   deletedAt,
		UpdatedAt:   deletedAt,
		DeletedAt:   &deletedAt,
	}

	result := handler.extractTenantFields(avroTenant)

	assert.Equal(t, "tenant-1", result.ID)
	assert.Equal(t, 2, result.Version)
	assert.Equal(t, "Tenant", result.Name)
	assert.Equal(t, "tenant@example.com", result.Email)
	assert.Equal(t, "desc", result.Description)
	assert.Equal(t, true, result.IsActive)
	assert.Equal(t, deletedAt, result.CreatedAt)
	assert.Equal(t, deletedAt, result.UpdatedAt)
	assert.Equal(t, &deletedAt, result.DeletedAt)
}

package dto

import (
	"testing"
	"time"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/stretchr/testify/assert"
)

func TestCommandStatusUpdateDTO_ToDomain(t *testing.T) {
	// Create a test DTO
	dto := CommandStatusUpdateDTO{
		CommandID:    "test-command-123",
		DeviceName:   "test-device",
		Status:       "queued",
		ErrorMessage: nil,
		Timestamp:    time.Now(),
	}

	// Convert to domain
	domainObj := dto.ToDomain()

	// Verify conversion
	assert.Equal(t, "test-command-123", domainObj.CommandID)
	assert.Equal(t, "test-device", domainObj.DeviceName)
	assert.Equal(t, domain.CommandStatusQueued, domainObj.Status)
	assert.Nil(t, domainObj.ErrorMessage)
	assert.Equal(t, dto.Timestamp, domainObj.Timestamp)
}

func TestCommandStatusUpdateDTO_FromDomain(t *testing.T) {
	// Create a test domain object
	domainObj := domain.CommandStatusUpdate{
		CommandID:    "test-command-456",
		DeviceName:   "test-device-2",
		Status:       domain.CommandStatusSent,
		ErrorMessage: nil,
		Timestamp:    time.Now(),
	}

	// Convert to DTO
	dto := FromDomain(domainObj)

	// Verify conversion
	assert.Equal(t, "test-command-456", dto.CommandID)
	assert.Equal(t, "test-device-2", dto.DeviceName)
	assert.Equal(t, "confirmed", dto.Status)
	assert.Nil(t, dto.ErrorMessage)
	assert.Equal(t, domainObj.Timestamp, dto.Timestamp)
}

func TestCommandStatusUpdateDTO_WithErrorMessage(t *testing.T) {
	errorMsg := "command failed"

	// Create a test DTO with error message
	dto := CommandStatusUpdateDTO{
		CommandID:    "test-command-789",
		DeviceName:   "test-device-3",
		Status:       "failed",
		ErrorMessage: &errorMsg,
		Timestamp:    time.Now(),
	}

	// Convert to domain
	domainObj := dto.ToDomain()

	// Verify conversion
	assert.Equal(t, "test-command-789", domainObj.CommandID)
	assert.Equal(t, "test-device-3", domainObj.DeviceName)
	assert.Equal(t, domain.CommandStatusFailed, domainObj.Status)
	assert.Equal(t, &errorMsg, domainObj.ErrorMessage)
	assert.Equal(t, dto.Timestamp, domainObj.Timestamp)
}

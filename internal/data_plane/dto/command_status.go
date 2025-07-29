package dto

import (
	"time"
	"zensor-server/internal/shared_kernel/domain"
)

// CommandStatusUpdateDTO represents a command status change event for JSON serialization
type CommandStatusUpdateDTO struct {
	CommandID    string    `json:"command_id,omitempty"`
	DeviceName   string    `json:"device_name"`
	Status       string    `json:"status"`
	ErrorMessage *string   `json:"error_message,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

// ToDomain converts CommandStatusUpdateDTO to domain.CommandStatusUpdate
func (dto CommandStatusUpdateDTO) ToDomain() domain.CommandStatusUpdate {
	return domain.CommandStatusUpdate{
		CommandID:    dto.CommandID,
		DeviceName:   dto.DeviceName,
		Status:       domain.CommandStatus(dto.Status),
		ErrorMessage: dto.ErrorMessage,
		Timestamp:    dto.Timestamp,
	}
}

// FromDomain converts domain.CommandStatusUpdate to CommandStatusUpdateDTO
func FromDomain(statusUpdate domain.CommandStatusUpdate) CommandStatusUpdateDTO {
	return CommandStatusUpdateDTO{
		CommandID:    statusUpdate.CommandID,
		DeviceName:   statusUpdate.DeviceName,
		Status:       string(statusUpdate.Status),
		ErrorMessage: statusUpdate.ErrorMessage,
		Timestamp:    statusUpdate.Timestamp,
	}
}

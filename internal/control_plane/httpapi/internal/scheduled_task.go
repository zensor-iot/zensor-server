package internal

import (
	"time"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"
)

// SchedulingConfigurationRequest represents the scheduling configuration in API requests
type SchedulingConfigurationRequest struct {
	Type          string     `json:"type"`                     // "cron" or "interval"
	Schedule      *string    `json:"schedule,omitempty"`       // Cron expression (for cron type)
	InitialDay    *time.Time `json:"initial_day,omitempty"`    // Starting day for interval scheduling
	DayInterval   *int       `json:"day_interval,omitempty"`   // Days between executions (for interval scheduling)
	ExecutionTime *string    `json:"execution_time,omitempty"` // Time of day (e.g., "02:00", "14:30")
}

// SchedulingConfigurationResponse represents the scheduling configuration in API responses
type SchedulingConfigurationResponse struct {
	Type          string     `json:"type"`                     // "cron" or "interval"
	Schedule      *string    `json:"schedule,omitempty"`       // Cron expression (for cron type)
	InitialDay    *time.Time `json:"initial_day,omitempty"`    // Starting day for interval scheduling
	DayInterval   *int       `json:"day_interval,omitempty"`   // Days between executions (for interval scheduling)
	ExecutionTime *string    `json:"execution_time,omitempty"` // Time of day (e.g., "02:00", "14:30")
	NextExecution *time.Time `json:"next_execution,omitempty"` // Calculated next execution time
}

type ScheduledTaskCreateRequest struct {
	Commands   []CommandSendPayloadRequest     `json:"commands"`
	Schedule   string                          `json:"schedule,omitempty"` // Deprecated: use Scheduling instead
	Scheduling *SchedulingConfigurationRequest `json:"scheduling,omitempty"`
	IsActive   bool                            `json:"is_active"`
}

type ScheduledTaskUpdateRequest struct {
	Commands   *[]CommandSendPayloadRequest    `json:"commands,omitempty"`
	Schedule   *string                         `json:"schedule,omitempty"` // Deprecated: use Scheduling instead
	Scheduling *SchedulingConfigurationRequest `json:"scheduling,omitempty"`
	IsActive   *bool                           `json:"is_active,omitempty"`
}

type ScheduledTaskResponse struct {
	ID         string                           `json:"id"`
	DeviceID   string                           `json:"device_id"`
	Commands   []CommandSendPayloadRequest      `json:"commands"`
	Schedule   string                           `json:"schedule,omitempty"` // Deprecated: use Scheduling instead
	Scheduling *SchedulingConfigurationResponse `json:"scheduling,omitempty"`
	IsActive   bool                             `json:"is_active"`
}

type ScheduledTaskListResponse struct {
	ScheduledTasks []ScheduledTaskResponse `json:"scheduled_tasks"`
}

// ToSchedulingConfiguration converts a request to domain SchedulingConfiguration
func (req *SchedulingConfigurationRequest) ToSchedulingConfiguration() domain.SchedulingConfiguration {
	config := domain.SchedulingConfiguration{
		Type: domain.SchedulingType(req.Type),
	}

	// Note: For cron type, the schedule field is handled in the controller
	// by setting the Schedule field on the ScheduledTask domain object

	if req.InitialDay != nil {
		config.InitialDay = &utils.Time{Time: *req.InitialDay}
	}

	if req.DayInterval != nil {
		config.DayInterval = req.DayInterval
	}

	if req.ExecutionTime != nil {
		config.ExecutionTime = req.ExecutionTime
	}

	return config
}

// FromSchedulingConfiguration converts domain SchedulingConfiguration to response
func FromSchedulingConfiguration(config domain.SchedulingConfiguration, nextExecution *time.Time) *SchedulingConfigurationResponse {
	resp := &SchedulingConfigurationResponse{
		Type: string(config.Type),
	}

	if config.InitialDay != nil {
		resp.InitialDay = &config.InitialDay.Time
	}

	if config.DayInterval != nil {
		resp.DayInterval = config.DayInterval
	}

	if config.ExecutionTime != nil {
		resp.ExecutionTime = config.ExecutionTime
	}

	if nextExecution != nil {
		resp.NextExecution = nextExecution
	}

	return resp
}

package internal

import (
	"encoding/json"
	"time"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"
)

// SchedulingConfigurationData represents the scheduling configuration
// that should be stored in the database
type SchedulingConfigurationData struct {
	Type          string  `json:"type"`
	InitialDay    *string `json:"initial_day,omitempty"`    // RFC3339 formatted time
	DayInterval   *int    `json:"day_interval,omitempty"`   // Days between executions
	ExecutionTime *string `json:"execution_time,omitempty"` // Time of day (e.g., "02:00", "14:30")
}

// CommandTemplateData represents the essential command template information
// that should be stored in the database, without the full device object
type CommandTemplateData struct {
	Port     uint8  `json:"port"`
	Priority string `json:"priority"`
	Payload  struct {
		Index uint8 `json:"index"`
		Value uint8 `json:"value"`
	} `json:"payload"`
	WaitFor string `json:"wait_for"` // Duration as string (e.g., "5s")
}

// ToCommandTemplateData converts a domain CommandTemplate to CommandTemplateData
func ToCommandTemplateData(template domain.CommandTemplate) CommandTemplateData {
	return CommandTemplateData{
		Port:     uint8(template.Port),
		Priority: string(template.Priority),
		Payload: struct {
			Index uint8 `json:"index"`
			Value uint8 `json:"value"`
		}{
			Index: uint8(template.Payload.Index),
			Value: uint8(template.Payload.Value),
		},
		WaitFor: template.WaitFor.String(),
	}
}

// ToCommand converts a CommandTemplateData to a domain Command with calculated DispatchAfter
func (ctd CommandTemplateData) ToCommand(device domain.Device, task domain.Task, baseTime time.Time) domain.Command {
	waitFor, _ := time.ParseDuration(ctd.WaitFor)
	dispatchAfter := baseTime.Add(waitFor)

	return domain.Command{
		ID:       domain.ID(utils.GenerateUUID()),
		Version:  1,
		Device:   device,
		Task:     task,
		Port:     domain.Port(ctd.Port),
		Priority: domain.CommandPriority(ctd.Priority),
		Payload: domain.CommandPayload{
			Index: domain.Index(ctd.Payload.Index),
			Value: domain.CommandValue(ctd.Payload.Value),
		},
		DispatchAfter: utils.Time{Time: dispatchAfter},
		Ready:         false,
		Sent:          false,
	}
}

type ScheduledTask struct {
	ID               string      `json:"id" gorm:"primaryKey"`
	Version          uint        `json:"version"`
	TenantID         string      `json:"tenant_id"`
	DeviceID         string      `json:"device_id"`
	CommandTemplates string      `json:"command_templates"` // JSON array of command templates
	Schedule         string      `json:"schedule"`          // Deprecated: use SchedulingConfig instead
	SchedulingConfig string      `json:"scheduling_config"` // JSON scheduling configuration
	IsActive         bool        `json:"is_active"`
	CreatedAt        utils.Time  `json:"created_at"`
	UpdatedAt        utils.Time  `json:"updated_at"`
	LastExecutedAt   *utils.Time `json:"last_executed_at"`
	DeletedAt        *utils.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func (ScheduledTask) TableName() string {
	return "scheduled_tasks"
}

func FromScheduledTask(value domain.ScheduledTask) ScheduledTask {
	// Convert domain CommandTemplates to CommandTemplateData
	commandTemplateData := make([]CommandTemplateData, len(value.CommandTemplates))
	for i, template := range value.CommandTemplates {
		commandTemplateData[i] = ToCommandTemplateData(template)
	}

	commandTemplatesJSON, _ := json.Marshal(commandTemplateData)

	// Convert scheduling configuration to persistence format
	var schedulingConfigJSON []byte
	var schedulingConfigStr string
	if value.Scheduling.Type != "" {
		schedulingData := SchedulingConfigurationData{
			Type:          string(value.Scheduling.Type),
			DayInterval:   value.Scheduling.DayInterval,
			ExecutionTime: value.Scheduling.ExecutionTime,
		}

		if value.Scheduling.InitialDay != nil {
			initialDayStr := value.Scheduling.InitialDay.Time.Format(time.RFC3339)
			schedulingData.InitialDay = &initialDayStr
		}

		schedulingConfigJSON, _ = json.Marshal(schedulingData)
		schedulingConfigStr = string(schedulingConfigJSON)
	}

	return ScheduledTask{
		ID:               value.ID.String(),
		Version:          uint(value.Version),
		TenantID:         value.Tenant.ID.String(),
		DeviceID:         value.Device.ID.String(),
		CommandTemplates: string(commandTemplatesJSON),
		Schedule:         value.Schedule,
		SchedulingConfig: schedulingConfigStr,
		IsActive:         value.IsActive,
		CreatedAt:        value.CreatedAt,
		UpdatedAt:        value.UpdatedAt,
		LastExecutedAt:   value.LastExecutedAt,
		DeletedAt:        value.DeletedAt,
	}
}

func (s ScheduledTask) ToDomain() domain.ScheduledTask {
	var commandTemplateData []CommandTemplateData
	json.Unmarshal([]byte(s.CommandTemplates), &commandTemplateData)

	// Convert CommandTemplateData back to domain CommandTemplates
	commandTemplates := make([]domain.CommandTemplate, len(commandTemplateData))
	for i, data := range commandTemplateData {
		// Parse the WaitFor duration from the stored string
		waitFor, _ := time.ParseDuration(data.WaitFor)

		commandTemplates[i] = domain.CommandTemplate{
			Device:   domain.Device{ID: domain.ID(s.DeviceID)}, // Set device from scheduled task
			Port:     domain.Port(data.Port),
			Priority: domain.CommandPriority(data.Priority),
			Payload: domain.CommandPayload{
				Index: domain.Index(data.Payload.Index),
				Value: domain.CommandValue(data.Payload.Value),
			},
			WaitFor: waitFor,
		}
	}

	// Parse scheduling configuration
	var schedulingConfig domain.SchedulingConfiguration
	if s.SchedulingConfig != "" {
		var schedulingData SchedulingConfigurationData
		if err := json.Unmarshal([]byte(s.SchedulingConfig), &schedulingData); err == nil {
			schedulingConfig.Type = domain.SchedulingType(schedulingData.Type)
			schedulingConfig.DayInterval = schedulingData.DayInterval
			schedulingConfig.ExecutionTime = schedulingData.ExecutionTime

			if schedulingData.InitialDay != nil {
				parsedTime, err := time.Parse(time.RFC3339, *schedulingData.InitialDay)
				if err == nil {
					schedulingConfig.InitialDay = &utils.Time{Time: parsedTime}
				}
			}
		}
	}

	return domain.ScheduledTask{
		ID:               domain.ID(s.ID),
		Version:          domain.Version(s.Version),
		Tenant:           domain.Tenant{ID: domain.ID(s.TenantID)},
		Device:           domain.Device{ID: domain.ID(s.DeviceID)},
		CommandTemplates: commandTemplates,
		Schedule:         s.Schedule,
		Scheduling:       schedulingConfig,
		IsActive:         s.IsActive,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
		LastExecutedAt:   s.LastExecutedAt,
		DeletedAt:        s.DeletedAt,
	}
}

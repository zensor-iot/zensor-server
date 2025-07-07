package avro

import (
	"encoding/json"
	"time"

	"zensor-server/internal/control_plane/domain"
)

// Avro-compatible message structs that match the Avro schemas

// AvroCommand represents the Avro-compatible Command message
type AvroCommand struct {
	ID            string    `avro:"id"`
	Version       int       `avro:"version"`
	DeviceName    string    `avro:"device_name"`
	DeviceID      string    `avro:"device_id"`
	TaskID        string    `avro:"task_id"`
	PayloadIndex  int       `avro:"payload_index"`
	PayloadValue  int       `avro:"payload_value"`
	DispatchAfter time.Time `avro:"dispatch_after"`
	Port          int       `avro:"port"`
	Priority      string    `avro:"priority"`
	CreatedAt     time.Time `avro:"created_at"`
	Ready         bool      `avro:"ready"`
	Sent          bool      `avro:"sent"`
	SentAt        time.Time `avro:"sent_at"`
}

// AvroTask represents the Avro-compatible Task message
type AvroTask struct {
	ID              string    `avro:"id"`
	DeviceID        string    `avro:"device_id"`
	ScheduledTaskID *string   `avro:"scheduled_task_id"`
	Version         int64     `avro:"version"`
	CreatedAt       time.Time `avro:"created_at"`
	UpdatedAt       time.Time `avro:"updated_at"`
}

// AvroDevice represents the Avro-compatible Device message
type AvroDevice struct {
	ID                    string     `avro:"id"`
	Version               int        `avro:"version"`
	Name                  string     `avro:"name"`
	DisplayName           string     `avro:"display_name"`
	AppEUI                string     `avro:"app_eui"`
	DevEUI                string     `avro:"dev_eui"`
	AppKey                string     `avro:"app_key"`
	TenantID              *string    `avro:"tenant_id"`
	LastMessageReceivedAt *time.Time `avro:"last_message_received_at"`
	CreatedAt             time.Time  `avro:"created_at"`
	UpdatedAt             time.Time  `avro:"updated_at"`
}

// AvroScheduledTask represents the Avro-compatible ScheduledTask message
type AvroScheduledTask struct {
	ID               string     `avro:"id"`
	Version          int64      `avro:"version"`
	TenantID         string     `avro:"tenant_id"`
	DeviceID         string     `avro:"device_id"`
	CommandTemplates string     `avro:"command_templates"`
	Schedule         string     `avro:"schedule"`
	IsActive         bool       `avro:"is_active"`
	CreatedAt        time.Time  `avro:"created_at"`
	UpdatedAt        time.Time  `avro:"updated_at"`
	LastExecutedAt   *time.Time `avro:"last_executed_at"`
	DeletedAt        *time.Time `avro:"deleted_at"`
}

// AvroTenant represents the Avro-compatible Tenant message
type AvroTenant struct {
	ID          string     `avro:"id"`
	Version     int        `avro:"version"`
	Name        string     `avro:"name"`
	Email       string     `avro:"email"`
	Description string     `avro:"description"`
	IsActive    bool       `avro:"is_active"`
	CreatedAt   time.Time  `avro:"created_at"`
	UpdatedAt   time.Time  `avro:"updated_at"`
	DeletedAt   *time.Time `avro:"deleted_at"`
}

// AvroEvaluationRule represents the Avro-compatible EvaluationRule message
type AvroEvaluationRule struct {
	ID          string    `avro:"id"`
	DeviceID    string    `avro:"device_id"`
	Version     int       `avro:"version"`
	Description string    `avro:"description"`
	Kind        string    `avro:"kind"`
	Enabled     bool      `avro:"enabled"`
	Parameters  string    `avro:"parameters"`
	CreatedAt   time.Time `avro:"created_at"`
	UpdatedAt   time.Time `avro:"updated_at"`
}

// Conversion functions to convert from domain types to Avro types

// ToAvroCommand converts a domain.Command to an AvroCommand for serialization
func ToAvroCommand(cmd domain.Command) *AvroCommand {
	avroCmd := &AvroCommand{
		ID:            string(cmd.ID),
		Version:       int(cmd.Version),
		DeviceName:    cmd.Device.Name,
		DeviceID:      string(cmd.Device.ID),
		TaskID:        string(cmd.Task.ID),
		PayloadIndex:  int(cmd.Payload.Index),
		PayloadValue:  int(cmd.Payload.Value),
		DispatchAfter: cmd.DispatchAfter.Time,
		Port:          int(cmd.Port),
		Priority:      string(cmd.Priority),
		CreatedAt:     time.Now(), // Note: domain.Command doesn't have CreatedAt, using current time
		Ready:         cmd.Ready,
		Sent:          cmd.Sent,
		SentAt:        cmd.SentAt.Time,
	}

	return avroCmd
}

// ToAvroTask converts a domain.Task to AvroTask
func ToAvroTask(task domain.Task) *AvroTask {
	avroTask := &AvroTask{
		ID:        string(task.ID),
		DeviceID:  string(task.Device.ID),
		Version:   int64(task.Version),
		CreatedAt: task.CreatedAt.Time,
		UpdatedAt: task.CreatedAt.Time, // Note: domain.Task doesn't have UpdatedAt, using CreatedAt
	}

	// Handle optional ScheduledTaskID
	if task.ScheduledTask != nil {
		scheduledTaskID := string(task.ScheduledTask.ID)
		avroTask.ScheduledTaskID = &scheduledTaskID
	}

	return avroTask
}

// ToAvroDevice converts a domain.Device to AvroDevice
func ToAvroDevice(device domain.Device) *AvroDevice {
	avroDevice := &AvroDevice{
		ID:          string(device.ID),
		Version:     1, // Note: domain.Device doesn't have Version, using default
		Name:        device.Name,
		DisplayName: device.DisplayName,
		AppEUI:      device.AppEUI,
		DevEUI:      device.DevEUI,
		AppKey:      device.AppKey,
		CreatedAt:   time.Now(), // Note: domain.Device doesn't have CreatedAt, using current time
		UpdatedAt:   time.Now(), // Note: domain.Device doesn't have UpdatedAt, using current time
	}

	// Handle optional TenantID
	if device.TenantID != nil {
		tenantID := string(*device.TenantID)
		avroDevice.TenantID = &tenantID
	}

	// Handle optional LastMessageReceivedAt
	if !device.LastMessageReceivedAt.IsZero() {
		lastMessageTime := device.LastMessageReceivedAt.Time
		avroDevice.LastMessageReceivedAt = &lastMessageTime
	}

	return avroDevice
}

// serializeCommandTemplates converts a slice of CommandTemplate to a JSON string
func serializeCommandTemplates(templates []domain.CommandTemplate) string {
	if len(templates) == 0 {
		return "[]"
	}

	// Create a slice of maps to represent the command templates
	var templateMaps []map[string]any
	for _, template := range templates {
		templateMap := map[string]any{
			"device": map[string]any{
				"id": template.Device.ID.String(),
			},
			"port":     int(template.Port),
			"priority": string(template.Priority),
			"payload": map[string]any{
				"index": int(template.Payload.Index),
				"value": int(template.Payload.Value),
			},
			"wait_for": template.WaitFor.String(),
		}
		templateMaps = append(templateMaps, templateMap)
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(templateMaps)
	if err != nil {
		// Return empty array if marshaling fails
		return "[]"
	}

	return string(jsonData)
}

// ToAvroScheduledTask converts a domain.ScheduledTask to AvroScheduledTask
func ToAvroScheduledTask(scheduledTask domain.ScheduledTask) *AvroScheduledTask {
	avroScheduledTask := &AvroScheduledTask{
		ID:        string(scheduledTask.ID),
		Version:   int64(scheduledTask.Version),
		TenantID:  string(scheduledTask.Tenant.ID),
		DeviceID:  string(scheduledTask.Device.ID),
		Schedule:  scheduledTask.Schedule,
		IsActive:  scheduledTask.IsActive,
		CreatedAt: scheduledTask.CreatedAt.Time,
		UpdatedAt: scheduledTask.UpdatedAt.Time,
	}

	// Handle CommandTemplates serialization
	avroScheduledTask.CommandTemplates = serializeCommandTemplates(scheduledTask.CommandTemplates)

	// Handle optional LastExecutedAt
	if scheduledTask.LastExecutedAt != nil {
		lastExecutedTime := scheduledTask.LastExecutedAt.Time
		avroScheduledTask.LastExecutedAt = &lastExecutedTime
	}

	// Handle optional DeletedAt
	if scheduledTask.DeletedAt != nil {
		deletedTime := scheduledTask.DeletedAt.Time
		avroScheduledTask.DeletedAt = &deletedTime
	}

	return avroScheduledTask
}

// ToAvroTenant converts a domain.Tenant to AvroTenant
func ToAvroTenant(tenant domain.Tenant) *AvroTenant {
	avroTenant := &AvroTenant{
		ID:          string(tenant.ID),
		Version:     tenant.Version,
		Name:        tenant.Name,
		Email:       tenant.Email,
		Description: tenant.Description,
		IsActive:    tenant.IsActive,
		CreatedAt:   tenant.CreatedAt,
		UpdatedAt:   tenant.UpdatedAt,
	}

	// Handle optional DeletedAt
	if tenant.DeletedAt != nil {
		avroTenant.DeletedAt = tenant.DeletedAt
	}

	return avroTenant
}

// ToAvroEvaluationRule converts a domain.EvaluationRule to AvroEvaluationRule
func ToAvroEvaluationRule(evaluationRule domain.EvaluationRule) *AvroEvaluationRule {
	// Serialize parameters to JSON
	paramsJSON, _ := json.Marshal(evaluationRule.Parameters)

	avroEvaluationRule := &AvroEvaluationRule{
		ID:          string(evaluationRule.ID),
		DeviceID:    "", // Note: domain.EvaluationRule doesn't have DeviceID, using empty string
		Version:     int(evaluationRule.Version),
		Description: evaluationRule.Description,
		Kind:        evaluationRule.Kind,
		Enabled:     evaluationRule.Enabled,
		Parameters:  string(paramsJSON),
		CreatedAt:   time.Now(), // Note: domain.EvaluationRule doesn't have CreatedAt, using current time
		UpdatedAt:   time.Now(), // Note: domain.EvaluationRule doesn't have UpdatedAt, using current time
	}

	return avroEvaluationRule
}

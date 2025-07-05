package avro

import (
	"encoding/json"
	"reflect"
	"time"
)

// Avro-compatible message structs that match the Avro schemas

// AvroCommand represents the Avro-compatible Command message
type AvroCommand struct {
	ID            string             `avro:"id"`
	Version       int                `avro:"version"`
	DeviceName    string             `avro:"device_name"`
	DeviceID      string             `avro:"device_id"`
	TaskID        string             `avro:"task_id"`
	Payload       AvroCommandPayload `avro:"payload"`
	DispatchAfter time.Time          `avro:"dispatch_after"`
	Port          int                `avro:"port"`
	Priority      string             `avro:"priority"`
	CreatedAt     time.Time          `avro:"created_at"`
	Ready         bool               `avro:"ready"`
	Sent          bool               `avro:"sent"`
	SentAt        time.Time          `avro:"sent_at"`
}

// AvroCommandPayload represents the Avro-compatible CommandPayload
type AvroCommandPayload struct {
	Index int `avro:"index"`
	Value int `avro:"value"`
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
	DeletedAt        *string    `avro:"deleted_at"`
}

// AvroTenant represents the Avro-compatible Tenant message
type AvroTenant struct {
	ID          string    `avro:"id"`
	Version     int       `avro:"version"`
	Name        string    `avro:"name"`
	Email       string    `avro:"email"`
	Description string    `avro:"description"`
	IsActive    bool      `avro:"is_active"`
	CreatedAt   time.Time `avro:"created_at"`
	UpdatedAt   time.Time `avro:"updated_at"`
	DeletedAt   *string   `avro:"deleted_at"`
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

// ToAvroCommand converts a domain Command to AvroCommand
func ToAvroCommand(cmd interface{}) *AvroCommand {
	// Use reflection to extract fields from the original command
	val := reflect.ValueOf(cmd)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	avroCmd := &AvroCommand{}

	// Extract basic fields
	if idField := val.FieldByName("ID"); idField.IsValid() {
		avroCmd.ID = idField.Interface().(string)
	}
	if versionField := val.FieldByName("Version"); versionField.IsValid() {
		avroCmd.Version = versionField.Interface().(int)
	}
	if deviceNameField := val.FieldByName("DeviceName"); deviceNameField.IsValid() {
		avroCmd.DeviceName = deviceNameField.Interface().(string)
	}
	if deviceIDField := val.FieldByName("DeviceID"); deviceIDField.IsValid() {
		avroCmd.DeviceID = deviceIDField.Interface().(string)
	}
	if taskIDField := val.FieldByName("TaskID"); taskIDField.IsValid() {
		avroCmd.TaskID = taskIDField.Interface().(string)
	}
	if portField := val.FieldByName("Port"); portField.IsValid() {
		avroCmd.Port = int(portField.Interface().(uint8))
	}
	if priorityField := val.FieldByName("Priority"); priorityField.IsValid() {
		avroCmd.Priority = priorityField.Interface().(string)
	}
	if readyField := val.FieldByName("Ready"); readyField.IsValid() {
		avroCmd.Ready = readyField.Interface().(bool)
	}
	if sentField := val.FieldByName("Sent"); sentField.IsValid() {
		avroCmd.Sent = sentField.Interface().(bool)
	}

	// Handle time fields - convert to time.Time
	if dispatchAfterField := val.FieldByName("DispatchAfter"); dispatchAfterField.IsValid() {
		if timeField, ok := dispatchAfterField.Interface().(interface{ Time() time.Time }); ok {
			avroCmd.DispatchAfter = timeField.Time()
		}
	}
	if createdAtField := val.FieldByName("CreatedAt"); createdAtField.IsValid() {
		if timeField, ok := createdAtField.Interface().(interface{ Time() time.Time }); ok {
			avroCmd.CreatedAt = timeField.Time()
		}
	}
	if sentAtField := val.FieldByName("SentAt"); sentAtField.IsValid() {
		if timeField, ok := sentAtField.Interface().(interface{ Time() time.Time }); ok {
			avroCmd.SentAt = timeField.Time()
		}
	}

	// Handle payload
	if payloadField := val.FieldByName("Payload"); payloadField.IsValid() {
		payload := payloadField.Interface()
		payloadVal := reflect.ValueOf(payload)
		if payloadVal.Kind() == reflect.Ptr {
			payloadVal = payloadVal.Elem()
		}

		avroPayload := &AvroCommandPayload{}
		if indexField := payloadVal.FieldByName("Index"); indexField.IsValid() {
			avroPayload.Index = int(indexField.Interface().(uint8))
		}
		if valueField := payloadVal.FieldByName("Value"); valueField.IsValid() {
			avroPayload.Value = int(valueField.Interface().(uint8))
		}
		avroCmd.Payload = *avroPayload
	}

	return avroCmd
}

// ToAvroTask converts a domain Task to AvroTask
func ToAvroTask(task interface{}) *AvroTask {
	val := reflect.ValueOf(task)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	avroTask := &AvroTask{}

	if idField := val.FieldByName("ID"); idField.IsValid() {
		avroTask.ID = idField.Interface().(string)
	}
	if deviceIDField := val.FieldByName("DeviceID"); deviceIDField.IsValid() {
		avroTask.DeviceID = deviceIDField.Interface().(string)
	}
	if scheduledTaskIDField := val.FieldByName("ScheduledTaskID"); scheduledTaskIDField.IsValid() {
		if scheduledTaskID := scheduledTaskIDField.Interface().(string); scheduledTaskID != "" {
			avroTask.ScheduledTaskID = &scheduledTaskID
		}
	}
	if versionField := val.FieldByName("Version"); versionField.IsValid() {
		avroTask.Version = int64(versionField.Interface().(uint))
	}
	if createdAtField := val.FieldByName("CreatedAt"); createdAtField.IsValid() {
		if timeField, ok := createdAtField.Interface().(interface{ Time() time.Time }); ok {
			avroTask.CreatedAt = timeField.Time()
		}
	}
	if updatedAtField := val.FieldByName("UpdatedAt"); updatedAtField.IsValid() {
		if timeField, ok := updatedAtField.Interface().(interface{ Time() time.Time }); ok {
			avroTask.UpdatedAt = timeField.Time()
		}
	}

	return avroTask
}

// ToAvroDevice converts a domain Device to AvroDevice
func ToAvroDevice(device interface{}) *AvroDevice {
	val := reflect.ValueOf(device)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	avroDevice := &AvroDevice{}

	if idField := val.FieldByName("ID"); idField.IsValid() {
		avroDevice.ID = idField.Interface().(string)
	}
	if versionField := val.FieldByName("Version"); versionField.IsValid() {
		avroDevice.Version = versionField.Interface().(int)
	}
	if nameField := val.FieldByName("Name"); nameField.IsValid() {
		avroDevice.Name = nameField.Interface().(string)
	}
	if displayNameField := val.FieldByName("DisplayName"); displayNameField.IsValid() {
		avroDevice.DisplayName = displayNameField.Interface().(string)
	}
	if appEUIField := val.FieldByName("AppEUI"); appEUIField.IsValid() {
		avroDevice.AppEUI = appEUIField.Interface().(string)
	}
	if devEUIField := val.FieldByName("DevEUI"); devEUIField.IsValid() {
		avroDevice.DevEUI = devEUIField.Interface().(string)
	}
	if appKeyField := val.FieldByName("AppKey"); appKeyField.IsValid() {
		avroDevice.AppKey = appKeyField.Interface().(string)
	}
	if tenantIDField := val.FieldByName("TenantID"); tenantIDField.IsValid() {
		if tenantID := tenantIDField.Interface().(*string); tenantID != nil {
			avroDevice.TenantID = tenantID
		}
	}
	if lastMessageReceivedAtField := val.FieldByName("LastMessageReceivedAt"); lastMessageReceivedAtField.IsValid() {
		if timeField, ok := lastMessageReceivedAtField.Interface().(interface{ Time() time.Time }); ok {
			timeVal := timeField.Time()
			avroDevice.LastMessageReceivedAt = &timeVal
		}
	}
	if createdAtField := val.FieldByName("CreatedAt"); createdAtField.IsValid() {
		avroDevice.CreatedAt = createdAtField.Interface().(time.Time)
	}
	if updatedAtField := val.FieldByName("UpdatedAt"); updatedAtField.IsValid() {
		avroDevice.UpdatedAt = updatedAtField.Interface().(time.Time)
	}

	return avroDevice
}

// ToAvroScheduledTask converts a domain ScheduledTask to AvroScheduledTask
func ToAvroScheduledTask(scheduledTask interface{}) *AvroScheduledTask {
	val := reflect.ValueOf(scheduledTask)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	avroScheduledTask := &AvroScheduledTask{}

	if idField := val.FieldByName("ID"); idField.IsValid() {
		avroScheduledTask.ID = idField.Interface().(string)
	}
	if versionField := val.FieldByName("Version"); versionField.IsValid() {
		avroScheduledTask.Version = int64(versionField.Interface().(uint))
	}
	if tenantIDField := val.FieldByName("TenantID"); tenantIDField.IsValid() {
		avroScheduledTask.TenantID = tenantIDField.Interface().(string)
	}
	if deviceIDField := val.FieldByName("DeviceID"); deviceIDField.IsValid() {
		avroScheduledTask.DeviceID = deviceIDField.Interface().(string)
	}
	if commandTemplatesField := val.FieldByName("CommandTemplates"); commandTemplatesField.IsValid() {
		avroScheduledTask.CommandTemplates = commandTemplatesField.Interface().(string)
	}
	if scheduleField := val.FieldByName("Schedule"); scheduleField.IsValid() {
		avroScheduledTask.Schedule = scheduleField.Interface().(string)
	}
	if isActiveField := val.FieldByName("IsActive"); isActiveField.IsValid() {
		avroScheduledTask.IsActive = isActiveField.Interface().(bool)
	}
	if createdAtField := val.FieldByName("CreatedAt"); createdAtField.IsValid() {
		if timeField, ok := createdAtField.Interface().(interface{ Time() time.Time }); ok {
			avroScheduledTask.CreatedAt = timeField.Time()
		}
	}
	if updatedAtField := val.FieldByName("UpdatedAt"); updatedAtField.IsValid() {
		if timeField, ok := updatedAtField.Interface().(interface{ Time() time.Time }); ok {
			avroScheduledTask.UpdatedAt = timeField.Time()
		}
	}
	if lastExecutedAtField := val.FieldByName("LastExecutedAt"); lastExecutedAtField.IsValid() {
		if lastExecutedAt := lastExecutedAtField.Interface().(*interface{ Time() time.Time }); lastExecutedAt != nil {
			if timeField, ok := (*lastExecutedAt).(interface{ Time() time.Time }); ok {
				timeVal := timeField.Time()
				avroScheduledTask.LastExecutedAt = &timeVal
			}
		}
	}
	if deletedAtField := val.FieldByName("DeletedAt"); deletedAtField.IsValid() {
		if deletedAt := deletedAtField.Interface().(*interface{ Time() time.Time }); deletedAt != nil {
			if timeField, ok := (*deletedAt).(interface{ Time() time.Time }); ok {
				timeStr := timeField.Time().Format(time.RFC3339)
				avroScheduledTask.DeletedAt = &timeStr
			}
		}
	}

	return avroScheduledTask
}

// ToAvroTenant converts a domain Tenant to AvroTenant
func ToAvroTenant(tenant interface{}) *AvroTenant {
	val := reflect.ValueOf(tenant)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	avroTenant := &AvroTenant{}

	if idField := val.FieldByName("ID"); idField.IsValid() {
		avroTenant.ID = idField.Interface().(string)
	}
	if versionField := val.FieldByName("Version"); versionField.IsValid() {
		avroTenant.Version = versionField.Interface().(int)
	}
	if nameField := val.FieldByName("Name"); nameField.IsValid() {
		avroTenant.Name = nameField.Interface().(string)
	}
	if emailField := val.FieldByName("Email"); emailField.IsValid() {
		avroTenant.Email = emailField.Interface().(string)
	}
	if descriptionField := val.FieldByName("Description"); descriptionField.IsValid() {
		avroTenant.Description = descriptionField.Interface().(string)
	}
	if isActiveField := val.FieldByName("IsActive"); isActiveField.IsValid() {
		avroTenant.IsActive = isActiveField.Interface().(bool)
	}
	if createdAtField := val.FieldByName("CreatedAt"); createdAtField.IsValid() {
		avroTenant.CreatedAt = createdAtField.Interface().(time.Time)
	}
	if updatedAtField := val.FieldByName("UpdatedAt"); updatedAtField.IsValid() {
		avroTenant.UpdatedAt = updatedAtField.Interface().(time.Time)
	}
	if deletedAtField := val.FieldByName("DeletedAt"); deletedAtField.IsValid() {
		if deletedAt := deletedAtField.Interface().(*time.Time); deletedAt != nil {
			timeStr := deletedAt.Format(time.RFC3339)
			avroTenant.DeletedAt = &timeStr
		}
	}

	return avroTenant
}

// ToAvroEvaluationRule converts a domain EvaluationRule to AvroEvaluationRule
func ToAvroEvaluationRule(evaluationRule interface{}) *AvroEvaluationRule {
	val := reflect.ValueOf(evaluationRule)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	avroEvaluationRule := &AvroEvaluationRule{}

	if idField := val.FieldByName("ID"); idField.IsValid() {
		avroEvaluationRule.ID = idField.Interface().(string)
	}
	if deviceIDField := val.FieldByName("DeviceID"); deviceIDField.IsValid() {
		avroEvaluationRule.DeviceID = deviceIDField.Interface().(string)
	}
	if versionField := val.FieldByName("Version"); versionField.IsValid() {
		avroEvaluationRule.Version = versionField.Interface().(int)
	}
	if descriptionField := val.FieldByName("Description"); descriptionField.IsValid() {
		avroEvaluationRule.Description = descriptionField.Interface().(string)
	}
	if kindField := val.FieldByName("Kind"); kindField.IsValid() {
		avroEvaluationRule.Kind = kindField.Interface().(string)
	}
	if enabledField := val.FieldByName("Enabled"); enabledField.IsValid() {
		avroEvaluationRule.Enabled = enabledField.Interface().(bool)
	}
	if parametersField := val.FieldByName("Parameters"); parametersField.IsValid() {
		// Convert parameters to JSON string
		if params, ok := parametersField.Interface().(map[string]any); ok {
			if jsonData, err := json.Marshal(params); err == nil {
				avroEvaluationRule.Parameters = string(jsonData)
			}
		}
	}
	if createdAtField := val.FieldByName("CreatedAt"); createdAtField.IsValid() {
		if timeField, ok := createdAtField.Interface().(interface{ Time() time.Time }); ok {
			avroEvaluationRule.CreatedAt = timeField.Time()
		}
	}
	if updatedAtField := val.FieldByName("UpdatedAt"); updatedAtField.IsValid() {
		if timeField, ok := updatedAtField.Interface().(interface{ Time() time.Time }); ok {
			avroEvaluationRule.UpdatedAt = timeField.Time()
		}
	}

	return avroEvaluationRule
}

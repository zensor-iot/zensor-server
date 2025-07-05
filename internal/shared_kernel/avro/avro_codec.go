package avro

import (
	"fmt"
	"reflect"
	"time"

	"github.com/hamba/avro/v2"
)

// AvroCodec implements Codec interface using schemas from the schema registry
type AvroCodec struct {
	prototype         any
	schemaRegistry    SchemaRegistry
	schemas           map[string]avro.Schema
	subjectNameSuffix string
}

// NewAvroCodec creates a new Avro codec that fetches schemas from the registry
func NewAvroCodec(prototype any, schemaRegistryURL string) *AvroCodec {
	return &AvroCodec{
		prototype:         prototype,
		schemaRegistry:    NewSchemaRegistryClient(schemaRegistryURL),
		schemas:           make(map[string]avro.Schema),
		subjectNameSuffix: "-value",
	}
}

// getSchemaForMessage returns the appropriate Avro schema for the given message
func (c *AvroCodec) getSchemaForMessage(message any) (avro.Schema, error) {
	messageType := reflect.TypeOf(message)
	if messageType.Kind() == reflect.Ptr {
		messageType = messageType.Elem()
	}

	schemaName := messageType.Name()
	switch schemaName {
	case "Command", "AvroCommand":
		schemaName = "commands"
	case "Task", "AvroTask":
		schemaName = "tasks"
	case "Device", "AvroDevice":
		schemaName = "devices"
	case "ScheduledTask", "AvroScheduledTask":
		schemaName = "scheduled-tasks"
	case "Tenant", "AvroTenant":
		schemaName = "tenants"
	case "EvaluationRule", "AvroEvaluationRule":
		schemaName = "evaluation-rules"
	default:
		return nil, fmt.Errorf("no Avro schema found for message type: %s", schemaName)
	}

	// Check if schema is already cached
	if schema, exists := c.schemas[schemaName]; exists {
		return schema, nil
	}

	// Fetch schema from registry
	subject := schemaName + c.subjectNameSuffix
	schemaInfo, err := c.schemaRegistry.GetLatestSchema(subject)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schema for subject %s: %w", subject, err)
	}

	// Parse the schema
	schema, err := avro.Parse(schemaInfo.Schema)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema for subject %s: %w", subject, err)
	}

	// Cache the schema
	c.schemas[schemaName] = schema

	return schema, nil
}

// Encode encodes a value into Avro format
func (c *AvroCodec) Encode(value any) ([]byte, error) {
	// Convert the original struct to Avro-compatible struct
	avroValue, err := c.convertToAvroStruct(value)
	if err != nil {
		return nil, fmt.Errorf("converting to Avro struct: %w", err)
	}

	schema, err := c.getSchemaForMessage(value)
	if err != nil {
		return nil, fmt.Errorf("getting schema: %w", err)
	}

	data, err := avro.Marshal(schema, avroValue)
	if err != nil {
		return nil, fmt.Errorf("marshaling Avro: %w", err)
	}

	return data, nil
}

// Decode decodes Avro data into a value
func (c *AvroCodec) Decode(data []byte) (any, error) {
	// For decoding, we need to determine the schema from the prototype
	schema, err := c.getSchemaForMessage(c.prototype)
	if err != nil {
		return nil, fmt.Errorf("getting schema: %w", err)
	}

	// Create a new instance of the prototype type
	prototypeType := reflect.TypeOf(c.prototype)
	if prototypeType.Kind() == reflect.Ptr {
		prototypeType = prototypeType.Elem()
	}
	result := reflect.New(prototypeType).Interface()

	err = avro.Unmarshal(schema, data, result)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling Avro: %w", err)
	}

	// Convert from Avro struct back to original struct if needed
	converted, err := c.convertFromAvroStruct(result)
	if err != nil {
		return nil, fmt.Errorf("converting from Avro struct: %w", err)
	}

	return converted, nil
}

// convertToAvroStruct converts any struct to its Avro-compatible version
func (c *AvroCodec) convertToAvroStruct(value any) (any, error) {
	valueType := reflect.TypeOf(value)
	if valueType.Kind() == reflect.Ptr {
		valueType = valueType.Elem()
		value = reflect.ValueOf(value).Elem().Interface()
	}

	// Handle different struct types
	switch v := value.(type) {
	case *AvroCommand, AvroCommand:
		return v, nil
	case *AvroTask, AvroTask:
		return v, nil
	case *AvroDevice, AvroDevice:
		return v, nil
	case *AvroScheduledTask, AvroScheduledTask:
		return v, nil
	case *AvroTenant, AvroTenant:
		return v, nil
	case *AvroEvaluationRule, AvroEvaluationRule:
		return v, nil
	default:
		// Try to convert using reflection
		return c.convertUsingReflection(value)
	}
}

// convertFromAvroStruct converts Avro struct back to original struct
func (c *AvroCodec) convertFromAvroStruct(value any) (any, error) {
	// For now, return the value as-is since we're using Avro structs directly
	return value, nil
}

// convertUsingReflection converts any struct to Avro struct using reflection
func (c *AvroCodec) convertUsingReflection(value any) (any, error) {
	valueType := reflect.TypeOf(value)
	valueValue := reflect.ValueOf(value)

	if valueType.Kind() == reflect.Ptr {
		valueType = valueType.Elem()
		valueValue = valueValue.Elem()
	}

	// Map internal types to Avro types
	switch valueType.Name() {
	case "Command":
		return c.convertCommandToAvro(valueValue)
	case "Task":
		return c.convertTaskToAvro(valueValue)
	case "Device":
		return c.convertDeviceToAvro(valueValue)
	case "ScheduledTask":
		return c.convertScheduledTaskToAvro(valueValue)
	case "Tenant":
		return c.convertTenantToAvro(valueValue)
	case "EvaluationRule":
		return c.convertEvaluationRuleToAvro(valueValue)
	default:
		return nil, fmt.Errorf("unsupported type for conversion: %s", valueType.Name())
	}
}

// Helper conversion functions
func (c *AvroCodec) convertCommandToAvro(v reflect.Value) (*AvroCommand, error) {
	cmd := &AvroCommand{}

	if id := v.FieldByName("ID"); id.IsValid() {
		cmd.ID = id.String()
	}
	if version := v.FieldByName("Version"); version.IsValid() {
		cmd.Version = int(version.Int())
	}
	if deviceName := v.FieldByName("DeviceName"); deviceName.IsValid() {
		cmd.DeviceName = deviceName.String()
	}
	if deviceID := v.FieldByName("DeviceID"); deviceID.IsValid() {
		cmd.DeviceID = deviceID.String()
	}
	if taskID := v.FieldByName("TaskID"); taskID.IsValid() {
		cmd.TaskID = taskID.String()
	}
	if payload := v.FieldByName("Payload"); payload.IsValid() {
		cmd.Payload = AvroCommandPayload{
			Index: int(payload.FieldByName("Index").Int()),
			Value: int(payload.FieldByName("Value").Int()),
		}
	}
	if dispatchAfter := v.FieldByName("DispatchAfter"); dispatchAfter.IsValid() {
		cmd.DispatchAfter = dispatchAfter.Interface().(time.Time)
	}
	if port := v.FieldByName("Port"); port.IsValid() {
		cmd.Port = int(port.Int())
	}
	if priority := v.FieldByName("Priority"); priority.IsValid() {
		cmd.Priority = priority.String()
	}
	if createdAt := v.FieldByName("CreatedAt"); createdAt.IsValid() {
		cmd.CreatedAt = createdAt.Interface().(time.Time)
	}
	if ready := v.FieldByName("Ready"); ready.IsValid() {
		cmd.Ready = ready.Bool()
	}
	if sent := v.FieldByName("Sent"); sent.IsValid() {
		cmd.Sent = sent.Bool()
	}
	if sentAt := v.FieldByName("SentAt"); sentAt.IsValid() {
		cmd.SentAt = sentAt.Interface().(time.Time)
	}

	return cmd, nil
}

func (c *AvroCodec) convertTaskToAvro(v reflect.Value) (*AvroTask, error) {
	task := &AvroTask{}

	if id := v.FieldByName("ID"); id.IsValid() {
		task.ID = id.String()
	}
	if deviceID := v.FieldByName("DeviceID"); deviceID.IsValid() {
		task.DeviceID = deviceID.String()
	}
	if scheduledTaskID := v.FieldByName("ScheduledTaskID"); scheduledTaskID.IsValid() {
		if scheduledTaskID.IsNil() {
			task.ScheduledTaskID = nil
		} else {
			val := scheduledTaskID.String()
			task.ScheduledTaskID = &val
		}
	}
	if version := v.FieldByName("Version"); version.IsValid() {
		task.Version = version.Int()
	}
	if createdAt := v.FieldByName("CreatedAt"); createdAt.IsValid() {
		task.CreatedAt = createdAt.Interface().(time.Time)
	}
	if updatedAt := v.FieldByName("UpdatedAt"); updatedAt.IsValid() {
		task.UpdatedAt = updatedAt.Interface().(time.Time)
	}

	return task, nil
}

func (c *AvroCodec) convertDeviceToAvro(v reflect.Value) (*AvroDevice, error) {
	device := &AvroDevice{}

	if id := v.FieldByName("ID"); id.IsValid() {
		device.ID = id.String()
	}
	if version := v.FieldByName("Version"); version.IsValid() {
		device.Version = int(version.Int())
	}
	if name := v.FieldByName("Name"); name.IsValid() {
		device.Name = name.String()
	}
	if displayName := v.FieldByName("DisplayName"); displayName.IsValid() {
		device.DisplayName = displayName.String()
	}
	if appEUI := v.FieldByName("AppEUI"); appEUI.IsValid() {
		device.AppEUI = appEUI.String()
	}
	if devEUI := v.FieldByName("DevEUI"); devEUI.IsValid() {
		device.DevEUI = devEUI.String()
	}
	if appKey := v.FieldByName("AppKey"); appKey.IsValid() {
		device.AppKey = appKey.String()
	}
	if tenantID := v.FieldByName("TenantID"); tenantID.IsValid() {
		if tenantID.IsNil() {
			device.TenantID = nil
		} else {
			val := tenantID.String()
			device.TenantID = &val
		}
	}
	if lastMessageReceivedAt := v.FieldByName("LastMessageReceivedAt"); lastMessageReceivedAt.IsValid() {
		if lastMessageReceivedAt.IsNil() {
			device.LastMessageReceivedAt = nil
		} else {
			val := lastMessageReceivedAt.Interface().(time.Time)
			device.LastMessageReceivedAt = &val
		}
	}
	if createdAt := v.FieldByName("CreatedAt"); createdAt.IsValid() {
		device.CreatedAt = createdAt.Interface().(time.Time)
	}
	if updatedAt := v.FieldByName("UpdatedAt"); updatedAt.IsValid() {
		device.UpdatedAt = updatedAt.Interface().(time.Time)
	}

	return device, nil
}

func (c *AvroCodec) convertScheduledTaskToAvro(v reflect.Value) (*AvroScheduledTask, error) {
	scheduledTask := &AvroScheduledTask{}

	if id := v.FieldByName("ID"); id.IsValid() {
		scheduledTask.ID = id.String()
	}
	if version := v.FieldByName("Version"); version.IsValid() {
		scheduledTask.Version = version.Int()
	}
	if tenantID := v.FieldByName("TenantID"); tenantID.IsValid() {
		scheduledTask.TenantID = tenantID.String()
	}
	if deviceID := v.FieldByName("DeviceID"); deviceID.IsValid() {
		scheduledTask.DeviceID = deviceID.String()
	}
	if commandTemplates := v.FieldByName("CommandTemplates"); commandTemplates.IsValid() {
		scheduledTask.CommandTemplates = commandTemplates.String()
	}
	if schedule := v.FieldByName("Schedule"); schedule.IsValid() {
		scheduledTask.Schedule = schedule.String()
	}
	if isActive := v.FieldByName("IsActive"); isActive.IsValid() {
		scheduledTask.IsActive = isActive.Bool()
	}
	if createdAt := v.FieldByName("CreatedAt"); createdAt.IsValid() {
		scheduledTask.CreatedAt = createdAt.Interface().(time.Time)
	}
	if updatedAt := v.FieldByName("UpdatedAt"); updatedAt.IsValid() {
		scheduledTask.UpdatedAt = updatedAt.Interface().(time.Time)
	}
	if lastExecutedAt := v.FieldByName("LastExecutedAt"); lastExecutedAt.IsValid() {
		if lastExecutedAt.IsNil() {
			scheduledTask.LastExecutedAt = nil
		} else {
			val := lastExecutedAt.Interface().(time.Time)
			scheduledTask.LastExecutedAt = &val
		}
	}
	if deletedAt := v.FieldByName("DeletedAt"); deletedAt.IsValid() {
		if deletedAt.IsNil() {
			scheduledTask.DeletedAt = nil
		} else {
			val := deletedAt.String()
			scheduledTask.DeletedAt = &val
		}
	}

	return scheduledTask, nil
}

func (c *AvroCodec) convertTenantToAvro(v reflect.Value) (*AvroTenant, error) {
	tenant := &AvroTenant{}

	if id := v.FieldByName("ID"); id.IsValid() {
		tenant.ID = id.String()
	}
	if version := v.FieldByName("Version"); version.IsValid() {
		tenant.Version = int(version.Int())
	}
	if name := v.FieldByName("Name"); name.IsValid() {
		tenant.Name = name.String()
	}
	if email := v.FieldByName("Email"); email.IsValid() {
		tenant.Email = email.String()
	}
	if description := v.FieldByName("Description"); description.IsValid() {
		tenant.Description = description.String()
	}
	if isActive := v.FieldByName("IsActive"); isActive.IsValid() {
		tenant.IsActive = isActive.Bool()
	}
	if createdAt := v.FieldByName("CreatedAt"); createdAt.IsValid() {
		tenant.CreatedAt = createdAt.Interface().(time.Time)
	}
	if updatedAt := v.FieldByName("UpdatedAt"); updatedAt.IsValid() {
		tenant.UpdatedAt = updatedAt.Interface().(time.Time)
	}
	if deletedAt := v.FieldByName("DeletedAt"); deletedAt.IsValid() {
		if deletedAt.IsNil() {
			tenant.DeletedAt = nil
		} else {
			val := deletedAt.String()
			tenant.DeletedAt = &val
		}
	}

	return tenant, nil
}

func (c *AvroCodec) convertEvaluationRuleToAvro(v reflect.Value) (*AvroEvaluationRule, error) {
	evaluationRule := &AvroEvaluationRule{}

	if id := v.FieldByName("ID"); id.IsValid() {
		evaluationRule.ID = id.String()
	}
	if deviceID := v.FieldByName("DeviceID"); deviceID.IsValid() {
		evaluationRule.DeviceID = deviceID.String()
	}
	if version := v.FieldByName("Version"); version.IsValid() {
		evaluationRule.Version = int(version.Int())
	}
	if description := v.FieldByName("Description"); description.IsValid() {
		evaluationRule.Description = description.String()
	}
	if kind := v.FieldByName("Kind"); kind.IsValid() {
		evaluationRule.Kind = kind.String()
	}
	if enabled := v.FieldByName("Enabled"); enabled.IsValid() {
		evaluationRule.Enabled = enabled.Bool()
	}
	if parameters := v.FieldByName("Parameters"); parameters.IsValid() {
		evaluationRule.Parameters = parameters.String()
	}
	if createdAt := v.FieldByName("CreatedAt"); createdAt.IsValid() {
		evaluationRule.CreatedAt = createdAt.Interface().(time.Time)
	}
	if updatedAt := v.FieldByName("UpdatedAt"); updatedAt.IsValid() {
		evaluationRule.UpdatedAt = updatedAt.Interface().(time.Time)
	}

	return evaluationRule, nil
}

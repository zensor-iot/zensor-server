package avro

import (
	"fmt"
	"reflect"

	"github.com/hamba/avro/v2"
)

// AvroCodec implements Codec interface using static Avro schemas
type AvroCodec struct {
	prototype any
	schemas   map[string]avro.Schema
}

// Static Avro schemas for all message types
const (
	// Command schema
	commandSchema = `{
		"type": "record",
		"name": "Command",
		"fields": [
			{"name": "id", "type": "string"},
			{"name": "version", "type": "int"},
			{"name": "device_name", "type": "string"},
			{"name": "device_id", "type": "string"},
			{"name": "task_id", "type": "string"},
			{"name": "payload", "type": {
				"type": "record",
				"name": "CommandPayload",
				"fields": [
					{"name": "index", "type": "int"},
					{"name": "value", "type": "int"}
				]
			}},
			{"name": "dispatch_after", "type": {"type": "long", "logicalType": "timestamp-millis"}},
			{"name": "port", "type": "int"},
			{"name": "priority", "type": "string"},
			{"name": "created_at", "type": {"type": "long", "logicalType": "timestamp-millis"}},
			{"name": "ready", "type": "boolean"},
			{"name": "sent", "type": "boolean"},
			{"name": "sent_at", "type": {"type": "long", "logicalType": "timestamp-millis"}}
		]
	}`

	// Task schema
	taskSchema = `{
		"type": "record",
		"name": "Task",
		"fields": [
			{"name": "id", "type": "string"},
			{"name": "device_id", "type": "string"},
			{"name": "scheduled_task_id", "type": ["null", "string"]},
			{"name": "version", "type": "long"},
			{"name": "created_at", "type": {"type": "long", "logicalType": "timestamp-millis"}},
			{"name": "updated_at", "type": {"type": "long", "logicalType": "timestamp-millis"}}
		]
	}`

	// Device schema
	deviceSchema = `{
		"type": "record",
		"name": "Device",
		"fields": [
			{"name": "id", "type": "string"},
			{"name": "version", "type": "int"},
			{"name": "name", "type": "string"},
			{"name": "display_name", "type": "string"},
			{"name": "app_eui", "type": "string"},
			{"name": "dev_eui", "type": "string"},
			{"name": "app_key", "type": "string"},
			{"name": "tenant_id", "type": ["null", "string"]},
			{"name": "last_message_received_at", "type": ["null", {"type": "long", "logicalType": "timestamp-millis"}]},
			{"name": "created_at", "type": {"type": "long", "logicalType": "timestamp-millis"}},
			{"name": "updated_at", "type": {"type": "long", "logicalType": "timestamp-millis"}}
		]
	}`

	// ScheduledTask schema
	scheduledTaskSchema = `{
		"type": "record",
		"name": "ScheduledTask",
		"fields": [
			{"name": "id", "type": "string"},
			{"name": "version", "type": "long"},
			{"name": "tenant_id", "type": "string"},
			{"name": "device_id", "type": "string"},
			{"name": "command_templates", "type": "string"},
			{"name": "schedule", "type": "string"},
			{"name": "is_active", "type": "boolean"},
			{"name": "created_at", "type": {"type": "long", "logicalType": "timestamp-millis"}},
			{"name": "updated_at", "type": {"type": "long", "logicalType": "timestamp-millis"}},
			{"name": "last_executed_at", "type": ["null", {"type": "long", "logicalType": "timestamp-millis"}]},
			{"name": "deleted_at", "type": ["null", "string"]}
		]
	}`

	// Tenant schema
	tenantSchema = `{
		"type": "record",
		"name": "Tenant",
		"fields": [
			{"name": "id", "type": "string"},
			{"name": "version", "type": "int"},
			{"name": "name", "type": "string"},
			{"name": "email", "type": "string"},
			{"name": "description", "type": "string"},
			{"name": "is_active", "type": "boolean"},
			{"name": "created_at", "type": {"type": "long", "logicalType": "timestamp-millis"}},
			{"name": "updated_at", "type": {"type": "long", "logicalType": "timestamp-millis"}},
			{"name": "deleted_at", "type": ["null", "string"]}
		]
	}`

	// EvaluationRule schema
	evaluationRuleSchema = `{
		"type": "record",
		"name": "EvaluationRule",
		"fields": [
			{"name": "id", "type": "string"},
			{"name": "device_id", "type": "string"},
			{"name": "version", "type": "int"},
			{"name": "description", "type": "string"},
			{"name": "kind", "type": "string"},
			{"name": "enabled", "type": "boolean"},
			{"name": "parameters", "type": "string"},
			{"name": "created_at", "type": {"type": "long", "logicalType": "timestamp-millis"}},
			{"name": "updated_at", "type": {"type": "long", "logicalType": "timestamp-millis"}}
		]
	}`
)

// newAvroCodec creates a new Avro codec with static schemas
func NewAvroCodec(prototype any) *AvroCodec {
	schemas := make(map[string]avro.Schema)

	// Parse all schemas
	commandAvroSchema, _ := avro.Parse(commandSchema)
	taskAvroSchema, _ := avro.Parse(taskSchema)
	deviceAvroSchema, _ := avro.Parse(deviceSchema)
	scheduledTaskAvroSchema, _ := avro.Parse(scheduledTaskSchema)
	tenantAvroSchema, _ := avro.Parse(tenantSchema)
	evaluationRuleAvroSchema, _ := avro.Parse(evaluationRuleSchema)

	schemas["Command"] = commandAvroSchema
	schemas["Task"] = taskAvroSchema
	schemas["Device"] = deviceAvroSchema
	schemas["ScheduledTask"] = scheduledTaskAvroSchema
	schemas["Tenant"] = tenantAvroSchema
	schemas["EvaluationRule"] = evaluationRuleAvroSchema

	return &AvroCodec{
		prototype: prototype,
		schemas:   schemas,
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
		schemaName = "Command"
	case "Task", "AvroTask":
		schemaName = "Task"
	case "Device", "AvroDevice":
		schemaName = "Device"
	case "ScheduledTask", "AvroScheduledTask":
		schemaName = "ScheduledTask"
	case "Tenant", "AvroTenant":
		schemaName = "Tenant"
	case "EvaluationRule", "AvroEvaluationRule":
		schemaName = "EvaluationRule"
	default:
		return nil, fmt.Errorf("no Avro schema found for message type: %s", schemaName)
	}

	schema, exists := c.schemas[schemaName]
	if !exists {
		return nil, fmt.Errorf("no Avro schema found for message type: %s", schemaName)
	}

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
		return nil, fmt.Errorf("marshaling to Avro: %w", err)
	}

	return data, nil
}

// Decode decodes an Avro message back to the original value
func (c *AvroCodec) Decode(data []byte) (any, error) {
	// For decoding, we need to determine the schema from the prototype
	prototypeType := reflect.TypeOf(c.prototype)
	if prototypeType.Kind() == reflect.Ptr {
		prototypeType = prototypeType.Elem()
	}

	schemaName := prototypeType.Name()
	switch schemaName {
	case "Command", "AvroCommand":
		schemaName = "Command"
	case "Task", "AvroTask":
		schemaName = "Task"
	case "Device", "AvroDevice":
		schemaName = "Device"
	case "ScheduledTask", "AvroScheduledTask":
		schemaName = "ScheduledTask"
	case "Tenant", "AvroTenant":
		schemaName = "Tenant"
	case "EvaluationRule", "AvroEvaluationRule":
		schemaName = "EvaluationRule"
	default:
		return nil, fmt.Errorf("no Avro schema found for prototype type: %s", schemaName)
	}

	schema, exists := c.schemas[schemaName]
	if !exists {
		return nil, fmt.Errorf("no Avro schema found for prototype type: %s", schemaName)
	}

	// Create a new instance of the prototype type
	instance := reflect.New(prototypeType).Interface()

	err := avro.Unmarshal(schema, data, instance)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling from Avro: %w", err)
	}

	return instance, nil
}

// convertToAvroStruct converts an original struct to its Avro-compatible equivalent
func (c *AvroCodec) convertToAvroStruct(value any) (any, error) {
	valueType := reflect.TypeOf(value)
	if valueType.Kind() == reflect.Ptr {
		valueType = valueType.Elem()
	}

	switch valueType.Name() {
	case "AvroCommand", "AvroTask", "AvroDevice", "AvroScheduledTask", "AvroTenant", "AvroEvaluationRule":
		return value, nil
	case "Command":
		return ToAvroCommand(value), nil
	case "Task":
		return ToAvroTask(value), nil
	case "Device":
		return ToAvroDevice(value), nil
	case "ScheduledTask":
		return ToAvroScheduledTask(value), nil
	case "Tenant":
		return ToAvroTenant(value), nil
	case "EvaluationRule":
		return ToAvroEvaluationRule(value), nil
	default:
		return nil, fmt.Errorf("unsupported message type for Avro conversion: %s", valueType.Name())
	}
}

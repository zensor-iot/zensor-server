package avro

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"time"

	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel"

	"github.com/linkedin/goavro/v2"
	"github.com/riferrei/srclient"
)

// SchemaRegistry defines the interface for schema registry operations
type SchemaRegistry interface {
	GetLatestSchema(subject string) (*srclient.Schema, error)
	CreateSchema(subject string, schema string, schemaType srclient.SchemaType, references ...srclient.Reference) (*srclient.Schema, error)
	GetSchema(schemaID int) (*srclient.Schema, error)
}

// ConfluentAvroCodec implements Codec interface using Confluent Avro wire format and Schema Registry
type ConfluentAvroCodec struct {
	prototype      any
	schemas        map[string]string
	codecs         map[string]*goavro.Codec
	schemaRegistry SchemaRegistry
	subjectToID    map[string]int
	idToCodec      map[int]*goavro.Codec
	subjectSuffix  string
}

// NewConfluentAvroCodec creates a new Confluent Avro codec with schema registry
func NewConfluentAvroCodec(prototype any, schemaRegistry SchemaRegistry) *ConfluentAvroCodec {
	return &ConfluentAvroCodec{
		prototype:      prototype,
		schemas:        make(map[string]string),
		codecs:         make(map[string]*goavro.Codec),
		schemaRegistry: schemaRegistry,
		subjectToID:    make(map[string]int),
		idToCodec:      make(map[int]*goavro.Codec),
		subjectSuffix:  "-value",
	}
}

// getSchemaForMessage returns the Avro schema name for the given message
func (c *ConfluentAvroCodec) getSchemaForMessage(message any) (string, error) {
	messageType := reflect.TypeOf(message)
	if messageType.Kind() == reflect.Ptr {
		messageType = messageType.Elem()
	}

	schemaName := messageType.Name()
	switch schemaName {
	case "Command", "AvroCommand":
		return "commands", nil
	case "Task", "AvroTask":
		return "tasks", nil
	case "Device", "AvroDevice":
		return "devices", nil
	case "ScheduledTask", "AvroScheduledTask":
		return "scheduled-tasks", nil
	case "Tenant", "AvroTenant":
		return "tenants", nil
	case "EvaluationRule", "AvroEvaluationRule":
		return "evaluation-rules", nil
	default:
		return "", fmt.Errorf("no Avro schema found for message type: %s", schemaName)
	}
}

// getOrRegisterSchemaID gets or registers the schema in the registry and returns its ID
func (c *ConfluentAvroCodec) getOrRegisterSchemaID(schemaName string) (int, error) {
	subject := schemaName + c.subjectSuffix
	if id, ok := c.subjectToID[subject]; ok {
		return id, nil
	}

	// Try to get existing schema from registry
	registered, err := c.schemaRegistry.GetLatestSchema(subject)
	if err == nil && registered != nil {
		c.subjectToID[subject] = registered.ID()
		c.schemas[schemaName] = registered.Schema()

		// Create codec for this schema
		codec, err := goavro.NewCodec(registered.Schema())
		if err == nil {
			c.idToCodec[registered.ID()] = codec
			c.codecs[schemaName] = codec
		}
		return registered.ID(), nil
	}

	// If schema doesn't exist, we need to load it from the schemas folder
	schema, err := c.loadSchemaFromFile(schemaName)
	if err != nil {
		return 0, fmt.Errorf("loading schema from file: %w", err)
	}

	// Register new schema
	newSchema, err := c.schemaRegistry.CreateSchema(subject, schema, srclient.Avro)
	if err != nil {
		return 0, fmt.Errorf("registering schema: %w", err)
	}

	c.subjectToID[subject] = newSchema.ID()
	c.schemas[schemaName] = schema

	codec, err := goavro.NewCodec(schema)
	if err == nil {
		c.idToCodec[newSchema.ID()] = codec
		c.codecs[schemaName] = codec
	}

	return newSchema.ID(), nil
}

// getCodecByID fetches the codec for a schema ID from the registry if not cached
func (c *ConfluentAvroCodec) getCodecByID(schemaID int) (*goavro.Codec, error) {
	if codec, ok := c.idToCodec[schemaID]; ok {
		return codec, nil
	}
	schema, err := c.schemaRegistry.GetSchema(schemaID)
	if err != nil {
		return nil, fmt.Errorf("fetching schema from registry: %w", err)
	}
	codec, err := goavro.NewCodec(schema.Schema())
	if err != nil {
		return nil, fmt.Errorf("creating codec from schema: %w", err)
	}
	c.idToCodec[schemaID] = codec
	return codec, nil
}

// loadSchemaFromFile loads a schema from the schemas folder
func (c *ConfluentAvroCodec) loadSchemaFromFile(schemaName string) (string, error) {
	// Map schema names to file names
	schemaFileMap := map[string]string{
		"commands":         "command.avsc",
		"tasks":            "task.avsc",
		"devices":          "device.avsc",
		"scheduled-tasks":  "scheduled_task.avsc",
		"tenants":          "tenant.avsc",
		"evaluation-rules": "evaluation_rule.avsc",
	}

	fileName, exists := schemaFileMap[schemaName]
	if !exists {
		return "", fmt.Errorf("no schema file mapping for %s", schemaName)
	}

	// Read schema from file
	schemaPath := "../../../schemas/" + fileName
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return "", fmt.Errorf("reading schema file %s: %w", schemaPath, err)
	}

	return string(schemaBytes), nil
}

// Encode encodes a value into Confluent Avro format with schema registry
func (c *ConfluentAvroCodec) Encode(value any) ([]byte, error) {
	avroValue, err := c.convertToAvroStruct(value)
	if err != nil {
		return nil, fmt.Errorf("converting to Avro struct: %w", err)
	}

	schemaName, err := c.getSchemaForMessage(value)
	if err != nil {
		return nil, fmt.Errorf("getting schema for message: %w", err)
	}

	schemaID, err := c.getOrRegisterSchemaID(schemaName)
	if err != nil {
		return nil, fmt.Errorf("getting schema ID: %w", err)
	}

	codec, err := c.getCodecByID(schemaID)
	if err != nil {
		return nil, fmt.Errorf("getting codec by schema ID: %w", err)
	}

	avroData, err := codec.BinaryFromNative(nil, avroValue)
	if err != nil {
		return nil, fmt.Errorf("encoding to Avro: %w", err)
	}

	result := make([]byte, 5+len(avroData))
	result[0] = 0 // Magic byte
	binary.BigEndian.PutUint32(result[1:5], uint32(schemaID))
	copy(result[5:], avroData)

	return result, nil
}

// Decode decodes a value from Confluent Avro format with schema registry
func (c *ConfluentAvroCodec) Decode(data []byte) (any, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("invalid Avro data: too short")
	}
	if data[0] != 0 {
		return nil, fmt.Errorf("invalid magic byte: expected 0, got %d", data[0])
	}
	schemaID := int(binary.BigEndian.Uint32(data[1:5]))
	avroData := data[5:]

	codec, err := c.getCodecByID(schemaID)
	if err != nil {
		return nil, fmt.Errorf("getting codec by schema ID: %w", err)
	}

	native, _, err := codec.NativeFromBinary(avroData)
	if err != nil {
		return nil, fmt.Errorf("decoding Avro data: %w", err)
	}

	result, err := c.convertFromAvroStruct(native)
	if err != nil {
		return nil, fmt.Errorf("converting from Avro struct: %w", err)
	}

	return result, nil
}

// convertToAvroStruct converts original struct to Avro-compatible struct
func (c *ConfluentAvroCodec) convertToAvroStruct(value any) (any, error) {
	// Handle Avro-compatible structs directly by converting to map
	switch v := value.(type) {
	case *AvroCommand:
		return map[string]interface{}{
			"id":             v.ID,
			"version":        v.Version,
			"device_name":    v.DeviceName,
			"device_id":      v.DeviceID,
			"task_id":        v.TaskID,
			"payload":        map[string]interface{}{"index": v.Payload.Index, "value": v.Payload.Value},
			"dispatch_after": v.DispatchAfter,
			"port":           v.Port,
			"priority":       v.Priority,
			"created_at":     v.CreatedAt,
			"ready":          v.Ready,
			"sent":           v.Sent,
			"sent_at":        v.SentAt,
		}, nil
	case *AvroTask:
		result := map[string]interface{}{
			"id":         v.ID,
			"device_id":  v.DeviceID,
			"version":    v.Version,
			"created_at": v.CreatedAt,
			"updated_at": v.UpdatedAt,
		}

		// Handle nullable scheduled_task_id field for Avro union type
		if v.ScheduledTaskID != nil {
			result["scheduled_task_id"] = map[string]interface{}{"string": *v.ScheduledTaskID}
		} else {
			result["scheduled_task_id"] = nil
		}

		return result, nil
	case *AvroDevice:
		result := map[string]interface{}{
			"id":           v.ID,
			"version":      v.Version,
			"name":         v.Name,
			"display_name": v.DisplayName,
			"app_eui":      v.AppEUI,
			"dev_eui":      v.DevEUI,
			"app_key":      v.AppKey,
			"created_at":   v.CreatedAt,
			"updated_at":   v.UpdatedAt,
		}

		// Handle nullable tenant_id field for Avro union type
		if v.TenantID != nil {
			result["tenant_id"] = map[string]interface{}{"string": *v.TenantID}
		} else {
			result["tenant_id"] = nil
		}

		// Handle nullable last_message_received_at field for Avro union type
		if v.LastMessageReceivedAt != nil {
			result["last_message_received_at"] = map[string]interface{}{"string": *v.LastMessageReceivedAt}
		} else {
			result["last_message_received_at"] = nil
		}

		return result, nil
	case *AvroScheduledTask:
		result := map[string]interface{}{
			"id":                v.ID,
			"version":           v.Version,
			"tenant_id":         v.TenantID,
			"device_id":         v.DeviceID,
			"command_templates": v.CommandTemplates,
			"schedule":          v.Schedule,
			"is_active":         v.IsActive,
			"created_at":        v.CreatedAt,
			"updated_at":        v.UpdatedAt,
		}

		// Handle nullable last_executed_at field for Avro union type
		if v.LastExecutedAt != nil {
			result["last_executed_at"] = map[string]interface{}{"string": *v.LastExecutedAt}
		} else {
			result["last_executed_at"] = nil
		}

		// Handle nullable deleted_at field for Avro union type
		if v.DeletedAt != nil {
			result["deleted_at"] = map[string]interface{}{"string": *v.DeletedAt}
		} else {
			result["deleted_at"] = nil
		}

		return result, nil
	case *AvroTenant:
		result := map[string]interface{}{
			"id":          v.ID,
			"version":     v.Version,
			"name":        v.Name,
			"email":       v.Email,
			"description": v.Description,
			"is_active":   v.IsActive,
			"created_at":  v.CreatedAt,
			"updated_at":  v.UpdatedAt,
		}

		// Handle nullable deleted_at field for Avro union type
		if v.DeletedAt != nil {
			result["deleted_at"] = map[string]interface{}{"string": *v.DeletedAt}
		} else {
			result["deleted_at"] = nil
		}

		return result, nil
	case *AvroEvaluationRule:
		return map[string]interface{}{
			"id":          v.ID,
			"device_id":   v.DeviceID,
			"version":     v.Version,
			"description": v.Description,
			"kind":        v.Kind,
			"enabled":     v.Enabled,
			"parameters":  v.Parameters,
			"created_at":  v.CreatedAt,
			"updated_at":  v.UpdatedAt,
		}, nil
	case AvroCommand:
		return map[string]interface{}{
			"id":             v.ID,
			"version":        v.Version,
			"device_name":    v.DeviceName,
			"device_id":      v.DeviceID,
			"task_id":        v.TaskID,
			"payload":        map[string]interface{}{"index": v.Payload.Index, "value": v.Payload.Value},
			"dispatch_after": v.DispatchAfter,
			"port":           v.Port,
			"priority":       v.Priority,
			"created_at":     v.CreatedAt,
			"ready":          v.Ready,
			"sent":           v.Sent,
			"sent_at":        v.SentAt,
		}, nil
	case AvroTask:
		result := map[string]interface{}{
			"id":         v.ID,
			"device_id":  v.DeviceID,
			"version":    v.Version,
			"created_at": v.CreatedAt,
			"updated_at": v.UpdatedAt,
		}

		// Handle nullable scheduled_task_id field for Avro union type
		if v.ScheduledTaskID != nil {
			result["scheduled_task_id"] = map[string]interface{}{"string": *v.ScheduledTaskID}
		} else {
			result["scheduled_task_id"] = nil
		}

		return result, nil
	case AvroDevice:
		result := map[string]interface{}{
			"id":           v.ID,
			"version":      v.Version,
			"name":         v.Name,
			"display_name": v.DisplayName,
			"app_eui":      v.AppEUI,
			"dev_eui":      v.DevEUI,
			"app_key":      v.AppKey,
			"created_at":   v.CreatedAt,
			"updated_at":   v.UpdatedAt,
		}

		// Handle nullable tenant_id field for Avro union type
		if v.TenantID != nil {
			result["tenant_id"] = map[string]interface{}{"string": *v.TenantID}
		} else {
			result["tenant_id"] = nil
		}

		// Handle nullable last_message_received_at field for Avro union type
		if v.LastMessageReceivedAt != nil {
			result["last_message_received_at"] = map[string]interface{}{"string": *v.LastMessageReceivedAt}
		} else {
			result["last_message_received_at"] = nil
		}

		return result, nil
	case AvroScheduledTask:
		result := map[string]interface{}{
			"id":                v.ID,
			"version":           v.Version,
			"tenant_id":         v.TenantID,
			"device_id":         v.DeviceID,
			"command_templates": v.CommandTemplates,
			"schedule":          v.Schedule,
			"is_active":         v.IsActive,
			"created_at":        v.CreatedAt,
			"updated_at":        v.UpdatedAt,
		}

		// Handle nullable last_executed_at field for Avro union type
		if v.LastExecutedAt != nil {
			result["last_executed_at"] = map[string]interface{}{"string": *v.LastExecutedAt}
		} else {
			result["last_executed_at"] = nil
		}

		// Handle nullable deleted_at field for Avro union type
		if v.DeletedAt != nil {
			result["deleted_at"] = map[string]interface{}{"string": *v.DeletedAt}
		} else {
			result["deleted_at"] = nil
		}

		return result, nil
	case AvroTenant:
		result := map[string]interface{}{
			"id":          v.ID,
			"version":     v.Version,
			"name":        v.Name,
			"email":       v.Email,
			"description": v.Description,
			"is_active":   v.IsActive,
			"created_at":  v.CreatedAt,
			"updated_at":  v.UpdatedAt,
		}

		// Handle nullable deleted_at field for Avro union type
		if v.DeletedAt != nil {
			result["deleted_at"] = map[string]interface{}{"string": *v.DeletedAt}
		} else {
			result["deleted_at"] = nil
		}

		return result, nil
	case AvroEvaluationRule:
		return map[string]interface{}{
			"id":          v.ID,
			"device_id":   v.DeviceID,
			"version":     v.Version,
			"description": v.Description,
			"kind":        v.Kind,
			"enabled":     v.Enabled,
			"parameters":  v.Parameters,
			"created_at":  v.CreatedAt,
			"updated_at":  v.UpdatedAt,
		}, nil
	}

	// Convert original structs to Avro structs
	switch v := value.(type) {
	case *shared_kernel.Command:
		return map[string]interface{}{
			"id":             v.ID,
			"version":        v.Version,
			"device_name":    v.DeviceName,
			"device_id":      v.DeviceID,
			"task_id":        v.TaskID,
			"payload":        map[string]interface{}{"index": int(v.Payload.Index), "value": int(v.Payload.Value)},
			"dispatch_after": v.DispatchAfter.Time.Format(time.RFC3339),
			"port":           int(v.Port),
			"priority":       v.Priority,
			"created_at":     v.CreatedAt.Time.Format(time.RFC3339),
			"ready":          v.Ready,
			"sent":           v.Sent,
			"sent_at":        v.SentAt.Time.Format(time.RFC3339),
		}, nil
	}

	// Convert internal persistence types to Avro structs using reflection
	valueType := reflect.TypeOf(value)
	if valueType.Kind() == reflect.Ptr {
		valueType = valueType.Elem()
	}

	// Handle internal persistence types by their struct name
	switch valueType.Name() {
	case "Tenant":
		return c.convertInternalTenant(value)
	case "Device":
		return c.convertInternalDevice(value)
	case "Task":
		return c.convertInternalTask(value)
	case "ScheduledTask":
		return c.convertInternalScheduledTask(value)
	case "EvaluationRule":
		return c.convertInternalEvaluationRule(value)
	case "Command":
		return c.convertInternalCommand(value)
	default:
		return nil, fmt.Errorf("unsupported type for Avro conversion: %T", value)
	}
}

// convertFromAvroStruct converts Avro struct back to original struct
func (c *ConfluentAvroCodec) convertFromAvroStruct(value any) (any, error) {
	// Handle map[string]interface{} from Avro decoding
	if mapValue, ok := value.(map[string]interface{}); ok {
		// Try to determine the type from the map structure
		if _, hasID := mapValue["id"]; hasID {
			if _, hasDeviceName := mapValue["device_name"]; hasDeviceName {
				// Command
				dispatchAfter, _ := time.Parse(time.RFC3339, getString(mapValue, "dispatch_after"))
				createdAt, _ := time.Parse(time.RFC3339, getString(mapValue, "created_at"))
				sentAt, _ := time.Parse(time.RFC3339, getString(mapValue, "sent_at"))

				return &shared_kernel.Command{
					ID:            getString(mapValue, "id"),
					Version:       getInt(mapValue, "version"),
					DeviceName:    getString(mapValue, "device_name"),
					DeviceID:      getString(mapValue, "device_id"),
					TaskID:        getString(mapValue, "task_id"),
					Payload:       shared_kernel.CommandPayload{Index: uint8(getInt(mapValue, "payload_index")), Value: uint8(getInt(mapValue, "payload_value"))},
					DispatchAfter: utils.Time{Time: dispatchAfter},
					Port:          uint8(getInt(mapValue, "port")),
					Priority:      getString(mapValue, "priority"),
					CreatedAt:     utils.Time{Time: createdAt},
					Ready:         getBool(mapValue, "ready"),
					Sent:          getBool(mapValue, "sent"),
					SentAt:        utils.Time{Time: sentAt},
				}, nil
			} else if _, hasDeviceID := mapValue["device_id"]; hasDeviceID {
				if _, hasScheduledTaskID := mapValue["scheduled_task_id"]; hasScheduledTaskID {
					// Task
					return &AvroTask{
						ID:              getString(mapValue, "id"),
						DeviceID:        getString(mapValue, "device_id"),
						ScheduledTaskID: getStringPtr(mapValue, "scheduled_task_id"),
						Version:         getInt64(mapValue, "version"),
						CreatedAt:       parseTimeRFC3339(getString(mapValue, "created_at")),
						UpdatedAt:       parseTimeRFC3339(getString(mapValue, "updated_at")),
					}, nil
				} else if _, hasName := mapValue["name"]; hasName {
					// Device
					return &AvroDevice{
						ID:                    getString(mapValue, "id"),
						Version:               getInt(mapValue, "version"),
						Name:                  getString(mapValue, "name"),
						DisplayName:           getString(mapValue, "display_name"),
						AppEUI:                getString(mapValue, "app_eui"),
						DevEUI:                getString(mapValue, "dev_eui"),
						AppKey:                getString(mapValue, "app_key"),
						TenantID:              getStringPtr(mapValue, "tenant_id"),
						LastMessageReceivedAt: parseTimePtrRFC3339(getStringPtr(mapValue, "last_message_received_at")),
						CreatedAt:             parseTimeRFC3339(getString(mapValue, "created_at")),
						UpdatedAt:             parseTimeRFC3339(getString(mapValue, "updated_at")),
					}, nil
				} else if _, hasKind := mapValue["kind"]; hasKind {
					// EvaluationRule
					return &AvroEvaluationRule{
						ID:          getString(mapValue, "id"),
						DeviceID:    getString(mapValue, "device_id"),
						Version:     getInt(mapValue, "version"),
						Description: getString(mapValue, "description"),
						Kind:        getString(mapValue, "kind"),
						Enabled:     getBool(mapValue, "enabled"),
						Parameters:  getString(mapValue, "parameters"),
						CreatedAt:   parseTimeRFC3339(getString(mapValue, "created_at")),
						UpdatedAt:   parseTimeRFC3339(getString(mapValue, "updated_at")),
					}, nil
				}
			} else if _, hasName := mapValue["name"]; hasName {
				if _, hasEmail := mapValue["email"]; hasEmail {
					// Tenant
					return &AvroTenant{
						ID:          getString(mapValue, "id"),
						Version:     getInt(mapValue, "version"),
						Name:        getString(mapValue, "name"),
						Email:       getString(mapValue, "email"),
						Description: getString(mapValue, "description"),
						IsActive:    getBool(mapValue, "is_active"),
						CreatedAt:   parseTimeRFC3339(getString(mapValue, "created_at")),
						UpdatedAt:   parseTimeRFC3339(getString(mapValue, "updated_at")),
						DeletedAt:   parseTimePtrRFC3339(getStringPtr(mapValue, "deleted_at")),
					}, nil
				}
			} else if _, hasTenantID := mapValue["tenant_id"]; hasTenantID {
				if _, hasCommandTemplates := mapValue["command_templates"]; hasCommandTemplates {
					// ScheduledTask
					return &AvroScheduledTask{
						ID:               getString(mapValue, "id"),
						Version:          getInt64(mapValue, "version"),
						TenantID:         getString(mapValue, "tenant_id"),
						DeviceID:         getString(mapValue, "device_id"),
						CommandTemplates: getString(mapValue, "command_templates"),
						Schedule:         getString(mapValue, "schedule"),
						IsActive:         getBool(mapValue, "is_active"),
						CreatedAt:        parseTimeRFC3339(getString(mapValue, "created_at")),
						UpdatedAt:        parseTimeRFC3339(getString(mapValue, "updated_at")),
						LastExecutedAt:   parseTimePtrRFC3339(getStringPtr(mapValue, "last_executed_at")),
						DeletedAt:        parseTimePtrRFC3339(getStringPtr(mapValue, "deleted_at")),
					}, nil
				}
			}
		}
		return nil, fmt.Errorf("unable to determine type from map structure")
	}

	// Handle Avro structs directly
	switch v := value.(type) {
	case *AvroCommand:
		return &shared_kernel.Command{
			ID:            v.ID,
			Version:       v.Version,
			DeviceName:    v.DeviceName,
			DeviceID:      v.DeviceID,
			TaskID:        v.TaskID,
			Payload:       shared_kernel.CommandPayload{Index: uint8(v.Payload.Index), Value: uint8(v.Payload.Value)},
			DispatchAfter: utils.Time{Time: v.DispatchAfter},
			Port:          uint8(v.Port),
			Priority:      v.Priority,
			CreatedAt:     utils.Time{Time: v.CreatedAt},
			Ready:         v.Ready,
			Sent:          v.Sent,
			SentAt:        utils.Time{Time: v.SentAt},
		}, nil
	case *AvroTask:
		return &AvroTask{
			ID:              v.ID,
			DeviceID:        v.DeviceID,
			ScheduledTaskID: v.ScheduledTaskID,
			Version:         v.Version,
			CreatedAt:       v.CreatedAt,
			UpdatedAt:       v.UpdatedAt,
		}, nil
	case *AvroDevice:
		return &AvroDevice{
			ID:                    v.ID,
			Version:               v.Version,
			Name:                  v.Name,
			DisplayName:           v.DisplayName,
			AppEUI:                v.AppEUI,
			DevEUI:                v.DevEUI,
			AppKey:                v.AppKey,
			TenantID:              v.TenantID,
			LastMessageReceivedAt: v.LastMessageReceivedAt,
			CreatedAt:             v.CreatedAt,
			UpdatedAt:             v.UpdatedAt,
		}, nil
	case *AvroScheduledTask:
		return &AvroScheduledTask{
			ID:               v.ID,
			Version:          v.Version,
			TenantID:         v.TenantID,
			DeviceID:         v.DeviceID,
			CommandTemplates: v.CommandTemplates,
			Schedule:         v.Schedule,
			IsActive:         v.IsActive,
			CreatedAt:        v.CreatedAt,
			UpdatedAt:        v.UpdatedAt,
			LastExecutedAt:   v.LastExecutedAt,
			DeletedAt:        v.DeletedAt,
		}, nil
	case *AvroTenant:
		return &AvroTenant{
			ID:          v.ID,
			Version:     v.Version,
			Name:        v.Name,
			Email:       v.Email,
			Description: v.Description,
			IsActive:    v.IsActive,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
			DeletedAt:   v.DeletedAt,
		}, nil
	case *AvroEvaluationRule:
		return &AvroEvaluationRule{
			ID:          v.ID,
			DeviceID:    v.DeviceID,
			Version:     v.Version,
			Description: v.Description,
			Kind:        v.Kind,
			Enabled:     v.Enabled,
			Parameters:  v.Parameters,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported type for conversion from Avro: %T", value)
	}
}

// Helper functions for map conversion
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// parseTimeRFC3339 returns a time.Time parsed from s, or zero time if s is empty or invalid.
func parseTimeRFC3339(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// parseTimePtrRFC3339 returns a *time.Time parsed from s, or nil if s is nil/empty/invalid.
func parseTimePtrRFC3339(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil
	}
	return &t
}

func getStringPtr(m map[string]interface{}, key string) *string {
	if v, ok := m[key]; ok {
		// Handle Avro union type format: map[string]interface{}{"string": "value"}
		if unionMap, ok := v.(map[string]interface{}); ok {
			if s, ok := unionMap["string"].(string); ok {
				return &s
			}
		}
		// Handle direct string value
		if s, ok := v.(string); ok {
			return &s
		}
	}
	return nil
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		if i, ok := v.(int); ok {
			return i
		}
		if f, ok := v.(float64); ok {
			return int(f)
		}
	}
	return 0
}

func getInt64(m map[string]interface{}, key string) int64 {
	if v, ok := m[key]; ok {
		if i, ok := v.(int64); ok {
			return i
		}
		if f, ok := v.(float64); ok {
			return int64(f)
		}
	}
	return 0
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// Helper methods for converting internal persistence types using reflection

func (c *ConfluentAvroCodec) convertInternalTenant(value any) (any, error) {
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	result := map[string]interface{}{
		"id":          getStringField(val, "ID"),
		"version":     getIntField(val, "Version"),
		"name":        getStringField(val, "Name"),
		"email":       getStringField(val, "Email"),
		"description": getStringField(val, "Description"),
		"is_active":   getBoolField(val, "IsActive"),
		"created_at":  getTimeField(val, "CreatedAt").Format(time.RFC3339),
		"updated_at":  getTimeField(val, "UpdatedAt").Format(time.RFC3339),
	}

	// Handle optional DeletedAt field
	if deletedAtField := val.FieldByName("DeletedAt"); deletedAtField.IsValid() && !deletedAtField.IsNil() {
		deletedAt := deletedAtField.Elem().Interface().(time.Time)
		deletedAtStr := deletedAt.Format(time.RFC3339)
		result["deleted_at"] = &deletedAtStr
	}

	return result, nil
}

func (c *ConfluentAvroCodec) convertInternalDevice(value any) (any, error) {
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	result := map[string]interface{}{
		"id":           getStringField(val, "ID"),
		"version":      getIntField(val, "Version"),
		"name":         getStringField(val, "Name"),
		"display_name": getStringField(val, "DisplayName"),
		"app_eui":      getStringField(val, "AppEUI"),
		"dev_eui":      getStringField(val, "DevEUI"),
		"app_key":      getStringField(val, "AppKey"),
		"created_at":   getTimeField(val, "CreatedAt").Format(time.RFC3339),
		"updated_at":   getTimeField(val, "UpdatedAt").Format(time.RFC3339),
	}

	// Handle optional TenantID field
	if tenantIDField := val.FieldByName("TenantID"); tenantIDField.IsValid() && !tenantIDField.IsNil() {
		tenantID := tenantIDField.Elem().Interface().(string)
		result["tenant_id"] = &tenantID
	}

	// Handle optional LastMessageReceivedAt field
	if lastMessageField := val.FieldByName("LastMessageReceivedAt"); lastMessageField.IsValid() {
		lastMessageTime := lastMessageField.FieldByName("Time").Interface().(time.Time)
		if lastMessageTime != (time.Time{}) {
			lastMessageStr := lastMessageTime.Format(time.RFC3339)
			result["last_message_received_at"] = &lastMessageStr
		}
	}

	return result, nil
}

func (c *ConfluentAvroCodec) convertInternalTask(value any) (any, error) {
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	result := map[string]interface{}{
		"id":         getStringField(val, "ID"),
		"device_id":  getStringField(val, "DeviceID"),
		"version":    int64(getUintField(val, "Version")),
		"created_at": getTimeField(val, "CreatedAt").Format(time.RFC3339),
		"updated_at": getTimeField(val, "UpdatedAt").Format(time.RFC3339),
	}

	// Handle optional ScheduledTaskID field
	if scheduledTaskID := getStringField(val, "ScheduledTaskID"); scheduledTaskID != "" {
		result["scheduled_task_id"] = &scheduledTaskID
	}

	return result, nil
}

func (c *ConfluentAvroCodec) convertInternalScheduledTask(value any) (any, error) {
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	result := map[string]interface{}{
		"id":                getStringField(val, "ID"),
		"version":           int64(getUintField(val, "Version")),
		"tenant_id":         getStringField(val, "TenantID"),
		"device_id":         getStringField(val, "DeviceID"),
		"command_templates": getStringField(val, "CommandTemplates"),
		"schedule":          getStringField(val, "Schedule"),
		"is_active":         getBoolField(val, "IsActive"),
		"created_at":        getTimeField(val, "CreatedAt").Format(time.RFC3339),
		"updated_at":        getTimeField(val, "UpdatedAt").Format(time.RFC3339),
	}

	// Handle optional LastExecutedAt field
	if lastExecutedField := val.FieldByName("LastExecutedAt"); lastExecutedField.IsValid() && !lastExecutedField.IsNil() {
		lastExecutedTime := lastExecutedField.Elem().FieldByName("Time").Interface().(time.Time)
		lastExecutedStr := lastExecutedTime.Format(time.RFC3339)
		result["last_executed_at"] = &lastExecutedStr
	}

	// Handle optional DeletedAt field
	if deletedAtField := val.FieldByName("DeletedAt"); deletedAtField.IsValid() && !deletedAtField.IsNil() {
		deletedAtTime := deletedAtField.Elem().FieldByName("Time").Interface().(time.Time)
		deletedAtStr := deletedAtTime.Format(time.RFC3339)
		result["deleted_at"] = &deletedAtStr
	}

	return result, nil
}

func (c *ConfluentAvroCodec) convertInternalEvaluationRule(value any) (any, error) {
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Get parameters field and marshal it to JSON
	parametersField := val.FieldByName("Parameters")
	var parametersStr string
	if parametersField.IsValid() {
		parameters := parametersField.Interface()
		if parametersBytes, err := json.Marshal(parameters); err == nil {
			parametersStr = string(parametersBytes)
		}
	}

	result := map[string]interface{}{
		"id":          getStringField(val, "ID"),
		"device_id":   getStringField(val, "DeviceID"),
		"version":     getIntField(val, "Version"),
		"description": getStringField(val, "Description"),
		"kind":        getStringField(val, "Kind"),
		"enabled":     getBoolField(val, "Enabled"),
		"parameters":  parametersStr,
		"created_at":  getTimeField(val, "CreatedAt").Format(time.RFC3339),
		"updated_at":  getTimeField(val, "UpdatedAt").Format(time.RFC3339),
	}

	return result, nil
}

func (c *ConfluentAvroCodec) convertInternalCommand(value any) (any, error) {
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Get payload field
	payloadField := val.FieldByName("Payload")
	payloadIndex := 0
	payloadValue := 0
	if payloadField.IsValid() {
		payloadIndex = int(payloadField.FieldByName("Index").Interface().(uint8))
		payloadValue = int(payloadField.FieldByName("Value").Interface().(uint8))
	}

	result := map[string]interface{}{
		"id":             getStringField(val, "ID"),
		"version":        getIntField(val, "Version"),
		"device_name":    getStringField(val, "DeviceName"),
		"device_id":      getStringField(val, "DeviceID"),
		"task_id":        getStringField(val, "TaskID"),
		"payload":        map[string]interface{}{"index": payloadIndex, "value": payloadValue},
		"dispatch_after": getTimeField(val, "DispatchAfter").Format(time.RFC3339),
		"port":           int(getUint8Field(val, "Port")),
		"priority":       getStringField(val, "Priority"),
		"created_at":     getTimeField(val, "CreatedAt").Format(time.RFC3339),
		"ready":          getBoolField(val, "Ready"),
		"sent":           getBoolField(val, "Sent"),
		"sent_at":        getTimeField(val, "SentAt").Format(time.RFC3339),
	}

	return result, nil
}

// Helper functions for reflection-based field access

func getStringField(val reflect.Value, fieldName string) string {
	if field := val.FieldByName(fieldName); field.IsValid() {
		if str, ok := field.Interface().(string); ok {
			return str
		}
	}
	return ""
}

func getIntField(val reflect.Value, fieldName string) int {
	if field := val.FieldByName(fieldName); field.IsValid() {
		if i, ok := field.Interface().(int); ok {
			return i
		}
	}
	return 0
}

func getUintField(val reflect.Value, fieldName string) uint {
	if field := val.FieldByName(fieldName); field.IsValid() {
		if u, ok := field.Interface().(uint); ok {
			return u
		}
	}
	return 0
}

func getUint8Field(val reflect.Value, fieldName string) uint8 {
	if field := val.FieldByName(fieldName); field.IsValid() {
		if u, ok := field.Interface().(uint8); ok {
			return u
		}
	}
	return 0
}

func getBoolField(val reflect.Value, fieldName string) bool {
	if field := val.FieldByName(fieldName); field.IsValid() {
		if b, ok := field.Interface().(bool); ok {
			return b
		}
	}
	return false
}

func getTimeField(val reflect.Value, fieldName string) time.Time {
	if field := val.FieldByName(fieldName); field.IsValid() {
		if t, ok := field.Interface().(time.Time); ok {
			return t
		}
		// Handle utils.Time wrapper
		if utilsTimeField := field.FieldByName("Time"); utilsTimeField.IsValid() {
			if t, ok := utilsTimeField.Interface().(time.Time); ok {
				return t
			}
		}
	}
	return time.Time{}
}

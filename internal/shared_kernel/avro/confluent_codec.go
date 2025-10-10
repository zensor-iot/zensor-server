package avro

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"

	"zensor-server/internal/infra/cache"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/device"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/linkedin/goavro/v2"
	"github.com/riferrei/srclient"
)

const (
	_defaultSchemaCacheTTL = 5 * time.Minute
	_defaultCodecCacheTTL  = 5 * time.Minute
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
	schemaRegistry SchemaRegistry
	subjectSuffix  string
	schemaCache    cache.Cache
	codecCache     cache.Cache
}

// NewConfluentAvroCodec creates a new Confluent Avro codec with schema registry
func NewConfluentAvroCodec(_ any, schemaRegistry SchemaRegistry) *ConfluentAvroCodec {
	// Create schema cache with 1-minute TTL
	schemaCache, _ := cache.New(&cache.CacheConfig{
		MaxCost:     1 << 20, // 1MB
		NumCounters: 1e6,     // 1M
		BufferItems: 64,
	})

	// Create codec cache with 1-minute TTL
	codecCache, _ := cache.New(&cache.CacheConfig{
		MaxCost:     1 << 20, // 1MB
		NumCounters: 1e6,     // 1M
		BufferItems: 64,
	})

	return &ConfluentAvroCodec{
		schemaRegistry: schemaRegistry,
		subjectSuffix:  "-value",
		schemaCache:    schemaCache,
		codecCache:     codecCache,
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
		return "device_commands", nil
	case "Task", "AvroTask":
		return "tasks", nil
	case "Device", "AvroDevice":
		return "devices", nil
	case "ScheduledTask", "AvroScheduledTask":
		return "scheduled_tasks", nil
	case "Tenant", "AvroTenant":
		return "tenants", nil
	case "EvaluationRule", "AvroEvaluationRule":
		return "evaluation_rules", nil
	case "TenantConfiguration", "AvroTenantConfiguration":
		return "tenant_configurations", nil
	default:
		return "", fmt.Errorf("no Avro schema found for message type: %s", schemaName)
	}
}

// getOrRegisterSchemaID gets or registers the schema in the registry and returns its ID
func (c *ConfluentAvroCodec) getOrRegisterSchemaID(schemaName string) (int, error) {
	subject := schemaName + c.subjectSuffix

	ctx := context.Background()
	if cached, found := c.schemaCache.Get(ctx, subject); found {
		if id, ok := cached.(int); ok {
			return id, nil
		}
	}

	registered, err := c.schemaRegistry.GetLatestSchema(subject)
	if err == nil && registered != nil {
		c.schemaCache.Set(ctx, subject, registered.ID(), _defaultSchemaCacheTTL)
		return registered.ID(), nil
	}

	schema, err := c.loadSchemaFromFile(schemaName)
	if err != nil {
		return 0, fmt.Errorf("loading schema from file: %w", err)
	}

	newSchema, err := c.schemaRegistry.CreateSchema(subject, schema, srclient.Avro)
	if err != nil {
		return 0, fmt.Errorf("registering schema: %w", err)
	}

	c.schemaCache.Set(ctx, subject, newSchema.ID(), _defaultSchemaCacheTTL)
	return newSchema.ID(), nil
}

// getCodecByID fetches the codec for a schema ID from the registry if not cached
func (c *ConfluentAvroCodec) getCodecByID(schemaID int) (*goavro.Codec, error) {
	ctx := context.Background()
	schemaIDKey := fmt.Sprintf("schema_%d", schemaID)

	if cached, found := c.codecCache.Get(ctx, schemaIDKey); found {
		if codec, ok := cached.(*goavro.Codec); ok {
			return codec, nil
		}
	}

	schema, err := c.schemaRegistry.GetSchema(schemaID)
	if err != nil {
		return nil, fmt.Errorf("fetching schema from registry: %w", err)
	}
	codec, err := goavro.NewCodec(schema.Schema())
	if err != nil {
		return nil, fmt.Errorf("creating codec from schema: %w", err)
	}
	c.codecCache.Set(ctx, schemaIDKey, codec, _defaultCodecCacheTTL)
	return codec, nil
}

// loadSchemaFromFile loads a schema from the schemas folder
func (c *ConfluentAvroCodec) loadSchemaFromFile(schemaName string) (string, error) {
	// Map schema names to file names
	schemaFileMap := map[string]string{
		"tasks":                 "task.avsc",
		"devices":               "device.avsc",
		"scheduled_tasks":       "scheduled_task.avsc",
		"tenants":               "tenant.avsc",
		"evaluation_rules":      "evaluation_rule.avsc",
		"device_commands":       "device_command.avsc",
		"tenant_configurations": "tenant_configuration.avsc",
	}

	fileName, exists := schemaFileMap[schemaName]
	if !exists {
		return "", fmt.Errorf("no schema file mapping for %s", schemaName)
	}

	// Read schema from file
	schemaPath := "./schemas/" + fileName
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
		result := map[string]any{
			"id":             v.ID,
			"version":        v.Version,
			"device_name":    v.DeviceName,
			"device_id":      v.DeviceID,
			"task_id":        v.TaskID,
			"payload_index":  v.PayloadIndex,
			"payload_value":  v.PayloadValue,
			"dispatch_after": v.DispatchAfter,
			"port":           v.Port,
			"priority":       v.Priority,
			"created_at":     v.CreatedAt,
			"ready":          v.Ready,
			"sent":           v.Sent,
			"sent_at":        v.SentAt,

			// Response tracking fields
			"status": v.Status,
		}

		// Handle nullable error_message field for Avro union type
		if v.ErrorMessage != nil {
			result["error_message"] = map[string]any{"string": *v.ErrorMessage}
		} else {
			result["error_message"] = nil
		}

		// Handle optional timestamp fields
		if v.QueuedAt != nil {
			result["queued_at"] = map[string]any{
				"long.timestamp-millis": v.QueuedAt.UnixMilli(),
			}
		} else {
			result["queued_at"] = nil
		}
		if v.AckedAt != nil {
			result["acked_at"] = map[string]any{
				"long.timestamp-millis": v.AckedAt.UnixMilli(),
			}
		} else {
			result["acked_at"] = nil
		}
		if v.FailedAt != nil {
			result["failed_at"] = map[string]any{
				"long.timestamp-millis": v.FailedAt.UnixMilli(),
			}
		} else {
			result["failed_at"] = nil
		}

		return result, nil
	case *AvroTask:
		result := map[string]any{
			"id":         v.ID,
			"device_id":  v.DeviceID,
			"version":    v.Version,
			"created_at": v.CreatedAt,
			"updated_at": v.UpdatedAt,
		}

		// Handle nullable scheduled_task_id field for Avro union type
		if v.ScheduledTaskID != nil {
			result["scheduled_task_id"] = map[string]any{"string": *v.ScheduledTaskID}
		} else {
			result["scheduled_task_id"] = nil
		}

		return result, nil
	case *AvroDevice:
		result := map[string]any{
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
			result["tenant_id"] = map[string]any{"string": *v.TenantID}
		} else {
			result["tenant_id"] = nil
		}

		if v.LastMessageReceivedAt != nil {
			result["last_message_received_at"] = map[string]any{
				"long.timestamp-millis": v.LastMessageReceivedAt.UnixMilli(),
			}
		} else {
			result["last_message_received_at"] = nil
		}

		return result, nil
	case *AvroScheduledTask:
		result := map[string]any{
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

		if v.SchedulingConfig != nil {
			result["scheduling_config"] = map[string]any{"string": *v.SchedulingConfig}
		} else {
			result["scheduling_config"] = nil
		}

		if v.LastExecutedAt != nil {
			result["last_executed_at"] = map[string]any{
				"long.timestamp-millis": v.LastExecutedAt.UnixMilli(),
			}
		} else {
			result["last_executed_at"] = nil
		}

		if v.DeletedAt != nil {
			result["deleted_at"] = map[string]any{
				"long.timestamp-millis": v.DeletedAt.UnixMilli(),
			}
		} else {
			result["deleted_at"] = nil
		}

		return result, nil
	case *AvroTenant:
		result := map[string]any{
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
			result["deleted_at"] = map[string]any{
				"long.timestamp-millis": v.DeletedAt.UnixMilli(),
			}
		} else {
			result["deleted_at"] = nil
		}

		return result, nil
	case *AvroEvaluationRule:
		return map[string]any{
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
		result := map[string]any{
			"id":             v.ID,
			"version":        v.Version,
			"device_name":    v.DeviceName,
			"device_id":      v.DeviceID,
			"task_id":        v.TaskID,
			"payload_index":  v.PayloadIndex,
			"payload_value":  v.PayloadValue,
			"dispatch_after": v.DispatchAfter,
			"port":           v.Port,
			"priority":       v.Priority,
			"created_at":     v.CreatedAt,
			"ready":          v.Ready,
			"sent":           v.Sent,
			"sent_at":        v.SentAt,

			// Response tracking fields
			"status": v.Status,
		}

		// Handle nullable error_message field for Avro union type
		if v.ErrorMessage != nil {
			result["error_message"] = map[string]any{"string": *v.ErrorMessage}
		} else {
			result["error_message"] = nil
		}

		// Handle optional timestamp fields
		if v.QueuedAt != nil {
			result["queued_at"] = map[string]any{
				"long.timestamp-millis": v.QueuedAt.UnixMilli(),
			}
		} else {
			result["queued_at"] = nil
		}
		if v.AckedAt != nil {
			result["acked_at"] = map[string]any{
				"long.timestamp-millis": v.AckedAt.UnixMilli(),
			}
		} else {
			result["acked_at"] = nil
		}
		if v.FailedAt != nil {
			result["failed_at"] = map[string]any{
				"long.timestamp-millis": v.FailedAt.UnixMilli(),
			}
		} else {
			result["failed_at"] = nil
		}

		return result, nil
	case AvroTask:
		result := map[string]any{
			"id":         v.ID,
			"device_id":  v.DeviceID,
			"version":    v.Version,
			"created_at": v.CreatedAt,
			"updated_at": v.UpdatedAt,
		}

		// Handle nullable scheduled_task_id field for Avro union type
		if v.ScheduledTaskID != nil {
			result["scheduled_task_id"] = map[string]any{"string": *v.ScheduledTaskID}
		} else {
			result["scheduled_task_id"] = nil
		}

		return result, nil
	case AvroDevice:
		result := map[string]any{
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

		if v.TenantID != nil {
			result["tenant_id"] = map[string]any{"string": *v.TenantID}
		} else {
			result["tenant_id"] = nil
		}

		if v.LastMessageReceivedAt != nil {
			result["last_message_received_at"] = map[string]any{
				"long.timestamp-millis": v.LastMessageReceivedAt.UnixMilli(),
			}
		} else {
			result["last_message_received_at"] = nil
		}

		return result, nil
	case AvroScheduledTask:
		result := map[string]any{
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
			result["last_executed_at"] = map[string]any{
				"long.timestamp-millis": v.LastExecutedAt.UnixMilli(),
			}
		} else {
			result["last_executed_at"] = nil
		}

		// Handle nullable deleted_at field for Avro union type
		if v.DeletedAt != nil {
			result["deleted_at"] = map[string]any{
				"long.timestamp-millis": v.DeletedAt.UnixMilli(),
			}
		} else {
			result["deleted_at"] = nil
		}

		return result, nil
	case AvroTenant:
		result := map[string]any{
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
			result["deleted_at"] = map[string]any{
				"long.timestamp-millis": v.DeletedAt.UnixMilli(),
			}
		} else {
			result["deleted_at"] = nil
		}

		return result, nil
	case AvroEvaluationRule:
		return map[string]any{
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
	case *AvroTenantConfiguration:
		return map[string]any{
			"id":         v.ID,
			"tenant_id":  v.TenantID,
			"timezone":   v.Timezone,
			"version":    v.Version,
			"created_at": v.CreatedAt,
			"updated_at": v.UpdatedAt,
		}, nil
	case AvroTenantConfiguration:
		return map[string]any{
			"id":         v.ID,
			"tenant_id":  v.TenantID,
			"timezone":   v.Timezone,
			"version":    v.Version,
			"created_at": v.CreatedAt,
			"updated_at": v.UpdatedAt,
		}, nil
	}

	// Convert original structs to Avro structs
	switch v := value.(type) {
	case *device.Command:
		return map[string]any{
			"id":             v.ID,
			"version":        v.Version,
			"device_name":    v.DeviceName,
			"device_id":      v.DeviceID,
			"task_id":        v.TaskID,
			"payload_index":  int(v.Payload.Index),
			"payload_value":  int(v.Payload.Value),
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
		if device, ok := value.(*domain.Device); ok {
			return c.convertInternalDevice(device)
		}
		return nil, fmt.Errorf("expected *domain.Device, got %T", value)
	case "Task":
		if task, ok := value.(*domain.Task); ok {
			return c.convertInternalTask(task)
		}
		return nil, fmt.Errorf("expected *domain.Task, got %T", value)
	case "ScheduledTask":
		if st, ok := value.(*domain.ScheduledTask); ok {
			return c.convertInternalScheduledTask(st)
		}
		return nil, fmt.Errorf("expected *domain.ScheduledTask, got %T", value)
	case "EvaluationRule":
		if er, ok := value.(*domain.EvaluationRule); ok {
			return c.convertInternalEvaluationRule(er)
		}
		return nil, fmt.Errorf("expected *domain.EvaluationRule, got %T", value)
	case "Command":
		if cmd, ok := value.(*domain.Command); ok {
			return c.convertInternalCommand(cmd)
		}
		return nil, fmt.Errorf("expected *domain.Command, got %T", value)
	case "TenantConfiguration":
		if config, ok := value.(*domain.TenantConfiguration); ok {
			return c.convertInternalTenantConfiguration(config)
		}
		return nil, fmt.Errorf("expected *domain.TenantConfiguration, got %T", value)
	default:
		return nil, fmt.Errorf("unsupported type for Avro conversion: %T", value)
	}
}

// convertFromAvroStruct converts Avro struct back to original struct
func (c *ConfluentAvroCodec) convertFromAvroStruct(value any) (any, error) {
	// Handle map[string]any from Avro decoding
	if mapValue, ok := value.(map[string]any); ok {
		// Try to determine the type from the map structure
		if _, hasID := mapValue["id"]; hasID {
			if _, hasDeviceName := mapValue["device_name"]; hasDeviceName {
				// Command
				dispatchAfter := getTime(mapValue, "dispatch_after")
				createdAt := getTime(mapValue, "created_at")
				sentAt := getTime(mapValue, "sent_at")

				return &AvroCommand{
					ID:            getString(mapValue, "id"),
					Version:       getInt(mapValue, "version"),
					DeviceName:    getString(mapValue, "device_name"),
					DeviceID:      getString(mapValue, "device_id"),
					TaskID:        getString(mapValue, "task_id"),
					PayloadIndex:  getInt(mapValue, "payload_index"),
					PayloadValue:  getInt(mapValue, "payload_value"),
					DispatchAfter: dispatchAfter,
					Port:          getInt(mapValue, "port"),
					Priority:      getString(mapValue, "priority"),
					CreatedAt:     createdAt,
					Ready:         getBool(mapValue, "ready"),
					Sent:          getBool(mapValue, "sent"),
					SentAt:        sentAt,

					// Response tracking fields
					Status:       getString(mapValue, "status"),
					ErrorMessage: getStringPtr(mapValue, "error_message"),
					QueuedAt:     parseTimePtrRFC3339(getStringPtr(mapValue, "queued_at")),
					AckedAt:      parseTimePtrRFC3339(getStringPtr(mapValue, "acked_at")),
					FailedAt:     parseTimePtrRFC3339(getStringPtr(mapValue, "failed_at")),
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
						SchedulingConfig: getStringPtr(mapValue, "scheduling_config"),
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
		return &device.Command{
			ID:            v.ID,
			Version:       v.Version,
			DeviceName:    v.DeviceName,
			DeviceID:      v.DeviceID,
			TaskID:        v.TaskID,
			Payload:       device.CommandPayload{Index: uint8(v.PayloadIndex), Value: uint8(v.PayloadValue)},
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
			SchedulingConfig: v.SchedulingConfig,
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
func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// parseTimeFlexible tries multiple time formats and returns the first successful parse
func parseTimeFlexible(s string) time.Time {
	if s == "" {
		return time.Time{}
	}

	// Try RFC3339 format first
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}

	// Try Go's default format
	if t, err := time.Parse("2006-01-02 15:04:05.999 -0700 MST", s); err == nil {
		return t
	}

	return time.Time{}
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

func getStringPtr(m map[string]any, key string) *string {
	if v, ok := m[key]; ok {
		// Handle Avro union type format: map[string]any{"string": "value"}
		if unionMap, ok := v.(map[string]any); ok {
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

func getInt(m map[string]any, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int32:
			return int(val)
		case int64:
			return int(val)
		case float64:
			return int(val)
		case string:
			if i, err := strconv.Atoi(val); err == nil {
				return i
			}
		}
	}
	return 0
}

func getInt64(m map[string]any, key string) int64 {
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

func getBool(m map[string]any, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func getTime(m map[string]any, key string) time.Time {
	if v, ok := m[key]; ok {
		if t, ok := v.(time.Time); ok {
			return t
		}
	}
	return time.Time{}
}

// Helper methods for converting internal persistence types using reflection

func (c *ConfluentAvroCodec) convertInternalTenant(value any) (any, error) {
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	result := map[string]any{
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

func (c *ConfluentAvroCodec) convertInternalDevice(device *domain.Device) (*AvroDevice, error) {
	avroDevice := &AvroDevice{
		ID:          device.ID.String(),
		Version:     1, // Default version for domain devices
		Name:        device.Name,
		DisplayName: device.DisplayName,
		AppEUI:      device.AppEUI,
		DevEUI:      device.DevEUI,
		AppKey:      device.AppKey,
		CreatedAt:   time.Now(), // Default timestamp for domain devices
		UpdatedAt:   time.Now(), // Default timestamp for domain devices
	}

	// Handle optional TenantID field
	if device.TenantID != nil {
		tenantID := device.TenantID.String()
		avroDevice.TenantID = &tenantID
	}

	// Handle optional LastMessageReceivedAt field
	if !device.LastMessageReceivedAt.IsZero() {
		lastMessageTime := device.LastMessageReceivedAt.Time
		avroDevice.LastMessageReceivedAt = &lastMessageTime
	}

	return avroDevice, nil
}

func (c *ConfluentAvroCodec) convertInternalTask(task *domain.Task) (*AvroTask, error) {
	avroTask := &AvroTask{
		ID:        task.ID.String(),
		DeviceID:  task.Device.ID.String(),
		Version:   int64(task.Version),
		CreatedAt: task.CreatedAt.Time,
		UpdatedAt: task.CreatedAt.Time, // No UpdatedAt in domain.Task, use CreatedAt
	}
	if task.ScheduledTask != nil {
		id := task.ScheduledTask.ID.String()
		avroTask.ScheduledTaskID = &id
	}
	return avroTask, nil
}

func (c *ConfluentAvroCodec) convertInternalScheduledTask(st *domain.ScheduledTask) (*AvroScheduledTask, error) {
	avroST := &AvroScheduledTask{
		ID:               st.ID.String(),
		Version:          int64(st.Version),
		TenantID:         st.Tenant.ID.String(),
		DeviceID:         st.Device.ID.String(),
		CommandTemplates: c.serializeCommandTemplates(st.CommandTemplates),
		Schedule:         st.Schedule,
		IsActive:         st.IsActive,
		CreatedAt:        st.CreatedAt.Time,
		UpdatedAt:        st.UpdatedAt.Time,
	}

	if st.Scheduling.Type != "" {
		schedulingData := schedulingConfigurationData{
			Type:          string(st.Scheduling.Type),
			DayInterval:   st.Scheduling.DayInterval,
			ExecutionTime: st.Scheduling.ExecutionTime,
		}

		if st.Scheduling.InitialDay != nil {
			initialDayStr := st.Scheduling.InitialDay.Time.Format(time.RFC3339)
			schedulingData.InitialDay = &initialDayStr
		}

		schedulingConfigJSON, _ := json.Marshal(schedulingData)
		schedulingConfigStr := string(schedulingConfigJSON)
		avroST.SchedulingConfig = &schedulingConfigStr
	}

	if st.LastExecutedAt != nil {
		t := st.LastExecutedAt.Time
		avroST.LastExecutedAt = &t
	}
	if st.DeletedAt != nil {
		t := st.DeletedAt.Time
		avroST.DeletedAt = &t
	}
	return avroST, nil
}

func (c *ConfluentAvroCodec) convertInternalEvaluationRule(er *domain.EvaluationRule) (*AvroEvaluationRule, error) {
	paramsJSON, _ := json.Marshal(er.Parameters) // You may want to handle error
	avroER := &AvroEvaluationRule{
		ID:          er.ID.String(),
		DeviceID:    "", // Not present in domain.EvaluationRule, set as needed
		Version:     int(er.Version),
		Description: er.Description,
		Kind:        er.Kind,
		Enabled:     er.Enabled,
		Parameters:  string(paramsJSON),
		CreatedAt:   time.Time{}, // Not present in domain.EvaluationRule
		UpdatedAt:   time.Time{}, // Not present in domain.EvaluationRule
	}
	return avroER, nil
}

func (c *ConfluentAvroCodec) convertInternalCommand(cmd *domain.Command) (*AvroCommand, error) {
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

	return avroCmd, nil
}

func (c *ConfluentAvroCodec) convertInternalTenantConfiguration(config *domain.TenantConfiguration) (*AvroTenantConfiguration, error) {
	avroConfig := &AvroTenantConfiguration{
		ID:        string(config.ID),
		TenantID:  string(config.TenantID),
		Timezone:  config.Timezone,
		Version:   config.Version,
		CreatedAt: config.CreatedAt,
		UpdatedAt: config.UpdatedAt,
	}

	return avroConfig, nil
}

// serializeCommandTemplates converts a slice of CommandTemplate to a JSON string
// using only the essential template data without device information
func (c *ConfluentAvroCodec) serializeCommandTemplates(templates []domain.CommandTemplate) string {
	if len(templates) == 0 {
		return "[]"
	}

	// Create a slice of maps to represent the command templates
	var templateMaps []map[string]any
	for _, template := range templates {
		templateMap := map[string]any{
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

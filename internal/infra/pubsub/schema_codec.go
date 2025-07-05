package pubsub

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// SchemaMessage represents the wrapper structure for messages with schema
type SchemaMessage struct {
	Schema  map[string]any `json:"schema"`
	Payload any            `json:"payload"`
}

// SchemaCodec implements Codec interface with Kafka Connect JSON Schema inference
type SchemaCodec struct {
	prototype any
	schema    map[string]any
}

// newSchemaCodec creates a new schema-aware codec
func newSchemaCodec(prototype any) *SchemaCodec {
	return &SchemaCodec{
		prototype: prototype,
		schema:    inferSchema(prototype),
	}
}

// Encode encodes a value into a schema-wrapped JSON message
func (c *SchemaCodec) Encode(value any) ([]byte, error) {
	schemaMessage := SchemaMessage{
		Schema:  c.schema,
		Payload: value,
	}

	data, err := json.Marshal(schemaMessage)
	if err != nil {
		return nil, fmt.Errorf("marshaling schema message: %w", err)
	}

	return data, nil
}

// Decode decodes a schema-wrapped JSON message back to the original value
func (c *SchemaCodec) Decode(data []byte) (any, error) {
	// First, try to decode as a schema-wrapped message
	var schemaMessage SchemaMessage
	err := json.Unmarshal(data, &schemaMessage)
	if err == nil && schemaMessage.Schema != nil && schemaMessage.Payload != nil {
		// This is a schema-wrapped message
		pt := reflect.TypeOf(c.prototype)
		instance := reflect.New(pt).Interface()

		// Unmarshal the payload into the instance
		payloadData, err := json.Marshal(schemaMessage.Payload)
		if err != nil {
			return nil, fmt.Errorf("marshaling payload: %w", err)
		}

		err = json.Unmarshal(payloadData, &instance)
		if err != nil {
			return nil, fmt.Errorf("unmarshaling payload: %w", err)
		}

		return instance, nil
	}

	// If that fails, try to decode as a plain JSON message (backward compatibility)
	pt := reflect.TypeOf(c.prototype)
	instance := reflect.New(pt).Interface()

	err = json.Unmarshal(data, &instance)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling message: %w", err)
	}

	return instance, nil
}

// inferSchema generates a Kafka Connect JSON Schema from a Go struct or map
func inferSchema(prototype any) map[string]any {
	schema := map[string]any{
		"type":   "struct",
		"fields": []map[string]any{},
	}

	t := reflect.TypeOf(prototype)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Handle different types
	switch t.Kind() {
	case reflect.Struct:
		// Handle structs
		fields := []map[string]any{}

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				continue
			}

			// Extract field name from JSON tag
			fieldName := strings.Split(jsonTag, ",")[0]
			if fieldName == "" {
				fieldName = field.Name
			}

			// Infer field schema
			fieldSchema := inferFieldSchema(field.Type)

			// Check if field is optional (has omitempty tag)
			isOptional := strings.Contains(jsonTag, "omitempty")

			fieldDef := map[string]any{
				"field": fieldName,
				"type":  fieldSchema,
			}

			if isOptional {
				fieldDef["optional"] = true
			}

			fields = append(fields, fieldDef)
		}

		schema["fields"] = fields

	case reflect.Map:
		// Handle maps - create a generic object schema
		schema["fields"] = []map[string]any{
			{
				"field":    "value",
				"type":     "string",
				"optional": true,
			},
		}

	default:
		// For other types, create a generic object schema
		schema["fields"] = []map[string]any{
			{
				"field":    "value",
				"type":     "string",
				"optional": true,
			},
		}
	}

	return schema
}

// inferFieldSchema generates a Kafka Connect schema for a specific field type
func inferFieldSchema(t reflect.Type) any {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "int32"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return "int32"
	case reflect.Uint64:
		return "int64"
	case reflect.Float32:
		return "float32"
	case reflect.Float64:
		return "float64"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice, reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 {
			// Special case for []byte
			return "bytes"
		} else {
			// For arrays, return the element type
			return inferFieldSchema(t.Elem())
		}
	case reflect.Map:
		if t.Key().Kind() == reflect.String {
			// Map with string keys
			return map[string]any{
				"type":   "map",
				"values": inferFieldSchema(t.Elem()),
			}
		} else {
			// Generic map
			return "string"
		}
	case reflect.Struct:
		// Handle nested structs
		if t == reflect.TypeOf(json.RawMessage{}) {
			return "bytes"
		} else {
			// Recursively infer schema for nested struct
			return inferSchema(reflect.New(t).Interface())
		}
	case reflect.Ptr:
		// For pointers, infer schema from the underlying type
		return inferFieldSchema(t.Elem())
	case reflect.Interface:
		// For interfaces, we can't infer much
		return "string"
	default:
		return "string"
	}
}

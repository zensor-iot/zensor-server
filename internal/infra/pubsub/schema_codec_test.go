package pubsub

import (
	"encoding/json"
	"testing"
	"time"

	"zensor-server/internal/infra/utils"
)

// TestStruct represents a sample struct for testing
type TestStruct struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Age       int             `json:"age"`
	IsActive  bool            `json:"is_active"`
	Score     float64         `json:"score"`
	Tags      []string        `json:"tags"`
	Metadata  map[string]any  `json:"metadata"`
	CreatedAt utils.Time      `json:"created_at"`
	Optional  *string         `json:"optional,omitempty"`
	RawData   json.RawMessage `json:"raw_data"`
}

func TestSchemaCodec_EncodeDecode(t *testing.T) {
	// Create a test instance
	testData := TestStruct{
		ID:        "test-123",
		Name:      "Test Device",
		Age:       25,
		IsActive:  true,
		Score:     95.5,
		Tags:      []string{"sensor", "temperature"},
		Metadata:  map[string]any{"location": "room-1", "floor": 2},
		CreatedAt: utils.Time{Time: time.Now()},
		Optional:  nil,
		RawData:   json.RawMessage(`{"key": "value"}`),
	}

	// Create schema codec
	codec := newSchemaCodec(TestStruct{})

	// Encode the data
	encoded, err := codec.Encode(testData)
	if err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}

	// Verify the encoded data has the expected structure
	var schemaMessage SchemaMessage
	err = json.Unmarshal(encoded, &schemaMessage)
	if err != nil {
		t.Fatalf("Failed to unmarshal encoded data: %v", err)
	}

	// Check that schema and payload are present
	if schemaMessage.Schema == nil {
		t.Error("Schema is nil")
	}
	if schemaMessage.Payload == nil {
		t.Error("Payload is nil")
	}

	// Verify schema structure (Kafka Connect format)
	schema := schemaMessage.Schema
	if schema["type"] != "struct" {
		t.Errorf("Expected schema type 'struct', got %v", schema["type"])
	}

	// The fields might be []any instead of []map[string]any
	fieldsInterface, ok := schema["fields"]
	if !ok {
		t.Fatal("Schema fields not found")
	}

	fields, ok := fieldsInterface.([]any)
	if !ok {
		t.Fatalf("Schema fields is not a slice, got %T", fieldsInterface)
	}

	// Check that required fields are present
	foundID := false
	for _, fieldInterface := range fields {
		field, ok := fieldInterface.(map[string]any)
		if !ok {
			continue
		}
		if field["field"] == "id" {
			foundID = true
			if field["type"] != "string" {
				t.Errorf("Expected ID type 'string', got %v", field["type"])
			}
			break
		}
	}
	if !foundID {
		t.Error("ID field not found in schema")
	}

	// Decode the data
	decoded, err := codec.Decode(encoded)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	// Verify the decoded data matches the original
	decodedStruct, ok := decoded.(*TestStruct)
	if !ok {
		t.Fatalf("Decoded data is not *TestStruct, got %T", decoded)
	}

	if decodedStruct.ID != testData.ID {
		t.Errorf("ID mismatch: expected %s, got %s", testData.ID, decodedStruct.ID)
	}
	if decodedStruct.Name != testData.Name {
		t.Errorf("Name mismatch: expected %s, got %s", testData.Name, decodedStruct.Name)
	}
	if decodedStruct.Age != testData.Age {
		t.Errorf("Age mismatch: expected %d, got %d", testData.Age, decodedStruct.Age)
	}
	if decodedStruct.IsActive != testData.IsActive {
		t.Errorf("IsActive mismatch: expected %t, got %t", testData.IsActive, decodedStruct.IsActive)
	}
}

func TestSchemaCodec_BackwardCompatibility(t *testing.T) {
	// Create a test instance
	testData := TestStruct{
		ID:       "test-123",
		Name:     "Test Device",
		Age:      25,
		IsActive: true,
	}

	// Create schema codec
	codec := newSchemaCodec(TestStruct{})

	// Encode using the old JSON codec format (plain JSON)
	oldJSONCodec := &JSONCodec{prototype: TestStruct{}}
	encoded, err := oldJSONCodec.Encode(testData)
	if err != nil {
		t.Fatalf("Failed to encode with old codec: %v", err)
	}

	// Decode using the new schema codec (should handle backward compatibility)
	decoded, err := codec.Decode(encoded)
	if err != nil {
		t.Fatalf("Failed to decode with schema codec: %v", err)
	}

	// Verify the decoded data matches the original
	decodedStruct, ok := decoded.(*TestStruct)
	if !ok {
		t.Fatalf("Decoded data is not *TestStruct, got %T", decoded)
	}

	if decodedStruct.ID != testData.ID {
		t.Errorf("ID mismatch: expected %s, got %s", testData.ID, decodedStruct.ID)
	}
	if decodedStruct.Name != testData.Name {
		t.Errorf("Name mismatch: expected %s, got %s", testData.Name, decodedStruct.Name)
	}
}

func TestSchemaCodec_SchemaInference(t *testing.T) {
	codec := newSchemaCodec(TestStruct{})

	// Verify the inferred schema (Kafka Connect format)
	schema := codec.schema
	if schema["type"] != "struct" {
		t.Errorf("Expected schema type 'struct', got %v", schema["type"])
	}

	fieldsInterface, ok := schema["fields"]
	if !ok {
		t.Fatal("Schema fields not found")
	}

	var fieldsAsAny []any
	var fieldsAsMap []map[string]any
	if tmp, ok := fieldsInterface.([]any); ok {
		fieldsAsAny = tmp
	} else if tmp, ok := fieldsInterface.([]map[string]any); ok {
		fieldsAsMap = tmp
	} else {
		t.Fatalf("Schema fields is not a recognized slice type, got %T", fieldsInterface)
	}

	getField := func(name string) (map[string]any, bool) {
		if fieldsAsAny != nil {
			for _, fieldInterface := range fieldsAsAny {
				field, ok := fieldInterface.(map[string]any)
				if !ok {
					continue
				}
				if field["field"] == name {
					return field, true
				}
			}
		} else {
			for _, field := range fieldsAsMap {
				if field["field"] == name {
					return field, true
				}
			}
		}
		return nil, false
	}

	// Check specific field types
	expectedTypes := map[string]string{
		"id":        "string",
		"name":      "string",
		"age":       "int32",
		"is_active": "boolean",
		"score":     "float64",
		"tags":      "string", // arrays become the element type
		"raw_data":  "bytes",  // json.RawMessage becomes bytes
	}

	for fieldName, expectedType := range expectedTypes {
		field, found := getField(fieldName)
		if !found {
			t.Errorf("Field %s not found in schema", fieldName)
			continue
		}
		fieldType := field["type"]
		if fieldType != expectedType {
			t.Errorf("Field %s: expected type %s, got %v", fieldName, expectedType, fieldType)
		}
	}

	// Check complex types that are objects
	complexTypes := map[string]string{
		"metadata":   "map",    // maps with string keys
		"created_at": "struct", // utils.Time is a struct
	}

	for fieldName, expectedType := range complexTypes {
		field, found := getField(fieldName)
		if !found {
			t.Errorf("Field %s not found in schema", fieldName)
			continue
		}
		fieldType := field["type"]
		fieldTypeMap, isMap := fieldType.(map[string]any)
		if !isMap {
			t.Errorf("Field %s: expected type to be a map, got %T", fieldName, fieldType)
			continue
		}
		if fieldTypeMap["type"] != expectedType {
			t.Errorf("Field %s: expected type %s, got %v", fieldName, expectedType, fieldTypeMap["type"])
		}
	}

	// Check optional fields (fields with omitempty)
	optionalFields := []string{"optional"}
	for _, expectedField := range optionalFields {
		field, found := getField(expectedField)
		if !found {
			t.Errorf("Optional field %s not found", expectedField)
			continue
		}
		if optional, exists := field["optional"]; !exists || !optional.(bool) {
			t.Errorf("Field %s should be optional", expectedField)
		}
	}
}

func TestSchemaCodec_MapPrototype(t *testing.T) {
	// Test with map[string]any prototype (like the one used in lora_integration.go)
	codec := newSchemaCodec(map[string]any{})

	// Create a test message
	testMessage := map[string]any{
		"id":          "test-123",
		"device_name": "test-device",
		"payload": map[string]any{
			"index": 1,
			"value": 100,
		},
	}

	// Encode the message
	encoded, err := codec.Encode(testMessage)
	if err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}

	// Verify the encoded data has the expected structure
	var schemaMessage SchemaMessage
	err = json.Unmarshal(encoded, &schemaMessage)
	if err != nil {
		t.Fatalf("Failed to unmarshal encoded data: %v", err)
	}

	// Check that schema and payload are present
	if schemaMessage.Schema == nil {
		t.Error("Schema is nil")
	}
	if schemaMessage.Payload == nil {
		t.Error("Payload is nil")
	}

	// Verify schema structure for map (Kafka Connect format)
	schema := schemaMessage.Schema
	if schema["type"] != "struct" {
		t.Errorf("Expected schema type 'struct', got %v", schema["type"])
	}

	fieldsInterface, ok := schema["fields"]
	if !ok {
		t.Fatal("Schema fields not found")
	}

	fields, ok := fieldsInterface.([]any)
	if !ok {
		t.Fatalf("Schema fields is not a slice, got %T", fieldsInterface)
	}

	// For maps, we expect a single field with generic type
	if len(fields) != 1 {
		t.Errorf("Expected 1 field for map schema, got %d", len(fields))
	}

	field, ok := fields[0].(map[string]any)
	if !ok {
		t.Fatal("First field is not a map")
	}

	if field["field"] != "value" {
		t.Errorf("Expected field name 'value', got %v", field["field"])
	}

	if field["type"] != "string" {
		t.Errorf("Expected field type 'string', got %v", field["type"])
	}

	// Decode the message
	decoded, err := codec.Decode(encoded)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	// Verify the decoded data matches the original
	decodedMap, ok := decoded.(*map[string]any)
	if !ok {
		t.Fatalf("Decoded data is not *map[string]any, got %T", decoded)
	}

	if (*decodedMap)["id"] != testMessage["id"] {
		t.Errorf("ID mismatch: expected %s, got %v", testMessage["id"], (*decodedMap)["id"])
	}
	if (*decodedMap)["device_name"] != testMessage["device_name"] {
		t.Errorf("DeviceName mismatch: expected %s, got %v", testMessage["device_name"], (*decodedMap)["device_name"])
	}
}

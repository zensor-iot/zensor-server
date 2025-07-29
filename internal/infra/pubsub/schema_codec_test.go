package pubsub_test

import (
	"encoding/json"
	"time"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/utils"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
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

var _ = ginkgo.Describe("Schema Codec", func() {
	ginkgo.Context("EncodeDecode", func() {
		var (
			testData TestStruct
			codec    *pubsub.SchemaCodec
		)

		ginkgo.BeforeEach(func() {
			testData = TestStruct{
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

			codec = pubsub.NewSchemaCodec(TestStruct{})
		})

		ginkgo.It("should encode and decode data correctly", func() {
			// Encode the data
			encoded, err := codec.Encode(testData)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Verify the encoded data has the expected structure
			var schemaMessage pubsub.SchemaMessage
			err = json.Unmarshal(encoded, &schemaMessage)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Check that schema and payload are present
			gomega.Expect(schemaMessage.Schema).NotTo(gomega.BeNil())
			gomega.Expect(schemaMessage.Payload).NotTo(gomega.BeNil())

			// Verify schema structure (Kafka Connect format)
			schema := schemaMessage.Schema
			gomega.Expect(schema["type"]).To(gomega.Equal("struct"))

			// The fields might be []any instead of []map[string]any
			fieldsInterface, ok := schema["fields"]
			gomega.Expect(ok).To(gomega.BeTrue())

			fields, ok := fieldsInterface.([]any)
			gomega.Expect(ok).To(gomega.BeTrue())

			// Check that required fields are present
			foundID := false
			for _, fieldInterface := range fields {
				field, ok := fieldInterface.(map[string]any)
				if !ok {
					continue
				}
				if field["field"] == "id" {
					foundID = true
					gomega.Expect(field["type"]).To(gomega.Equal("string"))
					break
				}
			}
			gomega.Expect(foundID).To(gomega.BeTrue())

			// Decode the data
			decoded, err := codec.Decode(encoded)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Verify decoded data matches original
			decodedStruct, ok := decoded.(*TestStruct)
			gomega.Expect(ok).To(gomega.BeTrue())
			gomega.Expect(decodedStruct.ID).To(gomega.Equal(testData.ID))
			gomega.Expect(decodedStruct.Name).To(gomega.Equal(testData.Name))
			gomega.Expect(decodedStruct.Age).To(gomega.Equal(testData.Age))
			gomega.Expect(decodedStruct.IsActive).To(gomega.Equal(testData.IsActive))
			gomega.Expect(decodedStruct.Score).To(gomega.Equal(testData.Score))
			gomega.Expect(decodedStruct.Tags).To(gomega.Equal(testData.Tags))
			// Note: JSON unmarshaling converts numbers to float64, so we need to check values individually
			gomega.Expect(decodedStruct.Metadata["location"]).To(gomega.Equal(testData.Metadata["location"]))
			gomega.Expect(decodedStruct.Metadata["floor"]).To(gomega.Equal(float64(2))) // JSON numbers are float64
		})
	})

	ginkgo.Context("BackwardCompatibility", func() {
		var codec *pubsub.SchemaCodec

		ginkgo.BeforeEach(func() {
			codec = pubsub.NewSchemaCodec(TestStruct{})
		})

		ginkgo.It("should handle backward compatibility scenarios", func() {
			// Create data with optional fields
			testData := TestStruct{
				ID:       "test-123",
				Name:     "Test Device",
				Age:      25,
				IsActive: true,
				Score:    95.5,
				Tags:     []string{"sensor"},
				Metadata: map[string]any{"location": "room-1"},
			}

			// Encode the data
			encoded, err := codec.Encode(testData)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Create a new struct with additional fields (simulating schema evolution)
			type ExtendedTestStruct struct {
				TestStruct
				NewField string `json:"new_field,omitempty"`
			}

			// Decode into extended struct
			decoded, err := codec.Decode(encoded)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Verify original fields are preserved
			decodedStruct, ok := decoded.(*TestStruct)
			gomega.Expect(ok).To(gomega.BeTrue())
			gomega.Expect(decodedStruct.ID).To(gomega.Equal(testData.ID))
			gomega.Expect(decodedStruct.Name).To(gomega.Equal(testData.Name))
			gomega.Expect(decodedStruct.Age).To(gomega.Equal(testData.Age))
		})
	})

	ginkgo.Context("SchemaInference", func() {
		var codec *pubsub.SchemaCodec

		ginkgo.BeforeEach(func() {
			codec = pubsub.NewSchemaCodec(TestStruct{})
		})

		ginkgo.It("should infer schema correctly from struct tags", func() {
			// Encode empty struct to get schema
			encoded, err := codec.Encode(TestStruct{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			var schemaMessage pubsub.SchemaMessage
			err = json.Unmarshal(encoded, &schemaMessage)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			schema := schemaMessage.Schema
			fields, ok := schema["fields"].([]any)
			gomega.Expect(ok).To(gomega.BeTrue())

			// Check that all fields are present in schema
			expectedFields := []string{"id", "name", "age", "is_active", "score", "tags", "metadata", "created_at", "raw_data"}

			for _, fieldName := range expectedFields {
				found := false
				for _, fieldInterface := range fields {
					field, ok := fieldInterface.(map[string]any)
					if !ok {
						continue
					}
					if field["field"] == fieldName {
						found = true
						break
					}
				}
				gomega.Expect(found).To(gomega.BeTrue(), "Field %s not found in schema", fieldName)
			}
		})
	})

	ginkgo.Context("MapPrototype", func() {
		var codec *pubsub.SchemaCodec

		ginkgo.BeforeEach(func() {
			codec = pubsub.NewSchemaCodec(map[string]any{})
		})

		ginkgo.It("should handle map prototype correctly", func() {
			// Create test data
			testData := map[string]any{
				"id":          "test-123",
				"device_name": "test-device",
				"payload": map[string]any{
					"index": 1,
					"value": 100,
				},
			}

			// Encode the data
			encoded, err := codec.Encode(testData)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Decode as map
			decoded, err := codec.Decode(encoded)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Verify map structure
			decodedMap, ok := decoded.(*map[string]any)
			gomega.Expect(ok).To(gomega.BeTrue())
			gomega.Expect((*decodedMap)["id"]).To(gomega.Equal(testData["id"]))
			gomega.Expect((*decodedMap)["device_name"]).To(gomega.Equal(testData["device_name"]))

			// Verify nested structures
			payload, ok := (*decodedMap)["payload"].(map[string]any)
			gomega.Expect(ok).To(gomega.BeTrue())
			gomega.Expect(payload["index"]).To(gomega.Equal(float64(1))) // JSON numbers are float64
			gomega.Expect(payload["value"]).To(gomega.Equal(float64(100)))
		})
	})
})

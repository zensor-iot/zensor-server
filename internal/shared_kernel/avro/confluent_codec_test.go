package avro

import (
	"fmt"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/riferrei/srclient"

	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"
)

var _ = ginkgo.Describe("ConfluentAvroCodec", func() {
	ginkgo.Context("NewConfluentAvroCodec", func() {
		ginkgo.It("should create a new ConfluentAvroCodec", func() {
			// Test that ConfluentAvroCodec can be created
			schemaRegistry := srclient.CreateSchemaRegistryClient("http://localhost:8081")
			codec := NewConfluentAvroCodec(&AvroCommand{}, schemaRegistry)
			gomega.Expect(codec).NotTo(gomega.BeNil())
			gomega.Expect(codec.schemaCache).NotTo(gomega.BeNil())
			gomega.Expect(codec.codecCache).NotTo(gomega.BeNil())
			gomega.Expect(codec.subjectSuffix).To(gomega.Equal("-value"))
		})
	})

	ginkgo.Context("StructValidation", func() {
		ginkgo.When("validating AvroCommand", func() {
			ginkgo.It("should validate AvroCommand struct", func() {
				cmd := &AvroCommand{
					ID:            "cmd-123",
					Version:       1,
					DeviceName:    "test-device",
					DeviceID:      "dev-123",
					TaskID:        "task-123",
					PayloadIndex:  1,
					PayloadValue:  100,
					DispatchAfter: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					Port:          80,
					Priority:      "high",
					CreatedAt:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					Ready:         true,
					Sent:          false,
					SentAt:        time.Time{},
				}
				gomega.Expect(cmd.ID).To(gomega.Equal("cmd-123"))
				gomega.Expect(cmd.Version).To(gomega.Equal(1))
				gomega.Expect(cmd.DeviceName).To(gomega.Equal("test-device"))
			})
		})

		ginkgo.When("validating AvroTask", func() {
			ginkgo.It("should validate AvroTask struct", func() {
				scheduledTaskID := "scheduled-task-123"
				task := &AvroTask{
					ID:              "task-123",
					DeviceID:        "dev-123",
					ScheduledTaskID: &scheduledTaskID,
					Version:         1,
					CreatedAt:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				}
				gomega.Expect(task.ID).To(gomega.Equal("task-123"))
				gomega.Expect(task.DeviceID).To(gomega.Equal("dev-123"))
				gomega.Expect(task.ScheduledTaskID).To(gomega.Equal(&scheduledTaskID))
			})
		})

		ginkgo.When("validating AvroDevice", func() {
			ginkgo.It("should validate AvroDevice struct", func() {
				tenantID := "tenant-123"
				lastMessageReceivedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
				device := &AvroDevice{
					ID:                    "dev-123",
					Version:               1,
					Name:                  "test-device",
					DisplayName:           "Test Device",
					AppEUI:                "app-eui-123",
					DevEUI:                "dev-eui-123",
					AppKey:                "app-key-123",
					TenantID:              &tenantID,
					LastMessageReceivedAt: &lastMessageReceivedAt,
					CreatedAt:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				}
				gomega.Expect(device.ID).To(gomega.Equal("dev-123"))
				gomega.Expect(device.Version).To(gomega.Equal(1))
				gomega.Expect(device.Name).To(gomega.Equal("test-device"))
				gomega.Expect(device.TenantID).To(gomega.Equal(&tenantID))
			})
		})

		ginkgo.When("validating AvroScheduledTask", func() {
			ginkgo.It("should validate AvroScheduledTask struct", func() {
				lastExecutedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
				scheduledTask := &AvroScheduledTask{
					ID:               "scheduled-task-123",
					Version:          1,
					TenantID:         "tenant-123",
					DeviceID:         "dev-123",
					CommandTemplates: "template-123",
					Schedule:         "0 0 * * *",
					IsActive:         true,
					CreatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					LastExecutedAt:   &lastExecutedAt,
					DeletedAt:        nil,
				}
				gomega.Expect(scheduledTask.ID).To(gomega.Equal("scheduled-task-123"))
				gomega.Expect(scheduledTask.Version).To(gomega.Equal(1))
				gomega.Expect(scheduledTask.TenantID).To(gomega.Equal("tenant-123"))
				gomega.Expect(scheduledTask.DeviceID).To(gomega.Equal("dev-123"))
				gomega.Expect(scheduledTask.CommandTemplates).To(gomega.Equal("template-123"))
				gomega.Expect(scheduledTask.Schedule).To(gomega.Equal("0 0 * * *"))
				gomega.Expect(scheduledTask.IsActive).To(gomega.BeTrue())
				gomega.Expect(scheduledTask.LastExecutedAt).To(gomega.Equal(&lastExecutedAt))
				gomega.Expect(scheduledTask.DeletedAt).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Context("InvalidData", func() {
		ginkgo.It("should handle invalid data gracefully", func() {
			schemaRegistry := srclient.CreateSchemaRegistryClient("http://localhost:8081")
			codec := NewConfluentAvroCodec(&AvroCommand{}, schemaRegistry)

			// Test with invalid data
			invalidData := []byte("invalid avro data")
			_, err := codec.Decode(invalidData)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})

	ginkgo.Context("UnsupportedType", func() {
		ginkgo.It("should handle unsupported types", func() {
			schemaRegistry := srclient.CreateSchemaRegistryClient("http://localhost:8081")
			codec := NewConfluentAvroCodec(&AvroCommand{}, schemaRegistry)

			// Test with unsupported type
			unsupportedData := "unsupported string data"
			_, err := codec.Encode(unsupportedData)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})

	ginkgo.Context("DeviceWithLastMessageReceivedAt", func() {
		ginkgo.It("should handle device with last message received at", func() {
			schemaRegistry := srclient.CreateSchemaRegistryClient("http://localhost:8081")
			codec := NewConfluentAvroCodec(&AvroDevice{}, schemaRegistry)

			lastMessageReceivedAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
			tenantID := "tenant-123"

			device := &AvroDevice{
				ID:                    "dev-123",
				Version:               1,
				Name:                  "test-device",
				DisplayName:           "Test Device",
				AppEUI:                "app-eui-123",
				DevEUI:                "dev-eui-123",
				AppKey:                "app-key-123",
				TenantID:              &tenantID,
				LastMessageReceivedAt: &lastMessageReceivedAt,
				CreatedAt:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			}

			// Encode the device
			encoded, err := codec.Encode(device)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(encoded).NotTo(gomega.BeNil())

			// Decode the device
			decoded, err := codec.Decode(encoded)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(decoded).NotTo(gomega.BeNil())

			// Verify the decoded device
			decodedDevice := decoded.(*AvroDevice)
			gomega.Expect(decodedDevice.ID).To(gomega.Equal(device.ID))
			gomega.Expect(decodedDevice.Name).To(gomega.Equal(device.Name))
			gomega.Expect(decodedDevice.LastMessageReceivedAt).NotTo(gomega.BeNil())
			gomega.Expect(decodedDevice.LastMessageReceivedAt.Unix()).To(gomega.Equal(lastMessageReceivedAt.Unix()))
		})
	})

	ginkgo.Context("ConvertDomainDevice", func() {
		ginkgo.It("should convert domain device to Avro device", func() {
			schemaRegistry := srclient.CreateSchemaRegistryClient("http://localhost:8081")
			codec := NewConfluentAvroCodec(&AvroDevice{}, schemaRegistry)

			lastMessageReceivedAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
			tenantID := domain.ID("tenant-123")

			domainDevice := domain.Device{
				ID:                    domain.ID("dev-123"),
				Name:                  "test-device",
				DisplayName:           "Test Device",
				AppEUI:                "app-eui-123",
				DevEUI:                "dev-eui-123",
				AppKey:                "app-key-123",
				TenantID:              &tenantID,
				LastMessageReceivedAt: utils.Time{Time: lastMessageReceivedAt},
			}

			// Convert to Avro device
			avroDevice, err := codec.convertInternalDevice(&domainDevice)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Verify the conversion
			gomega.Expect(avroDevice.ID).To(gomega.Equal(string(domainDevice.ID)))
			gomega.Expect(avroDevice.Name).To(gomega.Equal(domainDevice.Name))
			gomega.Expect(avroDevice.DisplayName).To(gomega.Equal(domainDevice.DisplayName))
			gomega.Expect(avroDevice.AppEUI).To(gomega.Equal(domainDevice.AppEUI))
			gomega.Expect(avroDevice.DevEUI).To(gomega.Equal(domainDevice.DevEUI))
			gomega.Expect(avroDevice.AppKey).To(gomega.Equal(domainDevice.AppKey))
			gomega.Expect(*avroDevice.TenantID).To(gomega.Equal(string(*domainDevice.TenantID)))
			gomega.Expect(avroDevice.LastMessageReceivedAt).NotTo(gomega.BeNil())
			gomega.Expect(avroDevice.LastMessageReceivedAt.Unix()).To(gomega.Equal(lastMessageReceivedAt.Unix()))
		})
	})

	ginkgo.Context("SerializeCommandTemplates", func() {
		ginkgo.It("should serialize command templates correctly", func() {
			schemaRegistry := srclient.CreateSchemaRegistryClient("http://localhost:8081")
			codec := NewConfluentAvroCodec(&AvroScheduledTask{}, schemaRegistry)

			commandTemplates := []domain.CommandTemplate{
				{
					Index: 1,
					Value: 100,
					Port:  15,
					Delay: 1000,
				},
				{
					Index: 2,
					Value: 200,
					Port:  16,
					Delay: 2000,
				},
			}

			// Serialize command templates
			serialized := codec.SerializeCommandTemplates(commandTemplates)

			// Verify the serialization
			gomega.Expect(serialized).NotTo(gomega.BeEmpty())
			gomega.Expect(serialized).To(gomega.ContainSubstring("1"))
			gomega.Expect(serialized).To(gomega.ContainSubstring("100"))
			gomega.Expect(serialized).To(gomega.ContainSubstring("15"))
			gomega.Expect(serialized).To(gomega.ContainSubstring("1000"))
			gomega.Expect(serialized).To(gomega.ContainSubstring("2"))
			gomega.Expect(serialized).To(gomega.ContainSubstring("200"))
			gomega.Expect(serialized).To(gomega.ContainSubstring("16"))
			gomega.Expect(serialized).To(gomega.ContainSubstring("2000"))
		})
	})

	ginkgo.Context("MapConversion", func() {
		ginkgo.It("should handle map conversion correctly", func() {
			schemaRegistry := srclient.CreateSchemaRegistryClient("http://localhost:8081")
			codec := NewConfluentAvroCodec(&AvroCommand{}, schemaRegistry)

			// Create a map with various types
			testMap := map[string]interface{}{
				"string": "test",
				"int":    123,
				"float":  123.45,
				"bool":   true,
				"null":   nil,
				"array":  []interface{}{1, 2, 3},
				"nested": map[string]interface{}{
					"key": "value",
				},
			}

			// Test map conversion
			converted := codec.ConvertMap(testMap)

			// Verify the conversion
			gomega.Expect(converted).NotTo(gomega.BeNil())
			gomega.Expect(converted["string"]).To(gomega.Equal("test"))
			gomega.Expect(converted["int"]).To(gomega.Equal(123))
			gomega.Expect(converted["float"]).To(gomega.Equal(123.45))
			gomega.Expect(converted["bool"]).To(gomega.Equal(true))
			gomega.Expect(converted["null"]).To(gomega.BeNil())
			gomega.Expect(converted["array"]).To(gomega.Equal([]interface{}{1, 2, 3}))
			gomega.Expect(converted["nested"]).To(gomega.Equal(map[string]interface{}{"key": "value"}))
		})
	})

	ginkgo.Context("MapConversionWithInt32", func() {
		ginkgo.It("should handle map conversion with int32 correctly", func() {
			schemaRegistry := srclient.CreateSchemaRegistryClient("http://localhost:8081")
			codec := NewConfluentAvroCodec(&AvroCommand{}, schemaRegistry)

			// Create a map with int32 values
			testMap := map[string]interface{}{
				"int32_1": int32(123),
				"int32_2": int32(456),
				"int64_1": int64(789),
				"int_1":   999,
			}

			// Test map conversion
			converted := codec.ConvertMap(testMap)

			// Verify the conversion
			gomega.Expect(converted).NotTo(gomega.BeNil())
			gomega.Expect(converted["int32_1"]).To(gomega.Equal(int32(123)))
			gomega.Expect(converted["int32_2"]).To(gomega.Equal(int32(456)))
			gomega.Expect(converted["int64_1"]).To(gomega.Equal(int64(789)))
			gomega.Expect(converted["int_1"]).To(gomega.Equal(999))
		})
	})

	ginkgo.Context("WithMockSchemaRegistry", func() {
		ginkgo.It("should work with mock schema registry", func() {
			mockRegistry := NewMockSchemaRegistry()
			codec := NewConfluentAvroCodec(&AvroCommand{}, mockRegistry)

			gomega.Expect(codec).NotTo(gomega.BeNil())
			gomega.Expect(codec.schemaRegistry).To(gomega.Equal(mockRegistry))
		})
	})

	ginkgo.Context("DeletedAtFieldEncoding", func() {
		ginkgo.It("should handle DeletedAt field encoding correctly", func() {
			schemaRegistry := srclient.CreateSchemaRegistryClient("http://localhost:8081")
			codec := NewConfluentAvroCodec(&AvroScheduledTask{}, schemaRegistry)

			// Test with nil DeletedAt
			scheduledTask1 := &AvroScheduledTask{
				ID:               "task-1",
				Version:          1,
				TenantID:         "tenant-123",
				DeviceID:         "dev-123",
				CommandTemplates: "template-123",
				Schedule:         "0 0 * * *",
				IsActive:         true,
				CreatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				LastExecutedAt:   nil,
				DeletedAt:        nil,
			}

			// Encode and decode
			encoded1, err := codec.Encode(scheduledTask1)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			decoded1, err := codec.Decode(encoded1)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			decodedTask1 := decoded1.(*AvroScheduledTask)
			gomega.Expect(decodedTask1.DeletedAt).To(gomega.BeNil())

			// Test with non-nil DeletedAt
			deletedAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
			scheduledTask2 := &AvroScheduledTask{
				ID:               "task-2",
				Version:          1,
				TenantID:         "tenant-123",
				DeviceID:         "dev-123",
				CommandTemplates: "template-123",
				Schedule:         "0 0 * * *",
				IsActive:         false,
				CreatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				LastExecutedAt:   nil,
				DeletedAt:        &deletedAt,
			}

			// Encode and decode
			encoded2, err := codec.Encode(scheduledTask2)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			decoded2, err := codec.Decode(encoded2)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			decodedTask2 := decoded2.(*AvroScheduledTask)
			gomega.Expect(decodedTask2.DeletedAt).NotTo(gomega.BeNil())
			gomega.Expect(decodedTask2.DeletedAt.Unix()).To(gomega.Equal(deletedAt.Unix()))
		})
	})
})

// MockSchemaRegistry is a mock implementation of the schema registry for testing
type MockSchemaRegistry struct {
	schemas map[string]*srclient.Schema
}

func NewMockSchemaRegistry() *MockSchemaRegistry {
	return &MockSchemaRegistry{
		schemas: make(map[string]*srclient.Schema),
	}
}

func (m *MockSchemaRegistry) GetLatestSchema(subject string) (*srclient.Schema, error) {
	if schema, exists := m.schemas[subject]; exists {
		return schema, nil
	}
	return nil, fmt.Errorf("schema not found for subject: %s", subject)
}

func (m *MockSchemaRegistry) CreateSchema(subject string, schema string, schemaType srclient.SchemaType, references ...srclient.Reference) (*srclient.Schema, error) {
	newSchema := &srclient.Schema{
		ID:      len(m.schemas) + 1,
		Subject: subject,
		Schema:  schema,
		Version: 1,
	}
	m.schemas[subject] = newSchema
	return newSchema, nil
}

func (m *MockSchemaRegistry) GetSchema(schemaID int) (*srclient.Schema, error) {
	for _, schema := range m.schemas {
		if schema.ID == schemaID {
			return schema, nil
		}
	}
	return nil, fmt.Errorf("schema not found for ID: %d", schemaID)
}

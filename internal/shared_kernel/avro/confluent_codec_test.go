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
					CommandTemplates: `[{"port":15,"priority":"NORMAL","payload":{"index":1,"value":100},"wait_for":"5s"}]`,
					Schedule:         "* * * * *",
					IsActive:         true,
					LastExecutedAt:   &lastExecutedAt,
					CreatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				}
				gomega.Expect(scheduledTask.ID).To(gomega.Equal("scheduled-task-123"))
				gomega.Expect(scheduledTask.Version).To(gomega.Equal(int64(1)))
				gomega.Expect(scheduledTask.TenantID).To(gomega.Equal("tenant-123"))
				gomega.Expect(scheduledTask.DeviceID).To(gomega.Equal("dev-123"))
			})
		})

		ginkgo.When("validating AvroTenant", func() {
			ginkgo.It("should validate AvroTenant struct", func() {
				tenant := &AvroTenant{
					ID:        "tenant-123",
					Version:   1,
					Name:      "Test Tenant",
					Email:     "test@example.com",
					IsActive:  true,
					CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				}
				gomega.Expect(tenant.ID).To(gomega.Equal("tenant-123"))
				gomega.Expect(tenant.Version).To(gomega.Equal(1))
				gomega.Expect(tenant.Name).To(gomega.Equal("Test Tenant"))
				gomega.Expect(tenant.Email).To(gomega.Equal("test@example.com"))
			})
		})

		ginkgo.When("validating AvroEvaluationRule", func() {
			ginkgo.It("should validate AvroEvaluationRule struct", func() {
				evaluationRule := &AvroEvaluationRule{
					ID:          "rule-123",
					DeviceID:    "device-123",
					Version:     1,
					Description: "Test Description",
					Kind:        "temperature_alert",
					Enabled:     true,
					Parameters:  `{"threshold": 25}`,
					CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				}
				gomega.Expect(evaluationRule.ID).To(gomega.Equal("rule-123"))
				gomega.Expect(evaluationRule.Version).To(gomega.Equal(1))
				gomega.Expect(evaluationRule.DeviceID).To(gomega.Equal("device-123"))
				gomega.Expect(evaluationRule.Description).To(gomega.Equal("Test Description"))
			})
		})

		ginkgo.When("validating AvroTenantConfiguration", func() {
			ginkgo.It("should validate AvroTenantConfiguration struct", func() {
				tenantConfiguration := &AvroTenantConfiguration{
					ID:        "config-123",
					Version:   1,
					TenantID:  "tenant-123",
					Timezone:  "UTC",
					CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				}
				gomega.Expect(tenantConfiguration.ID).To(gomega.Equal("config-123"))
				gomega.Expect(tenantConfiguration.Version).To(gomega.Equal(1))
				gomega.Expect(tenantConfiguration.TenantID).To(gomega.Equal("tenant-123"))
				gomega.Expect(tenantConfiguration.Timezone).To(gomega.Equal("UTC"))
			})
		})
	})

	ginkgo.Context("DomainConversion", func() {
		ginkgo.When("converting domain.Command to AvroCommand", func() {
			ginkgo.It("should convert domain.Command correctly", func() {
				schemaRegistry := srclient.CreateSchemaRegistryClient("http://localhost:8081")
				codec := NewConfluentAvroCodec(&AvroCommand{}, schemaRegistry)

				// Create a domain command
				device := domain.Device{
					ID:   "dev-123",
					Name: "test-device",
				}
				task := domain.Task{
					ID: "task-123",
				}
				payload := domain.CommandPayload{
					Index: 1,
					Value: 100,
				}
				dispatchAfter := utils.Time{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
				createdAt := utils.Time{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}

				domainCommand := &domain.Command{
					ID:            "cmd-123",
					Version:       1,
					Device:        device,
					Task:          task,
					Port:          80,
					Priority:      "NORMAL",
					Payload:       payload,
					DispatchAfter: dispatchAfter,
					CreatedAt:     createdAt,
					Ready:         true,
					Sent:          false,
				}

				// Convert to AvroCommand
				avroCommand, err := codec.convertInternalCommand(domainCommand)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(avroCommand).NotTo(gomega.BeNil())

				// Verify the conversion
				gomega.Expect(avroCommand.ID).To(gomega.Equal("cmd-123"))
				gomega.Expect(avroCommand.Version).To(gomega.Equal(1))
				gomega.Expect(avroCommand.DeviceName).To(gomega.Equal("test-device"))
				gomega.Expect(avroCommand.DeviceID).To(gomega.Equal("dev-123"))
				gomega.Expect(avroCommand.TaskID).To(gomega.Equal("task-123"))
				gomega.Expect(avroCommand.PayloadIndex).To(gomega.Equal(1))
				gomega.Expect(avroCommand.PayloadValue).To(gomega.Equal(100))
				gomega.Expect(avroCommand.Port).To(gomega.Equal(80))
				gomega.Expect(avroCommand.Priority).To(gomega.Equal("NORMAL"))
				gomega.Expect(avroCommand.Ready).To(gomega.BeTrue())
				gomega.Expect(avroCommand.Sent).To(gomega.BeFalse())
			})
		})

		ginkgo.When("converting domain.Device to AvroDevice", func() {
			ginkgo.It("should convert domain.Device correctly", func() {
				schemaRegistry := srclient.CreateSchemaRegistryClient("http://localhost:8081")
				codec := NewConfluentAvroCodec(&AvroDevice{}, schemaRegistry)

				// Create a domain device
				tenantID := domain.ID("tenant-123")
				lastMessageReceivedAt := utils.Time{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}

				domainDevice := &domain.Device{
					ID:                    "dev-123",
					Name:                  "test-device",
					DisplayName:           "Test Device",
					AppEUI:                "app-eui-123",
					DevEUI:                "dev-eui-123",
					AppKey:                "app-key-123",
					TenantID:              &tenantID,
					LastMessageReceivedAt: lastMessageReceivedAt,
				}

				// Convert to AvroDevice
				avroDevice, err := codec.convertInternalDevice(domainDevice)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(avroDevice).NotTo(gomega.BeNil())

				// Verify the conversion
				gomega.Expect(avroDevice.ID).To(gomega.Equal("dev-123"))
				gomega.Expect(avroDevice.Version).To(gomega.Equal(1))
				gomega.Expect(avroDevice.Name).To(gomega.Equal("test-device"))
				gomega.Expect(avroDevice.DisplayName).To(gomega.Equal("Test Device"))
				gomega.Expect(avroDevice.AppEUI).To(gomega.Equal("app-eui-123"))
				gomega.Expect(avroDevice.DevEUI).To(gomega.Equal("dev-eui-123"))
				gomega.Expect(avroDevice.AppKey).To(gomega.Equal("app-key-123"))
				gomega.Expect(*avroDevice.TenantID).To(gomega.Equal("tenant-123"))
				gomega.Expect(avroDevice.LastMessageReceivedAt.Unix()).To(gomega.Equal(lastMessageReceivedAt.Unix()))
			})
		})
	})

	ginkgo.Context("CommandTemplateSerialization", func() {
		ginkgo.It("should serialize command templates correctly", func() {
			schemaRegistry := srclient.CreateSchemaRegistryClient("http://localhost:8081")
			codec := NewConfluentAvroCodec(&AvroScheduledTask{}, schemaRegistry)

			// Create command templates using the correct domain structure
			device := domain.Device{ID: "dev-123", Name: "test-device"}
			payload1 := domain.CommandPayload{Index: 1, Value: 100}
			payload2 := domain.CommandPayload{Index: 2, Value: 200}

			commandTemplates := []domain.CommandTemplate{
				{
					Device:   device,
					Port:     15,
					Priority: "NORMAL",
					Payload:  payload1,
					WaitFor:  5 * time.Second,
				},
				{
					Device:   device,
					Port:     16,
					Priority: "HIGH",
					Payload:  payload2,
					WaitFor:  10 * time.Second,
				},
			}

			// Serialize command templates using the private method
			serialized := codec.serializeCommandTemplates(commandTemplates)

			// Verify the serialization
			gomega.Expect(serialized).NotTo(gomega.BeEmpty())
			gomega.Expect(serialized).To(gomega.ContainSubstring("1"))
			gomega.Expect(serialized).To(gomega.ContainSubstring("100"))
			gomega.Expect(serialized).To(gomega.ContainSubstring("15"))
			gomega.Expect(serialized).To(gomega.ContainSubstring("5s"))
			gomega.Expect(serialized).To(gomega.ContainSubstring("2"))
			gomega.Expect(serialized).To(gomega.ContainSubstring("200"))
			gomega.Expect(serialized).To(gomega.ContainSubstring("16"))
			gomega.Expect(serialized).To(gomega.ContainSubstring("10s"))
		})
	})

	ginkgo.Context("WithMockSchemaRegistry", func() {
		ginkgo.It("should work with mock schema registry", func() {
			mockRegistry := NewMockSchemaRegistry()
			codec := NewConfluentAvroCodec(&AvroCommand{}, mockRegistry)

			// Test that the codec can be created with mock registry
			gomega.Expect(codec).NotTo(gomega.BeNil())
			gomega.Expect(codec.schemaRegistry).To(gomega.Equal(mockRegistry))
		})
	})
})

// MockSchemaRegistry is a mock implementation of SchemaRegistry for testing
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
	// Create a simple mock schema without using unexported fields
	mockSchema := &srclient.Schema{}
	m.schemas[subject] = mockSchema
	return mockSchema, nil
}

func (m *MockSchemaRegistry) GetSchema(schemaID int) (*srclient.Schema, error) {
	// For mock purposes, return a simple schema
	return &srclient.Schema{}, nil
}

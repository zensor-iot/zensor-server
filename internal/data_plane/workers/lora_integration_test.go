package workers

import (
	"context"
	"time"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/avro"
	"zensor-server/internal/shared_kernel/device"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("LoraIntegrationWorker", func() {
	ginkgo.Context("ConvertToSharedCommand", func() {
		var worker *LoraIntegrationWorker

		ginkgo.BeforeEach(func() {
			worker = &LoraIntegrationWorker{}
		})

		ginkgo.When("converting AvroCommand to shared command", func() {
			ginkgo.It("should convert AvroCommand correctly", func() {
				// Test with AvroCommand
				avroCmd := &avro.AvroCommand{
					ID:            "test-command-id",
					Version:       2,
					DeviceID:      "test-device-id",
					DeviceName:    "test-device",
					TaskID:        "test-task-id",
					PayloadIndex:  5,
					PayloadValue:  123,
					DispatchAfter: time.Now(),
					Port:          16,
					Priority:      "HIGH",
					CreatedAt:     time.Now(),
					Ready:         true,
					Sent:          false,
					SentAt:        time.Time{},
				}

				command, err := worker.convertToSharedCommand(context.TODO(), avroCmd)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Verify that all fields are preserved correctly
				gomega.Expect(command.ID).To(gomega.Equal("test-command-id"))
				gomega.Expect(command.DeviceID).To(gomega.Equal("test-device-id"))
				gomega.Expect(command.DeviceName).To(gomega.Equal("test-device"))
				gomega.Expect(command.TaskID).To(gomega.Equal("test-task-id"))
				gomega.Expect(command.Payload.Index).To(gomega.Equal(uint8(5)))
				gomega.Expect(command.Payload.Value).To(gomega.Equal(uint8(123)))
				gomega.Expect(command.Port).To(gomega.Equal(uint8(16)))
				gomega.Expect(command.Priority).To(gomega.Equal("HIGH"))
				gomega.Expect(command.Ready).To(gomega.BeTrue())
				gomega.Expect(command.Sent).To(gomega.BeFalse())
			})
		})

		ginkgo.When("converting struct message to shared command", func() {
			ginkgo.It("should convert device.Command correctly", func() {
				// Test with device.Command (fallback case)
				structMessage := device.Command{
					ID:         "test-456",
					Version:    2,
					DeviceID:   "device-456",
					DeviceName: "test-device-2",
					TaskID:     "task-456",
					Payload: device.CommandPayload{
						Index: 2,
						Value: 200,
					},
					DispatchAfter: utils.Time{Time: time.Now()},
					Port:          16,
					Priority:      "HIGH",
					CreatedAt:     utils.Time{Time: time.Now()},
					Ready:         false,
					Sent:          true,
					SentAt:        utils.Time{Time: time.Now()},
				}

				command, err := worker.convertToSharedCommand(context.TODO(), structMessage)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				gomega.Expect(command.ID).To(gomega.Equal("test-456"))
				gomega.Expect(command.DeviceName).To(gomega.Equal("test-device-2"))
				gomega.Expect(command.Payload.Index).To(gomega.Equal(uint8(2)))
				gomega.Expect(command.Payload.Value).To(gomega.Equal(uint8(200)))
				gomega.Expect(command.Port).To(gomega.Equal(uint8(16)))
			})
		})
	})
})

package workers

import (
	"time"

	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("LoraIntegrationWorker", func() {
	ginkgo.Context("domainCommandToDeviceCommand", func() {
		ginkgo.When("converting a domain.Command to a device.Command", func() {
			ginkgo.It("should preserve all fields correctly", func() {
				cmd := domain.Command{
					ID:      domain.ID("test-command-id"),
					Version: domain.Version(2),
					Device:  domain.Device{ID: domain.ID("test-device-id"), Name: "test-device"},
					Task:    domain.Task{ID: domain.ID("test-task-id")},
					Payload: domain.CommandPayload{
						Index: domain.Index(5),
						Value: domain.CommandValue(123),
					},
					DispatchAfter: utils.Time{Time: time.Now()},
					Port:          domain.Port(16),
					Priority:      domain.CommandPriority("HIGH"),
					CreatedAt:     utils.Time{Time: time.Now()},
					Ready:         true,
					Sent:          false,
				}

				command := domainCommandToDeviceCommand(cmd)

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
	})
})

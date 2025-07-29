package dto_test

import (
	"time"
	"zensor-server/internal/data_plane/dto"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("CommandStatusUpdateDTO", func() {
	ginkgo.Context("ToDomain", func() {
		var (
			commandStatusDTO dto.CommandStatusUpdateDTO
			errorMsg         string
		)

		ginkgo.When("converting DTO to domain object", func() {
			ginkgo.It("should convert successfully without error message", func() {
				commandStatusDTO = dto.CommandStatusUpdateDTO{
					CommandID:    "test-command-123",
					DeviceName:   "test-device",
					Status:       "queued",
					ErrorMessage: nil,
					Timestamp:    time.Now(),
				}

				// Convert to domain
				domainObj := commandStatusDTO.ToDomain()

				// Verify conversion
				gomega.Expect(domainObj.CommandID).To(gomega.Equal("test-command-123"))
				gomega.Expect(domainObj.DeviceName).To(gomega.Equal("test-device"))
				gomega.Expect(domainObj.Status).To(gomega.Equal(domain.CommandStatusQueued))
				gomega.Expect(domainObj.ErrorMessage).To(gomega.BeNil())
				gomega.Expect(domainObj.Timestamp).To(gomega.Equal(commandStatusDTO.Timestamp))
			})

			ginkgo.It("should convert successfully with error message", func() {
				errorMsg = "command failed"
				commandStatusDTO = dto.CommandStatusUpdateDTO{
					CommandID:    "test-command-789",
					DeviceName:   "test-device-3",
					Status:       "failed",
					ErrorMessage: &errorMsg,
					Timestamp:    time.Now(),
				}

				// Convert to domain
				domainObj := commandStatusDTO.ToDomain()

				// Verify conversion
				gomega.Expect(domainObj.CommandID).To(gomega.Equal("test-command-789"))
				gomega.Expect(domainObj.DeviceName).To(gomega.Equal("test-device-3"))
				gomega.Expect(domainObj.Status).To(gomega.Equal(domain.CommandStatusFailed))
				gomega.Expect(domainObj.ErrorMessage).To(gomega.Equal(&errorMsg))
				gomega.Expect(domainObj.Timestamp).To(gomega.Equal(commandStatusDTO.Timestamp))
			})
		})
	})

	ginkgo.Context("FromDomain", func() {
		var domainObj domain.CommandStatusUpdate

		ginkgo.When("converting domain object to DTO", func() {
			ginkgo.It("should convert successfully", func() {
				domainObj = domain.CommandStatusUpdate{
					CommandID:    "test-command-456",
					DeviceName:   "test-device-2",
					Status:       domain.CommandStatusSent,
					ErrorMessage: nil,
					Timestamp:    time.Now(),
				}

				// Convert to DTO
				dto := dto.FromDomain(domainObj)

				// Verify conversion
				gomega.Expect(dto.CommandID).To(gomega.Equal("test-command-456"))
				gomega.Expect(dto.DeviceName).To(gomega.Equal("test-device-2"))
				gomega.Expect(dto.Status).To(gomega.Equal("confirmed"))
				gomega.Expect(dto.ErrorMessage).To(gomega.BeNil())
				gomega.Expect(dto.Timestamp).To(gomega.Equal(domainObj.Timestamp))
			})
		})
	})
})

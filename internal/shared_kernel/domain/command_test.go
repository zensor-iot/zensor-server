package domain_test

import (
	"time"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Command", func() {
	ginkgo.Context("CommandBuilder", func() {
		var cmd domain.Command

		ginkgo.When("building a command", func() {
			ginkgo.It("should set CreatedAt correctly", func() {
				// Create a command using the builder
				var err error
				cmd, err = domain.NewCommandBuilder().
					WithDevice(domain.Device{ID: "test-device", Name: "Test Device"}).
					WithPayload(domain.CommandPayload{Index: 1, Value: 100}).
					Build()

				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Verify that CreatedAt is set and is recent
				gomega.Expect(cmd.CreatedAt.Time.IsZero()).To(gomega.BeFalse())

				// Verify that CreatedAt is within the last second
				now := time.Now()
				gomega.Expect(cmd.CreatedAt.Time.After(now)).To(gomega.BeFalse())
				gomega.Expect(now.Sub(cmd.CreatedAt.Time) <= time.Second).To(gomega.BeTrue())

				// Verify other fields are set correctly
				gomega.Expect(cmd.ID).NotTo(gomega.BeEmpty())
				gomega.Expect(cmd.Version).To(gomega.Equal(domain.Version(1)))
				gomega.Expect(cmd.Device.ID).To(gomega.Equal(domain.ID("test-device")))
				// Task field is not set by the builder, so we don't check it
				gomega.Expect(cmd.Payload.Index).To(gomega.Equal(domain.Index(1)))
				gomega.Expect(cmd.Payload.Value).To(gomega.Equal(domain.CommandValue(100)))
			})
		})
	})

	ginkgo.Context("UpdateStatus", func() {
		var cmd domain.Command

		ginkgo.When("updating command status to queued", func() {
			ginkgo.It("should set QueuedAt correctly", func() {
				// Create a command
				var err error
				cmd, err = domain.NewCommandBuilder().
					WithDevice(domain.Device{ID: "test-device", Name: "Test Device"}).
					WithPayload(domain.CommandPayload{Index: 1, Value: 100}).
					Build()

				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Initially, QueuedAt should be nil
				gomega.Expect(cmd.QueuedAt).To(gomega.BeNil())

				// Update status to queued
				cmd.UpdateStatus(domain.CommandStatusQueued, nil)

				// Verify that QueuedAt is now set
				gomega.Expect(cmd.QueuedAt).NotTo(gomega.BeNil())

				// Verify that QueuedAt is recent
				now := time.Now()
				gomega.Expect(cmd.QueuedAt.Time.After(now)).To(gomega.BeFalse())
				gomega.Expect(now.Sub(cmd.QueuedAt.Time) <= time.Second).To(gomega.BeTrue())

				// Verify that status is updated
				gomega.Expect(cmd.Status).To(gomega.Equal(domain.CommandStatusQueued))

				// Verify that version is incremented
				gomega.Expect(cmd.Version).To(gomega.Equal(domain.Version(2)))
			})
		})
	})
})

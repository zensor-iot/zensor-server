package persistence_test

import (
	"context"
	"zensor-server/internal/control_plane/persistence"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("DeviceRepository", func() {
	var (
		orm  sql.ORM
		repo usecases.DeviceRepository
		ctx  context.Context
	)

	ginkgo.BeforeEach(func() {
		var err error
		orm, err = sql.NewMemoryORM("migrations")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		repo, err = persistence.NewDeviceRepository(orm)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(repo).NotTo(gomega.BeNil())

		ctx = context.Background()
	})

	ginkgo.Context("CreateDevice", func() {
		var device domain.Device

		ginkgo.BeforeEach(func() {
			id := utils.GenerateUUID()
			device = domain.Device{
				ID:   domain.ID(id),
				Name: "device-" + id,
			}
		})

		ginkgo.It("should create the device in the database", func() {
			err := repo.CreateDevice(ctx, device)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			result, err := repo.FindByName(ctx, device.Name)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(result.ID).To(gomega.Equal(device.ID))
		})

		ginkgo.When("a device with the same name already exists", func() {
			ginkgo.BeforeEach(func() {
				err := repo.CreateDevice(ctx, device)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should return ErrDeviceDuplicated", func() {
				duplicate := device
				duplicate.ID = domain.ID(utils.GenerateUUID())
				err := repo.CreateDevice(ctx, duplicate)
				gomega.Expect(err).To(gomega.MatchError(usecases.ErrDeviceDuplicated))
			})
		})
	})

	ginkgo.Context("UpdateDevice", func() {
		var device domain.Device

		ginkgo.BeforeEach(func() {
			id := utils.GenerateUUID()
			device = domain.Device{
				ID:   domain.ID(id),
				Name: "device-" + id,
			}
			err := repo.CreateDevice(ctx, device)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.It("should update the device in the database", func() {
			device.DisplayName = "updated-display-name"
			err := repo.UpdateDevice(ctx, device)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			result, err := repo.FindByName(ctx, device.Name)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(result.DisplayName).To(gomega.Equal("updated-display-name"))
		})

		ginkgo.When("the device does not exist", func() {
			ginkgo.It("should return ErrDeviceNotFound", func() {
				nonExistent := domain.Device{
					ID:   domain.ID(utils.GenerateUUID()),
					Name: "non-existent-" + utils.GenerateUUID(),
				}
				err := repo.UpdateDevice(ctx, nonExistent)
				gomega.Expect(err).To(gomega.MatchError(usecases.ErrDeviceNotFound))
			})
		})
	})
})

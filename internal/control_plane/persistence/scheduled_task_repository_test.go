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

var _ = ginkgo.Describe("ScheduledTaskRepository", func() {
	var (
		orm  sql.ORM
		repo usecases.ScheduledTaskRepository
		ctx  context.Context
	)

	ginkgo.BeforeEach(func() {
		var err error
		orm, err = sql.NewMemoryORM("migrations")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		repo, err = persistence.NewScheduledTaskRepository(orm)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(repo).NotTo(gomega.BeNil())

		ctx = context.Background()
	})

	ginkgo.Context("Create", func() {
		var scheduledTask domain.ScheduledTask
		var tenantID domain.ID

		ginkgo.BeforeEach(func() {
			tenantID = domain.ID(utils.GenerateUUID())
			scheduledTask = domain.ScheduledTask{
				ID:       domain.ID(utils.GenerateUUID()),
				Version:  1,
				Tenant:   domain.Tenant{ID: tenantID},
				Device:   domain.Device{ID: domain.ID(utils.GenerateUUID())},
				Schedule: "0 0 * * *",
				IsActive: true,
			}
		})

		ginkgo.It("should create the scheduled task in the database", func() {
			err := repo.Create(ctx, scheduledTask)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			result, err := repo.GetByID(ctx, scheduledTask.ID)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(result.Tenant.ID).To(gomega.Equal(tenantID))
			gomega.Expect(result.Schedule).To(gomega.Equal(scheduledTask.Schedule))
		})
	})

	ginkgo.Context("Update", func() {
		var scheduledTask domain.ScheduledTask

		ginkgo.BeforeEach(func() {
			scheduledTask = domain.ScheduledTask{
				ID:       domain.ID(utils.GenerateUUID()),
				Version:  1,
				Tenant:   domain.Tenant{ID: domain.ID(utils.GenerateUUID())},
				Device:   domain.Device{ID: domain.ID(utils.GenerateUUID())},
				Schedule: "0 0 * * *",
				IsActive: true,
			}
			err := repo.Create(ctx, scheduledTask)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.It("should update the scheduled task in the database", func() {
			scheduledTask.IsActive = false
			err := repo.Update(ctx, scheduledTask)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			result, err := repo.GetByID(ctx, scheduledTask.ID)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(result.IsActive).To(gomega.BeFalse())
			gomega.Expect(result.Version).To(gomega.Equal(domain.Version(2)))
		})
	})

	ginkgo.Context("Delete", func() {
		var scheduledTask domain.ScheduledTask

		ginkgo.BeforeEach(func() {
			scheduledTask = domain.ScheduledTask{
				ID:       domain.ID(utils.GenerateUUID()),
				Version:  1,
				Tenant:   domain.Tenant{ID: domain.ID(utils.GenerateUUID())},
				Device:   domain.Device{ID: domain.ID(utils.GenerateUUID())},
				Schedule: "0 0 * * *",
				IsActive: true,
			}
			err := repo.Create(ctx, scheduledTask)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.It("should soft-delete the scheduled task", func() {
			err := repo.Delete(ctx, scheduledTask.ID)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			_, err = repo.GetByID(ctx, scheduledTask.ID)
			gomega.Expect(err).To(gomega.MatchError(usecases.ErrScheduledTaskNotFound))
		})
	})

	ginkgo.Context("GetByID", func() {
		ginkgo.When("scheduled task does not exist", func() {
			ginkgo.It("should return ErrScheduledTaskNotFound", func() {
				_, err := repo.GetByID(ctx, domain.ID(utils.GenerateUUID()))
				gomega.Expect(err).To(gomega.MatchError(usecases.ErrScheduledTaskNotFound))
			})
		})
	})
})

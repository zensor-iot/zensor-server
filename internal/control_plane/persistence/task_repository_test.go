package persistence_test

import (
	"context"

	"zensor-server/internal/control_plane/persistence"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("TaskRepository", func() {
	var (
		orm  sql.ORM
		repo usecases.TaskRepository
		ctx  context.Context
	)

	ginkgo.BeforeEach(func() {
		var err error
		orm, err = sql.NewMemoryORM("migrations")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		repo, err = persistence.NewTaskRepository(orm)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(repo).NotTo(gomega.BeNil())

		orm.Unscoped().Where("1=1").Delete(&internal.Command{})

		ctx = context.Background()
	})

	ginkgo.Context("Create", func() {
		var task domain.Task
		var device domain.Device

		ginkgo.BeforeEach(func() {
			device = domain.Device{ID: domain.ID(utils.GenerateUUID()), Name: "test-device"}
			taskID := domain.ID(utils.GenerateUUID())
			task = domain.Task{
				ID:      taskID,
				Version: 1,
				Device:  device,
				Commands: []domain.Command{
					{
						ID:      domain.ID(utils.GenerateUUID()),
						Version: 1,
						Device:  device,
						Task:    domain.Task{ID: taskID},
						Port:    domain.Port(15),
					},
					{
						ID:      domain.ID(utils.GenerateUUID()),
						Version: 1,
						Device:  device,
						Task:    domain.Task{ID: taskID},
						Port:    domain.Port(16),
					},
				},
				CreatedAt: utils.Time{},
			}
		})

		ginkgo.It("should create the task and all its commands in a single transaction", func() {
			err := repo.Create(ctx, task)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			tasks, total, err := repo.FindAllByDevice(ctx, device, usecases.Pagination{Limit: 10, Offset: 0})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(total).To(gomega.Equal(1))
			gomega.Expect(tasks).To(gomega.HaveLen(1))
			gomega.Expect(tasks[0].ID).To(gomega.Equal(task.ID))

			var commandCount int64
			err = orm.WithContext(ctx).Model(&internal.Command{}).Where("task_id = ?", task.ID.String()).Count(&commandCount).Error()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(commandCount).To(gomega.Equal(int64(2)))
		})
	})

	ginkgo.Context("FindAllByDevice", func() {
		ginkgo.When("no tasks exist for the device", func() {
			ginkgo.It("should return an empty list", func() {
				tasks, total, err := repo.FindAllByDevice(ctx, domain.Device{ID: domain.ID(utils.GenerateUUID())}, usecases.Pagination{Limit: 10, Offset: 0})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(tasks).To(gomega.BeEmpty())
				gomega.Expect(total).To(gomega.Equal(0))
			})
		})
	})
})

package persistence_test

import (
	"context"
	"time"
	"zensor-server/internal/control_plane/persistence"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("CommandRepository", func() {
	var (
		orm  sql.ORM
		repo usecases.CommandRepository
		ctx  context.Context
	)

	ginkgo.BeforeEach(func() {
		var err error
		orm, err = sql.NewMemoryORM("migrations")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		repo, err = persistence.NewCommandRepository(orm)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(repo).NotTo(gomega.BeNil())

		orm.Unscoped().Where("1=1").Delete(&internal.Command{})

		ctx = context.Background()
	})

	ginkgo.Context("NewCommandRepository", func() {
		ginkgo.When("creating a new command repository", func() {
			ginkgo.It("should create a valid repository instance", func() {
				gomega.Expect(repo).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Context("FindAllPending", func() {
		ginkgo.When("finding all pending commands", func() {
			ginkgo.It("should return empty list when no commands exist", func() {
				commands, err := repo.FindAllPending(ctx)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(commands).To(gomega.BeEmpty())
			})
		})
	})

	ginkgo.Context("FindPendingByDevice", func() {
		var deviceID domain.ID

		ginkgo.When("finding pending commands for a specific device", func() {
			ginkgo.BeforeEach(func() {
				deviceID = domain.ID("test-device-id")
			})

			ginkgo.It("should return empty list when no commands exist for device", func() {
				commands, err := repo.FindPendingByDevice(ctx, deviceID)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(commands).To(gomega.BeEmpty())
			})
		})
	})

	ginkgo.Context("FindByTaskID", func() {
		var taskID domain.ID

		ginkgo.When("finding commands by task ID", func() {
			ginkgo.BeforeEach(func() {
				taskID = domain.ID("test-task-id")
			})

			ginkgo.It("should return empty list when no commands exist for task", func() {
				commands, err := repo.FindByTaskID(ctx, taskID)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(commands).To(gomega.BeEmpty())
			})
		})
	})

	ginkgo.Context("Create", func() {
		var cmd domain.Command
		ginkgo.When("creating a new command", func() {
			ginkgo.BeforeEach(func() {
				id := utils.GenerateUUID()
				cmd = domain.Command{
					ID:       domain.ID(id),
					Version:  domain.Version(1),
					Device:   domain.Device{ID: domain.ID("test-device-id"), Name: "test-device"},
					Task:     domain.Task{ID: domain.ID("test-task-id-create")},
					Port:     domain.Port(15),
					Priority: domain.CommandPriority("NORMAL"),
					Payload: domain.CommandPayload{
						Index: domain.Index(0),
						Value: domain.CommandValue(100),
					},
					DispatchAfter: utils.Time{Time: time.Now()},
					Ready:         false,
					Sent:          false,
					CreatedAt:     utils.Time{Time: time.Now()},
				}
			})

			ginkgo.It("should create the command in the database", func() {
				err := repo.Create(ctx, cmd)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				result, err := repo.GetByID(ctx, cmd.ID)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result.Device.ID).To(gomega.Equal(cmd.Device.ID))
				gomega.Expect(result.Task.ID).To(gomega.Equal(cmd.Task.ID))
			})
		})
	})

	ginkgo.Context("Update", func() {
		var cmd domain.Command
		ginkgo.When("updating an existing command", func() {
			ginkgo.BeforeEach(func() {
				id := utils.GenerateUUID()
				cmd = domain.Command{
					ID:       domain.ID(id),
					Version:  domain.Version(1),
					Device:   domain.Device{ID: domain.ID("test-device-id"), Name: "test-device"},
					Task:     domain.Task{ID: domain.ID("test-task-id-update")},
					Port:     domain.Port(15),
					Priority: domain.CommandPriority("NORMAL"),
					Payload: domain.CommandPayload{
						Index: domain.Index(0),
						Value: domain.CommandValue(100),
					},
					DispatchAfter: utils.Time{Time: time.Now()},
					Ready:         false,
					Sent:          false,
					CreatedAt:     utils.Time{Time: time.Now()},
				}

				internalCmd := internal.FromCommand(cmd)
				err := orm.WithContext(ctx).Create(&internalCmd).Error()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should update the command in the database", func() {
				cmd.Ready = true
				cmd.Sent = true
				cmd.SentAt = utils.Time{Time: time.Now()}

				err := repo.Update(ctx, cmd)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				result, err := repo.GetByID(ctx, cmd.ID)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result.Ready).To(gomega.BeTrue())
				gomega.Expect(result.Sent).To(gomega.BeTrue())
			})
		})

		ginkgo.When("updating a non-existent command", func() {
			ginkgo.BeforeEach(func() {
				cmd = domain.Command{
					ID:       domain.ID("non-existent-command-id"),
					Version:  domain.Version(1),
					Device:   domain.Device{ID: domain.ID("test-device-id"), Name: "test-device"},
					Task:     domain.Task{ID: domain.ID("test-task-id")},
					Port:     domain.Port(15),
					Priority: domain.CommandPriority("NORMAL"),
					Payload: domain.CommandPayload{
						Index: domain.Index(0),
						Value: domain.CommandValue(100),
					},
					DispatchAfter: utils.Time{Time: time.Now()},
					Ready:         true,
					Sent:          true,
					CreatedAt:     utils.Time{Time: time.Now()},
				}
			})

			ginkgo.It("should return error for non-existent command", func() {
				err := repo.Update(ctx, cmd)
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("command not found"))
			})
		})
	})
})

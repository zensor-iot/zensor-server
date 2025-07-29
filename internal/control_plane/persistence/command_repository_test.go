package persistence_test

import (
	"context"
	"time"
	"zensor-server/internal/control_plane/persistence"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// TestCommand represents the device_commands_final table for testing
type TestCommand struct {
	ID            string `gorm:"primaryKey"`
	Version       int
	DeviceName    string
	DeviceID      string
	TaskID        string
	Payload       string `gorm:"type:json"`
	DispatchAfter string
	Port          uint8
	Priority      string
	CreatedAt     string
	Ready         bool
	Sent          bool
	SentAt        string
}

func (t TestCommand) TableName() string {
	return "device_commands"
}

// setupTestTables creates tables for testing that mimic the Materialized View structure
func setupTestTables(orm sql.ORM) error {
	err := orm.AutoMigrate(&TestCommand{})
	return err
}

var _ = ginkgo.Describe("CommandRepository", func() {
	var (
		orm         sql.ORM
		mockFactory pubsub.PublisherFactory
		repo        usecases.CommandRepository
		ctx         context.Context
	)

	ginkgo.BeforeEach(func() {
		var err error
		// Use a unique database name for each test to ensure isolation
		orm, err = sql.NewMemoryORM("migrations")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// Clear any existing data to ensure test isolation
		orm.Unscoped().Where("1=1").Delete(&TestCommand{})

		// Create test tables
		err = setupTestTables(orm)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// Create a mock publisher factory for testing
		mockFactory = pubsub.NewMemoryPublisherFactory()

		repo, err = persistence.NewCommandRepository(orm, mockFactory)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(repo).NotTo(gomega.BeNil())

		ctx = context.Background()
	})

	ginkgo.Context("NewCommandRepository", func() {
		ginkgo.When("creating a new command repository", func() {
			ginkgo.It("should create a valid repository instance", func() {
				// This is already tested in BeforeEach, but keeping for clarity
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
				cmd = domain.Command{
					ID:       domain.ID("test-command-id"),
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

			ginkgo.It("should create command successfully", func() {
				err := repo.Create(ctx, cmd)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Note: In the event-driven architecture, the database update happens through
				// the replication layer consuming Kafka events, not directly in the repository
			})
		})
	})

	ginkgo.Context("Update", func() {
		var cmd domain.Command
		ginkgo.When("updating an existing command", func() {
			ginkgo.BeforeEach(func() {
				cmd = domain.Command{
					ID:       domain.ID("test-command-id-update"),
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

				// First create the command in the database (simulating replication layer)
				internalCmd := internal.FromCommand(cmd)
				err := orm.WithContext(ctx).Create(&internalCmd).Error()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should update command successfully", func() {
				// Now update it through the repository (which publishes to Kafka)
				cmd.Ready = true
				cmd.Sent = true
				cmd.SentAt = utils.Time{Time: time.Now()}

				err := repo.Update(ctx, cmd)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Note: In the event-driven architecture, the database update happens through
				// the replication layer consuming Kafka events, not directly in the repository
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
				// Try to update a non-existent command
				err := repo.Update(ctx, cmd)
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("command not found"))
			})
		})
	})
})

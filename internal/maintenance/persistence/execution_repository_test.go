package persistence_test

import (
	"context"
	"time"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/infra/utils"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	maintenancePersistence "zensor-server/internal/maintenance/persistence"
	maintenancePersistenceInternal "zensor-server/internal/maintenance/persistence/internal"
	maintenanceUsecases "zensor-server/internal/maintenance/usecases"
	shareddomain "zensor-server/internal/shared_kernel/domain"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("MaintenanceExecutionRepository", func() {
	var (
		orm  sql.ORM
		repo maintenanceUsecases.ExecutionRepository
		ctx  context.Context
	)

	ginkgo.BeforeEach(func() {
		var err error
		orm, err = sql.NewMemoryORM("migrations")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// Migrate both tables for testing since execution queries join with activities
		err = orm.AutoMigrate(&maintenancePersistenceInternal.Activity{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		repo, err = maintenancePersistence.NewExecutionRepository(orm)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(repo).NotTo(gomega.BeNil())

		ctx = context.Background()
	})

	ginkgo.Context("Create", func() {
		var execution maintenanceDomain.Execution
		var activityID shareddomain.ID

		ginkgo.BeforeEach(func() {
			activityID = shareddomain.ID(utils.GenerateUUID())

			execution, _ = maintenanceDomain.NewExecutionBuilder().
				WithActivityID(activityID).
				WithScheduledDate(time.Now().AddDate(0, 0, 30)).
				WithFieldValues(map[string]any{"maintenance_type": "Filter Replacement"}).
				Build()
		})

		ginkgo.When("creating a valid execution", func() {
			ginkgo.It("should successfully create the execution", func() {
				err := repo.Create(ctx, execution)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})
	})

	ginkgo.Context("GetByID", func() {
		var execution maintenanceDomain.Execution

		ginkgo.When("execution exists in database", func() {
			ginkgo.BeforeEach(func() {
				activityID := shareddomain.ID(utils.GenerateUUID())

				execution, _ = maintenanceDomain.NewExecutionBuilder().
					WithActivityID(activityID).
					WithScheduledDate(time.Now().AddDate(0, 0, 30)).
					WithFieldValues(map[string]any{
						"maintenance_type": "Filter Replacement",
						"provider":         "ACME Plumbing",
						"cost":             float64(150.50),
					}).
					Build()

				// Insert directly into database for testing
				internalExecution := maintenancePersistenceInternal.FromExecution(execution)
				err := orm.WithContext(ctx).Create(&internalExecution).Error()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should return the execution", func() {
				result, err := repo.GetByID(ctx, execution.ID)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result.ID).To(gomega.Equal(execution.ID))
				gomega.Expect(result.ActivityID).To(gomega.Equal(execution.ActivityID))
				gomega.Expect(result.FieldValues).To(gomega.Equal(execution.FieldValues))
			})
		})

		ginkgo.When("execution does not exist", func() {
			ginkgo.It("should return ErrMaintenanceExecutionNotFound", func() {
				nonExistentID := shareddomain.ID(utils.GenerateUUID())
				result, err := repo.GetByID(ctx, nonExistentID)
				gomega.Expect(err).To(gomega.MatchError(maintenanceUsecases.ErrExecutionNotFound))
				gomega.Expect(result.ID).To(gomega.BeEmpty())
			})
		})
	})

	ginkgo.Context("FindAllByActivity", func() {
		var activityID shareddomain.ID
		var executions []maintenanceDomain.Execution

		ginkgo.BeforeEach(func() {
			activityID = shareddomain.ID(utils.GenerateUUID())
			executions = []maintenanceDomain.Execution{}

			for i := 0; i < 3; i++ {
				execution, _ := maintenanceDomain.NewExecutionBuilder().
					WithActivityID(activityID).
					WithScheduledDate(time.Now().AddDate(0, 0, 30+i*30)).
					WithFieldValues(map[string]any{"maintenance_type": "Filter Replacement"}).
					Build()
				executions = append(executions, execution)
				_ = repo.Create(ctx, execution)
			}
		})

		ginkgo.When("listing executions by activity", func() {
			ginkgo.It("should return empty list since data is not in database", func() {
				emptyActivityID := shareddomain.ID(utils.GenerateUUID())
				pagination := maintenanceUsecases.Pagination{Limit: 10, Offset: 0}
				result, total, err := repo.FindAllByActivity(ctx, emptyActivityID, pagination)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).To(gomega.BeEmpty())
				gomega.Expect(total).To(gomega.Equal(0))
			})
		})

		ginkgo.When("listing executions for different activity", func() {
			ginkgo.It("should return empty list", func() {
				otherActivityID := shareddomain.ID(utils.GenerateUUID())
				pagination := maintenanceUsecases.Pagination{Limit: 10, Offset: 0}
				result, total, err := repo.FindAllByActivity(ctx, otherActivityID, pagination)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).To(gomega.BeEmpty())
				gomega.Expect(total).To(gomega.Equal(0))
			})
		})
	})

	ginkgo.Context("Update", func() {
		var execution maintenanceDomain.Execution

		ginkgo.BeforeEach(func() {
			activityID := shareddomain.ID(utils.GenerateUUID())

			execution, _ = maintenanceDomain.NewExecutionBuilder().
				WithActivityID(activityID).
				WithScheduledDate(time.Now().AddDate(0, 0, 30)).
				WithFieldValues(map[string]any{"maintenance_type": "Filter Replacement"}).
				Build()

			_ = repo.Create(ctx, execution)
		})

		ginkgo.When("updating an execution", func() {
			ginkgo.It("should successfully update the execution", func() {
				err := repo.Update(ctx, execution)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})
	})

	ginkgo.Context("MarkCompleted", func() {
		var execution maintenanceDomain.Execution
		var completedBy string

		ginkgo.When("execution exists in database", func() {
			ginkgo.BeforeEach(func() {
				activityID := shareddomain.ID(utils.GenerateUUID())
				completedBy = "user@example.com"

				execution, _ = maintenanceDomain.NewExecutionBuilder().
					WithActivityID(activityID).
					WithScheduledDate(time.Now().AddDate(0, 0, -5)).
					WithFieldValues(map[string]any{"maintenance_type": "Filter Replacement"}).
					Build()

				// Insert directly into database for testing
				internalExecution := maintenancePersistenceInternal.FromExecution(execution)
				err := orm.WithContext(ctx).Create(&internalExecution).Error()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should successfully mark the execution completed", func() {
				err := repo.MarkCompleted(ctx, execution.ID, completedBy)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})

		ginkgo.When("execution does not exist", func() {
			ginkgo.It("should return ErrMaintenanceExecutionNotFound", func() {
				nonExistentID := shareddomain.ID(utils.GenerateUUID())
				err := repo.MarkCompleted(ctx, nonExistentID, "user@example.com")
				gomega.Expect(err).To(gomega.MatchError(maintenanceUsecases.ErrExecutionNotFound))
			})
		})
	})

	ginkgo.Context("FindOverdueExecutions", func() {
		ginkgo.When("there is an overdue execution for an active activity", func() {
			var tenantID shareddomain.ID
			var execution maintenanceDomain.Execution

			ginkgo.BeforeEach(func() {
				tenantID = shareddomain.ID(utils.GenerateUUID())
				activityID := utils.GenerateUUID()

				internalActivity := maintenancePersistenceInternal.Activity{
					ID:       activityID,
					Version:  1,
					TenantID: tenantID.String(),
					TypeName: "water_system",
					Name:     "Test Activity",
					Schedule: "0 0 1 * *",
					IsActive: true,
				}
				err := orm.WithContext(ctx).Create(&internalActivity).Error()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				execution, _ = maintenanceDomain.NewExecutionBuilder().
					WithActivityID(shareddomain.ID(activityID)).
					WithScheduledDate(time.Now().AddDate(0, 0, -5)).
					Build()

				internalExecution := maintenancePersistenceInternal.FromExecution(execution)
				err = orm.WithContext(ctx).Create(&internalExecution).Error()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should return the activity with the correct tenant ID", func() {
				results, err := repo.FindOverdueExecutions(ctx)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				var found *maintenanceUsecases.ExecutionWithActivity
				for i := range results {
					if results[i].Execution.ID == execution.ID {
						found = &results[i]
						break
					}
				}
				gomega.Expect(found).NotTo(gomega.BeNil(), "overdue execution should be returned")
				gomega.Expect(found.Activity.TenantID).To(gomega.Equal(tenantID))
			})
		})
	})

	ginkgo.Context("FindAllOverdue", func() {
		ginkgo.When("finding overdue executions", func() {
			ginkgo.It("should return empty list", func() {
				tenantID := shareddomain.ID(utils.GenerateUUID())
				result, err := repo.FindAllOverdue(ctx, tenantID)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).To(gomega.BeEmpty())
			})
		})
	})

	ginkgo.Context("FindAllDueSoon", func() {
		ginkgo.When("finding due soon executions", func() {
			ginkgo.It("should return empty list", func() {
				tenantID := shareddomain.ID(utils.GenerateUUID())
				result, err := repo.FindAllDueSoon(ctx, tenantID, 7)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).To(gomega.BeEmpty())
			})
		})
	})
})

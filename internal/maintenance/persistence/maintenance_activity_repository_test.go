package persistence_test

import (
	"context"
	"zensor-server/internal/infra/pubsub"
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

var _ = ginkgo.Describe("MaintenanceActivityRepository", func() {
	var (
		orm             sql.ORM
		mockFactory     pubsub.PublisherFactory
		repo            maintenanceUsecases.MaintenanceActivityRepository
		ctx             context.Context
	)

	ginkgo.BeforeEach(func() {
		var err error
		orm, err = sql.NewMemoryORM("migrations")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		mockFactory = pubsub.NewMemoryPublisherFactory()

		repo, err = maintenancePersistence.NewMaintenanceActivityRepository(mockFactory, orm)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(repo).NotTo(gomega.BeNil())

		ctx = context.Background()
	})

	ginkgo.Context("Create", func() {
		var activity maintenanceDomain.MaintenanceActivity

		ginkgo.BeforeEach(func() {
			activityType := maintenanceDomain.ActivityType{
				Name:         shareddomain.Name(maintenanceDomain.ActivityTypeWaterSystem),
				DisplayName:  shareddomain.DisplayName("Water System"),
				Description:  shareddomain.Description("Water system maintenance"),
				IsPredefined: true,
				Fields: []maintenanceDomain.FieldDefinition{
					{Name: shareddomain.Name("maintenance_type"), DisplayName: shareddomain.DisplayName("Maintenance Type"), Type: maintenanceDomain.FieldTypeText, IsRequired: true},
				},
			}

			activity, _ = maintenanceDomain.NewMaintenanceActivityBuilder().
				WithTenantID(shareddomain.ID(utils.GenerateUUID())).
				WithType(activityType).
				WithName("Monthly Water Filter Check").
				WithDescription("Replace or clean water filter").
				WithSchedule("0 0 1 * *").
				WithNotificationDaysBefore([]int{7, 1}).
				WithFields([]maintenanceDomain.FieldDefinition{}).
				Build()
		})

		ginkgo.When("creating a valid activity", func() {
			ginkgo.It("should successfully publish to kafka", func() {
				err := repo.Create(ctx, activity)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})
	})

	ginkgo.Context("GetByID", func() {
		var activity maintenanceDomain.MaintenanceActivity

		ginkgo.When("activity exists in database", func() {
			ginkgo.BeforeEach(func() {
				activityType := maintenanceDomain.ActivityType{
					Name:         shareddomain.Name(maintenanceDomain.ActivityTypeWaterSystem),
					DisplayName:  shareddomain.DisplayName("Water System"),
					Description:  shareddomain.Description("Water system maintenance"),
					IsPredefined: true,
					Fields: []maintenanceDomain.FieldDefinition{
						{Name: shareddomain.Name("maintenance_type"), DisplayName: shareddomain.DisplayName("Maintenance Type"), Type: maintenanceDomain.FieldTypeText, IsRequired: true},
						{Name: shareddomain.Name("provider"), DisplayName: shareddomain.DisplayName("Provider"), Type: maintenanceDomain.FieldTypeText, IsRequired: false},
					},
				}

				activity, _ = maintenanceDomain.NewMaintenanceActivityBuilder().
					WithTenantID(shareddomain.ID(utils.GenerateUUID())).
					WithType(activityType).
					WithName("Test Activity").
					WithDescription("Test Description").
					WithSchedule("0 0 1 * *").
					WithNotificationDaysBefore([]int{7, 3}).
					WithFields([]maintenanceDomain.FieldDefinition{}).
					Build()

				// Insert directly into database for testing
				internalActivity := maintenancePersistenceInternal.FromMaintenanceActivity(activity)
				err := orm.WithContext(ctx).Create(&internalActivity).Error()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should return the activity", func() {
				result, err := repo.GetByID(ctx, activity.ID)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result.ID).To(gomega.Equal(activity.ID))
				gomega.Expect(result.Name).To(gomega.Equal(activity.Name))
				gomega.Expect(result.Description).To(gomega.Equal(activity.Description))
				gomega.Expect(result.Schedule).To(gomega.Equal(activity.Schedule))
			})
		})

		ginkgo.When("activity does not exist", func() {
			ginkgo.It("should return ErrMaintenanceActivityNotFound", func() {
				nonExistentID := shareddomain.ID(utils.GenerateUUID())
				result, err := repo.GetByID(ctx, nonExistentID)
				gomega.Expect(err).To(gomega.MatchError(maintenanceUsecases.ErrMaintenanceActivityNotFound))
				gomega.Expect(result.ID).To(gomega.BeEmpty())
			})
		})
	})

	ginkgo.Context("FindAllByTenant", func() {
		var tenantID shareddomain.ID
		var activities []maintenanceDomain.MaintenanceActivity

		ginkgo.BeforeEach(func() {
			tenantID = shareddomain.ID(utils.GenerateUUID())
			activities = []maintenanceDomain.MaintenanceActivity{}

			activityType := maintenanceDomain.ActivityType{
				Name:         shareddomain.Name(maintenanceDomain.ActivityTypeWaterSystem),
				DisplayName:  shareddomain.DisplayName("Water System"),
				Description:  shareddomain.Description("Water system maintenance"),
				IsPredefined: true,
				Fields: []maintenanceDomain.FieldDefinition{},
			}

			for i := 0; i < 3; i++ {
				activity, _ := maintenanceDomain.NewMaintenanceActivityBuilder().
					WithTenantID(tenantID).
					WithType(activityType).
					WithName("Test Activity " + string(rune('0'+i))).
					WithDescription("Test Description").
					WithSchedule("0 0 1 * *").
					WithFields([]maintenanceDomain.FieldDefinition{}).
					Build()
				activities = append(activities, activity)
				_ = repo.Create(ctx, activity)
			}
		})

		ginkgo.When("listing activities by tenant", func() {
			ginkgo.It("should return empty list since data is not in database", func() {
				pagination := maintenanceUsecases.Pagination{Limit: 10, Offset: 0}
				result, total, err := repo.FindAllByTenant(ctx, tenantID, pagination)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).To(gomega.BeEmpty())
				gomega.Expect(total).To(gomega.Equal(0))
			})
		})

		ginkgo.When("listing activities for different tenant", func() {
			ginkgo.It("should return empty list", func() {
				otherTenantID := shareddomain.ID(utils.GenerateUUID())
				pagination := maintenanceUsecases.Pagination{Limit: 10, Offset: 0}
				result, total, err := repo.FindAllByTenant(ctx, otherTenantID, pagination)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).To(gomega.BeEmpty())
				gomega.Expect(total).To(gomega.Equal(0))
			})
		})
	})

	ginkgo.Context("Update", func() {
		var activity maintenanceDomain.MaintenanceActivity

		ginkgo.BeforeEach(func() {
			activityType := maintenanceDomain.ActivityType{
				Name:         shareddomain.Name(maintenanceDomain.ActivityTypeWaterSystem),
				DisplayName:  shareddomain.DisplayName("Water System"),
				Description:  shareddomain.Description("Water system maintenance"),
				IsPredefined: true,
				Fields: []maintenanceDomain.FieldDefinition{},
			}

			activity, _ = maintenanceDomain.NewMaintenanceActivityBuilder().
				WithTenantID(shareddomain.ID(utils.GenerateUUID())).
				WithType(activityType).
				WithName("Original Name").
				WithDescription("Original Description").
				WithSchedule("0 0 1 * *").
				WithFields([]maintenanceDomain.FieldDefinition{}).
				Build()

			_ = repo.Create(ctx, activity)
		})

		ginkgo.When("updating an activity", func() {
			ginkgo.It("should successfully publish update to kafka", func() {
				activity.Description = shareddomain.Description("Updated Description")
				err := repo.Update(ctx, activity)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})
	})

	ginkgo.Context("Delete", func() {
		var activity maintenanceDomain.MaintenanceActivity

		ginkgo.When("activity exists in database", func() {
			ginkgo.BeforeEach(func() {
				activityType := maintenanceDomain.ActivityType{
					Name:         shareddomain.Name(maintenanceDomain.ActivityTypeWaterSystem),
					DisplayName:  shareddomain.DisplayName("Water System"),
					Description:  shareddomain.Description("Water system maintenance"),
					IsPredefined: true,
					Fields: []maintenanceDomain.FieldDefinition{},
				}

				activity, _ = maintenanceDomain.NewMaintenanceActivityBuilder().
					WithTenantID(shareddomain.ID(utils.GenerateUUID())).
					WithType(activityType).
					WithName("Test Activity to Delete").
					WithDescription("Test Description").
					WithSchedule("0 0 1 * *").
					WithFields([]maintenanceDomain.FieldDefinition{}).
					Build()

				// Insert directly into database for testing
				internalActivity := maintenancePersistenceInternal.FromMaintenanceActivity(activity)
				err := orm.WithContext(ctx).Create(&internalActivity).Error()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should successfully publish soft delete to kafka", func() {
				err := repo.Delete(ctx, activity.ID)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Note: In production, Kafka Connect would sync this to the database
				// For now, we just verify the publish succeeded
				// The activity will still be visible in DB until sync
			})
		})

		ginkgo.When("activity does not exist", func() {
			ginkgo.It("should return ErrMaintenanceActivityNotFound", func() {
				nonExistentID := shareddomain.ID(utils.GenerateUUID())
				err := repo.Delete(ctx, nonExistentID)
				gomega.Expect(err).To(gomega.MatchError(maintenanceUsecases.ErrMaintenanceActivityNotFound))
			})
		})
	})
})


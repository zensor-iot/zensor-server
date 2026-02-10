package usecases_test

import (
	"context"
	"errors"
	controlPlaneUsecases "zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/utils"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	maintenanceUsecases "zensor-server/internal/maintenance/usecases"
	shareddomain "zensor-server/internal/shared_kernel/domain"
	mocksharedkernel "zensor-server/test/unit/doubles/shared_kernel/usecases"
	mockmaintenance "zensor-server/test/unit/doubles/maintenance/usecases"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("MaintenanceActivityService", func() {
	var (
		ctrl              *gomock.Controller
		mockRepository    *mockmaintenance.MockActivityRepository
		mockTenantService *mocksharedkernel.MockTenantService
		service           maintenanceUsecases.ActivityService
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockRepository = mockmaintenance.NewMockActivityRepository(ctrl)
		mockTenantService = mocksharedkernel.NewMockTenantService(ctrl)
		service = maintenanceUsecases.NewActivityService(mockRepository, mockTenantService)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("CreateActivity", func() {
		var activity maintenanceDomain.Activity

		BeforeEach(func() {
			activityType := maintenanceDomain.ActivityType{
				Name:         shareddomain.Name(maintenanceDomain.ActivityTypeWaterSystem),
				DisplayName:  shareddomain.DisplayName("Water System"),
				Description:  shareddomain.Description("Water system maintenance"),
				IsPredefined: true,
				Fields: []maintenanceDomain.FieldDefinition{
					{Name: shareddomain.Name("maintenance_type"), DisplayName: shareddomain.DisplayName("Maintenance Type"), Type: maintenanceDomain.FieldTypeText, IsRequired: true},
				},
			}

			activity, _ = maintenanceDomain.NewActivityBuilder().
				WithTenantID(shareddomain.ID(utils.GenerateUUID())).
				WithType(activityType).
				WithName("Monthly Water Filter Check").
				WithDescription("Replace or clean water filter").
				WithSchedule("0 0 1 * *").
				WithNotificationDaysBefore([]int{7, 1}).
				WithFields([]maintenanceDomain.FieldDefinition{}).
				Build()
		})

		When("creating a valid activity", func() {
			It("should successfully create the activity", func() {
				tenant := shareddomain.Tenant{
					ID:   activity.TenantID,
					Name: "Test Tenant",
				}
				mockTenantService.EXPECT().
					GetTenant(gomock.Any(), activity.TenantID).
					Return(tenant, nil)
				mockRepository.EXPECT().
					Create(gomock.Any(), activity).
					Return(nil)

				err := service.CreateActivity(context.Background(), activity)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("tenant does not exist", func() {
			It("should return an error", func() {
				mockTenantService.EXPECT().
					GetTenant(gomock.Any(), activity.TenantID).
					Return(shareddomain.Tenant{}, controlPlaneUsecases.ErrTenantNotFound)

				err := service.CreateActivity(context.Background(), activity)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("tenant not found"))
			})
		})

		When("repository returns an error", func() {
			It("should return the error", func() {
				tenant := shareddomain.Tenant{
					ID:   activity.TenantID,
					Name: "Test Tenant",
				}
				mockTenantService.EXPECT().
					GetTenant(gomock.Any(), activity.TenantID).
					Return(tenant, nil)
				mockRepository.EXPECT().
					Create(gomock.Any(), activity).
					Return(errors.New("database error"))

				err := service.CreateActivity(context.Background(), activity)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("creating maintenance activity"))
			})
		})
	})

	Context("GetActivity", func() {
		var activityID shareddomain.ID
		var activity maintenanceDomain.Activity

		BeforeEach(func() {
			activityID = shareddomain.ID(utils.GenerateUUID())
			activityType := maintenanceDomain.ActivityType{
				Name:         shareddomain.Name(maintenanceDomain.ActivityTypeWaterSystem),
				DisplayName:  shareddomain.DisplayName("Water System"),
				Description:  shareddomain.Description("Water system maintenance"),
				IsPredefined: true,
				Fields:       []maintenanceDomain.FieldDefinition{},
			}

			activity, _ = maintenanceDomain.NewActivityBuilder().
				WithTenantID(shareddomain.ID(utils.GenerateUUID())).
				WithType(activityType).
				WithName("Test Activity").
				WithDescription("Test Description").
				WithSchedule("0 0 1 * *").
				WithFields([]maintenanceDomain.FieldDefinition{}).
				Build()
			activity.ID = activityID
		})

		When("activity exists", func() {
			It("should return the activity", func() {
				mockRepository.EXPECT().
					GetByID(gomock.Any(), activityID).
					Return(activity, nil)

				result, err := service.GetActivity(context.Background(), activityID)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.ID).To(Equal(activityID))
				Expect(result.Name).To(Equal(activity.Name))
			})
		})

		When("activity does not exist", func() {
			It("should return ErrMaintenanceActivityNotFound", func() {
				mockRepository.EXPECT().
					GetByID(gomock.Any(), activityID).
					Return(maintenanceDomain.Activity{}, maintenanceUsecases.ErrActivityNotFound)

				result, err := service.GetActivity(context.Background(), activityID)
				Expect(err).To(MatchError(maintenanceUsecases.ErrActivityNotFound))
				Expect(result.ID).To(BeEmpty())
			})
		})

		When("repository returns an error", func() {
			It("should return the error", func() {
				mockRepository.EXPECT().
					GetByID(gomock.Any(), activityID).
					Return(maintenanceDomain.Activity{}, errors.New("database error"))

				result, err := service.GetActivity(context.Background(), activityID)
				Expect(err).To(HaveOccurred())
				Expect(result.ID).To(BeEmpty())
				Expect(err.Error()).To(ContainSubstring("getting maintenance activity"))
			})
		})
	})

	Context("ListActivitiesByTenant", func() {
		var tenantID shareddomain.ID
		var activities []maintenanceDomain.Activity

		BeforeEach(func() {
			tenantID = shareddomain.ID(utils.GenerateUUID())
			activities = []maintenanceDomain.Activity{}

			activityType := maintenanceDomain.ActivityType{
				Name:         shareddomain.Name(maintenanceDomain.ActivityTypeWaterSystem),
				DisplayName:  shareddomain.DisplayName("Water System"),
				Description:  shareddomain.Description("Water system maintenance"),
				IsPredefined: true,
				Fields:       []maintenanceDomain.FieldDefinition{},
			}

			for i := 0; i < 3; i++ {
				activity, _ := maintenanceDomain.NewActivityBuilder().
					WithTenantID(tenantID).
					WithType(activityType).
					WithName("Test Activity " + string(rune('0'+i))).
					WithDescription("Test Description").
					WithSchedule("0 0 1 * *").
					WithFields([]maintenanceDomain.FieldDefinition{}).
					Build()
				activities = append(activities, activity)
			}
		})

		When("listing activities", func() {
			It("should return all activities for the tenant", func() {
				pagination := maintenanceUsecases.Pagination{Limit: 10, Offset: 0}
				mockRepository.EXPECT().
					FindAllByTenant(gomock.Any(), tenantID, pagination).
					Return(activities, len(activities), nil)

				result, total, err := service.ListActivitiesByTenant(context.Background(), tenantID, pagination)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(HaveLen(3))
				Expect(total).To(Equal(3))
			})
		})

		When("repository returns an error", func() {
			It("should return the error", func() {
				pagination := maintenanceUsecases.Pagination{Limit: 10, Offset: 0}
				mockRepository.EXPECT().
					FindAllByTenant(gomock.Any(), tenantID, pagination).
					Return(nil, 0, errors.New("database error"))

				result, total, err := service.ListActivitiesByTenant(context.Background(), tenantID, pagination)
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
				Expect(total).To(Equal(0))
				Expect(err.Error()).To(ContainSubstring("listing maintenance activities"))
			})
		})
	})

	Context("UpdateActivity", func() {
		var activity maintenanceDomain.Activity

		BeforeEach(func() {
			activityID := shareddomain.ID(utils.GenerateUUID())
			activityType := maintenanceDomain.ActivityType{
				Name:         shareddomain.Name(maintenanceDomain.ActivityTypeWaterSystem),
				DisplayName:  shareddomain.DisplayName("Water System"),
				Description:  shareddomain.Description("Water system maintenance"),
				IsPredefined: true,
				Fields:       []maintenanceDomain.FieldDefinition{},
			}

			activity, _ = maintenanceDomain.NewActivityBuilder().
				WithTenantID(shareddomain.ID(utils.GenerateUUID())).
				WithType(activityType).
				WithName("Original Name").
				WithDescription("Original Description").
				WithSchedule("0 0 1 * *").
				WithFields([]maintenanceDomain.FieldDefinition{}).
				Build()
			activity.ID = activityID
		})

		When("updating an existing activity", func() {
			It("should successfully update the activity", func() {
				activity.Description = shareddomain.Description("Updated Description")
				mockRepository.EXPECT().
					GetByID(gomock.Any(), activity.ID).
					Return(activity, nil)
				mockRepository.EXPECT().
					Update(gomock.Any(), activity).
					Return(nil)

				err := service.UpdateActivity(context.Background(), activity)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("activity does not exist", func() {
			It("should return ErrMaintenanceActivityNotFound", func() {
				mockRepository.EXPECT().
					GetByID(gomock.Any(), activity.ID).
					Return(maintenanceDomain.Activity{}, maintenanceUsecases.ErrActivityNotFound)

				err := service.UpdateActivity(context.Background(), activity)
				Expect(err).To(MatchError(maintenanceUsecases.ErrActivityNotFound))
			})
		})

		When("activity is deleted", func() {
			It("should return an error", func() {
				activity.SoftDelete()
				mockRepository.EXPECT().
					GetByID(gomock.Any(), activity.ID).
					Return(activity, nil)

				err := service.UpdateActivity(context.Background(), activity)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("deleted"))
			})
		})
	})

	Context("DeleteActivity", func() {
		var activity maintenanceDomain.Activity

		BeforeEach(func() {
			activityID := shareddomain.ID(utils.GenerateUUID())
			activityType := maintenanceDomain.ActivityType{
				Name:         shareddomain.Name(maintenanceDomain.ActivityTypeWaterSystem),
				DisplayName:  shareddomain.DisplayName("Water System"),
				Description:  shareddomain.Description("Water system maintenance"),
				IsPredefined: true,
				Fields:       []maintenanceDomain.FieldDefinition{},
			}

			activity, _ = maintenanceDomain.NewActivityBuilder().
				WithTenantID(shareddomain.ID(utils.GenerateUUID())).
				WithType(activityType).
				WithName("Test Activity").
				WithDescription("Test Description").
				WithSchedule("0 0 1 * *").
				WithFields([]maintenanceDomain.FieldDefinition{}).
				Build()
			activity.ID = activityID
		})

		When("deleting an existing activity", func() {
			It("should successfully delete the activity", func() {
				mockRepository.EXPECT().
					GetByID(gomock.Any(), activity.ID).
					Return(activity, nil)
				mockRepository.EXPECT().
					Delete(gomock.Any(), activity.ID).
					Return(nil)

				err := service.DeleteActivity(context.Background(), activity.ID)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("activity does not exist", func() {
			It("should return ErrMaintenanceActivityNotFound", func() {
				mockRepository.EXPECT().
					GetByID(gomock.Any(), activity.ID).
					Return(maintenanceDomain.Activity{}, maintenanceUsecases.ErrActivityNotFound)

				err := service.DeleteActivity(context.Background(), activity.ID)
				Expect(err).To(MatchError(maintenanceUsecases.ErrActivityNotFound))
			})
		})

		When("activity is already deleted", func() {
			It("should return an error", func() {
				activity.SoftDelete()
				mockRepository.EXPECT().
					GetByID(gomock.Any(), activity.ID).
					Return(activity, nil)

				err := service.DeleteActivity(context.Background(), activity.ID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("already deleted"))
			})
		})
	})

	Context("ActivateActivity", func() {
		var activity maintenanceDomain.Activity

		BeforeEach(func() {
			activityID := shareddomain.ID(utils.GenerateUUID())
			activityType := maintenanceDomain.ActivityType{
				Name:         shareddomain.Name(maintenanceDomain.ActivityTypeWaterSystem),
				DisplayName:  shareddomain.DisplayName("Water System"),
				Description:  shareddomain.Description("Water system maintenance"),
				IsPredefined: true,
				Fields:       []maintenanceDomain.FieldDefinition{},
			}

			activity, _ = maintenanceDomain.NewActivityBuilder().
				WithTenantID(shareddomain.ID(utils.GenerateUUID())).
				WithType(activityType).
				WithName("Test Activity").
				WithDescription("Test Description").
				WithSchedule("0 0 1 * *").
				WithFields([]maintenanceDomain.FieldDefinition{}).
				Build()
			activity.ID = activityID
			activity.IsActive = false
		})

		When("activating an existing activity", func() {
			It("should successfully activate the activity", func() {
				mockRepository.EXPECT().
					GetByID(gomock.Any(), activity.ID).
					Return(activity, nil)
				mockRepository.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(nil)

				err := service.ActivateActivity(context.Background(), activity.ID)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("activity does not exist", func() {
			It("should return ErrMaintenanceActivityNotFound", func() {
				mockRepository.EXPECT().
					GetByID(gomock.Any(), activity.ID).
					Return(maintenanceDomain.Activity{}, maintenanceUsecases.ErrActivityNotFound)

				err := service.ActivateActivity(context.Background(), activity.ID)
				Expect(err).To(MatchError(maintenanceUsecases.ErrActivityNotFound))
			})
		})

		When("activity is deleted", func() {
			It("should return an error", func() {
				activity.SoftDelete()
				mockRepository.EXPECT().
					GetByID(gomock.Any(), activity.ID).
					Return(activity, nil)

				err := service.ActivateActivity(context.Background(), activity.ID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("deleted"))
			})
		})
	})

	Context("DeactivateActivity", func() {
		var activity maintenanceDomain.Activity

		BeforeEach(func() {
			activityID := shareddomain.ID(utils.GenerateUUID())
			activityType := maintenanceDomain.ActivityType{
				Name:         shareddomain.Name(maintenanceDomain.ActivityTypeWaterSystem),
				DisplayName:  shareddomain.DisplayName("Water System"),
				Description:  shareddomain.Description("Water system maintenance"),
				IsPredefined: true,
				Fields:       []maintenanceDomain.FieldDefinition{},
			}

			activity, _ = maintenanceDomain.NewActivityBuilder().
				WithTenantID(shareddomain.ID(utils.GenerateUUID())).
				WithType(activityType).
				WithName("Test Activity").
				WithDescription("Test Description").
				WithSchedule("0 0 1 * *").
				WithFields([]maintenanceDomain.FieldDefinition{}).
				Build()
			activity.ID = activityID
			activity.IsActive = true
		})

		When("deactivating an existing activity", func() {
			It("should successfully deactivate the activity", func() {
				mockRepository.EXPECT().
					GetByID(gomock.Any(), activity.ID).
					Return(activity, nil)
				mockRepository.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(nil)

				err := service.DeactivateActivity(context.Background(), activity.ID)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("activity does not exist", func() {
			It("should return ErrMaintenanceActivityNotFound", func() {
				mockRepository.EXPECT().
					GetByID(gomock.Any(), activity.ID).
					Return(maintenanceDomain.Activity{}, maintenanceUsecases.ErrActivityNotFound)

				err := service.DeactivateActivity(context.Background(), activity.ID)
				Expect(err).To(MatchError(maintenanceUsecases.ErrActivityNotFound))
			})
		})

		When("activity is deleted", func() {
			It("should return an error", func() {
				activity.SoftDelete()
				mockRepository.EXPECT().
					GetByID(gomock.Any(), activity.ID).
					Return(activity, nil)

				err := service.DeactivateActivity(context.Background(), activity.ID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("deleted"))
			})
		})
	})
})

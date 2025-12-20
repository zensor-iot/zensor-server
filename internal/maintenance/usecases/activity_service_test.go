package usecases_test

import (
	"context"
	"errors"
	"zensor-server/internal/infra/utils"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	maintenanceUsecases "zensor-server/internal/maintenance/usecases"
	shareddomain "zensor-server/internal/shared_kernel/domain"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MaintenanceActivityService", func() {
	var service maintenanceUsecases.ActivityService
	var mockRepository *mockMaintenanceActivityRepository

	BeforeEach(func() {
		mockRepository = newMockMaintenanceActivityRepository()
		service = maintenanceUsecases.NewActivityService(mockRepository)
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
				err := service.CreateActivity(context.Background(), activity)
				Expect(err).NotTo(HaveOccurred())
				Expect(mockRepository.createCalled).To(BeTrue())
				Expect(mockRepository.activities[activity.ID.String()]).To(Equal(activity))
			})
		})

		When("repository returns an error", func() {
			BeforeEach(func() {
				mockRepository.createError = errors.New("database error")
			})

			It("should return the error", func() {
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

			mockRepository.activities[activityID.String()] = activity
		})

		When("activity exists", func() {
			It("should return the activity", func() {
				result, err := service.GetActivity(context.Background(), activityID)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.ID).To(Equal(activityID))
				Expect(result.Name).To(Equal(activity.Name))
			})
		})

		When("activity does not exist", func() {
			BeforeEach(func() {
				mockRepository.getByIDError = maintenanceUsecases.ErrActivityNotFound
			})

			It("should return ErrMaintenanceActivityNotFound", func() {
				result, err := service.GetActivity(context.Background(), activityID)
				Expect(err).To(MatchError(maintenanceUsecases.ErrActivityNotFound))
				Expect(result.ID).To(BeEmpty())
			})
		})

		When("repository returns an error", func() {
			BeforeEach(func() {
				mockRepository.getByIDError = errors.New("database error")
			})

			It("should return the error", func() {
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

			mockRepository.activitiesByTenant[tenantID.String()] = activities
			mockRepository.totalByTenant[tenantID.String()] = len(activities)
		})

		When("listing activities", func() {
			It("should return all activities for the tenant", func() {
				pagination := maintenanceUsecases.Pagination{Limit: 10, Offset: 0}
				result, total, err := service.ListActivitiesByTenant(context.Background(), tenantID, pagination)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(HaveLen(3))
				Expect(total).To(Equal(3))
			})
		})

		When("repository returns an error", func() {
			BeforeEach(func() {
				mockRepository.findAllByTenantError = errors.New("database error")
			})

			It("should return the error", func() {
				pagination := maintenanceUsecases.Pagination{Limit: 10, Offset: 0}
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

			mockRepository.activities[activityID.String()] = activity
		})

		When("updating an existing activity", func() {
			It("should successfully update the activity", func() {
				activity.Description = shareddomain.Description("Updated Description")
				err := service.UpdateActivity(context.Background(), activity)
				Expect(err).NotTo(HaveOccurred())
				Expect(mockRepository.updateCalled).To(BeTrue())
			})
		})

		When("activity does not exist", func() {
			BeforeEach(func() {
				mockRepository.getByIDError = maintenanceUsecases.ErrActivityNotFound
			})

			It("should return ErrMaintenanceActivityNotFound", func() {
				err := service.UpdateActivity(context.Background(), activity)
				Expect(err).To(MatchError(maintenanceUsecases.ErrActivityNotFound))
			})
		})

		When("activity is deleted", func() {
			BeforeEach(func() {
				activity.SoftDelete()
				mockRepository.activities[activity.ID.String()] = activity
			})

			It("should return an error", func() {
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

			mockRepository.activities[activityID.String()] = activity
		})

		When("deleting an existing activity", func() {
			It("should successfully delete the activity", func() {
				err := service.DeleteActivity(context.Background(), activity.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(mockRepository.deleteCalled).To(BeTrue())
			})
		})

		When("activity does not exist", func() {
			BeforeEach(func() {
				mockRepository.getByIDError = maintenanceUsecases.ErrActivityNotFound
			})

			It("should return ErrMaintenanceActivityNotFound", func() {
				err := service.DeleteActivity(context.Background(), activity.ID)
				Expect(err).To(MatchError(maintenanceUsecases.ErrActivityNotFound))
			})
		})

		When("activity is already deleted", func() {
			BeforeEach(func() {
				activity.SoftDelete()
				mockRepository.activities[activity.ID.String()] = activity
			})

			It("should return an error", func() {
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

			mockRepository.activities[activityID.String()] = activity
		})

		When("activating an existing activity", func() {
			It("should successfully activate the activity", func() {
				err := service.ActivateActivity(context.Background(), activity.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(mockRepository.updateCalled).To(BeTrue())
			})
		})

		When("activity does not exist", func() {
			BeforeEach(func() {
				mockRepository.getByIDError = maintenanceUsecases.ErrActivityNotFound
			})

			It("should return ErrMaintenanceActivityNotFound", func() {
				err := service.ActivateActivity(context.Background(), activity.ID)
				Expect(err).To(MatchError(maintenanceUsecases.ErrActivityNotFound))
			})
		})

		When("activity is deleted", func() {
			BeforeEach(func() {
				activity.SoftDelete()
				mockRepository.activities[activity.ID.String()] = activity
			})

			It("should return an error", func() {
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

			mockRepository.activities[activityID.String()] = activity
		})

		When("deactivating an existing activity", func() {
			It("should successfully deactivate the activity", func() {
				err := service.DeactivateActivity(context.Background(), activity.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(mockRepository.updateCalled).To(BeTrue())
			})
		})

		When("activity does not exist", func() {
			BeforeEach(func() {
				mockRepository.getByIDError = maintenanceUsecases.ErrActivityNotFound
			})

			It("should return ErrMaintenanceActivityNotFound", func() {
				err := service.DeactivateActivity(context.Background(), activity.ID)
				Expect(err).To(MatchError(maintenanceUsecases.ErrActivityNotFound))
			})
		})

		When("activity is deleted", func() {
			BeforeEach(func() {
				activity.SoftDelete()
				mockRepository.activities[activity.ID.String()] = activity
			})

			It("should return an error", func() {
				err := service.DeactivateActivity(context.Background(), activity.ID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("deleted"))
			})
		})
	})
})

type mockMaintenanceActivityRepository struct {
	activities            map[string]maintenanceDomain.Activity
	activitiesByTenant    map[string][]maintenanceDomain.Activity
	totalByTenant         map[string]int
	createCalled          bool
	getByIDCalled         bool
	findAllByTenantCalled bool
	updateCalled          bool
	deleteCalled          bool
	createError           error
	getByIDError          error
	findAllByTenantError  error
	updateError           error
	deleteError           error
}

func newMockMaintenanceActivityRepository() *mockMaintenanceActivityRepository {
	return &mockMaintenanceActivityRepository{
		activities:         make(map[string]maintenanceDomain.Activity),
		activitiesByTenant: make(map[string][]maintenanceDomain.Activity),
		totalByTenant:      make(map[string]int),
	}
}

func (m *mockMaintenanceActivityRepository) Create(ctx context.Context, activity maintenanceDomain.Activity) error {
	m.createCalled = true
	if m.createError != nil {
		return m.createError
	}
	m.activities[activity.ID.String()] = activity
	return nil
}

func (m *mockMaintenanceActivityRepository) GetByID(ctx context.Context, id shareddomain.ID) (maintenanceDomain.Activity, error) {
	m.getByIDCalled = true
	if m.getByIDError != nil {
		return maintenanceDomain.Activity{}, m.getByIDError
	}
	if activity, ok := m.activities[id.String()]; ok {
		return activity, nil
	}
	return maintenanceDomain.Activity{}, maintenanceUsecases.ErrActivityNotFound
}

func (m *mockMaintenanceActivityRepository) FindAllByTenant(ctx context.Context, tenantID shareddomain.ID, pagination maintenanceUsecases.Pagination) ([]maintenanceDomain.Activity, int, error) {
	m.findAllByTenantCalled = true
	if m.findAllByTenantError != nil {
		return nil, 0, m.findAllByTenantError
	}
	if activities, ok := m.activitiesByTenant[tenantID.String()]; ok {
		total := m.totalByTenant[tenantID.String()]
		return activities, total, nil
	}
	return []maintenanceDomain.Activity{}, 0, nil
}

func (m *mockMaintenanceActivityRepository) Update(ctx context.Context, activity maintenanceDomain.Activity) error {
	m.updateCalled = true
	if m.updateError != nil {
		return m.updateError
	}
	m.activities[activity.ID.String()] = activity
	return nil
}

func (m *mockMaintenanceActivityRepository) Delete(ctx context.Context, id shareddomain.ID) error {
	m.deleteCalled = true
	if m.deleteError != nil {
		return m.deleteError
	}
	delete(m.activities, id.String())
	return nil
}

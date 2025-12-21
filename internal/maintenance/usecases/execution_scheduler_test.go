package usecases_test

import (
	"context"
	"errors"
	"time"
	controlPlaneUsecases "zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/utils"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	maintenanceUsecases "zensor-server/internal/maintenance/usecases"
	shareddomain "zensor-server/internal/shared_kernel/domain"
	mockcontrolplane "zensor-server/test/unit/doubles/control_plane/usecases"
	mockasync "zensor-server/test/unit/doubles/infra/async"
	mockmaintenance "zensor-server/test/unit/doubles/maintenance/usecases"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("ExecutionScheduler", func() {
	var (
		ctrl                           *gomock.Controller
		ticker                         *time.Ticker
		mockActivityRepository         *mockActivityRepository
		mockExecutionRepository        *mockExecutionRepository
		mockExecutionService           *mockmaintenance.MockExecutionService
		mockTenantService              *mockcontrolplane.MockTenantService
		mockTenantConfigurationService *mockcontrolplane.MockTenantConfigurationService
		mockBroker                     *mockasync.MockInternalBroker
		scheduler                      *maintenanceUsecases.ExecutionScheduler
		ctx                            context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ticker = time.NewTicker(100 * time.Millisecond)
		mockActivityRepository = newMockActivityRepository()
		mockExecutionRepository = newMockExecutionRepository()
		mockExecutionService = mockmaintenance.NewMockExecutionService(ctrl)
		mockTenantService = mockcontrolplane.NewMockTenantService(ctrl)
		mockTenantConfigurationService = mockcontrolplane.NewMockTenantConfigurationService(ctrl)
		mockBroker = mockasync.NewMockInternalBroker(ctrl)
		ctx = context.Background()

		scheduler = maintenanceUsecases.NewExecutionScheduler(
			ticker,
			mockActivityRepository,
			mockExecutionRepository,
			mockExecutionService,
			mockTenantService,
			mockTenantConfigurationService,
			mockBroker,
		)
	})

	AfterEach(func() {
		ticker.Stop()
		ctrl.Finish()
	})

	Context("scheduleExecutions", func() {
		var activity maintenanceDomain.Activity
		var tenant shareddomain.Tenant
		var tenantConfig shareddomain.TenantConfiguration

		BeforeEach(func() {
			tenantID := shareddomain.ID(utils.GenerateUUID())
			activityID := shareddomain.ID(utils.GenerateUUID())

			activityType := maintenanceDomain.ActivityType{
				Name:         shareddomain.Name(maintenanceDomain.ActivityTypeWaterSystem),
				DisplayName:  shareddomain.DisplayName("Water System"),
				Description:  shareddomain.Description("Water system maintenance"),
				IsPredefined: true,
				Fields: []maintenanceDomain.FieldDefinition{
					{
						Name:         shareddomain.Name("maintenance_type"),
						DisplayName:  shareddomain.DisplayName("Maintenance Type"),
						Type:         maintenanceDomain.FieldTypeText,
						IsRequired:   true,
						DefaultValue: func() *any { v := any("Filter Replacement"); return &v }(),
					},
				},
			}

			activity, _ = maintenanceDomain.NewActivityBuilder().
				WithTenantID(tenantID).
				WithType(activityType).
				WithName("Test Activity").
				WithDescription("Test Description").
				WithSchedule("0 0 * * *").
				WithFields(activityType.Fields).
				Build()
			activity.ID = activityID
			activity.Activate()

			tenant, _ = shareddomain.NewTenantBuilder().
				WithName("Test Tenant").
				Build()
			tenant.ID = tenantID

			tenantConfig, _ = shareddomain.NewTenantConfigurationBuilder().
				WithTenantID(tenantID).
				WithTimezone("America/New_York").
				Build()
		})

		When("no active activities exist", func() {
			It("should not create any executions", func() {
				mockActivityRepository.findAllActiveActivities = []maintenanceDomain.Activity{}
				mockExecutionService.EXPECT().CreateExecution(gomock.Any(), gomock.Any()).Times(0)
				scheduler.ScheduleExecutions(ctx)
			})
		})

		When("active activities exist", func() {
			BeforeEach(func() {
				mockActivityRepository.findAllActiveActivities = []maintenanceDomain.Activity{activity}
			})

			When("executions do not exist", func() {
				BeforeEach(func() {
					mockTenantService.EXPECT().
						GetTenant(gomock.Any(), activity.TenantID).
						Return(tenant, nil).
						AnyTimes()
					mockTenantConfigurationService.EXPECT().
						GetOrCreateTenantConfiguration(ctx, tenant, "UTC").
						Return(tenantConfig, nil).
						AnyTimes()
					mockExecutionRepository.findByActivityAndScheduledDateError = maintenanceUsecases.ErrExecutionNotFound
				})

				It("should create three future executions", func() {
					mockExecutionService.EXPECT().
						CreateExecution(gomock.Any(), gomock.Any()).
						Return(nil).
						Times(3)
					mockBroker.EXPECT().
						Publish(gomock.Any(), async.BrokerTopicName("maintenance_executions"), gomock.Any()).
						Return(nil).
						Times(3)

					scheduler.ScheduleExecutions(ctx)
				})

				It("should populate field values from activity template", func() {
					var capturedExecutions []maintenanceDomain.Execution
					mockExecutionService.EXPECT().
						CreateExecution(gomock.Any(), gomock.Any()).
						DoAndReturn(func(_ context.Context, exec maintenanceDomain.Execution) error {
							capturedExecutions = append(capturedExecutions, exec)
							return nil
						}).
						Times(3)
					mockBroker.EXPECT().
						Publish(gomock.Any(), async.BrokerTopicName("maintenance_executions"), gomock.Any()).
						Return(nil).
						Times(3)

					scheduler.ScheduleExecutions(ctx)

					Expect(capturedExecutions).To(HaveLen(3))
					for _, exec := range capturedExecutions {
						Expect(exec.FieldValues).To(HaveKey("maintenance_type"))
						Expect(exec.FieldValues["maintenance_type"]).To(Equal("Filter Replacement"))
					}
				})

				It("should publish success events for each created execution", func() {
					mockExecutionService.EXPECT().
						CreateExecution(gomock.Any(), gomock.Any()).
						Return(nil).
						Times(3)
					mockBroker.EXPECT().
						Publish(gomock.Any(), async.BrokerTopicName("maintenance_executions"), gomock.Any()).
						DoAndReturn(func(_ context.Context, topic async.BrokerTopicName, msg async.BrokerMessage) error {
							Expect(topic).To(Equal(async.BrokerTopicName("maintenance_executions")))
							Expect(msg.Event).To(Equal("execution_created"))
							Expect(msg.Value).To(HaveKey("activity_id"))
							Expect(msg.Value).To(HaveKey("scheduled_date"))
							return nil
						}).
						Times(3)

					scheduler.ScheduleExecutions(ctx)
				})
			})

			When("executions already exist", func() {
				BeforeEach(func() {
					existingExecution, _ := maintenanceDomain.NewExecutionBuilder().
						WithActivityID(activity.ID).
						WithScheduledDate(time.Now().Add(24 * time.Hour)).
						Build()
					mockExecutionRepository.executionsByActivityAndDate[activity.ID.String()] = existingExecution
					mockExecutionRepository.findByActivityAndScheduledDateError = nil
					mockTenantService.EXPECT().
						GetTenant(gomock.Any(), activity.TenantID).
						Return(tenant, nil).
						AnyTimes()
					mockTenantConfigurationService.EXPECT().
						GetOrCreateTenantConfiguration(ctx, tenant, "UTC").
						Return(tenantConfig, nil).
						AnyTimes()
				})

				It("should skip creating duplicate executions", func() {
					mockExecutionService.EXPECT().
						CreateExecution(gomock.Any(), gomock.Any()).
						Times(0)

					scheduler.ScheduleExecutions(ctx)
				})
			})

			When("activity has invalid cron schedule", func() {
				var invalidActivity maintenanceDomain.Activity

				BeforeEach(func() {
					invalidActivity = activity
					invalidActivity.Schedule = maintenanceDomain.Schedule("invalid cron expression that will definitely fail")
					mockActivityRepository.findAllActiveActivities = []maintenanceDomain.Activity{invalidActivity}
				})

				It("should publish failure event", func() {
					mockTenantService.EXPECT().
						GetTenant(gomock.Any(), invalidActivity.TenantID).
						Return(tenant, nil)
					mockTenantConfigurationService.EXPECT().
						GetOrCreateTenantConfiguration(ctx, tenant, "UTC").
						Return(tenantConfig, nil)
					mockExecutionService.EXPECT().
						CreateExecution(gomock.Any(), gomock.Any()).
						Times(0)
					mockBroker.EXPECT().
						Publish(gomock.Any(), async.BrokerTopicName("maintenance_executions"), gomock.Any()).
						DoAndReturn(func(_ context.Context, topic async.BrokerTopicName, msg async.BrokerMessage) error {
							Expect(topic).To(Equal(async.BrokerTopicName("maintenance_executions")))
							Expect(msg.Event).To(Equal("execution_creation_failed"))
							Expect(msg.Value).To(HaveKey("activity_id"))
							Expect(msg.Value).To(HaveKey("error"))
							return nil
						})

					scheduler.ScheduleExecutions(ctx)
				})
			})

			When("tenant is not found", func() {
				BeforeEach(func() {
					mockActivityRepository.findAllActiveActivities = []maintenanceDomain.Activity{activity}
				})

				It("should publish failure event", func() {
					mockTenantService.EXPECT().
						GetTenant(gomock.Any(), activity.TenantID).
						Return(shareddomain.Tenant{}, controlPlaneUsecases.ErrTenantNotFound)
					mockExecutionService.EXPECT().
						CreateExecution(gomock.Any(), gomock.Any()).
						Times(0)
					mockBroker.EXPECT().
						Publish(gomock.Any(), async.BrokerTopicName("maintenance_executions"), gomock.Any()).
						DoAndReturn(func(_ context.Context, topic async.BrokerTopicName, msg async.BrokerMessage) error {
							Expect(topic).To(Equal(async.BrokerTopicName("maintenance_executions")))
							Expect(msg.Event).To(Equal("execution_creation_failed"))
							Expect(msg.Value).To(HaveKey("activity_id"))
							Expect(msg.Value).To(HaveKey("error"))
							return nil
						})

					scheduler.ScheduleExecutions(ctx)
				})
			})

			When("tenant configuration retrieval fails", func() {
				BeforeEach(func() {
					mockTenantService.EXPECT().
						GetTenant(gomock.Any(), activity.TenantID).
						Return(tenant, nil).
						AnyTimes()
					mockTenantConfigurationService.EXPECT().
						GetOrCreateTenantConfiguration(ctx, tenant, "UTC").
						Return(shareddomain.TenantConfiguration{}, errors.New("config error")).
						AnyTimes()
					mockExecutionRepository.findByActivityAndScheduledDateError = maintenanceUsecases.ErrExecutionNotFound
				})

				It("should use default timezone and continue processing", func() {
					mockExecutionService.EXPECT().
						CreateExecution(gomock.Any(), gomock.Any()).
						Return(nil).
						Times(3)
					mockBroker.EXPECT().
						Publish(gomock.Any(), async.BrokerTopicName("maintenance_executions"), gomock.Any()).
						Return(nil).
						Times(3)

					scheduler.ScheduleExecutions(ctx)
				})
			})

			When("execution creation fails", func() {
				BeforeEach(func() {
					mockTenantService.EXPECT().
						GetTenant(gomock.Any(), activity.TenantID).
						Return(tenant, nil).
						AnyTimes()
					mockTenantConfigurationService.EXPECT().
						GetOrCreateTenantConfiguration(ctx, tenant, "UTC").
						Return(tenantConfig, nil).
						AnyTimes()
					mockExecutionRepository.findByActivityAndScheduledDateError = maintenanceUsecases.ErrExecutionNotFound
					mockExecutionService.EXPECT().
						CreateExecution(gomock.Any(), gomock.Any()).
						Return(errors.New("creation failed")).
						Times(3)
				})

				It("should publish failure events and continue processing", func() {
					mockBroker.EXPECT().
						Publish(gomock.Any(), async.BrokerTopicName("maintenance_executions"), gomock.Any()).
						DoAndReturn(func(_ context.Context, topic async.BrokerTopicName, msg async.BrokerMessage) error {
							Expect(msg.Event).To(Equal("execution_creation_failed"))
							return nil
						}).
						Times(3)

					scheduler.ScheduleExecutions(ctx)
				})
			})

			When("checking execution existence fails", func() {
				BeforeEach(func() {
					mockTenantService.EXPECT().
						GetTenant(gomock.Any(), activity.TenantID).
						Return(tenant, nil).
						AnyTimes()
					mockTenantConfigurationService.EXPECT().
						GetOrCreateTenantConfiguration(ctx, tenant, "UTC").
						Return(tenantConfig, nil).
						AnyTimes()
					mockExecutionRepository.findByActivityAndScheduledDateError = errors.New("database error")
				})

				It("should publish failure event and continue", func() {
					mockBroker.EXPECT().
						Publish(gomock.Any(), async.BrokerTopicName("maintenance_executions"), gomock.Any()).
						DoAndReturn(func(_ context.Context, topic async.BrokerTopicName, msg async.BrokerMessage) error {
							Expect(msg.Event).To(Equal("execution_creation_failed"))
							return nil
						}).
						Times(3)

					scheduler.ScheduleExecutions(ctx)
				})
			})

			When("activity has no default field values", func() {
				BeforeEach(func() {
					activityType := maintenanceDomain.ActivityType{
						Name:         shareddomain.Name(maintenanceDomain.ActivityTypeWaterSystem),
						DisplayName:  shareddomain.DisplayName("Water System"),
						Description:  shareddomain.Description("Water system maintenance"),
						IsPredefined: true,
						Fields: []maintenanceDomain.FieldDefinition{
							{
								Name:         shareddomain.Name("maintenance_type"),
								DisplayName:  shareddomain.DisplayName("Maintenance Type"),
								Type:         maintenanceDomain.FieldTypeText,
								IsRequired:   true,
								DefaultValue: nil,
							},
						},
					}
					newActivity, _ := maintenanceDomain.NewActivityBuilder().
						WithTenantID(activity.TenantID).
						WithType(activityType).
						WithName(string(activity.Name)).
						WithDescription(string(activity.Description)).
						WithSchedule(string(activity.Schedule)).
						WithFields(activityType.Fields).
						Build()
					newActivity.ID = activity.ID
					newActivity.Activate()
					activity = newActivity
					mockActivityRepository.findAllActiveActivities = []maintenanceDomain.Activity{activity}
					mockExecutionRepository.findByActivityAndScheduledDateError = maintenanceUsecases.ErrExecutionNotFound
					mockTenantService.EXPECT().
						GetTenant(gomock.Any(), activity.TenantID).
						Return(tenant, nil).
						AnyTimes()
					mockTenantConfigurationService.EXPECT().
						GetOrCreateTenantConfiguration(ctx, tenant, "UTC").
						Return(tenantConfig, nil).
						AnyTimes()
				})

				It("should create executions with empty field values", func() {
					var capturedExecutions []maintenanceDomain.Execution
					mockExecutionService.EXPECT().
						CreateExecution(gomock.Any(), gomock.Any()).
						DoAndReturn(func(_ context.Context, exec maintenanceDomain.Execution) error {
							capturedExecutions = append(capturedExecutions, exec)
							return nil
						}).
						Times(3)
					mockBroker.EXPECT().
						Publish(gomock.Any(), async.BrokerTopicName("maintenance_executions"), gomock.Any()).
						Return(nil).
						Times(3)

					scheduler.ScheduleExecutions(ctx)

					Expect(capturedExecutions).To(HaveLen(3))
					for _, exec := range capturedExecutions {
						Expect(exec.FieldValues).NotTo(HaveKey("maintenance_type"))
					}
				})
			})
		})

		When("finding active activities fails", func() {
			BeforeEach(func() {
				mockActivityRepository.findAllActiveError = errors.New("database error")
			})

			It("should not create any executions", func() {
				mockExecutionService.EXPECT().
					CreateExecution(gomock.Any(), gomock.Any()).
					Times(0)

				scheduler.ScheduleExecutions(ctx)
			})
		})
	})
})

type mockActivityRepository struct {
	findAllActiveActivities []maintenanceDomain.Activity
	findAllActiveError      error
}

func newMockActivityRepository() *mockActivityRepository {
	return &mockActivityRepository{
		findAllActiveActivities: []maintenanceDomain.Activity{},
	}
}

func (m *mockActivityRepository) Create(ctx context.Context, activity maintenanceDomain.Activity) error {
	return nil
}

func (m *mockActivityRepository) GetByID(ctx context.Context, id shareddomain.ID) (maintenanceDomain.Activity, error) {
	return maintenanceDomain.Activity{}, nil
}

func (m *mockActivityRepository) FindAllByTenant(ctx context.Context, tenantID shareddomain.ID, pagination maintenanceUsecases.Pagination) ([]maintenanceDomain.Activity, int, error) {
	return nil, 0, nil
}

func (m *mockActivityRepository) FindAllActive(ctx context.Context) ([]maintenanceDomain.Activity, error) {
	if m.findAllActiveError != nil {
		return nil, m.findAllActiveError
	}
	return m.findAllActiveActivities, nil
}

func (m *mockActivityRepository) Update(ctx context.Context, activity maintenanceDomain.Activity) error {
	return nil
}

func (m *mockActivityRepository) Delete(ctx context.Context, id shareddomain.ID) error {
	return nil
}

type mockExecutionRepository struct {
	executionsByActivityAndDate         map[string]maintenanceDomain.Execution
	findByActivityAndScheduledDateError error
}

func newMockExecutionRepository() *mockExecutionRepository {
	return &mockExecutionRepository{
		executionsByActivityAndDate:         make(map[string]maintenanceDomain.Execution),
		findByActivityAndScheduledDateError: maintenanceUsecases.ErrExecutionNotFound,
	}
}

func (m *mockExecutionRepository) Create(ctx context.Context, execution maintenanceDomain.Execution) error {
	return nil
}

func (m *mockExecutionRepository) GetByID(ctx context.Context, id shareddomain.ID) (maintenanceDomain.Execution, error) {
	return maintenanceDomain.Execution{}, nil
}

func (m *mockExecutionRepository) FindAllByActivity(ctx context.Context, activityID shareddomain.ID, pagination maintenanceUsecases.Pagination) ([]maintenanceDomain.Execution, int, error) {
	return nil, 0, nil
}

func (m *mockExecutionRepository) FindByActivityAndScheduledDate(ctx context.Context, activityID shareddomain.ID, scheduledDate time.Time) (maintenanceDomain.Execution, error) {
	if m.findByActivityAndScheduledDateError != nil {
		if errors.Is(m.findByActivityAndScheduledDateError, maintenanceUsecases.ErrExecutionNotFound) {
			return maintenanceDomain.Execution{}, maintenanceUsecases.ErrExecutionNotFound
		}
		return maintenanceDomain.Execution{}, m.findByActivityAndScheduledDateError
	}
	key := activityID.String()
	if exec, ok := m.executionsByActivityAndDate[key]; ok {
		return exec, nil
	}
	return maintenanceDomain.Execution{}, maintenanceUsecases.ErrExecutionNotFound
}

func (m *mockExecutionRepository) Update(ctx context.Context, execution maintenanceDomain.Execution) error {
	return nil
}

func (m *mockExecutionRepository) MarkCompleted(ctx context.Context, id shareddomain.ID, completedBy string) error {
	return nil
}

func (m *mockExecutionRepository) FindAllOverdue(ctx context.Context, tenantID shareddomain.ID) ([]maintenanceDomain.Execution, error) {
	return nil, nil
}

func (m *mockExecutionRepository) FindAllDueSoon(ctx context.Context, tenantID shareddomain.ID, days int) ([]maintenanceDomain.Execution, error) {
	return nil, nil
}

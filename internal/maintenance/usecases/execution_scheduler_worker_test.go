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
	mocksharedkernel "zensor-server/test/unit/doubles/shared_kernel/usecases"
	mockasync "zensor-server/test/unit/doubles/infra/async"
	mockmaintenance "zensor-server/test/unit/doubles/maintenance/usecases"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("ExecutionWorker", func() {
	var (
		ctrl                           *gomock.Controller
		ticker                         *time.Ticker
		mockActivityRepository         *mockmaintenance.MockActivityRepository
		mockExecutionRepository        *mockmaintenance.MockExecutionRepository
		mockExecutionService           *mockmaintenance.MockExecutionService
		mockTenantService              *mocksharedkernel.MockTenantService
		mockTenantConfigurationService *mocksharedkernel.MockTenantConfigurationService
		mockBroker                     *mockasync.MockInternalBroker
		worker                         *maintenanceUsecases.ExecutionWorker
		ctx                            context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ticker = time.NewTicker(100 * time.Millisecond)
		mockActivityRepository = mockmaintenance.NewMockActivityRepository(ctrl)
		mockExecutionRepository = mockmaintenance.NewMockExecutionRepository(ctrl)
		mockExecutionService = mockmaintenance.NewMockExecutionService(ctrl)
		mockTenantService = mocksharedkernel.NewMockTenantService(ctrl)
		mockTenantConfigurationService = mocksharedkernel.NewMockTenantConfigurationService(ctrl)
		mockBroker = mockasync.NewMockInternalBroker(ctrl)
		ctx = context.Background()

		worker = maintenanceUsecases.NewExecutionWorker(
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
				mockActivityRepository.EXPECT().
					FindAllActive(gomock.Any()).
					Return([]maintenanceDomain.Activity{}, nil)
				mockExecutionService.EXPECT().CreateExecution(gomock.Any(), gomock.Any()).Times(0)
				worker.ScheduleExecutions(ctx)
			})
		})

		When("active activities exist and executions do not exist", func() {
			BeforeEach(func() {
				mockActivityRepository.EXPECT().
					FindAllActive(gomock.Any()).
					Return([]maintenanceDomain.Activity{activity}, nil)
				mockTenantService.EXPECT().
					GetTenant(gomock.Any(), activity.TenantID).
					Return(tenant, nil).
					AnyTimes()
				mockTenantConfigurationService.EXPECT().
					GetOrCreateTenantConfiguration(ctx, tenant, "UTC").
					Return(tenantConfig, nil).
					AnyTimes()
				mockExecutionRepository.EXPECT().
					FindByActivityAndScheduledDate(gomock.Any(), activity.ID, gomock.Any()).
					Return(maintenanceDomain.Execution{}, maintenanceUsecases.ErrExecutionNotFound).
					AnyTimes()
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

				worker.ScheduleExecutions(ctx)
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

				worker.ScheduleExecutions(ctx)

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

				worker.ScheduleExecutions(ctx)
			})
		})

		When("active activities exist and executions already exist", func() {
			BeforeEach(func() {
				mockActivityRepository.EXPECT().
					FindAllActive(gomock.Any()).
					Return([]maintenanceDomain.Activity{activity}, nil)
				existingExecution, _ := maintenanceDomain.NewExecutionBuilder().
					WithActivityID(activity.ID).
					WithScheduledDate(time.Now().Add(24 * time.Hour)).
					Build()
				mockExecutionRepository.EXPECT().
					FindByActivityAndScheduledDate(gomock.Any(), activity.ID, gomock.Any()).
					Return(existingExecution, nil).
					AnyTimes()
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

				worker.ScheduleExecutions(ctx)
			})
		})

		When("active activities exist and activity has invalid cron schedule", func() {
			var invalidActivity maintenanceDomain.Activity

			BeforeEach(func() {
				invalidActivity = activity
				invalidActivity.Schedule = maintenanceDomain.Schedule("invalid cron expression that will definitely fail")
				mockActivityRepository.EXPECT().
					FindAllActive(gomock.Any()).
					Return([]maintenanceDomain.Activity{invalidActivity}, nil)
				mockTenantService.EXPECT().
					GetTenant(gomock.Any(), invalidActivity.TenantID).
					Return(tenant, nil)
				mockTenantConfigurationService.EXPECT().
					GetOrCreateTenantConfiguration(ctx, tenant, "UTC").
					Return(tenantConfig, nil)
			})

			It("should publish failure event", func() {
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

				worker.ScheduleExecutions(ctx)
			})
		})

		When("active activities exist and tenant is not found", func() {
			BeforeEach(func() {
				mockActivityRepository.EXPECT().
					FindAllActive(gomock.Any()).
					Return([]maintenanceDomain.Activity{activity}, nil)
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

				worker.ScheduleExecutions(ctx)
			})
		})

		When("active activities exist and tenant configuration retrieval fails", func() {
			BeforeEach(func() {
				mockActivityRepository.EXPECT().
					FindAllActive(gomock.Any()).
					Return([]maintenanceDomain.Activity{activity}, nil)
				mockTenantService.EXPECT().
					GetTenant(gomock.Any(), activity.TenantID).
					Return(tenant, nil).
					AnyTimes()
				mockTenantConfigurationService.EXPECT().
					GetOrCreateTenantConfiguration(ctx, tenant, "UTC").
					Return(shareddomain.TenantConfiguration{}, errors.New("config error")).
					AnyTimes()
				mockExecutionRepository.EXPECT().
					FindByActivityAndScheduledDate(gomock.Any(), activity.ID, gomock.Any()).
					Return(maintenanceDomain.Execution{}, maintenanceUsecases.ErrExecutionNotFound).
					AnyTimes()
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

				worker.ScheduleExecutions(ctx)
			})
		})

		When("active activities exist and execution creation fails", func() {
			BeforeEach(func() {
				mockActivityRepository.EXPECT().
					FindAllActive(gomock.Any()).
					Return([]maintenanceDomain.Activity{activity}, nil)
				mockTenantService.EXPECT().
					GetTenant(gomock.Any(), activity.TenantID).
					Return(tenant, nil).
					AnyTimes()
				mockTenantConfigurationService.EXPECT().
					GetOrCreateTenantConfiguration(ctx, tenant, "UTC").
					Return(tenantConfig, nil).
					AnyTimes()
				mockExecutionRepository.EXPECT().
					FindByActivityAndScheduledDate(gomock.Any(), activity.ID, gomock.Any()).
					Return(maintenanceDomain.Execution{}, maintenanceUsecases.ErrExecutionNotFound).
					AnyTimes()
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

				worker.ScheduleExecutions(ctx)
			})
		})

		When("active activities exist and checking execution existence fails", func() {
			BeforeEach(func() {
				mockActivityRepository.EXPECT().
					FindAllActive(gomock.Any()).
					Return([]maintenanceDomain.Activity{activity}, nil)
				mockTenantService.EXPECT().
					GetTenant(gomock.Any(), activity.TenantID).
					Return(tenant, nil).
					AnyTimes()
				mockTenantConfigurationService.EXPECT().
					GetOrCreateTenantConfiguration(ctx, tenant, "UTC").
					Return(tenantConfig, nil).
					AnyTimes()
				mockExecutionRepository.EXPECT().
					FindByActivityAndScheduledDate(gomock.Any(), activity.ID, gomock.Any()).
					Return(maintenanceDomain.Execution{}, errors.New("database error")).
					AnyTimes()
			})

			It("should publish failure event and continue", func() {
				mockBroker.EXPECT().
					Publish(gomock.Any(), async.BrokerTopicName("maintenance_executions"), gomock.Any()).
					DoAndReturn(func(_ context.Context, topic async.BrokerTopicName, msg async.BrokerMessage) error {
						Expect(msg.Event).To(Equal("execution_creation_failed"))
						return nil
					}).
					Times(3)

				worker.ScheduleExecutions(ctx)
			})
		})

		When("active activities exist and activity has no default field values", func() {
			var activityWithoutDefaults maintenanceDomain.Activity

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
				activityWithoutDefaults, _ = maintenanceDomain.NewActivityBuilder().
					WithTenantID(activity.TenantID).
					WithType(activityType).
					WithName(string(activity.Name)).
					WithDescription(string(activity.Description)).
					WithSchedule(string(activity.Schedule)).
					WithFields(activityType.Fields).
					Build()
				activityWithoutDefaults.ID = activity.ID
				activityWithoutDefaults.Activate()
				mockActivityRepository.EXPECT().
					FindAllActive(gomock.Any()).
					Return([]maintenanceDomain.Activity{activityWithoutDefaults}, nil)
				mockExecutionRepository.EXPECT().
					FindByActivityAndScheduledDate(gomock.Any(), activityWithoutDefaults.ID, gomock.Any()).
					Return(maintenanceDomain.Execution{}, maintenanceUsecases.ErrExecutionNotFound).
					AnyTimes()
				mockTenantService.EXPECT().
					GetTenant(gomock.Any(), activityWithoutDefaults.TenantID).
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

				worker.ScheduleExecutions(ctx)

				Expect(capturedExecutions).To(HaveLen(3))
				for _, exec := range capturedExecutions {
					Expect(exec.FieldValues).NotTo(HaveKey("maintenance_type"))
				}
			})
		})

		When("finding active activities fails", func() {
			BeforeEach(func() {
				mockActivityRepository.EXPECT().
					FindAllActive(gomock.Any()).
					Return(nil, errors.New("database error"))
			})

			It("should not create any executions", func() {
				mockExecutionService.EXPECT().
					CreateExecution(gomock.Any(), gomock.Any()).
					Times(0)

				worker.ScheduleExecutions(ctx)
			})
		})
	})
})

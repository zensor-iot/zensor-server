package usecases_test

import (
	"context"
	"errors"
	"sync"
	"time"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/notification"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"
	usecases_mocks "zensor-server/test/unit/doubles/control_plane/usecases"
	notification_mocks "zensor-server/test/unit/doubles/infra/notification"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("NotificationWorker", func() {
	var (
		ctrl                    *gomock.Controller
		ticker                  *time.Ticker
		mockNotificationClient  *notification_mocks.MockNotificationClient
		mockDeviceService       *usecases_mocks.MockDeviceService
		mockTenantConfigService *usecases_mocks.MockTenantConfigurationService
		mockTaskService         *usecases_mocks.MockTaskService
		worker                  *usecases.NotificationWorker
		ctx                     context.Context
		cancel                  context.CancelFunc
		realBroker              *async.LocalBroker
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ticker = time.NewTicker(1 * time.Second)
		mockNotificationClient = notification_mocks.NewMockNotificationClient(ctrl)
		mockDeviceService = usecases_mocks.NewMockDeviceService(ctrl)
		mockTenantConfigService = usecases_mocks.NewMockTenantConfigurationService(ctrl)
		mockTaskService = usecases_mocks.NewMockTaskService(ctrl)
		realBroker = async.NewLocalBroker()
		ctx, cancel = context.WithCancel(context.Background())

		worker = usecases.NewNotificationWorker(
			ticker,
			mockNotificationClient,
			mockDeviceService,
			mockTenantConfigService,
			mockTaskService,
			realBroker,
		)
	})

	AfterEach(func() {
		ticker.Stop()
		cancel()
		realBroker.Stop()
		ctrl.Finish()
	})

	Context("Run", func() {
		var (
			deviceID          domain.ID
			tenantID          domain.ID
			scheduledTaskID   domain.ID
			taskID            domain.ID
			notificationEmail string
			device            domain.Device
			tenantConfig      domain.TenantConfiguration
			scheduledTask     domain.ScheduledTask
			task              domain.Task
		)

		BeforeEach(func() {
			deviceID = domain.ID("device-123")
			tenantID = domain.ID("tenant-456")
			scheduledTaskID = domain.ID("scheduled-task-456")
			taskID = domain.ID("task-123")
			notificationEmail = "admin@example.com"

			device = domain.Device{
				ID:          deviceID,
				Name:        "Test Device",
				DisplayName: "Test Device Display",
				AppEUI:      "app-eui-123",
				DevEUI:      "dev-eui-123",
				TenantID:    &tenantID,
			}

			tenantConfig = domain.TenantConfiguration{
				ID:                domain.ID("config-789"),
				TenantID:          tenantID,
				Timezone:          "UTC",
				NotificationEmail: notificationEmail,
				Version:           1,
				CreatedAt:         time.Now(),
				UpdatedAt:         time.Now(),
			}

			scheduledTask = domain.ScheduledTask{
				ID:        scheduledTaskID,
				Device:    device,
				Tenant:    domain.Tenant{ID: tenantID},
				IsActive:  true,
				CreatedAt: utils.Time{Time: time.Now()},
				UpdatedAt: utils.Time{Time: time.Now()},
			}

			task = domain.Task{
				ID:            taskID,
				Device:        device,
				ScheduledTask: &scheduledTask,
				Commands:      []domain.Command{},
				CreatedAt:     utils.Time{Time: time.Now()},
			}
		})

		When("worker receives a valid scheduled task executed event", func() {
			BeforeEach(func() {
				mockTaskService.EXPECT().
					FindAllByScheduledTask(gomock.Any(), scheduledTaskID, gomock.Any()).
					Return([]domain.Task{task}, 1, nil).
					Times(1)

				mockDeviceService.EXPECT().
					GetDevice(gomock.Any(), deviceID).
					Return(device, nil).
					Times(1)

				mockTenantConfigService.EXPECT().
					GetTenantConfiguration(gomock.Any(), domain.Tenant{ID: tenantID}).
					Return(tenantConfig, nil).
					Times(1)

				mockNotificationClient.EXPECT().
					SendEmail(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, req notification.EmailRequest) error {
						Expect(req.To).To(Equal(notificationEmail))
						Expect(req.Subject).To(ContainSubstring("New Task Created for Device"))
						Expect(req.Body).To(ContainSubstring("Test Device Display"))
						return nil
					}).
					Times(1)
			})

			It("should process the message and send notification", func() {
				var wg sync.WaitGroup
				wg.Add(1)

				go func() {
					defer wg.Done()
					worker.Run(ctx, func() {})
				}()

				time.Sleep(100 * time.Millisecond)

				brokerMessage := async.BrokerMessage{
					Event: "scheduled_task_executed",
					Value: scheduledTask,
				}
				err := realBroker.Publish(ctx, async.BrokerTopicName("scheduled_tasks"), brokerMessage)
				Expect(err).To(BeNil())

				time.Sleep(200 * time.Millisecond)

				cancel()
				wg.Wait()
			})
		})

		When("device has no tenant assigned", func() {
			var deviceWithoutTenant domain.Device
			var taskWithoutTenant domain.Task

			BeforeEach(func() {
				deviceWithoutTenant = domain.Device{
					ID:          deviceID,
					Name:        "Test Device",
					DisplayName: "Test Device Display",
					TenantID:    nil,
				}

				taskWithoutTenant = domain.Task{
					ID:            taskID,
					Device:        deviceWithoutTenant,
					ScheduledTask: &scheduledTask,
					Commands:      []domain.Command{},
					CreatedAt:     utils.Time{Time: time.Now()},
				}

				mockTaskService.EXPECT().
					FindAllByScheduledTask(gomock.Any(), scheduledTaskID, gomock.Any()).
					Return([]domain.Task{taskWithoutTenant}, 1, nil).
					Times(1)

				mockDeviceService.EXPECT().
					GetDevice(gomock.Any(), deviceID).
					Return(deviceWithoutTenant, nil).
					Times(1)
			})

			It("should skip notification and not call other services", func() {
				var wg sync.WaitGroup
				wg.Add(1)

				go func() {
					defer wg.Done()
					worker.Run(ctx, func() {})
				}()

				time.Sleep(100 * time.Millisecond)

				brokerMessage := async.BrokerMessage{
					Event: "scheduled_task_executed",
					Value: scheduledTask,
				}
				err := realBroker.Publish(ctx, async.BrokerTopicName("scheduled_tasks"), brokerMessage)
				Expect(err).To(BeNil())

				time.Sleep(200 * time.Millisecond)
				cancel()
				wg.Wait()
			})
		})

		When("tenant has no notification email configured", func() {
			var tenantConfigWithoutEmail domain.TenantConfiguration

			BeforeEach(func() {
				tenantConfigWithoutEmail = domain.TenantConfiguration{
					ID:                domain.ID("config-789"),
					TenantID:          tenantID,
					Timezone:          "UTC",
					NotificationEmail: "",
					Version:           1,
					CreatedAt:         time.Now(),
					UpdatedAt:         time.Now(),
				}

				mockTaskService.EXPECT().
					FindAllByScheduledTask(gomock.Any(), scheduledTaskID, gomock.Any()).
					Return([]domain.Task{task}, 1, nil).
					Times(1)

				mockDeviceService.EXPECT().
					GetDevice(gomock.Any(), deviceID).
					Return(device, nil).
					Times(1)

				mockTenantConfigService.EXPECT().
					GetTenantConfiguration(gomock.Any(), domain.Tenant{ID: tenantID}).
					Return(tenantConfigWithoutEmail, nil).
					Times(1)
			})

			It("should skip notification and not call notification client", func() {
				var wg sync.WaitGroup
				wg.Add(1)

				go func() {
					defer wg.Done()
					worker.Run(ctx, func() {})
				}()

				time.Sleep(100 * time.Millisecond)

				brokerMessage := async.BrokerMessage{
					Event: "scheduled_task_executed",
					Value: scheduledTask,
				}
				err := realBroker.Publish(ctx, async.BrokerTopicName("scheduled_tasks"), brokerMessage)
				Expect(err).To(BeNil())

				time.Sleep(200 * time.Millisecond)
				cancel()
				wg.Wait()
			})
		})

		When("task service returns an error", func() {
			BeforeEach(func() {
				mockTaskService.EXPECT().
					FindAllByScheduledTask(gomock.Any(), scheduledTaskID, gomock.Any()).
					Return(nil, 0, errors.New("failed to find tasks")).
					Times(1)
			})

			It("should handle error gracefully", func() {
				var wg sync.WaitGroup
				wg.Add(1)

				go func() {
					defer wg.Done()
					worker.Run(ctx, func() {})
				}()

				time.Sleep(100 * time.Millisecond)

				brokerMessage := async.BrokerMessage{
					Event: "scheduled_task_executed",
					Value: scheduledTask,
				}
				err := realBroker.Publish(ctx, async.BrokerTopicName("scheduled_tasks"), brokerMessage)
				Expect(err).To(BeNil())

				time.Sleep(200 * time.Millisecond)
				cancel()
				wg.Wait()
			})
		})

		When("device service returns an error", func() {
			BeforeEach(func() {
				mockTaskService.EXPECT().
					FindAllByScheduledTask(gomock.Any(), scheduledTaskID, gomock.Any()).
					Return([]domain.Task{task}, 1, nil).
					Times(1)

				mockDeviceService.EXPECT().
					GetDevice(gomock.Any(), deviceID).
					Return(domain.Device{}, errors.New("device not found")).
					Times(1)
			})

			It("should handle error gracefully", func() {
				var wg sync.WaitGroup
				wg.Add(1)

				go func() {
					defer wg.Done()
					worker.Run(ctx, func() {})
				}()

				time.Sleep(100 * time.Millisecond)

				brokerMessage := async.BrokerMessage{
					Event: "scheduled_task_executed",
					Value: scheduledTask,
				}
				err := realBroker.Publish(ctx, async.BrokerTopicName("scheduled_tasks"), brokerMessage)
				Expect(err).To(BeNil())

				time.Sleep(200 * time.Millisecond)
				cancel()
				wg.Wait()
			})
		})

		When("notification client returns an error", func() {
			BeforeEach(func() {
				mockTaskService.EXPECT().
					FindAllByScheduledTask(gomock.Any(), scheduledTaskID, gomock.Any()).
					Return([]domain.Task{task}, 1, nil).
					Times(1)

				mockDeviceService.EXPECT().
					GetDevice(gomock.Any(), deviceID).
					Return(device, nil).
					Times(1)

				mockTenantConfigService.EXPECT().
					GetTenantConfiguration(gomock.Any(), domain.Tenant{ID: tenantID}).
					Return(tenantConfig, nil).
					Times(1)

				mockNotificationClient.EXPECT().
					SendEmail(gomock.Any(), gomock.Any()).
					Return(errors.New("failed to send email")).
					Times(1)
			})

			It("should handle error gracefully", func() {
				var wg sync.WaitGroup
				wg.Add(1)

				go func() {
					defer wg.Done()
					worker.Run(ctx, func() {})
				}()

				time.Sleep(100 * time.Millisecond)

				brokerMessage := async.BrokerMessage{
					Event: "scheduled_task_executed",
					Value: scheduledTask,
				}
				err := realBroker.Publish(ctx, async.BrokerTopicName("scheduled_tasks"), brokerMessage)
				Expect(err).To(BeNil())

				time.Sleep(200 * time.Millisecond)
				cancel()
				wg.Wait()
			})
		})
	})
})

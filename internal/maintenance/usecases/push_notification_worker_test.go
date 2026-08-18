package usecases_test

import (
	"context"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/config"
	"zensor-server/internal/infra/notification"
	"zensor-server/internal/shared_kernel/domain"

	maintenanceUsecases "zensor-server/internal/maintenance/usecases"

	mocknotification "zensor-server/test/unit/doubles/infra/notification"
	mocksharedusecases "zensor-server/test/unit/doubles/shared_kernel/usecases"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("PushNotificationWorker", func() {
	Context("Run", func() {
		var (
			ctrl               *gomock.Controller
			notificationClient *mocknotification.MockNotificationClient
			pushTokenService   *mocksharedusecases.MockPushTokenService
			userService        *mocksharedusecases.MockUserService
			broker             *async.LocalBroker
			worker             *maintenanceUsecases.PushNotificationWorker
			cancel             context.CancelFunc
			publishMessage     func()
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			notificationClient = mocknotification.NewMockNotificationClient(ctrl)
			pushTokenService = mocksharedusecases.NewMockPushTokenService(ctrl)
			userService = mocksharedusecases.NewMockUserService(ctrl)
			broker = async.NewLocalBroker()

			cfg := config.PushNotificationWorkerConfig{
				Name:      "test_notification",
				Topic:     "maintenance_executions",
				EventType: "execution_ready_for_notification",

				TenantIDPath: "tenant_id",
				UserIDPath:   "user_id",
				Title:        "Execution Reminder",
				Body:         "Scheduled execution is due soon",
				DeepLink:     "/maintenance/executions",
			}

			var err error
			worker, err = maintenanceUsecases.NewPushNotificationWorker(
				cfg, broker, notificationClient, pushTokenService, userService)
			Expect(err).NotTo(HaveOccurred())

			var ctx context.Context
			ctx, cancel = context.WithCancel(context.Background())
			go worker.Run(ctx, func() {})
			DeferCleanup(func() {
				cancel()
				broker.Stop()
			})

			publishMessage = func() {
				msg := async.BrokerMessage{
					Event: "execution_ready_for_notification",
					Value: map[string]any{
						"tenant_id": "tenant-1",
						"user_id":   "user-1",
					},
				}
				Eventually(func() error {
					return broker.Publish(context.Background(), "maintenance_executions", msg)
				}).Should(Succeed())
			}
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		When("the user has tokens for multiple platforms", func() {
			It("should propagate each token's platform on the request", func() {
				tokens := []domain.PushToken{
					{ID: "tok-a", UserID: "user-1", Token: "fcm-token", Platform: "android"},                       // #nosec G101 -- test fixture
					{ID: "tok-w", UserID: "user-1", Token: `{"endpoint":"https://push.example"}`, Platform: "web"}, // #nosec G101 -- test fixture
				}
				pushTokenService.EXPECT().
					ListTokensByUserID(gomock.Any(), domain.ID("user-1")).
					Return(tokens, nil)

				sent := make(chan notification.PushNotificationRequest, 2)
				notificationClient.EXPECT().
					SendPushNotification(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, req notification.PushNotificationRequest) error {
						sent <- req
						return nil
					}).
					Times(2)

				publishMessage()

				var first, second notification.PushNotificationRequest
				Eventually(sent).Should(Receive(&first))
				Eventually(sent).Should(Receive(&second))
				platforms := []string{first.Platform, second.Platform}
				Expect(platforms).To(ConsistOf("android", "web"))
			})
		})

		When("a web subscription is expired", func() {
			It("should unregister the dead token", func() {
				webToken := `{"endpoint":"https://push.example/expired"}` // #nosec G101 -- test fixture
				tokens := []domain.PushToken{
					{ID: "tok-w", UserID: "user-1", Token: webToken, Platform: "web"},
				}
				pushTokenService.EXPECT().
					ListTokensByUserID(gomock.Any(), domain.ID("user-1")).
					Return(tokens, nil)

				notificationClient.EXPECT().
					SendPushNotification(gomock.Any(), gomock.Any()).
					Return(notification.ErrSubscriptionExpired)

				unregistered := make(chan string, 1)
				pushTokenService.EXPECT().
					UnregisterToken(gomock.Any(), webToken).
					DoAndReturn(func(_ context.Context, token string) error {
						unregistered <- token
						return nil
					})

				publishMessage()

				Eventually(unregistered).Should(Receive(Equal(webToken)))
			})
		})
	})

	Context("TitleTemplate", func() {
		var (
			ctrl               *gomock.Controller
			notificationClient *mocknotification.MockNotificationClient
			pushTokenService   *mocksharedusecases.MockPushTokenService
			userService        *mocksharedusecases.MockUserService
			broker             *async.LocalBroker
			worker             *maintenanceUsecases.PushNotificationWorker
			cancel             context.CancelFunc
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			notificationClient = mocknotification.NewMockNotificationClient(ctrl)
			pushTokenService = mocksharedusecases.NewMockPushTokenService(ctrl)
			userService = mocksharedusecases.NewMockUserService(ctrl)
			broker = async.NewLocalBroker()

			cfg := config.PushNotificationWorkerConfig{
				Name:             "test_notification",
				Topic:            "maintenance_executions",
				EventType:        "execution_ready_for_notification",
				TenantIDPath:     "tenant_id",
				UserIDPath:       "user_id",
				Title:            "Execution Reminder",
				TitleTemplate:    "Execution Reminder: {{activity_name}}",
				Body:             "Scheduled execution is due soon",
				DeepLink:         "/maintenance/executions",
				DeepLinkTemplate: "/maintenance/executions/%s",
			}

			var err error
			worker, err = maintenanceUsecases.NewPushNotificationWorker(
				cfg, broker, notificationClient, pushTokenService, userService)
			Expect(err).NotTo(HaveOccurred())

			var ctx context.Context
			ctx, cancel = context.WithCancel(context.Background())
			go worker.Run(ctx, func() {})
			DeferCleanup(func() {
				cancel()
				broker.Stop()
			})
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		When("a reminder notification is published", func() {
			It("should include the specific activity name in the title", func() {
				tokens := []domain.PushToken{
					{ID: "tok-a", UserID: "user-1", Token: "fcm-token", Platform: "android"},
				}
				pushTokenService.EXPECT().
					ListTokensByUserID(gomock.Any(), domain.ID("user-1")).
					Return(tokens, nil)

				sent := make(chan notification.PushNotificationRequest, 1)
				notificationClient.EXPECT().
					SendPushNotification(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, req notification.PushNotificationRequest) error {
						sent <- req
						return nil
					}).
					Times(1)

				msg := async.BrokerMessage{
					Event: "execution_ready_for_notification",
					Value: map[string]any{
						"tenant_id":     "tenant-1",
						"user_id":       "user-1",
						"execution_id":  "execution-1",
						"activity_id":   "activity-1",
						"activity_name": "Filter Replacement",
					},
				}
				Eventually(func() error {
					return broker.Publish(context.Background(), "maintenance_executions", msg)
				}).Should(Succeed())

				var req notification.PushNotificationRequest
				Eventually(sent).Should(Receive(&req))
				Expect(req.Title).To(Equal("Execution Reminder: Filter Replacement"))
				Expect(req.DeepLink).To(Equal("/maintenance/executions/execution-1"))
			})
		})

		When("the activity name is missing from the message", func() {
			It("should fall back to the generic title", func() {
				tokens := []domain.PushToken{
					{ID: "tok-a", UserID: "user-1", Token: "fcm-token", Platform: "android"},
				}
				pushTokenService.EXPECT().
					ListTokensByUserID(gomock.Any(), domain.ID("user-1")).
					Return(tokens, nil)

				sent := make(chan notification.PushNotificationRequest, 1)
				notificationClient.EXPECT().
					SendPushNotification(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, req notification.PushNotificationRequest) error {
						sent <- req
						return nil
					}).
					Times(1)

				msg := async.BrokerMessage{
					Event: "execution_ready_for_notification",
					Value: map[string]any{
						"tenant_id":    "tenant-1",
						"user_id":      "user-1",
						"execution_id": "execution-1",
					},
				}
				Eventually(func() error {
					return broker.Publish(context.Background(), "maintenance_executions", msg)
				}).Should(Succeed())

				var req notification.PushNotificationRequest
				Eventually(sent).Should(Receive(&req))
				Expect(req.Title).To(Equal("Execution Reminder"))
			})
		})
	})
})

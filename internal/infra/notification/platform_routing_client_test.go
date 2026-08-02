package notification_test

import (
	"context"
	"errors"

	"zensor-server/internal/infra/notification"
	mocknotification "zensor-server/test/unit/doubles/infra/notification"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("PlatformRoutingPushClient", func() {
	var (
		ctrl    *gomock.Controller
		fcm     *mocknotification.MockNotificationClient
		webPush *mocknotification.MockNotificationClient
		client  *notification.PlatformRoutingPushClient
		request notification.PushNotificationRequest
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		fcm = mocknotification.NewMockNotificationClient(ctrl)
		webPush = mocknotification.NewMockNotificationClient(ctrl)
		client = notification.NewPlatformRoutingPushClient(fcm, webPush)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("SendPushNotification", func() {
		When("the platform is web", func() {
			BeforeEach(func() {
				request = notification.PushNotificationRequest{
					Token:    `{"endpoint":"https://push.example"}`,
					Title:    "t",
					Body:     "b",
					Platform: "web",
				}
			})

			It("should route to the web push client", func() {
				webPush.EXPECT().SendPushNotification(gomock.Any(), request).Return(nil)

				err := client.SendPushNotification(context.Background(), request)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should propagate web push client errors", func() {
				webPush.EXPECT().SendPushNotification(gomock.Any(), request).Return(notification.ErrSubscriptionExpired)

				err := client.SendPushNotification(context.Background(), request)
				Expect(err).To(MatchError(notification.ErrSubscriptionExpired))
			})
		})

		When("the platform is ios", func() {
			BeforeEach(func() {
				request = notification.PushNotificationRequest{Token: "fcm-token", Platform: "ios"}
			})

			It("should route to the FCM client", func() {
				fcm.EXPECT().SendPushNotification(gomock.Any(), request).Return(nil)

				err := client.SendPushNotification(context.Background(), request)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("the platform is android", func() {
			BeforeEach(func() {
				request = notification.PushNotificationRequest{Token: "fcm-token", Platform: "android"}
			})

			It("should route to the FCM client", func() {
				fcm.EXPECT().SendPushNotification(gomock.Any(), request).Return(nil)

				err := client.SendPushNotification(context.Background(), request)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("the platform is empty", func() {
			BeforeEach(func() {
				request = notification.PushNotificationRequest{Token: "fcm-token"}
			})

			It("should route to the FCM client", func() {
				fcm.EXPECT().SendPushNotification(gomock.Any(), request).Return(nil)

				err := client.SendPushNotification(context.Background(), request)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should propagate FCM client errors", func() {
				sendErr := errors.New("fcm unavailable")
				fcm.EXPECT().SendPushNotification(gomock.Any(), request).Return(sendErr)

				err := client.SendPushNotification(context.Background(), request)
				Expect(err).To(MatchError(sendErr))
			})
		})
	})

	Context("SendEmail", func() {
		It("should return an error", func() {
			err := client.SendEmail(context.Background(), notification.EmailRequest{To: "a@b.c"})
			Expect(err).To(HaveOccurred())
		})
	})
})

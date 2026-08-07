package notification_test

import (
	"context"
	"path/filepath"

	"zensor-server/internal/infra/notification"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("FCMPushClient", func() {
	ginkgo.Context("NewFCMPushClient", func() {
		var ctx context.Context
		var config notification.FCMConfig
		var client notification.NotificationClient

		ginkgo.BeforeEach(func() {
			ctx = context.Background()
		})

		ginkgo.JustBeforeEach(func() {
			client = notification.NewFCMPushClient(ctx, config)
		})

		ginkgo.When("the service account file is missing", func() {
			ginkgo.BeforeEach(func() {
				config = notification.FCMConfig{
					ProjectID:          "my-project",
					ServiceAccountPath: filepath.Join("testdata", "missing_service_account.json"),
				}
			})

			ginkgo.It("should return a usable client instead of failing", func() {
				gomega.Expect(client).NotTo(gomega.BeNil())
			})

			ginkgo.It("should report the credential failure when sending a push notification", func() {
				err := client.SendPushNotification(ctx, notification.PushNotificationRequest{
					Token: "device-token",
					Title: "title",
					Body:  "body",
				})

				gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("missing_service_account.json")))
			})

			ginkgo.It("should report the credential failure when sending an email", func() {
				err := client.SendEmail(ctx, notification.EmailRequest{
					To:      "someone@example.com",
					Subject: "subject",
					Body:    "body",
				})

				gomega.Expect(err).To(gomega.HaveOccurred())
			})
		})

		ginkgo.When("the service account file is valid", func() {
			ginkgo.BeforeEach(func() {
				config = notification.FCMConfig{
					ProjectID:          "ci-dummy-project",
					ServiceAccountPath: filepath.Join("testdata", "fcm_service_account_dummy.json"),
				}
			})

			ginkgo.It("should return a real FCM client", func() {
				gomega.Expect(client).To(gomega.BeAssignableToTypeOf(&notification.FCMClient{}))
			})
		})
	})
})

package usecases_test

import (
	"context"
	"errors"

	"zensor-server/internal/infra/notification"
	"zensor-server/internal/shared_kernel/domain"
	"zensor-server/internal/shared_kernel/usecases"
	mocknotification "zensor-server/test/unit/doubles/infra/notification"
	mockusecases "zensor-server/test/unit/doubles/shared_kernel/usecases"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("SimpleUserPushMessageSender", func() {
	Context("SendBroadcastToUser", func() {
		var (
			ctrl         *gomock.Controller
			tokenService *mockusecases.MockPushTokenService
			notifier     *mocknotification.MockNotificationClient
			sender       *usecases.SimpleUserPushMessageSender
			uid          domain.ID
			content      usecases.UserPushBroadcastContent
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			tokenService = mockusecases.NewMockPushTokenService(ctrl)
			notifier = mocknotification.NewMockNotificationClient(ctrl)
			sender = usecases.NewUserPushMessageSender(tokenService, notifier)
			uid = domain.ID("user-1")
			content = usecases.UserPushBroadcastContent{
				Title:    "Hello",
				Body:     "World",
				DeepLink: "zensor://x",
			}
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		When("body is empty", func() {
			BeforeEach(func() {
				content.Body = "   "
			})

			It("should return ErrUserPushBroadcastBodyRequired", func() {
				err := sender.SendBroadcastToUser(context.Background(), uid, content)
				Expect(err).To(Equal(usecases.ErrUserPushBroadcastBodyRequired))
			})
		})

		When("user has no tokens", func() {
			It("should return ErrPushTokenNotFound", func() {
				tokenService.EXPECT().
					ListTokensByUserID(gomock.Any(), uid).
					Return(nil, usecases.ErrPushTokenNotFound)

				err := sender.SendBroadcastToUser(context.Background(), uid, content)
				Expect(err).To(Equal(usecases.ErrPushTokenNotFound))
			})
		})

		When("one token and send succeeds", func() {
			It("should send one notification with expected fields", func() {
				tokens := []domain.PushToken{
					{ID: "tok-1", UserID: uid, Token: "fcm-1", Platform: "android"},
				}
				tokenService.EXPECT().
					ListTokensByUserID(gomock.Any(), uid).
					Return(tokens, nil)

				notifier.EXPECT().
					SendPushNotification(gomock.Any(), notification.PushNotificationRequest{
						Token:    "fcm-1",
						Title:    "Hello",
						Body:     "World",
						DeepLink: "zensor://x",
					}).
					Return(nil)

				err := sender.SendBroadcastToUser(context.Background(), uid, content)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("two tokens and one send fails", func() {
			It("should still succeed", func() {
				tokens := []domain.PushToken{
					{ID: "a", UserID: uid, Token: "fcm-a", Platform: "ios"},
					{ID: "b", UserID: uid, Token: "fcm-b", Platform: "ios"},
				}
				tokenService.EXPECT().
					ListTokensByUserID(gomock.Any(), uid).
					Return(tokens, nil)

				notifier.EXPECT().
					SendPushNotification(gomock.Any(), gomock.Any()).
					Return(errors.New("fcm error"))
				notifier.EXPECT().
					SendPushNotification(gomock.Any(), gomock.Any()).
					Return(nil)

				err := sender.SendBroadcastToUser(context.Background(), uid, content)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("all sends fail", func() {
			It("should return an error", func() {
				tokens := []domain.PushToken{
					{ID: "a", UserID: uid, Token: "fcm-a", Platform: "ios"},
				}
				tokenService.EXPECT().
					ListTokensByUserID(gomock.Any(), uid).
					Return(tokens, nil)

				sendErr := errors.New("unavailable")
				notifier.EXPECT().
					SendPushNotification(gomock.Any(), gomock.Any()).
					Return(sendErr)

				err := sender.SendBroadcastToUser(context.Background(), uid, content)
				Expect(err).To(HaveOccurred())
				Expect(errors.Is(err, sendErr)).To(BeTrue())
			})
		})
	})
})

package usecases_test

import (
	"context"

	"zensor-server/internal/shared_kernel/domain"
	"zensor-server/internal/shared_kernel/usecases"
	mockusecases "zensor-server/test/unit/doubles/shared_kernel/usecases"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("SimplePushTokenService", func() {
	Context("ListTokensByUserID", func() {
		var (
			ctrl *gomock.Controller
			repo *mockusecases.MockPushTokenRepository
			svc  *usecases.SimplePushTokenService
			uid  domain.ID
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			repo = mockusecases.NewMockPushTokenRepository(ctrl)
			svc = usecases.NewPushTokenService(repo)
			uid = domain.ID("user-1")
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		When("user ID is empty", func() {
			It("should return an error", func() {
				_, err := svc.ListTokensByUserID(context.Background(), "")
				Expect(err).To(HaveOccurred())
			})
		})

		When("repository returns no tokens", func() {
			It("should return ErrPushTokenNotFound", func() {
				repo.EXPECT().ListByUserID(gomock.Any(), uid).Return(nil, nil)
				_, err := svc.ListTokensByUserID(context.Background(), uid)
				Expect(err).To(Equal(usecases.ErrPushTokenNotFound))
			})
		})

		When("repository returns tokens", func() {
			It("should return them", func() {
				tokens := []domain.PushToken{
					{ID: "id-1", UserID: uid, Token: "t1", Platform: "ios"},
				}
				repo.EXPECT().ListByUserID(gomock.Any(), uid).Return(tokens, nil)
				got, err := svc.ListTokensByUserID(context.Background(), uid)
				Expect(err).NotTo(HaveOccurred())
				Expect(got).To(Equal(tokens))
			})
		})
	})
})

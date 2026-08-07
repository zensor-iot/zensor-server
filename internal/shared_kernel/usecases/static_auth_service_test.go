package usecases_test

import (
	"context"
	"time"

	"zensor-server/internal/shared_kernel/domain"
	"zensor-server/internal/shared_kernel/usecases"
	mockusecases "zensor-server/test/unit/doubles/shared_kernel/usecases"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = ginkgo.Describe("SimpleStaticAuthService", func() {
	var (
		ctrl     *gomock.Controller
		sessions *mockusecases.MockSessionStore
		svc      *usecases.SimpleStaticAuthService
		ctx      context.Context
	)

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		sessions = mockusecases.NewMockSessionStore(ctrl)
		svc = usecases.NewStaticAuthService(sessions, time.Hour, "admin", "secret")
		ctx = context.Background()
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Context("Login", func() {
		ginkgo.When("the credentials match the configured admin", func() {
			ginkgo.It("should create an admin session", func() {
				sessions.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, session domain.Session) error {
					gomega.Expect(session.IsAdmin).To(gomega.BeTrue())
					gomega.Expect(session.ID).NotTo(gomega.BeEmpty())
					return nil
				})

				session, err := svc.Login(ctx, "admin", "secret")

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(session.IsAdmin).To(gomega.BeTrue())
			})
		})

		ginkgo.When("the username does not match", func() {
			ginkgo.It("should return ErrInvalidCredentials", func() {
				_, err := svc.Login(ctx, "someone-else", "secret")

				gomega.Expect(err).To(gomega.MatchError(usecases.ErrInvalidCredentials))
			})
		})

		ginkgo.When("the password does not match", func() {
			ginkgo.It("should return ErrInvalidCredentials", func() {
				_, err := svc.Login(ctx, "admin", "wrong")

				gomega.Expect(err).To(gomega.MatchError(usecases.ErrInvalidCredentials))
			})
		})
	})

	ginkgo.Context("GetSession", func() {
		ginkgo.When("the session store returns a session", func() {
			ginkgo.It("should return it", func() {
				expected := domain.Session{ID: "session-1", IsAdmin: true}
				sessions.EXPECT().Get(ctx, "session-1").Return(expected, nil)

				session, err := svc.GetSession(ctx, "session-1")

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(session).To(gomega.Equal(expected))
			})
		})

		ginkgo.When("the session store returns ErrSessionNotFound", func() {
			ginkgo.It("should propagate it", func() {
				sessions.EXPECT().Get(ctx, "missing").Return(domain.Session{}, usecases.ErrSessionNotFound)

				_, err := svc.GetSession(ctx, "missing")

				gomega.Expect(err).To(gomega.MatchError(usecases.ErrSessionNotFound))
			})
		})
	})

	ginkgo.Context("Logout", func() {
		ginkgo.When("logging out", func() {
			ginkgo.It("should delete the session", func() {
				sessions.EXPECT().Delete(ctx, "session-1").Return(nil)

				gomega.Expect(svc.Logout(ctx, "session-1")).To(gomega.Succeed())
			})
		})
	})
})

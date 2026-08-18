package usecases_test

import (
	"context"
	"errors"
	"time"

	"zensor-server/internal/shared_kernel/domain"
	"zensor-server/internal/shared_kernel/usecases"

	mockusecases "zensor-server/test/unit/doubles/shared_kernel/usecases"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = ginkgo.Describe("SimpleAuthService", func() {
	var (
		ctrl     *gomock.Controller
		repo     *mockusecases.MockAllowedUserRepository
		sessions *mockusecases.MockSessionStore
		provider *mockusecases.MockOAuthProvider
		svc      *usecases.SimpleAuthService
		ctx      context.Context
	)

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		repo = mockusecases.NewMockAllowedUserRepository(ctrl)
		sessions = mockusecases.NewMockSessionStore(ctrl)
		provider = mockusecases.NewMockOAuthProvider(ctrl)
		svc = usecases.NewAuthService(repo, sessions, provider, time.Hour)
		ctx = context.Background()
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Context("AuthCodeURL", func() {
		ginkgo.When("requesting the auth URL", func() {
			ginkgo.It("should delegate to the provider", func() {
				provider.EXPECT().AuthCodeURL("state-123").Return("https://accounts.google.com/?state=state-123")

				gomega.Expect(svc.AuthCodeURL("state-123")).To(gomega.Equal("https://accounts.google.com/?state=state-123"))
			})
		})
	})

	ginkgo.Context("HandleCallback", func() {
		var identity usecases.OAuthIdentity

		ginkgo.When("the email is on the allowlist", func() {
			var allowedUser domain.AllowedUser

			ginkgo.BeforeEach(func() {
				identity = usecases.OAuthIdentity{Email: "User@Example.com", Name: "User Name", EmailVerified: true}
				allowedUser = domain.AllowedUser{ID: "user-1", Email: "user@example.com", IsAdmin: true}

				provider.EXPECT().ExchangeCode(gomock.Any(), "code-1").Return(identity, nil)
				repo.EXPECT().GetByEmail(gomock.Any(), "user@example.com").Return(allowedUser, nil)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
			})

			ginkgo.It("should create a session with the allowlist identity", func() {
				var created domain.Session
				sessions.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, s domain.Session) error {
						created = s
						return nil
					})

				session, err := svc.HandleCallback(ctx, "code-1")

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(session.ID).To(gomega.HaveLen(64))
				gomega.Expect(session.UserID).To(gomega.Equal(domain.ID("user-1")))
				gomega.Expect(session.Email).To(gomega.Equal("user@example.com"))
				gomega.Expect(session.Name).To(gomega.Equal("User Name"))
				gomega.Expect(session.IsAdmin).To(gomega.BeTrue())
				gomega.Expect(session.ExpiresAt).To(gomega.BeTemporally("~", time.Now().Add(time.Hour), time.Minute))
				gomega.Expect(created).To(gomega.Equal(session))
			})
		})

		ginkgo.When("the email is not on the allowlist", func() {
			ginkgo.BeforeEach(func() {
				identity = usecases.OAuthIdentity{Email: "intruder@example.com", EmailVerified: true}
				provider.EXPECT().ExchangeCode(gomock.Any(), "code-1").Return(identity, nil)
				repo.EXPECT().GetByEmail(gomock.Any(), "intruder@example.com").Return(domain.AllowedUser{}, usecases.ErrAllowedUserNotFound)
			})

			ginkgo.It("should return ErrEmailNotAllowed and create no session", func() {
				_, err := svc.HandleCallback(ctx, "code-1")

				gomega.Expect(err).To(gomega.MatchError(usecases.ErrEmailNotAllowed))
			})
		})

		ginkgo.When("the email is not verified", func() {
			ginkgo.BeforeEach(func() {
				identity = usecases.OAuthIdentity{Email: "user@example.com", EmailVerified: false}
				provider.EXPECT().ExchangeCode(gomock.Any(), "code-1").Return(identity, nil)
			})

			ginkgo.It("should return ErrEmailNotVerified", func() {
				_, err := svc.HandleCallback(ctx, "code-1")

				gomega.Expect(err).To(gomega.MatchError(usecases.ErrEmailNotVerified))
			})
		})

		ginkgo.When("the code exchange fails", func() {
			ginkgo.BeforeEach(func() {
				provider.EXPECT().ExchangeCode(gomock.Any(), "bad-code").Return(usecases.OAuthIdentity{}, errors.New("exchange failed"))
			})

			ginkgo.It("should return an error", func() {
				_, err := svc.HandleCallback(ctx, "bad-code")

				gomega.Expect(err).To(gomega.HaveOccurred())
			})
		})
	})

	ginkgo.Context("GetSession", func() {
		ginkgo.When("the session exists", func() {
			ginkgo.It("should return it", func() {
				stored := domain.Session{ID: "sess-1", UserID: "user-1"}
				sessions.EXPECT().Get(gomock.Any(), "sess-1").Return(stored, nil)

				session, err := svc.GetSession(ctx, "sess-1")

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(session).To(gomega.Equal(stored))
			})
		})

		ginkgo.When("the session does not exist", func() {
			ginkgo.It("should return ErrSessionNotFound", func() {
				sessions.EXPECT().Get(gomock.Any(), "missing").Return(domain.Session{}, usecases.ErrSessionNotFound)

				_, err := svc.GetSession(ctx, "missing")

				gomega.Expect(err).To(gomega.MatchError(usecases.ErrSessionNotFound))
			})
		})
	})

	ginkgo.Context("Logout", func() {
		ginkgo.When("logging out", func() {
			ginkgo.It("should delete the session", func() {
				sessions.EXPECT().Delete(gomock.Any(), "sess-1").Return(nil)

				gomega.Expect(svc.Logout(ctx, "sess-1")).To(gomega.Succeed())
			})
		})
	})

	ginkgo.Context("AddAllowedUser", func() {
		ginkgo.When("the email is new", func() {
			ginkgo.It("should create a normalized allowlist entry", func() {
				repo.EXPECT().GetByEmail(gomock.Any(), "new@example.com").Return(domain.AllowedUser{}, usecases.ErrAllowedUserNotFound)
				var created domain.AllowedUser
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, u domain.AllowedUser) error {
						created = u
						return nil
					})

				user, err := svc.AddAllowedUser(ctx, "New@Example.COM", true)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(user.Email).To(gomega.Equal("new@example.com"))
				gomega.Expect(user.IsAdmin).To(gomega.BeTrue())
				gomega.Expect(created).To(gomega.Equal(user))
			})
		})

		ginkgo.When("the email already exists", func() {
			ginkgo.It("should return ErrAllowedUserDuplicated", func() {
				repo.EXPECT().GetByEmail(gomock.Any(), "dup@example.com").Return(domain.AllowedUser{ID: "u1"}, nil)

				_, err := svc.AddAllowedUser(ctx, "dup@example.com", false)

				gomega.Expect(err).To(gomega.MatchError(usecases.ErrAllowedUserDuplicated))
			})
		})

		ginkgo.When("the email is invalid", func() {
			ginkgo.It("should return an error", func() {
				_, err := svc.AddAllowedUser(ctx, "not-an-email", false)

				gomega.Expect(err).To(gomega.HaveOccurred())
			})
		})
	})

	ginkgo.Context("UpdateAllowedUser", func() {
		ginkgo.When("the user exists", func() {
			ginkgo.It("should toggle the admin flag", func() {
				existing := domain.AllowedUser{ID: "user-1", Email: "user@example.com", IsAdmin: false}
				repo.EXPECT().GetByID(gomock.Any(), domain.ID("user-1")).Return(existing, nil)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, u domain.AllowedUser) error {
						gomega.Expect(u.IsAdmin).To(gomega.BeTrue())
						return nil
					})

				updated, err := svc.UpdateAllowedUser(ctx, "user-1", true)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(updated.IsAdmin).To(gomega.BeTrue())
			})
		})

		ginkgo.When("the user does not exist", func() {
			ginkgo.It("should return ErrAllowedUserNotFound", func() {
				repo.EXPECT().GetByID(gomock.Any(), domain.ID("missing")).Return(domain.AllowedUser{}, usecases.ErrAllowedUserNotFound)

				_, err := svc.UpdateAllowedUser(ctx, "missing", true)

				gomega.Expect(err).To(gomega.MatchError(usecases.ErrAllowedUserNotFound))
			})
		})
	})

	ginkgo.Context("RemoveAllowedUser", func() {
		ginkgo.When("removing a user", func() {
			ginkgo.It("should delete the user and revoke their sessions", func() {
				repo.EXPECT().Delete(gomock.Any(), domain.ID("user-1")).Return(nil)
				sessions.EXPECT().DeleteByUser(gomock.Any(), domain.ID("user-1")).Return(nil)

				gomega.Expect(svc.RemoveAllowedUser(ctx, "user-1")).To(gomega.Succeed())
			})
		})
	})

	ginkgo.Context("ListAllowedUsers", func() {
		ginkgo.When("listing users", func() {
			ginkgo.It("should return the repository result", func() {
				users := []domain.AllowedUser{{ID: "u1"}, {ID: "u2"}}
				repo.EXPECT().FindAll(gomock.Any()).Return(users, nil)

				result, err := svc.ListAllowedUsers(ctx)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).To(gomega.Equal(users))
			})
		})
	})

	ginkgo.Context("BootstrapAdmin", func() {
		ginkgo.When("the email is not on the allowlist", func() {
			ginkgo.It("should create it as admin", func() {
				repo.EXPECT().GetByEmail(gomock.Any(), "boss@example.com").Return(domain.AllowedUser{}, usecases.ErrAllowedUserNotFound)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, u domain.AllowedUser) error {
						gomega.Expect(u.Email).To(gomega.Equal("boss@example.com"))
						gomega.Expect(u.IsAdmin).To(gomega.BeTrue())
						return nil
					})

				gomega.Expect(svc.BootstrapAdmin(ctx, "Boss@Example.com")).To(gomega.Succeed())
			})
		})

		ginkgo.When("the email exists without admin", func() {
			ginkgo.It("should promote it to admin", func() {
				existing := domain.AllowedUser{ID: "u1", Email: "boss@example.com", IsAdmin: false}
				repo.EXPECT().GetByEmail(gomock.Any(), "boss@example.com").Return(existing, nil)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, u domain.AllowedUser) error {
						gomega.Expect(u.IsAdmin).To(gomega.BeTrue())
						return nil
					})

				gomega.Expect(svc.BootstrapAdmin(ctx, "boss@example.com")).To(gomega.Succeed())
			})
		})

		ginkgo.When("the email already is an admin", func() {
			ginkgo.It("should do nothing", func() {
				existing := domain.AllowedUser{ID: "u1", Email: "boss@example.com", IsAdmin: true}
				repo.EXPECT().GetByEmail(gomock.Any(), "boss@example.com").Return(existing, nil)

				gomega.Expect(svc.BootstrapAdmin(ctx, "boss@example.com")).To(gomega.Succeed())
			})
		})

		ginkgo.When("the email is empty", func() {
			ginkgo.It("should do nothing", func() {
				gomega.Expect(svc.BootstrapAdmin(ctx, "")).To(gomega.Succeed())
			})
		})
	})
})

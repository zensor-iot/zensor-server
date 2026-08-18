package auth_test

import (
	"context"
	"time"

	"zensor-server/internal/infra/auth"
	"zensor-server/internal/shared_kernel/domain"
	"zensor-server/internal/shared_kernel/usecases"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("MemorySessionStore", func() {
	var (
		store *auth.MemorySessionStore
		ctx   context.Context
	)

	ginkgo.BeforeEach(func() {
		store = auth.NewMemorySessionStore()
		ctx = context.Background()
	})

	newSession := func(id string, userID domain.ID, expiresAt time.Time) domain.Session {
		return domain.Session{
			ID:        id,
			UserID:    userID,
			Email:     "user@example.com",
			CreatedAt: time.Now(),
			ExpiresAt: expiresAt,
		}
	}

	ginkgo.Context("Create and Get", func() {
		ginkgo.When("the session exists and is not expired", func() {
			ginkgo.It("should return the stored session", func() {
				session := newSession("sess-1", "user-1", time.Now().Add(time.Hour))
				gomega.Expect(store.Create(ctx, session)).To(gomega.Succeed())

				result, err := store.Get(ctx, "sess-1")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).To(gomega.Equal(session))
			})
		})

		ginkgo.When("the session does not exist", func() {
			ginkgo.It("should return ErrSessionNotFound", func() {
				_, err := store.Get(ctx, "missing")
				gomega.Expect(err).To(gomega.MatchError(usecases.ErrSessionNotFound))
			})
		})

		ginkgo.When("the session is expired", func() {
			ginkgo.It("should return ErrSessionNotFound", func() {
				session := newSession("sess-1", "user-1", time.Now().Add(-time.Minute))
				gomega.Expect(store.Create(ctx, session)).To(gomega.Succeed())

				_, err := store.Get(ctx, "sess-1")
				gomega.Expect(err).To(gomega.MatchError(usecases.ErrSessionNotFound))
			})
		})
	})

	ginkgo.Context("Delete", func() {
		ginkgo.When("deleting a session", func() {
			ginkgo.It("should no longer be retrievable", func() {
				session := newSession("sess-1", "user-1", time.Now().Add(time.Hour))
				gomega.Expect(store.Create(ctx, session)).To(gomega.Succeed())

				gomega.Expect(store.Delete(ctx, "sess-1")).To(gomega.Succeed())

				_, err := store.Get(ctx, "sess-1")
				gomega.Expect(err).To(gomega.MatchError(usecases.ErrSessionNotFound))
			})
		})
	})

	ginkgo.Context("DeleteByUser", func() {
		ginkgo.When("a user has multiple sessions", func() {
			ginkgo.It("should remove all of them and keep other users' sessions", func() {
				gomega.Expect(store.Create(ctx, newSession("sess-1", "user-1", time.Now().Add(time.Hour)))).To(gomega.Succeed())
				gomega.Expect(store.Create(ctx, newSession("sess-2", "user-1", time.Now().Add(time.Hour)))).To(gomega.Succeed())
				gomega.Expect(store.Create(ctx, newSession("sess-3", "user-2", time.Now().Add(time.Hour)))).To(gomega.Succeed())

				gomega.Expect(store.DeleteByUser(ctx, "user-1")).To(gomega.Succeed())

				_, err := store.Get(ctx, "sess-1")
				gomega.Expect(err).To(gomega.MatchError(usecases.ErrSessionNotFound))
				_, err = store.Get(ctx, "sess-2")
				gomega.Expect(err).To(gomega.MatchError(usecases.ErrSessionNotFound))

				result, err := store.Get(ctx, "sess-3")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result.UserID).To(gomega.Equal(domain.ID("user-2")))
			})
		})
	})
})

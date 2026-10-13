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

var _ = ginkgo.Describe("MemoryAuthHandoffStore", func() {
	var (
		store *auth.MemoryAuthHandoffStore
		ctx   context.Context
	)

	ginkgo.BeforeEach(func() {
		store = auth.NewMemoryAuthHandoffStore()
		ctx = context.Background()
	})

	ginkgo.Context("Consume", func() {
		ginkgo.When("a handoff was saved", func() {
			ginkgo.It("should return it once and then report it as not found", func() {
				handoff := domain.AuthHandoff{UserID: "user-1", ExpiresAt: time.Now().Add(time.Minute)}
				gomega.Expect(store.Save(ctx, "hash-1", handoff)).To(gomega.Succeed())

				got, err := store.Consume(ctx, "hash-1")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(got).To(gomega.Equal(handoff))

				_, err = store.Consume(ctx, "hash-1")
				gomega.Expect(err).To(gomega.MatchError(usecases.ErrAuthHandoffNotFound))
			})
		})

		ginkgo.When("nothing was saved under the hash", func() {
			ginkgo.It("should return ErrAuthHandoffNotFound", func() {
				_, err := store.Consume(ctx, "missing")

				gomega.Expect(err).To(gomega.MatchError(usecases.ErrAuthHandoffNotFound))
			})
		})
	})
})

package persistence_test

import (
	"context"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/shared_kernel/domain"
	sharedPersistence "zensor-server/internal/shared_kernel/persistence"
	sharedUsecases "zensor-server/internal/shared_kernel/usecases"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("AllowedUserRepository", func() {
	var (
		orm  sql.ORM
		repo sharedUsecases.AllowedUserRepository
		ctx  context.Context
	)

	ginkgo.BeforeEach(func() {
		var err error
		orm, err = sql.NewMemoryORM("migrations")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		repo, err = sharedPersistence.NewAllowedUserRepository(orm)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ctx = context.Background()
	})

	newUser := func(email string, isAdmin bool) domain.AllowedUser {
		user, err := domain.NewAllowedUserBuilder().
			WithEmail(email).
			WithIsAdmin(isAdmin).
			Build()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return user
	}

	ginkgo.Context("Create and GetByEmail", func() {
		ginkgo.When("the user exists", func() {
			ginkgo.It("should be retrievable by its lowercase email", func() {
				user := newUser("someone@example.com", true)
				gomega.Expect(repo.Create(ctx, user)).To(gomega.Succeed())

				result, err := repo.GetByEmail(ctx, "someone@example.com")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result.ID).To(gomega.Equal(user.ID))
				gomega.Expect(result.Email).To(gomega.Equal("someone@example.com"))
				gomega.Expect(result.IsAdmin).To(gomega.BeTrue())
				gomega.Expect(result.LastLoginAt).To(gomega.BeNil())
			})
		})

		ginkgo.When("the email does not exist", func() {
			ginkgo.It("should return ErrAllowedUserNotFound", func() {
				_, err := repo.GetByEmail(ctx, "missing@example.com")
				gomega.Expect(err).To(gomega.MatchError(sharedUsecases.ErrAllowedUserNotFound))
			})
		})
	})

	ginkgo.Context("GetByID", func() {
		ginkgo.When("the user does not exist", func() {
			ginkgo.It("should return ErrAllowedUserNotFound", func() {
				_, err := repo.GetByID(ctx, domain.ID("missing"))
				gomega.Expect(err).To(gomega.MatchError(sharedUsecases.ErrAllowedUserNotFound))
			})
		})
	})

	ginkgo.Context("FindAll", func() {
		ginkgo.When("multiple users exist", func() {
			ginkgo.It("should return all of them", func() {
				userA := newUser("a@example.com", false)
				userB := newUser("b@example.com", true)
				gomega.Expect(repo.Create(ctx, userA)).To(gomega.Succeed())
				gomega.Expect(repo.Create(ctx, userB)).To(gomega.Succeed())

				users, err := repo.FindAll(ctx)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				emails := make([]string, 0, len(users))
				for _, user := range users {
					emails = append(emails, user.Email)
				}
				gomega.Expect(emails).To(gomega.ContainElements("a@example.com", "b@example.com"))
			})
		})
	})

	ginkgo.Context("Update", func() {
		ginkgo.When("recording a login", func() {
			ginkgo.It("should persist display name and last login", func() {
				user := newUser("login@example.com", false)
				gomega.Expect(repo.Create(ctx, user)).To(gomega.Succeed())

				user.RecordLogin("Login User")
				gomega.Expect(repo.Update(ctx, user)).To(gomega.Succeed())

				result, err := repo.GetByEmail(ctx, "login@example.com")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result.DisplayName).To(gomega.Equal("Login User"))
				gomega.Expect(result.LastLoginAt).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Context("Delete", func() {
		ginkgo.When("the user exists", func() {
			ginkgo.It("should remove it", func() {
				user := newUser("gone@example.com", false)
				gomega.Expect(repo.Create(ctx, user)).To(gomega.Succeed())

				gomega.Expect(repo.Delete(ctx, user.ID)).To(gomega.Succeed())

				_, err := repo.GetByEmail(ctx, "gone@example.com")
				gomega.Expect(err).To(gomega.MatchError(sharedUsecases.ErrAllowedUserNotFound))
			})
		})
	})
})

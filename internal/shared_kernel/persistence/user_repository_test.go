package persistence_test

import (
	"context"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/infra/utils"
	sharedPersistence "zensor-server/internal/shared_kernel/persistence"
	"zensor-server/internal/shared_kernel/domain"
	sharedUsecases "zensor-server/internal/shared_kernel/usecases"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("UserRepository", func() {
	var (
		orm  sql.ORM
		repo sharedUsecases.UserRepository
		ctx  context.Context
	)

	ginkgo.BeforeEach(func() {
		var err error
		orm, err = sql.NewMemoryORM("migrations")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		repo, err = sharedPersistence.NewUserRepository(orm)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(repo).NotTo(gomega.BeNil())

		ctx = context.Background()
	})

	ginkgo.Context("Upsert", func() {
		var user domain.User

		ginkgo.BeforeEach(func() {
			user = domain.User{
				ID:      domain.ID(utils.GenerateUUID()),
				Tenants: []domain.ID{"tenant-a"},
			}
		})

		ginkgo.When("the user does not exist", func() {
			ginkgo.It("should create the user in the database", func() {
				err := repo.Upsert(ctx, user)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				result, err := repo.GetByID(ctx, user.ID)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result.ID).To(gomega.Equal(user.ID))
				gomega.Expect(result.Tenants).To(gomega.Equal(user.Tenants))
			})
		})

		ginkgo.When("the user already exists", func() {
			ginkgo.BeforeEach(func() {
				err := repo.Upsert(ctx, user)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should update the existing user's tenants", func() {
				user.Tenants = []domain.ID{"tenant-a", "tenant-b"}
				err := repo.Upsert(ctx, user)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				result, err := repo.GetByID(ctx, user.ID)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result.Tenants).To(gomega.Equal(user.Tenants))
			})
		})
	})

	ginkgo.Context("GetByID", func() {
		ginkgo.When("user does not exist", func() {
			ginkgo.It("should return ErrUserNotFound", func() {
				_, err := repo.GetByID(ctx, domain.ID("non-existent"))
				gomega.Expect(err).To(gomega.MatchError(sharedUsecases.ErrUserNotFound))
			})
		})
	})

	ginkgo.Context("FindByTenant", func() {
		ginkgo.When("users are associated with the tenant", func() {
			ginkgo.It("should return only those users", func() {
				tenantA := domain.ID(utils.GenerateUUID())
				tenantB := domain.ID(utils.GenerateUUID())

				userA := domain.User{ID: domain.ID(utils.GenerateUUID()), Tenants: []domain.ID{tenantA}}
				userB := domain.User{ID: domain.ID(utils.GenerateUUID()), Tenants: []domain.ID{tenantA, tenantB}}
				userC := domain.User{ID: domain.ID(utils.GenerateUUID()), Tenants: []domain.ID{tenantB}}

				gomega.Expect(repo.Upsert(ctx, userA)).NotTo(gomega.HaveOccurred())
				gomega.Expect(repo.Upsert(ctx, userB)).NotTo(gomega.HaveOccurred())
				gomega.Expect(repo.Upsert(ctx, userC)).NotTo(gomega.HaveOccurred())

				users, err := repo.FindByTenant(ctx, tenantA)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(users).To(gomega.HaveLen(2))

				ids := make([]domain.ID, len(users))
				for i, u := range users {
					ids[i] = u.ID
				}
				gomega.Expect(ids).To(gomega.ConsistOf(userA.ID, userB.ID))
			})
		})

		ginkgo.When("no user is associated with the tenant", func() {
			ginkgo.It("should return an empty list", func() {
				tenantID := domain.ID(utils.GenerateUUID())
				users, err := repo.FindByTenant(ctx, tenantID)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(users).To(gomega.BeEmpty())
			})
		})
	})
})

package persistence_test

import (
	"context"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"

	sharedPersistence "zensor-server/internal/shared_kernel/persistence"
	sharedUsecases "zensor-server/internal/shared_kernel/usecases"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("TenantRepository", func() {
	var (
		orm  sql.ORM
		repo sharedUsecases.TenantRepository
		ctx  context.Context
	)

	ginkgo.BeforeEach(func() {
		var err error
		orm, err = sql.NewMemoryORM("migrations")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		repo, err = sharedPersistence.NewTenantRepository(orm)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(repo).NotTo(gomega.BeNil())

		ctx = context.Background()
	})

	ginkgo.Context("Create", func() {
		var tenant domain.Tenant

		ginkgo.BeforeEach(func() {
			id := utils.GenerateUUID()
			tenant = domain.Tenant{
				ID:      domain.ID(id),
				Name:    "acme-" + id,
				Email:   "acme@example.com",
				Version: 1,
			}
		})

		ginkgo.It("should create the tenant in the database", func() {
			err := repo.Create(ctx, tenant)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			result, err := repo.GetByID(ctx, tenant.ID)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(result.Name).To(gomega.Equal(tenant.Name))
			gomega.Expect(result.Email).To(gomega.Equal(tenant.Email))
		})
	})

	ginkgo.Context("Update", func() {
		var tenant domain.Tenant

		ginkgo.BeforeEach(func() {
			id := utils.GenerateUUID()
			tenant = domain.Tenant{
				ID:      domain.ID(id),
				Name:    "acme-" + id,
				Email:   "acme@example.com",
				Version: 1,
			}
			err := repo.Create(ctx, tenant)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.It("should update the tenant in the database", func() {
			updatedName := tenant.Name + "-updated"
			tenant.Name = updatedName
			err := repo.Update(ctx, tenant)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			result, err := repo.GetByID(ctx, tenant.ID)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(result.Name).To(gomega.Equal(updatedName))
		})
	})

	ginkgo.Context("GetByID", func() {
		ginkgo.When("tenant does not exist", func() {
			ginkgo.It("should return ErrTenantNotFound", func() {
				_, err := repo.GetByID(ctx, domain.ID("non-existent"))
				gomega.Expect(err).To(gomega.MatchError(sharedUsecases.ErrTenantNotFound))
			})
		})
	})
})

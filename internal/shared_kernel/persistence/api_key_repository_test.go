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

var _ = ginkgo.Describe("APIKeyRepository", func() {
	var (
		orm  sql.ORM
		repo sharedUsecases.APIKeyRepository
		ctx  context.Context
	)

	ginkgo.BeforeEach(func() {
		var err error
		orm, err = sql.NewMemoryORM("migrations")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		repo, err = sharedPersistence.NewAPIKeyRepository(orm)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ctx = context.Background()
	})

	newKey := func(name string) domain.APIKey {
		key, _, err := domain.NewAPIKeyBuilder().
			WithName(name).
			WithCreatedBy(domain.ID("admin-1")).
			Build()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return key
	}

	ginkgo.Context("Create and GetByHash", func() {
		ginkgo.When("the key exists", func() {
			ginkgo.It("should be retrievable by its hash", func() {
				key := newKey("hash-lookup")
				gomega.Expect(repo.Create(ctx, key)).To(gomega.Succeed())

				result, err := repo.GetByHash(ctx, key.KeyHash)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result.ID).To(gomega.Equal(key.ID))
				gomega.Expect(result.Name).To(gomega.Equal("hash-lookup"))
				gomega.Expect(result.KeyPrefix).To(gomega.Equal(key.KeyPrefix))
				gomega.Expect(result.CreatedBy).To(gomega.Equal(key.CreatedBy))
			})
		})

		ginkgo.When("the hash does not exist", func() {
			ginkgo.It("should return ErrAPIKeyNotFound", func() {
				_, err := repo.GetByHash(ctx, "missing-hash")
				gomega.Expect(err).To(gomega.MatchError(sharedUsecases.ErrAPIKeyNotFound))
			})
		})
	})

	ginkgo.Context("GetByID", func() {
		ginkgo.When("the key exists", func() {
			ginkgo.It("should return it", func() {
				key := newKey("id-lookup")
				gomega.Expect(repo.Create(ctx, key)).To(gomega.Succeed())

				result, err := repo.GetByID(ctx, key.ID)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result.KeyHash).To(gomega.Equal(key.KeyHash))
			})
		})

		ginkgo.When("the key does not exist", func() {
			ginkgo.It("should return ErrAPIKeyNotFound", func() {
				_, err := repo.GetByID(ctx, domain.ID("missing"))
				gomega.Expect(err).To(gomega.MatchError(sharedUsecases.ErrAPIKeyNotFound))
			})
		})
	})

	ginkgo.Context("GetByName", func() {
		ginkgo.When("the key exists", func() {
			ginkgo.It("should return it", func() {
				key := newKey("name-lookup")
				gomega.Expect(repo.Create(ctx, key)).To(gomega.Succeed())

				result, err := repo.GetByName(ctx, "name-lookup")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result.ID).To(gomega.Equal(key.ID))
			})
		})

		ginkgo.When("the name does not exist", func() {
			ginkgo.It("should return ErrAPIKeyNotFound", func() {
				_, err := repo.GetByName(ctx, "missing-name")
				gomega.Expect(err).To(gomega.MatchError(sharedUsecases.ErrAPIKeyNotFound))
			})
		})
	})

	ginkgo.Context("FindAll", func() {
		ginkgo.When("multiple keys exist", func() {
			ginkgo.It("should return all of them", func() {
				keyA := newKey("all-a")
				keyB := newKey("all-b")
				gomega.Expect(repo.Create(ctx, keyA)).To(gomega.Succeed())
				gomega.Expect(repo.Create(ctx, keyB)).To(gomega.Succeed())

				keys, err := repo.FindAll(ctx)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				names := make([]string, 0, len(keys))
				for _, key := range keys {
					names = append(names, key.Name)
				}
				gomega.Expect(names).To(gomega.ContainElements("all-a", "all-b"))
			})
		})
	})

	ginkgo.Context("Delete", func() {
		ginkgo.When("the key exists", func() {
			ginkgo.It("should remove it", func() {
				key := newKey("deletable")
				gomega.Expect(repo.Create(ctx, key)).To(gomega.Succeed())

				gomega.Expect(repo.Delete(ctx, key.ID)).To(gomega.Succeed())

				_, err := repo.GetByHash(ctx, key.KeyHash)
				gomega.Expect(err).To(gomega.MatchError(sharedUsecases.ErrAPIKeyNotFound))
			})
		})
	})
})

package usecases_test

import (
	"context"
	"errors"
	"time"

	"zensor-server/internal/infra/cache"
	"zensor-server/internal/shared_kernel/domain"
	"zensor-server/internal/shared_kernel/usecases"
	mockusecases "zensor-server/test/unit/doubles/shared_kernel/usecases"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = ginkgo.Describe("SimpleAPIKeyService", func() {
	var (
		ctrl       *gomock.Controller
		repository *mockusecases.MockAPIKeyRepository
		keyCache   cache.Cache
		service    *usecases.SimpleAPIKeyService
		ctx        context.Context
	)

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		repository = mockusecases.NewMockAPIKeyRepository(ctrl)

		var err error
		keyCache, err = cache.New(cache.DefaultConfig())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		service = usecases.NewAPIKeyService(repository, keyCache)
		ctx = context.Background()
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Context("Create", func() {
		ginkgo.When("the name is available", func() {
			ginkgo.It("should persist the hashed key and return the plaintext once", func() {
				repository.EXPECT().GetByName(gomock.Any(), "grafana-sync").
					Return(domain.APIKey{}, usecases.ErrAPIKeyNotFound)

				var persisted domain.APIKey
				repository.EXPECT().Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, key domain.APIKey) error {
						persisted = key
						return nil
					})

				key, plaintext, err := service.Create(ctx, "grafana-sync", domain.ID("admin-1"))

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(plaintext).To(gomega.HavePrefix("zsk_"))
				gomega.Expect(key.KeyHash).To(gomega.Equal(domain.HashAPIKey(plaintext)))
				gomega.Expect(persisted.KeyHash).To(gomega.Equal(key.KeyHash))
				gomega.Expect(persisted.Name).To(gomega.Equal("grafana-sync"))
				gomega.Expect(persisted.CreatedBy).To(gomega.Equal(domain.ID("admin-1")))
			})
		})

		ginkgo.When("the name is already taken", func() {
			ginkgo.It("should return ErrAPIKeyDuplicated", func() {
				repository.EXPECT().GetByName(gomock.Any(), "grafana-sync").
					Return(domain.APIKey{Name: "grafana-sync"}, nil)

				_, _, err := service.Create(ctx, "grafana-sync", domain.ID("admin-1"))

				gomega.Expect(err).To(gomega.MatchError(usecases.ErrAPIKeyDuplicated))
			})
		})

		ginkgo.When("the name is blank", func() {
			ginkgo.It("should return a validation error without touching the repository", func() {
				_, _, err := service.Create(ctx, "  ", domain.ID("admin-1"))

				gomega.Expect(err).To(gomega.HaveOccurred())
			})
		})
	})

	ginkgo.Context("Validate", func() {
		var (
			plaintext string
			stored    domain.APIKey
		)

		ginkgo.BeforeEach(func() {
			var err error
			stored, plaintext, err = domain.NewAPIKeyBuilder().
				WithName("grafana-sync").
				WithCreatedBy(domain.ID("admin-1")).
				Build()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.When("the key is unknown to the cache", func() {
			ginkgo.It("should fall back to the repository and then serve from cache", func() {
				repository.EXPECT().GetByHash(gomock.Any(), stored.KeyHash).
					Return(stored, nil).
					Times(1)

				result, err := service.Validate(ctx, plaintext)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result.ID).To(gomega.Equal(stored.ID))

				gomega.Eventually(func() bool {
					_, found := keyCache.Get(ctx, "apikey:"+stored.KeyHash)
					return found
				}).Should(gomega.BeTrue())

				result, err = service.Validate(ctx, plaintext)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result.ID).To(gomega.Equal(stored.ID))
			})
		})

		ginkgo.When("the key does not exist", func() {
			ginkgo.It("should return ErrAPIKeyNotFound without negative caching", func() {
				repository.EXPECT().GetByHash(gomock.Any(), gomock.Any()).
					Return(domain.APIKey{}, usecases.ErrAPIKeyNotFound).
					Times(2)

				_, err := service.Validate(ctx, "zsk_unknown")
				gomega.Expect(err).To(gomega.MatchError(usecases.ErrAPIKeyNotFound))

				_, err = service.Validate(ctx, "zsk_unknown")
				gomega.Expect(err).To(gomega.MatchError(usecases.ErrAPIKeyNotFound))
			})
		})

		ginkgo.When("the repository fails", func() {
			ginkgo.It("should propagate the error", func() {
				repository.EXPECT().GetByHash(gomock.Any(), gomock.Any()).
					Return(domain.APIKey{}, errors.New("database down"))

				_, err := service.Validate(ctx, plaintext)
				gomega.Expect(err).To(gomega.HaveOccurred())
			})
		})
	})

	ginkgo.Context("List", func() {
		ginkgo.When("keys exist", func() {
			ginkgo.It("should return them", func() {
				keys := []domain.APIKey{{ID: domain.ID("key-1"), Name: "one"}}
				repository.EXPECT().FindAll(gomock.Any()).Return(keys, nil)

				result, err := service.List(ctx)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).To(gomega.Equal(keys))
			})
		})
	})

	ginkgo.Context("Revoke", func() {
		var (
			plaintext string
			stored    domain.APIKey
		)

		ginkgo.BeforeEach(func() {
			var err error
			stored, plaintext, err = domain.NewAPIKeyBuilder().
				WithName("grafana-sync").
				WithCreatedBy(domain.ID("admin-1")).
				Build()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.When("the key exists and is cached", func() {
			ginkgo.It("should delete the row and purge the cache entry", func() {
				keyCache.Set(ctx, "apikey:"+stored.KeyHash, stored, time.Hour)
				gomega.Eventually(func() bool {
					_, found := keyCache.Get(ctx, "apikey:"+stored.KeyHash)
					return found
				}).Should(gomega.BeTrue())

				repository.EXPECT().GetByID(gomock.Any(), stored.ID).Return(stored, nil)
				repository.EXPECT().Delete(gomock.Any(), stored.ID).Return(nil)

				gomega.Expect(service.Revoke(ctx, stored.ID)).To(gomega.Succeed())

				gomega.Eventually(func() bool {
					_, found := keyCache.Get(ctx, "apikey:"+stored.KeyHash)
					return found
				}).Should(gomega.BeFalse())

				repository.EXPECT().GetByHash(gomock.Any(), stored.KeyHash).
					Return(domain.APIKey{}, usecases.ErrAPIKeyNotFound)
				_, err := service.Validate(ctx, plaintext)
				gomega.Expect(err).To(gomega.MatchError(usecases.ErrAPIKeyNotFound))
			})
		})

		ginkgo.When("the key does not exist", func() {
			ginkgo.It("should return ErrAPIKeyNotFound", func() {
				repository.EXPECT().GetByID(gomock.Any(), domain.ID("missing")).
					Return(domain.APIKey{}, usecases.ErrAPIKeyNotFound)

				err := service.Revoke(ctx, domain.ID("missing"))

				gomega.Expect(err).To(gomega.MatchError(usecases.ErrAPIKeyNotFound))
			})
		})
	})
})

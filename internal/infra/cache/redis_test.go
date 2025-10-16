package cache_test

import (
	"context"
	"time"
	"zensor-server/internal/infra/cache"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("RedisCache", func() {
	var (
		redisCache *cache.RedisCache
		ctx        context.Context
	)

	ginkgo.BeforeEach(func() {
		config := &cache.RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}

		var err error
		redisCache, err = cache.NewRedisCache(config)
		if err != nil {
			ginkgo.Skip("Redis not available, skipping test")
		}

		ctx = context.Background()
	})

	ginkgo.Context("NewRedisCache", func() {
		var config *cache.RedisConfig

		ginkgo.When("creating a new Redis cache with custom config", func() {
			ginkgo.BeforeEach(func() {
				config = &cache.RedisConfig{
					Addr:     "localhost:6379",
					Password: "",
					DB:       0,
				}
			})

			ginkgo.It("should create a valid Redis cache instance", func() {
				cacheInstance, err := cache.NewRedisCache(config)
				if err != nil {
					ginkgo.Skip("Redis not available, skipping test")
				}

				gomega.Expect(cacheInstance).NotTo(gomega.BeNil())
			})
		})

		ginkgo.When("creating a new Redis cache with default config", func() {
			ginkgo.It("should create cache with default configuration", func() {
				cacheInstance, err := cache.NewRedisCache(nil)
				if err != nil {
					ginkgo.Skip("Redis not available, skipping test")
				}

				gomega.Expect(cacheInstance).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Context("SetAndGet", func() {
		var (
			key   string
			value string
		)

		ginkgo.When("setting and getting a value", func() {
			ginkgo.BeforeEach(func() {
				key = "test_key"
				value = "test_value"
			})

			ginkgo.It("should store and retrieve the value correctly", func() {
				success := redisCache.Set(ctx, key, value, 0)
				gomega.Expect(success).To(gomega.BeTrue())

				retrieved, found := redisCache.Get(ctx, key)
				gomega.Expect(found).To(gomega.BeTrue())
				gomega.Expect(retrieved).To(gomega.Equal(value))
			})
		})
	})

	ginkgo.Context("SetWithTTL", func() {
		var (
			key   string
			value string
			ttl   time.Duration
		)

		ginkgo.When("setting a value with TTL", func() {
			ginkgo.BeforeEach(func() {
				key = "test_ttl_key"
				value = "test_ttl_value"
				ttl = 1 * time.Second
			})

			ginkgo.It("should expire the value after TTL", func() {
				success := redisCache.Set(ctx, key, value, ttl)
				gomega.Expect(success).To(gomega.BeTrue())

				retrieved, found := redisCache.Get(ctx, key)
				gomega.Expect(found).To(gomega.BeTrue())
				gomega.Expect(retrieved).To(gomega.Equal(value))

				time.Sleep(2 * time.Second)

				retrieved, found = redisCache.Get(ctx, key)
				gomega.Expect(found).To(gomega.BeFalse())
			})
		})
	})

	ginkgo.Context("Delete", func() {
		var (
			key   string
			value string
		)

		ginkgo.When("deleting a value", func() {
			ginkgo.BeforeEach(func() {
				key = "test_delete_key"
				value = "test_delete_value"
			})

			ginkgo.It("should remove the value from cache", func() {
				success := redisCache.Set(ctx, key, value, 0)
				gomega.Expect(success).To(gomega.BeTrue())

				retrieved, found := redisCache.Get(ctx, key)
				gomega.Expect(found).To(gomega.BeTrue())
				gomega.Expect(retrieved).To(gomega.Equal(value))

				redisCache.Delete(ctx, key)

				retrieved, found = redisCache.Get(ctx, key)
				gomega.Expect(found).To(gomega.BeFalse())
			})
		})
	})

	ginkgo.Context("GetOrSet", func() {
		var (
			key           string
			expectedValue string
			ttl           time.Duration
			loader        func() (any, error)
		)

		ginkgo.When("getting or setting a value", func() {
			ginkgo.BeforeEach(func() {
				key = "test_getorset_key"
				expectedValue = "test_getorset_value"
				ttl = 5 * time.Second
				loader = func() (any, error) {
					return expectedValue, nil
				}
			})

			ginkgo.It("should load and cache the value", func() {
				value, err := redisCache.GetOrSet(ctx, key, ttl, loader)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(value).To(gomega.Equal(expectedValue))

				value, err = redisCache.GetOrSet(ctx, key, ttl, loader)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(value).To(gomega.Equal(expectedValue))
			})
		})
	})

	ginkgo.Context("Keys", func() {
		ginkgo.When("getting keys matching a pattern", func() {
			ginkgo.It("should return matching keys", func() {
				redisCache.Set(ctx, "test_keys_1", "value1", 0)
				redisCache.Set(ctx, "test_keys_2", "value2", 0)
				redisCache.Set(ctx, "other_key", "value3", 0)

				keys, err := redisCache.Keys(ctx, "test_keys_*")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(keys).To(gomega.HaveLen(2))
				gomega.Expect(keys).To(gomega.ContainElement("test_keys_1"))
				gomega.Expect(keys).To(gomega.ContainElement("test_keys_2"))
			})
		})
	})

	ginkgo.Context("ContextMethods", func() {
		var (
			key   string
			value string
		)

		ginkgo.When("using context-aware methods", func() {
			ginkgo.BeforeEach(func() {
				key = "test_context_key"
				value = "test_context_value"
			})

			ginkgo.It("should handle context operations correctly", func() {
				success := redisCache.Set(ctx, key, value, 5*time.Second)
				gomega.Expect(success).To(gomega.BeTrue())

				retrieved, found := redisCache.Get(ctx, key)
				gomega.Expect(found).To(gomega.BeTrue())
				gomega.Expect(retrieved).To(gomega.Equal(value))

				redisCache.Delete(ctx, key)

				retrieved, found = redisCache.Get(ctx, key)
				gomega.Expect(found).To(gomega.BeFalse())
			})
		})
	})

	ginkgo.Context("Ping", func() {
		ginkgo.When("pinging the Redis server", func() {
			ginkgo.It("should respond to ping", func() {
				err := redisCache.Ping()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				err = redisCache.PingWithContext(ctx)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})
	})
})

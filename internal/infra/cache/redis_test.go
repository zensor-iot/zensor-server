package cache_test

import (
	"context"
	"time"
	"zensor-server/internal/infra/cache"
	mockcache "zensor-server/test/unit/doubles/infra/cache"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/redis/go-redis/v9"
	"go.uber.org/mock/gomock"
)

var _ = ginkgo.Describe("RedisCache", func() {
	var (
		redisCache      *cache.RedisCache
		mockCacheClient *mockcache.MockCacheClient
		ctrl            *gomock.Controller
		ctx             context.Context
	)

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		mockCacheClient = mockcache.NewMockCacheClient(ctrl)

		redisCache = cache.NewRedisCacheWithClient(mockCacheClient, &cache.RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})

		ctx = context.Background()
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
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
				// Mock Set call
				mockCacheClient.EXPECT().
					Set(gomock.Any(), key, gomock.Any(), time.Duration(0)).
					Return(redis.NewStatusCmd(ctx, "OK"))

				// Mock Get call
				cmd := redis.NewStringCmd(ctx, "get", key)
				cmd.SetVal(`"test_value"`)
				mockCacheClient.EXPECT().
					Get(gomock.Any(), key).
					Return(cmd)

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
				// Mock Set call
				mockCacheClient.EXPECT().
					Set(gomock.Any(), key, gomock.Any(), ttl).
					Return(redis.NewStatusCmd(ctx, "OK"))

				// Mock Get call (before expiration)
				cmd := redis.NewStringCmd(ctx, "get", key)
				cmd.SetVal(`"test_ttl_value"`)
				mockCacheClient.EXPECT().
					Get(gomock.Any(), key).
					Return(cmd)

				// Mock Get call (after expiration) - should return redis.Nil
				cmd = redis.NewStringCmd(ctx, "")
				cmd.SetErr(redis.Nil)
				mockCacheClient.EXPECT().
					Get(gomock.Any(), key).
					Return(cmd)

				success := redisCache.Set(ctx, key, value, ttl)
				gomega.Expect(success).To(gomega.BeTrue())

				retrieved, found := redisCache.Get(ctx, key)
				gomega.Expect(found).To(gomega.BeTrue())
				gomega.Expect(retrieved).To(gomega.Equal(value))

				time.Sleep(2 * time.Second)

				_, found = redisCache.Get(ctx, key)
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
				// Mock Set call
				mockCacheClient.EXPECT().
					Set(gomock.Any(), key, gomock.Any(), time.Duration(0)).
					Return(redis.NewStatusCmd(ctx, "OK"))

				// Mock Get call (before delete)
				cmd := redis.NewStringCmd(ctx, "get", key)
				cmd.SetVal(`"test_delete_value"`)
				mockCacheClient.EXPECT().
					Get(gomock.Any(), key).
					Return(cmd)

				// Mock Delete call
				mockCacheClient.EXPECT().
					Del(gomock.Any(), key).
					Return(redis.NewIntCmd(ctx, 1))

				// Mock Get call (after delete) - should return redis.Nil
				cmd = redis.NewStringCmd(ctx, "")
				cmd.SetErr(redis.Nil)
				mockCacheClient.EXPECT().
					Get(gomock.Any(), key).
					Return(cmd)

				success := redisCache.Set(ctx, key, value, 0)
				gomega.Expect(success).To(gomega.BeTrue())

				retrieved, found := redisCache.Get(ctx, key)
				gomega.Expect(found).To(gomega.BeTrue())
				gomega.Expect(retrieved).To(gomega.Equal(value))

				redisCache.Delete(ctx, key)

				_, found = redisCache.Get(ctx, key)
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
				// Mock Get call (first time - not found)
				cmd := redis.NewStringCmd(ctx, "")
				cmd.SetErr(redis.Nil)
				mockCacheClient.EXPECT().
					Get(gomock.Any(), key).
					Return(cmd)

				// Mock Set call (cache the loaded value)
				mockCacheClient.EXPECT().
					Set(gomock.Any(), key, gomock.Any(), ttl).
					Return(redis.NewStatusCmd(ctx, "OK"))

				// Mock Get call (second time - found in cache)
				cmd2 := redis.NewStringCmd(ctx, "get", key)
				cmd2.SetVal(`"test_getorset_value"`)
				mockCacheClient.EXPECT().
					Get(gomock.Any(), key).
					Return(cmd2)

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
				// Mock Set calls
				mockCacheClient.EXPECT().
					Set(gomock.Any(), "test_keys_1", gomock.Any(), time.Duration(0)).
					Return(redis.NewStatusCmd(ctx, "OK"))
				mockCacheClient.EXPECT().
					Set(gomock.Any(), "test_keys_2", gomock.Any(), time.Duration(0)).
					Return(redis.NewStatusCmd(ctx, "OK"))
				mockCacheClient.EXPECT().
					Set(gomock.Any(), "other_key", gomock.Any(), time.Duration(0)).
					Return(redis.NewStatusCmd(ctx, "OK"))

				// Mock Keys call
				cmd := redis.NewStringSliceCmd(ctx, "keys", "test_keys_*")
				cmd.SetVal([]string{"test_keys_1", "test_keys_2"})
				mockCacheClient.EXPECT().
					Keys(gomock.Any(), "test_keys_*").
					Return(cmd)

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
				// Mock Set call
				mockCacheClient.EXPECT().
					Set(gomock.Any(), key, gomock.Any(), 5*time.Second).
					Return(redis.NewStatusCmd(ctx, "OK"))

				// Mock Get call (before delete)
				cmd := redis.NewStringCmd(ctx, "get", key)
				cmd.SetVal(`"test_context_value"`)
				mockCacheClient.EXPECT().
					Get(gomock.Any(), key).
					Return(cmd)

				// Mock Delete call
				mockCacheClient.EXPECT().
					Del(gomock.Any(), key).
					Return(redis.NewIntCmd(ctx, 1))

				// Mock Get call (after delete) - should return redis.Nil
				cmd = redis.NewStringCmd(ctx, "")
				cmd.SetErr(redis.Nil)
				mockCacheClient.EXPECT().
					Get(gomock.Any(), key).
					Return(cmd)

				success := redisCache.Set(ctx, key, value, 5*time.Second)
				gomega.Expect(success).To(gomega.BeTrue())

				retrieved, found := redisCache.Get(ctx, key)
				gomega.Expect(found).To(gomega.BeTrue())
				gomega.Expect(retrieved).To(gomega.Equal(value))

				redisCache.Delete(ctx, key)

				_, found = redisCache.Get(ctx, key)
				gomega.Expect(found).To(gomega.BeFalse())
			})
		})
	})

	ginkgo.Context("Ping", func() {
		ginkgo.When("pinging the Redis server", func() {
			ginkgo.It("should respond to ping", func() {
				// Mock Ping calls
				mockCacheClient.EXPECT().
					Ping(gomock.Any()).
					Return(redis.NewStatusCmd(ctx, "PONG"))
				mockCacheClient.EXPECT().
					Ping(ctx).
					Return(redis.NewStatusCmd(ctx, "PONG"))

				err := redisCache.Ping()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				err = redisCache.PingWithContext(ctx)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})
	})
})

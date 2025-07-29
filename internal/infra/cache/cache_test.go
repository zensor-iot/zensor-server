package cache_test

import (
	"context"
	"time"
	"zensor-server/internal/infra/cache"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Cache", func() {
	var (
		cacheInstance cache.Cache
		ctx           context.Context
	)

	ginkgo.BeforeEach(func() {
		var err error
		cacheInstance, err = cache.New(nil)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		ctx = context.Background()
	})

	ginkgo.Context("New", func() {
		ginkgo.When("creating a new cache", func() {
			ginkgo.It("should create a valid cache instance", func() {
				gomega.Expect(cacheInstance).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Context("GetSet", func() {
		var (
			key   string
			value string
		)

		ginkgo.When("setting and getting a value", func() {
			ginkgo.BeforeEach(func() {
				key = "test-key"
				value = "test-value"
			})

			ginkgo.It("should store and retrieve the value correctly", func() {
				// Set value
				success := cacheInstance.Set(ctx, key, value, 0)
				gomega.Expect(success).To(gomega.BeTrue())

				// Small delay for Ristretto to process the value
				time.Sleep(10 * time.Millisecond)

				// Get value
				retrieved, found := cacheInstance.Get(ctx, key)
				gomega.Expect(found).To(gomega.BeTrue())
				gomega.Expect(retrieved).To(gomega.Equal(value))
			})
		})
	})

	ginkgo.Context("GetSetWithTTL", func() {
		var (
			key   string
			value string
			ttl   time.Duration
		)

		ginkgo.When("setting a value with TTL", func() {
			ginkgo.BeforeEach(func() {
				key = "test-key-ttl"
				value = "test-value-ttl"
				ttl = 100 * time.Millisecond
			})

			ginkgo.It("should expire the value after TTL", func() {
				// Set value with TTL
				success := cacheInstance.Set(ctx, key, value, ttl)
				gomega.Expect(success).To(gomega.BeTrue())

				// Small delay for Ristretto to process the value
				time.Sleep(10 * time.Millisecond)

				// Get value immediately
				retrieved, found := cacheInstance.Get(ctx, key)
				gomega.Expect(found).To(gomega.BeTrue())
				gomega.Expect(retrieved).To(gomega.Equal(value))

				// Wait for TTL to expire
				time.Sleep(ttl + 50*time.Millisecond)

				// Value should be expired
				retrieved, found = cacheInstance.Get(ctx, key)
				gomega.Expect(found).To(gomega.BeFalse())
				gomega.Expect(retrieved).To(gomega.BeNil())
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
				key = "test-key-delete"
				value = "test-value-delete"
			})

			ginkgo.It("should remove the value from cache", func() {
				// Set value
				cacheInstance.Set(ctx, key, value, 0)

				// Small delay for Ristretto to process the value
				time.Sleep(10 * time.Millisecond)

				// Verify it exists
				retrieved, found := cacheInstance.Get(ctx, key)
				gomega.Expect(found).To(gomega.BeTrue())
				gomega.Expect(retrieved).To(gomega.Equal(value))

				// Delete value
				cacheInstance.Delete(ctx, key)

				// Small delay for Ristretto to process the deletion
				time.Sleep(10 * time.Millisecond)

				// Verify it's gone
				retrieved, found = cacheInstance.Get(ctx, key)
				gomega.Expect(found).To(gomega.BeFalse())
				gomega.Expect(retrieved).To(gomega.BeNil())
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
				key = "test-key-getorset"
				expectedValue = "loaded-value"
				ttl = 1 * time.Second
				loader = func() (any, error) {
					return expectedValue, nil
				}
			})

			ginkgo.It("should load and cache the value", func() {
				// GetOrSet should load the value
				value, err := cacheInstance.GetOrSet(ctx, key, ttl, loader)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(value).To(gomega.Equal(expectedValue))

				// Second call should return cached value
				value, err = cacheInstance.GetOrSet(ctx, key, ttl, loader)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(value).To(gomega.Equal(expectedValue))
			})
		})
	})

	ginkgo.Context("GetOrSetWithContext", func() {
		var (
			key           string
			expectedValue string
			ttl           time.Duration
			loader        func() (any, error)
		)

		ginkgo.When("getting or setting with context", func() {
			ginkgo.BeforeEach(func() {
				key = "test-key-context"
				expectedValue = "loaded-value"
				ttl = 1 * time.Second
				loader = func() (any, error) {
					return expectedValue, nil
				}
			})

			ginkgo.It("should handle context correctly", func() {
				// GetOrSet should load the value
				value, err := cacheInstance.GetOrSet(ctx, key, ttl, loader)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(value).To(gomega.Equal(expectedValue))

				// Second call should return cached value
				value, err = cacheInstance.GetOrSet(ctx, key, ttl, loader)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(value).To(gomega.Equal(expectedValue))
			})
		})
	})

	ginkgo.Context("GetOrSetWithCancelledContext", func() {
		var (
			key    string
			ttl    time.Duration
			loader func() (any, error)
		)

		ginkgo.When("using a cancelled context", func() {
			ginkgo.BeforeEach(func() {
				key = "test-key-cancelled"
				ttl = 1 * time.Second
				loader = func() (any, error) {
					ginkgo.Fail("loader should not be called with cancelled context")
					return nil, nil
				}
			})

			ginkgo.It("should return context error", func() {
				// Create a cancelled context
				cancelledCtx, cancel := context.WithCancel(context.Background())
				cancel()

				// GetOrSet should return context error
				_, err := cacheInstance.GetOrSet(cancelledCtx, key, ttl, loader)
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err).To(gomega.Equal(context.Canceled))
			})
		})
	})

	ginkgo.Context("ConcurrentAccess", func() {
		var (
			key           string
			expectedValue string
			ttl           time.Duration
			loader        func() (any, error)
		)

		ginkgo.When("accessing the cache concurrently", func() {
			ginkgo.BeforeEach(func() {
				key = "test-key-concurrent"
				expectedValue = "concurrent-value"
				ttl = 1 * time.Second
				loader = func() (any, error) {
					time.Sleep(10 * time.Millisecond) // Simulate work
					return expectedValue, nil
				}
			})

			ginkgo.It("should handle concurrent operations safely", func() {
				// Run multiple concurrent GetOrSet operations
				const numGoroutines = 10
				results := make(chan any, numGoroutines)
				errors := make(chan error, numGoroutines)

				for i := 0; i < numGoroutines; i++ {
					go func() {
						value, err := cacheInstance.GetOrSet(ctx, key, ttl, loader)
						results <- value
						errors <- err
					}()
				}

				// Collect results
				for i := 0; i < numGoroutines; i++ {
					value := <-results
					err := <-errors
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					gomega.Expect(value).To(gomega.Equal(expectedValue))
				}
			})
		})
	})

	ginkgo.Context("Config", func() {
		var config *cache.CacheConfig

		ginkgo.When("creating cache with custom config", func() {
			ginkgo.BeforeEach(func() {
				config = &cache.CacheConfig{
					MaxCost:     1 << 20, // 1MB
					NumCounters: 1e6,     // 1M
					BufferItems: 32,
				}
			})

			ginkgo.It("should create cache with custom configuration", func() {
				customCache, err := cache.New(config)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(customCache).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Context("DefaultConfig", func() {
		ginkgo.When("getting default configuration", func() {
			ginkgo.It("should return valid default config", func() {
				config := cache.DefaultConfig()
				gomega.Expect(config).NotTo(gomega.BeNil())
				gomega.Expect(config.MaxCost).To(gomega.BeNumerically(">", int64(0)))
				gomega.Expect(config.NumCounters).To(gomega.BeNumerically(">", int64(0)))
				gomega.Expect(config.BufferItems).To(gomega.BeNumerically(">", int64(0)))
			})
		})
	})
})

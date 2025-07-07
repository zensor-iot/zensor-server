package cache

import (
	"context"
	"time"

	"github.com/dgraph-io/ristretto"
	"golang.org/x/sync/singleflight"
)

// Cache defines the interface for a generic cache with TTL support
type Cache interface {
	Get(key string) (any, bool)
	Set(key string, value any) bool
	SetWithTTL(key string, value any, ttl time.Duration) bool
	Delete(key string)
	Clear()
	GetOrSet(key string, ttl time.Duration, loader func() (any, error)) (any, error)
	GetOrSetWithContext(ctx context.Context, key string, ttl time.Duration, loader func() (any, error)) (any, error)
	Close()
	Wait()
}

// RistrettoCache provides a generic caching implementation with TTL support using Ristretto
type RistrettoCache struct {
	store       *ristretto.Cache
	singleGroup singleflight.Group
	config      *CacheConfig
}

// CacheConfig holds configuration for the cache
type CacheConfig struct {
	// MaxCost is the maximum cost of the cache (in bytes)
	MaxCost int64
	// NumCounters is the number of counters for the cache
	NumCounters int64
	// BufferItems is the number of items to buffer
	BufferItems int64
}

// DefaultConfig returns a default cache configuration
func DefaultConfig() *CacheConfig {
	return &CacheConfig{
		MaxCost:     1 << 30, // 1GB
		NumCounters: 1e7,     // 10M
		BufferItems: 64,
	}
}

// New creates a new RistrettoCache instance and returns the Cache interface
func New(config *CacheConfig) (*RistrettoCache, error) {
	if config == nil {
		config = DefaultConfig()
	}

	store, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: config.NumCounters,
		MaxCost:     config.MaxCost,
		BufferItems: config.BufferItems,
		OnEvict: func(item *ristretto.Item) {
			// Optional: log evictions or perform cleanup
		},
	})
	if err != nil {
		return nil, err
	}

	cache := &RistrettoCache{
		store:  store,
		config: config,
	}

	// Wait for the cache to be ready
	cache.Wait()

	return cache, nil
}

// Get retrieves a value from the cache
func (c *RistrettoCache) Get(key string) (any, bool) {
	return c.store.Get(key)
}

// Set stores a value in the cache with the default TTL
func (c *RistrettoCache) Set(key string, value any) bool {
	return c.store.Set(key, value, 1)
}

// SetWithTTL stores a value in the cache with a custom TTL
func (c *RistrettoCache) SetWithTTL(key string, value any, ttl time.Duration) bool {
	return c.store.SetWithTTL(key, value, 1, ttl)
}

// Delete removes a value from the cache
func (c *RistrettoCache) Delete(key string) {
	c.store.Del(key)
}

// Clear removes all values from the cache
func (c *RistrettoCache) Clear() {
	c.store.Clear()
}

// GetOrSet retrieves a value from the cache, or sets it if not found
// This method uses singleflight to prevent cache stampede
func (c *RistrettoCache) GetOrSet(key string, ttl time.Duration, loader func() (any, error)) (any, error) {
	// Try to get from cache first
	if value, found := c.Get(key); found {
		return value, nil
	}

	// Use singleflight to prevent multiple concurrent loads of the same key
	value, err, _ := c.singleGroup.Do(key, func() (any, error) {
		// Double-check cache after acquiring the lock
		if value, found := c.Get(key); found {
			return value, nil
		}

		// Load the value
		value, err := loader()
		if err != nil {
			return nil, err
		}

		// Store in cache
		c.SetWithTTL(key, value, ttl)
		return value, nil
	})

	return value, err
}

// GetOrSetWithContext is like GetOrSet but accepts a context for cancellation
func (c *RistrettoCache) GetOrSetWithContext(ctx context.Context, key string, ttl time.Duration, loader func() (any, error)) (any, error) {
	// Try to get from cache first
	if value, found := c.Get(key); found {
		return value, nil
	}

	// Use singleflight with context
	value, err, _ := c.singleGroup.Do(key, func() (any, error) {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Double-check cache after acquiring the lock
		if value, found := c.Get(key); found {
			return value, nil
		}

		// Load the value
		value, err := loader()
		if err != nil {
			return nil, err
		}

		// Store in cache
		c.SetWithTTL(key, value, ttl)
		return value, nil
	})

	return value, err
}

// Close closes the cache and frees resources
func (c *RistrettoCache) Close() {
	c.store.Close()
}

// Wait waits for the cache to be ready
func (c *RistrettoCache) Wait() {
	c.store.Wait()
}

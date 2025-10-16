package cache

import (
	"context"
	"time"

	"github.com/dgraph-io/ristretto"
	"golang.org/x/sync/singleflight"
)

// Cache defines the interface for a generic cache with TTL support
type Cache interface {
	Get(ctx context.Context, key string) (any, bool)
	Set(ctx context.Context, key string, value any, ttl time.Duration) bool
	Delete(ctx context.Context, key string)
	GetOrSet(ctx context.Context, key string, ttl time.Duration, loader func() (any, error)) (any, error)
	Keys(ctx context.Context, pattern string) ([]string, error)
}

// RistrettoCache provides a generic caching implementation with TTL support using Ristretto
type RistrettoCache struct {
	store       *ristretto.Cache
	singleGroup singleflight.Group
	config      *CacheConfig
}

// CacheConfig holds configuration for the cache
type CacheConfig struct {
	MaxCost     int64
	NumCounters int64
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

	cache.store.Wait()

	return cache, nil
}

// Get retrieves a value from the cache
func (c *RistrettoCache) Get(ctx context.Context, key string) (any, bool) {
	select {
	case <-ctx.Done():
		return nil, false
	default:
	}
	return c.store.Get(key)
}

// Set stores a value in the cache with TTL
func (c *RistrettoCache) Set(ctx context.Context, key string, value any, ttl time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}
	return c.store.SetWithTTL(key, value, 1, ttl)
}

// Delete removes a value from the cache
func (c *RistrettoCache) Delete(ctx context.Context, key string) {
	select {
	case <-ctx.Done():
		return
	default:
	}
	c.store.Del(key)
}

// GetOrSet retrieves a value from the cache, or sets it if not found
// This method uses singleflight to prevent cache stampede
func (c *RistrettoCache) GetOrSet(ctx context.Context, key string, ttl time.Duration, loader func() (any, error)) (any, error) {
	if value, found := c.Get(ctx, key); found {
		return value, nil
	}

	value, err, _ := c.singleGroup.Do(key, func() (any, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if value, found := c.Get(ctx, key); found {
			return value, nil
		}

		value, err := loader()
		if err != nil {
			return nil, err
		}

		c.Set(ctx, key, value, ttl)
		return value, nil
	})

	return value, err
}

// Keys returns all keys matching the pattern (not supported by Ristretto)
func (c *RistrettoCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return []string{}, nil
}

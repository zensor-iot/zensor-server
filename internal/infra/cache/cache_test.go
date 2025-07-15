package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_New(t *testing.T) {
	cache, err := New(nil)
	require.NoError(t, err)

	assert.NotNil(t, cache)
	assert.NotNil(t, cache.store)
}

func TestCache_GetSet(t *testing.T) {
	cache, err := New(nil)
	require.NoError(t, err)

	// Test basic get/set
	key := "test-key"
	value := "test-value"

	ctx := context.Background()

	// Set value
	success := cache.Set(ctx, key, value, 0)
	assert.True(t, success)

	// Small delay for Ristretto to process the value
	time.Sleep(10 * time.Millisecond)

	// Get value
	retrieved, found := cache.Get(ctx, key)
	assert.True(t, found)
	assert.Equal(t, value, retrieved)
}

func TestCache_GetSetWithTTL(t *testing.T) {
	cache, err := New(nil)
	require.NoError(t, err)

	key := "test-key-ttl"
	value := "test-value-ttl"
	ttl := 100 * time.Millisecond

	ctx := context.Background()

	// Set value with TTL
	success := cache.Set(ctx, key, value, ttl)
	assert.True(t, success)

	// Small delay for Ristretto to process the value
	time.Sleep(10 * time.Millisecond)

	// Get value immediately
	retrieved, found := cache.Get(ctx, key)
	assert.True(t, found)
	assert.Equal(t, value, retrieved)

	// Wait for TTL to expire
	time.Sleep(ttl + 50*time.Millisecond)

	// Value should be expired
	retrieved, found = cache.Get(ctx, key)
	assert.False(t, found)
	assert.Nil(t, retrieved)
}

func TestCache_Delete(t *testing.T) {
	cache, err := New(nil)
	require.NoError(t, err)

	key := "test-key-delete"
	value := "test-value-delete"

	ctx := context.Background()

	// Set value
	cache.Set(ctx, key, value, 0)

	// Small delay for Ristretto to process the value
	time.Sleep(10 * time.Millisecond)

	// Verify it exists
	retrieved, found := cache.Get(ctx, key)
	assert.True(t, found)
	assert.Equal(t, value, retrieved)

	// Delete value
	cache.Delete(ctx, key)

	// Small delay for Ristretto to process the deletion
	time.Sleep(10 * time.Millisecond)

	// Verify it's gone
	retrieved, found = cache.Get(ctx, key)
	assert.False(t, found)
	assert.Nil(t, retrieved)
}

func TestCache_GetOrSet(t *testing.T) {
	cache, err := New(nil)
	require.NoError(t, err)

	key := "test-key-getorset"
	expectedValue := "loaded-value"
	ttl := 1 * time.Second

	// Load function that returns the expected value
	loader := func() (any, error) {
		return expectedValue, nil
	}

	ctx := context.Background()

	// GetOrSet should load the value
	value, err := cache.GetOrSet(ctx, key, ttl, loader)
	require.NoError(t, err)
	assert.Equal(t, expectedValue, value)

	// Second call should return cached value
	value, err = cache.GetOrSet(ctx, key, ttl, loader)
	require.NoError(t, err)
	assert.Equal(t, expectedValue, value)
}

func TestCache_GetOrSetWithContext(t *testing.T) {
	cache, err := New(nil)
	require.NoError(t, err)

	key := "test-key-context"
	expectedValue := "loaded-value"
	ttl := 1 * time.Second

	// Load function that returns the expected value
	loader := func() (any, error) {
		return expectedValue, nil
	}

	ctx := context.Background()

	// GetOrSet should load the value
	value, err := cache.GetOrSet(ctx, key, ttl, loader)
	require.NoError(t, err)
	assert.Equal(t, expectedValue, value)

	// Second call should return cached value
	value, err = cache.GetOrSet(ctx, key, ttl, loader)
	require.NoError(t, err)
	assert.Equal(t, expectedValue, value)
}

func TestCache_GetOrSetWithCancelledContext(t *testing.T) {
	cache, err := New(nil)
	require.NoError(t, err)

	key := "test-key-cancelled"
	ttl := 1 * time.Second

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Load function that should not be called
	loader := func() (any, error) {
		t.Fatal("loader should not be called with cancelled context")
		return nil, nil
	}

	// GetOrSet should return context error
	_, err = cache.GetOrSet(ctx, key, ttl, loader)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestCache_ConcurrentAccess(t *testing.T) {
	cache, err := New(nil)
	require.NoError(t, err)

	key := "test-key-concurrent"
	expectedValue := "concurrent-value"
	ttl := 1 * time.Second

	// Load function that simulates some work
	loader := func() (any, error) {
		time.Sleep(10 * time.Millisecond) // Simulate work
		return expectedValue, nil
	}

	// Run multiple concurrent GetOrSet operations
	const numGoroutines = 10
	results := make(chan any, numGoroutines)
	errors := make(chan error, numGoroutines)

	ctx := context.Background()

	for i := 0; i < numGoroutines; i++ {
		go func() {
			value, err := cache.GetOrSet(ctx, key, ttl, loader)
			results <- value
			errors <- err
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		value := <-results
		err := <-errors
		require.NoError(t, err)
		assert.Equal(t, expectedValue, value)
	}
}

func TestCache_Config(t *testing.T) {
	config := &CacheConfig{
		MaxCost:     1 << 20, // 1MB
		NumCounters: 1e6,     // 1M
		BufferItems: 32,
	}

	cache, err := New(config)
	require.NoError(t, err)

	assert.NotNil(t, cache)
}

func TestCache_DefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.NotNil(t, config)
	assert.Greater(t, config.MaxCost, int64(0))
	assert.Greater(t, config.NumCounters, int64(0))
	assert.Greater(t, config.BufferItems, int64(0))
}

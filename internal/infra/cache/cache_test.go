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
	defer cache.Close()

	assert.NotNil(t, cache)
	assert.NotNil(t, cache.store)
}

func TestCache_GetSet(t *testing.T) {
	cache, err := New(nil)
	require.NoError(t, err)
	defer cache.Close()

	// Test basic get/set
	key := "test-key"
	value := "test-value"

	// Set value
	success := cache.Set(key, value)
	assert.True(t, success)
	cache.Wait() // Ensure value is available

	// Get value
	retrieved, found := cache.Get(key)
	assert.True(t, found)
	assert.Equal(t, value, retrieved)
}

func TestCache_GetSetWithTTL(t *testing.T) {
	cache, err := New(nil)
	require.NoError(t, err)
	defer cache.Close()

	key := "test-key-ttl"
	value := "test-value-ttl"
	ttl := 100 * time.Millisecond

	// Set value with TTL
	success := cache.SetWithTTL(key, value, ttl)
	assert.True(t, success)
	cache.Wait() // Ensure value is available

	// Get value immediately
	retrieved, found := cache.Get(key)
	assert.True(t, found)
	assert.Equal(t, value, retrieved)

	// Wait for TTL to expire
	time.Sleep(ttl + 50*time.Millisecond)

	// Value should be expired
	retrieved, found = cache.Get(key)
	assert.False(t, found)
	assert.Nil(t, retrieved)
}

func TestCache_Delete(t *testing.T) {
	cache, err := New(nil)
	require.NoError(t, err)
	defer cache.Close()

	key := "test-key-delete"
	value := "test-value-delete"

	// Set value
	cache.Set(key, value)
	cache.Wait() // Ensure value is available

	// Verify it exists
	retrieved, found := cache.Get(key)
	assert.True(t, found)
	assert.Equal(t, value, retrieved)

	// Delete value
	cache.Delete(key)
	cache.Wait() // Ensure deletion is processed

	// Verify it's gone
	retrieved, found = cache.Get(key)
	assert.False(t, found)
	assert.Nil(t, retrieved)
}

func TestCache_Clear(t *testing.T) {
	cache, err := New(nil)
	require.NoError(t, err)
	defer cache.Close()

	// Set multiple values
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")
	cache.Wait() // Ensure all values are available

	// Verify they exist
	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")
	_, found3 := cache.Get("key3")
	assert.True(t, found1)
	assert.True(t, found2)
	assert.True(t, found3)

	// Clear cache
	cache.Clear()
	cache.Wait() // Ensure clear is processed

	// Verify all are gone
	_, found1 = cache.Get("key1")
	_, found2 = cache.Get("key2")
	_, found3 = cache.Get("key3")
	assert.False(t, found1)
	assert.False(t, found2)
	assert.False(t, found3)
}

func TestCache_GetOrSet(t *testing.T) {
	cache, err := New(nil)
	require.NoError(t, err)
	defer cache.Close()

	key := "test-key-getorset"
	expectedValue := "loaded-value"
	ttl := 1 * time.Second

	// Load function that returns the expected value
	loader := func() (any, error) {
		return expectedValue, nil
	}

	// GetOrSet should load the value
	value, err := cache.GetOrSet(key, ttl, loader)
	require.NoError(t, err)
	assert.Equal(t, expectedValue, value)

	// Second call should return cached value
	value, err = cache.GetOrSet(key, ttl, loader)
	require.NoError(t, err)
	assert.Equal(t, expectedValue, value)
}

func TestCache_GetOrSetWithContext(t *testing.T) {
	cache, err := New(nil)
	require.NoError(t, err)
	defer cache.Close()

	key := "test-key-context"
	expectedValue := "loaded-value"
	ttl := 1 * time.Second

	// Load function that returns the expected value
	loader := func() (any, error) {
		return expectedValue, nil
	}

	ctx := context.Background()

	// GetOrSetWithContext should load the value
	value, err := cache.GetOrSetWithContext(ctx, key, ttl, loader)
	require.NoError(t, err)
	assert.Equal(t, expectedValue, value)

	// Second call should return cached value
	value, err = cache.GetOrSetWithContext(ctx, key, ttl, loader)
	require.NoError(t, err)
	assert.Equal(t, expectedValue, value)
}

func TestCache_GetOrSetWithCancelledContext(t *testing.T) {
	cache, err := New(nil)
	require.NoError(t, err)
	defer cache.Close()

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

	// GetOrSetWithContext should return context error
	_, err = cache.GetOrSetWithContext(ctx, key, ttl, loader)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestCache_ConcurrentAccess(t *testing.T) {
	cache, err := New(nil)
	require.NoError(t, err)
	defer cache.Close()

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

	for i := 0; i < numGoroutines; i++ {
		go func() {
			value, err := cache.GetOrSet(key, ttl, loader)
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
	defer cache.Close()

	assert.NotNil(t, cache)
}

func TestCache_DefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.NotNil(t, config)
	assert.Greater(t, config.MaxCost, int64(0))
	assert.Greater(t, config.NumCounters, int64(0))
	assert.Greater(t, config.BufferItems, int64(0))
}

package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRedisCache(t *testing.T) {
	// This test requires a Redis server running
	// Skip if Redis is not available
	config := &RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	cache, err := NewRedisCache(config)
	if err != nil {
		t.Skipf("Redis not available, skipping test: %v", err)
	}

	assert.NotNil(t, cache)
	assert.Equal(t, config, cache.config)
}

func TestNewRedisCache_DefaultConfig(t *testing.T) {
	cache, err := NewRedisCache(nil)
	if err != nil {
		t.Skipf("Redis not available, skipping test: %v", err)
	}

	assert.NotNil(t, cache)
	assert.Equal(t, "localhost:6379", cache.config.Addr)
	assert.Equal(t, 0, cache.config.DB)
}

func TestRedisCache_SetAndGet(t *testing.T) {
	config := &RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	cache, err := NewRedisCache(config)
	if err != nil {
		t.Skipf("Redis not available, skipping test: %v", err)
	}

	ctx := context.Background()

	// Test basic set and get
	key := "test_key"
	value := "test_value"

	success := cache.Set(ctx, key, value, 0)
	assert.True(t, success)

	retrieved, found := cache.Get(ctx, key)
	assert.True(t, found)
	assert.Equal(t, value, retrieved)
}

func TestRedisCache_SetWithTTL(t *testing.T) {
	config := &RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	cache, err := NewRedisCache(config)
	if err != nil {
		t.Skipf("Redis not available, skipping test: %v", err)
	}

	ctx := context.Background()

	// Test set with TTL
	key := "test_ttl_key"
	value := "test_ttl_value"
	ttl := 1 * time.Second

	success := cache.Set(ctx, key, value, ttl)
	assert.True(t, success)

	// Value should be available immediately
	retrieved, found := cache.Get(ctx, key)
	assert.True(t, found)
	assert.Equal(t, value, retrieved)

	// Wait for TTL to expire
	time.Sleep(2 * time.Second)

	// Value should be gone
	retrieved, found = cache.Get(ctx, key)
	assert.False(t, found)
}

func TestRedisCache_Delete(t *testing.T) {
	config := &RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	cache, err := NewRedisCache(config)
	if err != nil {
		t.Skipf("Redis not available, skipping test: %v", err)
	}

	ctx := context.Background()

	// Test delete
	key := "test_delete_key"
	value := "test_delete_value"

	success := cache.Set(ctx, key, value, 0)
	assert.True(t, success)

	// Verify value exists
	retrieved, found := cache.Get(ctx, key)
	assert.True(t, found)
	assert.Equal(t, value, retrieved)

	// Delete the value
	cache.Delete(ctx, key)

	// Verify value is gone
	retrieved, found = cache.Get(ctx, key)
	assert.False(t, found)
}

func TestRedisCache_GetOrSet(t *testing.T) {
	config := &RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	cache, err := NewRedisCache(config)
	if err != nil {
		t.Skipf("Redis not available, skipping test: %v", err)
	}

	// Test GetOrSet
	key := "test_getorset_key"
	expectedValue := "test_getorset_value"
	ttl := 5 * time.Second

	loader := func() (any, error) {
		return expectedValue, nil
	}

	ctx := context.Background()

	// First call should set the value
	value, err := cache.GetOrSet(ctx, key, ttl, loader)
	require.NoError(t, err)
	assert.Equal(t, expectedValue, value)

	// Second call should get the cached value
	value, err = cache.GetOrSet(ctx, key, ttl, loader)
	require.NoError(t, err)
	assert.Equal(t, expectedValue, value)
}

func TestRedisCache_Keys(t *testing.T) {
	config := &RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	cache, err := NewRedisCache(config)
	if err != nil {
		t.Skipf("Redis not available, skipping test: %v", err)
	}

	ctx := context.Background()

	// Set some test keys
	cache.Set(ctx, "test_keys_1", "value1", 0)
	cache.Set(ctx, "test_keys_2", "value2", 0)
	cache.Set(ctx, "other_key", "value3", 0)

	// Get keys matching pattern
	keys, err := cache.Keys(ctx, "test_keys_*")
	require.NoError(t, err)
	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "test_keys_1")
	assert.Contains(t, keys, "test_keys_2")
}

func TestRedisCache_ContextMethods(t *testing.T) {
	config := &RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	cache, err := NewRedisCache(config)
	if err != nil {
		t.Skipf("Redis not available, skipping test: %v", err)
	}

	ctx := context.Background()
	key := "test_context_key"
	value := "test_context_value"

	// Test Set
	success := cache.Set(ctx, key, value, 5*time.Second)
	assert.True(t, success)

	// Test Get
	retrieved, found := cache.Get(ctx, key)
	assert.True(t, found)
	assert.Equal(t, value, retrieved)

	// Test Delete
	cache.Delete(ctx, key)

	// Verify value is gone
	retrieved, found = cache.Get(ctx, key)
	assert.False(t, found)
}

func TestRedisCache_Ping(t *testing.T) {
	config := &RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	cache, err := NewRedisCache(config)
	if err != nil {
		t.Skipf("Redis not available, skipping test: %v", err)
	}

	// Test ping
	err = cache.Ping()
	assert.NoError(t, err)

	// Test ping with context
	ctx := context.Background()
	err = cache.PingWithContext(ctx)
	assert.NoError(t, err)
}

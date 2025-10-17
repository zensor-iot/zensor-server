package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// RedisCache provides a generic caching implementation with TTL support using Redis
type RedisCache struct {
	client CacheClient
	config *RedisConfig
}

// RedisConfig holds configuration for the Redis cache
type RedisConfig struct {
	// Addr is the Redis server address (e.g., "localhost:6379")
	Addr string
	// Password is the Redis password (optional)
	Password string
	// DB is the Redis database number (0-15)
	DB int
	// PoolSize is the maximum number of connections in the pool
	PoolSize int
	// MinIdleConns is the minimum number of idle connections in the pool
	MinIdleConns int
	// MaxRetries is the maximum number of retries for failed commands
	MaxRetries int
	// DialTimeout is the timeout for establishing new connections
	DialTimeout time.Duration
	// ReadTimeout is the timeout for socket reads
	ReadTimeout time.Duration
	// WriteTimeout is the timeout for socket writes
	WriteTimeout time.Duration
	// PoolTimeout is the timeout for getting a connection from the pool
	PoolTimeout time.Duration
	// IdleTimeout is the timeout for idle connections
	IdleTimeout time.Duration
	// IdleCheckFrequency is the frequency of idle connection checks
	IdleCheckFrequency time.Duration
}

// DefaultRedisConfig returns a default Redis configuration
func DefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		Addr:               "localhost:6379",
		Password:           "",
		DB:                 0,
		PoolSize:           10,
		MinIdleConns:       5,
		MaxRetries:         3,
		DialTimeout:        5 * time.Second,
		ReadTimeout:        3 * time.Second,
		WriteTimeout:       3 * time.Second,
		PoolTimeout:        4 * time.Second,
		IdleTimeout:        5 * time.Minute,
		IdleCheckFrequency: time.Minute,
	}
}

// NewRedisCacheWithClient creates a new RedisCache instance with a custom client (for testing)
func NewRedisCacheWithClient(client CacheClient, config *RedisConfig) *RedisCache {
	return &RedisCache{
		client: client,
		config: config,
	}
}

// NewRedisCache creates a new RedisCache instance
func NewRedisCache(config *RedisConfig) (*RedisCache, error) {
	if config == nil {
		config = DefaultRedisConfig()
	}

	client := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		PoolTimeout:  config.PoolTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	cache := &RedisCache{
		client: NewRedisClient(client),
		config: config,
	}

	slog.Info("Redis cache initialized",
		slog.String("addr", config.Addr),
		slog.Int("db", config.DB),
		slog.Int("pool_size", config.PoolSize))

	return cache, nil
}

// Get retrieves a value from the cache
func (c *RedisCache) Get(ctx context.Context, key string) (any, bool) {
	ctx, span := c.createSpan(ctx, "get", key)
	defer span.End()

	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			span.SetAttributes(attribute.Bool("cache.hit", false))
			return nil, false
		}
		span.RecordError(err)
		slog.Error("failed to get value from Redis cache",
			slog.String("key", key),
			slog.String("error", err.Error()))
		return nil, false
	}

	var value any
	if err := json.Unmarshal([]byte(result), &value); err != nil {
		span.SetAttributes(attribute.Bool("cache.hit", true))
		return result, true
	}

	span.SetAttributes(attribute.Bool("cache.hit", true))
	return value, true
}

// Set stores a value in the cache with TTL
func (c *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) bool {
	ctx, span := c.createSpan(ctx, "set", key)
	defer span.End()

	span.SetAttributes(attribute.Int64("db.redis.ttl_ms", ttl.Milliseconds()))

	data, err := json.Marshal(value)
	if err != nil {
		span.RecordError(err)
		slog.Error("failed to marshal value for Redis cache",
			slog.String("key", key),
			slog.String("error", err.Error()))
		return false
	}

	var cmd *redis.StatusCmd
	if ttl > 0 {
		cmd = c.client.Set(ctx, key, data, ttl)
	} else {
		cmd = c.client.Set(ctx, key, data, 0)
	}

	if err := cmd.Err(); err != nil {
		span.RecordError(err)
		slog.Error("failed to set value in Redis cache",
			slog.String("key", key),
			slog.String("error", err.Error()))
		return false
	}

	return true
}

// Delete removes a value from the cache
func (c *RedisCache) Delete(ctx context.Context, key string) {
	ctx, span := c.createSpan(ctx, "delete", key)
	defer span.End()

	if err := c.client.Del(ctx, key).Err(); err != nil {
		span.RecordError(err)
		slog.Error("failed to delete value from Redis cache",
			slog.String("key", key),
			slog.String("error", err.Error()))
	}
}

// GetOrSet retrieves a value from the cache, or sets it if not found
func (c *RedisCache) GetOrSet(ctx context.Context, key string, ttl time.Duration, loader func() (any, error)) (any, error) {
	if value, found := c.Get(ctx, key); found {
		return value, nil
	}

	value, err := loader()
	if err != nil {
		return nil, err
	}

	c.Set(ctx, key, value, ttl)
	return value, nil
}

// Keys returns all keys matching the pattern
func (c *RedisCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.client.Keys(ctx, pattern).Result()
}

// Ping tests the Redis connection
func (c *RedisCache) Ping() error {
	ctx := context.Background()
	return c.client.Ping(ctx).Err()
}

// PingWithContext tests the Redis connection with context
func (c *RedisCache) PingWithContext(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *RedisCache) createSpan(ctx context.Context, operation, key string) (context.Context, trace.Span) {
	tracer := otel.Tracer("zensor-server")
	return tracer.Start(ctx, "redis."+operation,
		trace.WithAttributes(
			attribute.String("span.kind", "client"),
			attribute.String("component", "redis-cache"),
			attribute.String("db.system", "redis"),
			attribute.String("db.operation", operation),
			attribute.String("db.redis.key", key),
		),
	)
}

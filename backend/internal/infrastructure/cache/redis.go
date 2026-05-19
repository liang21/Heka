// tasks.md: T097 | spec.md: Redis cache implementation
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheClient wraps go-redis client
type CacheClient struct {
	client *redis.Client
}

// NewCacheClient creates a new Redis cache client
func NewCacheClient(cfg interface{}) (*CacheClient, error) {
	// For now, we'll use a simple configuration
	// In production, this should accept the actual RedisConfig struct
	host := "localhost"
	port := 6379
	password := ""
	db := 0

	// Create Redis client with proper timeout configuration
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &CacheClient{
		client: rdb,
	}, nil
}

// Get retrieves a value from Redis
func (c *CacheClient) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key not found: %s", key)
		}
		return "", fmt.Errorf("failed to get key: %w", err)
	}

	return val, nil
}

// Set stores a value in Redis with an optional TTL
func (c *CacheClient) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if err := c.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set key: %w", err)
	}

	return nil
}

// Delete removes one or more keys from Redis
func (c *CacheClient) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete keys: %w", err)
	}

	return nil
}

// DeleteByPattern deletes all keys matching a pattern using SCAN + DEL
func (c *CacheClient) DeleteByPattern(ctx context.Context, pattern string) error {
	var keys []string

	// Use SCAN to iterate through matching keys
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	// Delete all matching keys
	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
	}

	return nil
}

// Exists checks if one or more keys exist
func (c *CacheClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	count, err := c.client.Exists(ctx, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to check key existence: %w", err)
	}

	return count, nil
}

// Expire sets a TTL on an existing key
func (c *CacheClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := c.client.Expire(ctx, key, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set expiry: %w", err)
	}

	return nil
}

// TTL returns the remaining time to live of a key
func (c *CacheClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}

	return ttl, nil
}

// Increment increments the numeric value of a key by 1
func (c *CacheClient) Increment(ctx context.Context, key string) (int64, error) {
	val, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment: %w", err)
	}

	return val, nil
}

// Decrement decrements the numeric value of a key by 1
func (c *CacheClient) Decrement(ctx context.Context, key string) (int64, error) {
	val, err := c.client.Decr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement: %w", err)
	}

	return val, nil
}

// GetSet sets a new value and returns the old value
func (c *CacheClient) GetSet(ctx context.Context, key string, value string) (string, error) {
	oldVal, err := c.client.GetSet(ctx, key, value).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key not found: %s", key)
		}
		return "", fmt.Errorf("failed to get/set: %w", err)
	}

	return oldVal, nil
}

// MGet retrieves multiple values at once
func (c *CacheClient) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	vals, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple keys: %w", err)
	}

	return vals, nil
}

// MSet stores multiple key-value pairs at once
func (c *CacheClient) MSet(ctx context.Context, pairs ...interface{}) error {
	if len(pairs)%2 != 0 {
		return fmt.Errorf("MSet requires even number of arguments (key-value pairs)")
	}

	if err := c.client.MSet(ctx, pairs...).Err(); err != nil {
		return fmt.Errorf("failed to set multiple keys: %w", err)
	}

	return nil
}

// Close closes the Redis connection
func (c *CacheClient) Close() error {
	return c.client.Close()
}

// Ping checks if the Redis server is responding
func (c *CacheClient) Ping(ctx context.Context) error {
	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis ping failed: %w", err)
	}

	return nil
}

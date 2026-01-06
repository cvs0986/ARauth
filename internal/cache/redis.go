package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache provides Redis-based caching functionality
type Cache struct {
	client *redis.Client
}

// NewCache creates a new Redis cache
// If client is nil, operations will fail gracefully
func NewCache(client *redis.Client) *Cache {
	return &Cache{client: client}
}

// IsAvailable checks if the cache is available
func (c *Cache) IsAvailable() bool {
	return c.client != nil
}

// Get retrieves a value from cache
func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {
	if c.client == nil {
		return fmt.Errorf("cache not available")
	}

	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("cache miss")
	}
	if err != nil {
		return fmt.Errorf("cache get error: %w", err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("cache unmarshal error: %w", err)
	}

	return nil
}

// Set stores a value in cache
func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if c.client == nil {
		return fmt.Errorf("cache not available")
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal error: %w", err)
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("cache set error: %w", err)
	}

	return nil
}

// Delete removes a value from cache
func (c *Cache) Delete(ctx context.Context, key string) error {
	if c.client == nil {
		return fmt.Errorf("cache not available")
	}

	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("cache delete error: %w", err)
	}
	return nil
}

// Exists checks if a key exists in cache
func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	if c.client == nil {
		return false, fmt.Errorf("cache not available")
	}

	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("cache exists error: %w", err)
	}
	return count > 0, nil
}

// SetNX sets a value only if it doesn't exist (for distributed locks)
func (c *Cache) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	if c.client == nil {
		return false, fmt.Errorf("cache not available")
	}

	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("cache marshal error: %w", err)
	}

	result, err := c.client.SetNX(ctx, key, data, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("cache setnx error: %w", err)
	}

	return result, nil
}


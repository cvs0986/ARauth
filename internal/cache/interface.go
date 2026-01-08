package cache

import (
	"context"
	"time"
)

// CacheInterface defines the interface for cache operations
// Both Cache (Redis) and MemoryCache implement this interface
type CacheInterface interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error)
	IsAvailable() bool
}


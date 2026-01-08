package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

var (
	// ErrNotFound is returned when a key is not found in cache
	ErrNotFound = fmt.Errorf("cache miss")
)

// MemoryCache is an in-memory cache implementation for when Redis is not available
// It implements the same interface as Cache but uses in-memory storage
type MemoryCache struct {
	data map[string]*cacheEntry
	mu   sync.RWMutex
}

type cacheEntry struct {
	value     []byte    // Store as JSON bytes for consistency with Redis
	expiresAt time.Time
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache() *MemoryCache {
	mc := &MemoryCache{
		data: make(map[string]*cacheEntry),
	}
	// Start background goroutine to clean up expired entries
	go mc.cleanup()
	return mc
}

// cleanup periodically removes expired entries
func (mc *MemoryCache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		mc.mu.Lock()
		now := time.Now()
		for key, entry := range mc.data {
			if now.After(entry.expiresAt) {
				delete(mc.data, key)
			}
		}
		mc.mu.Unlock()
	}
}

// Set stores a value in the cache with expiration
func (mc *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	// Marshal to JSON like Redis cache does
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal error: %w", err)
	}
	
	mc.data[key] = &cacheEntry{
		value:     data,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

// Get retrieves a value from the cache
func (mc *MemoryCache) Get(ctx context.Context, key string, dest interface{}) error {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	entry, exists := mc.data[key]
	if !exists {
		return ErrNotFound
	}
	
	if time.Now().After(entry.expiresAt) {
		delete(mc.data, key)
		return ErrNotFound
	}
	
	// Unmarshal from JSON like Redis cache does
	if err := json.Unmarshal(entry.value, dest); err != nil {
		return fmt.Errorf("cache unmarshal error: %w", err)
	}
	
	return nil
}

// Delete removes a value from the cache
func (mc *MemoryCache) Delete(ctx context.Context, key string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	delete(mc.data, key)
	return nil
}

// Exists checks if a key exists in the cache
func (mc *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	entry, exists := mc.data[key]
	if !exists {
		return false, nil
	}
	
	if time.Now().After(entry.expiresAt) {
		delete(mc.data, key)
		return false, nil
	}
	
	return true, nil
}

// SetNX sets a value only if it doesn't exist (for distributed locks)
func (mc *MemoryCache) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	// Check if key already exists and is not expired
	if entry, exists := mc.data[key]; exists {
		if time.Now().Before(entry.expiresAt) {
			return false, nil // Key exists and is valid
		}
		// Key exists but expired, delete it
		delete(mc.data, key)
	}
	
	// Marshal to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("cache marshal error: %w", err)
	}
	
	mc.data[key] = &cacheEntry{
		value:     data,
		expiresAt: time.Now().Add(ttl),
	}
	return true, nil
}

// IsAvailable always returns true for in-memory cache
func (mc *MemoryCache) IsAvailable() bool {
	return true
}

// Clear removes all entries from the cache
func (mc *MemoryCache) Clear(ctx context.Context) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.data = make(map[string]*cacheEntry)
	return nil
}


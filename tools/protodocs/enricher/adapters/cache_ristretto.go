package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"
)

// RistrettoCache implements EnrichmentCache using Ristretto
type RistrettoCache struct {
	cache *ristretto.Cache
}

// NewRistrettoCache creates a new Ristretto-based cache
func NewRistrettoCache(maxSize int64, numCounters int64) (*RistrettoCache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: numCounters,
		MaxCost:     maxSize,
		BufferItems: 64,
	})
	if err != nil {
		return nil, fmt.Errorf("create ristretto cache: %w", err)
	}

	return &RistrettoCache{
		cache: cache,
	}, nil
}

// Get retrieves a value from cache
func (c *RistrettoCache) Get(ctx context.Context, key string) (string, bool) {
	val, found := c.cache.Get(key)
	if !found {
		return "", false
	}

	str, ok := val.(string)
	if !ok {
		return "", false
	}

	return str, true
}

// Set stores a value in cache with TTL
func (c *RistrettoCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	cost := int64(len(value))
	success := c.cache.SetWithTTL(key, value, cost, ttl)
	if !success {
		return fmt.Errorf("failed to set cache key %s", key)
	}

	// Wait for value to pass through buffers
	c.cache.Wait()
	return nil
}

// Delete removes a value from cache
func (c *RistrettoCache) Delete(ctx context.Context, key string) error {
	c.cache.Del(key)
	return nil
}

// Clear clears all cache entries
func (c *RistrettoCache) Clear(ctx context.Context) error {
	c.cache.Clear()
	return nil
}

// Close closes the cache
func (c *RistrettoCache) Close() error {
	c.cache.Close()
	return nil
}

// NoOpCache is a no-op implementation for when caching is disabled
type NoOpCache struct{}

// NewNoOpCache creates a no-op cache
func NewNoOpCache() *NoOpCache {
	return &NoOpCache{}
}

// Get always returns not found
func (c *NoOpCache) Get(ctx context.Context, key string) (string, bool) {
	return "", false
}

// Set does nothing
func (c *NoOpCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return nil
}

// Delete does nothing
func (c *NoOpCache) Delete(ctx context.Context, key string) error {
	return nil
}

// Clear does nothing
func (c *NoOpCache) Clear(ctx context.Context) error {
	return nil
}

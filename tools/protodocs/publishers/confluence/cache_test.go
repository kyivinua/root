package confluence

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_BasicOperations(t *testing.T) {
	config := &CacheConfig{
		MaxSize:         10,
		TTL:             1 * time.Minute,
		CleanupInterval: 10 * time.Second,
	}
	cache := NewCache(config)
	defer cache.Close()

	// Test Set and Get
	cache.Set("key1", "value1")
	value, exists := cache.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", value)

	// Test non-existent key
	_, exists = cache.Get("nonexistent")
	assert.False(t, exists)

	// Test Delete
	cache.Delete("key1")
	_, exists = cache.Get("key1")
	assert.False(t, exists)
}

func TestCache_LRUEviction(t *testing.T) {
	config := &CacheConfig{
		MaxSize:         3,
		TTL:             1 * time.Minute,
		CleanupInterval: 10 * time.Second,
	}
	cache := NewCache(config)
	defer cache.Close()

	// Fill cache to capacity
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	assert.Equal(t, 3, cache.Size())

	// Add one more item, should evict oldest (key1)
	cache.Set("key4", "value4")

	assert.Equal(t, 3, cache.Size())
	_, exists := cache.Get("key1")
	assert.False(t, exists, "Oldest key should be evicted")

	// key2, key3, key4 should still exist
	_, exists = cache.Get("key2")
	assert.True(t, exists)
	_, exists = cache.Get("key3")
	assert.True(t, exists)
	_, exists = cache.Get("key4")
	assert.True(t, exists)
}

func TestCache_LRUOrdering(t *testing.T) {
	config := &CacheConfig{
		MaxSize:         3,
		TTL:             1 * time.Minute,
		CleanupInterval: 10 * time.Second,
	}
	cache := NewCache(config)
	defer cache.Close()

	// Add items
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Access key1 to make it most recently used
	_, _ = cache.Get("key1")

	// Add key4, should evict key2 (oldest)
	cache.Set("key4", "value4")

	_, exists := cache.Get("key1")
	assert.True(t, exists, "key1 should still exist (was accessed)")
	_, exists = cache.Get("key2")
	assert.False(t, exists, "key2 should be evicted (oldest)")
	_, exists = cache.Get("key3")
	assert.True(t, exists)
	_, exists = cache.Get("key4")
	assert.True(t, exists)
}

func TestCache_TTLExpiration(t *testing.T) {
	config := &CacheConfig{
		MaxSize:         10,
		TTL:             50 * time.Millisecond,
		CleanupInterval: 10 * time.Millisecond,
	}
	cache := NewCache(config)
	defer cache.Close()

	cache.Set("key1", "value1")

	// Should exist immediately
	_, exists := cache.Get("key1")
	assert.True(t, exists)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	_, exists = cache.Get("key1")
	assert.False(t, exists)
}

func TestCache_UpdateRefreshesExpiration(t *testing.T) {
	config := &CacheConfig{
		MaxSize:         10,
		TTL:             100 * time.Millisecond,
		CleanupInterval: 10 * time.Millisecond,
	}
	cache := NewCache(config)
	defer cache.Close()

	cache.Set("key1", "value1")

	// Wait half the TTL
	time.Sleep(60 * time.Millisecond)

	// Update the value
	cache.Set("key1", "value2")

	// Wait another 60ms (total 120ms from initial set, but only 60ms from update)
	time.Sleep(60 * time.Millisecond)

	// Should still exist because update refreshed expiration
	value, exists := cache.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "value2", value)
}

func TestCache_Clear(t *testing.T) {
	config := &CacheConfig{
		MaxSize:         10,
		TTL:             1 * time.Minute,
		CleanupInterval: 10 * time.Second,
	}
	cache := NewCache(config)
	defer cache.Close()

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	assert.Equal(t, 3, cache.Size())

	cache.Clear()

	assert.Equal(t, 0, cache.Size())
	_, exists := cache.Get("key1")
	assert.False(t, exists)
}

func TestCache_ConcurrentAccess(t *testing.T) {
	config := &CacheConfig{
		MaxSize:         100,
		TTL:             1 * time.Minute,
		CleanupInterval: 10 * time.Second,
	}
	cache := NewCache(config)
	defer cache.Close()

	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				cache.Set(key, j)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				_, _ = cache.Get(key)
			}
		}(i)
	}

	wg.Wait()

	// Cache should be within max size
	assert.LessOrEqual(t, cache.Size(), 100)
}

func TestCache_GetOrCompute(t *testing.T) {
	config := &CacheConfig{
		MaxSize:         10,
		TTL:             1 * time.Minute,
		CleanupInterval: 10 * time.Second,
	}
	cache := NewCache(config)
	defer cache.Close()

	ctx := context.Background()
	computeCount := 0

	compute := func(ctx context.Context) (interface{}, error) {
		computeCount++
		return "computed-value", nil
	}

	// First call should compute
	value, err := cache.GetOrCompute(ctx, "key1", compute)
	require.NoError(t, err)
	assert.Equal(t, "computed-value", value)
	assert.Equal(t, 1, computeCount)

	// Second call should use cache
	value, err = cache.GetOrCompute(ctx, "key1", compute)
	require.NoError(t, err)
	assert.Equal(t, "computed-value", value)
	assert.Equal(t, 1, computeCount, "Compute should not be called again")
}

func TestCache_GetOrComputeError(t *testing.T) {
	config := &CacheConfig{
		MaxSize:         10,
		TTL:             1 * time.Minute,
		CleanupInterval: 10 * time.Second,
	}
	cache := NewCache(config)
	defer cache.Close()

	ctx := context.Background()
	expectedErr := fmt.Errorf("compute error")

	compute := func(ctx context.Context) (interface{}, error) {
		return nil, expectedErr
	}

	// Should return error and not cache
	_, err := cache.GetOrCompute(ctx, "key1", compute)
	assert.Equal(t, expectedErr, err)

	// Value should not be cached
	_, exists := cache.Get("key1")
	assert.False(t, exists)
}

func TestPageCache_BasicOperations(t *testing.T) {
	config := &CacheConfig{
		MaxSize:         10,
		TTL:             1 * time.Minute,
		CleanupInterval: 10 * time.Second,
	}
	pageCache := NewPageCache(config)
	defer pageCache.Close()

	page := &Page{
		ID:    "123",
		Title: "Test Page",
		Space: Space{Key: "TEST"},
	}

	// Test SetPage and GetPage
	pageCache.SetPage(page)
	retrieved, exists := pageCache.GetPage("123")
	assert.True(t, exists)
	assert.Equal(t, page.ID, retrieved.ID)
	assert.Equal(t, page.Title, retrieved.Title)

	// Test DeletePage
	pageCache.DeletePage("123")
	_, exists = pageCache.GetPage("123")
	assert.False(t, exists)
}

func TestPageCache_ByTitle(t *testing.T) {
	config := &CacheConfig{
		MaxSize:         10,
		TTL:             1 * time.Minute,
		CleanupInterval: 10 * time.Second,
	}
	pageCache := NewPageCache(config)
	defer pageCache.Close()

	page := &Page{
		ID:    "123",
		Title: "Test Page",
		Space: Space{Key: "TEST"},
	}

	// Test SetPageByTitle and GetPageByTitle
	pageCache.SetPageByTitle("TEST", "Test Page", page)
	retrieved, exists := pageCache.GetPageByTitle("TEST", "Test Page")
	assert.True(t, exists)
	assert.Equal(t, page.ID, retrieved.ID)

	// Test DeletePageByTitle
	pageCache.DeletePageByTitle("TEST", "Test Page")
	_, exists = pageCache.GetPageByTitle("TEST", "Test Page")
	assert.False(t, exists)
}

func TestPageCache_InvalidatePage(t *testing.T) {
	config := &CacheConfig{
		MaxSize:         10,
		TTL:             1 * time.Minute,
		CleanupInterval: 10 * time.Second,
	}
	pageCache := NewPageCache(config)
	defer pageCache.Close()

	page := &Page{
		ID:    "123",
		Title: "Test Page",
		Space: Space{Key: "TEST"},
	}

	// Cache page both by ID and by title
	pageCache.SetPage(page)
	pageCache.SetPageByTitle("TEST", "Test Page", page)

	// Verify both are cached
	_, exists := pageCache.GetPage("123")
	assert.True(t, exists)
	_, exists = pageCache.GetPageByTitle("TEST", "Test Page")
	assert.True(t, exists)

	// Invalidate should remove both
	pageCache.InvalidatePage(page)

	_, exists = pageCache.GetPage("123")
	assert.False(t, exists)
	_, exists = pageCache.GetPageByTitle("TEST", "Test Page")
	assert.False(t, exists)
}

func TestPageCache_GetOrFetchPage(t *testing.T) {
	config := &CacheConfig{
		MaxSize:         10,
		TTL:             1 * time.Minute,
		CleanupInterval: 10 * time.Second,
	}
	pageCache := NewPageCache(config)
	defer pageCache.Close()

	ctx := context.Background()
	fetchCount := 0

	fetch := func(ctx context.Context, pageID string) (*Page, error) {
		fetchCount++
		return &Page{
			ID:    pageID,
			Title: "Fetched Page",
		}, nil
	}

	// First call should fetch
	page, err := pageCache.GetOrFetchPage(ctx, "123", fetch)
	require.NoError(t, err)
	assert.Equal(t, "123", page.ID)
	assert.Equal(t, 1, fetchCount)

	// Second call should use cache
	page, err = pageCache.GetOrFetchPage(ctx, "123", fetch)
	require.NoError(t, err)
	assert.Equal(t, "123", page.ID)
	assert.Equal(t, 1, fetchCount, "Fetch should not be called again")
}

func TestPageCache_GetOrFindPageByTitle(t *testing.T) {
	config := &CacheConfig{
		MaxSize:         10,
		TTL:             1 * time.Minute,
		CleanupInterval: 10 * time.Second,
	}
	pageCache := NewPageCache(config)
	defer pageCache.Close()

	ctx := context.Background()
	findCount := 0

	find := func(ctx context.Context, spaceKey, title string) (*Page, error) {
		findCount++
		return &Page{
			ID:    "123",
			Title: title,
			Space: Space{Key: spaceKey},
		}, nil
	}

	// First call should find
	page, err := pageCache.GetOrFindPageByTitle(ctx, "TEST", "Test Page", find)
	require.NoError(t, err)
	assert.Equal(t, "Test Page", page.Title)
	assert.Equal(t, 1, findCount)

	// Second call should use cache
	page, err = pageCache.GetOrFindPageByTitle(ctx, "TEST", "Test Page", find)
	require.NoError(t, err)
	assert.Equal(t, "Test Page", page.Title)
	assert.Equal(t, 1, findCount, "Find should not be called again")
}

func TestCache_CleanupExpiredEntries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cleanup test in short mode")
	}

	config := &CacheConfig{
		MaxSize:         10,
		TTL:             50 * time.Millisecond,
		CleanupInterval: 30 * time.Millisecond,
	}
	cache := NewCache(config)
	defer cache.Close()

	// Add multiple entries
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	assert.Equal(t, 3, cache.Size())

	// Wait for cleanup to run
	time.Sleep(150 * time.Millisecond)

	// All entries should be expired and cleaned up
	assert.Equal(t, 0, cache.Size())
}

func TestCache_MetricsIntegration(t *testing.T) {
	metrics := NewMetrics("test")
	config := &CacheConfig{
		MaxSize:         3,
		TTL:             1 * time.Minute,
		CleanupInterval: 10 * time.Second,
		Metrics:         metrics,
	}
	cache := NewCache(config)
	defer cache.Close()

	// Set values (should update cache size)
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Get existing key (cache hit)
	_, exists := cache.Get("key1")
	assert.True(t, exists)

	// Get non-existent key (cache miss)
	_, exists = cache.Get("nonexistent")
	assert.False(t, exists)

	// Add one more to trigger eviction
	cache.Set("key4", "value4")

	// Note: We can't easily verify the metrics values without exposing
	// prometheus internals, but we can verify the operations don't panic
}

func TestPageCache_ClearAndSize(t *testing.T) {
	config := &CacheConfig{
		MaxSize:         10,
		TTL:             1 * time.Minute,
		CleanupInterval: 10 * time.Second,
	}
	pageCache := NewPageCache(config)
	defer pageCache.Close()

	// Add some pages
	for i := 0; i < 5; i++ {
		page := &Page{
			ID:    fmt.Sprintf("page-%d", i),
			Title: fmt.Sprintf("Page %d", i),
		}
		pageCache.SetPage(page)
	}

	assert.Equal(t, 5, pageCache.Size())

	pageCache.Clear()

	assert.Equal(t, 0, pageCache.Size())
}

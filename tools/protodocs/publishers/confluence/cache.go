// Package confluence provides a cache implementation for Confluence API responses
package confluence

import (
	"container/list"
	"context"
	"sync"
	"time"
)

// CacheEntry represents a cached item with expiration
type CacheEntry struct {
	Key       string
	Value     interface{}
	ExpiresAt time.Time
	Element   *list.Element
}

// Cache implements an LRU cache with TTL support
type Cache struct {
	mu          sync.RWMutex
	maxSize     int
	ttl         time.Duration
	items       map[string]*CacheEntry
	lruList     *list.List
	metrics     *Metrics
	stopCleanup chan struct{}
}

// CacheConfig holds configuration for the cache
type CacheConfig struct {
	MaxSize        int           // Maximum number of items in cache
	TTL            time.Duration // Time to live for cache entries
	CleanupInterval time.Duration // How often to clean expired entries
	Metrics        *Metrics      // Metrics collector
}

// DefaultCacheConfig returns sensible defaults for cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		MaxSize:        1000,
		TTL:            10 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		Metrics:        nil,
	}
}

// NewCache creates a new LRU cache with TTL support
func NewCache(config *CacheConfig) *Cache {
	if config == nil {
		config = DefaultCacheConfig()
	}

	cache := &Cache{
		maxSize:     config.MaxSize,
		ttl:         config.TTL,
		items:       make(map[string]*CacheEntry, config.MaxSize),
		lruList:     list.New(),
		metrics:     config.Metrics,
		stopCleanup: make(chan struct{}),
	}

	// Start background cleanup goroutine
	go cache.cleanupExpired(config.CleanupInterval)

	return cache
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.items[key]
	if !exists {
		if c.metrics != nil {
			c.metrics.RecordCacheMiss()
		}
		return nil, false
	}

	// Check if entry has expired
	if time.Now().After(entry.ExpiresAt) {
		c.removeEntry(entry)
		if c.metrics != nil {
			c.metrics.RecordCacheMiss()
		}
		return nil, false
	}

	// Move to front of LRU list (mark as recently used)
	c.lruList.MoveToFront(entry.Element)

	if c.metrics != nil {
		c.metrics.RecordCacheHit()
	}

	return entry.Value, true
}

// Set adds or updates a value in the cache
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if key already exists
	if entry, exists := c.items[key]; exists {
		// Update existing entry
		entry.Value = value
		entry.ExpiresAt = time.Now().Add(c.ttl)
		c.lruList.MoveToFront(entry.Element)
		return
	}

	// Check if we need to evict an entry
	if c.lruList.Len() >= c.maxSize {
		c.evictOldest()
	}

	// Add new entry
	entry := &CacheEntry{
		Key:       key,
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}
	entry.Element = c.lruList.PushFront(entry)
	c.items[key] = entry

	if c.metrics != nil {
		c.metrics.UpdateCacheSize(c.lruList.Len())
	}
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists := c.items[key]; exists {
		c.removeEntry(entry)
	}
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*CacheEntry, c.maxSize)
	c.lruList.Init()

	if c.metrics != nil {
		c.metrics.UpdateCacheSize(0)
	}
}

// Size returns the current number of items in the cache
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lruList.Len()
}

// Close stops the background cleanup goroutine
func (c *Cache) Close() {
	close(c.stopCleanup)
}

// GetOrCompute retrieves a value from cache or computes it if not present
func (c *Cache) GetOrCompute(ctx context.Context, key string, compute func(context.Context) (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	if value, exists := c.Get(key); exists {
		return value, nil
	}

	// Compute the value
	value, err := compute(ctx)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.Set(key, value)

	return value, nil
}

// removeEntry removes an entry from the cache (caller must hold lock)
func (c *Cache) removeEntry(entry *CacheEntry) {
	c.lruList.Remove(entry.Element)
	delete(c.items, entry.Key)

	if c.metrics != nil {
		c.metrics.UpdateCacheSize(c.lruList.Len())
	}
}

// evictOldest removes the least recently used entry (caller must hold lock)
func (c *Cache) evictOldest() {
	oldest := c.lruList.Back()
	if oldest != nil {
		entry := oldest.Value.(*CacheEntry)
		c.removeEntry(entry)

		if c.metrics != nil {
			c.metrics.RecordCacheEviction()
		}
	}
}

// cleanupExpired periodically removes expired entries
func (c *Cache) cleanupExpired(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.removeExpiredEntries()
		case <-c.stopCleanup:
			return
		}
	}
}

// removeExpiredEntries removes all expired entries from the cache
func (c *Cache) removeExpiredEntries() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	var toRemove []*CacheEntry

	// Collect expired entries
	for _, entry := range c.items {
		if now.After(entry.ExpiresAt) {
			toRemove = append(toRemove, entry)
		}
	}

	// Remove them
	for _, entry := range toRemove {
		c.removeEntry(entry)
		if c.metrics != nil {
			c.metrics.RecordCacheEviction()
		}
	}
}

// PageCache provides typed cache methods for Confluence pages
type PageCache struct {
	cache *Cache
}

// NewPageCache creates a new cache for Confluence pages
func NewPageCache(config *CacheConfig) *PageCache {
	return &PageCache{
		cache: NewCache(config),
	}
}

// GetPage retrieves a page by ID from cache
func (pc *PageCache) GetPage(pageID string) (*Page, bool) {
	value, exists := pc.cache.Get("page:" + pageID)
	if !exists {
		return nil, false
	}
	if page, ok := value.(*Page); ok {
		return page, true
	}
	return nil, false
}

// SetPage stores a page in cache
func (pc *PageCache) SetPage(page *Page) {
	pc.cache.Set("page:"+page.ID, page)
}

// DeletePage removes a page from cache
func (pc *PageCache) DeletePage(pageID string) {
	pc.cache.Delete("page:" + pageID)
}

// GetPageByTitle retrieves a page by space and title from cache
func (pc *PageCache) GetPageByTitle(spaceKey, title string) (*Page, bool) {
	key := "page:" + spaceKey + ":" + title
	value, exists := pc.cache.Get(key)
	if !exists {
		return nil, false
	}
	if page, ok := value.(*Page); ok {
		return page, true
	}
	return nil, false
}

// SetPageByTitle stores a page by space and title in cache
func (pc *PageCache) SetPageByTitle(spaceKey, title string, page *Page) {
	key := "page:" + spaceKey + ":" + title
	pc.cache.Set(key, page)
}

// DeletePageByTitle removes a page by space and title from cache
func (pc *PageCache) DeletePageByTitle(spaceKey, title string) {
	key := "page:" + spaceKey + ":" + title
	pc.cache.Delete(key)
}

// InvalidatePage removes all cached entries for a page
func (pc *PageCache) InvalidatePage(page *Page) {
	pc.DeletePage(page.ID)
	// Also invalidate by title if we have the space key
	if page.Space.Key != "" {
		pc.DeletePageByTitle(page.Space.Key, page.Title)
	}
}

// Clear removes all entries from the cache
func (pc *PageCache) Clear() {
	pc.cache.Clear()
}

// Size returns the current number of items in the cache
func (pc *PageCache) Size() int {
	return pc.cache.Size()
}

// Close stops the background cleanup goroutine
func (pc *PageCache) Close() {
	pc.cache.Close()
}

// GetOrFetchPage retrieves a page from cache or fetches it
func (pc *PageCache) GetOrFetchPage(ctx context.Context, pageID string, fetch func(context.Context, string) (*Page, error)) (*Page, error) {
	// Try cache first
	if page, exists := pc.GetPage(pageID); exists {
		return page, nil
	}

	// Fetch from API
	page, err := fetch(ctx, pageID)
	if err != nil {
		return nil, err
	}

	// Store in cache
	pc.SetPage(page)

	return page, nil
}

// GetOrFindPageByTitle retrieves a page from cache or finds it
func (pc *PageCache) GetOrFindPageByTitle(ctx context.Context, spaceKey, title string, find func(context.Context, string, string) (*Page, error)) (*Page, error) {
	// Try cache first
	if page, exists := pc.GetPageByTitle(spaceKey, title); exists {
		return page, nil
	}

	// Find via API
	page, err := find(ctx, spaceKey, title)
	if err != nil {
		return nil, err
	}

	// Store in cache
	pc.SetPageByTitle(spaceKey, title, page)

	return page, nil
}

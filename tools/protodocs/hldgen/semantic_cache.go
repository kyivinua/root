package hldgen

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// SemanticCache provides caching for LLM responses.
// Uses content-based hashing for exact match detection.
// Thread-safe for concurrent use.
type SemanticCache struct {
	store      map[string]*CacheEntry
	mu         sync.RWMutex
	ttl        time.Duration
	maxEntries int
	enabled    bool
	shutdown   chan struct{} // Channel to signal goroutine shutdown

	// Statistics
	hits   int64
	misses int64
}

// CacheEntry represents a cached LLM response
type CacheEntry struct {
	Response  *LLMResponse
	CreatedAt time.Time
	LastUsed  time.Time
	UseCount  int
}

// NewSemanticCache creates a new semantic cache with the given configuration.
// The cache automatically starts a background cleanup goroutine if enabled.
// Call Shutdown() to properly clean up resources when done.
func NewSemanticCache(cfg SemanticCacheConfig) *SemanticCache {
	// Parse TTL string to duration
	ttl := 1 * time.Hour // Default
	if cfg.TTL != "" {
		if parsed, err := time.ParseDuration(cfg.TTL); err == nil {
			ttl = parsed
		}
	}

	maxEntries := cfg.MaxEntries
	if maxEntries == 0 {
		maxEntries = 1000 // Default max entries
	}

	cache := &SemanticCache{
		store:      make(map[string]*CacheEntry),
		ttl:        ttl,
		maxEntries: maxEntries,
		enabled:    cfg.Enabled,
		shutdown:   make(chan struct{}),
	}

	// Start background cleanup goroutine
	if cfg.Enabled {
		go cache.cleanupLoop()
	}

	return cache
}

// Get retrieves a response from cache.
// Returns the cached response and true if found and not expired, otherwise nil and false.
// Thread-safe for concurrent access.
func (sc *SemanticCache) Get(ctx context.Context, prompt string, config LLMConfig) (*LLMResponse, bool) {
	// Check context cancellation first
	select {
	case <-ctx.Done():
		return nil, false
	default:
	}

	if !sc.enabled {
		return nil, false
	}

	key := sc.generateKey(prompt, config)

	// Use single lock to avoid race conditions
	sc.mu.Lock()
	defer sc.mu.Unlock()

	entry, exists := sc.store[key]
	if !exists {
		sc.misses++
		return nil, false
	}

	// Check if entry is expired
	if time.Since(entry.CreatedAt) > sc.ttl {
		delete(sc.store, key)
		sc.misses++
		return nil, false
	}

	// Update usage statistics
	entry.LastUsed = time.Now()
	entry.UseCount++
	sc.hits++

	// Return a copy to avoid mutation
	return sc.copyResponse(entry.Response), true
}

// Put stores a response in cache.
// Thread-safe for concurrent access.
func (sc *SemanticCache) Put(ctx context.Context, prompt string, config LLMConfig, response *LLMResponse) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return
	default:
	}

	if !sc.enabled {
		return
	}

	if response == nil {
		return // Don't cache nil responses
	}

	key := sc.generateKey(prompt, config)

	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Check if we need to evict entries
	if len(sc.store) >= sc.maxEntries {
		sc.evictOldest()
	}

	// Store the entry
	sc.store[key] = &CacheEntry{
		Response:  sc.copyResponse(response),
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		UseCount:  1,
	}
}

// generateKey creates a cache key from prompt and config.
// Uses SHA-256 hash of JSON-serialized data for deterministic keys.
func (sc *SemanticCache) generateKey(prompt string, config LLMConfig) string {
	// Create a deterministic key from prompt + relevant config
	data := struct {
		Prompt      string
		Temperature float64
		MaxTokens   int
	}{
		Prompt:      prompt,
		Temperature: config.Temperature,
		MaxTokens:   config.MaxTokens,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		// Fallback to simple hash if marshal fails
		// This should never happen with our simple struct, but handle it anyway
		hash := sha256.Sum256([]byte(prompt))
		return hex.EncodeToString(hash[:])
	}

	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

// copyResponse creates a copy of LLMResponse to avoid mutation
func (sc *SemanticCache) copyResponse(resp *LLMResponse) *LLMResponse {
	if resp == nil {
		return nil
	}

	return &LLMResponse{
		Content:      resp.Content,
		TokensUsed:   resp.TokensUsed,
		Model:        resp.Model,
		Provider:     resp.Provider,
		Confidence:   resp.Confidence,
		FinishReason: resp.FinishReason,
	}
}

// evictOldest removes the least recently used entry
func (sc *SemanticCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	// Find the least recently used entry
	for key, entry := range sc.store {
		if oldestKey == "" || entry.LastUsed.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.LastUsed
		}
	}

	if oldestKey != "" {
		delete(sc.store, oldestKey)
	}
}

// cleanupLoop periodically removes expired entries.
// Runs in the background and stops when Shutdown() is called.
func (sc *SemanticCache) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sc.cleanup()
		case <-sc.shutdown:
			return
		}
	}
}

// cleanup removes expired entries
func (sc *SemanticCache) cleanup() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	now := time.Now()
	for key, entry := range sc.store {
		if now.Sub(entry.CreatedAt) > sc.ttl {
			delete(sc.store, key)
		}
	}
}

// GetStats returns cache statistics
func (sc *SemanticCache) GetStats() CacheStats {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	hitRate := 0.0
	total := sc.hits + sc.misses
	if total > 0 {
		hitRate = float64(sc.hits) / float64(total) * 100
	}

	return CacheStats{
		Hits:       sc.hits,
		Misses:     sc.misses,
		HitRate:    hitRate,
		Size:       len(sc.store),
		MaxEntries: sc.maxEntries,
	}
}

// CacheStats holds cache statistics
type CacheStats struct {
	Hits       int64
	Misses     int64
	HitRate    float64
	Size       int
	MaxEntries int
}

// String returns a human-readable representation of cache stats
func (cs CacheStats) String() string {
	return fmt.Sprintf(
		"Cache: %d/%d entries, %.1f%% hit rate (%d hits, %d misses)",
		cs.Size, cs.MaxEntries, cs.HitRate, cs.Hits, cs.Misses,
	)
}

// Clear removes all entries from cache
func (sc *SemanticCache) Clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.store = make(map[string]*CacheEntry)
	sc.hits = 0
	sc.misses = 0
}

// CachedLLMClient wraps an LLMClient with semantic caching
type CachedLLMClient struct {
	client LLMClient
	cache  *SemanticCache
}

// NewCachedLLMClient creates a new cached LLM client
func NewCachedLLMClient(client LLMClient, cache *SemanticCache) *CachedLLMClient {
	return &CachedLLMClient{
		client: client,
		cache:  cache,
	}
}

// Generate implements LLMClient.Generate with caching
func (c *CachedLLMClient) Generate(ctx context.Context, prompt string, config LLMConfig) (*LLMResponse, error) {
	// Try to get from cache first
	if cached, hit := c.cache.Get(ctx, prompt, config); hit {
		// Add cache hit indicator to response
		cached.FinishReason = "cache_hit"
		return cached, nil
	}

	// Cache miss - call underlying client
	resp, err := c.client.Generate(ctx, prompt, config)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.cache.Put(ctx, prompt, config, resp)

	return resp, nil
}

// GetModelName implements LLMClient.GetModelName
func (c *CachedLLMClient) GetModelName() string {
	return c.client.GetModelName()
}

// GetProviderName implements LLMClient.GetProviderName
func (c *CachedLLMClient) GetProviderName() string {
	return c.client.GetProviderName()
}

// Shutdown stops the background cleanup goroutine and cleans up resources.
// Should be called when the cache is no longer needed.
// Thread-safe and can be called multiple times.
func (sc *SemanticCache) Shutdown() {
	if sc.shutdown != nil {
		close(sc.shutdown)
	}
}

// WrapWithCache wraps an LLM client with caching if enabled
func WrapWithCache(client LLMClient, cache *SemanticCache) LLMClient {
	if cache == nil || !cache.enabled {
		return client
	}
	return NewCachedLLMClient(client, cache)
}

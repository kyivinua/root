package hldgen

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSemanticCache(t *testing.T) {
	tests := []struct {
		name           string
		cfg            SemanticCacheConfig
		wantEnabled    bool
		wantTTL        time.Duration
		wantMaxEntries int
	}{
		{
			name: "with explicit values",
			cfg: SemanticCacheConfig{
				Enabled:    true,
				TTL:        "30m",
				MaxEntries: 500,
			},
			wantEnabled:    true,
			wantTTL:        30 * time.Minute,
			wantMaxEntries: 500,
		},
		{
			name: "with defaults",
			cfg: SemanticCacheConfig{
				Enabled: true,
			},
			wantEnabled:    true,
			wantTTL:        1 * time.Hour,
			wantMaxEntries: 1000,
		},
		{
			name: "disabled",
			cfg: SemanticCacheConfig{
				Enabled: false,
			},
			wantEnabled:    false,
			wantTTL:        1 * time.Hour,
			wantMaxEntries: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewSemanticCache(tt.cfg)
			assert.Equal(t, tt.wantEnabled, cache.enabled)
			assert.Equal(t, tt.wantTTL, cache.ttl)
			assert.Equal(t, tt.wantMaxEntries, cache.maxEntries)
			assert.NotNil(t, cache.store)
		})
	}
}

func TestSemanticCache_GetPut(t *testing.T) {
	cache := NewSemanticCache(SemanticCacheConfig{
		Enabled:    true,
		TTL:        "1h",
		MaxEntries: 10,
	})

	ctx := context.Background()
	prompt := "What is gRPC?"
	config := LLMConfig{Temperature: 0.7, MaxTokens: 100}

	t.Run("cache miss on first get", func(t *testing.T) {
		resp, hit := cache.Get(ctx, prompt, config)
		assert.False(t, hit)
		assert.Nil(t, resp)

		stats := cache.GetStats()
		assert.Equal(t, int64(1), stats.Misses)
		assert.Equal(t, int64(0), stats.Hits)
	})

	t.Run("put and get", func(t *testing.T) {
		originalResp := &LLMResponse{
			Content:    "gRPC is a high-performance RPC framework",
			TokensUsed: 50,
			Provider:   "test",
			Model:      "test-model",
			Confidence: 0.9,
		}

		cache.Put(ctx, prompt, config, originalResp)

		// Should get cache hit now
		cached, hit := cache.Get(ctx, prompt, config)
		assert.True(t, hit)
		require.NotNil(t, cached)
		assert.Equal(t, originalResp.Content, cached.Content)
		assert.Equal(t, originalResp.TokensUsed, cached.TokensUsed)
		assert.Equal(t, originalResp.Provider, cached.Provider)

		stats := cache.GetStats()
		assert.Equal(t, int64(1), stats.Hits)
		assert.Equal(t, 1, stats.Size)
	})

	t.Run("different config produces cache miss", func(t *testing.T) {
		differentConfig := LLMConfig{Temperature: 0.9, MaxTokens: 200}
		resp, hit := cache.Get(ctx, prompt, differentConfig)
		assert.False(t, hit)
		assert.Nil(t, resp)
	})
}

func TestSemanticCache_Disabled(t *testing.T) {
	cache := NewSemanticCache(SemanticCacheConfig{
		Enabled: false,
	})

	ctx := context.Background()
	prompt := "test prompt"
	config := LLMConfig{}

	// Put should do nothing
	cache.Put(ctx, prompt, config, &LLMResponse{Content: "test"})

	// Get should always miss
	resp, hit := cache.Get(ctx, prompt, config)
	assert.False(t, hit)
	assert.Nil(t, resp)

	stats := cache.GetStats()
	assert.Equal(t, int64(0), stats.Misses)
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, 0, stats.Size)
}

func TestSemanticCache_TTL(t *testing.T) {
	cache := NewSemanticCache(SemanticCacheConfig{
		Enabled:    true,
		TTL:        "100ms",
		MaxEntries: 10,
	})

	ctx := context.Background()
	prompt := "test prompt"
	config := LLMConfig{}

	// Put an entry
	cache.Put(ctx, prompt, config, &LLMResponse{Content: "test"})

	// Should hit immediately
	resp, hit := cache.Get(ctx, prompt, config)
	assert.True(t, hit)
	assert.NotNil(t, resp)

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Should miss after TTL
	resp, hit = cache.Get(ctx, prompt, config)
	assert.False(t, hit)
	assert.Nil(t, resp)
}

func TestSemanticCache_Eviction(t *testing.T) {
	cache := NewSemanticCache(SemanticCacheConfig{
		Enabled:    true,
		TTL:        "1h",
		MaxEntries: 3,
	})

	ctx := context.Background()
	config := LLMConfig{}

	// Fill cache to capacity
	for i := 0; i < 3; i++ {
		prompt := fmt.Sprintf("prompt %d", i)
		cache.Put(ctx, prompt, config, &LLMResponse{Content: fmt.Sprintf("response %d", i)})
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	stats := cache.GetStats()
	assert.Equal(t, 3, stats.Size)

	// Add one more - should evict the oldest
	cache.Put(ctx, "prompt 4", config, &LLMResponse{Content: "response 4"})

	stats = cache.GetStats()
	assert.Equal(t, 3, stats.Size)

	// The first entry should have been evicted
	resp, hit := cache.Get(ctx, "prompt 0", config)
	assert.False(t, hit)
	assert.Nil(t, resp)

	// The new entry should exist
	resp, hit = cache.Get(ctx, "prompt 4", config)
	assert.True(t, hit)
	assert.NotNil(t, resp)
}

func TestSemanticCache_Clear(t *testing.T) {
	cache := NewSemanticCache(SemanticCacheConfig{
		Enabled:    true,
		TTL:        "1h",
		MaxEntries: 10,
	})

	ctx := context.Background()
	config := LLMConfig{}

	// Add entries
	for i := 0; i < 5; i++ {
		cache.Put(ctx, fmt.Sprintf("prompt %d", i), config, &LLMResponse{Content: "test"})
	}

	stats := cache.GetStats()
	assert.Equal(t, 5, stats.Size)

	// Clear cache
	cache.Clear()

	stats = cache.GetStats()
	assert.Equal(t, 0, stats.Size)
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)
}

func TestSemanticCache_Stats(t *testing.T) {
	cache := NewSemanticCache(SemanticCacheConfig{
		Enabled:    true,
		TTL:        "1h",
		MaxEntries: 10,
	})

	ctx := context.Background()
	prompt := "test"
	config := LLMConfig{}

	// 2 misses
	cache.Get(ctx, prompt, config)
	cache.Get(ctx, prompt, config)

	// Add entry
	cache.Put(ctx, prompt, config, &LLMResponse{Content: "test"})

	// 3 hits
	cache.Get(ctx, prompt, config)
	cache.Get(ctx, prompt, config)
	cache.Get(ctx, prompt, config)

	stats := cache.GetStats()
	assert.Equal(t, int64(3), stats.Hits)
	assert.Equal(t, int64(2), stats.Misses)
	assert.Equal(t, 60.0, stats.HitRate) // 3/(3+2) * 100 = 60%
	assert.Equal(t, 1, stats.Size)

	str := stats.String()
	assert.Contains(t, str, "60.0%")
	assert.Contains(t, str, "3 hits")
	assert.Contains(t, str, "2 misses")
}

func TestCachedLLMClient_Generate(t *testing.T) {
	mockClient := NewMockLLMClient("test-provider", "test-model")

	cache := NewSemanticCache(SemanticCacheConfig{
		Enabled:    true,
		TTL:        "1h",
		MaxEntries: 10,
	})

	cachedClient := NewCachedLLMClient(mockClient, cache)

	ctx := context.Background()
	prompt := "test prompt"
	config := LLMConfig{}

	t.Run("first call - cache miss", func(t *testing.T) {
		resp, err := cachedClient.Generate(ctx, prompt, config)
		require.NoError(t, err)
		assert.Equal(t, "Mock LLM response", resp.Content)
		assert.NotEqual(t, "cache_hit", resp.FinishReason)

		stats := cache.GetStats()
		assert.Equal(t, int64(1), stats.Misses)
		assert.Equal(t, int64(0), stats.Hits)
	})

	t.Run("second call - cache hit", func(t *testing.T) {
		resp, err := cachedClient.Generate(ctx, prompt, config)
		require.NoError(t, err)
		assert.Equal(t, "Mock LLM response", resp.Content)
		assert.Equal(t, "cache_hit", resp.FinishReason)

		stats := cache.GetStats()
		assert.Equal(t, int64(1), stats.Hits)
	})

	t.Run("GetModelName", func(t *testing.T) {
		assert.Equal(t, "test-model", cachedClient.GetModelName())
	})

	t.Run("GetProviderName", func(t *testing.T) {
		assert.Equal(t, "test-provider", cachedClient.GetProviderName())
	})
}

func TestWrapWithCache(t *testing.T) {
	mockClient := NewMockLLMClient("test", "test-model")

	t.Run("with enabled cache", func(t *testing.T) {
		cache := NewSemanticCache(SemanticCacheConfig{Enabled: true})
		wrapped := WrapWithCache(mockClient, cache)
		assert.IsType(t, &CachedLLMClient{}, wrapped)
	})

	t.Run("with disabled cache", func(t *testing.T) {
		cache := NewSemanticCache(SemanticCacheConfig{Enabled: false})
		wrapped := WrapWithCache(mockClient, cache)
		assert.Equal(t, mockClient, wrapped)
	})

	t.Run("with nil cache", func(t *testing.T) {
		wrapped := WrapWithCache(mockClient, nil)
		assert.Equal(t, mockClient, wrapped)
	})
}

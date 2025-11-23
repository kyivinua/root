package template

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

// EnrichmentResult represents the result of LLM enrichment
type EnrichmentResult struct {
	ChunkID     string
	EnrichType  string
	Original    string
	Enriched    string
	Confidence  float64
	Duration    time.Duration
	Cached      bool
	Error       error
}

// LLMProvider interface for pluggable LLM backends
type LLMProvider interface {
	// Enrich generates enriched content for the given prompt and context
	Enrich(ctx context.Context, enrichType, prompt string, context map[string]interface{}) (string, error)

	// Name returns the provider name
	Name() string
}

// Enricher handles LLM enrichment of template chunks
type Enricher struct {
	provider LLMProvider
	cache    *EnrichmentCache
	config   EnricherConfig
}

// EnricherConfig holds enricher configuration
type EnricherConfig struct {
	EnableCache      bool
	CacheTTL         time.Duration
	MaxConcurrency   int
	Timeout          time.Duration
	RetryAttempts    int
	RetryDelay       time.Duration
	EnableFallback   bool
	FallbackProvider LLMProvider
}

// DefaultEnricherConfig returns default configuration
func DefaultEnricherConfig() EnricherConfig {
	return EnricherConfig{
		EnableCache:    true,
		CacheTTL:       24 * time.Hour,
		MaxConcurrency: 5,
		Timeout:        30 * time.Second,
		RetryAttempts:  3,
		RetryDelay:     2 * time.Second,
		EnableFallback: false,
	}
}

// NewEnricher creates a new enricher
func NewEnricher(provider LLMProvider, config EnricherConfig) *Enricher {
	return &Enricher{
		provider: provider,
		cache:    NewEnrichmentCache(config.CacheTTL),
		config:   config,
	}
}

// EnrichChunks enriches all enrichable chunks in parallel
func (e *Enricher) EnrichChunks(ctx context.Context, graph *ChunkGraph) ([]*EnrichmentResult, error) {
	enrichableChunks := graph.GetEnrichableChunks()
	if len(enrichableChunks) == 0 {
		return []*EnrichmentResult{}, nil
	}

	results := make([]*EnrichmentResult, len(enrichableChunks))
	errChan := make(chan error, len(enrichableChunks))

	// Use semaphore for concurrency control
	sem := make(chan struct{}, e.config.MaxConcurrency)

	var wg sync.WaitGroup
	for i, chunk := range enrichableChunks {
		wg.Add(1)
		go func(index int, ch *Chunk) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			result := e.enrichChunk(ctx, ch)
			results[index] = result

			if result.Error != nil {
				errChan <- result.Error
			}
		}(i, chunk)
	}

	wg.Wait()
	close(errChan)

	// Collect errors
	errors := []error{}
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return results, fmt.Errorf("enrichment errors: %v", errors)
	}

	return results, nil
}

// enrichChunk enriches a single chunk
func (e *Enricher) enrichChunk(ctx context.Context, chunk *Chunk) *EnrichmentResult {
	result := &EnrichmentResult{
		ChunkID:    chunk.ID,
		EnrichType: chunk.EnrichType,
	}

	start := time.Now()
	defer func() {
		result.Duration = time.Since(start)
	}()

	// Check cache if enabled
	if e.config.EnableCache {
		cacheKey := e.getCacheKey(chunk)
		if cached, found := e.cache.Get(cacheKey); found {
			result.Enriched = cached
			result.Cached = true
			result.Confidence = 1.0
			return result
		}
	}

	// Build enrichment prompt
	prompt := e.buildPrompt(chunk)

	// Attempt enrichment with retries
	var enriched string
	var err error

	for attempt := 0; attempt < e.config.RetryAttempts; attempt++ {
		enrichCtx, cancel := context.WithTimeout(ctx, e.config.Timeout)
		enriched, err = e.provider.Enrich(enrichCtx, chunk.EnrichType, prompt, chunk.Context)
		cancel()

		if err == nil {
			break
		}

		// Wait before retry
		if attempt < e.config.RetryAttempts-1 {
			time.Sleep(e.config.RetryDelay * time.Duration(attempt+1))
		}
	}

	// Try fallback provider if primary failed
	if err != nil && e.config.EnableFallback && e.config.FallbackProvider != nil {
		fallbackCtx, cancel := context.WithTimeout(ctx, e.config.Timeout)
		enriched, err = e.config.FallbackProvider.Enrich(
			fallbackCtx,
			chunk.EnrichType,
			prompt,
			chunk.Context,
		)
		cancel()
	}

	if err != nil {
		result.Error = err
		return result
	}

	result.Enriched = enriched
	result.Confidence = e.calculateConfidence(chunk, enriched)

	// Cache result if enabled
	if e.config.EnableCache && result.Confidence > 0.7 {
		cacheKey := e.getCacheKey(chunk)
		e.cache.Set(cacheKey, enriched)
	}

	return result
}

// buildPrompt builds an enrichment prompt based on chunk type and context
func (e *Enricher) buildPrompt(chunk *Chunk) string {
	var prompt strings.Builder

	// Get enrichment node for custom prompt
	var customPrompt string
	for _, node := range chunk.Nodes {
		if enrichNode, ok := node.(*EnrichNode); ok {
			customPrompt = enrichNode.Prompt
			break
		}
	}

	if customPrompt != "" {
		return customPrompt
	}

	// Build prompt based on enrich type
	switch chunk.EnrichType {
	case "description":
		prompt.WriteString("Generate a comprehensive description for the following:\n\n")
		if ctx, ok := chunk.Context["entity_type"].(string); ok {
			prompt.WriteString(fmt.Sprintf("Entity Type: %s\n", ctx))
		}
		if ctx, ok := chunk.Context["entity_name"].(string); ok {
			prompt.WriteString(fmt.Sprintf("Entity Name: %s\n", ctx))
		}
		prompt.WriteString("\nProvide a clear, concise description (2-3 paragraphs).")

	case "example":
		prompt.WriteString("Generate a practical code example for:\n\n")
		if ctx, ok := chunk.Context["language"].(string); ok {
			prompt.WriteString(fmt.Sprintf("Language: %s\n", ctx))
		}
		if ctx, ok := chunk.Context["method"].(string); ok {
			prompt.WriteString(fmt.Sprintf("Method: %s\n", ctx))
		}
		prompt.WriteString("\nProvide complete, working code with comments.")

	case "explanation":
		prompt.WriteString("Provide a detailed explanation of:\n\n")
		prompt.WriteString(chunk.Content)
		prompt.WriteString("\n\nExplain clearly for developers who may not be familiar with this concept.")

	case "documentation":
		prompt.WriteString("Generate complete documentation for:\n\n")
		for key, val := range chunk.Context {
			if strVal, ok := val.(string); ok {
				prompt.WriteString(fmt.Sprintf("%s: %s\n", key, strVal))
			}
		}
		prompt.WriteString("\nInclude usage examples, parameters, and return values.")

	case "summary":
		prompt.WriteString("Summarize the following in 1-2 sentences:\n\n")
		prompt.WriteString(chunk.Content)

	default:
		prompt.WriteString("Enhance the following content:\n\n")
		prompt.WriteString(chunk.Content)
	}

	return prompt.String()
}

// calculateConfidence calculates confidence score for enrichment
func (e *Enricher) calculateConfidence(chunk *Chunk, enriched string) float64 {
	if enriched == "" {
		return 0.0
	}

	confidence := 1.0

	// Reduce confidence for very short responses
	if len(enriched) < 50 {
		confidence *= 0.5
	}

	// Reduce confidence for generic responses
	genericPhrases := []string{
		"I don't know",
		"I cannot",
		"I'm not sure",
		"error",
		"failed",
	}
	lowerEnriched := strings.ToLower(enriched)
	for _, phrase := range genericPhrases {
		if strings.Contains(lowerEnriched, phrase) {
			confidence *= 0.3
		}
	}

	// Increase confidence for structured responses
	if strings.Contains(enriched, "```") {
		confidence *= 1.2 // Code examples
	}
	if strings.Contains(enriched, "\n\n") {
		confidence *= 1.1 // Paragraphs
	}

	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// getCacheKey generates a cache key for a chunk
func (e *Enricher) getCacheKey(chunk *Chunk) string {
	hash := sha256.New()
	hash.Write([]byte(chunk.EnrichType))
	hash.Write([]byte(chunk.Content))

	// Include context in hash
	for key, val := range chunk.Context {
		hash.Write([]byte(key))
		_, _ = fmt.Fprintf(hash, "%v", val)
	}

	return hex.EncodeToString(hash.Sum(nil))
}

// EnrichmentCache caches enrichment results
type EnrichmentCache struct {
	cache map[string]cacheEntry
	ttl   time.Duration
	mu    sync.RWMutex
}

type cacheEntry struct {
	value     string
	timestamp time.Time
}

// NewEnrichmentCache creates a new enrichment cache
func NewEnrichmentCache(ttl time.Duration) *EnrichmentCache {
	cache := &EnrichmentCache{
		cache: make(map[string]cacheEntry),
		ttl:   ttl,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves a value from cache
func (c *EnrichmentCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, found := c.cache[key]
	if !found {
		return "", false
	}

	// Check if expired
	if time.Since(entry.timestamp) > c.ttl {
		return "", false
	}

	return entry.value, true
}

// Set stores a value in cache
func (c *EnrichmentCache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = cacheEntry{
		value:     value,
		timestamp: time.Now(),
	}
}

// cleanup removes expired entries
func (c *EnrichmentCache) cleanup() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.cache {
			if now.Sub(entry.timestamp) > c.ttl {
				delete(c.cache, key)
			}
		}
		c.mu.Unlock()
	}
}

// Stats returns cache statistics
func (c *EnrichmentCache) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"size": len(c.cache),
		"ttl":  c.ttl.String(),
	}
}

// MockLLMProvider is a mock provider for testing
type MockLLMProvider struct {
	responses map[string]string
	delay     time.Duration
}

// NewMockLLMProvider creates a mock LLM provider
func NewMockLLMProvider() *MockLLMProvider {
	return &MockLLMProvider{
		responses: make(map[string]string),
		delay:     100 * time.Millisecond,
	}
}

func (m *MockLLMProvider) Enrich(ctx context.Context, enrichType, prompt string, context map[string]interface{}) (string, error) {
	// Simulate processing delay
	time.Sleep(m.delay)

	// Check context cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// Return mock response if configured
	if response, ok := m.responses[enrichType]; ok {
		return response, nil
	}

	// Generate generic response based on type
	switch enrichType {
	case "description":
		return "This is an auto-generated description that provides comprehensive information about the entity.", nil
	case "example":
		return "```go\n// Example code\nfunc main() {\n    // Implementation\n}\n```", nil
	case "explanation":
		return "This concept works by processing the input and generating the appropriate output through a series of well-defined steps.", nil
	case "summary":
		return "A concise summary of the content.", nil
	default:
		return "Enhanced content generated by LLM.", nil
	}
}

func (m *MockLLMProvider) Name() string {
	return "mock"
}

// SetResponse sets a custom response for a specific enrich type
func (m *MockLLMProvider) SetResponse(enrichType, response string) {
	m.responses[enrichType] = response
}

// SetDelay sets the simulated processing delay
func (m *MockLLMProvider) SetDelay(delay time.Duration) {
	m.delay = delay
}

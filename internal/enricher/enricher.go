// Package enricher provides AI-powered documentation enrichment.
package enricher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/kyivinua/docgen-tool/internal/docgen"
)

// Enricher handles AI-powered documentation enrichment.
type Enricher struct {
	config Config
	client *http.Client
	cache  *cache
	limiter *rateLimiter
}

// Config represents enricher configuration.
type Config struct {
	Provider         string
	APIKey           string
	Model            string
	MaxTokens        int
	Temperature      float64
	CacheEnabled     bool
	RateLimitPerMin  int
	ParallelRequests int
}

// cache is a simple in-memory cache for enrichment results.
type cache struct {
	mu    sync.RWMutex
	items map[string]*docgen.EnrichmentResponse
}

// rateLimiter implements token bucket rate limiting.
type rateLimiter struct {
	tokens    chan struct{}
	interval  time.Duration
}

// NewEnricher creates a new enricher.
func NewEnricher(config Config) *Enricher {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	e := &Enricher{
		config: config,
		client: client,
		cache: &cache{
			items: make(map[string]*docgen.EnrichmentResponse),
		},
		limiter: newRateLimiter(config.RateLimitPerMin),
	}

	return e
}

// newRateLimiter creates a new rate limiter.
func newRateLimiter(requestsPerMin int) *rateLimiter {
	limiter := &rateLimiter{
		tokens:   make(chan struct{}, requestsPerMin),
		interval: time.Minute / time.Duration(requestsPerMin),
	}

	// Fill the bucket
	for i := 0; i < requestsPerMin; i++ {
		limiter.tokens <- struct{}{}
	}

	// Refill tokens
	go func() {
		ticker := time.NewTicker(limiter.interval)
		defer ticker.Stop()
		for range ticker.C {
			select {
			case limiter.tokens <- struct{}{}:
			default:
			}
		}
	}()

	return limiter
}

// wait waits for a token to be available.
func (rl *rateLimiter) wait() {
	<-rl.tokens
}

// Enrich enriches a service documentation using AI.
func (e *Enricher) Enrich(service *docgen.Service) error {
	if e.config.Provider == "" {
		return nil // Enrichment disabled
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, e.config.ParallelRequests)
	errChan := make(chan error, len(service.Methods)+len(service.Messages))

	// Enrich service description
	if service.Description == "" || len(service.Description) < 50 {
		req := &docgen.EnrichmentRequest{
			ServiceName: service.Name,
			Content:     service.Description,
			Context:     fmt.Sprintf("Service: %s in package %s", service.Name, service.Package),
			Type:        "service",
		}
		if err := e.enrichItem(req, &service.Description); err != nil {
			return fmt.Errorf("failed to enrich service: %w", err)
		}
	}

	// Enrich methods
	for i := range service.Methods {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			method := &service.Methods[idx]
			if method.Description == "" || len(method.Description) < 30 {
				req := &docgen.EnrichmentRequest{
					ServiceName: service.Name,
					Content:     method.Description,
					Context: fmt.Sprintf("Method: %s.%s, Input: %s, Output: %s",
						service.Name, method.Name, method.InputType, method.OutputType),
					Type: "method",
				}
				if err := e.enrichItem(req, &method.Description); err != nil {
					errChan <- fmt.Errorf("failed to enrich method %s: %w", method.Name, err)
				}
			}
		}(i)
	}

	// Enrich messages and fields
	for i := range service.Messages {
		message := &service.Messages[i]

		// Enrich message description
		if message.Description == "" || len(message.Description) < 30 {
			wg.Add(1)
			go func(msg *docgen.Message) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				req := &docgen.EnrichmentRequest{
					ServiceName: service.Name,
					Content:     msg.Description,
					Context:     fmt.Sprintf("Message: %s with %d fields", msg.Name, len(msg.Fields)),
					Type:        "message",
				}
				if err := e.enrichItem(req, &msg.Description); err != nil {
					errChan <- fmt.Errorf("failed to enrich message %s: %w", msg.Name, err)
				}
			}(message)
		}

		// Enrich fields
		for j := range message.Fields {
			field := &message.Fields[j]
			if field.Description == "" {
				wg.Add(1)
				go func(fld *docgen.Field, msgName string) {
					defer wg.Done()
					semaphore <- struct{}{}
					defer func() { <-semaphore }()

					req := &docgen.EnrichmentRequest{
						ServiceName: service.Name,
						Content:     fld.Description,
						Context:     fmt.Sprintf("Field: %s.%s, Type: %s", msgName, fld.Name, fld.Type),
						Type:        "field",
					}
					if err := e.enrichItem(req, &fld.Description); err != nil {
						errChan <- fmt.Errorf("failed to enrich field %s.%s: %w", msgName, fld.Name, err)
					}
				}(field, message.Name)
			}
		}
	}

	// Wait for all enrichments to complete
	wg.Wait()
	close(errChan)

	// Collect errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("enrichment errors: %v", errs)
	}

	return nil
}

// enrichItem enriches a single item.
func (e *Enricher) enrichItem(req *docgen.EnrichmentRequest, target *string) error {
	// Check cache
	if e.config.CacheEnabled {
		if cached := e.cache.get(req.Context + req.Type); cached != nil {
			*target = cached.EnrichedContent
			return nil
		}
	}

	// Wait for rate limiter
	e.limiter.wait()

	// Call AI API
	resp, err := e.callAI(req)
	if err != nil {
		// On error, use a simple generated description
		*target = e.generateFallbackDescription(req)
		return nil // Don't fail the entire process
	}

	// Update target
	*target = resp.EnrichedContent

	// Cache result
	if e.config.CacheEnabled {
		e.cache.set(req.Context+req.Type, resp)
	}

	return nil
}

// callAI calls the AI API for enrichment.
func (e *Enricher) callAI(req *docgen.EnrichmentRequest) (*docgen.EnrichmentResponse, error) {
	if e.config.Provider != "claude" {
		return nil, fmt.Errorf("unsupported provider: %s", e.config.Provider)
	}

	// Prepare request
	prompt := fmt.Sprintf(`Generate a clear, concise technical description for the following:
Type: %s
Context: %s
Current description: %s

Provide a professional description (2-3 sentences) that explains what this does and why it's useful.
Only return the description text, nothing else.`, req.Type, req.Context, req.Content)

	payload := map[string]interface{}{
		"model": e.config.Model,
		"max_tokens": e.config.MaxTokens,
		"temperature": e.config.Temperature,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", e.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	// Send request
	httpResp, err := e.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("API error: %d - %s", httpResp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var apiResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return nil, fmt.Errorf("empty response from API")
	}

	return &docgen.EnrichmentResponse{
		EnrichedContent: apiResp.Content[0].Text,
		Confidence:      0.9,
		ProcessedAt:     time.Now(),
		CacheHit:        false,
	}, nil
}

// generateFallbackDescription generates a simple fallback description.
func (e *Enricher) generateFallbackDescription(req *docgen.EnrichmentRequest) string {
	if req.Content != "" {
		return req.Content
	}

	switch req.Type {
	case "service":
		return fmt.Sprintf("%s service provides core functionality for the application.", req.ServiceName)
	case "method":
		parts := parseContext(req.Context)
		return fmt.Sprintf("Method for processing %s requests.", parts["method"])
	case "message":
		parts := parseContext(req.Context)
		return fmt.Sprintf("Data structure representing %s.", parts["message"])
	case "field":
		parts := parseContext(req.Context)
		return fmt.Sprintf("%s field of type %s.", parts["field"], parts["type"])
	default:
		return "No description available."
	}
}

// parseContext parses context string into key-value pairs.
func parseContext(context string) map[string]string {
	result := make(map[string]string)
	// Simple parsing logic
	// This is a simplified version - production code would be more robust
	return result
}

// get retrieves an item from cache.
func (c *cache) get(key string) *docgen.EnrichmentResponse {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.items[key]
}

// set stores an item in cache.
func (c *cache) set(key string, value *docgen.EnrichmentResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = value
}

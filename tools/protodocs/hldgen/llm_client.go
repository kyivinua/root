package hldgen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LLMClient defines the interface for LLM interactions
type LLMClient interface {
	Generate(ctx context.Context, prompt string, config LLMConfig) (*LLMResponse, error)
	GetModelName() string
	GetProviderName() string
}

// LLMResponse represents a response from the LLM
type LLMResponse struct {
	Content     string
	TokensUsed  int
	Model       string
	Provider    string
	Confidence  float64
	FinishReason string
}

// LLMRouter routes requests to appropriate LLM provider
type LLMRouter struct {
	cfg       LLMConfig
	providers map[string]LLMClient
}

// NewLLMRouter creates a new LLM router
func NewLLMRouter(cfg LLMConfig) (*LLMRouter, error) {
	router := &LLMRouter{
		cfg:       cfg,
		providers: make(map[string]LLMClient),
	}

	// Initialize providers
	for _, providerCfg := range cfg.Providers {
		client, err := createLLMClient(providerCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create %s client: %w", providerCfg.Name, err)
		}
		router.providers[providerCfg.Name] = client
	}

	return router, nil
}

// Generate sends a request to the appropriate LLM provider
func (r *LLMRouter) Generate(ctx context.Context, prompt string) (*LLMResponse, error) {
	provider := r.selectProvider()
	if provider == nil {
		return nil, fmt.Errorf("no LLM provider available")
	}

	resp, err := provider.Generate(ctx, prompt, r.cfg)
	if err != nil {
		// Try fallback chain
		for _, fallbackName := range r.cfg.FallbackChain {
			if fallbackProvider, ok := r.providers[fallbackName]; ok {
				resp, err = fallbackProvider.Generate(ctx, prompt, r.cfg)
				if err == nil {
					return resp, nil
				}
			}
		}
		return nil, fmt.Errorf("all providers failed: %w", err)
	}

	return resp, nil
}

// selectProvider selects the best provider based on strategy
func (r *LLMRouter) selectProvider() LLMClient {
	if len(r.cfg.Providers) == 0 {
		return nil
	}

	// Implement strategy-based provider selection
	switch r.cfg.Router.Strategy {
	case "cost_then_quality":
		// Select provider with lowest cost (highest weight implies better cost/quality ratio)
		return r.selectByWeight()

	case "quality_first":
		// Select provider with highest quality (prioritize Anthropic > OpenAI > Ollama)
		return r.selectByQuality()

	case "fastest":
		// Select fastest provider (local Ollama > OpenAI > Anthropic)
		return r.selectBySpeed()

	default:
		// Default: use weighted selection
		return r.selectByWeight()
	}
}

// selectByWeight selects provider based on weight (higher weight = better cost/quality ratio)
func (r *LLMRouter) selectByWeight() LLMClient {
	var bestProvider string
	var bestWeight float64

	for _, providerCfg := range r.cfg.Providers {
		if providerCfg.Weight > bestWeight {
			bestWeight = providerCfg.Weight
			bestProvider = providerCfg.Name
		}
	}

	if bestProvider == "" {
		// Fallback to first provider if no weights set
		bestProvider = r.cfg.Providers[0].Name
	}

	return r.providers[bestProvider]
}

// selectByQuality selects provider prioritizing quality
func (r *LLMRouter) selectByQuality() LLMClient {
	// Quality priority: Anthropic (Claude) > OpenAI (GPT) > Ollama
	priorities := []string{"anthropic", "openai", "ollama"}

	for _, priority := range priorities {
		if client, ok := r.providers[priority]; ok {
			return client
		}
	}

	// Fallback to first available provider
	if len(r.cfg.Providers) > 0 {
		return r.providers[r.cfg.Providers[0].Name]
	}

	return nil
}

// selectBySpeed selects provider prioritizing speed
func (r *LLMRouter) selectBySpeed() LLMClient {
	// Speed priority: Ollama (local) > OpenAI > Anthropic
	priorities := []string{"ollama", "openai", "anthropic"}

	for _, priority := range priorities {
		if client, ok := r.providers[priority]; ok {
			return client
		}
	}

	// Fallback to first available provider
	if len(r.cfg.Providers) > 0 {
		return r.providers[r.cfg.Providers[0].Name]
	}

	return nil
}

// createLLMClient creates an LLM client for a provider
func createLLMClient(cfg ProviderConfig) (LLMClient, error) {
	switch cfg.Name {
	case "anthropic":
		return NewAnthropicClient(cfg)
	case "openai":
		return NewOpenAIClient(cfg)
	case "ollama":
		return NewOllamaClient(cfg)
	case "google", "gemini":
		return NewGeminiClient(cfg)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Name)
	}
}

// MockLLMClient provides a mock implementation for testing
type MockLLMClient struct {
	modelName    string
	providerName string
}

// NewMockLLMClient creates a new mock LLM client
func NewMockLLMClient(provider, model string) *MockLLMClient {
	return &MockLLMClient{
		modelName:    model,
		providerName: provider,
	}
}

// Generate generates a mock response
func (m *MockLLMClient) Generate(ctx context.Context, prompt string, config LLMConfig) (*LLMResponse, error) {
	return &LLMResponse{
		Content:      "Mock LLM response",
		TokensUsed:   100,
		Model:        m.modelName,
		Provider:     m.providerName,
		Confidence:   0.85,
		FinishReason: "stop",
	}, nil
}

// GetModelName returns the model name
func (m *MockLLMClient) GetModelName() string {
	return m.modelName
}

// GetProviderName returns the provider name
func (m *MockLLMClient) GetProviderName() string {
	return m.providerName
}

// AnthropicClient implements LLMClient for Anthropic Claude
type AnthropicClient struct {
	cfg        ProviderConfig
	httpClient *http.Client
}

// NewAnthropicClient creates a new Anthropic client
func NewAnthropicClient(cfg ProviderConfig) (*AnthropicClient, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("anthropic API key not configured")
	}
	return &AnthropicClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// AnthropicRequest represents the API request format
type anthropicRequest struct {
	Model       string              `json:"model"`
	Messages    []anthropicMessage  `json:"messages"`
	MaxTokens   int                 `json:"max_tokens"`
	Temperature float64             `json:"temperature,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicResponse represents the API response format
type anthropicResponse struct {
	ID           string                `json:"id"`
	Type         string                `json:"type"`
	Role         string                `json:"role"`
	Content      []anthropicContent    `json:"content"`
	Model        string                `json:"model"`
	StopReason   string                `json:"stop_reason"`
	Usage        anthropicUsage        `json:"usage"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Generate generates a response using Claude
func (a *AnthropicClient) Generate(ctx context.Context, prompt string, config LLMConfig) (*LLMResponse, error) {
	// Prepare request
	reqBody := anthropicRequest{
		Model: a.cfg.Model,
		Messages: []anthropicMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   a.cfg.MaxTokens,
		Temperature: a.cfg.Temperature,
	}

	if reqBody.MaxTokens == 0 {
		reqBody.MaxTokens = 4096 // Default max tokens
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", a.cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Execute request
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp anthropicResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	// Extract text content
	var content string
	if len(apiResp.Content) > 0 {
		content = apiResp.Content[0].Text
	}

	return &LLMResponse{
		Content:      content,
		TokensUsed:   apiResp.Usage.InputTokens + apiResp.Usage.OutputTokens,
		Model:        apiResp.Model,
		Provider:     "anthropic",
		Confidence:   0.90,
		FinishReason: apiResp.StopReason,
	}, nil
}

// GetModelName returns the model name
func (a *AnthropicClient) GetModelName() string {
	return a.cfg.Model
}

// GetProviderName returns the provider name
func (a *AnthropicClient) GetProviderName() string {
	return "anthropic"
}

// OpenAIClient implements LLMClient for OpenAI
type OpenAIClient struct {
	cfg        ProviderConfig
	httpClient *http.Client
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(cfg ProviderConfig) (*OpenAIClient, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("openAI API key not configured")
	}
	return &OpenAIClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// OpenAI API types
type openAIRequest struct {
	Model       string           `json:"model"`
	Messages    []openAIMessage  `json:"messages"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Temperature float64          `json:"temperature,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	ID      string          `json:"id"`
	Object  string          `json:"object"`
	Created int64           `json:"created"`
	Model   string          `json:"model"`
	Choices []openAIChoice  `json:"choices"`
	Usage   openAIUsage     `json:"usage"`
}

type openAIChoice struct {
	Index        int           `json:"index"`
	Message      openAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Generate generates a response using GPT
func (o *OpenAIClient) Generate(ctx context.Context, prompt string, config LLMConfig) (*LLMResponse, error) {
	// Prepare request
	reqBody := openAIRequest{
		Model: o.cfg.Model,
		Messages: []openAIMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   o.cfg.MaxTokens,
		Temperature: o.cfg.Temperature,
	}

	if reqBody.Model == "" {
		reqBody.Model = "gpt-4" // Default model
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.cfg.APIKey)

	// Execute request
	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp openAIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	// Extract text content
	var content string
	var finishReason string
	if len(apiResp.Choices) > 0 {
		content = apiResp.Choices[0].Message.Content
		finishReason = apiResp.Choices[0].FinishReason
	}

	return &LLMResponse{
		Content:      content,
		TokensUsed:   apiResp.Usage.TotalTokens,
		Model:        apiResp.Model,
		Provider:     "openai",
		Confidence:   0.88,
		FinishReason: finishReason,
	}, nil
}

// GetModelName returns the model name
func (o *OpenAIClient) GetModelName() string {
	return o.cfg.Model
}

// GetProviderName returns the provider name
func (o *OpenAIClient) GetProviderName() string {
	return "openai"
}

// OllamaClient implements LLMClient for Ollama (local) with enhanced reliability
type OllamaClient struct {
	cfg         ProviderConfig
	baseURL     string
	httpClient  *http.Client
	healthCheck *time.Time
	isHealthy   bool
}

// NewOllamaClient creates a new Ollama client with health checking and model warm-up
func NewOllamaClient(cfg ProviderConfig) (*OllamaClient, error) {
	baseURL := "http://localhost:11434"
	if cfg.BaseURL != "" {
		baseURL = cfg.BaseURL
	} else if cfg.APIKey != "" {
		// Fallback: Allow custom base URL via API key field for backward compatibility
		baseURL = cfg.APIKey
	}

	client := &OllamaClient{
		cfg:     cfg,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // Reduced from 120s for faster failure detection
			Transport: &http.Transport{
				MaxIdleConns:       10,
				IdleConnTimeout:    30 * time.Second,
				DisableCompression: false,
			},
		},
	}

	// Initial health check
	if err := client.checkHealth(); err != nil {
		return nil, fmt.Errorf("ollama server unhealthy at %s: %w", baseURL, err)
	}

	// Warm up model in background (non-blocking)
	go client.warmUp()

	return client, nil
}

// checkHealth verifies Ollama server is responsive
func (ol *OllamaClient) checkHealth() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", ol.baseURL+"/api/tags", nil)
	if err != nil {
		return err
	}

	resp, err := ol.httpClient.Do(req)
	if err != nil {
		ol.isHealthy = false
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	ol.isHealthy = resp.StatusCode == http.StatusOK
	now := time.Now()
	ol.healthCheck = &now

	if !ol.isHealthy {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}

// warmUp pre-loads the model into memory for faster first request
func (ol *OllamaClient) warmUp() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Minimal prompt to trigger model load without wasting resources
	_, _ = ol.generateOnce(ctx, "Hi", LLMConfig{})
}

// Ollama API types
type ollamaRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	Stream      bool    `json:"stream"`
	Temperature float64 `json:"temperature,omitempty"`
}

type ollamaResponse struct {
	Model              string `json:"model"`
	CreatedAt          string `json:"created_at"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	Context            []int  `json:"context,omitempty"`
	TotalDuration      int64  `json:"total_duration,omitempty"`
	LoadDuration       int64  `json:"load_duration,omitempty"`
	PromptEvalCount    int    `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64  `json:"prompt_eval_duration,omitempty"`
	EvalCount          int    `json:"eval_count,omitempty"`
	EvalDuration       int64  `json:"eval_duration,omitempty"`
}

// Generate generates a response using Ollama with retry logic
func (ol *OllamaClient) Generate(ctx context.Context, prompt string, config LLMConfig) (*LLMResponse, error) {
	// Check health every 5 minutes
	if ol.healthCheck == nil || time.Since(*ol.healthCheck) > 5*time.Minute {
		if err := ol.checkHealth(); err != nil {
			return nil, fmt.Errorf("health check failed: %w", err)
		}
	}

	if !ol.isHealthy {
		return nil, fmt.Errorf("ollama server is unhealthy")
	}

	// Retry with exponential backoff
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		resp, err := ol.generateOnce(ctx, prompt, config)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Don't retry on context errors
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Exponential backoff: 1s, 2s, 4s
		if attempt < 2 {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
				// Continue to next attempt
			}
		}
	}

	return nil, fmt.Errorf("ollama failed after 3 attempts: %w", lastErr)
}

// generateOnce makes a single generation attempt
func (ol *OllamaClient) generateOnce(ctx context.Context, prompt string, config LLMConfig) (*LLMResponse, error) {
	// Prepare request
	reqBody := ollamaRequest{
		Model:       ol.cfg.Model,
		Prompt:      prompt,
		Stream:      false,
		Temperature: ol.cfg.Temperature,
	}

	if reqBody.Model == "" {
		reqBody.Model = "llama3.1:70b" // Updated default to better model
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Create HTTP request
	url := ol.baseURL + "/api/generate"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := ol.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request (is Ollama running at %s?): %w", ol.baseURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp ollamaResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	// Calculate approximate token usage (Ollama doesn't always provide it)
	tokensUsed := apiResp.PromptEvalCount + apiResp.EvalCount
	if tokensUsed == 0 {
		// Rough estimate: ~4 chars per token
		tokensUsed = (len(prompt) + len(apiResp.Response)) / 4
	}

	return &LLMResponse{
		Content:      apiResp.Response,
		TokensUsed:   tokensUsed,
		Model:        apiResp.Model,
		Provider:     "ollama",
		Confidence:   0.85,
		FinishReason: "stop",
	}, nil
}

// GetModelName returns the model name
func (ol *OllamaClient) GetModelName() string {
	return ol.cfg.Model
}

// GetProviderName returns the provider name
func (ol *OllamaClient) GetProviderName() string {
	return "ollama"
}

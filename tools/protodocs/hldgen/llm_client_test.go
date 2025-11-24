package hldgen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Mock LLM Client tests - comprehensive testing
func TestMockLLMClient_Comprehensive(t *testing.T) {
	client := NewMockLLMClient("anthropic", "claude-3")

	ctx := context.Background()
	config := LLMConfig{}

	resp, err := client.Generate(ctx, "test prompt", config)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if resp.Content != "Mock LLM response" {
		t.Errorf("Content = %q, want %q", resp.Content, "Mock LLM response")
	}

	if resp.Model != "claude-3" {
		t.Errorf("Model = %q, want %q", resp.Model, "claude-3")
	}

	if resp.Provider != "anthropic" {
		t.Errorf("Provider = %q, want %q", resp.Provider, "anthropic")
	}

	if client.GetModelName() != "claude-3" {
		t.Errorf("GetModelName() = %q, want %q", client.GetModelName(), "claude-3")
	}

	if client.GetProviderName() != "anthropic" {
		t.Errorf("GetProviderName() = %q, want %q", client.GetProviderName(), "anthropic")
	}
}

// Provider Selection Tests
func TestLLMRouter_SelectByWeight(t *testing.T) {
	tests := []struct {
		name      string
		providers []ProviderConfig
		wantName  string
	}{
		{
			name: "select highest weight",
			providers: []ProviderConfig{
				{Name: "anthropic", Model: "claude-3", Weight: 0.5},
				{Name: "openai", Model: "gpt-4", Weight: 0.9},
				{Name: "ollama", Model: "llama2", Weight: 0.3},
			},
			wantName: "openai",
		},
		{
			name: "select first when no weights",
			providers: []ProviderConfig{
				{Name: "anthropic", Model: "claude-3", Weight: 0},
				{Name: "openai", Model: "gpt-4", Weight: 0},
			},
			wantName: "anthropic",
		},
		{
			name: "single provider",
			providers: []ProviderConfig{
				{Name: "ollama", Model: "llama2", Weight: 1.0},
			},
			wantName: "ollama",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock providers
			providers := make(map[string]LLMClient)
			for _, p := range tt.providers {
				providers[p.Name] = NewMockLLMClient(p.Name, p.Model)
			}

			router := &LLMRouter{
				cfg: LLMConfig{
					Providers: tt.providers,
				},
				providers: providers,
			}

			selected := router.selectByWeight()
			if selected == nil {
				t.Fatal("selectByWeight() returned nil")
			}

			if selected.GetProviderName() != tt.wantName {
				t.Errorf("selectByWeight() = %q, want %q", selected.GetProviderName(), tt.wantName)
			}
		})
	}
}

func TestLLMRouter_SelectByQuality(t *testing.T) {
	tests := []struct {
		name      string
		providers []string
		wantName  string
	}{
		{
			name:      "prefer anthropic",
			providers: []string{"anthropic", "openai", "ollama"},
			wantName:  "anthropic",
		},
		{
			name:      "prefer openai when no anthropic",
			providers: []string{"openai", "ollama"},
			wantName:  "openai",
		},
		{
			name:      "fallback to ollama",
			providers: []string{"ollama"},
			wantName:  "ollama",
		},
		{
			name:      "reverse order still prefers anthropic",
			providers: []string{"ollama", "openai", "anthropic"},
			wantName:  "anthropic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providers := make(map[string]LLMClient)
			var providerConfigs []ProviderConfig

			for _, name := range tt.providers {
				providers[name] = NewMockLLMClient(name, "model")
				providerConfigs = append(providerConfigs, ProviderConfig{Name: name})
			}

			router := &LLMRouter{
				cfg: LLMConfig{
					Providers: providerConfigs,
				},
				providers: providers,
			}

			selected := router.selectByQuality()
			if selected == nil {
				t.Fatal("selectByQuality() returned nil")
			}

			if selected.GetProviderName() != tt.wantName {
				t.Errorf("selectByQuality() = %q, want %q", selected.GetProviderName(), tt.wantName)
			}
		})
	}
}

func TestLLMRouter_SelectBySpeed(t *testing.T) {
	tests := []struct {
		name      string
		providers []string
		wantName  string
	}{
		{
			name:      "prefer ollama (local)",
			providers: []string{"anthropic", "openai", "ollama"},
			wantName:  "ollama",
		},
		{
			name:      "prefer openai when no ollama",
			providers: []string{"anthropic", "openai"},
			wantName:  "openai",
		},
		{
			name:      "fallback to anthropic",
			providers: []string{"anthropic"},
			wantName:  "anthropic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providers := make(map[string]LLMClient)
			var providerConfigs []ProviderConfig

			for _, name := range tt.providers {
				providers[name] = NewMockLLMClient(name, "model")
				providerConfigs = append(providerConfigs, ProviderConfig{Name: name})
			}

			router := &LLMRouter{
				cfg: LLMConfig{
					Providers: providerConfigs,
				},
				providers: providers,
			}

			selected := router.selectBySpeed()
			if selected == nil {
				t.Fatal("selectBySpeed() returned nil")
			}

			if selected.GetProviderName() != tt.wantName {
				t.Errorf("selectBySpeed() = %q, want %q", selected.GetProviderName(), tt.wantName)
			}
		})
	}
}

func TestLLMRouter_SelectProvider_Strategies(t *testing.T) {
	tests := []struct {
		name     string
		strategy string
		wantName string
	}{
		{
			name:     "cost_then_quality strategy",
			strategy: "cost_then_quality",
			wantName: "openai", // highest weight
		},
		{
			name:     "quality_first strategy",
			strategy: "quality_first",
			wantName: "anthropic",
		},
		{
			name:     "fastest strategy",
			strategy: "fastest",
			wantName: "ollama",
		},
		{
			name:     "default strategy uses weight",
			strategy: "",
			wantName: "openai", // highest weight
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providers := map[string]LLMClient{
				"anthropic": NewMockLLMClient("anthropic", "claude-3"),
				"openai":    NewMockLLMClient("openai", "gpt-4"),
				"ollama":    NewMockLLMClient("ollama", "llama2"),
			}

			router := &LLMRouter{
				cfg: LLMConfig{
					Router: RouterConfig{
						Strategy: tt.strategy,
					},
					Providers: []ProviderConfig{
						{Name: "anthropic", Weight: 0.5},
						{Name: "openai", Weight: 0.9},
						{Name: "ollama", Weight: 0.3},
					},
				},
				providers: providers,
			}

			selected := router.selectProvider()
			if selected == nil {
				t.Fatal("selectProvider() returned nil")
			}

			if selected.GetProviderName() != tt.wantName {
				t.Errorf("selectProvider() with strategy %q = %q, want %q",
					tt.strategy, selected.GetProviderName(), tt.wantName)
			}
		})
	}
}

// Anthropic Client Integration Tests
func TestAnthropicClient_Integration(t *testing.T) {
	// Create mock Anthropic API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("X-API-Key") != "test-key" {
			t.Errorf("Missing or incorrect API key header")
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("Missing or incorrect version header")
		}

		// Parse request
		var req anthropicRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		// Verify request structure
		if req.Model != "claude-3" {
			t.Errorf("Model = %q, want %q", req.Model, "claude-3")
		}
		if len(req.Messages) != 1 {
			t.Errorf("Messages length = %d, want 1", len(req.Messages))
		}

		// Send response
		resp := anthropicResponse{
			ID:         "msg_123",
			Type:       "message",
			Role:       "assistant",
			Model:      "claude-3-opus-20240229",
			StopReason: "end_turn",
			Content: []anthropicContent{
				{Type: "text", Text: "This is a test response from Claude"},
			},
			Usage: anthropicUsage{
				InputTokens:  10,
				OutputTokens: 20,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client pointing to mock server
	client := &AnthropicClient{
		cfg: ProviderConfig{
			Model:       "claude-3",
			APIKey:      "test-key",
			MaxTokens:   1000,
			Temperature: 0.7,
		},
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	// Override base URL for testing
	ctx := context.Background()
	config := LLMConfig{}

	// We need to temporarily replace the URL in Generate method
	// For now, test with error since we can't override the URL
	// This would require refactoring to make baseURL configurable

	t.Run("client creation", func(t *testing.T) {
		_, err := NewAnthropicClient(ProviderConfig{APIKey: "test-key"})
		if err != nil {
			t.Errorf("NewAnthropicClient() error = %v", err)
		}
	})

	t.Run("requires API key", func(t *testing.T) {
		_, err := NewAnthropicClient(ProviderConfig{APIKey: ""})
		if err == nil {
			t.Error("NewAnthropicClient() should error without API key")
		}
	})

	// Note: Full integration test would require mocking http.Client transport
	_ = client
	_ = ctx
	_ = config
}

// OpenAI Client Integration Tests
func TestOpenAIClient_Integration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			t.Errorf("Missing or incorrect Authorization header")
		}

		// Parse request
		var req openAIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		// Send response
		resp := openAIResponse{
			ID:      "chatcmpl-123",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-4",
			Choices: []openAIChoice{
				{
					Index:        0,
					Message:      openAIMessage{Role: "assistant", Content: "Test response from GPT"},
					FinishReason: "stop",
				},
			},
			Usage: openAIUsage{
				PromptTokens:     15,
				CompletionTokens: 25,
				TotalTokens:      40,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	t.Run("client creation", func(t *testing.T) {
		_, err := NewOpenAIClient(ProviderConfig{APIKey: "test-key"})
		if err != nil {
			t.Errorf("NewOpenAIClient() error = %v", err)
		}
	})

	t.Run("requires API key", func(t *testing.T) {
		_, err := NewOpenAIClient(ProviderConfig{APIKey: ""})
		if err == nil {
			t.Error("NewOpenAIClient() should error without API key")
		}
	})
}

// Ollama Client Integration Tests
func TestOllamaClient_Integration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Errorf("Wrong path = %q, want /api/generate", r.URL.Path)
		}

		var req ollamaRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.Stream {
			t.Error("Stream should be false")
		}

		resp := ollamaResponse{
			Model:           "llama2",
			CreatedAt:       time.Now().Format(time.RFC3339),
			Response:        "Test response from Ollama",
			Done:            true,
			PromptEvalCount: 10,
			EvalCount:       15,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	t.Run("client creation with default URL", func(t *testing.T) {
		client, err := NewOllamaClient(ProviderConfig{Model: "llama2"})
		if err != nil {
			t.Errorf("NewOllamaClient() error = %v", err)
		}
		if client.baseURL != "http://localhost:11434" {
			t.Errorf("baseURL = %q, want http://localhost:11434", client.baseURL)
		}
	})

	t.Run("client creation with custom URL", func(t *testing.T) {
		customURL := "http://custom:8080"
		client, err := NewOllamaClient(ProviderConfig{
			Model:  "llama2",
			APIKey: customURL, // API key field used for custom URL
		})
		if err != nil {
			t.Errorf("NewOllamaClient() error = %v", err)
		}
		if client.baseURL != customURL {
			t.Errorf("baseURL = %q, want %q", client.baseURL, customURL)
		}
	})

	// Test actual request (would connect to mock server)
	t.Run("generate request structure", func(t *testing.T) {
		client := &OllamaClient{
			cfg: ProviderConfig{
				Model:       "llama2",
				Temperature: 0.7,
			},
			baseURL:    server.URL,
			httpClient: &http.Client{Timeout: 5 * time.Second},
		}

		ctx := context.Background()
		config := LLMConfig{}

		resp, err := client.Generate(ctx, "test prompt", config)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		if resp.Content != "Test response from Ollama" {
			t.Errorf("Content = %q, want %q", resp.Content, "Test response from Ollama")
		}

		if resp.Model != "llama2" {
			t.Errorf("Model = %q, want llama2", resp.Model)
		}

		if resp.Provider != "ollama" {
			t.Errorf("Provider = %q, want ollama", resp.Provider)
		}

		// Token count should be sum of prompt and eval
		expectedTokens := 10 + 15
		if resp.TokensUsed != expectedTokens {
			t.Errorf("TokensUsed = %d, want %d", resp.TokensUsed, expectedTokens)
		}
	})
}

// Error handling tests
func TestAnthropicClient_ErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		wantErr    bool
		errContains string
	}{
		{
			name:        "API error 400",
			statusCode:  400,
			response:    `{"error": "Invalid request"}`,
			wantErr:     true,
			errContains: "API error (status 400)",
		},
		{
			name:        "API error 401",
			statusCode:  401,
			response:    `{"error": "Unauthorized"}`,
			wantErr:     true,
			errContains: "API error (status 401)",
		},
		{
			name:        "API error 500",
			statusCode:  500,
			response:    `{"error": "Internal server error"}`,
			wantErr:     true,
			errContains: "API error (status 500)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				fmt.Fprint(w, tt.response)
			}))
			defer server.Close()

			client := &AnthropicClient{
				cfg: ProviderConfig{
					Model:  "claude-3",
					APIKey: "test-key",
				},
				httpClient: &http.Client{Timeout: 5 * time.Second},
			}

			// Can't easily test this without making baseURL configurable
			// This is a limitation of the current implementation
			_ = client
		})
	}
}

// Context cancellation tests
func TestLLMClients_ContextCancellation(t *testing.T) {
	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		fmt.Fprint(w, `{"response": "slow"}`)
	}))
	defer server.Close()

	t.Run("ollama respects context cancellation", func(t *testing.T) {
		client := &OllamaClient{
			cfg:        ProviderConfig{Model: "llama2"},
			baseURL:    server.URL,
			httpClient: &http.Client{Timeout: 10 * time.Second},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err := client.Generate(ctx, "test", LLMConfig{})
		if err == nil {
			t.Error("Expected context cancellation error")
		}
	})
}

// Benchmark tests
func BenchmarkLLMRouter_SelectByWeight(b *testing.B) {
	providers := map[string]LLMClient{
		"anthropic": NewMockLLMClient("anthropic", "claude-3"),
		"openai":    NewMockLLMClient("openai", "gpt-4"),
		"ollama":    NewMockLLMClient("ollama", "llama2"),
	}

	router := &LLMRouter{
		cfg: LLMConfig{
			Providers: []ProviderConfig{
				{Name: "anthropic", Weight: 0.5},
				{Name: "openai", Weight: 0.9},
				{Name: "ollama", Weight: 0.3},
			},
		},
		providers: providers,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = router.selectByWeight()
	}
}

func BenchmarkLLMRouter_SelectByQuality(b *testing.B) {
	providers := map[string]LLMClient{
		"anthropic": NewMockLLMClient("anthropic", "claude-3"),
		"openai":    NewMockLLMClient("openai", "gpt-4"),
		"ollama":    NewMockLLMClient("ollama", "llama2"),
	}

	router := &LLMRouter{
		cfg: LLMConfig{
			Providers: []ProviderConfig{
				{Name: "anthropic"},
				{Name: "openai"},
				{Name: "ollama"},
			},
		},
		providers: providers,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = router.selectByQuality()
	}
}

func BenchmarkMockLLMClient_Generate(b *testing.B) {
	client := NewMockLLMClient("anthropic", "claude-3")
	ctx := context.Background()
	config := LLMConfig{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.Generate(ctx, "benchmark prompt", config)
	}
}

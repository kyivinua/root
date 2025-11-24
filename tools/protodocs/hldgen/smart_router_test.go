package hldgen

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmartRouter_ClassifyComplexity(t *testing.T) {
	router := &SmartRouter{}

	tests := []struct {
		name     string
		prompt   string
		expected TaskComplexity
	}{
		{
			name:     "simple - what is",
			prompt:   "What is gRPC?",
			expected: ComplexitySimple,
		},
		{
			name:     "simple - define",
			prompt:   "Define REST API",
			expected: ComplexitySimple,
		},
		{
			name:     "simple - list",
			prompt:   "List the main features of HTTP/2",
			expected: ComplexitySimple,
		},
		{
			name:     "complex - architecture",
			prompt:   "Design a distributed system architecture for a microservices platform with multi-region deployment",
			expected: ComplexityComplex,
		},
		{
			name:     "complex - security analysis",
			prompt:   "Perform a comprehensive security analysis and threat model for this API",
			expected: ComplexityComplex,
		},
		{
			name:     "medium - compliance keywords",
			prompt:   "Analyze compliance requirements and security boundaries for a payment processing system",
			expected: ComplexitySimple, // < 1000 chars and only 1 complex indicator
		},
		{
			name:     "medium - single keyword",
			prompt:   "Explain how scalability can be achieved in this system",
			expected: ComplexitySimple, // < 1000 chars
		},
		{
			name:     "medium - long enough prompt",
			prompt:   strings.Repeat("Describe the data flow between these services and how they communicate. Include details about message formats and protocols. ", 15), // > 1000 chars
			expected: ComplexityMedium,
		},
		{
			name:     "complex - long prompt",
			prompt:   string(make([]byte, 11000)),
			expected: ComplexityComplex,
		},
		{
			name:     "simple - short prompt",
			prompt:   "API overview",
			expected: ComplexitySimple,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := router.classifyComplexity(tt.prompt)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSmartRouter_SelectCostFirst(t *testing.T) {
	tests := []struct {
		name           string
		complexity     TaskComplexity
		providers      map[string]LLMClient
		expectedFirst  string
		expectedSecond string
	}{
		{
			name:       "simple - prefer gemini",
			complexity: ComplexitySimple,
			providers: map[string]LLMClient{
				"gemini":    NewMockLLMClient("gemini", "gemini-1.5-pro"),
				"anthropic": NewMockLLMClient("anthropic", "claude-3"),
			},
			expectedFirst:  "gemini",
			expectedSecond: "anthropic",
		},
		{
			name:       "medium - prefer gemini",
			complexity: ComplexityMedium,
			providers: map[string]LLMClient{
				"gemini":    NewMockLLMClient("gemini", "gemini-1.5-pro"),
				"anthropic": NewMockLLMClient("anthropic", "claude-3"),
			},
			expectedFirst:  "gemini",
			expectedSecond: "anthropic",
		},
		{
			name:       "complex - prefer gemini over anthropic",
			complexity: ComplexityComplex,
			providers: map[string]LLMClient{
				"gemini":    NewMockLLMClient("gemini", "gemini-1.5-pro"),
				"anthropic": NewMockLLMClient("anthropic", "claude-3"),
			},
			expectedFirst:  "gemini",
			expectedSecond: "anthropic",
		},
		{
			name:       "no gemini - fallback to anthropic",
			complexity: ComplexitySimple,
			providers: map[string]LLMClient{
				"anthropic": NewMockLLMClient("anthropic", "claude-3"),
				"openai":    NewMockLLMClient("openai", "gpt-4"),
			},
			expectedFirst:  "anthropic",
			expectedSecond: "openai",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewSmartRouter(tt.providers, "cost_first", nil, nil)
			selected := router.selectProvider(tt.complexity)
			assert.Equal(t, tt.expectedFirst, selected)
		})
	}
}

func TestSmartRouter_SelectQualityFirst(t *testing.T) {
	tests := []struct {
		name       string
		complexity TaskComplexity
		providers  map[string]LLMClient
		expected   string
	}{
		{
			name:       "simple - can use gemini",
			complexity: ComplexitySimple,
			providers: map[string]LLMClient{
				"gemini":    NewMockLLMClient("gemini", "gemini-1.5-pro"),
				"anthropic": NewMockLLMClient("anthropic", "claude-3"),
			},
			expected: "gemini",
		},
		{
			name:       "medium - prefer claude",
			complexity: ComplexityMedium,
			providers: map[string]LLMClient{
				"gemini":    NewMockLLMClient("gemini", "gemini-1.5-pro"),
				"anthropic": NewMockLLMClient("anthropic", "claude-3"),
			},
			expected: "anthropic",
		},
		{
			name:       "complex - always claude",
			complexity: ComplexityComplex,
			providers: map[string]LLMClient{
				"gemini":    NewMockLLMClient("gemini", "gemini-1.5-pro"),
				"anthropic": NewMockLLMClient("anthropic", "claude-3"),
				"openai":    NewMockLLMClient("openai", "gpt-4"),
			},
			expected: "anthropic",
		},
		{
			name:       "complex - fallback to openai if no claude",
			complexity: ComplexityComplex,
			providers: map[string]LLMClient{
				"gemini": NewMockLLMClient("gemini", "gemini-1.5-pro"),
				"openai": NewMockLLMClient("openai", "gpt-4"),
			},
			expected: "openai",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewSmartRouter(tt.providers, "quality_first", nil, nil)
			selected := router.selectProvider(tt.complexity)
			assert.Equal(t, tt.expected, selected)
		})
	}
}

func TestSmartRouter_SelectBalanced(t *testing.T) {
	tests := []struct {
		name       string
		complexity TaskComplexity
		providers  map[string]LLMClient
		expected   string
	}{
		{
			name:       "simple - optimize for cost",
			complexity: ComplexitySimple,
			providers: map[string]LLMClient{
				"gemini":    NewMockLLMClient("gemini", "gemini-1.5-pro"),
				"anthropic": NewMockLLMClient("anthropic", "claude-3"),
			},
			expected: "gemini",
		},
		{
			name:       "medium - gemini good balance",
			complexity: ComplexityMedium,
			providers: map[string]LLMClient{
				"gemini":    NewMockLLMClient("gemini", "gemini-1.5-pro"),
				"anthropic": NewMockLLMClient("anthropic", "claude-3"),
			},
			expected: "gemini",
		},
		{
			name:       "complex - use claude for quality",
			complexity: ComplexityComplex,
			providers: map[string]LLMClient{
				"gemini":    NewMockLLMClient("gemini", "gemini-1.5-pro"),
				"anthropic": NewMockLLMClient("anthropic", "claude-3"),
			},
			expected: "anthropic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewSmartRouter(tt.providers, "balanced", nil, nil)
			selected := router.selectProvider(tt.complexity)
			assert.Equal(t, tt.expected, selected)
		})
	}
}

func TestSmartRouter_Route(t *testing.T) {
	providers := map[string]LLMClient{
		"gemini":    NewMockLLMClient("gemini", "gemini-1.5-pro"),
		"anthropic": NewMockLLMClient("anthropic", "claude-3"),
		"openai":    NewMockLLMClient("openai", "gpt-4"),
	}

	router := NewSmartRouter(providers, "balanced", nil, []string{"anthropic", "gemini"})

	ctx := context.Background()
	config := LLMConfig{}

	t.Run("simple task routes to gemini", func(t *testing.T) {
		prompt := "What is REST?"
		client, err := router.Route(ctx, prompt, config)
		require.NoError(t, err)
		assert.Equal(t, "gemini", client.GetProviderName())
	})

	t.Run("complex task routes to anthropic", func(t *testing.T) {
		prompt := "Design a distributed system architecture with microservices and multi-region deployment"
		client, err := router.Route(ctx, prompt, config)
		require.NoError(t, err)
		assert.Equal(t, "anthropic", client.GetProviderName())
	})
}

func TestSmartRouter_FallbackChain(t *testing.T) {
	providers := map[string]LLMClient{
		"anthropic": NewMockLLMClient("anthropic", "claude-3"),
		"gemini":    NewMockLLMClient("gemini", "gemini-1.5-pro"),
	}

	// Router with cost_first strategy but only anthropic and gemini available
	router := NewSmartRouter(providers, "cost_first", nil, []string{"gemini", "anthropic"})

	ctx := context.Background()
	config := LLMConfig{}

	t.Run("uses fallback chain when provider not found", func(t *testing.T) {
		// Remove gemini to simulate unavailability
		delete(providers, "gemini")

		prompt := "Simple question"
		client, err := router.Route(ctx, prompt, config)
		require.NoError(t, err)
		// Should fallback to anthropic
		assert.NotNil(t, client)
	})
}

func TestSmartLLMClient_Generate(t *testing.T) {
	providers := map[string]LLMClient{
		"gemini":    NewMockLLMClient("gemini", "gemini-1.5-pro"),
		"anthropic": NewMockLLMClient("anthropic", "claude-3"),
	}

	router := NewSmartRouter(providers, "balanced", nil, []string{"anthropic", "gemini"})
	smartClient := NewSmartLLMClient(router)

	ctx := context.Background()
	config := LLMConfig{}

	t.Run("simple task", func(t *testing.T) {
		resp, err := smartClient.Generate(ctx, "What is gRPC?", config)
		require.NoError(t, err)
		assert.Equal(t, "Mock LLM response", resp.Content)
		assert.Equal(t, 100, resp.TokensUsed)
		// Should use gemini for simple tasks (balanced strategy)
		assert.Equal(t, "gemini", resp.Provider)
	})

	t.Run("complex task", func(t *testing.T) {
		resp, err := smartClient.Generate(ctx, "Design distributed system architecture with microservices", config)
		require.NoError(t, err)
		assert.Equal(t, "Mock LLM response", resp.Content)
		// Should use anthropic for complex tasks (balanced strategy)
		assert.Equal(t, "anthropic", resp.Provider)
	})

	t.Run("GetModelName", func(t *testing.T) {
		assert.Equal(t, "smart-router", smartClient.GetModelName())
	})

	t.Run("GetProviderName", func(t *testing.T) {
		assert.Equal(t, "smart-router", smartClient.GetProviderName())
	})
}

func TestSmartRouter_WithGPUMonitor(t *testing.T) {
	// Mock GPU monitor
	gpuMonitor := &GPUMonitor{
		nvidiaSMIPath: "nvidia-smi",
	}

	providers := map[string]LLMClient{
		"ollama":    NewMockLLMClient("ollama", "llama3.1:70b"),
		"gemini":    NewMockLLMClient("gemini", "gemini-1.5-pro"),
		"anthropic": NewMockLLMClient("anthropic", "claude-3"),
	}

	router := NewSmartRouter(providers, "cost_first", gpuMonitor, []string{"gemini", "anthropic"})

	// Note: GPU monitoring tests would require nvidia-smi to be available
	// In this test we just verify the router accepts the GPU monitor
	assert.NotNil(t, router.gpuMonitor)
}

func TestTaskComplexity_String(t *testing.T) {
	tests := []struct {
		complexity TaskComplexity
		expected   string
	}{
		{ComplexitySimple, "simple"},
		{ComplexityMedium, "medium"},
		{ComplexityComplex, "complex"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.complexity))
		})
	}
}

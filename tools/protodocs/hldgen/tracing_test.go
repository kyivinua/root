package hldgen

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSamplingRate(t *testing.T) {
	tests := []struct {
		name     string
		sampling string
		want     float64
	}{
		{
			name:     "always",
			sampling: "always",
			want:     1.0,
		},
		{
			name:     "never",
			sampling: "never",
			want:     0.0,
		},
		{
			name:     "empty string (default)",
			sampling: "",
			want:     1.0,
		},
		{
			name:     "valid float 0.5",
			sampling: "0.5",
			want:     0.5,
		},
		{
			name:     "valid float 0.1",
			sampling: "0.1",
			want:     0.1,
		},
		{
			name:     "valid float 1.0",
			sampling: "1.0",
			want:     1.0,
		},
		{
			name:     "invalid float",
			sampling: "invalid",
			want:     1.0, // defaults to always
		},
		{
			name:     "out of range high",
			sampling: "2.0",
			want:     1.0, // defaults to always
		},
		{
			name:     "out of range low",
			sampling: "-0.5",
			want:     1.0, // defaults to always
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSamplingRate(tt.sampling)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInitTracing(t *testing.T) {
	t.Run("disabled tracing", func(t *testing.T) {
		cfg := TracingConfig{
			Enabled: false,
		}

		shutdown, err := InitTracing(cfg)
		require.NoError(t, err)
		assert.NotNil(t, shutdown)

		// Shutdown should be no-op
		err = shutdown(context.Background())
		assert.NoError(t, err)
	})

	t.Run("enabled tracing with invalid endpoint", func(t *testing.T) {
		cfg := TracingConfig{
			Enabled:  true,
			Endpoint: "invalid-endpoint-that-does-not-exist:9999",
			Sampling: "always",
		}

		shutdown, err := InitTracing(cfg)
		// InitTracing should succeed even with invalid endpoint
		// Errors will occur when trying to export traces
		require.NoError(t, err)
		assert.NotNil(t, shutdown)

		// Clean up
		_ = shutdown(context.Background())
	})
}

func TestCalculateCost(t *testing.T) {
	tests := []struct {
		name       string
		provider   string
		tokensUsed int
		want       float64
	}{
		{
			name:       "anthropic",
			provider:   "anthropic",
			tokensUsed: 1_000_000,
			want:       3.0,
		},
		{
			name:       "openai",
			provider:   "openai",
			tokensUsed: 1_000_000,
			want:       10.0,
		},
		{
			name:       "google",
			provider:   "google",
			tokensUsed: 1_000_000,
			want:       1.25,
		},
		{
			name:       "gemini",
			provider:   "gemini",
			tokensUsed: 1_000_000,
			want:       1.25,
		},
		{
			name:       "ollama (free)",
			provider:   "ollama",
			tokensUsed: 1_000_000,
			want:       0.0,
		},
		{
			name:       "unknown provider",
			provider:   "unknown",
			tokensUsed: 1_000_000,
			want:       5.0, // default
		},
		{
			name:       "half million tokens",
			provider:   "anthropic",
			tokensUsed: 500_000,
			want:       1.5,
		},
		{
			name:       "small request",
			provider:   "google",
			tokensUsed: 1_000,
			want:       0.00125,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateCost(tt.provider, tt.tokensUsed)
			assert.InDelta(t, tt.want, got, 0.0001)
		})
	}
}

func TestTracedLLMClient_Generate(t *testing.T) {
	// Create a mock LLM client
	mockClient := NewMockLLMClient("test-provider", "test-model")

	// Wrap with tracing
	tracedClient := NewTracedLLMClient(mockClient)

	t.Run("successful generation", func(t *testing.T) {
		ctx := context.Background()
		resp, err := tracedClient.Generate(ctx, "test prompt", LLMConfig{})

		require.NoError(t, err)
		assert.Equal(t, "Mock LLM response", resp.Content)
		assert.Equal(t, 100, resp.TokensUsed)
		assert.Equal(t, "test-provider", resp.Provider)
	})

	t.Run("GetModelName", func(t *testing.T) {
		assert.Equal(t, "test-model", tracedClient.GetModelName())
	})

	t.Run("GetProviderName", func(t *testing.T) {
		assert.Equal(t, "test-provider", tracedClient.GetProviderName())
	})
}

func TestWrapLLMClient(t *testing.T) {
	mockClient := NewMockLLMClient("test", "test-model")

	t.Run("with tracing enabled", func(t *testing.T) {
		wrapped := WrapLLMClient(mockClient, true)
		assert.IsType(t, &TracedLLMClient{}, wrapped)
	})

	t.Run("with tracing disabled", func(t *testing.T) {
		wrapped := WrapLLMClient(mockClient, false)
		assert.Equal(t, mockClient, wrapped)
	})
}

func TestTraceAgentThink(t *testing.T) {
	t.Run("successful agent think", func(t *testing.T) {
		expectedResp := &AgentResponse{
			Role:       RoleArchitect,
			Content:    "Test content",
			TokensUsed: 500,
			Confidence: 0.85,
			Diagrams:   map[string]string{"component": "diagram"},
		}

		fn := func(ctx context.Context) (*AgentResponse, error) {
			return expectedResp, nil
		}

		ctx := context.Background()
		resp, err := TraceAgentThink(ctx, RoleArchitect, fn)

		require.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
	})

	t.Run("agent think with error", func(t *testing.T) {
		fn := func(ctx context.Context) (*AgentResponse, error) {
			return nil, assert.AnError
		}

		ctx := context.Background()
		resp, err := TraceAgentThink(ctx, RoleArchitect, fn)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

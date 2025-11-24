package hldgen

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGeminiClient(t *testing.T) {
	tests := []struct {
		name    string
		cfg     ProviderConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: ProviderConfig{
				Name:   "google",
				Model:  "gemini-1.5-pro",
				APIKey: "test-key",
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			cfg: ProviderConfig{
				Name:  "google",
				Model: "gemini-1.5-pro",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewGeminiClient(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, "google", client.GetProviderName())
			}
		})
	}
}

func TestGeminiClient_GetModelName(t *testing.T) {
	tests := []struct {
		name      string
		model     string
		wantModel string
	}{
		{
			name:      "explicit model",
			model:     "gemini-1.5-flash",
			wantModel: "gemini-1.5-flash",
		},
		{
			name:      "default model",
			model:     "",
			wantModel: "gemini-1.5-pro",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := NewGeminiClient(ProviderConfig{
				APIKey: "test",
				Model:  tt.model,
			})
			assert.Equal(t, tt.wantModel, client.GetModelName())
		})
	}
}

func TestGeminiClient_Integration(t *testing.T) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_API_KEY not set, skipping integration test")
	}

	tests := []struct {
		name          string
		model         string
		prompt        string
		temperature   float64
		maxTokens     int
		checkResponse func(t *testing.T, resp *LLMResponse)
	}{
		{
			name:        "simple generation with flash",
			model:       "gemini-1.5-flash",
			prompt:      "In one sentence, what is gRPC?",
			temperature: 0.0,
			maxTokens:   100,
			checkResponse: func(t *testing.T, resp *LLMResponse) {
				assert.NotEmpty(t, resp.Content)
				assert.Contains(t, strings.ToLower(resp.Content), "rpc")
				assert.Greater(t, resp.TokensUsed, 0)
				assert.Equal(t, "google", resp.Provider)
				assert.Equal(t, "gemini-1.5-flash", resp.Model)
				assert.Equal(t, 0.87, resp.Confidence)
			},
		},
		{
			name:        "technical prompt with pro",
			model:       "gemini-1.5-pro",
			prompt:      "Explain the architecture of a microservice with gRPC in 50 words",
			temperature: 0.2,
			maxTokens:   200,
			checkResponse: func(t *testing.T, resp *LLMResponse) {
				assert.NotEmpty(t, resp.Content)
				assert.Contains(t, strings.ToLower(resp.Content), "grpc")
				assert.Greater(t, resp.TokensUsed, 20)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ProviderConfig{
				Name:        "google",
				Model:       tt.model,
				APIKey:      apiKey,
				Temperature: tt.temperature,
				MaxTokens:   tt.maxTokens,
			}

			client, err := NewGeminiClient(cfg)
			require.NoError(t, err)

			ctx := context.Background()
			resp, err := client.Generate(ctx, tt.prompt, LLMConfig{})
			require.NoError(t, err)

			tt.checkResponse(t, resp)

			t.Logf("Response: %s", resp.Content)
			t.Logf("Tokens used: %d", resp.TokensUsed)
		})
	}
}

func TestGeminiClient_ErrorHandling(t *testing.T) {
	tests := []struct {
		name      string
		cfg       ProviderConfig
		prompt    string
		wantError string
	}{
		{
			name: "invalid API key",
			cfg: ProviderConfig{
				Name:   "google",
				Model:  "gemini-1.5-flash",
				APIKey: "invalid_key_12345",
			},
			prompt:    "test",
			wantError: "API error",
		},
		{
			name: "empty API key",
			cfg: ProviderConfig{
				Name:   "google",
				Model:  "gemini-1.5-flash",
				APIKey: "",
			},
			prompt:    "test",
			wantError: "API key not configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewGeminiClient(tt.cfg)
			if tt.cfg.APIKey == "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantError)
				return
			}

			require.NoError(t, err)

			ctx := context.Background()
			_, err = client.Generate(ctx, tt.prompt, LLMConfig{})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantError)
		})
	}
}

func TestGeminiClient_ContextCancellation(t *testing.T) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_API_KEY not set")
	}

	cfg := ProviderConfig{
		Name:   "google",
		Model:  "gemini-1.5-flash",
		APIKey: apiKey,
	}

	client, err := NewGeminiClient(cfg)
	require.NoError(t, err)

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for context to expire
	time.Sleep(10 * time.Millisecond)

	_, err = client.Generate(ctx, "Long prompt that won't complete in time", LLMConfig{})
	assert.Error(t, err)
	// Should contain context deadline exceeded or similar
}

func TestGeminiClient_LargeContext(t *testing.T) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_API_KEY not set")
	}

	// Test Gemini's large context window capability
	cfg := ProviderConfig{
		Name:        "google",
		Model:       "gemini-1.5-pro", // Pro has 2M context
		APIKey:      apiKey,
		Temperature: 0.0,
		MaxTokens:   500,
	}

	client, err := NewGeminiClient(cfg)
	require.NoError(t, err)

	// Create a prompt with multiple proto definitions
	largePrompt := `
	Analyze these gRPC services:

	service UserService {
		rpc GetUser(UserRequest) returns (User);
		rpc CreateUser(CreateUserRequest) returns (User);
		rpc UpdateUser(UpdateUserRequest) returns (User);
		rpc DeleteUser(DeleteUserRequest) returns (Empty);
		rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
	}

	service AuthService {
		rpc Login(LoginRequest) returns (LoginResponse);
		rpc Logout(LogoutRequest) returns (Empty);
		rpc RefreshToken(RefreshTokenRequest) returns (TokenResponse);
	}

	service PaymentService {
		rpc ProcessPayment(PaymentRequest) returns (PaymentResponse);
		rpc RefundPayment(RefundRequest) returns (RefundResponse);
	}

	Summarize the architecture in 2 sentences.
	`

	ctx := context.Background()
	resp, err := client.Generate(ctx, largePrompt, LLMConfig{})
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Content)
	assert.Greater(t, resp.TokensUsed, 100) // Should use significant tokens
	t.Logf("Large context response: %s", resp.Content)
	t.Logf("Tokens used: %d", resp.TokensUsed)
}

// BenchmarkGeminiClient_Generate benchmarks generation speed
func BenchmarkGeminiClient_Generate(b *testing.B) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		b.Skip("GOOGLE_API_KEY not set")
	}

	cfg := ProviderConfig{
		Name:        "google",
		Model:       "gemini-1.5-flash", // Fastest model
		APIKey:      apiKey,
		Temperature: 0.0,
		MaxTokens:   100,
	}

	client, err := NewGeminiClient(cfg)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	prompt := "What is gRPC?"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.Generate(ctx, prompt, LLMConfig{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

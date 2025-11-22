package adapters

import (
	"context"
	"fmt"

	"github.com/teilomillet/gollm"
)

// GollmLLMClient implements LLMClient using gollm for multi-provider support
type GollmLLMClient struct {
	llm      gollm.LLM
	model    string
	provider string
}

// NewGollmLLMClient creates a new gollm-based LLM client
func NewGollmLLMClient(provider, model, apiKey, baseURL string, temperature float64, maxTokens int) (*GollmLLMClient, error) {
	// Build gollm configuration
	opts := []gollm.ConfigOption{
		gollm.SetProvider(provider),
		gollm.SetModel(model),
		gollm.SetAPIKey(apiKey),
		gollm.SetMaxTokens(maxTokens),
		gollm.SetTemperature(temperature),
	}

	if baseURL != "" {
		opts = append(opts, gollm.SetExtraHeaders(map[string]string{
			"base_url": baseURL,
		}))
	}

	// Create gollm instance
	llm, err := gollm.NewLLM(opts...)
	if err != nil {
		return nil, fmt.Errorf("create gollm instance: %w", err)
	}

	return &GollmLLMClient{
		llm:      llm,
		model:    model,
		provider: provider,
	}, nil
}

// GenerateCompletion generates a completion for the given prompt
func (c *GollmLLMClient) GenerateCompletion(ctx context.Context, prompt string) (string, error) {
	response, err := c.llm.Generate(ctx, gollm.NewPrompt(prompt))
	if err != nil {
		return "", fmt.Errorf("gollm generate: %w", err)
	}
	return response, nil
}

// GenerateCompletionWithConfig generates a completion with custom configuration
func (c *GollmLLMClient) GenerateCompletionWithConfig(ctx context.Context, prompt string, config map[string]interface{}) (string, error) {
	// Build custom options from config
	opts := []gollm.ConfigOption{}

	if temp, ok := config["temperature"].(float64); ok {
		opts = append(opts, gollm.SetTemperature(temp))
	}
	if maxTokens, ok := config["max_tokens"].(int); ok {
		opts = append(opts, gollm.SetMaxTokens(maxTokens))
	}
	if topP, ok := config["top_p"].(float64); ok {
		opts = append(opts, gollm.SetTopP(topP))
	}

	// Create temporary LLM with custom config
	customLLM, err := gollm.NewLLM(append([]gollm.ConfigOption{
		gollm.SetProvider(c.provider),
		gollm.SetModel(c.model),
	}, opts...)...)
	if err != nil {
		return "", fmt.Errorf("create custom gollm instance: %w", err)
	}

	response, err := customLLM.Generate(ctx, gollm.NewPrompt(prompt))
	if err != nil {
		return "", fmt.Errorf("gollm generate with config: %w", err)
	}
	return response, nil
}

// GetModelName returns the model name
func (c *GollmLLMClient) GetModelName() string {
	return c.model
}

// GetProviderName returns the provider name
func (c *GollmLLMClient) GetProviderName() string {
	return c.provider
}

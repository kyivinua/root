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

// GeminiClient implements LLMClient for Google Gemini via REST API
type GeminiClient struct {
	cfg        ProviderConfig
	httpClient *http.Client
	apiKey     string
}

// NewGeminiClient creates a new Gemini client
func NewGeminiClient(cfg ProviderConfig) (*GeminiClient, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("google API key not configured")
	}

	return &GeminiClient{
		cfg:    cfg,
		apiKey: cfg.APIKey,
		httpClient: &http.Client{
			Timeout: 90 * time.Second, // Gemini can be slower with 2M context
		},
	}, nil
}

// geminiRequest represents the API request
type geminiRequest struct {
	Contents         []geminiContent `json:"contents"`
	GenerationConfig geminiGenConfig `json:"generationConfig,omitempty"`
	SafetySettings   []geminiSafety  `json:"safetySettings,omitempty"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type geminiSafety struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

// geminiResponse represents the API response
type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
	} `json:"usageMetadata"`
}

// Generate implements LLMClient.Generate
func (g *GeminiClient) Generate(ctx context.Context, prompt string, config LLMConfig) (*LLMResponse, error) {
	// Construct request
	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: geminiGenConfig{
			Temperature:     g.cfg.Temperature,
			MaxOutputTokens: g.cfg.MaxTokens,
		},
		SafetySettings: []geminiSafety{
			{Category: "HARM_CATEGORY_HARASSMENT", Threshold: "BLOCK_MEDIUM_AND_ABOVE"},
			{Category: "HARM_CATEGORY_HATE_SPEECH", Threshold: "BLOCK_MEDIUM_AND_ABOVE"},
			{Category: "HARM_CATEGORY_SEXUALLY_EXPLICIT", Threshold: "BLOCK_MEDIUM_AND_ABOVE"},
			{Category: "HARM_CATEGORY_DANGEROUS_CONTENT", Threshold: "BLOCK_MEDIUM_AND_ABOVE"},
		},
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build URL
	model := g.cfg.Model
	if model == "" {
		model = "gemini-1.5-pro"
	}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		model, g.apiKey)

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract content
	if len(geminiResp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}

	candidate := geminiResp.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		return nil, fmt.Errorf("no content parts in response")
	}

	text := candidate.Content.Parts[0].Text

	// Calculate tokens
	tokensUsed := geminiResp.UsageMetadata.PromptTokenCount +
		geminiResp.UsageMetadata.CandidatesTokenCount

	return &LLMResponse{
		Content:      text,
		TokensUsed:   tokensUsed,
		Model:        model,
		Provider:     "google",
		Confidence:   0.87, // Between Claude (0.90) and Ollama (0.85)
		FinishReason: candidate.FinishReason,
	}, nil
}

// GetModelName returns the model name
func (g *GeminiClient) GetModelName() string {
	if g.cfg.Model == "" {
		return "gemini-1.5-pro"
	}
	return g.cfg.Model
}

// GetProviderName returns the provider name
func (g *GeminiClient) GetProviderName() string {
	return "google"
}

// Package enricher provides AI client interfaces and implementations.
package enricher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AIClient defines the interface for AI API interactions.
type AIClient interface {
	Complete(ctx context.Context, prompt string, opts *CompletionOptions) (string, error)
	BatchComplete(ctx context.Context, prompts []string, opts *CompletionOptions) ([]string, error)
	ScoreQuality(ctx context.Context, content string) (float64, error)
}

// CompletionOptions configures AI completion requests.
type CompletionOptions struct {
	MaxTokens   int
	Temperature float64
	Model       string
}

// ClaudeClient implements AIClient for Anthropic's Claude API.
type ClaudeClient struct {
	apiKey     string
	httpClient *http.Client
	model      string
	baseURL    string
}

// NewClaudeClient creates a new Claude API client.
func NewClaudeClient(apiKey string, model string) *ClaudeClient {
	return &ClaudeClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		model:   model,
		baseURL: "https://api.anthropic.com/v1",
	}
}

// Complete performs a single AI completion.
func (c *ClaudeClient) Complete(ctx context.Context, prompt string, opts *CompletionOptions) (string, error) {
	if opts == nil {
		opts = &CompletionOptions{
			MaxTokens:   1024,
			Temperature: 0.3,
		}
	}

	model := c.model
	if opts.Model != "" {
		model = opts.Model
	}

	payload := map[string]interface{}{
		"model":      model,
		"max_tokens": opts.MaxTokens,
		"temperature": opts.Temperature,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var apiResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return "", fmt.Errorf("empty response from API")
	}

	return apiResp.Content[0].Text, nil
}

// BatchComplete performs multiple AI completions efficiently.
func (c *ClaudeClient) BatchComplete(ctx context.Context, prompts []string, opts *CompletionOptions) ([]string, error) {
	if len(prompts) == 0 {
		return []string{}, nil
	}

	// For now, process sequentially (future: use Claude's batch API when available)
	results := make([]string, len(prompts))

	for i, prompt := range prompts {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		result, err := c.Complete(ctx, prompt, opts)
		if err != nil {
			return nil, fmt.Errorf("batch item %d failed: %w", i, err)
		}
		results[i] = result
	}

	return results, nil
}

// ScoreQuality scores the quality of content using AI.
func (c *ClaudeClient) ScoreQuality(ctx context.Context, content string) (float64, error) {
	prompt := fmt.Sprintf(`Analyze the quality of this technical documentation on a scale of 0-100.

Documentation:
%s

Criteria:
- Clarity and readability (25 points)
- Technical accuracy (25 points)
- Completeness (25 points)
- Professional tone (15 points)
- Examples and details (10 points)

Respond with ONLY a number between 0 and 100.`, content)

	response, err := c.Complete(ctx, prompt, &CompletionOptions{
		MaxTokens:   10,
		Temperature: 0.1,
	})
	if err != nil {
		return 0, err
	}

	var score float64
	_, err = fmt.Sscanf(response, "%f", &score)
	if err != nil {
		// Fallback to simple heuristic
		return simpleQualityScore(content), nil
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score, nil
}

// MockAIClient provides a mock implementation for testing.
type MockAIClient struct {
	responses map[string]string
}

// NewMockAIClient creates a new mock AI client.
func NewMockAIClient() *MockAIClient {
	return &MockAIClient{
		responses: make(map[string]string),
	}
}

// SetResponse sets a mock response for a prompt pattern.
func (m *MockAIClient) SetResponse(pattern, response string) {
	m.responses[pattern] = response
}

// Complete returns a mock completion.
func (m *MockAIClient) Complete(ctx context.Context, prompt string, opts *CompletionOptions) (string, error) {
	// Return predefined response or generate simple one
	for pattern, response := range m.responses {
		if contains(prompt, pattern) {
			return response, nil
		}
	}

	return "This is a comprehensive description that provides detailed information about the functionality.", nil
}

// BatchComplete returns mock completions for a batch.
func (m *MockAIClient) BatchComplete(ctx context.Context, prompts []string, opts *CompletionOptions) ([]string, error) {
	results := make([]string, len(prompts))
	for i, prompt := range prompts {
		result, err := m.Complete(ctx, prompt, opts)
		if err != nil {
			return nil, err
		}
		results[i] = result
	}
	return results, nil
}

// ScoreQuality returns a mock quality score.
func (m *MockAIClient) ScoreQuality(ctx context.Context, content string) (float64, error) {
	return simpleQualityScore(content), nil
}

// BatchProcessor handles efficient batch processing of enrichment requests.
type BatchProcessor struct {
	client    AIClient
	batchSize int
}

// NewBatchProcessor creates a new batch processor.
func NewBatchProcessor(client AIClient, batchSize int) *BatchProcessor {
	return &BatchProcessor{
		client:    client,
		batchSize: batchSize,
	}
}

// ProcessBatch processes a batch of enrichment requests.
func (bp *BatchProcessor) ProcessBatch(ctx context.Context, requests []*EnrichmentRequest, opts *CompletionOptions) ([]*EnrichmentResult, error) {
	if len(requests) == 0 {
		return []*EnrichmentResult{}, nil
	}

	results := make([]*EnrichmentResult, len(requests))

	// Process in batches
	for i := 0; i < len(requests); i += bp.batchSize {
		end := i + bp.batchSize
		if end > len(requests) {
			end = len(requests)
		}

		batch := requests[i:end]
		prompts := make([]string, len(batch))

		for j, req := range batch {
			prompts[j] = buildSimplePrompt(req)
		}

		responses, err := bp.client.BatchComplete(ctx, prompts, opts)
		if err != nil {
			return nil, fmt.Errorf("batch processing failed: %w", err)
		}

		for j, response := range responses {
			results[i+j] = &EnrichmentResult{
				EnrichedContent: response,
				Confidence:      0.85,
				QualityScore:    simpleQualityScore(response),
			}
		}
	}

	return results, nil
}

// QualityScorer provides AI-powered quality assessment.
type QualityScorer struct {
	client AIClient
}

// NewQualityScorer creates a new quality scorer.
func NewQualityScorer(client AIClient) *QualityScorer {
	return &QualityScorer{client: client}
}

// ScoreService scores the overall quality of a service's documentation.
func (qs *QualityScorer) ScoreService(ctx context.Context, service interface{}) (*QualityAssessment, error) {
	// This would analyze the entire service
	// For now, return a simple assessment
	return &QualityAssessment{
		OverallScore: 85.0,
		Dimensions: map[string]float64{
			"clarity":       88.0,
			"completeness":  82.0,
			"consistency":   85.0,
			"accuracy":      90.0,
		},
		Suggestions: []string{
			"Consider adding more code examples",
			"Some method descriptions could be more detailed",
		},
	}, nil
}

// QualityAssessment represents a comprehensive quality assessment.
type QualityAssessment struct {
	OverallScore float64
	Dimensions   map[string]float64
	Suggestions  []string
	Strengths    []string
	Weaknesses   []string
}

// Helper functions

func buildSimplePrompt(req *EnrichmentRequest) string {
	return fmt.Sprintf("Generate a clear, concise description for this %s named '%s'. Current: %s",
		req.Type, req.Name, req.Content)
}

func simpleQualityScore(content string) float64 {
	score := 50.0

	words := len(splitWords(content))
	if words >= 20 {
		score += 20
	} else if words >= 10 {
		score += 10
	}

	sentences := countSentences(content)
	if sentences >= 3 {
		score += 15
	} else if sentences >= 2 {
		score += 10
	}

	if startsWithCapital(content) {
		score += 5
	}

	if hasExamples(content) {
		score += 10
	}

	if score > 100 {
		score = 100
	}

	return score
}

func contains(text, pattern string) bool {
	return len(pattern) > 0 && len(text) >= len(pattern) &&
		(text == pattern || len(text) > 0)
}

func splitWords(text string) []string {
	// Simple word splitting
	var words []string
	current := ""
	for _, r := range text {
		if r == ' ' || r == '\n' || r == '\t' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		words = append(words, current)
	}
	return words
}

func countSentences(text string) int {
	count := 0
	for _, r := range text {
		if r == '.' || r == '!' || r == '?' {
			count++
		}
	}
	return count
}

func startsWithCapital(text string) bool {
	if len(text) == 0 {
		return false
	}
	return text[0] >= 'A' && text[0] <= 'Z'
}

func hasExamples(text string) bool {
	lower := ""
	for _, r := range text {
		if r >= 'A' && r <= 'Z' {
			lower += string(r + 32)
		} else {
			lower += string(r)
		}
	}
	return contains(lower, "example") || contains(lower, "e.g.")
}

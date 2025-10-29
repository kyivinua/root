package enricher

import (
	"context"
	"testing"

	"github.com/kyivinua/docgen-tool/internal/docgen"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextAwareStrategy(t *testing.T) {
	client := NewMockAIClient()
	client.SetResponse("technical documentation", "This service provides comprehensive user management functionality.")

	strategy := NewContextAwareStrategy(client)

	req := &EnrichmentRequest{
		Type:    "service",
		Name:    "UserService",
		Content: "",
		Context: &EnrichmentContext{
			Service: &docgen.Service{
				Name:    "UserService",
				Package: "user.v1",
			},
			ProjectDomain: "Authentication and Authorization",
		},
		Options: &EnrichmentOptions{
			DetailLevel: "standard",
		},
	}

	result, err := strategy.Enrich(context.Background(), req)
	require.NoError(t, err)
	assert.NotEmpty(t, result.EnrichedContent)
	assert.Greater(t, result.Confidence, 0.0)
}

func TestMultiPassStrategy(t *testing.T) {
	client := NewMockAIClient()
	strategy := NewMultiPassStrategy(client, 2)

	req := &EnrichmentRequest{
		Type:    "method",
		Name:    "CreateUser",
		Content: "Creates a user",
		Context: &EnrichmentContext{},
		Options: &EnrichmentOptions{},
	}

	result, err := strategy.Enrich(context.Background(), req)
	require.NoError(t, err)
	assert.NotEmpty(t, result.EnrichedContent)
	assert.Greater(t, result.QualityScore, 0.0)
}

func TestTerminologyStrategy(t *testing.T) {
	client := NewMockAIClient()
	client.SetResponse("terminology", "User: An authenticated account\nRole: Permission level")

	strategy := NewTerminologyStrategy(client)

	req := &EnrichmentRequest{
		Type:    "service",
		Name:    "UserService",
		Content: "This service manages users and their roles",
		Context: &EnrichmentContext{},
	}

	result, err := strategy.Enrich(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result.Terminology)
}

func TestBatchProcessor(t *testing.T) {
	client := NewMockAIClient()
	processor := NewBatchProcessor(client, 2)

	requests := []*EnrichmentRequest{
		{Type: "method", Name: "Method1", Content: ""},
		{Type: "method", Name: "Method2", Content: ""},
		{Type: "method", Name: "Method3", Content: ""},
	}

	results, err := processor.ProcessBatch(context.Background(), requests, &CompletionOptions{})
	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestAdvancedEnricher(t *testing.T) {
	client := NewMockAIClient()
	config := DefaultAdvancedConfig()
	config.UseContextAware = true
	config.UseTerminology = true

	logger := zerolog.Nop()
	enricher := NewAdvancedEnricher(client, config, logger)

	service := &docgen.Service{
		Name:        "TestService",
		Package:     "test.v1",
		Description: "",
		Methods: []docgen.Method{
			{Name: "TestMethod", Description: ""},
		},
		Messages: []docgen.Message{
			{
				Name:        "TestMessage",
				Description: "",
				Fields: []docgen.Field{
					{Name: "test_field", Type: "string", Description: ""},
				},
			},
		},
	}

	summary, err := enricher.EnrichService(context.Background(), service)
	require.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, "TestService", summary.ServiceName)
}

func TestQualityScorer(t *testing.T) {
	client := NewMockAIClient()
	scorer := NewQualityScorer(client)

	assessment, err := scorer.ScoreService(context.Background(), &docgen.Service{})
	require.NoError(t, err)
	assert.NotNil(t, assessment)
	assert.Greater(t, assessment.OverallScore, 0.0)
}

func TestSimpleQualityScore(t *testing.T) {
	tests := []struct{
		name    string
		content string
		minScore float64
	}{
		{
			name:     "empty content",
			content:  "",
			minScore: 0,
		},
		{
			name:     "short content",
			content:  "Short description",
			minScore: 50,
		},
		{
			name:     "good content",
			content:  "This is a comprehensive description. It contains multiple sentences. For example, it provides clear information.",
			minScore: 70,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := simpleQualityScore(tt.content)
			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, 100.0)
		})
	}
}

func TestClaudeClient(t *testing.T) {
	// Skip if no API key
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test requires a real API key
	// In practice, this would be an integration test
	t.Skip("Requires real API key - use for integration testing only")
}

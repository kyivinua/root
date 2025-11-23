package enricher

import (
	"testing"
)

// TestEnricherConfig tests enricher configuration
func TestEnricherConfig(t *testing.T) {
	config := &EnrichmentConfig{
		Provider: "anthropic",
		APIKey:   "test-key",
		Model:    "claude-3-opus",
	}

	if config.Provider != "anthropic" {
		t.Errorf("Expected Provider to be anthropic, got %s", config.Provider)
	}
	if config.Model != "claude-3-opus" {
		t.Errorf("Expected Model to be claude-3-opus, got %s", config.Model)
	}
}

// TestEnricherConfigValidation tests config validation
func TestEnricherConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *EnrichmentConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &EnrichmentConfig{
				Provider: "anthropic",
				APIKey:   "test-key",
				Model:    "claude-3-opus",
			},
			wantErr: false,
		},
		{
			name: "missing provider",
			config: &EnrichmentConfig{
				Provider: "",
				APIKey:   "test-key",
				Model:    "claude-3-opus",
			},
			wantErr: true,
		},
		{
			name: "missing model",
			config: &EnrichmentConfig{
				Provider: "anthropic",
				APIKey:   "test-key",
				Model:    "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestRAGConfig tests RAG configuration
func TestRAGConfig(t *testing.T) {
	config := &RAGConfig{
		Enabled:     true,
		VectorStore: "weaviate",
		TopK:        5,
		MinScore:    0.7,
	}

	if !config.Enabled {
		t.Error("RAG should be enabled")
	}
	if config.VectorStore != "weaviate" {
		t.Errorf("Expected VectorStore to be weaviate, got %s", config.VectorStore)
	}
	if config.TopK != 5 {
		t.Errorf("Expected TopK to be 5, got %d", config.TopK)
	}
	if config.MinScore != 0.7 {
		t.Errorf("Expected MinScore to be 0.7, got %f", config.MinScore)
	}
}

// TestSafetyConfig tests safety configuration
func TestSafetyConfig(t *testing.T) {
	config := &SafetyConfig{
		Enabled:            true,
		PIIChecks:          true,
		UseSemanticEntropy: true,
		EntropyThreshold:   0.8,
		UseLLMJudge:        true,
		PCIDSSChecks:       true,
	}

	if !config.Enabled {
		t.Error("Safety should be enabled")
	}
	if !config.PIIChecks {
		t.Error("PII checks should be enabled")
	}
	if !config.UseSemanticEntropy {
		t.Error("Semantic entropy should be enabled")
	}
	if config.EntropyThreshold != 0.8 {
		t.Errorf("Expected EntropyThreshold to be 0.8, got %f", config.EntropyThreshold)
	}
}

// TestEnrichmentManifest tests enrichment manifest
func TestEnrichmentManifest(t *testing.T) {
	manifest := NewEnrichmentManifest()

	if manifest.Version != "1.0" {
		t.Errorf("Expected Version to be 1.0, got %s", manifest.Version)
	}
	if manifest.Targets == nil {
		t.Error("Targets map should be initialized")
	}

	// Test recording enrichment
	manifest.RecordEnrichment("test-target", true, false, 100)

	if manifest.Statistics.TotalTargets != 1 {
		t.Errorf("Expected TotalTargets to be 1, got %d", manifest.Statistics.TotalTargets)
	}
	if manifest.Statistics.EnrichedTargets != 1 {
		t.Errorf("Expected EnrichedTargets to be 1, got %d", manifest.Statistics.EnrichedTargets)
	}
	if manifest.Statistics.TotalTokensUsed != 100 {
		t.Errorf("Expected TotalTokensUsed to be 100, got %d", manifest.Statistics.TotalTokensUsed)
	}
}

// BenchmarkConfigValidation benchmarks config validation
func BenchmarkConfigValidation(b *testing.B) {
	config := &EnrichmentConfig{
		Provider: "anthropic",
		APIKey:   "test-key",
		Model:    "claude-3-opus",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}

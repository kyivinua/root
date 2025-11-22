package hldgen

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestValidateConfig tests configuration validation
func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "valid config",
			config: &Config{
				Enabled: true,
				Version: "7.0",
				Refinement: RefinementConfig{
					MaxRounds:              3,
					ConsensusThreshold:     0.88,
					MinImprovementPerRound: 0.05,
				},
				Intelligence: IntelligenceConfig{
					ConsensusThreshold: 0.88,
					MultiAgent: MultiAgentConfig{
						Enabled: true,
						Agents: []AgentConfig{
							{Role: "architect", Weight: 0.30, Model: "claude"},
							{Role: "security", Weight: 0.25, Model: "claude"},
							{Role: "sre", Weight: 0.20, Model: "claude"},
							{Role: "pm", Weight: 0.15, Model: "claude"},
							{Role: "qa", Weight: 0.10, Model: "claude"},
						},
					},
				},
				LLM: LLMConfig{
					Providers: []ProviderConfig{
						{Name: "anthropic", Model: "claude-3-5-sonnet-20241022"},
					},
				},
				Performance: PerformanceConfig{
					Concurrency:    10,
					TimeoutSeconds: 300,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid max rounds",
			config: &Config{
				Refinement: RefinementConfig{
					MaxRounds: 100, // Invalid - too high
				},
			},
			wantErr: true,
		},
		{
			name: "invalid consensus threshold",
			config: &Config{
				Refinement: RefinementConfig{
					MaxRounds:          3,
					ConsensusThreshold: 1.5, // Invalid - > 1.0
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateConsolidatedDocs tests docs validation
func TestValidateConsolidatedDocs(t *testing.T) {
	tests := []struct {
		name    string
		docs    *ConsolidatedDocs
		wantErr bool
	}{
		{
			name:    "nil docs",
			docs:    nil,
			wantErr: true,
		},
		{
			name: "valid docs",
			docs: &ConsolidatedDocs{
				ModuleName: "test.module",
				Services: []ServiceDoc{
					{
						Name: "TestService",
						Methods: []MethodDoc{
							{
								Name:       "TestMethod",
								InputType:  "Request",
								OutputType: "Response",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty module name",
			docs: &ConsolidatedDocs{
				ModuleName: "",
				Services: []ServiceDoc{
					{Name: "Service"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConsolidatedDocs(tt.docs)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConsolidatedDocs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestMockLLMClient tests the mock LLM client
func TestMockLLMClient(t *testing.T) {
	client := NewMockLLMClient("test-provider", "test-model")

	if client.GetProviderName() != "test-provider" {
		t.Errorf("Expected provider 'test-provider', got '%s'", client.GetProviderName())
	}

	if client.GetModelName() != "test-model" {
		t.Errorf("Expected model 'test-model', got '%s'", client.GetModelName())
	}

	ctx := context.Background()
	resp, err := client.Generate(ctx, "test prompt", LLMConfig{})
	if err != nil {
		t.Errorf("Generate() returned unexpected error: %v", err)
	}

	if resp.TokensUsed != 100 {
		t.Errorf("Expected 100 tokens, got %d", resp.TokensUsed)
	}

	if resp.Confidence != 0.85 {
		t.Errorf("Expected confidence 0.85, got %.2f", resp.Confidence)
	}
}

// TestResponseMerger tests response merging logic
func TestResponseMerger(t *testing.T) {
	merger := NewResponseMerger("priority_merge")

	responses := map[string]*AgentResponse{
		"architect": {
			Role:       RoleArchitect,
			Content:    "Architecture content",
			Confidence: 0.85,
			GeneratedAt: time.Now(),
		},
		"security": {
			Role:       RoleSecurity,
			Content:    "Security content",
			Confidence: 0.90,
			GeneratedAt: time.Now(),
		},
	}

	output := merger.Merge(responses)

	if output == nil {
		t.Fatal("Merge() returned nil")
	}

	if output.Architecture == "" {
		t.Error("Architecture section is empty")
	}

	if output.Security == "" {
		t.Error("Security section is empty")
	}

	if output.Markdown == "" {
		t.Error("Markdown output is empty")
	}
}

// TestCriticAgentScoring tests critic agent scoring
func TestCriticAgentScoring(t *testing.T) {
	cfg := AgentConfig{
		Role:   "critic",
		Model:  "test-model",
		Weight: 1.0,
	}

	critic := NewCriticAgent(cfg, 0.88)

	criticisms := []Criticism{
		{Agent: "architect", Score: 0.90, Confidence: 0.85},
		{Agent: "security", Score: 0.92, Confidence: 0.90},
		{Agent: "sre", Score: 0.88, Confidence: 0.85},
		{Agent: "pm", Score: 0.85, Confidence: 0.80},
		{Agent: "qa", Score: 0.87, Confidence: 0.85},
	}

	consensus := critic.EvaluateConsensus(criticisms)

	if consensus < 0.85 || consensus > 0.95 {
		t.Errorf("Expected consensus between 0.85-0.95, got %.2f", consensus)
	}
}

// TestSanitizeInput tests input sanitization
func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "null bytes",
			input:    "test\x00string",
			expected: "teststring",
		},
		{
			name:     "carriage return",
			input:    "line1\r\nline2",
			expected: "line1\nline2",
		},
		{
			name:     "long input",
			input:    strings.Repeat("a", 15000),
			expected: strings.Repeat("a", 10000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeInput(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
			}
		})
	}
}

// TestValidateMode tests mode validation
func TestValidateMode(t *testing.T) {
	validModes := []Mode{
		ModeBasic,
		ModeAdvanced,
		ModeBusiness,
		ModeCompliance,
		ModeUltraAdvanced,
		ModeCustom,
	}

	for _, mode := range validModes {
		if err := ValidateMode(mode); err != nil {
			t.Errorf("ValidateMode(%s) returned error: %v", mode, err)
		}
	}

	if err := ValidateMode("invalid_mode"); err == nil {
		t.Error("ValidateMode() should have returned error for invalid mode")
	}
}

// BenchmarkValidateConfig benchmarks config validation
func BenchmarkValidateConfig(b *testing.B) {
	cfg := DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateConfig(cfg)
	}
}

// BenchmarkMerge benchmarks response merging
func BenchmarkMerge(b *testing.B) {
	merger := NewResponseMerger("priority_merge")

	responses := map[string]*AgentResponse{
		"architect": {Role: RoleArchitect, Content: "Test", GeneratedAt: time.Now()},
		"security":  {Role: RoleSecurity, Content: "Test", GeneratedAt: time.Now()},
		"sre":       {Role: RoleSRE, Content: "Test", GeneratedAt: time.Now()},
		"pm":        {Role: RolePM, Content: "Test", GeneratedAt: time.Now()},
		"qa":        {Role: RoleQA, Content: "Test", GeneratedAt: time.Now()},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = merger.Merge(responses)
	}
}

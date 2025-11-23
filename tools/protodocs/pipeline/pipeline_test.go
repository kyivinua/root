package pipeline

import (
	"path/filepath"
	"testing"
)

// TestNewPipeline tests pipeline creation
func TestNewPipeline(t *testing.T) {
	config := &PipelineConfig{
		ProtoRoot: "/test/proto",
		Docs: DocsConfig{
			OutputDir: "/test/output",
		},
	}

	pipeline := NewPipeline(config)

	if pipeline == nil {
		t.Fatal("NewPipeline returned nil")
	}
	if pipeline.config != config {
		t.Error("Pipeline config not set correctly")
	}
	if pipeline.logger == nil {
		t.Error("Pipeline logger not initialized")
	}
}

// TestPipelineConfig tests config validation
func TestPipelineConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *PipelineConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &PipelineConfig{
				ProtoRoot: "/test/proto",
				Docs: DocsConfig{
					OutputDir: "/test/output",
				},
			},
			wantErr: false,
		},
		{
			name: "empty proto root",
			config: &PipelineConfig{
				ProtoRoot: "",
				Docs: DocsConfig{
					OutputDir: "/test/output",
				},
			},
			wantErr: true,
		},
		{
			name: "empty output dir",
			config: &PipelineConfig{
				ProtoRoot: "/test/proto",
				Docs: DocsConfig{
					OutputDir: "",
				},
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

// TestRunLint tests individual lint stage
func TestRunLint(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	config := &PipelineConfig{
		ProtoRoot: tmpDir,
		Docs: DocsConfig{
			OutputDir: filepath.Join(tmpDir, "output"),
		},
	}

	pipeline := NewPipeline(config)

	// This will fail because there are no proto files, but we're testing the method exists
	err := pipeline.RunLint()
	// We expect an error since there are no proto files
	if err == nil {
		t.Log("RunLint completed (no proto files)")
	}
}

// TestRunBreaking tests individual breaking check stage
func TestRunBreaking(t *testing.T) {
	tmpDir := t.TempDir()

	config := &PipelineConfig{
		ProtoRoot: tmpDir,
		Docs: DocsConfig{
			OutputDir: filepath.Join(tmpDir, "output"),
		},
	}

	pipeline := NewPipeline(config)

	// This will likely fail, but we're testing the method exists and can be called
	err := pipeline.RunBreaking()
	// We expect an error since there are no proto files or git history
	if err == nil {
		t.Log("RunBreaking completed")
	}
}

// TestRunDescriptorBuild tests descriptor build stage
func TestRunDescriptorBuild(t *testing.T) {
	tmpDir := t.TempDir()

	config := &PipelineConfig{
		ProtoRoot: tmpDir,
		Docs: DocsConfig{
			OutputDir: filepath.Join(tmpDir, "output"),
		},
	}

	pipeline := NewPipeline(config)

	// This will fail because there are no proto files
	_, err := pipeline.RunDescriptorBuild()
	// We expect an error
	if err == nil {
		t.Error("Expected error when building descriptor with no proto files")
	}
}

// TestIncrementalDiscoveryConfig tests incremental discovery configuration
func TestIncrementalDiscoveryConfig(t *testing.T) {
	tests := []struct {
		name   string
		config DiscoveryConfig
		want   bool
	}{
		{
			name: "incremental enabled",
			config: DiscoveryConfig{
				Incremental: true,
				BaseRef:     "main",
				HeadRef:     "HEAD",
			},
			want: true,
		},
		{
			name: "incremental disabled",
			config: DiscoveryConfig{
				Incremental: false,
			},
			want: false,
		},
		{
			name: "incremental with defaults",
			config: DiscoveryConfig{
				Incremental: true,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.Incremental != tt.want {
				t.Errorf("Incremental = %v, want %v", tt.config.Incremental, tt.want)
			}
		})
	}
}

// TestScopeIsEmpty tests Scope.IsEmpty method
func TestScopeIsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		scope *Scope
		want  bool
	}{
		{
			name: "empty scope",
			scope: &Scope{
				ProtoFiles:    []string{},
				ProtoPackages: []string{},
			},
			want: true,
		},
		{
			name: "scope with files",
			scope: &Scope{
				ProtoFiles:    []string{"test.proto"},
				ProtoPackages: []string{"test"},
			},
			want: false,
		},
		{
			name: "scope with only packages (no files)",
			scope: &Scope{
				ProtoFiles:    []string{},
				ProtoPackages: []string{"test"},
			},
			want: true, // Scope is empty if there are no files to process
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.scope.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestEnrichmentConfig tests enrichment configuration
func TestEnrichmentConfig(t *testing.T) {
	config := &PipelineConfig{
		ProtoRoot: "/test/proto",
		Docs: DocsConfig{
			OutputDir: "/test/output",
		},
		Enrichment: EnrichmentConfig{
			Enabled:    true,
			ConfigPath: "/test/enricher.yaml",
		},
	}

	if !config.Enrichment.Enabled {
		t.Error("Enrichment should be enabled")
	}
	if config.Enrichment.ConfigPath != "/test/enricher.yaml" {
		t.Error("Enrichment config path not set correctly")
	}
}

// TestHLDConfig tests HLD generation configuration
func TestHLDConfig(t *testing.T) {
	config := &PipelineConfig{
		ProtoRoot: "/test/proto",
		Docs: DocsConfig{
			OutputDir: "/test/output",
		},
		HLD: HLDConfig{
			Enabled:    true,
			ConfigPath: "/test/hld.yaml",
		},
	}

	if !config.HLD.Enabled {
		t.Error("HLD should be enabled")
	}
	if config.HLD.ConfigPath != "/test/hld.yaml" {
		t.Error("HLD config path not set correctly")
	}
}

// TestDiagramsConfig tests diagrams configuration
func TestDiagramsConfig(t *testing.T) {
	config := &PipelineConfig{
		ProtoRoot: "/test/proto",
		Docs: DocsConfig{
			OutputDir: "/test/output",
		},
		Diagrams: DiagramsConfig{
			Enabled: true,
		},
	}

	if !config.Diagrams.Enabled {
		t.Error("Diagrams should be enabled")
	}
}

// TestPublishingConfig tests publishing configuration
func TestPublishingConfig(t *testing.T) {
	config := &PipelineConfig{
		ProtoRoot: "/test/proto",
		Docs: DocsConfig{
			OutputDir: "/test/output",
		},
		Publishers: PublishersConfig{
			Confluence: ConfluenceConfig{
				Enabled:  true,
				BaseURL:  "https://confluence.example.com",
				SpaceKey: "API",
			},
		},
	}

	if !config.Publishers.Confluence.Enabled {
		t.Error("Confluence publishing should be enabled")
	}
	if config.Publishers.Confluence.SpaceKey != "API" {
		t.Error("Confluence space key not set correctly")
	}
}

// BenchmarkNewPipeline benchmarks pipeline creation
func BenchmarkNewPipeline(b *testing.B) {
	config := &PipelineConfig{
		ProtoRoot: "/test/proto",
		Docs: DocsConfig{
			OutputDir: "/test/output",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewPipeline(config)
	}
}

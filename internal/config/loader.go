// Package config provides configuration management for docgen-tool.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure.
type Config struct {
	Version     string            `yaml:"version"`
	ProjectName string            `yaml:"project_name"`
	OutputDir   string            `yaml:"output_dir"`
	Services    []ServiceConfig   `yaml:"services"`
	Quality     QualityConfig     `yaml:"quality"`
	Enricher    EnricherConfig    `yaml:"enricher"`
	Diagrams    DiagramConfig     `yaml:"diagrams"`
	Templates   TemplateConfig    `yaml:"templates"`
	Logging     LoggingConfig     `yaml:"logging"`
}

// ServiceConfig represents service-specific configuration.
type ServiceConfig struct {
	Name       string   `yaml:"name"`
	ProtoFiles []string `yaml:"proto_files"`
	Enabled    bool     `yaml:"enabled"`
	Package    string   `yaml:"package"`
	Version    string   `yaml:"version"`
}

// QualityConfig represents quality validation settings.
type QualityConfig struct {
	CoverageTarget            float64 `yaml:"coverage_target"`
	DescriptionQualityTarget  float64 `yaml:"description_quality_target"`
	AutoFix                   bool    `yaml:"auto_fix"`
	FailOnQualityGate         bool    `yaml:"fail_on_quality_gate"`
	MinMethodCoverage         float64 `yaml:"min_method_coverage"`
	MinFieldCoverage          float64 `yaml:"min_field_coverage"`
}

// EnricherConfig represents AI enrichment settings.
type EnricherConfig struct {
	Provider         string   `yaml:"provider"`
	APIKey           string   `yaml:"api_key"`
	Enabled          bool     `yaml:"enabled"`
	Model            string   `yaml:"model"`
	MaxTokens        int      `yaml:"max_tokens"`
	Temperature      float64  `yaml:"temperature"`
	CacheEnabled     bool     `yaml:"cache_enabled"`
	RateLimitPerMin  int      `yaml:"rate_limit_per_min"`
	ParallelRequests int      `yaml:"parallel_requests"`

	// Advanced features
	Advanced         AdvancedEnricherConfig `yaml:"advanced"`
}

// AdvancedEnricherConfig represents advanced AI enrichment settings.
type AdvancedEnricherConfig struct {
	Enabled              bool     `yaml:"enabled"`
	UseContextAware      bool     `yaml:"use_context_aware"`
	UseExampleGeneration bool     `yaml:"use_example_generation"`
	UseTerminology       bool     `yaml:"use_terminology"`
	UseBestPractices     bool     `yaml:"use_best_practices"`
	UseMultiPass         bool     `yaml:"use_multi_pass"`
	MultiPassCount       int      `yaml:"multi_pass_count"`
	BatchSize            int      `yaml:"batch_size"`
	DetailLevel          string   `yaml:"detail_level"`
	Tone                 string   `yaml:"tone"`
	IncludeCodeExamples  bool     `yaml:"include_code_examples"`
	LanguagesForExamples []string `yaml:"languages_for_examples"`
	ProjectDomain        string   `yaml:"project_domain"`
	MinQualityScore      float64  `yaml:"min_quality_score"`
	EnableQualityScoring bool     `yaml:"enable_quality_scoring"`
}

// DiagramConfig represents diagram generation settings.
type DiagramConfig struct {
	Enabled      bool `yaml:"enabled"`
	ServiceGraph bool `yaml:"service_graph"`
	MessageGraph bool `yaml:"message_graph"`
	SequenceGraph bool `yaml:"sequence_graph"`
	OverviewGraph bool `yaml:"overview_graph"`
}

// TemplateConfig represents template settings.
type TemplateConfig struct {
	Dir          string            `yaml:"dir"`
	ServiceTmpl  string            `yaml:"service_template"`
	MethodTmpl   string            `yaml:"method_template"`
	MessageTmpl  string            `yaml:"message_template"`
	CustomFuncs  map[string]string `yaml:"custom_functions"`
}

// LoggingConfig represents logging settings.
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"` // json, console
	OutputFile string `yaml:"output_file"`
}

// Load loads configuration from a YAML file.
func Load(path string) (*Config, error) {
	// Expand environment variables in path
	path = os.ExpandEnv(path)

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Expand environment variables in config
	if err := expandEnvVars(&config); err != nil {
		return nil, fmt.Errorf("failed to expand environment variables: %w", err)
	}

	// Validate configuration
	if err := Validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Apply defaults
	applyDefaults(&config)

	return &config, nil
}

// Validate validates the configuration.
func Validate(config *Config) error {
	if config.ProjectName == "" {
		return fmt.Errorf("project_name is required")
	}

	if config.OutputDir == "" {
		return fmt.Errorf("output_dir is required")
	}

	if len(config.Services) == 0 {
		return fmt.Errorf("at least one service must be configured")
	}

	for i, svc := range config.Services {
		if svc.Name == "" {
			return fmt.Errorf("service[%d]: name is required", i)
		}
		if len(svc.ProtoFiles) == 0 {
			return fmt.Errorf("service[%d]: at least one proto_file is required", i)
		}
		for _, pf := range svc.ProtoFiles {
			if pf == "" {
				return fmt.Errorf("service[%d]: proto_file path cannot be empty", i)
			}
		}
	}

	// Validate quality thresholds
	if config.Quality.CoverageTarget < 0 || config.Quality.CoverageTarget > 100 {
		return fmt.Errorf("quality.coverage_target must be between 0 and 100")
	}
	if config.Quality.DescriptionQualityTarget < 0 || config.Quality.DescriptionQualityTarget > 100 {
		return fmt.Errorf("quality.description_quality_target must be between 0 and 100")
	}

	// Validate enricher config if enabled
	if config.Enricher.Enabled {
		if config.Enricher.Provider == "" {
			return fmt.Errorf("enricher.provider is required when enricher is enabled")
		}
		if config.Enricher.APIKey == "" {
			return fmt.Errorf("enricher.api_key is required when enricher is enabled")
		}
	}

	return nil
}

// applyDefaults applies default values to the configuration.
func applyDefaults(config *Config) {
	if config.Version == "" {
		config.Version = "1.0"
	}

	// Quality defaults
	if config.Quality.CoverageTarget == 0 {
		config.Quality.CoverageTarget = 80
	}
	if config.Quality.DescriptionQualityTarget == 0 {
		config.Quality.DescriptionQualityTarget = 85
	}
	if config.Quality.MinMethodCoverage == 0 {
		config.Quality.MinMethodCoverage = 75
	}
	if config.Quality.MinFieldCoverage == 0 {
		config.Quality.MinFieldCoverage = 70
	}

	// Enricher defaults
	if config.Enricher.Enabled {
		if config.Enricher.Model == "" {
			config.Enricher.Model = "claude-3-5-sonnet-20241022"
		}
		if config.Enricher.MaxTokens == 0 {
			config.Enricher.MaxTokens = 1024
		}
		if config.Enricher.Temperature == 0 {
			config.Enricher.Temperature = 0.3
		}
		if config.Enricher.RateLimitPerMin == 0 {
			config.Enricher.RateLimitPerMin = 50
		}
		if config.Enricher.ParallelRequests == 0 {
			config.Enricher.ParallelRequests = 5
		}

		// Advanced enricher defaults
		if config.Enricher.Advanced.Enabled {
			if config.Enricher.Advanced.MultiPassCount == 0 {
				config.Enricher.Advanced.MultiPassCount = 2
			}
			if config.Enricher.Advanced.BatchSize == 0 {
				config.Enricher.Advanced.BatchSize = 5
			}
			if config.Enricher.Advanced.DetailLevel == "" {
				config.Enricher.Advanced.DetailLevel = "standard"
			}
			if config.Enricher.Advanced.Tone == "" {
				config.Enricher.Advanced.Tone = "technical"
			}
			if len(config.Enricher.Advanced.LanguagesForExamples) == 0 {
				config.Enricher.Advanced.LanguagesForExamples = []string{"go", "python"}
			}
			if config.Enricher.Advanced.MinQualityScore == 0 {
				config.Enricher.Advanced.MinQualityScore = 70.0
			}
		}
	}

	// Logging defaults
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "json"
	}

	// Service defaults
	for i := range config.Services {
		if !config.Services[i].Enabled {
			config.Services[i].Enabled = true
		}
	}
}

// expandEnvVars expands environment variables in configuration fields.
func expandEnvVars(config *Config) error {
	config.OutputDir = os.ExpandEnv(config.OutputDir)
	config.Enricher.APIKey = os.ExpandEnv(config.Enricher.APIKey)
	config.Templates.Dir = os.ExpandEnv(config.Templates.Dir)
	config.Logging.OutputFile = os.ExpandEnv(config.Logging.OutputFile)

	for i := range config.Services {
		for j := range config.Services[i].ProtoFiles {
			config.Services[i].ProtoFiles[j] = os.ExpandEnv(config.Services[i].ProtoFiles[j])
		}
	}

	return nil
}

// InitConfig creates a default configuration file.
func InitConfig(outputPath string) error {
	defaultConfig := &Config{
		Version:     "1.0",
		ProjectName: "my-project",
		OutputDir:   "./docs/generated",
		Services: []ServiceConfig{
			{
				Name: "ExampleService",
				ProtoFiles: []string{
					"./proto/example/v1/service.proto",
				},
				Enabled: true,
				Package: "example.v1",
				Version: "1.0",
			},
		},
		Quality: QualityConfig{
			CoverageTarget:           80,
			DescriptionQualityTarget: 85,
			AutoFix:                  true,
			FailOnQualityGate:        false,
			MinMethodCoverage:        75,
			MinFieldCoverage:         70,
		},
		Enricher: EnricherConfig{
			Provider:         "claude",
			APIKey:           "${ANTHROPIC_API_KEY}",
			Enabled:          false,
			Model:            "claude-3-5-sonnet-20241022",
			MaxTokens:        1024,
			Temperature:      0.3,
			CacheEnabled:     true,
			RateLimitPerMin:  50,
			ParallelRequests: 5,
		},
		Diagrams: DiagramConfig{
			Enabled:       true,
			ServiceGraph:  true,
			MessageGraph:  true,
			SequenceGraph: false,
			OverviewGraph: true,
		},
		Templates: TemplateConfig{
			Dir:         "./templates",
			ServiceTmpl: "service.md.tmpl",
			MethodTmpl:  "method.md.tmpl",
			MessageTmpl: "message.md.tmpl",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetServiceByName retrieves a service configuration by name.
func (c *Config) GetServiceByName(name string) *ServiceConfig {
	for i := range c.Services {
		if strings.EqualFold(c.Services[i].Name, name) {
			return &c.Services[i]
		}
	}
	return nil
}

// GetEnabledServices returns all enabled services.
func (c *Config) GetEnabledServices() []ServiceConfig {
	var enabled []ServiceConfig
	for _, svc := range c.Services {
		if svc.Enabled {
			enabled = append(enabled, svc)
		}
	}
	return enabled
}

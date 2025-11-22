package enricher

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// EnrichmentConfig holds all enrichment configuration
type EnrichmentConfig struct {
	// LLM Provider Configuration
	Provider    string  `yaml:"provider"`     // "anthropic", "openai", "ollama"
	Model       string  `yaml:"model"`        // "claude-3-5-sonnet-20241022", etc.
	Temperature float64 `yaml:"temperature"`  // 0.0 for determinism
	MaxTokens   int     `yaml:"max_tokens"`   // Response limit
	TopP        float64 `yaml:"top_p"`        // Nucleus sampling
	APIKey      string  `yaml:"api_key"`      // Or use env var
	BaseURL     string  `yaml:"base_url"`     // For custom endpoints
	Timeout     int     `yaml:"timeout_secs"` // Request timeout

	// RAG Configuration
	RAG RAGConfig `yaml:"rag"`

	// Safety Configuration
	Safety SafetyConfig `yaml:"safety"`

	// Cache Configuration
	Cache CacheConfig `yaml:"cache"`

	// Metrics Configuration
	Metrics MetricsConfig `yaml:"metrics"`

	// Policy Configuration
	Policy PolicyConfig `yaml:"policy"`

	// Concurrency Configuration
	Concurrency ConcurrencyConfig `yaml:"concurrency"`

	// Template Configuration
	Templates TemplateConfig `yaml:"templates"`

	// Trace Configuration
	Trace TraceConfig `yaml:"trace"`
}

// RAGConfig holds RAG-specific configuration
type RAGConfig struct {
	Enabled      bool          `yaml:"enabled"`
	VectorStore  string        `yaml:"vector_store"`  // "weaviate", "qdrant"
	EmbedModel   string        `yaml:"embed_model"`   // Embedding model name
	TopK         int           `yaml:"top_k"`         // Default retrieval count
	MinScore     float64       `yaml:"min_score"`     // Minimum relevance score
	UseAdaptive  bool          `yaml:"use_adaptive"`  // Adaptive RAG based on confidence
	Weaviate     WeaviateConfig `yaml:"weaviate"`
}

// WeaviateConfig holds Weaviate-specific settings
type WeaviateConfig struct {
	Host       string `yaml:"host"`
	Scheme     string `yaml:"scheme"`
	ClassName  string `yaml:"class_name"`
	APIKey     string `yaml:"api_key"`
}

// SafetyConfig holds safety validation settings
type SafetyConfig struct {
	Enabled             bool    `yaml:"enabled"`
	UseSemanticEntropy  bool    `yaml:"use_semantic_entropy"`
	UseLLMJudge         bool    `yaml:"use_llm_judge"`
	JudgeModel          string  `yaml:"judge_model"`
	EntropyThreshold    float64 `yaml:"entropy_threshold"`
	ConfidenceThreshold float64 `yaml:"confidence_threshold"`
	PIIChecks           bool    `yaml:"pii_checks"`
	PCIDSSChecks        bool    `yaml:"pci_dss_checks"`
}

// CacheConfig holds caching settings
type CacheConfig struct {
	Enabled     bool          `yaml:"enabled"`
	Type        string        `yaml:"type"`         // "ristretto", "redis"
	MaxSize     int64         `yaml:"max_size"`     // Max cache size in bytes
	TTL         time.Duration `yaml:"ttl"`          // Cache TTL
	NumCounters int64         `yaml:"num_counters"` // Ristretto counters
}

// MetricsConfig holds metrics settings
type MetricsConfig struct {
	Enabled        bool   `yaml:"enabled"`
	Type           string `yaml:"type"`           // "prometheus", "nop"
	PrometheusPort int    `yaml:"prometheus_port"`
	Namespace      string `yaml:"namespace"`
	Subsystem      string `yaml:"subsystem"`
}

// PolicyConfig holds policy engine settings
type PolicyConfig struct {
	Enabled     bool   `yaml:"enabled"`
	Type        string `yaml:"type"`        // "static", "remote"
	ConfigPath  string `yaml:"config_path"` // Path to tenant policies
	DefaultTenant string `yaml:"default_tenant"`
}

// ConcurrencyConfig holds concurrency settings
type ConcurrencyConfig struct {
	MaxConcurrent int `yaml:"max_concurrent"` // Max parallel enrichments
	RateLimitQPS  int `yaml:"rate_limit_qps"` // Queries per second limit
}

// TemplateConfig holds prompt template settings
type TemplateConfig struct {
	TemplateDir   string `yaml:"template_dir"`   // Directory with templates
	DefaultFormat string `yaml:"default_format"` // "xml", "cot", "plain"
	UseCoT        bool   `yaml:"use_cot"`        // Use Chain-of-Thought
}

// TraceConfig holds audit trail settings
type TraceConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Type       string `yaml:"type"`       // "file", "database", "nop"
	OutputPath string `yaml:"output_path"` // For file-based traces
}

// LoadConfig loads enrichment configuration from YAML file
func LoadConfig(path string) (*EnrichmentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg EnrichmentConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config YAML: %w", err)
	}

	// Apply environment variable overrides
	if apiKey := os.Getenv("LLM_API_KEY"); apiKey != "" {
		cfg.APIKey = apiKey
	}
	if weaviateKey := os.Getenv("WEAVIATE_API_KEY"); weaviateKey != "" {
		cfg.RAG.Weaviate.APIKey = weaviateKey
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// Validate checks configuration validity
func (c *EnrichmentConfig) Validate() error {
	if c.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if c.Model == "" {
		return fmt.Errorf("model is required")
	}
	if c.Temperature < 0 || c.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}
	if c.MaxTokens <= 0 {
		c.MaxTokens = 4096 // Default
	}
	if c.Timeout <= 0 {
		c.Timeout = 60 // Default 60 seconds
	}

	// Validate RAG config if enabled
	if c.RAG.Enabled {
		if c.RAG.VectorStore == "" {
			return fmt.Errorf("rag.vector_store is required when RAG is enabled")
		}
		if c.RAG.TopK <= 0 {
			c.RAG.TopK = 5 // Default
		}
		if c.RAG.VectorStore == "weaviate" {
			if c.RAG.Weaviate.Host == "" {
				return fmt.Errorf("rag.weaviate.host is required for Weaviate")
			}
			if c.RAG.Weaviate.ClassName == "" {
				c.RAG.Weaviate.ClassName = "ProtoDocumentation" // Default
			}
		}
	}

	// Validate safety config
	if c.Safety.Enabled {
		if c.Safety.UseLLMJudge && c.Safety.JudgeModel == "" {
			return fmt.Errorf("safety.judge_model is required when LLM judge is enabled")
		}
		if c.Safety.EntropyThreshold <= 0 {
			c.Safety.EntropyThreshold = 0.5 // Default
		}
		if c.Safety.ConfidenceThreshold <= 0 {
			c.Safety.ConfidenceThreshold = 0.7 // Default
		}
	}

	// Validate cache config
	if c.Cache.Enabled {
		if c.Cache.Type == "" {
			c.Cache.Type = "ristretto" // Default
		}
		if c.Cache.MaxSize <= 0 {
			c.Cache.MaxSize = 100 * 1024 * 1024 // 100MB default
		}
		if c.Cache.TTL <= 0 {
			c.Cache.TTL = 24 * time.Hour // Default 24 hours
		}
	}

	// Validate concurrency config
	if c.Concurrency.MaxConcurrent <= 0 {
		c.Concurrency.MaxConcurrent = 10 // Default
	}
	if c.Concurrency.RateLimitQPS <= 0 {
		c.Concurrency.RateLimitQPS = 10 // Default
	}

	return nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *EnrichmentConfig {
	return &EnrichmentConfig{
		Provider:    "anthropic",
		Model:       "claude-3-5-sonnet-20241022",
		Temperature: 0.0,
		MaxTokens:   4096,
		TopP:        1.0,
		Timeout:     60,
		RAG: RAGConfig{
			Enabled:     true,
			VectorStore: "weaviate",
			EmbedModel:  "text-embedding-3-small",
			TopK:        5,
			MinScore:    0.7,
			UseAdaptive: true,
			Weaviate: WeaviateConfig{
				Host:      "localhost:8080",
				Scheme:    "http",
				ClassName: "ProtoDocumentation",
			},
		},
		Safety: SafetyConfig{
			Enabled:             true,
			UseSemanticEntropy:  true,
			UseLLMJudge:         true,
			JudgeModel:          "claude-3-haiku-20240307",
			EntropyThreshold:    0.5,
			ConfidenceThreshold: 0.7,
			PIIChecks:           true,
			PCIDSSChecks:        true,
		},
		Cache: CacheConfig{
			Enabled:     true,
			Type:        "ristretto",
			MaxSize:     100 * 1024 * 1024,
			TTL:         24 * time.Hour,
			NumCounters: 1000000,
		},
		Metrics: MetricsConfig{
			Enabled:        true,
			Type:           "prometheus",
			PrometheusPort: 9090,
			Namespace:      "protodocs",
			Subsystem:      "enricher",
		},
		Policy: PolicyConfig{
			Enabled:       false,
			Type:          "static",
			DefaultTenant: "default",
		},
		Concurrency: ConcurrencyConfig{
			MaxConcurrent: 10,
			RateLimitQPS:  10,
		},
		Templates: TemplateConfig{
			DefaultFormat: "xml",
			UseCoT:        true,
		},
		Trace: TraceConfig{
			Enabled: true,
			Type:    "file",
		},
	}
}

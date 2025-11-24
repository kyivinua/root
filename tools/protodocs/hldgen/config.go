package hldgen

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the complete HLD Generator configuration
type Config struct {
	Enabled       bool               `yaml:"enabled"`
	Version       string             `yaml:"version"`
	DefaultMode   Mode               `yaml:"default_mode"`
	Modes         map[Mode]ModeConfig `yaml:"modes"`
	Context       ContextConfig      `yaml:"context"`
	Style         StyleConfig        `yaml:"style"`
	DSL           DSLConfig          `yaml:"dsl"`
	Intelligence  IntelligenceConfig `yaml:"intelligence"`
	Refinement    RefinementConfig   `yaml:"refinement"`
	LLM           LLMConfig          `yaml:"llm"`
	Output        OutputConfig       `yaml:"output"`
	Performance   PerformanceConfig  `yaml:"performance"`
	Cache         CacheConfig        `yaml:"cache"`
	Observability ObservabilityConfig `yaml:"observability"`
	Security      SecurityConfig     `yaml:"security"`
}

// ModeConfig represents configuration for a specific generation mode
type ModeConfig struct {
	Enabled    bool     `yaml:"enabled"`
	LLM        bool     `yaml:"llm"`
	Diagrams   bool     `yaml:"diagrams"`
	RAG        bool     `yaml:"rag"`
	Canvas     string   `yaml:"canvas,omitempty"`
	Standards  []string `yaml:"standards,omitempty"`
	Reasoning  string   `yaml:"reasoning,omitempty"`
	Output     string   `yaml:"output,omitempty"`
	PromptPath string   `yaml:"prompt_path,omitempty"`
}

// ContextConfig represents Context Engine configuration
type ContextConfig struct {
	Enabled bool            `yaml:"enabled"`
	Sources []ContextSource `yaml:"sources"`
	RAG     RAGConfig       `yaml:"rag"`
}

// RAGConfig represents RAG configuration
type RAGConfig struct {
	Enabled       bool   `yaml:"enabled"`
	VectorDB      string `yaml:"vector_db"`
	WeaviateURL   string `yaml:"weaviate_url,omitempty"`
	ClassName     string `yaml:"class_name"`
	TopK          int    `yaml:"top_k"`
	HybridSearch  bool   `yaml:"hybrid_search"`
}

// StyleConfig represents styling configuration
type StyleConfig struct {
	Preset string       `yaml:"preset"`
	Custom CustomStyle  `yaml:"custom"`
}

// CustomStyle represents custom style configuration
type CustomStyle struct {
	Enabled     bool              `yaml:"enabled"`
	TemplateDir string            `yaml:"template_dir"`
	Fonts       FontConfig        `yaml:"fonts"`
	Colors      ColorConfig       `yaml:"colors"`
	Tone        string            `yaml:"tone"`
	Language    string            `yaml:"language"`
}

type FontConfig struct {
	Heading string `yaml:"heading"`
	Body    string `yaml:"body"`
}

type ColorConfig struct {
	Primary    string `yaml:"primary"`
	Accent     string `yaml:"accent"`
	Background string `yaml:"background"`
}

// DSLConfig represents DSL configuration
type DSLConfig struct {
	Enabled            bool                   `yaml:"enabled"`
	SyntaxHighlighting bool                   `yaml:"syntax_highlighting"`
	Extensions         map[string]bool        `yaml:"extensions"`
}

// IntelligenceConfig represents AI intelligence configuration
type IntelligenceConfig struct {
	Reasoning              ReasoningConfig `yaml:"reasoning"`
	SelfCritique           bool            `yaml:"self_critique"`
	MultiAgent             MultiAgentConfig `yaml:"multi_agent"`
	ConsensusThreshold     float64         `yaml:"consensus_threshold"`
	AutoRefinement         bool            `yaml:"auto_refinement"`
	HallucinationMitigation string         `yaml:"hallucination_mitigation"`
	QualityGates           QualityGates    `yaml:"quality_gates"`
}

type ReasoningConfig struct {
	Type  string `yaml:"type"`
	Depth int    `yaml:"depth"`
}

type MultiAgentConfig struct {
	Enabled bool          `yaml:"enabled"`
	Agents  []AgentConfig `yaml:"agents"`
}

type AgentConfig struct {
	Role        string  `yaml:"role"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	Weight      float64 `yaml:"weight"`
	MaxTokens   int     `yaml:"max_tokens,omitempty"`
}

type QualityGates struct {
	ConsistencyScore      float64 `yaml:"consistency_score"`
	BusinessValueCoverage float64 `yaml:"business_value_coverage"`
	SecurityCoverage      float64 `yaml:"security_coverage"`
}

// RefinementConfig represents refinement loop configuration
type RefinementConfig struct {
	MaxRounds              int     `yaml:"max_rounds"`
	ConsensusThreshold     float64 `yaml:"consensus_threshold"`
	MinImprovementPerRound float64 `yaml:"min_improvement_per_round"`
	CriticalIssueOverride  bool    `yaml:"critical_issue_override"`
	FallbackOnStagnation   bool    `yaml:"fallback_on_stagnation"`
	EnableSelfReflection   bool    `yaml:"enable_self_reflection"`
	FinalValidation        bool    `yaml:"final_validation"`
}

// LLMConfig represents LLM configuration
type LLMConfig struct {
	Router        RouterConfig    `yaml:"router"`
	Providers     []ProviderConfig `yaml:"providers"`
	FallbackChain []string        `yaml:"fallback_chain"`
}

type RouterConfig struct {
	Strategy string `yaml:"strategy"` // cost_then_quality | quality_first | fastest
}

type ProviderConfig struct {
	Name        string  `yaml:"name"`
	Model       string  `yaml:"model"`
	Weight      float64 `yaml:"weight"`
	Temperature float64 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens"`
	APIKey      string  `yaml:"api_key"`
	BaseURL     string  `yaml:"base_url,omitempty"` // Optional base URL (for Ollama, custom endpoints)
}

// OutputConfig represents output configuration
type OutputConfig struct {
	Formats    []string         `yaml:"formats"`
	Confluence ConfluenceConfig `yaml:"confluence"`
	Git        GitConfig        `yaml:"git"`
}

type ConfluenceConfig struct {
	Enabled        bool   `yaml:"enabled"`
	SpaceKey       string `yaml:"space_key"`
	ParentPage     string `yaml:"parent_page"`
	AutoPublish    bool   `yaml:"auto_publish"`
	Versioning     string `yaml:"versioning"`
	ConfluenceURL  string `yaml:"confluence_url"`
	ConfluenceToken string `yaml:"confluence_token"`
}

type GitConfig struct {
	CommitEnabled bool   `yaml:"commit_enabled"`
	Path          string `yaml:"path"`
	Branch        string `yaml:"branch"`
	RepoURL       string `yaml:"repo_url"`
}

// PerformanceConfig represents performance configuration
type PerformanceConfig struct {
	Concurrency    int         `yaml:"concurrency"`
	BatchSize      int         `yaml:"batch_size"`
	TimeoutSeconds int         `yaml:"timeout_seconds"`
	RateLimit      RateLimitConfig `yaml:"rate_limit"`
}

type RateLimitConfig struct {
	RPS   int `yaml:"rps"`
	Burst int `yaml:"burst"`
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	Primary   string            `yaml:"primary"`
	Secondary string            `yaml:"secondary,omitempty"`
	RedisURL  string            `yaml:"redis_url,omitempty"`
	TTL       map[string]string `yaml:"ttl"`
}

// ObservabilityConfig represents observability configuration
type ObservabilityConfig struct {
	Tracing TracingConfig `yaml:"tracing"`
	Metrics MetricsConfig `yaml:"metrics"`
	Logging LoggingConfig `yaml:"logging"`
}

type TracingConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Exporter string `yaml:"exporter"`
	Endpoint string `yaml:"endpoint"`
	Sampling string `yaml:"sampling"`
}

type MetricsConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Namespace string `yaml:"namespace"`
}

type LoggingConfig struct {
	Level           string `yaml:"level"`
	IncludePrompt   bool   `yaml:"include_prompt"`
	IncludeResponse bool   `yaml:"include_response"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	PIIDetection                      string   `yaml:"pii_detection"`
	SecretDetection                   bool     `yaml:"secret_detection"`
	AllowedDomains                    []string `yaml:"allowed_domains"`
	ContentPolicy                     ContentPolicyConfig `yaml:"content_policy"`
}

type ContentPolicyConfig struct {
	BlockPatterns                       []string `yaml:"block_patterns"`
	RequireApprovalForCustomPrompts     bool     `yaml:"require_approval_for_custom_prompts"`
}

// LoadConfig loads HLD Generator configuration from file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Apply environment variable overrides
	cfg.applyEnvOverrides()

	// Set defaults
	cfg.setDefaults()

	return &cfg, nil
}

// applyEnvOverrides applies environment variable overrides
func (c *Config) applyEnvOverrides() {
	// LLM providers
	for i := range c.LLM.Providers {
		provider := &c.LLM.Providers[i]
		switch provider.Name {
		case "anthropic":
			if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
				provider.APIKey = key
			}
		case "openai":
			if key := os.Getenv("OPENAI_API_KEY"); key != "" {
				provider.APIKey = key
			}
		case "groq":
			if key := os.Getenv("GROQ_API_KEY"); key != "" {
				provider.APIKey = key
			}
		}
	}

	// Context sources
	for i := range c.Context.Sources {
		source := &c.Context.Sources[i]
		switch source.Type {
		case "jira_tickets":
			if token := os.Getenv("JIRA_TOKEN"); token != "" {
				if source.Config == nil {
					source.Config = make(map[string]interface{})
				}
				source.Config["jira_token"] = token
			}
		case "slo_dashboard":
			if token := os.Getenv("GRAFANA_TOKEN"); token != "" {
				if source.Config == nil {
					source.Config = make(map[string]interface{})
				}
				source.Config["grafana_token"] = token
			}
		case "security_policies":
			if token := os.Getenv("VAULT_TOKEN"); token != "" {
				if source.Config == nil {
					source.Config = make(map[string]interface{})
				}
				source.Config["vault_token"] = token
			}
		}
	}

	// Confluence
	if token := os.Getenv("CONFLUENCE_TOKEN"); token != "" {
		c.Output.Confluence.ConfluenceToken = token
	}
}

// setDefaults sets default values for missing configuration
func (c *Config) setDefaults() {
	if c.Version == "" {
		c.Version = "7.0"
	}
	if c.DefaultMode == "" {
		c.DefaultMode = ModeUltraAdvanced
	}
	if c.Intelligence.ConsensusThreshold == 0 {
		c.Intelligence.ConsensusThreshold = 0.88
	}
	if c.Refinement.MaxRounds == 0 {
		c.Refinement.MaxRounds = 3
	}
	if c.Refinement.ConsensusThreshold == 0 {
		c.Refinement.ConsensusThreshold = 0.88
	}
	if c.Refinement.MinImprovementPerRound == 0 {
		c.Refinement.MinImprovementPerRound = 0.05
	}
	if c.Context.RAG.TopK == 0 {
		c.Context.RAG.TopK = 15
	}
	if c.Performance.Concurrency == 0 {
		c.Performance.Concurrency = 25
	}
	if c.Performance.TimeoutSeconds == 0 {
		c.Performance.TimeoutSeconds = 600
	}
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:     true,
		Version:     "7.0",
		DefaultMode: ModeUltraAdvanced,
		Refinement: RefinementConfig{
			MaxRounds:              3,
			ConsensusThreshold:     0.88,
			MinImprovementPerRound: 0.05,
			CriticalIssueOverride:  true,
			FallbackOnStagnation:   true,
			EnableSelfReflection:   false,
			FinalValidation:        true,
		},
		Intelligence: IntelligenceConfig{
			ConsensusThreshold:      0.88,
			AutoRefinement:          true,
			HallucinationMitigation: "rag+judge+entropy",
			QualityGates: QualityGates{
				ConsistencyScore:      0.88,
				BusinessValueCoverage: 0.90,
				SecurityCoverage:      0.95,
			},
		},
		Performance: PerformanceConfig{
			Concurrency:    25,
			BatchSize:      8,
			TimeoutSeconds: 600,
		},
	}
}

package pipeline

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// PipelineConfig holds the configuration for the documentation pipeline.
type PipelineConfig struct {
	ProtoRoot string `yaml:"proto_root"`
	UseBuf    bool   `yaml:"use_buf"`

	Lint         LintConfig         `yaml:"lint"`
	Breaking     BreakingConfig     `yaml:"breaking"`
	Enrichment   EnrichmentConfig   `yaml:"enrichment"`
	Notifications NotificationsConfig `yaml:"notifications"`

	Descriptors DescriptorsConfig `yaml:"descriptors"`
	Docs        DocsConfig        `yaml:"docs"`
	OpenAPI     OpenAPIConfig     `yaml:"openapi"`
	Site        SiteConfig        `yaml:"site"`
}

// LintConfig holds lint configuration.
type LintConfig struct {
	EnableBufLint      bool `yaml:"enable_buf_lint"`
	EnableCommentsCheck bool `yaml:"enable_comments_check"`
}

// BreakingConfig holds breaking check configuration.
type BreakingConfig struct {
	Enable bool   `yaml:"enable"`
	Target string `yaml:"target"` // e.g., ".git#branch=main"
}

// DescriptorsConfig holds descriptor build configuration.
type DescriptorsConfig struct {
	OutputPath string `yaml:"output_path"`
}

// DocsConfig holds docs generation configuration.
type DocsConfig struct {
	OutputFormat     string   `yaml:"output_format"`      // "markdown" or "html"
	OutputDir        string   `yaml:"output_dir"`
	VisibilityFilter []string `yaml:"visibility_filter"`  // ["PUBLIC", "PARTNER"]
}

// OpenAPIConfig holds OpenAPI generation configuration.
type OpenAPIConfig struct {
	Enabled          bool     `yaml:"enabled"`
	Plugin           string   `yaml:"plugin"`             // "openapiv3"
	OutputDir        string   `yaml:"output_dir"`
	VisibilityFilter []string `yaml:"visibility_filter"`  // ["PUBLIC"]
}

// SiteConfig holds site assembly configuration.
type SiteConfig struct {
	Generator  string `yaml:"generator"`   // "mkdocs" or "docusaurus"
	ConfigPath string `yaml:"config_path"` // path to mkdocs.yml
	OutputDir  string `yaml:"output_dir"`
}

// EnrichmentConfig holds enrichment configuration.
type EnrichmentConfig struct {
	Enabled        bool   `yaml:"enabled"`
	ConfigPath     string `yaml:"config_path"`      // path to enricher.config.yaml
	ManifestPath   string `yaml:"manifest_path"`    // path to output manifest
	Tenant         string `yaml:"tenant"`           // tenant ID for policy
	OutputModelPath string `yaml:"output_model_path"` // path to enriched model
}

// NotificationsConfig holds notifications configuration.
type NotificationsConfig struct {
	Enabled    bool        `yaml:"enabled"`
	Slack      SlackConfig `yaml:"slack"`
}

// SlackConfig holds Slack-specific configuration.
type SlackConfig struct {
	Enabled                bool   `yaml:"enabled"`
	WebhookURL             string `yaml:"webhook_url"`              // Can be set via SLACK_WEBHOOK_URL
	BotToken               string `yaml:"bot_token"`                // Can be set via SLACK_BOT_TOKEN
	Channel                string `yaml:"channel"`
	Username               string `yaml:"username"`
	IconEmoji              string `yaml:"icon_emoji"`
	NotifyOnStart          bool   `yaml:"notify_on_start"`
	NotifyOnComplete       bool   `yaml:"notify_on_complete"`
	NotifyOnFailure        bool   `yaml:"notify_on_failure"`
	NotifyOnBreaking       bool   `yaml:"notify_on_breaking"`
	NotifyOnEnrichment     bool   `yaml:"notify_on_enrichment"`
	NotifyReleaseNotes     bool   `yaml:"notify_release_notes"`
	ReleaseNotesVersion    string `yaml:"release_notes_version"`
	ReleaseNotesFromRef    string `yaml:"release_notes_from_ref"`
}

// LoadConfig loads pipeline configuration from a YAML file.
func LoadConfig(path string) (*PipelineConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg PipelineConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Set defaults
	if cfg.ProtoRoot == "" {
		cfg.ProtoRoot = "./proto"
	}
	if cfg.Descriptors.OutputPath == "" {
		cfg.Descriptors.OutputPath = "api-docs/descriptors/image.bin"
	}
	if cfg.Docs.OutputFormat == "" {
		cfg.Docs.OutputFormat = "markdown"
	}
	if cfg.Docs.OutputDir == "" {
		cfg.Docs.OutputDir = "./api-docs/proto-docs"
	}
	if cfg.OpenAPI.OutputDir == "" {
		cfg.OpenAPI.OutputDir = "./api-docs/openapi"
	}
	if cfg.Site.Generator == "" {
		cfg.Site.Generator = "mkdocs"
	}
	if cfg.Site.OutputDir == "" {
		cfg.Site.OutputDir = "./api-docs/site"
	}

	return &cfg, nil
}

// DefaultConfig returns a default pipeline configuration.
func DefaultConfig() *PipelineConfig {
	return &PipelineConfig{
		ProtoRoot: "./proto",
		UseBuf:    true,
		Lint: LintConfig{
			EnableBufLint:      true,
			EnableCommentsCheck: true,
		},
		Breaking: BreakingConfig{
			Enable: true,
			Target: ".git#branch=main",
		},
		Enrichment: EnrichmentConfig{
			Enabled:         false,
			ConfigPath:      "configs/enricher.config.yaml",
			ManifestPath:    "api-docs/enrichment-manifest.json",
			Tenant:          "default",
			OutputModelPath: "api-docs/model/api-doc-model-enriched.json",
		},
		Notifications: NotificationsConfig{
			Enabled: false,
			Slack: SlackConfig{
				Enabled:            false,
				Username:           "ProtoDocs Bot",
				IconEmoji:          ":book:",
				NotifyOnStart:      true,
				NotifyOnComplete:   true,
				NotifyOnFailure:    true,
				NotifyOnBreaking:   true,
				NotifyOnEnrichment: true,
				NotifyReleaseNotes: false,
			},
		},
		Descriptors: DescriptorsConfig{
			OutputPath: "api-docs/descriptors/image.bin",
		},
		Docs: DocsConfig{
			OutputFormat:     "markdown",
			OutputDir:        "./api-docs/proto-docs",
			VisibilityFilter: []string{"PUBLIC", "PARTNER"},
		},
		OpenAPI: OpenAPIConfig{
			Enabled:          true,
			Plugin:           "openapiv3",
			OutputDir:        "./api-docs/openapi",
			VisibilityFilter: []string{"PUBLIC"},
		},
		Site: SiteConfig{
			Generator:  "mkdocs",
			ConfigPath:  "./docs-site/mkdocs.yml",
			OutputDir:  "./api-docs/site",
		},
	}
}

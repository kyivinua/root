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

	Lint          LintConfig          `yaml:"lint"`
	Breaking      BreakingConfig      `yaml:"breaking"`
	Enrichment    EnrichmentConfig    `yaml:"enrichment"`
	HLD           HLDConfig           `yaml:"hld"`
	Notifications NotificationsConfig `yaml:"notifications"`

	Descriptors DescriptorsConfig `yaml:"descriptors"`
	Docs        DocsConfig        `yaml:"docs"`
	OpenAPI     OpenAPIConfig     `yaml:"openapi"`
	Diagrams    DiagramsConfig    `yaml:"diagrams"`
	Site        SiteConfig        `yaml:"site"`
	Publishers  PublishersConfig  `yaml:"publishers"`
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

// PublishersConfig holds configuration for documentation publishers.
type PublishersConfig struct {
	Confluence ConfluenceConfig `yaml:"confluence"`
}

// ConfluenceConfig holds Confluence publishing configuration.
type ConfluenceConfig struct {
	Enabled      bool   `yaml:"enabled"`
	BaseURL      string `yaml:"base_url"`       // e.g., "https://company.atlassian.net/wiki"
	Username     string `yaml:"username"`       // Can be set via CONFLUENCE_USERNAME
	APIToken     string `yaml:"api_token"`      // Can be set via CONFLUENCE_API_TOKEN
	SpaceKey     string `yaml:"space_key"`      // Confluence space key
	ParentPageID string `yaml:"parent_page_id"` // Parent page ID for documentation

	// Page structure options
	CreatePagePerService bool   `yaml:"create_page_per_service"` // true = one page per service, false = single consolidated page
	PageTitlePrefix      string `yaml:"page_title_prefix"`       // Prefix for page titles
	IncludeTOC           bool   `yaml:"include_toc"`             // Include table of contents
	IncludeDiagrams      bool   `yaml:"include_diagrams"`        // Include Mermaid diagrams
	IncludeCodeExamples  bool   `yaml:"include_code_examples"`   // Include code examples

	// Update behavior
	UpdateExisting bool   `yaml:"update_existing"`  // Update existing pages or create new versions
	VersionLabel   string `yaml:"version_label"`    // Version label for pages

	// Content filtering
	VisibilityFilter []string `yaml:"visibility_filter"` // ["PUBLIC", "PARTNER", "INTERNAL"]
}

// EnrichmentConfig holds enrichment configuration.
type EnrichmentConfig struct {
	Enabled        bool   `yaml:"enabled"`
	ConfigPath     string `yaml:"config_path"`      // path to enricher.config.yaml
	ManifestPath   string `yaml:"manifest_path"`    // path to output manifest
	Tenant         string `yaml:"tenant"`           // tenant ID for policy
	OutputModelPath string `yaml:"output_model_path"` // path to enriched model
}

// HLDConfig holds High-Level Design generation configuration.
type HLDConfig struct {
	Enabled         bool   `yaml:"enabled"`
	ConfigPath      string `yaml:"config_path"`        // path to hld_generator.yaml
	InputModelPath  string `yaml:"input_model_path"`   // path to enriched model (from enricher)
	OutputDir       string `yaml:"output_dir"`         // output directory for HLD
	ModuleName      string `yaml:"module_name"`        // optional module name override
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

// DiagramsConfig holds diagram generation configuration.
type DiagramsConfig struct {
	Enabled                bool   `yaml:"enabled"`
	OutputDir              string `yaml:"output_dir"`
	EnablePipeline         bool   `yaml:"enable_pipeline"`
	EnableEnricher         bool   `yaml:"enable_enricher"`
	EnableComponent        bool   `yaml:"enable_component"`
	EnableTransform        bool   `yaml:"enable_transform"`
	EnableDeploy           bool   `yaml:"enable_deploy"`
	EnableDataModel        bool   `yaml:"enable_data_model"`
	EnableServiceMap       bool   `yaml:"enable_service_map"`
	EnableMessageHierarchy bool   `yaml:"enable_message_hierarchy"`
	GenerateIndex          bool   `yaml:"generate_index"`
	Theme                  string `yaml:"theme"` // default, forest, dark, neutral
	MaxServicesPerDiagram  int    `yaml:"max_services_per_diagram"`
	MaxMessagesPerDiagram  int    `yaml:"max_messages_per_diagram"`
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
	if cfg.Diagrams.OutputDir == "" {
		cfg.Diagrams.OutputDir = "./api-docs/diagrams"
	}
	if cfg.Diagrams.Theme == "" {
		cfg.Diagrams.Theme = "default"
	}
	if cfg.Diagrams.MaxServicesPerDiagram == 0 {
		cfg.Diagrams.MaxServicesPerDiagram = 20
	}
	if cfg.Diagrams.MaxMessagesPerDiagram == 0 {
		cfg.Diagrams.MaxMessagesPerDiagram = 30
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
		Diagrams: DiagramsConfig{
			Enabled:                true,
			OutputDir:              "./api-docs/diagrams",
			EnablePipeline:         true,
			EnableEnricher:         true,
			EnableComponent:        true,
			EnableTransform:        true,
			EnableDeploy:           true,
			EnableDataModel:        true,
			EnableServiceMap:       true,
			EnableMessageHierarchy: false,
			GenerateIndex:          true,
			Theme:                  "default",
			MaxServicesPerDiagram:  20,
			MaxMessagesPerDiagram:  30,
		},
		Site: SiteConfig{
			Generator:  "mkdocs",
			ConfigPath: "./docs-site/mkdocs.yml",
			OutputDir:  "./api-docs/site",
		},
		Publishers: PublishersConfig{
			Confluence: ConfluenceConfig{
				Enabled:              false,
				CreatePagePerService: true,
				PageTitlePrefix:      "API Documentation -",
				IncludeTOC:           true,
				IncludeDiagrams:      true,
				IncludeCodeExamples:  true,
				UpdateExisting:       true,
				VisibilityFilter:     []string{"PUBLIC", "PARTNER"},
			},
		},
	}
}

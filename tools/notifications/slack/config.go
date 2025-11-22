package slack

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds Slack notification configuration
type Config struct {
	Enabled bool `yaml:"enabled"`

	// Webhook configuration
	WebhookURL string `yaml:"webhook_url"` // Can be set via SLACK_WEBHOOK_URL env var

	// Bot API configuration (alternative to webhook)
	BotToken string `yaml:"bot_token"` // Can be set via SLACK_BOT_TOKEN env var
	Channel  string `yaml:"channel"`

	// Message customization
	Username  string `yaml:"username"`
	IconEmoji string `yaml:"icon_emoji"`

	// Notification preferences
	NotifyOnStart          bool `yaml:"notify_on_start"`
	NotifyOnComplete       bool `yaml:"notify_on_complete"`
	NotifyOnFailure        bool `yaml:"notify_on_failure"`
	NotifyOnBreaking       bool `yaml:"notify_on_breaking"`
	NotifyOnEnrichment     bool `yaml:"notify_on_enrichment"`
	NotifyReleaseNotes     bool `yaml:"notify_release_notes"`

	// Release notes configuration
	ReleaseNotesVersion string `yaml:"release_notes_version"` // e.g., "v1.2.3"
	ReleaseNotesFromRef string `yaml:"release_notes_from_ref"` // e.g., "v1.2.2"
}

// LoadConfig loads Slack configuration from YAML file
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
	if webhookURL := os.Getenv("SLACK_WEBHOOK_URL"); webhookURL != "" {
		cfg.WebhookURL = webhookURL
	}
	if botToken := os.Getenv("SLACK_BOT_TOKEN"); botToken != "" {
		cfg.BotToken = botToken
	}
	if channel := os.Getenv("SLACK_CHANNEL"); channel != "" {
		cfg.Channel = channel
	}

	return &cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if !c.Enabled {
		return nil // Not enabled, skip validation
	}

	if c.WebhookURL == "" && c.BotToken == "" {
		return fmt.Errorf("either webhook_url or bot_token must be configured")
	}

	if c.BotToken != "" && c.Channel == "" {
		return fmt.Errorf("channel is required when using bot_token")
	}

	return nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:                false,
		Username:               "ProtoDocs Bot",
		IconEmoji:              ":book:",
		NotifyOnStart:          true,
		NotifyOnComplete:       true,
		NotifyOnFailure:        true,
		NotifyOnBreaking:       true,
		NotifyOnEnrichment:     true,
		NotifyReleaseNotes:     false,
		ReleaseNotesVersion:    "",
		ReleaseNotesFromRef:    "",
	}
}

package pipeline

import (
	"os"
	"time"

	"github.com/kyivinua/root/tools/notifications/slack"
)

// NotificationManager manages pipeline notifications
type NotificationManager struct {
	config        *NotificationsConfig
	slackNotifier *slack.Notifier
	startTime     time.Time
	commit        string
	branch        string
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(config *NotificationsConfig) *NotificationManager {
	if !config.Enabled || !config.Slack.Enabled {
		return nil
	}

	// Apply environment variable overrides
	webhookURL := config.Slack.WebhookURL
	if envWebhook := os.Getenv("SLACK_WEBHOOK_URL"); envWebhook != "" {
		webhookURL = envWebhook
	}

	botToken := config.Slack.BotToken
	if envToken := os.Getenv("SLACK_BOT_TOKEN"); envToken != "" {
		botToken = envToken
	}

	channel := config.Slack.Channel
	if envChannel := os.Getenv("SLACK_CHANNEL"); envChannel != "" {
		channel = envChannel
	}

	notifier := slack.NewNotifier(
		webhookURL,
		botToken,
		channel,
		config.Slack.Username,
		config.Slack.IconEmoji,
	)

	return &NotificationManager{
		config:        config,
		slackNotifier: notifier,
	}
}

// NotifyPipelineStart sends a notification when pipeline starts
func (nm *NotificationManager) NotifyPipelineStart(commit, branch string) {
	if nm == nil || !nm.config.Slack.NotifyOnStart {
		return
	}

	nm.startTime = time.Now()
	nm.commit = commit
	nm.branch = branch

	_ = nm.slackNotifier.SendPipelineStart(commit, branch)
}

// NotifyPipelineComplete sends a notification when pipeline completes
func (nm *NotificationManager) NotifyPipelineComplete(success bool, model *ApiDocModel, errors []string) {
	if nm == nil || (!nm.config.Slack.NotifyOnComplete && success) || (!nm.config.Slack.NotifyOnFailure && !success) {
		return
	}

	duration := time.Since(nm.startTime)

	result := &slack.PipelineResult{
		Success:       success,
		Duration:      duration,
		Commit:        nm.commit,
		Branch:        nm.branch,
		TotalModules:  0,
		TotalServices: 0,
		TotalMessages: 0,
		Errors:        errors,
		Warnings:      []string{},
	}

	if model != nil {
		result.TotalModules = len(model.Modules)
		if stats, ok := model.Statistics["total_services"].(int64); ok {
			result.TotalServices = int(stats)
		}
		if stats, ok := model.Statistics["total_messages"].(int64); ok {
			result.TotalMessages = int(stats)
		}
	}

	_ = nm.slackNotifier.SendPipelineComplete(result)
}

// NotifyEnrichmentComplete sends a notification when enrichment completes
func (nm *NotificationManager) NotifyEnrichmentComplete(manifestPath string) {
	if nm == nil || !nm.config.Slack.NotifyOnEnrichment {
		return
	}

	// Load manifest
	manifest, err := loadManifestFromPath(manifestPath)
	if err != nil {
		return
	}

	result := &slack.EnrichmentResult{
		Success:         !manifest.HasFailures(),
		Duration:        manifest.Duration,
		TotalTargets:    manifest.Statistics.TotalTargets,
		EnrichedTargets: manifest.Statistics.EnrichedTargets,
		FailedTargets:   manifest.Statistics.FailedTargets,
		CacheHits:       manifest.Statistics.CacheHits,
		TotalTokens:     manifest.Statistics.TotalTokensUsed,
		Model:           manifest.ModelUsed,
		Provider:        manifest.ProviderUsed,
		SuccessRate:     manifest.Statistics.SuccessRate,
		CacheHitRate:    manifest.Statistics.CacheHitRate,
		Errors:          []string{},
	}

	_ = nm.slackNotifier.SendEnrichmentComplete(result)
}

// NotifyBreakingChanges sends a notification about breaking changes
func (nm *NotificationManager) NotifyBreakingChanges(changes []string) {
	if nm == nil || !nm.config.Slack.NotifyOnBreaking || len(changes) == 0 {
		return
	}

	_ = nm.slackNotifier.SendBreakingChanges(changes, nm.commit, nm.branch)
}

// NotifyReleaseNotes sends release notes to Slack
func (nm *NotificationManager) NotifyReleaseNotes(model *ApiDocModel, manifestPath string) {
	if nm == nil || !nm.config.Slack.NotifyReleaseNotes {
		return
	}

	// Generate release notes
	generator := slack.NewReleaseNotesGenerator(".")

	stats := slack.ReleaseStatistics{
		TotalServices: 0,
		TotalMessages: 0,
		TotalEnums:    0,
	}

	if model != nil {
		stats.TotalServices = len(collectAllServices(model))
		stats.TotalMessages = len(collectAllMessages(model))
		stats.TotalEnums = len(collectAllEnums(model))

		// Load enrichment stats if available
		if manifestPath != "" {
			if manifest, err := loadManifestFromPath(manifestPath); err == nil {
				stats.EnrichedTargets = manifest.Statistics.EnrichedTargets
				stats.CacheHitRate = manifest.Statistics.CacheHitRate
				stats.TotalTokensUsed = manifest.Statistics.TotalTokensUsed
				stats.SuccessRate = manifest.Statistics.SuccessRate
			}
		}
	}

	version := nm.config.Slack.ReleaseNotesVersion
	if version == "" {
		version = "latest"
	}

	notes, err := generator.GenerateFromCurrentState(version, stats)
	if err != nil {
		return
	}

	_ = nm.slackNotifier.SendReleaseNotes(notes)
}

// Helper to collect all services from model
func collectAllServices(model *ApiDocModel) []interface{} {
	var services []interface{}
	for _, module := range model.Modules {
		for _, svc := range module.Services {
			services = append(services, svc)
		}
	}
	return services
}

// Helper to collect all messages from model
func collectAllMessages(model *ApiDocModel) []interface{} {
	var messages []interface{}
	for _, module := range model.Modules {
		for _, msg := range module.Messages {
			messages = append(messages, msg)
		}
	}
	return messages
}

// Helper to collect all enums from model
func collectAllEnums(model *ApiDocModel) []interface{} {
	var enums []interface{}
	for _, module := range model.Modules {
		for _, enum := range module.Enums {
			enums = append(enums, enum)
		}
	}
	return enums
}

// Helper to load manifest from file
func loadManifestFromPath(path string) (*EnrichmentManifest, error) {
	// This would load the actual manifest file
	// For now, return a stub
	return &EnrichmentManifest{
		Statistics: ManifestStatistics{},
	}, nil
}

// Stub types for manifest (would be imported from enricher package)
type EnrichmentManifest struct {
	Duration       time.Duration
	ModelUsed      string
	ProviderUsed   string
	Statistics     ManifestStatistics
}

type ManifestStatistics struct {
	TotalTargets     int
	EnrichedTargets  int
	FailedTargets    int
	CacheHits        int
	TotalTokensUsed  int
	SuccessRate      float64
	CacheHitRate     float64
}

func (m *EnrichmentManifest) HasFailures() bool {
	return m.Statistics.FailedTargets > 0
}

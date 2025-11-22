package slack

import "time"

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationPipelineStart    NotificationType = "pipeline_start"
	NotificationPipelineComplete NotificationType = "pipeline_complete"
	NotificationPipelineFailed   NotificationType = "pipeline_failed"
	NotificationEnrichmentStart  NotificationType = "enrichment_start"
	NotificationEnrichmentComplete NotificationType = "enrichment_complete"
	NotificationEnrichmentFailed NotificationType = "enrichment_failed"
	NotificationBreakingChanges  NotificationType = "breaking_changes"
	NotificationReleaseNotes     NotificationType = "release_notes"
)

// Notification represents a notification to send to Slack
type Notification struct {
	Type      NotificationType
	Title     string
	Message   string
	Fields    []Field
	Color     string // "good", "warning", "danger", or hex color
	Timestamp time.Time
	Footer    string
	AuthorName string
	AuthorIcon string
}

// Field represents a field in a Slack message
type Field struct {
	Title string
	Value string
	Short bool
}

// WebhookMessage represents a Slack webhook message payload
type WebhookMessage struct {
	Text        string       `json:"text,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	Channel     string       `json:"channel,omitempty"`
	Username    string       `json:"username,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	IconURL     string       `json:"icon_url,omitempty"`
}

// Attachment represents a Slack message attachment
type Attachment struct {
	Fallback   string  `json:"fallback"`
	Color      string  `json:"color,omitempty"`
	Pretext    string  `json:"pretext,omitempty"`
	AuthorName string  `json:"author_name,omitempty"`
	AuthorLink string  `json:"author_link,omitempty"`
	AuthorIcon string  `json:"author_icon,omitempty"`
	Title      string  `json:"title,omitempty"`
	TitleLink  string  `json:"title_link,omitempty"`
	Text       string  `json:"text,omitempty"`
	Fields     []Field `json:"fields,omitempty"`
	Footer     string  `json:"footer,omitempty"`
	FooterIcon string  `json:"footer_icon,omitempty"`
	Timestamp  int64   `json:"ts,omitempty"`
	MarkdownIn []string `json:"mrkdwn_in,omitempty"`
}

// ReleaseNotes represents structured release notes
type ReleaseNotes struct {
	Version       string
	Date          time.Time
	Commit        string
	Branch        string
	NewFeatures   []string
	Improvements  []string
	BugFixes      []string
	BreakingChanges []string
	Deprecations  []string
	Statistics    ReleaseStatistics
}

// ReleaseStatistics contains metrics about the release
type ReleaseStatistics struct {
	TotalServices    int
	TotalMessages    int
	TotalEnums       int
	EnrichedTargets  int
	CacheHitRate     float64
	TotalTokensUsed  int
	SuccessRate      float64
}

// PipelineResult contains pipeline execution results
type PipelineResult struct {
	Success       bool
	Duration      time.Duration
	Commit        string
	Branch        string
	TotalModules  int
	TotalServices int
	TotalMessages int
	Errors        []string
	Warnings      []string
}

// EnrichmentResult contains enrichment execution results
type EnrichmentResult struct {
	Success         bool
	Duration        time.Duration
	TotalTargets    int
	EnrichedTargets int
	FailedTargets   int
	CacheHits       int
	TotalTokens     int
	Model           string
	Provider        string
	SuccessRate     float64
	CacheHitRate    float64
	Errors          []string
}

package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Notifier sends notifications to Slack
type Notifier struct {
	webhookURL string
	botToken   string
	channel    string
	username   string
	iconEmoji  string
	client     *http.Client
}

// NewNotifier creates a new Slack notifier
func NewNotifier(webhookURL, botToken, channel, username, iconEmoji string) *Notifier {
	return &Notifier{
		webhookURL: webhookURL,
		botToken:   botToken,
		channel:    channel,
		username:   username,
		iconEmoji:  iconEmoji,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendNotification sends a notification to Slack
func (n *Notifier) SendNotification(notification *Notification) error {
	message := n.buildWebhookMessage(notification)

	if n.webhookURL != "" {
		return n.sendViaWebhook(message)
	}

	if n.botToken != "" {
		return n.sendViaBotAPI(message)
	}

	return fmt.Errorf("no Slack webhook URL or bot token configured")
}

// buildWebhookMessage builds a Slack webhook message from a notification
func (n *Notifier) buildWebhookMessage(notification *Notification) *WebhookMessage {
	attachment := Attachment{
		Fallback:   notification.Title + ": " + notification.Message,
		Color:      notification.Color,
		Title:      notification.Title,
		Text:       notification.Message,
		Fields:     notification.Fields,
		Footer:     notification.Footer,
		Timestamp:  notification.Timestamp.Unix(),
		MarkdownIn: []string{"text", "fields"},
	}

	if notification.AuthorName != "" {
		attachment.AuthorName = notification.AuthorName
		attachment.AuthorIcon = notification.AuthorIcon
	}

	message := &WebhookMessage{
		Attachments: []Attachment{attachment},
	}

	if n.channel != "" {
		message.Channel = n.channel
	}
	if n.username != "" {
		message.Username = n.username
	}
	if n.iconEmoji != "" {
		message.IconEmoji = n.iconEmoji
	}

	return message
}

// sendViaWebhook sends a message via Slack webhook
func (n *Notifier) sendViaWebhook(message *WebhookMessage) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal webhook message: %w", err)
	}

	resp, err := n.client.Post(n.webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("send webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("slack webhook error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// sendViaBotAPI sends a message via Slack Bot API
func (n *Notifier) sendViaBotAPI(message *WebhookMessage) error {
	// For Bot API, we use chat.postMessage endpoint
	apiURL := "https://slack.com/api/chat.postMessage"

	payload, err := json.Marshal(map[string]interface{}{
		"channel":     n.channel,
		"text":        message.Text,
		"attachments": message.Attachments,
		"username":    message.Username,
		"icon_emoji":  message.IconEmoji,
	})
	if err != nil {
		return fmt.Errorf("marshal bot API message: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("create bot API request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+n.botToken)

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("send bot API request: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode bot API response: %w", err)
	}

	if ok, _ := result["ok"].(bool); !ok {
		errorMsg, _ := result["error"].(string)
		return fmt.Errorf("slack bot API error: %s", errorMsg)
	}

	return nil
}

// SendPipelineStart sends a notification when pipeline starts
func (n *Notifier) SendPipelineStart(commit, branch string) error {
	notification := &Notification{
		Type:      NotificationPipelineStart,
		Title:     "🚀 Documentation Pipeline Started",
		Message:   fmt.Sprintf("Building documentation for commit `%s` on branch `%s`", commit, branch),
		Color:     "#36a64f",
		Timestamp: time.Now(),
		Footer:    "ProtoDocs Pipeline",
		Fields: []Field{
			{Title: "Commit", Value: commit, Short: true},
			{Title: "Branch", Value: branch, Short: true},
		},
	}
	return n.SendNotification(notification)
}

// SendPipelineComplete sends a notification when pipeline completes
func (n *Notifier) SendPipelineComplete(result *PipelineResult) error {
	color := "good"
	if !result.Success {
		color = "danger"
	}

	fields := []Field{
		{Title: "Duration", Value: result.Duration.String(), Short: true},
		{Title: "Commit", Value: result.Commit, Short: true},
		{Title: "Branch", Value: result.Branch, Short: true},
		{Title: "Modules", Value: fmt.Sprintf("%d", result.TotalModules), Short: true},
		{Title: "Services", Value: fmt.Sprintf("%d", result.TotalServices), Short: true},
		{Title: "Messages", Value: fmt.Sprintf("%d", result.TotalMessages), Short: true},
	}

	message := fmt.Sprintf("Documentation pipeline completed successfully")
	if !result.Success {
		message = fmt.Sprintf("Documentation pipeline failed")
		if len(result.Errors) > 0 {
			message += fmt.Sprintf("\n\n*Errors:*\n```\n%s\n```", result.Errors[0])
		}
	}

	notification := &Notification{
		Type:      NotificationPipelineComplete,
		Title:     "✅ Documentation Pipeline Complete",
		Message:   message,
		Color:     color,
		Timestamp: time.Now(),
		Footer:    "ProtoDocs Pipeline",
		Fields:    fields,
	}

	return n.SendNotification(notification)
}

// SendEnrichmentComplete sends a notification when enrichment completes
func (n *Notifier) SendEnrichmentComplete(result *EnrichmentResult) error {
	color := "good"
	if !result.Success || result.SuccessRate < 0.9 {
		color = "warning"
	}

	fields := []Field{
		{Title: "Duration", Value: result.Duration.String(), Short: true},
		{Title: "Model", Value: fmt.Sprintf("%s/%s", result.Provider, result.Model), Short: true},
		{Title: "Total Targets", Value: fmt.Sprintf("%d", result.TotalTargets), Short: true},
		{Title: "Enriched", Value: fmt.Sprintf("%d", result.EnrichedTargets), Short: true},
		{Title: "Failed", Value: fmt.Sprintf("%d", result.FailedTargets), Short: true},
		{Title: "Cache Hits", Value: fmt.Sprintf("%d", result.CacheHits), Short: true},
		{Title: "Total Tokens", Value: fmt.Sprintf("%d", result.TotalTokens), Short: true},
		{Title: "Success Rate", Value: fmt.Sprintf("%.1f%%", result.SuccessRate*100), Short: true},
		{Title: "Cache Hit Rate", Value: fmt.Sprintf("%.1f%%", result.CacheHitRate*100), Short: true},
	}

	message := fmt.Sprintf("LLM enrichment completed with %.1f%% success rate", result.SuccessRate*100)
	if result.FailedTargets > 0 {
		message += fmt.Sprintf("\n\n⚠️ %d targets failed enrichment", result.FailedTargets)
	}

	notification := &Notification{
		Type:      NotificationEnrichmentComplete,
		Title:     "🤖 LLM Enrichment Complete",
		Message:   message,
		Color:     color,
		Timestamp: time.Now(),
		Footer:    "ProtoDocs Enricher",
		Fields:    fields,
	}

	return n.SendNotification(notification)
}

// SendBreakingChanges sends a notification about breaking changes
func (n *Notifier) SendBreakingChanges(changes []string, commit, branch string) error {
	changesText := ""
	for _, change := range changes {
		changesText += fmt.Sprintf("• %s\n", change)
	}

	notification := &Notification{
		Type:      NotificationBreakingChanges,
		Title:     "⚠️ Breaking Changes Detected",
		Message:   fmt.Sprintf("*Breaking changes detected in commit `%s`:*\n\n%s", commit, changesText),
		Color:     "danger",
		Timestamp: time.Now(),
		Footer:    "ProtoDocs Breaking Check",
		Fields: []Field{
			{Title: "Commit", Value: commit, Short: true},
			{Title: "Branch", Value: branch, Short: true},
			{Title: "Changes", Value: fmt.Sprintf("%d", len(changes)), Short: true},
		},
	}

	return n.SendNotification(notification)
}

// SendReleaseNotes sends release notes to Slack
func (n *Notifier) SendReleaseNotes(notes *ReleaseNotes) error {
	message := n.formatReleaseNotes(notes)

	fields := []Field{
		{Title: "Version", Value: notes.Version, Short: true},
		{Title: "Date", Value: notes.Date.Format("2006-01-02"), Short: true},
		{Title: "Commit", Value: notes.Commit, Short: true},
		{Title: "Branch", Value: notes.Branch, Short: true},
		{Title: "Services", Value: fmt.Sprintf("%d", notes.Statistics.TotalServices), Short: true},
		{Title: "Messages", Value: fmt.Sprintf("%d", notes.Statistics.TotalMessages), Short: true},
	}

	if notes.Statistics.EnrichedTargets > 0 {
		fields = append(fields, Field{
			Title: "LLM Enriched",
			Value: fmt.Sprintf("%d targets", notes.Statistics.EnrichedTargets),
			Short: true,
		})
	}

	color := "good"
	if len(notes.BreakingChanges) > 0 {
		color = "warning"
	}

	notification := &Notification{
		Type:      NotificationReleaseNotes,
		Title:     fmt.Sprintf("📝 Release Notes - %s", notes.Version),
		Message:   message,
		Color:     color,
		Timestamp: time.Now(),
		Footer:    "ProtoDocs Release",
		Fields:    fields,
	}

	return n.SendNotification(notification)
}

// formatReleaseNotes formats release notes for Slack
func (n *Notifier) formatReleaseNotes(notes *ReleaseNotes) string {
	message := ""

	if len(notes.NewFeatures) > 0 {
		message += "*✨ New Features:*\n"
		for _, feature := range notes.NewFeatures {
			message += fmt.Sprintf("• %s\n", feature)
		}
		message += "\n"
	}

	if len(notes.Improvements) > 0 {
		message += "*🔧 Improvements:*\n"
		for _, improvement := range notes.Improvements {
			message += fmt.Sprintf("• %s\n", improvement)
		}
		message += "\n"
	}

	if len(notes.BugFixes) > 0 {
		message += "*🐛 Bug Fixes:*\n"
		for _, fix := range notes.BugFixes {
			message += fmt.Sprintf("• %s\n", fix)
		}
		message += "\n"
	}

	if len(notes.BreakingChanges) > 0 {
		message += "*⚠️ Breaking Changes:*\n"
		for _, change := range notes.BreakingChanges {
			message += fmt.Sprintf("• %s\n", change)
		}
		message += "\n"
	}

	if len(notes.Deprecations) > 0 {
		message += "*🗑️ Deprecations:*\n"
		for _, deprecation := range notes.Deprecations {
			message += fmt.Sprintf("• %s\n", deprecation)
		}
	}

	return message
}

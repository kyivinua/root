package security

import (
	"regexp"
	"strings"
	"testing"
)

func TestSecretDetector_Scan(t *testing.T) {
	detector := NewSecretDetector()

	tests := []struct {
		name          string
		text          string
		wantDetections int
		wantType       string
		wantSeverity   string
	}{
		{
			name:          "AWS Access Key",
			text:          "aws_access_key_id = AKIAIOSFODNN7EXAMPLE",
			wantDetections: 1,
			wantType:       "AWS Access Key ID",
			wantSeverity:   "critical",
		},
		{
			name:          "GitHub Token",
			text:          "ghp_" + strings.Repeat("a", 36),
			wantDetections: 1,
			wantType:       "GitHub Token",
			wantSeverity:   "critical",
		},
		{
			name:          "Slack Token",
			text:          "slack_token=xoxb-EXAMPLE-TEST-NOT-REAL",
			wantDetections: 1,
			wantType:       "Slack Token",
			wantSeverity:   "critical",
		},
		{
			name:          "Anthropic API Key",
			text:          "ANTHROPIC_API_KEY=sk-ant-api03-" + strings.Repeat("a", 95),
			wantDetections: 1,
			wantType:       "Anthropic API Key",
			wantSeverity:   "critical",
		},
		{
			name:          "OpenAI API Key",
			text:          "OPENAI_API_KEY=sk-" + strings.Repeat("a", 48),
			wantDetections: 1,
			wantType:       "OpenAI API Key",
			wantSeverity:   "critical",
		},
		{
			name:          "Private Key",
			text:          "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEA...",
			wantDetections: 1,
			wantType:       "Private Key",
			wantSeverity:   "critical",
		},
		{
			name:          "Password in URL",
			text:          "jdbc:postgresql://user:password123@localhost:5432/db",
			wantDetections: 1,
			wantType:       "Password in URL",
			wantSeverity:   "critical",
		},
		{
			name:          "No secrets",
			text:          "This is a normal text without any secrets",
			wantDetections: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detections := detector.Scan(tt.text)

			if len(detections) != tt.wantDetections {
				t.Errorf("Scan() found %d detections, want %d", len(detections), tt.wantDetections)
			}

			if tt.wantDetections > 0 && len(detections) > 0 {
				if detections[0].Type != tt.wantType {
					t.Errorf("Scan() type = %s, want %s", detections[0].Type, tt.wantType)
				}
				if detections[0].Severity != tt.wantSeverity {
					t.Errorf("Scan() severity = %s, want %s", detections[0].Severity, tt.wantSeverity)
				}
				if detections[0].Redacted == "" {
					t.Error("Scan() redacted string is empty")
				}
			}
		})
	}
}

func TestSecretDetector_ScanFile(t *testing.T) {
	detector := NewSecretDetector()

	content := "api_key: sk-" + strings.Repeat("a", 48) + "\nslack_webhook: https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX\n"
	filename := "config.yaml"

	detections := detector.ScanFile(content, filename)

	if len(detections) != 2 {
		t.Errorf("ScanFile() found %d detections, want 2", len(detections))
	}

	for _, d := range detections {
		if !contains(d.Description, filename) {
			t.Errorf("ScanFile() detection description doesn't contain filename: %s", d.Description)
		}
	}
}

func TestHasCriticalSecrets(t *testing.T) {
	tests := []struct {
		name       string
		detections []Detection
		want       bool
	}{
		{
			name: "has critical",
			detections: []Detection{
				{Severity: "critical"},
				{Severity: "high"},
			},
			want: true,
		},
		{
			name: "no critical",
			detections: []Detection{
				{Severity: "high"},
				{Severity: "medium"},
			},
			want: false,
		},
		{
			name:       "empty",
			detections: []Detection{},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasCriticalSecrets(tt.detections); got != tt.want {
				t.Errorf("HasCriticalSecrets() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterBySeverity(t *testing.T) {
	detections := []Detection{
		{Type: "secret1", Severity: "critical"},
		{Type: "secret2", Severity: "high"},
		{Type: "secret3", Severity: "critical"},
		{Type: "secret4", Severity: "medium"},
	}

	filtered := FilterBySeverity(detections, "critical")

	if len(filtered) != 2 {
		t.Errorf("FilterBySeverity() returned %d detections, want 2", len(filtered))
	}

	for _, d := range filtered {
		if d.Severity != "critical" {
			t.Errorf("FilterBySeverity() returned detection with severity %s, want critical", d.Severity)
		}
	}
}

func TestAddPattern(t *testing.T) {
	detector := NewSecretDetector()
	initialCount := len(detector.patterns)

	customPattern := SecretPattern{
		Name:        "Custom Secret",
		Pattern:     regexp.MustCompile(`custom-secret-[0-9]+`),
		Description: "Custom secret pattern",
		Severity:    "high",
	}

	detector.AddPattern(customPattern)

	if len(detector.patterns) != initialCount+1 {
		t.Errorf("AddPattern() patterns count = %d, want %d", len(detector.patterns), initialCount+1)
	}

	// Test that the custom pattern works
	detections := detector.Scan("my custom-secret-12345 here")
	if len(detections) != 1 {
		t.Errorf("Custom pattern scan found %d detections, want 1", len(detections))
	}
}

func TestIsSecretAllowed(t *testing.T) {
	detector := NewSecretDetector()

	allowlist := []string{
		"EXAMPLE",
		"TEST_KEY",
		"placeholder",
	}

	tests := []struct {
		name   string
		secret string
		want   bool
	}{
		{
			name:   "allowed - contains EXAMPLE",
			secret: "AKIAIOSFODNN7EXAMPLE",
			want:   true,
		},
		{
			name:   "allowed - contains TEST_KEY",
			secret: "TEST_KEY_12345",
			want:   true,
		},
		{
			name:   "not allowed",
			secret: "sk-real-api-key-12345",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detector.IsSecretAllowed(tt.secret, allowlist); got != tt.want {
				t.Errorf("IsSecretAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

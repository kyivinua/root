package security

import (
	"regexp"
	"strings"
)

// SecretPattern represents a pattern for detecting secrets
type SecretPattern struct {
	Name        string
	Pattern     *regexp.Regexp
	Description string
	Severity    string // "critical", "high", "medium", "low"
}

// SecretDetector detects secrets in strings
type SecretDetector struct {
	patterns []SecretPattern
}

// Detection represents a detected secret
type Detection struct {
	Type        string
	Match       string
	Line        int
	Column      int
	Description string
	Severity    string
	Redacted    string // The string with the secret redacted
}

// NewSecretDetector creates a new secret detector with default patterns
func NewSecretDetector() *SecretDetector {
	return &SecretDetector{
		patterns: defaultPatterns(),
	}
}

// defaultPatterns returns common secret patterns
// NOTE: Order matters! More specific patterns should come first.
func defaultPatterns() []SecretPattern {
	return []SecretPattern{
		{
			Name:        "AWS Access Key ID",
			Pattern:     regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
			Description: "AWS Access Key ID detected",
			Severity:    "critical",
		},
		{
			Name:        "AWS Secret Access Key",
			Pattern:     regexp.MustCompile(`(?i)aws(.{0,20})?['\"][0-9a-zA-Z/+]{40}['\"]`),
			Description: "AWS Secret Access Key detected",
			Severity:    "critical",
		},
		{
			Name:        "GitHub Token",
			Pattern:     regexp.MustCompile(`ghp_[0-9a-zA-Z]{36}`),
			Description: "GitHub personal access token detected",
			Severity:    "critical",
		},
		{
			Name:        "Anthropic API Key",
			Pattern:     regexp.MustCompile(`sk-ant-api03-[0-9a-zA-Z-_]{95,}`),
			Description: "Anthropic API key detected",
			Severity:    "critical",
		},
		{
			Name:        "OpenAI API Key",
			Pattern:     regexp.MustCompile(`sk-[0-9a-zA-Z]{48}`),
			Description: "OpenAI API key detected",
			Severity:    "critical",
		},
		{
			Name:        "Slack Token",
			Pattern:     regexp.MustCompile(`xox[baprs]-[0-9a-zA-Z-]{10,}`),
			Description: "Slack token detected",
			Severity:    "critical",
		},
		{
			Name:        "Slack Webhook",
			Pattern:     regexp.MustCompile(`https://hooks\.slack\.com/services/[A-Z0-9/]+`),
			Description: "Slack webhook URL detected",
			Severity:    "high",
		},
		{
			Name:        "Google API Key",
			Pattern:     regexp.MustCompile(`AIza[0-9A-Za-z-_]{35}`),
			Description: "Google API key detected",
			Severity:    "critical",
		},
		{
			Name:        "Private Key",
			Pattern:     regexp.MustCompile(`-----BEGIN\s+(RSA\s+)?PRIVATE KEY-----`),
			Description: "Private key detected",
			Severity:    "critical",
		},
		{
			Name:        "Basic Auth",
			Pattern:     regexp.MustCompile(`(?i)basic\s+[a-zA-Z0-9+/=]{20,}`),
			Description: "Basic authentication credentials detected",
			Severity:    "high",
		},
		{
			Name:        "Bearer Token",
			Pattern:     regexp.MustCompile(`(?i)bearer\s+[a-zA-Z0-9\-_\.~+/=]{20,}`),
			Description: "Bearer token detected",
			Severity:    "high",
		},
		{
			Name:        "Password in URL",
			Pattern:     regexp.MustCompile(`[a-zA-Z]+://[^:]+:([^@]+)@`),
			Description: "Password in URL detected",
			Severity:    "critical",
		},
		{
			Name:        "Confluence API Token",
			Pattern:     regexp.MustCompile(`(?i)confluence[_-]?token['\"]?\s*[:=]\s*['\"]?([0-9a-zA-Z]{24,})['\"]?`),
			Description: "Confluence API token detected",
			Severity:    "high",
		},
	}
}

// Scan scans the given text for secrets
func (d *SecretDetector) Scan(text string) []Detection {
	var detections []Detection

	lines := strings.Split(text, "\n")
	for lineNum, line := range lines {
		for _, pattern := range d.patterns {
			matches := pattern.Pattern.FindAllStringIndex(line, -1)
			for _, match := range matches {
				detection := Detection{
					Type:        pattern.Name,
					Match:       line[match[0]:match[1]],
					Line:        lineNum + 1,
					Column:      match[0] + 1,
					Description: pattern.Description,
					Severity:    pattern.Severity,
					Redacted:    redactSecret(line, match[0], match[1]),
				}
				detections = append(detections, detection)
			}
		}
	}

	return detections
}

// ScanFile scans a file for secrets
func (d *SecretDetector) ScanFile(content string, filename string) []Detection {
	detections := d.Scan(content)
	// Add filename context to each detection
	for i := range detections {
		detections[i].Description = filename + ": " + detections[i].Description
	}
	return detections
}

// redactSecret redacts a secret from a string
func redactSecret(line string, start, end int) string {
	before := line[:start]
	after := line[end:]
	redacted := strings.Repeat("*", end-start)
	return before + redacted + after
}

// IsSecretAllowed checks if a secret is in the allowlist
func (d *SecretDetector) IsSecretAllowed(secret string, allowlist []string) bool {
	for _, allowed := range allowlist {
		if strings.Contains(secret, allowed) {
			return true
		}
	}
	return false
}

// AddPattern adds a custom pattern to the detector
func (d *SecretDetector) AddPattern(pattern SecretPattern) {
	d.patterns = append(d.patterns, pattern)
}

// HasCriticalSecrets returns true if any critical secrets were detected
func HasCriticalSecrets(detections []Detection) bool {
	for _, d := range detections {
		if d.Severity == "critical" {
			return true
		}
	}
	return false
}

// FilterByS everity filters detections by severity
func FilterBySeverity(detections []Detection, severity string) []Detection {
	var filtered []Detection
	for _, d := range detections {
		if d.Severity == severity {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

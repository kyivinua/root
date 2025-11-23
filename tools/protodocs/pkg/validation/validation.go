package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// MaxFileSize is the maximum allowed file size (100 MB)
	MaxFileSize = 100 * 1024 * 1024

	// MaxPathLength is the maximum allowed path length
	MaxPathLength = 4096
)

var (
	// apiKeyPattern matches common API key formats
	apiKeyPattern = regexp.MustCompile(`^[A-Za-z0-9_\-\.]+$`)

	// disallowedPathPatterns contains patterns that should not be in file paths
	disallowedPathPatterns = []string{
		"..",      // Directory traversal
		"~",       // Home directory
		"$",       // Environment variables
		"|",       // Command piping
		";",       // Command chaining
		"&",       // Command chaining
		"`",       // Command substitution
		"$(", ")", // Command substitution
	}
)

// ValidateFilePath validates and sanitizes a file path.
// Returns cleaned path and error if validation fails.
func ValidateFilePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	if len(path) > MaxPathLength {
		return "", fmt.Errorf("path too long: %d > %d", len(path), MaxPathLength)
	}

	// Check for disallowed patterns
	for _, pattern := range disallowedPathPatterns {
		if strings.Contains(path, pattern) {
			return "", fmt.Errorf("path contains disallowed pattern: %s", pattern)
		}
	}

	// Clean and make absolute
	cleanPath := filepath.Clean(path)

	// Ensure path doesn't escape the working directory
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("cannot resolve absolute path: %w", err)
	}

	return absPath, nil
}

// ValidateProtoFile validates a proto file path and checks if it exists.
func ValidateProtoFile(path string) (string, error) {
	cleanPath, err := ValidateFilePath(path)
	if err != nil {
		return "", err
	}

	// Check extension
	if filepath.Ext(cleanPath) != ".proto" {
		return "", fmt.Errorf("file must have .proto extension: %s", cleanPath)
	}

	// Check if file exists
	info, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("proto file does not exist: %s", cleanPath)
		}
		return "", fmt.Errorf("cannot stat file: %w", err)
	}

	// Check if it's a file
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory, not a file: %s", cleanPath)
	}

	// Check file size
	if info.Size() > MaxFileSize {
		return "", fmt.Errorf("file too large: %d > %d bytes", info.Size(), MaxFileSize)
	}

	return cleanPath, nil
}

// ValidateDirectory validates a directory path and checks if it exists.
func ValidateDirectory(path string) (string, error) {
	cleanPath, err := ValidateFilePath(path)
	if err != nil {
		return "", err
	}

	// Check if directory exists
	info, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("directory does not exist: %s", cleanPath)
		}
		return "", fmt.Errorf("cannot stat directory: %w", err)
	}

	// Check if it's a directory
	if !info.IsDir() {
		return "", fmt.Errorf("path is not a directory: %s", cleanPath)
	}

	return cleanPath, nil
}

// ValidateAPIKey validates an API key format.
func ValidateAPIKey(key string) error {
	if key == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	if len(key) < 16 {
		return fmt.Errorf("API key too short (minimum 16 characters)")
	}

	if len(key) > 512 {
		return fmt.Errorf("API key too long (maximum 512 characters)")
	}

	if !apiKeyPattern.MatchString(key) {
		return fmt.Errorf("API key contains invalid characters (allowed: A-Z, a-z, 0-9, _, -, .)")
	}

	return nil
}

// ValidateURL validates a URL format.
func ValidateURL(url string) error {
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}

	if len(url) > 2048 {
		return fmt.Errorf("URL too long (maximum 2048 characters)")
	}

	// Check for common injection patterns
	dangerous := []string{"javascript:", "data:", "file:", "vbscript:"}
	lowerURL := strings.ToLower(url)
	for _, pattern := range dangerous {
		if strings.Contains(lowerURL, pattern) {
			return fmt.Errorf("URL contains dangerous protocol: %s", pattern)
		}
	}

	return nil
}

// SanitizeString removes potentially dangerous characters from a string.
func SanitizeString(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters
	var sanitized strings.Builder
	for _, r := range input {
		// Keep printable characters and common whitespace
		if r >= 32 && r < 127 || r == '\n' || r == '\r' || r == '\t' {
			sanitized.WriteRune(r)
		}
	}

	return sanitized.String()
}

// ValidateCommandArgs validates command line arguments to prevent injection.
func ValidateCommandArgs(args []string) error {
	for i, arg := range args {
		// Check for command injection patterns
		dangerous := []string{";", "|", "&", "$", "`", "(", ")", "<", ">", "\n", "\r"}
		for _, pattern := range dangerous {
			if strings.Contains(arg, pattern) {
				return fmt.Errorf("argument %d contains dangerous character: %s", i, pattern)
			}
		}
	}
	return nil
}

// MaskSecret masks an API key or secret for logging.
// Shows first 4 and last 4 characters, masks the middle.
func MaskSecret(secret string) string {
	if len(secret) <= 8 {
		return "****"
	}

	if len(secret) <= 16 {
		return secret[:4] + "****"
	}

	masked := secret[:4] + strings.Repeat("*", len(secret)-8) + secret[len(secret)-4:]
	return masked
}

// ValidateSpaceKey validates Confluence space key format.
func ValidateSpaceKey(key string) error {
	if key == "" {
		return fmt.Errorf("space key cannot be empty")
	}

	// Confluence space keys are typically 2-255 alphanumeric characters
	if len(key) < 2 || len(key) > 255 {
		return fmt.Errorf("space key must be between 2 and 255 characters")
	}

	// Only allow alphanumeric and underscores
	matched, err := regexp.MatchString(`^[A-Z0-9_]+$`, key)
	if err != nil {
		return fmt.Errorf("regex error: %w", err)
	}

	if !matched {
		return fmt.Errorf("space key must contain only uppercase letters, numbers, and underscores")
	}

	return nil
}

// ValidatePageID validates Confluence page ID format.
func ValidatePageID(id string) error {
	if id == "" {
		return fmt.Errorf("page ID cannot be empty")
	}

	// Confluence page IDs are numeric strings
	matched, err := regexp.MatchString(`^\d+$`, id)
	if err != nil {
		return fmt.Errorf("regex error: %w", err)
	}

	if !matched {
		return fmt.Errorf("page ID must be numeric")
	}

	return nil
}

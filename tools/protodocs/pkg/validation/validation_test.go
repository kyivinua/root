package validation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFilePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid path", "test.proto", false},
		{"empty path", "", true},
		{"path traversal", "../../../etc/passwd", true},
		{"home directory", "~/test.proto", true},
		{"command injection", "test.proto;rm -rf /", true},
		{"pipe character", "test|other.proto", true},
		{"env variable", "$HOME/test.proto", true},
		{"command substitution", "$(cat test).proto", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateFilePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateProtoFile(t *testing.T) {
	// Create a temporary proto file
	tmpDir := t.TempDir()
	validProto := filepath.Join(tmpDir, "test.proto")
	if err := os.WriteFile(validProto, []byte("syntax = \"proto3\";"), 0644); err != nil {
		t.Fatal(err)
	}

	wrongExt := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(wrongExt, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid proto file", validProto, false},
		{"non-existent file", filepath.Join(tmpDir, "nonexistent.proto"), true},
		{"wrong extension", wrongExt, true},
		{"directory instead of file", tmpDir, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateProtoFile(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProtoFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid directory", tmpDir, false},
		{"non-existent directory", filepath.Join(tmpDir, "nonexistent"), true},
		{"file instead of directory", tmpFile, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateDirectory(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDirectory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAPIKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{"valid key", "sk-1234567890abcdef", false},
		{"empty key", "", true},
		{"too short", "short", true},
		{"too long", string(make([]byte, 513)), true},
		{"invalid characters", "key with spaces", true},
		{"valid with dash", "sk-test_key.12345", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAPIKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAPIKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid https", "https://example.com", false},
		{"valid http", "http://example.com", false},
		{"empty url", "", true},
		{"no protocol", "example.com", true},
		{"javascript protocol", "javascript:alert(1)", true},
		{"data protocol", "data:text/html,<script>alert(1)</script>", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"clean string", "hello world", "hello world"},
		{"with null bytes", "hello\x00world", "helloworld"},
		{"with control chars", "hello\x01\x02world", "helloworld"},
		{"with newlines", "hello\nworld", "hello\nworld"},
		{"with tabs", "hello\tworld", "hello\tworld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeString(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMaskSecret(t *testing.T) {
	tests := []struct {
		name   string
		secret string
		want   string
	}{
		{"short secret", "short", "****"},
		{"medium secret", "sk-1234567890", "sk-1****"},
		{"long secret", "sk-1234567890abcdef", "sk-1***********cdef"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskSecret(tt.secret)
			if got != tt.want {
				t.Errorf("MaskSecret() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateSpaceKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{"valid key", "APIDOCS", false},
		{"valid with numbers", "API123", false},
		{"valid with underscore", "API_DOCS", false},
		{"empty key", "", true},
		{"too short", "A", true},
		{"lowercase", "apidocs", true},
		{"with spaces", "API DOCS", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSpaceKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSpaceKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePageID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"valid ID", "123456789", false},
		{"empty ID", "", true},
		{"non-numeric", "abc123", true},
		{"with spaces", "123 456", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePageID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePageID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Benchmark tests
func BenchmarkValidateFilePath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ValidateFilePath("test/path/to/file.proto")
	}
}

func BenchmarkValidateAPIKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ValidateAPIKey("sk-1234567890abcdef")
	}
}

func BenchmarkMaskSecret(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MaskSecret("sk-1234567890abcdef1234567890")
	}
}

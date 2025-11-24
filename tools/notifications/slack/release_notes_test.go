package slack

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseChangelogFile(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		wantCount     int
		wantFirstVer  string
		wantFirstDate string
		wantErr       bool
	}{
		{
			name: "valid changelog with multiple releases",
			content: `# v1.2.0 (2025-01-15)

Commit: abc123 | Branch: main

## ✨ New Features

- Feature A
- Feature B

## 🐛 Bug Fixes

- Fix X

---

# v1.1.0 (2025-01-01)

Commit: def456 | Branch: main

## 🔧 Improvements

- Improvement Y
`,
			wantCount:     2,
			wantFirstVer:  "v1.2.0",
			wantFirstDate: "2025-01-15",
			wantErr:       false,
		},
		{
			name: "changelog with breaking changes",
			content: `# v2.0.0 (2025-02-01)

Commit: xyz789 | Branch: main

## ⚠️ BREAKING CHANGES

- API endpoint changed
- Configuration format updated

## ✨ New Features

- New feature C
`,
			wantCount:     1,
			wantFirstVer:  "v2.0.0",
			wantFirstDate: "2025-02-01",
			wantErr:       false,
		},
		{
			name:      "empty changelog",
			content:   "",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "changelog without commit info",
			content: `# v1.0.0 (2025-01-01)

## ✨ New Features

- Initial release
`,
			wantCount:     1,
			wantFirstVer:  "v1.0.0",
			wantFirstDate: "2025-01-01",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")

			if tt.content != "" {
				if err := os.WriteFile(changelogPath, []byte(tt.content), 0644); err != nil {
					t.Fatalf("Failed to create test changelog: %v", err)
				}
			}

			// Parse changelog
			releases, err := ParseChangelogFile(changelogPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseChangelogFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(releases) != tt.wantCount {
				t.Errorf("ParseChangelogFile() got %d releases, want %d", len(releases), tt.wantCount)
				return
			}

			if tt.wantCount > 0 {
				first := releases[0]
				if first.Version != tt.wantFirstVer {
					t.Errorf("First release version = %q, want %q", first.Version, tt.wantFirstVer)
				}

				expectedDate, _ := time.Parse("2006-01-02", tt.wantFirstDate)
				if !first.Date.Equal(expectedDate) {
					t.Errorf("First release date = %v, want %v", first.Date, expectedDate)
				}
			}
		})
	}
}

func TestParseChangelogFile_NonExistent(t *testing.T) {
	releases, err := ParseChangelogFile("/nonexistent/CHANGELOG.md")
	if err != nil {
		t.Errorf("ParseChangelogFile() should return empty slice for non-existent file, got error: %v", err)
	}
	if len(releases) != 0 {
		t.Errorf("ParseChangelogFile() got %d releases, want 0", len(releases))
	}
}

func TestParseChangelogFile_Sections(t *testing.T) {
	content := `# v1.0.0 (2025-01-01)

Commit: abc123 | Branch: main

## ⚠️ BREAKING CHANGES

- Breaking change 1
- Breaking change 2

## ✨ New Features

- Feature 1
- Feature 2
- Feature 3

## 🔧 Improvements

- Improvement 1

## 🐛 Bug Fixes

- Fix 1
- Fix 2

## 🗑️ Deprecations

- Deprecated API
`

	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")
	if err := os.WriteFile(changelogPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test changelog: %v", err)
	}

	releases, err := ParseChangelogFile(changelogPath)
	if err != nil {
		t.Fatalf("ParseChangelogFile() error = %v", err)
	}

	if len(releases) != 1 {
		t.Fatalf("Expected 1 release, got %d", len(releases))
	}

	release := releases[0]

	// Check breaking changes
	if len(release.BreakingChanges) != 2 {
		t.Errorf("BreakingChanges count = %d, want 2", len(release.BreakingChanges))
	}

	// Check new features
	if len(release.NewFeatures) != 3 {
		t.Errorf("NewFeatures count = %d, want 3", len(release.NewFeatures))
	}

	// Check improvements
	if len(release.Improvements) != 1 {
		t.Errorf("Improvements count = %d, want 1", len(release.Improvements))
	}

	// Check bug fixes
	if len(release.BugFixes) != 2 {
		t.Errorf("BugFixes count = %d, want 2", len(release.BugFixes))
	}

	// Check deprecations
	if len(release.Deprecations) != 1 {
		t.Errorf("Deprecations count = %d, want 1", len(release.Deprecations))
	}

	// Verify commit and branch
	if release.Commit != "abc123" {
		t.Errorf("Commit = %q, want %q", release.Commit, "abc123")
	}
	if release.Branch != "main" {
		t.Errorf("Branch = %q, want %q", release.Branch, "main")
	}
}

func TestAppendToChangelog(t *testing.T) {
	tests := []struct {
		name            string
		existingContent string
		newRelease      *ReleaseNotes
		wantContains    []string
	}{
		{
			name:            "append to empty file",
			existingContent: "",
			newRelease: &ReleaseNotes{
				Version: "v1.0.0",
				Date:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				Commit:  "abc123",
				Branch:  "main",
				NewFeatures: []string{
					"Initial release",
				},
			},
			wantContains: []string{
				"# v1.0.0 (2025-01-01)",
				"Commit: abc123 | Branch: main",
				"## ✨ New Features",
				"- Initial release",
			},
		},
		{
			name: "append to existing changelog",
			existingContent: `# v1.0.0 (2025-01-01)

Commit: old123 | Branch: main

## ✨ New Features

- Old feature
`,
			newRelease: &ReleaseNotes{
				Version: "v1.1.0",
				Date:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
				Commit:  "new456",
				Branch:  "main",
				NewFeatures: []string{
					"New feature",
				},
				BugFixes: []string{
					"Bug fix",
				},
			},
			wantContains: []string{
				"# v1.1.0 (2025-01-15)",
				"Commit: new456 | Branch: main",
				"- New feature",
				"- Bug fix",
				"---",
				"# v1.0.0 (2025-01-01)",
				"- Old feature",
			},
		},
		{
			name:            "append with breaking changes",
			existingContent: "",
			newRelease: &ReleaseNotes{
				Version: "v2.0.0",
				Date:    time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
				Commit:  "brk789",
				Branch:  "main",
				BreakingChanges: []string{
					"API changed",
					"Config format changed",
				},
				NewFeatures: []string{
					"New API version",
				},
			},
			wantContains: []string{
				"# v2.0.0 (2025-02-01)",
				"## ⚠️ BREAKING CHANGES",
				"- API changed",
				"- Config format changed",
				"## ✨ New Features",
				"- New API version",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")

			// Create existing changelog if provided
			if tt.existingContent != "" {
				if err := os.WriteFile(changelogPath, []byte(tt.existingContent), 0644); err != nil {
					t.Fatalf("Failed to create existing changelog: %v", err)
				}
			}

			// Create generator and append
			gen := NewReleaseNotesGenerator(".")
			if err := gen.AppendToChangelog(tt.newRelease, changelogPath); err != nil {
				t.Fatalf("AppendToChangelog() error = %v", err)
			}

			// Read result
			content, err := os.ReadFile(changelogPath)
			if err != nil {
				t.Fatalf("Failed to read changelog: %v", err)
			}

			contentStr := string(content)

			// Check all expected strings are present
			for _, want := range tt.wantContains {
				if !strings.Contains(contentStr, want) {
					t.Errorf("Changelog does not contain %q\nGot:\n%s", want, contentStr)
				}
			}

			// Check order: new release should come before old content
			if tt.existingContent != "" {
				newIdx := strings.Index(contentStr, tt.newRelease.Version)
				oldIdx := strings.Index(contentStr, "v1.0.0")
				if newIdx > oldIdx {
					t.Errorf("New release should come before old release")
				}
			}
		})
	}
}

func TestAppendToChangelog_PreservesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")

	// Create initial changelog
	initial := `# v1.0.0 (2025-01-01)

Commit: initial | Branch: main

## ✨ New Features

- Feature A
- Feature B
`

	if err := os.WriteFile(changelogPath, []byte(initial), 0644); err != nil {
		t.Fatalf("Failed to create initial changelog: %v", err)
	}

	// Append new release
	gen := NewReleaseNotesGenerator(".")
	newRelease := &ReleaseNotes{
		Version: "v1.1.0",
		Date:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Commit:  "new",
		Branch:  "main",
		BugFixes: []string{
			"Fix C",
		},
	}

	if err := gen.AppendToChangelog(newRelease, changelogPath); err != nil {
		t.Fatalf("AppendToChangelog() error = %v", err)
	}

	// Verify content
	content, err := os.ReadFile(changelogPath)
	if err != nil {
		t.Fatalf("Failed to read changelog: %v", err)
	}

	contentStr := string(content)

	// All original content should be preserved
	if !strings.Contains(contentStr, "Feature A") {
		t.Error("Original Feature A not preserved")
	}
	if !strings.Contains(contentStr, "Feature B") {
		t.Error("Original Feature B not preserved")
	}

	// New content should be present
	if !strings.Contains(contentStr, "Fix C") {
		t.Error("New fix C not found")
	}

	// Should have separator
	if !strings.Contains(contentStr, "---") {
		t.Error("Separator not found")
	}
}

func TestGetChangelog(t *testing.T) {
	gen := NewReleaseNotesGenerator(".")

	notes := &ReleaseNotes{
		Version: "v1.5.0",
		Date:    time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
		Commit:  "commit123",
		Branch:  "develop",
		BreakingChanges: []string{
			"API v1 removed",
		},
		NewFeatures: []string{
			"Feature X",
			"Feature Y",
		},
		Improvements: []string{
			"Performance boost",
		},
		BugFixes: []string{
			"Memory leak fixed",
		},
		Deprecations: []string{
			"Old method deprecated",
		},
		Statistics: ReleaseStatistics{
			TotalServices:   5,
			TotalMessages:   20,
			TotalEnums:      3,
			EnrichedTargets: 10,
			SuccessRate:     0.95,
		},
	}

	markdown := gen.GetChangelog(notes)

	// Check all sections are present
	expectedSections := []string{
		"# v1.5.0 (2025-03-01)",
		"Commit: commit123 | Branch: develop",
		"## ⚠️ BREAKING CHANGES",
		"- API v1 removed",
		"## ✨ New Features",
		"- Feature X",
		"- Feature Y",
		"## 🔧 Improvements",
		"- Performance boost",
		"## 🐛 Bug Fixes",
		"- Memory leak fixed",
		"## 🗑️ Deprecations",
		"- Old method deprecated",
		"## 📊 Statistics",
		"- Services: 5",
		"- Messages: 20",
		"- Enums: 3",
		"- LLM Enriched Targets: 10",
		"- Enrichment Success Rate: 95.0%",
	}

	for _, expected := range expectedSections {
		if !strings.Contains(markdown, expected) {
			t.Errorf("Changelog missing expected section: %q\nGot:\n%s", expected, markdown)
		}
	}
}

func TestGetChangelog_EmptySections(t *testing.T) {
	gen := NewReleaseNotesGenerator(".")

	notes := &ReleaseNotes{
		Version: "v1.0.0",
		Date:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Commit:  "abc",
		Branch:  "main",
		// Only new features, everything else empty
		NewFeatures: []string{
			"Initial release",
		},
	}

	markdown := gen.GetChangelog(notes)

	// Should have new features
	if !strings.Contains(markdown, "## ✨ New Features") {
		t.Error("Missing New Features section")
	}

	// Should NOT have empty sections
	if strings.Contains(markdown, "## 🐛 Bug Fixes") {
		t.Error("Should not include empty Bug Fixes section")
	}
	if strings.Contains(markdown, "## ⚠️ BREAKING CHANGES") {
		t.Error("Should not include empty Breaking Changes section")
	}
	if strings.Contains(markdown, "## 📊 Statistics") {
		t.Error("Should not include empty Statistics section")
	}
}

// Benchmark tests
func BenchmarkParseChangelogFile(b *testing.B) {
	content := `# v1.2.0 (2025-01-15)

Commit: abc123 | Branch: main

## ✨ New Features

- Feature A
- Feature B
- Feature C

## 🐛 Bug Fixes

- Fix X
- Fix Y

---

# v1.1.0 (2025-01-01)

Commit: def456 | Branch: main

## 🔧 Improvements

- Improvement Y
- Improvement Z
`

	tmpDir := b.TempDir()
	changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")
	if err := os.WriteFile(changelogPath, []byte(content), 0644); err != nil {
		b.Fatalf("Failed to create test changelog: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseChangelogFile(changelogPath)
	}
}

func BenchmarkAppendToChangelog(b *testing.B) {
	notes := &ReleaseNotes{
		Version: "v1.0.0",
		Date:    time.Now(),
		Commit:  "abc123",
		Branch:  "main",
		NewFeatures: []string{
			"Feature 1",
			"Feature 2",
		},
		BugFixes: []string{
			"Fix 1",
		},
	}

	gen := NewReleaseNotesGenerator(".")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmpDir := b.TempDir()
		changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")
		_ = gen.AppendToChangelog(notes, changelogPath)
	}
}

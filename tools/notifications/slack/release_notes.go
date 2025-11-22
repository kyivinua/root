package slack

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// ReleaseNotesGenerator generates release notes from git commits
type ReleaseNotesGenerator struct {
	repoPath string
}

// NewReleaseNotesGenerator creates a new release notes generator
func NewReleaseNotesGenerator(repoPath string) *ReleaseNotesGenerator {
	return &ReleaseNotesGenerator{
		repoPath: repoPath,
	}
}

// GenerateFromCommits generates release notes from git commits between two refs
func (g *ReleaseNotesGenerator) GenerateFromCommits(fromRef, toRef, version string) (*ReleaseNotes, error) {
	// Get commits between refs
	commits, err := g.getCommitsBetween(fromRef, toRef)
	if err != nil {
		return nil, fmt.Errorf("get commits: %w", err)
	}

	// Get current branch
	branch, err := g.getCurrentBranch()
	if err != nil {
		branch = "unknown"
	}

	// Get current commit
	commit, err := g.getCurrentCommit()
	if err != nil {
		commit = "unknown"
	}

	notes := &ReleaseNotes{
		Version: version,
		Date:    time.Now(),
		Commit:  commit,
		Branch:  branch,
		NewFeatures:     []string{},
		Improvements:    []string{},
		BugFixes:        []string{},
		BreakingChanges: []string{},
		Deprecations:    []string{},
	}

	// Parse commits and categorize
	for _, commitMsg := range commits {
		g.categorizeCommit(commitMsg, notes)
	}

	return notes, nil
}

// GenerateFromCurrentState generates release notes for current state
func (g *ReleaseNotesGenerator) GenerateFromCurrentState(version string, stats ReleaseStatistics) (*ReleaseNotes, error) {
	branch, err := g.getCurrentBranch()
	if err != nil {
		branch = "unknown"
	}

	commit, err := g.getCurrentCommit()
	if err != nil {
		commit = "unknown"
	}

	// Get recent commits (last 10)
	commits, err := g.getRecentCommits(10)
	if err != nil {
		commits = []string{}
	}

	notes := &ReleaseNotes{
		Version:         version,
		Date:            time.Now(),
		Commit:          commit,
		Branch:          branch,
		NewFeatures:     []string{},
		Improvements:    []string{},
		BugFixes:        []string{},
		BreakingChanges: []string{},
		Deprecations:    []string{},
		Statistics:      stats,
	}

	// Parse recent commits
	for _, commitMsg := range commits {
		g.categorizeCommit(commitMsg, notes)
	}

	return notes, nil
}

// categorizeCommit categorizes a commit message into release notes sections
func (g *ReleaseNotesGenerator) categorizeCommit(commitMsg string, notes *ReleaseNotes) {
	// Parse conventional commit format: type(scope): message
	// Examples:
	// feat: add new feature
	// feat!: breaking change
	// fix: bug fix
	// BREAKING CHANGE: description

	lines := strings.Split(commitMsg, "\n")
	if len(lines) == 0 {
		return
	}

	firstLine := strings.TrimSpace(lines[0])

	// Check for conventional commit format
	conventionalRegex := regexp.MustCompile(`^(\w+)(\([\w-]+\))?(!)?:\s*(.+)$`)
	matches := conventionalRegex.FindStringSubmatch(firstLine)

	if len(matches) > 0 {
		commitType := matches[1]
		hasBreaking := matches[3] == "!"
		message := matches[4]

		// Check for breaking change in body
		bodyHasBreaking := false
		for _, line := range lines[1:] {
			if strings.HasPrefix(strings.TrimSpace(line), "BREAKING CHANGE:") {
				bodyHasBreaking = true
				message = strings.TrimPrefix(strings.TrimSpace(line), "BREAKING CHANGE:")
				message = strings.TrimSpace(message)
				break
			}
		}

		if hasBreaking || bodyHasBreaking {
			notes.BreakingChanges = append(notes.BreakingChanges, message)
			return
		}

		switch strings.ToLower(commitType) {
		case "feat", "feature":
			notes.NewFeatures = append(notes.NewFeatures, message)
		case "fix":
			notes.BugFixes = append(notes.BugFixes, message)
		case "perf", "refactor", "improvement", "enhance":
			notes.Improvements = append(notes.Improvements, message)
		case "deprecate":
			notes.Deprecations = append(notes.Deprecations, message)
		}
	}
}

// getCommitsBetween gets commit messages between two refs
func (g *ReleaseNotesGenerator) getCommitsBetween(fromRef, toRef string) ([]string, error) {
	cmd := exec.Command("git", "log", fmt.Sprintf("%s..%s", fromRef, toRef), "--pretty=format:%B%n---COMMIT---")
	cmd.Dir = g.repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}

	commits := strings.Split(string(output), "---COMMIT---")
	result := []string{}
	for _, commit := range commits {
		trimmed := strings.TrimSpace(commit)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result, nil
}

// getRecentCommits gets recent N commit messages
func (g *ReleaseNotesGenerator) getRecentCommits(n int) ([]string, error) {
	cmd := exec.Command("git", "log", fmt.Sprintf("-%d", n), "--pretty=format:%B%n---COMMIT---")
	cmd.Dir = g.repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}

	commits := strings.Split(string(output), "---COMMIT---")
	result := []string{}
	for _, commit := range commits {
		trimmed := strings.TrimSpace(commit)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result, nil
}

// getCurrentBranch gets the current git branch
func (g *ReleaseNotesGenerator) getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = g.repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git branch: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// getCurrentCommit gets the current git commit hash
func (g *ReleaseNotesGenerator) getCurrentCommit() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = g.repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git commit: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetChangelog generates a markdown changelog
func (g *ReleaseNotesGenerator) GetChangelog(notes *ReleaseNotes) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("# %s (%s)\n\n", notes.Version, notes.Date.Format("2006-01-02")))
	builder.WriteString(fmt.Sprintf("Commit: %s | Branch: %s\n\n", notes.Commit, notes.Branch))

	if len(notes.BreakingChanges) > 0 {
		builder.WriteString("## ⚠️ BREAKING CHANGES\n\n")
		for _, change := range notes.BreakingChanges {
			builder.WriteString(fmt.Sprintf("- %s\n", change))
		}
		builder.WriteString("\n")
	}

	if len(notes.NewFeatures) > 0 {
		builder.WriteString("## ✨ New Features\n\n")
		for _, feature := range notes.NewFeatures {
			builder.WriteString(fmt.Sprintf("- %s\n", feature))
		}
		builder.WriteString("\n")
	}

	if len(notes.Improvements) > 0 {
		builder.WriteString("## 🔧 Improvements\n\n")
		for _, improvement := range notes.Improvements {
			builder.WriteString(fmt.Sprintf("- %s\n", improvement))
		}
		builder.WriteString("\n")
	}

	if len(notes.BugFixes) > 0 {
		builder.WriteString("## 🐛 Bug Fixes\n\n")
		for _, fix := range notes.BugFixes {
			builder.WriteString(fmt.Sprintf("- %s\n", fix))
		}
		builder.WriteString("\n")
	}

	if len(notes.Deprecations) > 0 {
		builder.WriteString("## 🗑️ Deprecations\n\n")
		for _, deprecation := range notes.Deprecations {
			builder.WriteString(fmt.Sprintf("- %s\n", deprecation))
		}
		builder.WriteString("\n")
	}

	// Add statistics if available
	if notes.Statistics.TotalServices > 0 {
		builder.WriteString("## 📊 Statistics\n\n")
		builder.WriteString(fmt.Sprintf("- Services: %d\n", notes.Statistics.TotalServices))
		builder.WriteString(fmt.Sprintf("- Messages: %d\n", notes.Statistics.TotalMessages))
		builder.WriteString(fmt.Sprintf("- Enums: %d\n", notes.Statistics.TotalEnums))
		if notes.Statistics.EnrichedTargets > 0 {
			builder.WriteString(fmt.Sprintf("- LLM Enriched Targets: %d\n", notes.Statistics.EnrichedTargets))
			builder.WriteString(fmt.Sprintf("- Enrichment Success Rate: %.1f%%\n", notes.Statistics.SuccessRate*100))
		}
	}

	return builder.String()
}

// ParseChangelogFile parses an existing CHANGELOG.md file
func ParseChangelogFile(filepath string) ([]*ReleaseNotes, error) {
	// TODO: Implement changelog parsing
	return nil, fmt.Errorf("not implemented")
}

// AppendToChangelog appends release notes to CHANGELOG.md
func (g *ReleaseNotesGenerator) AppendToChangelog(notes *ReleaseNotes, changelogPath string) error {
	// Generate markdown
	markdown := g.GetChangelog(notes)

	// Read existing changelog if it exists
	// Prepend new release notes to the top
	// Write back to file

	// For now, just write the markdown
	// TODO: Implement proper changelog management
	fmt.Println(markdown)
	return nil
}

package hldgen

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestNewContextEngine(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	cfg := ContextConfig{
		RAG: RAGConfig{
			Enabled: true,
			TopK:    5,
		},
	}

	engine := NewContextEngine(cfg, logger)
	if engine == nil {
		t.Fatal("NewContextEngine() returned nil")
	}

	if engine.cfg.RAG.TopK != 5 {
		t.Errorf("TopK = %d, want 5", engine.cfg.RAG.TopK)
	}
}

func TestFetchGitHistory_Success(t *testing.T) {
	// Check if git is available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available, skipping test")
	}

	// Create temporary git repository
	tmpDir := t.TempDir()

	// Initialize git repo
	initCmd := exec.Command("git", "init")
	initCmd.Dir = tmpDir
	if err := initCmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git
	configCmds := [][]string{
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test User"},
	}
	for _, args := range configCmds {
		cmd := exec.Command("git", args...)
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to configure git: %v", err)
		}
	}

	// Create a proto file and commit
	protoFile := filepath.Join(tmpDir, "test.proto")
	if err := os.WriteFile(protoFile, []byte("syntax = \"proto3\";"), 0644); err != nil {
		t.Fatalf("Failed to create proto file: %v", err)
	}

	addCmd := exec.Command("git", "add", "test.proto")
	addCmd.Dir = tmpDir
	if err := addCmd.Run(); err != nil {
		t.Fatalf("Failed to git add: %v", err)
	}

	commitCmd := exec.Command("git", "commit", "--no-gpg-sign", "-m", "Add test proto file")
	commitCmd.Dir = tmpDir
	if err := commitCmd.Run(); err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}

	// Change to temp directory for test
	oldDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	// Test fetchGitHistory
	logger := zerolog.New(os.Stdout)
	engine := NewContextEngine(ContextConfig{}, logger)

	ctx := context.Background()
	enriched := &EnrichedContext{}

	err := engine.fetchGitHistory(ctx, enriched)
	if err != nil {
		t.Fatalf("fetchGitHistory() error = %v", err)
	}

	if enriched.GitHistory == "" {
		t.Error("GitHistory should not be empty")
	}

	// Should contain commit message
	if !strings.Contains(enriched.GitHistory, "Add test proto file") {
		t.Errorf("GitHistory should contain commit message, got: %s", enriched.GitHistory)
	}
}

func TestFetchGitHistory_NoGitRepo(t *testing.T) {
	// Create temp directory without git
	tmpDir := t.TempDir()

	oldDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	logger := zerolog.New(os.Stdout)
	engine := NewContextEngine(ContextConfig{}, logger)

	ctx := context.Background()
	enriched := &EnrichedContext{}

	// Should not error, but should set fallback message
	err := engine.fetchGitHistory(ctx, enriched)
	if err != nil {
		t.Fatalf("fetchGitHistory() should not error on non-git repo: %v", err)
	}

	if enriched.GitHistory == "" {
		t.Error("GitHistory should have fallback message")
	}

	if !strings.Contains(enriched.GitHistory, "No recent git history available") {
		t.Errorf("GitHistory should contain fallback message, got: %s", enriched.GitHistory)
	}
}

func TestFetchGitHistory_NoRecentChanges(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available, skipping test")
	}

	tmpDir := t.TempDir()

	// Initialize git repo
	initCmd := exec.Command("git", "init")
	initCmd.Dir = tmpDir
	if err := initCmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git
	configCmds := [][]string{
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test User"},
	}
	for _, args := range configCmds {
		cmd := exec.Command("git", args...)
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to configure git: %v", err)
		}
	}

	// Create non-proto file
	txtFile := filepath.Join(tmpDir, "readme.txt")
	if err := os.WriteFile(txtFile, []byte("readme"), 0644); err != nil {
		t.Fatalf("Failed to create txt file: %v", err)
	}

	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = tmpDir
	_ = addCmd.Run()

	commitCmd := exec.Command("git", "commit", "-m", "Initial commit")
	commitCmd.Dir = tmpDir
	_ = commitCmd.Run()

	oldDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	logger := zerolog.New(os.Stdout)
	engine := NewContextEngine(ContextConfig{}, logger)

	ctx := context.Background()
	enriched := &EnrichedContext{}

	err := engine.fetchGitHistory(ctx, enriched)
	if err != nil {
		t.Fatalf("fetchGitHistory() error = %v", err)
	}

	// Should have message about no git history available
	if !strings.Contains(enriched.GitHistory, "No recent git history") {
		t.Errorf("Should indicate no git history, got: %s", enriched.GitHistory)
	}
}

func TestFetchOwnership_WithCodeowners(t *testing.T) {
	tmpDir := t.TempDir()

	// Create CODEOWNERS file
	codeownersContent := `# Code owners
* @team-platform @user1
*.proto @api-team @user2
/internal/ @internal-team
`
	codeownersPath := filepath.Join(tmpDir, "CODEOWNERS")
	if err := os.WriteFile(codeownersPath, []byte(codeownersContent), 0644); err != nil {
		t.Fatalf("Failed to create CODEOWNERS: %v", err)
	}

	oldDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	logger := zerolog.New(os.Stdout)
	engine := NewContextEngine(ContextConfig{}, logger)

	ctx := context.Background()
	enriched := &EnrichedContext{}

	err := engine.fetchOwnership(ctx, enriched)
	if err != nil {
		t.Fatalf("fetchOwnership() error = %v", err)
	}

	if enriched.Ownership == nil {
		t.Fatal("Ownership should not be nil")
	}

	// Check that owners were parsed
	if len(enriched.Ownership.Owners) == 0 {
		t.Error("Should have parsed owners")
	}

	// Check for specific owners
	foundPlatform := false
	foundUser1 := false
	foundApiTeam := false

	for _, owner := range enriched.Ownership.Owners {
		if owner == "team-platform" {
			foundPlatform = true
		}
		if owner == "user1" {
			foundUser1 = true
		}
		if owner == "api-team" {
			foundApiTeam = true
		}
	}

	if !foundPlatform {
		t.Error("Should find team-platform owner")
	}
	if !foundUser1 {
		t.Error("Should find user1 owner")
	}
	if !foundApiTeam {
		t.Error("Should find api-team owner")
	}

	// Should have team and slack
	if enriched.Ownership.Team == "" {
		t.Error("Team should be set")
	}
}

func TestFetchOwnership_GitHubLocation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .github directory
	githubDir := filepath.Join(tmpDir, ".github")
	if err := os.MkdirAll(githubDir, 0755); err != nil {
		t.Fatalf("Failed to create .github dir: %v", err)
	}

	// Create CODEOWNERS in .github
	codeownersContent := `* @github-team`
	codeownersPath := filepath.Join(githubDir, "CODEOWNERS")
	if err := os.WriteFile(codeownersPath, []byte(codeownersContent), 0644); err != nil {
		t.Fatalf("Failed to create CODEOWNERS: %v", err)
	}

	oldDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	logger := zerolog.New(os.Stdout)
	engine := NewContextEngine(ContextConfig{}, logger)

	ctx := context.Background()
	enriched := &EnrichedContext{}

	err := engine.fetchOwnership(ctx, enriched)
	if err != nil {
		t.Fatalf("fetchOwnership() error = %v", err)
	}

	// Should find github-team
	found := false
	for _, owner := range enriched.Ownership.Owners {
		if owner == "github-team" {
			found = true
		}
	}

	if !found {
		t.Errorf("Should find github-team, got owners: %v", enriched.Ownership.Owners)
	}
}

func TestFetchOwnership_NoCodeowners(t *testing.T) {
	tmpDir := t.TempDir()

	oldDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	logger := zerolog.New(os.Stdout)
	engine := NewContextEngine(ContextConfig{}, logger)

	ctx := context.Background()
	enriched := &EnrichedContext{}

	err := engine.fetchOwnership(ctx, enriched)
	if err != nil {
		t.Fatalf("fetchOwnership() error = %v", err)
	}

	// Should have defaults
	if enriched.Ownership == nil {
		t.Fatal("Ownership should not be nil")
	}

	if len(enriched.Ownership.Owners) == 0 {
		t.Error("Should have default owners")
	}

	if enriched.Ownership.Team == "" {
		t.Error("Should have default team")
	}
}

func TestFetchOwnership_Deduplication(t *testing.T) {
	tmpDir := t.TempDir()

	// Create CODEOWNERS with duplicate owners
	codeownersContent := `* @team-a @user1
*.proto @team-a @user2
/internal/ @user1 @team-b
`
	codeownersPath := filepath.Join(tmpDir, "CODEOWNERS")
	if err := os.WriteFile(codeownersPath, []byte(codeownersContent), 0644); err != nil {
		t.Fatalf("Failed to create CODEOWNERS: %v", err)
	}

	oldDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	logger := zerolog.New(os.Stdout)
	engine := NewContextEngine(ContextConfig{}, logger)

	ctx := context.Background()
	enriched := &EnrichedContext{}

	err := engine.fetchOwnership(ctx, enriched)
	if err != nil {
		t.Fatalf("fetchOwnership() error = %v", err)
	}

	// Count occurrences
	ownerCount := make(map[string]int)
	for _, owner := range enriched.Ownership.Owners {
		ownerCount[owner]++
	}

	// Check for duplicates
	for owner, count := range ownerCount {
		if count > 1 {
			t.Errorf("Owner %q appears %d times, should be deduplicated", owner, count)
		}
	}

	// Should have exactly 4 unique owners
	expectedOwners := map[string]bool{
		"team-a": true,
		"user1":  true,
		"user2":  true,
		"team-b": true,
	}

	if len(enriched.Ownership.Owners) != len(expectedOwners) {
		t.Errorf("Expected %d unique owners, got %d", len(expectedOwners), len(enriched.Ownership.Owners))
	}
}

func TestFetchOwnership_TeamExtraction(t *testing.T) {
	tests := []struct {
		name          string
		owner         string
		expectedTeam  string
		expectedSlack string
	}{
		{
			name:          "platform team",
			owner:         "team-platform",
			expectedTeam:  "Platform Engineering",
			expectedSlack: "#platform-eng",
		},
		{
			name:          "platform in owner name",
			owner:         "platform-team",
			expectedTeam:  "Platform Engineering",
			expectedSlack: "#platform-eng",
		},
		{
			name:  "other team",
			owner: "api-team",
			// Will use defaults
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			codeownersContent := fmt.Sprintf("* @%s", tt.owner)
			codeownersPath := filepath.Join(tmpDir, "CODEOWNERS")
			if err := os.WriteFile(codeownersPath, []byte(codeownersContent), 0644); err != nil {
				t.Fatalf("Failed to create CODEOWNERS: %v", err)
			}

			oldDir, _ := os.Getwd()
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("Failed to chdir: %v", err)
			}
			defer func() { _ = os.Chdir(oldDir) }()

			logger := zerolog.New(os.Stdout)
			engine := NewContextEngine(ContextConfig{}, logger)

			ctx := context.Background()
			enriched := &EnrichedContext{}

			err := engine.fetchOwnership(ctx, enriched)
			if err != nil {
				t.Fatalf("fetchOwnership() error = %v", err)
			}

			if tt.expectedTeam != "" && enriched.Ownership.Team != tt.expectedTeam {
				t.Errorf("Team = %q, want %q", enriched.Ownership.Team, tt.expectedTeam)
			}

			if tt.expectedSlack != "" && enriched.Ownership.Slack != tt.expectedSlack {
				t.Errorf("Slack = %q, want %q", enriched.Ownership.Slack, tt.expectedSlack)
			}
		})
	}
}

func TestBuildRAGQuery(t *testing.T) {
	docs := &ConsolidatedDocs{
		ModuleName: "user-service",
		Services: []ServiceDoc{
			{Name: "UserService"},
			{Name: "AuthService"},
			{Name: "ProfileService"},
		},
	}

	query := buildRAGQuery(docs)

	if !strings.Contains(query, "user-service") {
		t.Error("Query should contain module name")
	}

	if !strings.Contains(query, "UserService") {
		t.Error("Query should contain service names")
	}

	// Should include all 3 service names
	expectedServices := []string{"UserService", "AuthService", "ProfileService"}
	for _, svc := range expectedServices {
		if !strings.Contains(query, svc) {
			t.Errorf("Query should contain service %s", svc)
		}
	}
}

func TestBuildRAGQuery_EmptyServices(t *testing.T) {
	docs := &ConsolidatedDocs{
		ModuleName: "empty-module",
		Services:   []ServiceDoc{},
	}

	query := buildRAGQuery(docs)

	if !strings.Contains(query, "empty-module") {
		t.Error("Query should contain module name")
	}
}

// Benchmark tests
func BenchmarkFetchOwnership(b *testing.B) {
	tmpDir := b.TempDir()

	codeownersContent := `* @team-a @team-b @user1 @user2
*.proto @api-team
/internal/ @internal-team
`
	codeownersPath := filepath.Join(tmpDir, "CODEOWNERS")
	if err := os.WriteFile(codeownersPath, []byte(codeownersContent), 0644); err != nil {
		b.Fatalf("Failed to create CODEOWNERS: %v", err)
	}

	oldDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		b.Fatalf("Failed to chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	logger := zerolog.New(os.Stdout)
	engine := NewContextEngine(ContextConfig{}, logger)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enriched := &EnrichedContext{}
		_ = engine.fetchOwnership(ctx, enriched)
	}
}

func BenchmarkBuildRAGQuery(b *testing.B) {
	docs := &ConsolidatedDocs{
		ModuleName: "test-module",
		Services: []ServiceDoc{
			{Name: "Service1"},
			{Name: "Service2"},
			{Name: "Service3"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildRAGQuery(docs)
	}
}

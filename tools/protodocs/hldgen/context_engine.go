package hldgen

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// ContextEngine enriches documentation with external context sources
type ContextEngine struct {
	cfg    ContextConfig
	logger zerolog.Logger
}

// NewContextEngine creates a new context engine
func NewContextEngine(cfg ContextConfig, logger zerolog.Logger) *ContextEngine {
	return &ContextEngine{
		cfg:    cfg,
		logger: logger,
	}
}

// Enrich enriches the consolidated docs with external context
func (ce *ContextEngine) Enrich(ctx context.Context, docs *ConsolidatedDocs) (*EnrichedContext, error) {
	if !ce.cfg.Enabled {
		ce.logger.Debug().Msg("Context engine disabled")
		return &EnrichedContext{
			Docs:        docs,
			RAGContext:  []RAGDocument{},
			Sources:     make(map[string]ContextSource),
			GeneratedAt: time.Now(),
		}, nil
	}

	enriched := &EnrichedContext{
		Docs:        docs,
		RAGContext:  []RAGDocument{},
		Sources:     make(map[string]ContextSource),
		GeneratedAt: time.Now(),
	}

	ce.logger.Info().
		Int("sources", len(ce.cfg.Sources)).
		Msg("Starting context enrichment")

	// Process each context source
	for _, source := range ce.cfg.Sources {
		if err := ce.processSource(ctx, source, enriched); err != nil {
			ce.logger.Warn().
				Err(err).
				Str("source_type", source.Type).
				Msg("Failed to process context source")
			// Continue with other sources
			continue
		}
		enriched.Sources[source.Type] = source
	}

	// Fetch RAG context if enabled
	if ce.cfg.RAG.Enabled {
		ragDocs, err := ce.fetchRAGContext(ctx, docs)
		if err != nil {
			ce.logger.Warn().Err(err).Msg("RAG fetch failed")
		} else {
			enriched.RAGContext = ragDocs
			ce.logger.Info().
				Int("documents", len(ragDocs)).
				Msg("RAG context retrieved")
		}
	}

	ce.logger.Info().
		Int("rag_docs", len(enriched.RAGContext)).
		Int("sources", len(enriched.Sources)).
		Msg("Context enrichment completed")

	return enriched, nil
}

// processSource processes a specific context source
func (ce *ContextEngine) processSource(ctx context.Context, source ContextSource, enriched *EnrichedContext) error {
	ce.logger.Debug().
		Str("source_type", source.Type).
		Float64("weight", source.Weight).
		Msg("Processing context source")

	switch source.Type {
	case "git_history":
		return ce.fetchGitHistory(ctx, enriched)
	case "jira_tickets":
		return ce.fetchJIRATickets(ctx, source, enriched)
	case "ownership":
		return ce.fetchOwnership(ctx, enriched)
	case "slo_dashboard":
		return ce.fetchSLODashboard(ctx, source, enriched)
	case "security_policies":
		return ce.fetchSecurityPolicies(ctx, source, enriched)
	default:
		return fmt.Errorf("unknown source type: %s", source.Type)
	}
}

// fetchRAGContext retrieves relevant documents from vector store
func (ce *ContextEngine) fetchRAGContext(ctx context.Context, docs *ConsolidatedDocs) ([]RAGDocument, error) {
	if ce.cfg.RAG.VectorDB != "weaviate" {
		return nil, fmt.Errorf("unsupported vector DB: %s", ce.cfg.RAG.VectorDB)
	}

	// Build query from module and service names
	query := buildRAGQuery(docs)

	ce.logger.Debug().
		Str("query", query).
		Int("top_k", ce.cfg.RAG.TopK).
		Msg("Fetching RAG context")

	// TODO: Implement actual Weaviate integration
	// For now, return empty slice
	return []RAGDocument{}, nil
}

// fetchGitHistory fetches git commit history
func (ce *ContextEngine) fetchGitHistory(ctx context.Context, enriched *EnrichedContext) error {
	// Get git log for recent changes (last month)
	cmd := exec.CommandContext(ctx, "git", "log",
		"--oneline",
		"--since=1 month ago",
		"--",
		"*.proto", "**/*.proto")

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Git command failed, but this is not critical
		ce.logger.Warn().
			Err(err).
			Str("output", string(output)).
			Msg("Failed to fetch git history")
		enriched.GitHistory = "No recent git history available"
		return nil
	}

	history := strings.TrimSpace(string(output))
	if history == "" {
		enriched.GitHistory = "No proto file changes in the last month"
	} else {
		// Format the history nicely
		lines := strings.Split(history, "\n")
		if len(lines) > 20 {
			// Limit to 20 most recent commits
			lines = lines[:20]
			history = strings.Join(lines, "\n") + "\n... (truncated)"
		}
		enriched.GitHistory = history
	}

	return nil
}

// fetchJIRATickets fetches JIRA tickets related to the module
func (ce *ContextEngine) fetchJIRATickets(ctx context.Context, source ContextSource, enriched *EnrichedContext) error {
	jiraToken, ok := source.Config["jira_token"].(string)
	if !ok || jiraToken == "" {
		return fmt.Errorf("JIRA token not configured")
	}

	// TODO: Implement JIRA API integration
	// Search for tickets with labels matching module name
	enriched.JIRATickets = []JIRATicket{}
	return nil
}

// fetchOwnership fetches code ownership information
func (ce *ContextEngine) fetchOwnership(ctx context.Context, enriched *EnrichedContext) error {
	// Try to find CODEOWNERS file in common locations
	codeownersFiles := []string{
		"CODEOWNERS",
		".github/CODEOWNERS",
		".gitlab/CODEOWNERS",
		"docs/CODEOWNERS",
	}

	var owners []string
	var team string
	var slack string

	for _, path := range codeownersFiles {
		data, err := os.ReadFile(path)
		if err != nil {
			continue // Try next location
		}

		// Parse CODEOWNERS file
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())

			// Skip comments and empty lines
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			// Parse line: pattern @owner1 @owner2 ...
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}

			// Collect all @ mentions
			for _, part := range parts[1:] {
				if strings.HasPrefix(part, "@") {
					owner := strings.TrimPrefix(part, "@")
					// Deduplicate
					found := false
					for _, existing := range owners {
						if existing == owner {
							found = true
							break
						}
					}
					if !found {
						owners = append(owners, owner)

						// Try to extract team and slack from common patterns
						if strings.Contains(strings.ToLower(owner), "platform") {
							team = "Platform Engineering"
							slack = "#platform-eng"
						}
					}
				}
			}
		}

		if len(owners) > 0 {
			break // Found owners, stop searching
		}
	}

	// Set defaults if no CODEOWNERS found
	if len(owners) == 0 {
		owners = []string{"team-platform"}
		team = "Platform Engineering"
		slack = "#platform-eng"
	}

	enriched.Ownership = &OwnershipInfo{
		Team:   team,
		Owners: owners,
		Slack:  slack,
	}

	return nil
}

// fetchSLODashboard fetches SLO metrics from Grafana/similar
func (ce *ContextEngine) fetchSLODashboard(ctx context.Context, source ContextSource, enriched *EnrichedContext) error {
	grafanaToken, ok := source.Config["grafana_token"].(string)
	if !ok || grafanaToken == "" {
		return fmt.Errorf("grafana token not configured")
	}

	// TODO: Implement Grafana API integration
	enriched.SLODashboard = make(map[string]interface{})
	return nil
}

// fetchSecurityPolicies fetches security policies from Vault/OPA
func (ce *ContextEngine) fetchSecurityPolicies(ctx context.Context, source ContextSource, enriched *EnrichedContext) error {
	vaultToken, ok := source.Config["vault_token"].(string)
	if !ok || vaultToken == "" {
		return fmt.Errorf("vault token not configured")
	}

	// TODO: Implement Vault/OPA integration
	enriched.SecurityPolicies = []SecurityPolicy{
		{
			Name:        "Zero Trust",
			Type:        "network",
			Description: "All service-to-service communication must use mTLS",
			Rules:       []string{"REQUIRE_MTLS", "DENY_HTTP"},
		},
	}
	return nil
}

// buildRAGQuery builds a query string for RAG retrieval
func buildRAGQuery(docs *ConsolidatedDocs) string {
	query := fmt.Sprintf("Module: %s", docs.ModuleName)

	if len(docs.Services) > 0 {
		query += " Services:"
		for i, svc := range docs.Services {
			if i >= 3 {
				break
			}
			query += " " + svc.Name
		}
	}

	return query
}

// ValidateContext validates enriched context
func ValidateContext(ctx *EnrichedContext) error {
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}
	if ctx.Docs == nil {
		return fmt.Errorf("consolidated docs is nil")
	}
	return nil
}

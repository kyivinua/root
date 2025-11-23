package hldgen

import (
	"context"
	"fmt"
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
	// TODO: Implement git history fetching
	// git log --oneline --since="1 month ago" -- <proto files>
	enriched.GitHistory = "Git history enrichment not yet implemented"
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
	// TODO: Implement ownership parsing from CODEOWNERS or similar
	enriched.Ownership = &OwnershipInfo{
		Team:   "Platform Engineering",
		Owners: []string{"team-platform"},
		Slack:  "#platform-eng",
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

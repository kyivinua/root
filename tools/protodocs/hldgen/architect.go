package hldgen

import (
	"context"
	"fmt"
	"strings"
	"time"

)

// ArchitectAgent represents the Principal Architect agent
type ArchitectAgent struct {
	BaseAgent
}

// NewArchitectAgent creates a new Architect agent
func NewArchitectAgent(cfg AgentConfig) *ArchitectAgent {
	return &ArchitectAgent{
		BaseAgent: BaseAgent{
			role:        RoleArchitect,
			weight:      cfg.Weight,
			model:       cfg.Model,
			temperature: cfg.Temperature,
			maxTokens:   cfg.MaxTokens,
		},
	}
}

// Think generates architectural design based on input
func (a *ArchitectAgent) Think(ctx context.Context, input *AgentInput) (*AgentResponse, error) {
	// Build prompt for architect (for production LLM integration)
	_ = a.buildPrompt(input)

	// For this prototype, generate a structured response
	// In production, this would call LLM with instructor for structured output
	content := a.generateArchitecture(input)

	diagrams := a.generateDiagrams(input)

	return &AgentResponse{
		Role:        a.role,
		Content:     content,
		Diagrams:    diagrams,
		Confidence:  0.85,
		TokensUsed:  2500,
		GeneratedAt: time.Now(),
	}, nil
}

// buildPrompt constructs the prompt for the architect
func (a *ArchitectAgent) buildPrompt(input *AgentInput) string {
	var sb strings.Builder

	sb.WriteString("<system>\n")
	sb.WriteString("You are a Principal Architect. Generate High-Level Design following Stripe/Cloudflare standards.\n")
	sb.WriteString("Use Chain-of-Thought in <thinking>, output in <output> as structured content.\n")
	sb.WriteString("</system>\n\n")

	sb.WriteString("<user>\n")
	sb.WriteString(fmt.Sprintf("Module: %s\n", input.Docs.ModuleName))
	sb.WriteString(fmt.Sprintf("Services: %d\n", len(input.Docs.Services)))
	sb.WriteString(fmt.Sprintf("Messages: %d\n", len(input.Docs.Messages)))

	if input.EnrichedCtx != nil && len(input.EnrichedCtx.RAGContext) > 0 {
		sb.WriteString("\nRAG Context:\n")
		for _, doc := range input.EnrichedCtx.RAGContext {
			sb.WriteString(fmt.Sprintf("- %s (score: %.2f)\n", doc.Source, doc.Score))
		}
	}

	if input.PreviousDraft != nil && input.Round > 1 {
		sb.WriteString("\n<previous_draft>\n")
		sb.WriteString(input.PreviousDraft.Architecture)
		sb.WriteString("\n</previous_draft>\n")
	}

	if len(input.Criticism) > 0 {
		sb.WriteString("\n<criticism>\n")
		for _, crit := range input.Criticism {
			if crit.Agent == string(a.role) {
				sb.WriteString(fmt.Sprintf("Score: %.2f\n", crit.Score))
				sb.WriteString(fmt.Sprintf("Issues: %d\n", len(crit.Issues)))
				sb.WriteString(fmt.Sprintf("Suggestion: %s\n", crit.Suggestion))
			}
		}
		sb.WriteString("</criticism>\n")
	}

	sb.WriteString("\n<thinking>\n")
	sb.WriteString("1. Identify bounded contexts\n")
	sb.WriteString("2. Design component architecture\n")
	sb.WriteString("3. Define data flows\n")
	sb.WriteString("4. Specify async patterns (Kafka/SQS)\n")
	sb.WriteString("5. Determine critical paths\n")
	sb.WriteString("6. Plan multi-region resilience\n")
	sb.WriteString("</thinking>\n")

	sb.WriteString("\n<output>\n")
	sb.WriteString("Generate structured architecture section\n")
	sb.WriteString("</output>\n")
	sb.WriteString("</user>\n")

	return sb.String()
}

// generateArchitecture generates architecture content (prototype)
func (a *ArchitectAgent) generateArchitecture(input *AgentInput) string {
	var sb strings.Builder

	sb.WriteString("## Architecture\n\n")
	sb.WriteString("### Bounded Contexts\n\n")

	// Extract bounded contexts from services
	contexts := a.extractBoundedContexts(input.Docs)
	for _, ctx := range contexts {
		sb.WriteString(fmt.Sprintf("- **%s**: %s\n", ctx.Name, ctx.Description))
	}

	sb.WriteString("\n### Component Architecture\n\n")
	sb.WriteString("```mermaid\n")
	sb.WriteString(a.generateComponentDiagram(input.Docs))
	sb.WriteString("```\n\n")

	sb.WriteString("### Data Flows\n\n")
	for _, svc := range input.Docs.Services {
		sb.WriteString(fmt.Sprintf("**%s**:\n", svc.Name))
		for _, method := range svc.Methods {
			sb.WriteString(fmt.Sprintf("- %s: %s → %s\n", method.Name, method.InputType, method.OutputType))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("### Async Patterns\n\n")
	sb.WriteString("- **Event Bus**: Kafka for inter-service communication\n")
	sb.WriteString("- **Message Queue**: SQS for async processing\n")
	sb.WriteString("- **Circuit Breaker**: Resilience4j on all external calls\n")
	sb.WriteString("- **Retry**: Exponential backoff with max 3 attempts\n\n")

	return sb.String()
}

// extractBoundedContexts extracts bounded contexts from services
func (a *ArchitectAgent) extractBoundedContexts(docs *ConsolidatedDocs) []BoundedContext {
	contexts := make([]BoundedContext, 0)

	// Simple heuristic: group by service prefix
	serviceGroups := make(map[string][]ServiceDoc)
	for _, svc := range docs.Services {
		prefix := a.extractPrefix(svc.Name)
		serviceGroups[prefix] = append(serviceGroups[prefix], svc)
	}

	for prefix, services := range serviceGroups {
		contexts = append(contexts, BoundedContext{
			Name:        prefix,
			Description: fmt.Sprintf("%s domain with %d services", prefix, len(services)),
			Services:    services,
		})
	}

	return contexts
}

// extractPrefix extracts prefix from service name
func (a *ArchitectAgent) extractPrefix(name string) string {
	parts := strings.Split(name, "Service")
	if len(parts) > 0 {
		return parts[0]
	}
	return name
}

// generateDiagrams generates Mermaid diagrams
func (a *ArchitectAgent) generateDiagrams(input *AgentInput) map[string]string {
	diagrams := make(map[string]string)

	diagrams["component"] = a.generateComponentDiagram(input.Docs)
	diagrams["sequence"] = a.generateSequenceDiagram(input.Docs)

	return diagrams
}

// generateComponentDiagram generates a component diagram
func (a *ArchitectAgent) generateComponentDiagram(docs *ConsolidatedDocs) string {
	var sb strings.Builder

	sb.WriteString("graph TB\n")
	sb.WriteString("    subgraph \"Frontend\"\n")
	sb.WriteString("        Client[Client Application]\n")
	sb.WriteString("    end\n")
	sb.WriteString("    subgraph \"API Layer\"\n")
	sb.WriteString("        Gateway[API Gateway]\n")
	sb.WriteString("    end\n")
	sb.WriteString("    subgraph \"Core Services\"\n")

	for i, svc := range docs.Services {
		sb.WriteString(fmt.Sprintf("        S%d[%s]\n", i, svc.Name))
	}

	sb.WriteString("    end\n")
	sb.WriteString("    subgraph \"Data Layer\"\n")
	sb.WriteString("        DB[(Database)]\n")
	sb.WriteString("        Cache[(Redis Cache)]\n")
	sb.WriteString("    end\n\n")

	// Connections
	sb.WriteString("    Client --> Gateway\n")
	for i := range docs.Services {
		sb.WriteString(fmt.Sprintf("    Gateway --> S%d\n", i))
		sb.WriteString(fmt.Sprintf("    S%d --> DB\n", i))
		sb.WriteString(fmt.Sprintf("    S%d --> Cache\n", i))
	}

	return sb.String()
}

// generateSequenceDiagram generates a sequence diagram for critical path
func (a *ArchitectAgent) generateSequenceDiagram(docs *ConsolidatedDocs) string {
	var sb strings.Builder

	sb.WriteString("sequenceDiagram\n")
	sb.WriteString("    participant Client\n")
	sb.WriteString("    participant Gateway\n")

	if len(docs.Services) > 0 {
		svc := docs.Services[0]
		sb.WriteString(fmt.Sprintf("    participant %s\n", svc.Name))
		sb.WriteString("    participant Database\n\n")

		if len(svc.Methods) > 0 {
			method := svc.Methods[0]
			sb.WriteString(fmt.Sprintf("    Client->>Gateway: %s\n", method.Name))
			sb.WriteString(fmt.Sprintf("    Gateway->>%s: %s\n", svc.Name, method.Name))
			sb.WriteString(fmt.Sprintf("    %s->>Database: Query\n", svc.Name))
			sb.WriteString(fmt.Sprintf("    Database-->>%s: Result\n", svc.Name))
			sb.WriteString(fmt.Sprintf("    %s-->>Gateway: %s\n", svc.Name, method.OutputType))
			sb.WriteString("    Gateway-->>Client: Response\n")
		}
	}

	return sb.String()
}

// BoundedContext represents a bounded context in DDD
type BoundedContext struct {
	Name        string
	Description string
	Services    []ServiceDoc
}

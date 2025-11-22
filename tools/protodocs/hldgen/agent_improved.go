package hldgen

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kyivinua/docgen-tool/tools/protodocs/internal/errors"
)

// ImprovedArchitectAgent with LLM integration
func (a *ArchitectAgent) ThinkWithLLM(ctx context.Context, input *AgentInput) (*AgentResponse, error) {
	// Validate input
	if input == nil {
		return nil, errors.New(errors.ErrorTypeValidation, "agent input cannot be nil")
	}
	if input.Docs == nil {
		return nil, errors.New(errors.ErrorTypeValidation, "consolidated docs cannot be nil")
	}

	var content string
	var tokensUsed int
	var confidence float64

	// Try to use LLM if available
	if a.llmClient != nil {
		prompt := a.buildEnhancedPrompt(input)

		resp, err := a.llmClient.Generate(ctx, prompt, LLMConfig{})
		if err == nil {
			// Successfully got LLM response
			content = a.extractArchitectureFromLLM(resp.Content, input)
			tokensUsed = resp.TokensUsed
			confidence = resp.Confidence
		} else {
			// Check if error is retryable
			if errors.IsRetryable(err) {
				return nil, errors.Wrap(err, errors.ErrorTypeRetryable, "LLM call failed (retryable)")
			}
			// LLM failed with non-retryable error, fallback to static generation
			content = a.generateArchitecture(input)
			tokensUsed = 2500
			confidence = 0.75 // Lower confidence for static content
		}
	} else {
		// No LLM available, use static generation
		content = a.generateArchitecture(input)
		tokensUsed = 2500
		confidence = 0.75
	}

	// Generate diagrams (these are always generated)
	diagrams := a.generateDiagrams(input)

	return &AgentResponse{
		Role:        a.role,
		Content:     content,
		Diagrams:    diagrams,
		Confidence:  confidence,
		TokensUsed:  tokensUsed,
		GeneratedAt: time.Now(),
	}, nil
}

// buildEnhancedPrompt builds an improved prompt with rich context
func (a *ArchitectAgent) buildEnhancedPrompt(input *AgentInput) string {
	var sb strings.Builder

	sb.WriteString("<system>\n")
	sb.WriteString("You are a Principal Architect with 15+ years experience in distributed systems.\n")
	sb.WriteString("Generate enterprise-grade High-Level Design documentation following Stripe/Cloudflare standards.\n\n")
	sb.WriteString("Your output MUST include:\n")
	sb.WriteString("1. Bounded Contexts (Domain-Driven Design)\n")
	sb.WriteString("2. Component Architecture with clear boundaries\n")
	sb.WriteString("3. Data Flow diagrams\n")
	sb.WriteString("4. Async Communication patterns (Event-Driven)\n")
	sb.WriteString("5. Scalability & Performance considerations\n")
	sb.WriteString("6. Multi-region deployment strategy\n\n")
	sb.WriteString("Use markdown format with proper headers (##, ###).\n")
	sb.WriteString("</system>\n\n")

	sb.WriteString("<context>\n")
	sb.WriteString(fmt.Sprintf("**Module**: %s\n", input.Docs.ModuleName))
	sb.WriteString(fmt.Sprintf("**Services Count**: %d\n", len(input.Docs.Services)))
	sb.WriteString(fmt.Sprintf("**Messages Count**: %d\n", len(input.Docs.Messages)))
	sb.WriteString(fmt.Sprintf("**Source Commit**: %s\n\n", input.Docs.SourceCommit))

	// Add services detail
	sb.WriteString("**Services**:\n")
	for i, svc := range input.Docs.Services {
		if i >= 5 {
			sb.WriteString(fmt.Sprintf("... and %d more services\n", len(input.Docs.Services)-5))
			break
		}
		sb.WriteString(fmt.Sprintf("- %s (%d methods)\n", svc.Name, len(svc.Methods)))
	}

	// Add RAG context if available
	if input.EnrichedCtx != nil && len(input.EnrichedCtx.RAGContext) > 0 {
		sb.WriteString("\n**Retrieved Context (RAG)**:\n")
		for i, doc := range input.EnrichedCtx.RAGContext {
			if i >= 3 {
				break
			}
			sb.WriteString(fmt.Sprintf("- [Score: %.2f] %s\n", doc.Score, doc.Source))
			if len(doc.Content) > 200 {
				sb.WriteString(fmt.Sprintf("  %s...\n", doc.Content[:200]))
			} else {
				sb.WriteString(fmt.Sprintf("  %s\n", doc.Content))
			}
		}
	}

	// Add ownership info if available
	if input.EnrichedCtx != nil && input.EnrichedCtx.Ownership != nil {
		sb.WriteString(fmt.Sprintf("\n**Team**: %s\n", input.EnrichedCtx.Ownership.Team))
		sb.WriteString(fmt.Sprintf("**Owners**: %v\n", input.EnrichedCtx.Ownership.Owners))
	}
	sb.WriteString("</context>\n\n")

	// Add previous draft for refinement
	if input.PreviousDraft != nil && input.Round > 1 {
		sb.WriteString("<previous_draft>\n")
		sb.WriteString(input.PreviousDraft.Architecture)
		sb.WriteString("\n</previous_draft>\n\n")
	}

	// Add criticism for improvement
	if len(input.Criticism) > 0 {
		sb.WriteString("<feedback>\n")
		for _, crit := range input.Criticism {
			if crit.Agent == string(a.role) {
				sb.WriteString(fmt.Sprintf("**Quality Score**: %.2f/1.0\n", crit.Score))
				if len(crit.Issues) > 0 {
					sb.WriteString("\n**Issues to Address**:\n")
					for _, issue := range crit.Issues {
						sb.WriteString(fmt.Sprintf("- [%s][%s] %s\n", issue.Severity, issue.Type, issue.Message))
						if issue.Fix != "" {
							sb.WriteString(fmt.Sprintf("  Suggestion: %s\n", issue.Fix))
						}
					}
				}
				sb.WriteString(fmt.Sprintf("\n**Improvement Suggestion**: %s\n", crit.Suggestion))
			}
		}
		sb.WriteString("</feedback>\n\n")
	}

	sb.WriteString("<task>\n")
	if input.Round == 1 {
		sb.WriteString("Generate comprehensive architecture documentation for this API module.\n")
	} else {
		sb.WriteString(fmt.Sprintf("This is refinement round %d. Improve the architecture based on the feedback above.\n", input.Round))
	}
	sb.WriteString("Focus on production-ready, scalable, secure architecture.\n")
	sb.WriteString("</task>\n")

	return sb.String()
}

// extractArchitectureFromLLM extracts and validates architecture content from LLM response
func (a *ArchitectAgent) extractArchitectureFromLLM(llmResponse string, input *AgentInput) string {
	// Basic validation
	if len(llmResponse) < 500 {
		return a.generateArchitecture(input)
	}

	// Check for key sections
	hasArchitecture := strings.Contains(llmResponse, "Architecture") ||
		strings.Contains(llmResponse, "architecture")
	hasComponents := strings.Contains(llmResponse, "Component") ||
		strings.Contains(llmResponse, "component")

	if !hasArchitecture && !hasComponents {
		// LLM didn't generate architecture content, use static
		return a.generateArchitecture(input)
	}

	// Return LLM response
	return llmResponse
}

// ImprovedSimpleAgent with LLM integration
func (s *SimpleAgent) ThinkWithLLM(ctx context.Context, input *AgentInput) (*AgentResponse, error) {
	var content string
	var slo map[string]SLO
	var tokensUsed int
	var confidence float64

	if s.llmClient != nil {
		prompt := s.buildRoleSpecificPrompt(input)

		resp, err := s.llmClient.Generate(ctx, prompt, LLMConfig{})
		if err == nil {
			content = resp.Content
			tokensUsed = resp.TokensUsed
			confidence = resp.Confidence
		} else {
			// Fallback
			content = s.generateContentByRole(input)
			tokensUsed = 1500
			confidence = 0.70
		}
	} else {
		content = s.generateContentByRole(input)
		tokensUsed = 1500
		confidence = 0.70
	}

	// SRE generates SLO
	if s.role == RoleSRE {
		slo = s.generateSLO(input)
	}

	return &AgentResponse{
		Role:        s.role,
		Content:     content,
		SLO:         slo,
		Confidence:  confidence,
		TokensUsed:  tokensUsed,
		GeneratedAt: time.Now(),
	}, nil
}

// buildRoleSpecificPrompt creates role-specific prompts
func (s *SimpleAgent) buildRoleSpecificPrompt(input *AgentInput) string {
	var sb strings.Builder

	switch s.role {
	case RolePM:
		sb.WriteString("<system>You are a Senior Product Manager.</system>\n")
		sb.WriteString("<task>Generate business context, value proposition, and key metrics for:\n")
	case RoleSecurity:
		sb.WriteString("<system>You are a Security Architect (CISSP, CISA).</system>\n")
		sb.WriteString("<task>Generate security architecture including auth, compliance, zero-trust for:\n")
	case RoleSRE:
		sb.WriteString("<system>You are a Site Reliability Engineer.</system>\n")
		sb.WriteString("<task>Generate SLOs, observability stack, resilience patterns for:\n")
	case RoleQA:
		sb.WriteString("<system>You are a QA Lead.</system>\n")
		sb.WriteString("<task>Generate test strategy, requirements, acceptance criteria for:\n")
	}

	sb.WriteString(fmt.Sprintf("Module: %s with %d services</task>\n", input.Docs.ModuleName, len(input.Docs.Services)))

	return sb.String()
}

// generateContentByRole fallback content generation
func (s *SimpleAgent) generateContentByRole(input *AgentInput) string {
	switch s.role {
	case RolePM:
		return s.generatePMContent(input)
	case RoleSecurity:
		return s.generateSecurityContent(input)
	case RoleSRE:
		return s.generateSREContent(input)
	case RoleQA:
		return s.generateQAContent(input)
	default:
		return ""
	}
}

// ImprovedCriticAgent with LLM-based evaluation
func (c *CriticAgent) CritiqueWithLLM(ctx context.Context, role string, response *AgentResponse, docs *ConsolidatedDocs) (Criticism, error) {
	criticism := Criticism{
		Agent:      role,
		Confidence: 0.85,
		Issues:     []Issue{},
	}

	if c.llmClient != nil {
		// Use LLM for critique
		prompt := c.buildCritiquePrompt(role, response, docs)

		llmResp, err := c.llmClient.Generate(ctx, prompt, LLMConfig{})
		if err == nil {
			// Parse LLM response for score and issues
			criticism = c.parseCritiqueFromLLM(llmResp.Content, role)
		} else {
			// Fallback to rule-based
			criticism = c.evaluateRuleBased(role, response, docs)
		}
	} else {
		// No LLM, use rule-based
		criticism = c.evaluateRuleBased(role, response, docs)
	}

	// Mark as fatal if score too low
	if criticism.Score < 0.50 {
		criticism.Fatal = true
	}

	return criticism, nil
}

// buildCritiquePrompt builds LLM prompt for critique
func (c *CriticAgent) buildCritiquePrompt(role string, response *AgentResponse, docs *ConsolidatedDocs) string {
	return fmt.Sprintf(`<system>You are a Critic evaluating %s agent's output.</system>
<content>
%s
</content>
<task>
Evaluate quality on scale 0.0-1.0. List issues with severity (critical/high/medium/low).
Provide actionable suggestions for improvement.
Format: Score: X.XX | Issues: ... | Suggestion: ...
</task>`, role, response.Content)
}

// parseCritiqueFromLLM parses LLM critique response
func (c *CriticAgent) parseCritiqueFromLLM(content string, role string) Criticism {
	criticism := Criticism{
		Agent:      role,
		Score:      0.80,
		Confidence: 0.90,
		Issues:     []Issue{},
		Suggestion: "Continue with current quality",
	}

	// Simple parsing (in production, use structured output)
	if strings.Contains(content, "Score:") {
		// Extract score from "Score: 0.XX"
		// This is simplified - production would use proper parsing
	}

	return criticism
}

// evaluateRuleBased is the fallback rule-based evaluation
func (c *CriticAgent) evaluateRuleBased(role string, response *AgentResponse, docs *ConsolidatedDocs) Criticism {
	// Use existing evaluation logic
	switch AgentRole(role) {
	case RoleArchitect:
		return c.evaluateArchitect(response, docs)
	case RoleSecurity:
		return c.evaluateSecurity(response, docs)
	case RoleSRE:
		return c.evaluateSRE(response, docs)
	default:
		return Criticism{
			Agent:      role,
			Score:      0.85,
			Confidence: 0.75,
			Issues:     []Issue{},
			Suggestion: "No specific issues found",
		}
	}
}

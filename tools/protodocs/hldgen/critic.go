package hldgen

import (
	"context"
	"fmt"
	"math"
	"strings"
)

// CriticAgent evaluates agent responses and provides structured criticism
type CriticAgent struct {
	BaseAgent
	consensusThreshold float64
}

// NewCriticAgent creates a new Critic agent
func NewCriticAgent(cfg AgentConfig, consensusThreshold float64) *CriticAgent {
	return &CriticAgent{
		BaseAgent: BaseAgent{
			role:        RoleCritic,
			weight:      1.0, // Critic doesn't participate in consensus
			model:       cfg.Model,
			temperature: 0.0, // Deterministic evaluation
			maxTokens:   cfg.MaxTokens,
		},
		consensusThreshold: consensusThreshold,
	}
}

// Think is not used for Critic (use CritiqueAgent instead)
func (c *CriticAgent) Think(ctx context.Context, input *AgentInput) (*AgentResponse, error) {
	return nil, fmt.Errorf("critic agent should use CritiqueAgent method")
}

// CritiqueAgent evaluates a single agent's response
func (c *CriticAgent) CritiqueAgent(
	ctx context.Context,
	role string,
	response *AgentResponse,
	allResponses map[string]*AgentResponse,
	docs *ConsolidatedDocs,
) (Criticism, error) {

	criticism := Criticism{
		Agent:      role,
		Confidence: 0.9,
		Issues:     make([]Issue, 0),
	}

	// Evaluate based on role-specific criteria
	switch AgentRole(role) {
	case RoleArchitect:
		criticism = c.evaluateArchitect(response, docs)
	case RoleSecurity:
		criticism = c.evaluateSecurity(response, docs)
	case RoleSRE:
		criticism = c.evaluateSRE(response, docs)
	case RolePM:
		criticism = c.evaluatePM(response, docs)
	case RoleQA:
		criticism = c.evaluateQA(response, docs)
	}

	// Check for critical issues
	for _, issue := range criticism.Issues {
		if issue.Severity == SeverityCritical {
			criticism.Fatal = true
			break
		}
	}

	return criticism, nil
}

// evaluateArchitect evaluates architect's response
func (c *CriticAgent) evaluateArchitect(response *AgentResponse, docs *ConsolidatedDocs) Criticism {
	criticism := Criticism{
		Agent:  string(RoleArchitect),
		Score:  0.9, // Start with high score
		Issues: make([]Issue, 0),
	}

	// Check for component diagram
	if _, hasComponentDiagram := response.Diagrams["component"]; !hasComponentDiagram {
		criticism.Issues = append(criticism.Issues, Issue{
			Type:     IssueTypeMissingComponent,
			Severity: SeverityHigh,
			Path:     "diagrams.component",
			Message:  "Missing component diagram",
			Evidence: "No component diagram found in response",
			Fix:      "Add Mermaid component diagram showing system architecture",
		})
		criticism.Score -= 0.1
	}

	// Check for data flows
	if !c.hasDataFlows(response.Content) {
		criticism.Issues = append(criticism.Issues, Issue{
			Type:     IssueTypeIncorrectFlow,
			Severity: SeverityMedium,
			Path:     "architecture.data_flows",
			Message:  "Data flows not clearly defined",
			Evidence: "No explicit data flow documentation found",
			Fix:      "Add data flow diagrams and descriptions",
		})
		criticism.Score -= 0.05
	}

	// Generate suggestion
	if len(criticism.Issues) > 0 {
		criticism.Suggestion = c.generateArchitectSuggestion(criticism.Issues)
	} else {
		criticism.Suggestion = "Architecture looks comprehensive. Consider adding failure modes and recovery patterns."
	}

	return criticism
}

// evaluateSecurity evaluates security agent's response
func (c *CriticAgent) evaluateSecurity(response *AgentResponse, docs *ConsolidatedDocs) Criticism {
	criticism := Criticism{
		Agent:  string(RoleSecurity),
		Score:  0.9,
		Issues: make([]Issue, 0),
	}

	// Check for mTLS
	if !c.hasMTLS(response.Content) {
		criticism.Issues = append(criticism.Issues, Issue{
			Type:     IssueTypeSecurityGap,
			Severity: SeverityCritical,
			Path:     "security.transport",
			Message:  "Missing mTLS configuration",
			Evidence: "No mTLS (mutual TLS) mentioned for service-to-service communication",
			Fix:      "Implement mTLS for all inter-service communication",
		})
		criticism.Score -= 0.2
		criticism.Fatal = true
	}

	// Check for auth/authz
	if !c.hasAuth(response.Content) {
		criticism.Issues = append(criticism.Issues, Issue{
			Type:     IssueTypeSecurityGap,
			Severity: SeverityHigh,
			Path:     "security.authentication",
			Message:  "Authentication/Authorization not specified",
			Evidence: "No OAuth2, JWT, or authorization mechanism described",
			Fix:      "Add OAuth2 + JWT authentication with RBAC authorization",
		})
		criticism.Score -= 0.15
	}

	if len(criticism.Issues) > 0 {
		criticism.Suggestion = "Security gaps identified. Prioritize mTLS and auth/authz implementation."
	} else {
		criticism.Suggestion = "Security design is solid. Consider adding rate limiting and DDoS protection."
	}

	return criticism
}

// evaluateSRE evaluates SRE agent's response
func (c *CriticAgent) evaluateSRE(response *AgentResponse, docs *ConsolidatedDocs) Criticism{
	criticism := Criticism{
		Agent:  string(RoleSRE),
		Score:  0.9,
		Issues: make([]Issue, 0),
	}

	// Check for SLO
	if len(response.SLO) == 0 {
		criticism.Issues = append(criticism.Issues, Issue{
			Type:     IssueTypeMissingSLO,
			Severity: SeverityHigh,
			Path:     "observability.slo",
			Message:  "No SLO (Service Level Objectives) defined",
			Evidence: "SLO map is empty",
			Fix:      "Define SLO for availability, latency, and error rate",
		})
		criticism.Score -= 0.15
	}

	// Check for observability
	if !c.hasObservability(response.Content) {
		criticism.Issues = append(criticism.Issues, Issue{
			Type:     IssueTypeMissingComponent,
			Severity: SeverityMedium,
			Path:     "observability.tracing",
			Message:  "Observability stack not defined",
			Evidence: "No mention of tracing, metrics, or logging",
			Fix:      "Add OpenTelemetry for tracing, Prometheus for metrics, structured logging",
		})
		criticism.Score -= 0.1
	}

	if len(criticism.Issues) > 0 {
		criticism.Suggestion = "SRE design needs SLO and observability improvements."
	} else {
		criticism.Suggestion = "SRE design is comprehensive. Consider chaos engineering plans."
	}

	return criticism
}

// evaluatePM evaluates PM agent's response
func (c *CriticAgent) evaluatePM(response *AgentResponse, docs *ConsolidatedDocs) Criticism {
	criticism := Criticism{
		Agent:  string(RolePM),
		Score:  0.85,
		Issues: make([]Issue, 0),
	}

	// PM evaluation is more lenient
	if !c.hasBusinessContext(response.Content) {
		criticism.Issues = append(criticism.Issues, Issue{
			Type:     IssueTypeMissingComponent,
			Severity: SeverityLow,
			Path:     "business_context",
			Message:  "Business context could be more detailed",
			Evidence: "Value proposition not clearly articulated",
			Fix:      "Add customer segments, revenue streams, and key metrics",
		})
		criticism.Score -= 0.05
	}

	criticism.Suggestion = "Business context looks good. Consider adding user journeys."

	return criticism
}

// evaluateQA evaluates QA agent's response
func (c *CriticAgent) evaluateQA(response *AgentResponse, docs *ConsolidatedDocs) Criticism {
	criticism := Criticism{
		Agent:  string(RoleQA),
		Score:  0.85,
		Issues: make([]Issue, 0),
	}

	if !c.hasTestStrategy(response.Content) {
		criticism.Issues = append(criticism.Issues, Issue{
			Type:     IssueTypeMissingComponent,
			Severity: SeverityMedium,
			Path:     "requirements.test_strategy",
			Message:  "Test strategy not defined",
			Evidence: "No testing approach specified",
			Fix:      "Add unit, integration, and e2e test strategy",
		})
		criticism.Score -= 0.1
	}

	criticism.Suggestion = "QA design is acceptable. Consider contract testing for APIs."

	return criticism
}

// EvaluateConsensus calculates weighted consensus score
func (c *CriticAgent) EvaluateConsensus(criticisms []Criticism) float64 {
	var totalScore float64
	var totalWeight float64
	var criticalCount int

	weights := map[string]float64{
		string(RoleArchitect): 0.30,
		string(RoleSecurity):  0.25,
		string(RoleSRE):       0.20,
		string(RolePM):        0.15,
		string(RoleQA):        0.10,
	}

	for _, crit := range criticisms {
		weight := weights[crit.Agent]
		totalScore += crit.Score * weight
		totalWeight += weight

		if crit.Fatal {
			criticalCount++
		}
	}

	consensus := totalScore / totalWeight

	// Any critical issue forces refinement
	if criticalCount > 0 {
		return 0.0
	}

	return math.Round(consensus*100) / 100
}

// Helper methods for content checking
func (c *CriticAgent) hasDataFlows(content string) bool {
	keywords := []string{"data flow", "dataflow", "flow diagram", "sequence"}
	return c.containsAny(content, keywords)
}

func (c *CriticAgent) hasMTLS(content string) bool {
	keywords := []string{"mtls", "mutual tls", "m-tls", "mTLS"}
	return c.containsAny(content, keywords)
}

func (c *CriticAgent) hasAuth(content string) bool {
	keywords := []string{"oauth", "jwt", "authentication", "authorization", "auth"}
	return c.containsAny(content, keywords)
}

func (c *CriticAgent) hasObservability(content string) bool {
	keywords := []string{"observability", "tracing", "metrics", "logging", "opentelemetry"}
	return c.containsAny(content, keywords)
}

func (c *CriticAgent) hasBusinessContext(content string) bool {
	keywords := []string{"business", "value proposition", "customer", "revenue"}
	return c.containsAny(content, keywords)
}

func (c *CriticAgent) hasTestStrategy(content string) bool {
	keywords := []string{"test", "testing", "qa", "quality assurance"}
	return c.containsAny(content, keywords)
}

func (c *CriticAgent) containsAny(content string, keywords []string) bool {
	contentLower := strings.ToLower(content)
	for _, keyword := range keywords {
		if strings.Contains(contentLower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func (c *CriticAgent) generateArchitectSuggestion(issues []Issue) string {
	if len(issues) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("Improvements needed:\n")
	for i, issue := range issues {
		sb.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, issue.Type, issue.Fix))
	}
	return sb.String()
}

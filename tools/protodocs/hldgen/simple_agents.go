package hldgen

import (
	"context"
	"fmt"
	"strings"
	"time"

)

// SimpleAgent provides a simple implementation for PM, Security, SRE, QA agents
type SimpleAgent struct {
	BaseAgent
}

// NewSimpleAgent creates a new simple agent
func NewSimpleAgent(role AgentRole, cfg AgentConfig) *SimpleAgent {
	return &SimpleAgent{
		BaseAgent: BaseAgent{
			role:        role,
			weight:      cfg.Weight,
			model:       cfg.Model,
			temperature: cfg.Temperature,
			maxTokens:   cfg.MaxTokens,
		},
	}
}

// Think generates a response based on role
func (s *SimpleAgent) Think(ctx context.Context, input *AgentInput) (*AgentResponse, error) {
	var content string
	var slo map[string]SLO

	switch s.role {
	case RolePM:
		content = s.generatePMContent(input)
	case RoleSecurity:
		content = s.generateSecurityContent(input)
	case RoleSRE:
		content = s.generateSREContent(input)
		slo = s.generateSLO(input)
	case RoleQA:
		content = s.generateQAContent(input)
	}

	return &AgentResponse{
		Role:        s.role,
		Content:     content,
		SLO:         slo,
		Confidence:  0.80,
		TokensUsed:  1500,
		GeneratedAt: time.Now(),
	}, nil
}

// generatePMContent generates Product Manager content
func (s *SimpleAgent) generatePMContent(input *AgentInput) string {
	var sb strings.Builder

	sb.WriteString("## Business Context\n\n")
	sb.WriteString(fmt.Sprintf("### %s API\n\n", input.Docs.ModuleName))

	sb.WriteString("**Value Proposition**:\n")
	sb.WriteString(fmt.Sprintf("- Core API for %s domain\n", input.Docs.ModuleName))
	sb.WriteString(fmt.Sprintf("- Serves %d distinct services\n", len(input.Docs.Services)))
	sb.WriteString("- Enables critical business workflows\n\n")

	sb.WriteString("**Customer Segments**:\n")
	sb.WriteString("- Internal services\n")
	sb.WriteString("- Partner integrations\n")
	sb.WriteString("- Public API consumers\n\n")

	sb.WriteString("**Key Metrics**:\n")
	sb.WriteString("- API Calls: Target 10M+ requests/day\n")
	sb.WriteString("- Success Rate: 99.9%+\n")
	sb.WriteString("- Response Time: p95 < 200ms\n\n")

	return sb.String()
}

// generateSecurityContent generates Security content
func (s *SimpleAgent) generateSecurityContent(input *AgentInput) string {
	var sb strings.Builder

	sb.WriteString("## Security & Compliance\n\n")

	sb.WriteString("### Authentication & Authorization\n")
	sb.WriteString("- **Authentication**: OAuth2 with JWT tokens\n")
	sb.WriteString("- **Authorization**: RBAC (Role-Based Access Control)\n")
	sb.WriteString("- **Session Management**: Redis-backed sessions with 15min TTL\n\n")

	sb.WriteString("### Transport Security\n")
	sb.WriteString("- **Inter-service**: mTLS (mutual TLS) for all service-to-service communication\n")
	sb.WriteString("- **Client-facing**: TLS 1.3+\n")
	sb.WriteString("- **Certificate Management**: Automated rotation via cert-manager\n\n")

	sb.WriteString("### Compliance\n")
	sb.WriteString("- **GDPR**: Data residency in EU, user consent management\n")
	sb.WriteString("- **PCI-DSS**: No card data stored, tokenization via Stripe\n")
	sb.WriteString("- **SOC2**: Automated audit logging, access controls\n\n")

	sb.WriteString("### Security Boundaries\n")
	sb.WriteString("- Zero Trust Architecture\n")
	sb.WriteString("- Network policies enforced via Kubernetes NetworkPolicies\n")
	sb.WriteString("- Secret management via HashiCorp Vault\n\n")

	return sb.String()
}

// generateSREContent generates SRE content
func (s *SimpleAgent) generateSREContent(input *AgentInput) string {
	var sb strings.Builder

	sb.WriteString("## Observability & Resilience\n\n")

	sb.WriteString("### Service Level Objectives (SLO)\n")
	sb.WriteString("- **Availability**: 99.99% uptime (52min downtime/year)\n")
	sb.WriteString("- **Latency**: p95 < 150ms, p99 < 300ms\n")
	sb.WriteString("- **Error Rate**: < 0.1% of requests\n\n")

	sb.WriteString("### Observability Stack\n")
	sb.WriteString("- **Tracing**: OpenTelemetry → Jaeger\n")
	sb.WriteString("- **Metrics**: Prometheus + Grafana\n")
	sb.WriteString("- **Logging**: Structured JSON logs → Loki\n")
	sb.WriteString("- **Alerting**: PagerDuty for critical issues\n\n")

	sb.WriteString("### Resilience Patterns\n")
	sb.WriteString("- **Circuit Breaker**: Resilience4j on all external calls\n")
	sb.WriteString("- **Retry**: Exponential backoff (max 3 attempts)\n")
	sb.WriteString("- **Timeout**: 30s per request\n")
	sb.WriteString("- **Bulkhead**: Thread pool isolation\n")
	sb.WriteString("- **Rate Limiting**: 1000 RPS per client\n\n")

	sb.WriteString("### Capacity Planning\n")
	sb.WriteString("- **Horizontal Scaling**: Auto-scale 2-20 pods based on CPU/memory\n")
	sb.WriteString("- **Database**: Read replicas for scaling reads\n")
	sb.WriteString("- **Cache**: Redis cluster with 100GB capacity\n\n")

	return sb.String()
}

// generateQAContent generates QA content
func (s *SimpleAgent) generateQAContent(input *AgentInput) string {
	var sb strings.Builder

	sb.WriteString("## Requirements & Testing\n\n")

	sb.WriteString("### Functional Requirements\n")
	for i, svc := range input.Docs.Services {
		if i >= 3 {
			break // Limit to first 3 services
		}
		sb.WriteString(fmt.Sprintf("**%s**:\n", svc.Name))
		for j, method := range svc.Methods {
			if j >= 3 {
				break
			}
			sb.WriteString(fmt.Sprintf("- FR-%d.%d: %s\n", i+1, j+1, method.Description))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("### Test Strategy\n")
	sb.WriteString("**Unit Tests**:\n")
	sb.WriteString("- Coverage target: 80%+\n")
	sb.WriteString("- Framework: JUnit 5 / pytest\n")
	sb.WriteString("- Mock external dependencies\n\n")

	sb.WriteString("**Integration Tests**:\n")
	sb.WriteString("- Test service-to-service contracts\n")
	sb.WriteString("- Use Testcontainers for dependencies\n")
	sb.WriteString("- Contract testing with Pact\n\n")

	sb.WriteString("**E2E Tests**:\n")
	sb.WriteString("- Critical user journeys\n")
	sb.WriteString("- Run in staging environment\n")
	sb.WriteString("- Playwright for UI, Postman for API\n\n")

	sb.WriteString("### Acceptance Criteria\n")
	sb.WriteString("- All critical paths have automated tests\n")
	sb.WriteString("- Performance tests meet SLO targets\n")
	sb.WriteString("- Security scans pass (SAST + DAST)\n\n")

	return sb.String()
}

// generateSLO generates Service Level Objectives
func (s *SimpleAgent) generateSLO(input *AgentInput) map[string]SLO {
	slo := make(map[string]SLO)

	slo["availability"] = SLO{
		Name:        "Availability",
		Description: "Service uptime",
		Target:      0.9999,
		Metric:      "availability",
		Threshold:   "99.99%",
	}

	slo["latency_p95"] = SLO{
		Name:        "Latency (p95)",
		Description: "95th percentile response time",
		Target:      0.15,
		Metric:      "latency_p95",
		Threshold:   "< 150ms",
	}

	slo["error_rate"] = SLO{
		Name:        "Error Rate",
		Description: "Percentage of failed requests",
		Target:      0.001,
		Metric:      "error_rate",
		Threshold:   "< 0.1%",
	}

	return slo
}

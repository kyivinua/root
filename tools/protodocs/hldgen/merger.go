package hldgen

import (
	"fmt"
	"strings"
	"time"
)

// ResponseMerger merges agent responses into a unified HLD output
type ResponseMerger struct {
	strategy string
}

// NewResponseMerger creates a new response merger
func NewResponseMerger(strategy string) *ResponseMerger {
	if strategy == "" {
		strategy = "priority_merge"
	}
	return &ResponseMerger{strategy: strategy}
}

// Merge merges multiple agent responses into a single HLD output
func (m *ResponseMerger) Merge(responses map[string]*AgentResponse) *HLDOutput {
	output := &HLDOutput{
		Metadata: Metadata{
			Version:        "7.0",
			GenerationMode: "multi_agent",
			GeneratedAt:    time.Now(),
		},
		Diagrams: make(map[string]string),
		SLO:      make(map[string]SLO),
		Warnings: make([]string, 0),
	}

	// Priority order for merging
	priorityOrder := []AgentRole{
		RoleArchitect,
		RoleSecurity,
		RoleSRE,
		RolePM,
		RoleQA,
	}

	// Merge content by priority
	for _, role := range priorityOrder {
		if resp, exists := responses[string(role)]; exists {
			m.mergeResponse(output, role, resp)
		}
	}

	// Generate final markdown
	output.Markdown = m.generateMarkdown(output)

	return output
}

// mergeResponse merges a single agent response
func (m *ResponseMerger) mergeResponse(output *HLDOutput, role AgentRole, resp *AgentResponse) {
	switch role {
	case RoleArchitect:
		output.Architecture = resp.Content
		// Merge diagrams
		for key, diagram := range resp.Diagrams {
			output.Diagrams[key] = diagram
		}

	case RoleSecurity:
		output.Security = resp.Content
		// Check for critical security issues
		if m.hasCriticalSecurityIssues(resp) {
			output.Warnings = append(output.Warnings, "CRITICAL SECURITY ISSUES DETECTED")
		}

	case RoleSRE:
		output.Observability = resp.Content
		// Merge SLO
		for key, slo := range resp.SLO {
			output.SLO[key] = slo
		}

	case RolePM:
		output.BusinessContext = resp.Content

	case RoleQA:
		output.Requirements = resp.Content
	}
}

// hasCriticalSecurityIssues checks for critical security issues
func (m *ResponseMerger) hasCriticalSecurityIssues(resp *AgentResponse) bool {
	// In production, this would check structured data
	content := strings.ToLower(resp.Content)
	criticalKeywords := []string{"no mtls", "no auth", "security gap"}
	for _, keyword := range criticalKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	return false
}

// generateMarkdown generates the final markdown document
func (m *ResponseMerger) generateMarkdown(output *HLDOutput) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("# High-Level Design: %s\n\n", output.Metadata.ModuleName))
	sb.WriteString(fmt.Sprintf("**Generated**: %s | **Version**: %s | **Mode**: %s\n\n",
		output.Metadata.GeneratedAt.Format(time.RFC3339),
		output.Metadata.Version,
		output.Metadata.GenerationMode))

	if output.ConsensusScore > 0 {
		sb.WriteString(fmt.Sprintf("**Quality Score**: %.2f | **Rounds**: %d\n\n",
			output.ConsensusScore, output.RoundsCompleted))
	}

	// Warnings
	if len(output.Warnings) > 0 {
		sb.WriteString("!!! warning \"Warnings\"\n")
		for _, warning := range output.Warnings {
			sb.WriteString(fmt.Sprintf("    - %s\n", warning))
		}
		sb.WriteString("\n")
	}

	// Table of Contents
	sb.WriteString("## Table of Contents\n\n")
	sb.WriteString("1. [Business Context](#business-context)\n")
	sb.WriteString("2. [Architecture](#architecture)\n")
	sb.WriteString("3. [Security](#security)\n")
	sb.WriteString("4. [Observability](#observability)\n")
	sb.WriteString("5. [Requirements](#requirements)\n")
	if len(output.SLO) > 0 {
		sb.WriteString("6. [Service Level Objectives](#service-level-objectives)\n")
	}
	sb.WriteString("\n---\n\n")

	// Business Context
	if output.BusinessContext != "" {
		sb.WriteString(output.BusinessContext)
		sb.WriteString("\n---\n\n")
	}

	// Architecture
	if output.Architecture != "" {
		sb.WriteString(output.Architecture)
		sb.WriteString("\n---\n\n")
	}

	// Security
	if output.Security != "" {
		sb.WriteString(output.Security)
		sb.WriteString("\n---\n\n")
	}

	// Observability
	if output.Observability != "" {
		sb.WriteString(output.Observability)
		sb.WriteString("\n---\n\n")
	}

	// Requirements
	if output.Requirements != "" {
		sb.WriteString(output.Requirements)
		sb.WriteString("\n---\n\n")
	}

	// SLO
	if len(output.SLO) > 0 {
		sb.WriteString("## Service Level Objectives\n\n")
		sb.WriteString("| Name | Metric | Target | Threshold |\n")
		sb.WriteString("|------|--------|--------|----------|\n")
		for _, slo := range output.SLO {
			sb.WriteString(fmt.Sprintf("| %s | %s | %.2f%% | %s |\n",
				slo.Name, slo.Metric, slo.Target*100, slo.Threshold))
		}
		sb.WriteString("\n")
	}

	// Refinement Log
	if output.RefinementLog != "" {
		sb.WriteString("## Refinement Log\n\n")
		sb.WriteString("<details>\n")
		sb.WriteString("<summary>Click to expand refinement history</summary>\n\n")
		sb.WriteString(output.RefinementLog)
		sb.WriteString("</details>\n\n")
	}

	// Footer
	sb.WriteString("---\n\n")
	sb.WriteString("*Generated by ProtoDocs HLD Generator v7.0*\n")

	return sb.String()
}

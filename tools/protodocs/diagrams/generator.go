package diagrams

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DiagramGenerator orchestrates diagram generation
type DiagramGenerator struct {
	config *DiagramConfig
}

// NewDiagramGenerator creates a new diagram generator
func NewDiagramGenerator(config *DiagramConfig) *DiagramGenerator {
	if config == nil {
		config = DefaultDiagramConfig()
	}
	return &DiagramGenerator{
		config: config,
	}
}

// GenerateAll generates all enabled diagrams
func (g *DiagramGenerator) GenerateAll(model *ApiDocModel) ([]GenerationResult, error) {
	// Ensure output directory exists
	if err := os.MkdirAll(g.config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}

	var results []GenerationResult

	// Generate static diagrams (don't require model)
	if g.config.EnablePipeline {
		result := g.generatePipelineDiagram()
		results = append(results, result)
	}

	if g.config.EnableEnricher {
		result := g.generateEnricherDiagram()
		results = append(results, result)
	}

	if g.config.EnableComponent {
		result := g.generateComponentDiagram()
		results = append(results, result)
	}

	if g.config.EnableTransform {
		result := g.generateTransformDiagram()
		results = append(results, result)
	}

	if g.config.EnableDeploy {
		result := g.generateDeployDiagram()
		results = append(results, result)
	}

	// Generate dynamic diagrams (require model)
	if model != nil {
		if g.config.EnableDataModel {
			result := g.generateDataModelDiagram(model)
			results = append(results, result)
		}

		if g.config.EnableServiceMap {
			result := g.generateServiceMapDiagram(model)
			results = append(results, result)
		}

		if g.config.EnableMessageHierarchy {
			result := g.generateMessageHierarchyDiagram(model)
			results = append(results, result)
		}
	}

	// Write all diagrams to files
	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("Warning: Failed to generate %s diagram: %v\n", result.Metadata.Type, result.Error)
			continue
		}

		filename := filepath.Join(g.config.OutputDir, result.Metadata.Filename)
		content := g.wrapDiagram(result)

		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			result.Error = fmt.Errorf("write diagram file: %w", err)
			fmt.Printf("Warning: Failed to write %s: %v\n", filename, err)
		} else {
			fmt.Printf("Generated diagram: %s\n", filename)
		}
	}

	// Generate index if enabled
	if g.config.GenerateIndex {
		if err := g.generateIndex(results); err != nil {
			fmt.Printf("Warning: Failed to generate index: %v\n", err)
		}
	}

	return results, nil
}

// wrapDiagram wraps Mermaid diagram with markdown and metadata
func (g *DiagramGenerator) wrapDiagram(result GenerationResult) string {
	var sb strings.Builder

	// Title and description
	sb.WriteString(fmt.Sprintf("# %s\n\n", result.Metadata.Title))
	if result.Metadata.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", result.Metadata.Description))
	}

	// Timestamp if enabled
	if g.config.IncludeTimestamp {
		sb.WriteString(fmt.Sprintf("*Generated: %s*\n\n", result.Metadata.GeneratedAt.Format(time.RFC3339)))
	}

	// Mermaid diagram
	sb.WriteString("```mermaid\n")

	// Apply theme if specified
	if g.config.Theme != "" && g.config.Theme != "default" {
		sb.WriteString(fmt.Sprintf("%%{init: {'theme':'%s'}}%%\n", g.config.Theme))
	}

	sb.WriteString(result.Content)
	sb.WriteString("\n```\n")

	return sb.String()
}

// generateIndex creates an index page with links to all diagrams
func (g *DiagramGenerator) generateIndex(results []GenerationResult) error {
	var sb strings.Builder

	sb.WriteString("# ProtoDocs Architecture Diagrams\n\n")
	sb.WriteString("Auto-generated Mermaid diagrams for the ProtoDocs system.\n\n")

	if g.config.IncludeTimestamp {
		sb.WriteString(fmt.Sprintf("*Generated: %s*\n\n", time.Now().Format(time.RFC3339)))
	}

	sb.WriteString("## Table of Contents\n\n")

	// Group diagrams by category
	staticDiagrams := []GenerationResult{}
	dynamicDiagrams := []GenerationResult{}

	for _, result := range results {
		if result.Error != nil {
			continue
		}

		switch result.Metadata.Type {
		case DiagramTypePipeline, DiagramTypeEnricher, DiagramTypeComponent, DiagramTypeTransform, DiagramTypeDeploy:
			staticDiagrams = append(staticDiagrams, result)
		default:
			dynamicDiagrams = append(dynamicDiagrams, result)
		}
	}

	// Static diagrams section
	if len(staticDiagrams) > 0 {
		sb.WriteString("### Architecture Diagrams\n\n")
		for _, result := range staticDiagrams {
			sb.WriteString(fmt.Sprintf("- [%s](%s) - %s\n",
				result.Metadata.Title,
				result.Metadata.Filename,
				result.Metadata.Description))
		}
		sb.WriteString("\n")
	}

	// Dynamic diagrams section
	if len(dynamicDiagrams) > 0 {
		sb.WriteString("### Model-Based Diagrams\n\n")
		for _, result := range dynamicDiagrams {
			sb.WriteString(fmt.Sprintf("- [%s](%s) - %s\n",
				result.Metadata.Title,
				result.Metadata.Filename,
				result.Metadata.Description))
		}
		sb.WriteString("\n")
	}

	// Add quick reference
	sb.WriteString("## Diagram Types\n\n")
	sb.WriteString("| Diagram | Type | Purpose |\n")
	sb.WriteString("|---------|------|----------|\n")

	for _, result := range results {
		if result.Error != nil {
			continue
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
			result.Metadata.Title,
			result.Metadata.MermaidType,
			result.Metadata.Description))
	}

	indexPath := filepath.Join(g.config.OutputDir, "README.md")
	return os.WriteFile(indexPath, []byte(sb.String()), 0644)
}

// sanitizeName converts names to valid Mermaid IDs
func sanitizeName(name string) string {
	// Replace dots, slashes, and special chars with underscores
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, " ", "_")
	// Remove angle brackets (for generics)
	name = strings.ReplaceAll(name, "<", "")
	name = strings.ReplaceAll(name, ">", "")
	return name
}

// escapeString escapes special characters for Mermaid
func escapeString(s string) string {
	s = strings.ReplaceAll(s, "\"", "'")
	s = strings.ReplaceAll(s, "\n", " ")
	// Truncate long descriptions
	if len(s) > 100 {
		s = s[:97] + "..."
	}
	return s
}

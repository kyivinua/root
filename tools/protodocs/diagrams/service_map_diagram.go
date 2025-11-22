package diagrams

import (
	"fmt"
	"strings"
	"time"
)

// generateServiceMapDiagram creates a diagram showing service and message relationships
func (g *DiagramGenerator) generateServiceMapDiagram(model *ApiDocModel) GenerationResult {
	metadata := DiagramMetadata{
		Type:        DiagramTypeServiceMap,
		Title:       "Service Relationship Map",
		Description: fmt.Sprintf("Services and their message dependencies across all modules"),
		Filename:    "service-map.md",
		GeneratedAt: time.Now(),
		MermaidType: "graph TB",
	}

	var sb strings.Builder
	sb.WriteString("graph TB\n")

	// Define styles
	sb.WriteString("    classDef serviceNode fill:#e1f5ff,stroke:#01579b,stroke-width:3px\n")
	sb.WriteString("    classDef messageNode fill:#e8f5e9,stroke:#1b5e20,stroke-width:2px\n")
	sb.WriteString("    classDef enumNode fill:#fff3e0,stroke:#e65100,stroke-width:2px\n")
	sb.WriteString("    classDef methodNode fill:#f3e5f5,stroke:#4a148c,stroke-width:1px\n")
	sb.WriteString("\n")

	serviceCount := 0
	messageCount := 0

	// Build message and enum lookup for reference checking
	allMessages := make(map[string]bool)
	allEnums := make(map[string]bool)

	for _, module := range model.Modules {
		for _, msg := range module.Messages {
			allMessages[msg.FullName] = true
		}
		for _, enum := range module.Enums {
			allEnums[enum.FullName] = true
		}
	}

	// Process each module
	for _, module := range model.Modules {
		if len(module.Services) == 0 {
			continue // Skip modules without services
		}

		// Check if we should limit the number of services
		if g.config.MaxServicesPerDiagram > 0 && serviceCount >= g.config.MaxServicesPerDiagram {
			sb.WriteString(fmt.Sprintf("    MoreServices[\"... and %d more services\"]:::serviceNode\n",
				countTotalServices(model)-serviceCount))
			break
		}

		modulePrefix := sanitizeName(module.Package)
		sb.WriteString(fmt.Sprintf("    subgraph %s [%s]\n", modulePrefix, module.Package))

		// Services and methods
		for _, service := range module.Services {
			if g.config.MaxServicesPerDiagram > 0 && serviceCount >= g.config.MaxServicesPerDiagram {
				break
			}

			// Skip private types if configured
			if !g.config.IncludePrivateTypes && service.Visibility == "INTERNAL" {
				continue
			}

			serviceName := sanitizeName(service.FullName)
			sb.WriteString(fmt.Sprintf("        %s[🔧 %s]:::serviceNode\n", serviceName, service.Name))
			serviceCount++

			// Process methods and their message references
			for _, method := range service.Methods {
				methodName := sanitizeName(method.FullName)

				// Skip private methods if configured
				if !g.config.IncludePrivateTypes && method.Visibility == "INTERNAL" {
					continue
				}

				// Create method node
				streamInfo := ""
				if method.ClientStreaming && method.ServerStreaming {
					streamInfo = "↔️ "
				} else if method.ClientStreaming {
					streamInfo = "↑ "
				} else if method.ServerStreaming {
					streamInfo = "↓ "
				}

				sb.WriteString(fmt.Sprintf("        %s[%s%s]:::methodNode\n",
					methodName, streamInfo, method.Name))

				// Connect service to method
				sb.WriteString(fmt.Sprintf("        %s --> %s\n", serviceName, methodName))
			}
		}

		sb.WriteString("    end\n")
		sb.WriteString("\n")
	}

	// Now create message and enum nodes, and connect methods to them
	messageNodes := make(map[string]bool)
	enumNodes := make(map[string]bool)

	for _, module := range model.Modules {
		for _, service := range module.Services {
			if !g.config.IncludePrivateTypes && service.Visibility == "INTERNAL" {
				continue
			}

			for _, method := range service.Methods {
				if !g.config.IncludePrivateTypes && method.Visibility == "INTERNAL" {
					continue
				}

				methodName := sanitizeName(method.FullName)

				// Check if we should limit messages
				if g.config.MaxMessagesPerDiagram > 0 && messageCount >= g.config.MaxMessagesPerDiagram {
					continue
				}

				// Input type
				if allMessages[method.InputType] {
					inputName := sanitizeName(method.InputType)
					if !messageNodes[inputName] {
						shortName := getShortName(method.InputType)
						sb.WriteString(fmt.Sprintf("    %s[📦 %s]:::messageNode\n", inputName, shortName))
						messageNodes[inputName] = true
						messageCount++
					}
					sb.WriteString(fmt.Sprintf("    %s -.->|input| %s\n", methodName, inputName))
				}

				// Output type
				if allMessages[method.OutputType] {
					outputName := sanitizeName(method.OutputType)
					if !messageNodes[outputName] {
						shortName := getShortName(method.OutputType)
						sb.WriteString(fmt.Sprintf("    %s[📦 %s]:::messageNode\n", outputName, shortName))
						messageNodes[outputName] = true
						messageCount++
					}
					sb.WriteString(fmt.Sprintf("    %s -.->|output| %s\n", methodName, outputName))
				}
			}
		}
	}

	// Add selected enums that are referenced by visible messages
	if len(enumNodes) > 0 {
		sb.WriteString("\n")
		for _, module := range model.Modules {
			for _, enum := range module.Enums {
				if !g.config.IncludePrivateTypes && enum.Visibility == "INTERNAL" {
					continue
				}

				enumName := sanitizeName(enum.FullName)
				if enumNodes[enumName] {
					shortName := getShortName(enum.FullName)
					sb.WriteString(fmt.Sprintf("    %s{{🔢 %s}}:::enumNode\n", enumName, shortName))
				}
			}
		}
	}

	// Update description with actual counts
	metadata.Description = fmt.Sprintf("Service relationships map showing %d services and %d messages",
		serviceCount, messageCount)

	return GenerationResult{
		Metadata: metadata,
		Content:  sb.String(),
	}
}

// getShortName extracts the short name from a fully qualified name
func getShortName(fullName string) string {
	parts := strings.Split(fullName, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullName
}

// countTotalServices counts all services across all modules
func countTotalServices(model *ApiDocModel) int {
	count := 0
	for _, module := range model.Modules {
		count += len(module.Services)
	}
	return count
}

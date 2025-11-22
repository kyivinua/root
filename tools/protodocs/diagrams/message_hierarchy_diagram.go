package diagrams

import (
	"fmt"
	"strings"
	"time"
)

// generateMessageHierarchyDiagram creates a diagram showing message dependencies
func (g *DiagramGenerator) generateMessageHierarchyDiagram(model *ApiDocModel) GenerationResult {
	metadata := DiagramMetadata{
		Type:        DiagramTypeMessageHierarchy,
		Title:       "Message Hierarchy and Dependencies",
		Description: "Message type dependencies showing which messages reference other messages and enums",
		Filename:    "message-hierarchy.md",
		GeneratedAt: time.Now(),
		MermaidType: "graph TD",
	}

	var sb strings.Builder
	sb.WriteString("graph TD\n")

	// Define styles
	sb.WriteString("    classDef messageNode fill:#e1f5ff,stroke:#01579b,stroke-width:2px\n")
	sb.WriteString("    classDef enumNode fill:#fff3e0,stroke:#e65100,stroke-width:2px\n")
	sb.WriteString("    classDef scalarNode fill:#e8f5e9,stroke:#1b5e20,stroke-width:1px\n")
	sb.WriteString("\n")

	// Build message and enum lookup
	allMessages := make(map[string]*DocMessage)
	allEnums := make(map[string]*DocEnum)

	for _, module := range model.Modules {
		for i := range module.Messages {
			msg := &module.Messages[i]
			allMessages[msg.FullName] = msg
		}
		for i := range module.Enums {
			enum := &module.Enums[i]
			allEnums[enum.FullName] = enum
		}
	}

	messageCount := 0
	processedMessages := make(map[string]bool)
	processedEnums := make(map[string]bool)

	// Process each module
	for _, module := range model.Modules {
		if len(module.Messages) == 0 {
			continue
		}

		modulePrefix := sanitizeName(module.Package)
		sb.WriteString(fmt.Sprintf("    subgraph %s [📦 %s]\n", modulePrefix, module.Package))

		for _, message := range module.Messages {
			// Check message limit
			if g.config.MaxMessagesPerDiagram > 0 && messageCount >= g.config.MaxMessagesPerDiagram {
				sb.WriteString(fmt.Sprintf("        MoreMessages[\"... and %d more messages\"]:::messageNode\n",
					countTotalMessages(model)-messageCount))
				break
			}

			// Skip private types if configured
			if !g.config.IncludePrivateTypes && message.Visibility == "INTERNAL" {
				continue
			}

			messageName := sanitizeName(message.FullName)
			fieldCount := len(message.Fields)
			oneofCount := countOneofGroups(message.Fields)

			label := message.Name
			if oneofCount > 0 {
				label = fmt.Sprintf("%s<br/>%d fields, %d oneofs", message.Name, fieldCount, oneofCount)
			} else if fieldCount > 0 {
				label = fmt.Sprintf("%s<br/>%d fields", message.Name, fieldCount)
			}

			sb.WriteString(fmt.Sprintf("        %s[📄 %s]:::messageNode\n", messageName, label))
			processedMessages[messageName] = true
			messageCount++
		}

		sb.WriteString("    end\n")
		sb.WriteString("\n")
	}

	// Now process field references to create edges
	for _, module := range model.Modules {
		for _, message := range module.Messages {
			if !g.config.IncludePrivateTypes && message.Visibility == "INTERNAL" {
				continue
			}

			messageName := sanitizeName(message.FullName)
			if !processedMessages[messageName] {
				continue // Skip if message wasn't added due to limits
			}

			for _, field := range message.Fields {
				// Check if field references another message
				if field.TypeName != "" && allMessages[field.TypeName] != nil {
					refMessageName := sanitizeName(field.TypeName)

					// Add referenced message if not already added
					if !processedMessages[refMessageName] {
						refMsg := allMessages[field.TypeName]
						if g.config.IncludePrivateTypes || refMsg.Visibility != "INTERNAL" {
							shortName := getShortName(field.TypeName)
							sb.WriteString(fmt.Sprintf("    %s[📄 %s]:::messageNode\n", refMessageName, shortName))
							processedMessages[refMessageName] = true
						}
					}

					// Create edge
					edgeLabel := field.Name
					if field.Label == "repeated" {
						edgeLabel = fmt.Sprintf("%s[]", field.Name)
					}

					sb.WriteString(fmt.Sprintf("    %s -->|%s| %s\n",
						messageName, edgeLabel, refMessageName))
				}

				// Check if field references an enum
				if field.TypeName != "" && allEnums[field.TypeName] != nil {
					refEnumName := sanitizeName(field.TypeName)

					// Add referenced enum if not already added
					if !processedEnums[refEnumName] {
						refEnum := allEnums[field.TypeName]
						if g.config.IncludePrivateTypes || refEnum.Visibility != "INTERNAL" {
							shortName := getShortName(field.TypeName)
							valueCount := len(refEnum.Values)
							sb.WriteString(fmt.Sprintf("    %s{{🔢 %s<br/>%d values}}:::enumNode\n",
								refEnumName, shortName, valueCount))
							processedEnums[refEnumName] = true
						}
					}

					// Create edge
					sb.WriteString(fmt.Sprintf("    %s -.->|%s| %s\n",
						messageName, field.Name, refEnumName))
				}
			}
		}
	}

	// Update description with actual count
	metadata.Description = fmt.Sprintf("Message dependencies showing %d messages and their type references",
		messageCount)

	return GenerationResult{
		Metadata: metadata,
		Content:  sb.String(),
	}
}

// countOneofGroups counts unique oneof groups in fields
func countOneofGroups(fields []DocField) int {
	groups := make(map[string]bool)
	for _, field := range fields {
		if field.OneofGroup != "" {
			groups[field.OneofGroup] = true
		}
	}
	return len(groups)
}

// countTotalMessages counts all messages across all modules
func countTotalMessages(model *ApiDocModel) int {
	count := 0
	for _, module := range model.Modules {
		count += len(module.Messages)
	}
	return count
}

// Package diagrams provides diagram generation functionality.
package diagrams

import (
	"fmt"
	"strings"

	"github.com/kyivinua/docgen-tool/internal/docgen"
)

// Generator handles diagram generation.
type Generator struct {
	config Config
}

// Config represents diagram generator configuration.
type Config struct {
	ServiceGraph  bool
	MessageGraph  bool
	SequenceGraph bool
	OverviewGraph bool
}

// NewGenerator creates a new diagram generator.
func NewGenerator(config Config) *Generator {
	return &Generator{config: config}
}

// GenerateServiceDiagram generates a service architecture diagram.
func (g *Generator) GenerateServiceDiagram(service *docgen.Service) (string, error) {
	if !g.config.ServiceGraph {
		return "", nil
	}

	var sb strings.Builder
	sb.WriteString("```mermaid\n")
	sb.WriteString("graph TD\n")

	// Service node
	serviceID := sanitizeID(service.Name)
	sb.WriteString(fmt.Sprintf("    %s[%s]\n", serviceID, service.Name))

	// Add methods
	for i, method := range service.Methods {
		methodID := fmt.Sprintf("%s_m%d", serviceID, i)
		streamType := ""
		if method.ClientStreaming && method.ServerStreaming {
			streamType = " (BiDi Stream)"
		} else if method.ClientStreaming {
			streamType = " (Client Stream)"
		} else if method.ServerStreaming {
			streamType = " (Server Stream)"
		}
		sb.WriteString(fmt.Sprintf("    %s[%s%s]\n", methodID, method.Name, streamType))
		sb.WriteString(fmt.Sprintf("    %s --> %s\n", serviceID, methodID))

		// Add input/output
		inputID := fmt.Sprintf("%s_in", methodID)
		outputID := fmt.Sprintf("%s_out", methodID)
		sb.WriteString(fmt.Sprintf("    %s[%s]\n", inputID, method.InputType))
		sb.WriteString(fmt.Sprintf("    %s[%s]\n", outputID, method.OutputType))
		sb.WriteString(fmt.Sprintf("    %s -.-> %s\n", inputID, methodID))
		sb.WriteString(fmt.Sprintf("    %s -.-> %s\n", methodID, outputID))
	}

	sb.WriteString("```\n")
	return sb.String(), nil
}

// GenerateMessageDiagram generates a message structure diagram.
func (g *Generator) GenerateMessageDiagram(message *docgen.Message) (string, error) {
	if !g.config.MessageGraph {
		return "", nil
	}

	var sb strings.Builder
	sb.WriteString("```mermaid\n")
	sb.WriteString("classDiagram\n")

	// Message class
	messageID := sanitizeID(message.Name)
	sb.WriteString(fmt.Sprintf("    class %s {\n", messageID))

	// Add fields
	for _, field := range message.Fields {
		fieldType := field.Type
		if field.Label == "repeated" {
			fieldType = "[]" + fieldType
		}
		sb.WriteString(fmt.Sprintf("        +%s %s\n", fieldType, field.Name))
	}

	sb.WriteString("    }\n")
	sb.WriteString("```\n")
	return sb.String(), nil
}

// GenerateSequenceDiagram generates a sequence diagram for a method.
func (g *Generator) GenerateSequenceDiagram(service *docgen.Service, method *docgen.Method) (string, error) {
	if !g.config.SequenceGraph {
		return "", nil
	}

	var sb strings.Builder
	sb.WriteString("```mermaid\n")
	sb.WriteString("sequenceDiagram\n")
	sb.WriteString("    participant Client\n")
	sb.WriteString(fmt.Sprintf("    participant %s\n", service.Name))

	if method.ClientStreaming && method.ServerStreaming {
		sb.WriteString(fmt.Sprintf("    Client->>+%s: Stream %s\n", service.Name, method.InputType))
		sb.WriteString("    loop Bidirectional Streaming\n")
		sb.WriteString(fmt.Sprintf("        %s-->>Client: Stream %s\n", service.Name, method.OutputType))
		sb.WriteString(fmt.Sprintf("        Client->>%s: Stream %s\n", service.Name, method.InputType))
		sb.WriteString("    end\n")
		sb.WriteString(fmt.Sprintf("    %s-->>-Client: Complete\n", service.Name))
	} else if method.ClientStreaming {
		sb.WriteString("    loop Client Streaming\n")
		sb.WriteString(fmt.Sprintf("        Client->>%s: Stream %s\n", service.Name, method.InputType))
		sb.WriteString("    end\n")
		sb.WriteString(fmt.Sprintf("    %s-->>Client: %s\n", service.Name, method.OutputType))
	} else if method.ServerStreaming {
		sb.WriteString(fmt.Sprintf("    Client->>%s: %s\n", service.Name, method.InputType))
		sb.WriteString("    loop Server Streaming\n")
		sb.WriteString(fmt.Sprintf("        %s-->>Client: Stream %s\n", service.Name, method.OutputType))
		sb.WriteString("    end\n")
	} else {
		sb.WriteString(fmt.Sprintf("    Client->>+%s: %s\n", service.Name, method.InputType))
		sb.WriteString(fmt.Sprintf("    %s-->>-Client: %s\n", service.Name, method.OutputType))
	}

	sb.WriteString("```\n")
	return sb.String(), nil
}

// GenerateOverviewDiagram generates an overview diagram of all services.
func (g *Generator) GenerateOverviewDiagram(services []docgen.Service) (string, error) {
	if !g.config.OverviewGraph {
		return "", nil
	}

	var sb strings.Builder
	sb.WriteString("```mermaid\n")
	sb.WriteString("graph LR\n")

	for i, service := range services {
		serviceID := fmt.Sprintf("s%d", i)
		sb.WriteString(fmt.Sprintf("    %s[%s]\n", serviceID, service.Name))

		// Group methods
		for j, method := range service.Methods {
			methodID := fmt.Sprintf("%s_m%d", serviceID, j)
			sb.WriteString(fmt.Sprintf("    %s[%s]\n", methodID, method.Name))
			sb.WriteString(fmt.Sprintf("    %s --> %s\n", serviceID, methodID))
		}
	}

	sb.WriteString("```\n")
	return sb.String(), nil
}

// GenerateAllDiagrams generates all enabled diagrams for a service.
func (g *Generator) GenerateAllDiagrams(service *docgen.Service) (map[string]string, error) {
	diagrams := make(map[string]string)

	// Service diagram
	if g.config.ServiceGraph {
		diagram, err := g.GenerateServiceDiagram(service)
		if err != nil {
			return nil, fmt.Errorf("failed to generate service diagram: %w", err)
		}
		if diagram != "" {
			diagrams["service_architecture"] = diagram
		}
	}

	// Message diagrams
	if g.config.MessageGraph {
		for _, message := range service.Messages {
			diagram, err := g.GenerateMessageDiagram(&message)
			if err != nil {
				return nil, fmt.Errorf("failed to generate message diagram for %s: %w", message.Name, err)
			}
			if diagram != "" {
				diagrams[fmt.Sprintf("message_%s", strings.ToLower(message.Name))] = diagram
			}
		}
	}

	// Sequence diagrams
	if g.config.SequenceGraph {
		for _, method := range service.Methods {
			diagram, err := g.GenerateSequenceDiagram(service, &method)
			if err != nil {
				return nil, fmt.Errorf("failed to generate sequence diagram for %s: %w", method.Name, err)
			}
			if diagram != "" {
				diagrams[fmt.Sprintf("sequence_%s", strings.ToLower(method.Name))] = diagram
			}
		}
	}

	return diagrams, nil
}

// sanitizeID sanitizes an identifier for use in Mermaid diagrams.
func sanitizeID(id string) string {
	// Replace special characters with underscores
	id = strings.ReplaceAll(id, ".", "_")
	id = strings.ReplaceAll(id, "-", "_")
	id = strings.ReplaceAll(id, " ", "_")
	id = strings.ReplaceAll(id, "/", "_")
	return id
}

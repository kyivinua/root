package docgen

import (
	"fmt"
	"strings"
	"time"
)

// ConsolidatedDocGenerator creates comprehensive, single-file documentation for each service
type ConsolidatedDocGenerator struct {
	config ConsolidatedConfig
}

// ConsolidatedConfig holds configuration for consolidated documentation
type ConsolidatedConfig struct {
	// Documentation structure
	IncludeTOC            bool
	TOCDepth              int
	IncludeDiagrams       bool
	DiagramPosition       string // "inline", "appendix", "section"
	IncludeCrossReferences bool
	IncludeAnchors        bool

	// Diagram types to include
	IncludeArchitecture   bool
	IncludeSequence       bool
	IncludeMessageGraph   bool
	IncludeDataFlow       bool

	// Content sections
	IncludeOverview       bool
	IncludeAuthentication bool
	IncludeExamples       bool
	IncludeErrorCodes     bool
	IncludeChangelog      bool

	// Formatting
	UseEmojis             bool
	CodeHighlighting      string // "protobuf", "json", "yaml"
	DiagramTheme          string // "default", "forest", "dark", "neutral"
}

// DefaultConsolidatedConfig returns default configuration
func DefaultConsolidatedConfig() ConsolidatedConfig {
	return ConsolidatedConfig{
		IncludeTOC:            true,
		TOCDepth:              3,
		IncludeDiagrams:       true,
		DiagramPosition:       "section",
		IncludeCrossReferences: true,
		IncludeAnchors:        true,
		IncludeArchitecture:   true,
		IncludeSequence:       true,
		IncludeMessageGraph:   true,
		IncludeDataFlow:       true,
		IncludeOverview:       true,
		IncludeAuthentication: true,
		IncludeExamples:       true,
		IncludeErrorCodes:     true,
		IncludeChangelog:      false,
		UseEmojis:             true,
		CodeHighlighting:      "protobuf",
		DiagramTheme:          "default",
	}
}

// NewConsolidatedDocGenerator creates a new consolidated documentation generator
func NewConsolidatedDocGenerator(config ConsolidatedConfig) *ConsolidatedDocGenerator {
	return &ConsolidatedDocGenerator{
		config: config,
	}
}

// ServiceDocumentation represents complete documentation for a service
type ServiceDocumentation struct {
	Service      *ServiceDoc
	Methods      []MethodDoc
	Messages     []MessageDoc
	Enums        []EnumDoc
	Diagrams     map[string]string // diagram type -> mermaid content
	Metadata     DocumentMetadata
}

// ServiceDoc contains service information
type ServiceDoc struct {
	Name        string
	FullName    string
	Package     string
	Description string
	Version     string
	ProtoFile   string
}

// MethodDoc contains method information
type MethodDoc struct {
	Name            string
	FullName        string
	Description     string
	InputType       string
	OutputType      string
	ClientStreaming bool
	ServerStreaming bool
	HTTPBindings    []HTTPBinding
	Examples        []Example
}

// MessageDoc contains message information
type MessageDoc struct {
	Name        string
	FullName    string
	Description string
	Fields      []FieldDoc
	NestedTypes []string
}

// FieldDoc contains field information
type FieldDoc struct {
	Name        string
	Number      int32
	Type        string
	TypeName    string // For message/enum references
	Label       string // optional, required, repeated
	Description string
	OneofGroup  string
	DefaultValue string
}

// EnumDoc contains enum information
type EnumDoc struct {
	Name        string
	FullName    string
	Description string
	Values      []EnumValueDoc
}

// EnumValueDoc contains enum value information
type EnumValueDoc struct {
	Name        string
	Number      int32
	Description string
}

// HTTPBinding contains HTTP method binding
type HTTPBinding struct {
	Method string // GET, POST, PUT, DELETE, PATCH
	Path   string
	Body   string
}

// Example contains code example
type Example struct {
	Title       string
	Description string
	Language    string
	Code        string
}

// DocumentMetadata contains document metadata
type DocumentMetadata struct {
	Title       string
	Author      string
	Version     string
	Generated   time.Time
	LastUpdated time.Time
	Tags        []string
}

// GenerateConsolidatedDoc generates a complete, consolidated documentation file
func (g *ConsolidatedDocGenerator) GenerateConsolidatedDoc(doc *ServiceDocumentation) string {
	var sb strings.Builder

	// Document header
	g.writeHeader(&sb, doc)

	// Table of Contents
	if g.config.IncludeTOC {
		g.writeTOC(&sb, doc)
	}

	// Overview section
	if g.config.IncludeOverview {
		g.writeOverview(&sb, doc)
	}

	// Architecture diagram
	if g.config.IncludeDiagrams && g.config.IncludeArchitecture {
		g.writeArchitectureDiagram(&sb, doc)
	}

	// Methods section
	g.writeMethods(&sb, doc)

	// Messages section
	g.writeMessages(&sb, doc)

	// Enums section
	if len(doc.Enums) > 0 {
		g.writeEnums(&sb, doc)
	}

	// Diagrams appendix
	if g.config.IncludeDiagrams && g.config.DiagramPosition == "appendix" {
		g.writeDiagramsAppendix(&sb, doc)
	}

	// Error codes
	if g.config.IncludeErrorCodes {
		g.writeErrorCodes(&sb, doc)
	}

	// Examples
	if g.config.IncludeExamples {
		g.writeExamplesSection(&sb, doc)
	}

	// Footer
	g.writeFooter(&sb, doc)

	return sb.String()
}

// writeHeader writes the document header
func (g *ConsolidatedDocGenerator) writeHeader(sb *strings.Builder, doc *ServiceDocumentation) {
	emoji := ""
	if g.config.UseEmojis {
		emoji = "📚 "
	}

	sb.WriteString(fmt.Sprintf("# %s%s API Documentation\n\n", emoji, doc.Service.Name))

	// Metadata table
	sb.WriteString("| **Attribute** | **Value** |\n")
	sb.WriteString("|---------------|----------|\n")
	sb.WriteString(fmt.Sprintf("| **Service Name** | `%s` |\n", doc.Service.Name))
	sb.WriteString(fmt.Sprintf("| **Package** | `%s` |\n", doc.Service.Package))
	sb.WriteString(fmt.Sprintf("| **Version** | %s |\n", doc.Service.Version))
	sb.WriteString(fmt.Sprintf("| **Proto File** | `%s` |\n", doc.Service.ProtoFile))
	sb.WriteString(fmt.Sprintf("| **Generated** | %s |\n", doc.Metadata.Generated.Format(time.RFC3339)))
	sb.WriteString("\n")

	if doc.Service.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", doc.Service.Description))
	}

	sb.WriteString("---\n\n")
}

// writeTOC writes the table of contents
func (g *ConsolidatedDocGenerator) writeTOC(sb *strings.Builder, doc *ServiceDocumentation) {
	emoji := ""
	if g.config.UseEmojis {
		emoji = "📑 "
	}

	sb.WriteString(fmt.Sprintf("## %sTable of Contents\n\n", emoji))

	// Overview
	if g.config.IncludeOverview {
		sb.WriteString("- [Overview](#overview)\n")
	}

	// Architecture
	if g.config.IncludeDiagrams && g.config.IncludeArchitecture {
		sb.WriteString("- [Architecture](#architecture)\n")
	}

	// Methods
	sb.WriteString("- [Methods](#methods)\n")
	for _, method := range doc.Methods {
		anchor := strings.ToLower(strings.ReplaceAll(method.Name, "_", "-"))
		sb.WriteString(fmt.Sprintf("  - [%s](#%s)\n", method.Name, anchor))
	}

	// Messages
	sb.WriteString("- [Messages](#messages)\n")
	for _, msg := range doc.Messages {
		anchor := strings.ToLower(strings.ReplaceAll(msg.Name, "_", "-"))
		sb.WriteString(fmt.Sprintf("  - [%s](#%s)\n", msg.Name, anchor))
	}

	// Enums
	if len(doc.Enums) > 0 {
		sb.WriteString("- [Enumerations](#enumerations)\n")
	}

	// Error codes
	if g.config.IncludeErrorCodes {
		sb.WriteString("- [Error Codes](#error-codes)\n")
	}

	// Examples
	if g.config.IncludeExamples {
		sb.WriteString("- [Examples](#examples)\n")
	}

	// Diagrams appendix
	if g.config.IncludeDiagrams && g.config.DiagramPosition == "appendix" {
		sb.WriteString("- [Diagrams](#diagrams)\n")
	}

	sb.WriteString("\n---\n\n")
}

// writeOverview writes the overview section
func (g *ConsolidatedDocGenerator) writeOverview(sb *strings.Builder, doc *ServiceDocumentation) {
	emoji := ""
	if g.config.UseEmojis {
		emoji = "📖 "
	}

	sb.WriteString(fmt.Sprintf("## %sOverview\n\n", emoji))

	if g.config.IncludeAnchors {
		sb.WriteString("<a name=\"overview\"></a>\n\n")
	}

	// Service statistics
	sb.WriteString("### Service Statistics\n\n")
	sb.WriteString("| Metric | Count |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| **RPC Methods** | %d |\n", len(doc.Methods)))
	sb.WriteString(fmt.Sprintf("| **Message Types** | %d |\n", len(doc.Messages)))
	sb.WriteString(fmt.Sprintf("| **Enumerations** | %d |\n", len(doc.Enums)))

	// Count streaming methods
	streamingCount := 0
	for _, m := range doc.Methods {
		if m.ClientStreaming || m.ServerStreaming {
			streamingCount++
		}
	}
	sb.WriteString(fmt.Sprintf("| **Streaming RPCs** | %d |\n", streamingCount))
	sb.WriteString("\n")

	// Quick start
	sb.WriteString("### Quick Start\n\n")
	sb.WriteString("This service provides the following capabilities:\n\n")

	for i, method := range doc.Methods {
		if i >= 5 {
			sb.WriteString(fmt.Sprintf("- ... and %d more methods\n", len(doc.Methods)-5))
			break
		}

		streamInfo := ""
		if method.ClientStreaming && method.ServerStreaming {
			streamInfo = " (bidirectional streaming)"
		} else if method.ClientStreaming {
			streamInfo = " (client streaming)"
		} else if method.ServerStreaming {
			streamInfo = " (server streaming)"
		}

		anchor := strings.ToLower(strings.ReplaceAll(method.Name, "_", "-"))
		sb.WriteString(fmt.Sprintf("- [`%s`](#%s)%s: %s\n",
			method.Name, anchor, streamInfo, method.Description))
	}

	sb.WriteString("\n---\n\n")
}

// writeArchitectureDiagram writes the architecture diagram
func (g *ConsolidatedDocGenerator) writeArchitectureDiagram(sb *strings.Builder, doc *ServiceDocumentation) {
	emoji := ""
	if g.config.UseEmojis {
		emoji = "🏗️ "
	}

	sb.WriteString(fmt.Sprintf("## %sArchitecture\n\n", emoji))

	if g.config.IncludeAnchors {
		sb.WriteString("<a name=\"architecture\"></a>\n\n")
	}

	if archDiagram, ok := doc.Diagrams["architecture"]; ok {
		sb.WriteString("### Service Architecture Diagram\n\n")
		sb.WriteString("```mermaid\n")
		if g.config.DiagramTheme != "default" {
			sb.WriteString(fmt.Sprintf("%%{init: {'theme':'%s'}}%%\n", g.config.DiagramTheme))
		}
		sb.WriteString(archDiagram)
		sb.WriteString("\n```\n\n")
	} else {
		// Generate default architecture diagram
		g.generateDefaultArchitectureDiagram(sb, doc)
	}

	sb.WriteString("---\n\n")
}

// generateDefaultArchitectureDiagram generates a default architecture diagram
func (g *ConsolidatedDocGenerator) generateDefaultArchitectureDiagram(sb *strings.Builder, doc *ServiceDocumentation) {
	sb.WriteString("```mermaid\n")
	if g.config.DiagramTheme != "default" {
		sb.WriteString(fmt.Sprintf("%%{init: {'theme':'%s'}}%%\n", g.config.DiagramTheme))
	}
	sb.WriteString("graph TB\n")
	sb.WriteString("    classDef serviceClass fill:#e1f5ff,stroke:#01579b,stroke-width:3px\n")
	sb.WriteString("    classDef methodClass fill:#f3e5f5,stroke:#4a148c,stroke-width:2px\n")
	sb.WriteString("    classDef messageClass fill:#e8f5e9,stroke:#1b5e20,stroke-width:2px\n\n")

	serviceName := sanitizeName(doc.Service.Name)
	sb.WriteString(fmt.Sprintf("    %s[🔧 %s]:::serviceClass\n\n", serviceName, doc.Service.Name))

	for _, method := range doc.Methods {
		methodName := sanitizeName(method.Name)

		streamIcon := ""
		if method.ClientStreaming && method.ServerStreaming {
			streamIcon = "↔️ "
		} else if method.ClientStreaming {
			streamIcon = "↑ "
		} else if method.ServerStreaming {
			streamIcon = "↓ "
		}

		sb.WriteString(fmt.Sprintf("    %s[%s%s]:::methodClass\n", methodName, streamIcon, method.Name))
		sb.WriteString(fmt.Sprintf("    %s --> %s\n", serviceName, methodName))

		// Input/Output messages
		sb.WriteString(fmt.Sprintf("    %s_in[📥 %s]:::messageClass\n", methodName, getShortName(method.InputType)))
		sb.WriteString(fmt.Sprintf("    %s_out[📤 %s]:::messageClass\n", methodName, getShortName(method.OutputType)))
		sb.WriteString(fmt.Sprintf("    %s_in -.->|input| %s\n", methodName, methodName))
		sb.WriteString(fmt.Sprintf("    %s -.->|output| %s_out\n", methodName, methodName))
	}

	sb.WriteString("```\n\n")
}

// Helper functions
func sanitizeName(name string) string {
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "<", "")
	name = strings.ReplaceAll(name, ">", "")
	return name
}

func getShortName(fullName string) string {
	parts := strings.Split(fullName, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullName
}

package docgen

import (
	"fmt"
	"strings"
)

// writeMethods writes the methods section with detailed information
func (g *ConsolidatedDocGenerator) writeMethods(sb *strings.Builder, doc *ServiceDocumentation) {
	emoji := ""
	if g.config.UseEmojis {
		emoji = "⚙️ "
	}

	sb.WriteString(fmt.Sprintf("## %sMethods\n\n", emoji))

	if g.config.IncludeAnchors {
		sb.WriteString("<a name=\"methods\"></a>\n\n")
	}

	sb.WriteString(fmt.Sprintf("This service defines **%d RPC methods**:\n\n", len(doc.Methods)))

	for _, method := range doc.Methods {
		g.writeMethod(sb, method, doc)
	}
}

// writeMethod writes detailed information for a single method
func (g *ConsolidatedDocGenerator) writeMethod(sb *strings.Builder, method MethodDoc, doc *ServiceDocumentation) {
	anchor := strings.ToLower(strings.ReplaceAll(method.Name, "_", "-"))

	// Method header
	sb.WriteString(fmt.Sprintf("### %s\n\n", method.Name))

	if g.config.IncludeAnchors {
		sb.WriteString(fmt.Sprintf("<a name=\"%s\"></a>\n\n", anchor))
	}

	// Description
	if method.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", method.Description))
	}

	// Method signature
	sb.WriteString("#### Method Signature\n\n")
	sb.WriteString("```protobuf\n")

	// Streaming annotations
	if method.ClientStreaming && method.ServerStreaming {
		sb.WriteString("// Bidirectional streaming RPC\n")
		sb.WriteString(fmt.Sprintf("rpc %s(stream %s) returns (stream %s);\n",
			method.Name, getShortName(method.InputType), getShortName(method.OutputType)))
	} else if method.ClientStreaming {
		sb.WriteString("// Client streaming RPC\n")
		sb.WriteString(fmt.Sprintf("rpc %s(stream %s) returns (%s);\n",
			method.Name, getShortName(method.InputType), getShortName(method.OutputType)))
	} else if method.ServerStreaming {
		sb.WriteString("// Server streaming RPC\n")
		sb.WriteString(fmt.Sprintf("rpc %s(%s) returns (stream %s);\n",
			method.Name, getShortName(method.InputType), getShortName(method.OutputType)))
	} else {
		sb.WriteString("// Unary RPC\n")
		sb.WriteString(fmt.Sprintf("rpc %s(%s) returns (%s);\n",
			method.Name, getShortName(method.InputType), getShortName(method.OutputType)))
	}

	sb.WriteString("```\n\n")

	// Method details table
	sb.WriteString("#### Method Details\n\n")
	sb.WriteString("| Attribute | Value |\n")
	sb.WriteString("|-----------|-------|\n")
	sb.WriteString(fmt.Sprintf("| **Full Name** | `%s` |\n", method.FullName))

	inputAnchor := strings.ToLower(strings.ReplaceAll(getShortName(method.InputType), "_", "-"))
	outputAnchor := strings.ToLower(strings.ReplaceAll(getShortName(method.OutputType), "_", "-"))

	if g.config.IncludeCrossReferences {
		sb.WriteString(fmt.Sprintf("| **Input Type** | [`%s`](#%s) |\n",
			getShortName(method.InputType), inputAnchor))
		sb.WriteString(fmt.Sprintf("| **Output Type** | [`%s`](#%s) |\n",
			getShortName(method.OutputType), outputAnchor))
	} else {
		sb.WriteString(fmt.Sprintf("| **Input Type** | `%s` |\n", getShortName(method.InputType)))
		sb.WriteString(fmt.Sprintf("| **Output Type** | `%s` |\n", getShortName(method.OutputType)))
	}

	// Streaming type
	streamingType := "Unary"
	if method.ClientStreaming && method.ServerStreaming {
		streamingType = "Bidirectional Streaming"
	} else if method.ClientStreaming {
		streamingType = "Client Streaming"
	} else if method.ServerStreaming {
		streamingType = "Server Streaming"
	}
	sb.WriteString(fmt.Sprintf("| **Streaming Type** | %s |\n", streamingType))

	sb.WriteString("\n")

	// HTTP bindings if available
	if len(method.HTTPBindings) > 0 {
		sb.WriteString("#### HTTP/REST Bindings\n\n")
		sb.WriteString("| Method | Path | Body |\n")
		sb.WriteString("|--------|------|------|\n")
		for _, binding := range method.HTTPBindings {
			body := binding.Body
			if body == "" {
				body = "-"
			}
			sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` |\n", binding.Method, binding.Path, body))
		}
		sb.WriteString("\n")
	}

	// Sequence diagram for this method
	if g.config.IncludeDiagrams && g.config.IncludeSequence {
		g.writeMethodSequenceDiagram(sb, method)
	}

	// Examples for this method
	if len(method.Examples) > 0 {
		sb.WriteString("#### Usage Examples\n\n")
		for _, example := range method.Examples {
			if example.Title != "" {
				sb.WriteString(fmt.Sprintf("**%s**\n\n", example.Title))
			}
			if example.Description != "" {
				sb.WriteString(fmt.Sprintf("%s\n\n", example.Description))
			}
			sb.WriteString(fmt.Sprintf("```%s\n%s\n```\n\n", example.Language, example.Code))
		}
	}

	sb.WriteString("---\n\n")
}

// writeMethodSequenceDiagram writes a sequence diagram for a method
func (g *ConsolidatedDocGenerator) writeMethodSequenceDiagram(sb *strings.Builder, method MethodDoc) {
	sb.WriteString("##### Sequence Diagram\n\n")
	sb.WriteString("```mermaid\n")

	if g.config.DiagramTheme != "default" {
		sb.WriteString(fmt.Sprintf("%%{init: {'theme':'%s'}}%%\n", g.config.DiagramTheme))
	}

	sb.WriteString("sequenceDiagram\n")
	sb.WriteString("    participant Client\n")
	sb.WriteString("    participant Service\n\n")

	inputType := getShortName(method.InputType)
	outputType := getShortName(method.OutputType)

	if method.ClientStreaming && method.ServerStreaming {
		// Bidirectional streaming
		sb.WriteString("    Note over Client,Service: Bidirectional Streaming\n")
		sb.WriteString(fmt.Sprintf("    Client->>+Service: %s (stream)\n", method.Name))
		sb.WriteString("    loop Stream Messages\n")
		sb.WriteString(fmt.Sprintf("        Client->>Service: %s\n", inputType))
		sb.WriteString(fmt.Sprintf("        Service-->>Client: %s\n", outputType))
		sb.WriteString("    end\n")
		sb.WriteString("    Service-->>-Client: End Stream\n")
	} else if method.ClientStreaming {
		// Client streaming
		sb.WriteString("    Note over Client,Service: Client Streaming\n")
		sb.WriteString(fmt.Sprintf("    Client->>+Service: %s (stream)\n", method.Name))
		sb.WriteString("    loop Stream Messages\n")
		sb.WriteString(fmt.Sprintf("        Client->>Service: %s\n", inputType))
		sb.WriteString("    end\n")
		sb.WriteString(fmt.Sprintf("    Service-->>-Client: %s\n", outputType))
	} else if method.ServerStreaming {
		// Server streaming
		sb.WriteString("    Note over Client,Service: Server Streaming\n")
		sb.WriteString(fmt.Sprintf("    Client->>+Service: %s\n", method.Name))
		sb.WriteString(fmt.Sprintf("    Client->>Service: %s\n", inputType))
		sb.WriteString("    loop Stream Messages\n")
		sb.WriteString(fmt.Sprintf("        Service-->>Client: %s\n", outputType))
		sb.WriteString("    end\n")
		sb.WriteString("    Service-->>-Client: End Stream\n")
	} else {
		// Unary
		sb.WriteString(fmt.Sprintf("    Client->>+Service: %s\n", method.Name))
		sb.WriteString(fmt.Sprintf("    Note right of Service: %s\n", inputType))
		sb.WriteString("    Service-->>-Client: Response\n")
		sb.WriteString(fmt.Sprintf("    Note left of Client: %s\n", outputType))
	}

	sb.WriteString("```\n\n")
}

// writeMessages writes the messages section
func (g *ConsolidatedDocGenerator) writeMessages(sb *strings.Builder, doc *ServiceDocumentation) {
	emoji := ""
	if g.config.UseEmojis {
		emoji = "📦 "
	}

	sb.WriteString(fmt.Sprintf("## %sMessages\n\n", emoji))

	if g.config.IncludeAnchors {
		sb.WriteString("<a name=\"messages\"></a>\n\n")
	}

	sb.WriteString(fmt.Sprintf("This service defines **%d message types**:\n\n", len(doc.Messages)))

	for _, message := range doc.Messages {
		g.writeMessage(sb, message, doc)
	}
}

// writeMessage writes detailed information for a single message
func (g *ConsolidatedDocGenerator) writeMessage(sb *strings.Builder, message MessageDoc, doc *ServiceDocumentation) {
	anchor := strings.ToLower(strings.ReplaceAll(message.Name, "_", "-"))

	// Message header
	sb.WriteString(fmt.Sprintf("### %s\n\n", message.Name))

	if g.config.IncludeAnchors {
		sb.WriteString(fmt.Sprintf("<a name=\"%s\"></a>\n\n", anchor))
	}

	// Description
	if message.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", message.Description))
	}

	// Message details
	sb.WriteString("| Attribute | Value |\n")
	sb.WriteString("|-----------|-------|\n")
	sb.WriteString(fmt.Sprintf("| **Full Name** | `%s` |\n", message.FullName))
	sb.WriteString(fmt.Sprintf("| **Field Count** | %d |\n", len(message.Fields)))
	if len(message.NestedTypes) > 0 {
		sb.WriteString(fmt.Sprintf("| **Nested Types** | %d |\n", len(message.NestedTypes)))
	}
	sb.WriteString("\n")

	// Fields table
	if len(message.Fields) > 0 {
		sb.WriteString("#### Fields\n\n")
		sb.WriteString("| # | Name | Type | Label | Description |\n")
		sb.WriteString("|---|------|------|-------|-------------|\n")

		for _, field := range message.Fields {
			fieldType := field.Type
			if field.TypeName != "" {
				if g.config.IncludeCrossReferences {
					typeAnchor := strings.ToLower(strings.ReplaceAll(getShortName(field.TypeName), "_", "-"))
					fieldType = fmt.Sprintf("[`%s`](#%s)", getShortName(field.TypeName), typeAnchor)
				} else {
					fieldType = fmt.Sprintf("`%s`", getShortName(field.TypeName))
				}
			}

			label := field.Label
			if field.OneofGroup != "" {
				label = fmt.Sprintf("oneof `%s`", field.OneofGroup)
			}

			description := field.Description
			if description == "" {
				description = "-"
			}

			sb.WriteString(fmt.Sprintf("| %d | `%s` | %s | %s | %s |\n",
				field.Number, field.Name, fieldType, label, description))
		}
		sb.WriteString("\n")
	}

	// Proto definition
	sb.WriteString("#### Proto Definition\n\n")
	sb.WriteString(fmt.Sprintf("```%s\n", g.config.CodeHighlighting))
	sb.WriteString(fmt.Sprintf("message %s {\n", message.Name))

	// Group fields by oneof
	oneofFields := make(map[string][]FieldDoc)
	regularFields := []FieldDoc{}

	for _, field := range message.Fields {
		if field.OneofGroup != "" {
			oneofFields[field.OneofGroup] = append(oneofFields[field.OneofGroup], field)
		} else {
			regularFields = append(regularFields, field)
		}
	}

	// Write regular fields
	for _, field := range regularFields {
		fieldType := field.Type
		if field.TypeName != "" {
			fieldType = getShortName(field.TypeName)
		}

		if field.Description != "" {
			sb.WriteString(fmt.Sprintf("  // %s\n", field.Description))
		}

		sb.WriteString(fmt.Sprintf("  %s %s %s = %d;\n",
			field.Label, fieldType, field.Name, field.Number))
	}

	// Write oneof groups
	for oneofName, fields := range oneofFields {
		sb.WriteString(fmt.Sprintf("\n  oneof %s {\n", oneofName))
		for _, field := range fields {
			fieldType := field.Type
			if field.TypeName != "" {
				fieldType = getShortName(field.TypeName)
			}

			if field.Description != "" {
				sb.WriteString(fmt.Sprintf("    // %s\n", field.Description))
			}

			sb.WriteString(fmt.Sprintf("    %s %s = %d;\n",
				fieldType, field.Name, field.Number))
		}
		sb.WriteString("  }\n")
	}

	sb.WriteString("}\n```\n\n")

	// Message graph diagram
	if g.config.IncludeDiagrams && g.config.IncludeMessageGraph {
		g.writeMessageDiagram(sb, message)
	}

	sb.WriteString("---\n\n")
}

// writeMessageDiagram writes a diagram for message structure
func (g *ConsolidatedDocGenerator) writeMessageDiagram(sb *strings.Builder, message MessageDoc) {
	sb.WriteString("##### Message Structure\n\n")
	sb.WriteString("```mermaid\n")

	if g.config.DiagramTheme != "default" {
		sb.WriteString(fmt.Sprintf("%%{init: {'theme':'%s'}}%%\n", g.config.DiagramTheme))
	}

	sb.WriteString("classDiagram\n")
	sb.WriteString(fmt.Sprintf("    class %s {\n", message.Name))

	for _, field := range message.Fields {
		fieldType := field.Type
		if field.TypeName != "" {
			fieldType = getShortName(field.TypeName)
		}

		label := ""
		if field.Label == "repeated" {
			label = "[]"
		}

		sb.WriteString(fmt.Sprintf("        +%s%s %s\n", fieldType, label, field.Name))
	}

	sb.WriteString("    }\n")

	// Add relationships for message type fields
	for _, field := range message.Fields {
		if field.TypeName != "" {
			shortType := getShortName(field.TypeName)
			rel := "-->"
			if field.Label == "repeated" {
				rel = "\"1\" --> \"*\""
			}
			sb.WriteString(fmt.Sprintf("    %s %s %s\n", message.Name, rel, shortType))
		}
	}

	sb.WriteString("```\n\n")
}

// writeEnums writes the enumerations section
func (g *ConsolidatedDocGenerator) writeEnums(sb *strings.Builder, doc *ServiceDocumentation) {
	emoji := ""
	if g.config.UseEmojis {
		emoji = "🔢 "
	}

	sb.WriteString(fmt.Sprintf("## %sEnumerations\n\n", emoji))

	if g.config.IncludeAnchors {
		sb.WriteString("<a name=\"enumerations\"></a>\n\n")
	}

	sb.WriteString(fmt.Sprintf("This service defines **%d enumeration types**:\n\n", len(doc.Enums)))

	for _, enum := range doc.Enums {
		g.writeEnum(sb, enum)
	}
}

// writeEnum writes detailed information for a single enum
func (g *ConsolidatedDocGenerator) writeEnum(sb *strings.Builder, enum EnumDoc) {
	anchor := strings.ToLower(strings.ReplaceAll(enum.Name, "_", "-"))

	sb.WriteString(fmt.Sprintf("### %s\n\n", enum.Name))

	if g.config.IncludeAnchors {
		sb.WriteString(fmt.Sprintf("<a name=\"%s\"></a>\n\n", anchor))
	}

	if enum.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", enum.Description))
	}

	sb.WriteString("| Value | Number | Description |\n")
	sb.WriteString("|-------|--------|-------------|\n")

	for _, value := range enum.Values {
		description := value.Description
		if description == "" {
			description = "-"
		}
		sb.WriteString(fmt.Sprintf("| `%s` | %d | %s |\n", value.Name, value.Number, description))
	}

	sb.WriteString("\n")

	// Proto definition
	sb.WriteString("#### Proto Definition\n\n")
	sb.WriteString(fmt.Sprintf("```%s\n", g.config.CodeHighlighting))
	sb.WriteString(fmt.Sprintf("enum %s {\n", enum.Name))

	for _, value := range enum.Values {
		if value.Description != "" {
			sb.WriteString(fmt.Sprintf("  // %s\n", value.Description))
		}
		sb.WriteString(fmt.Sprintf("  %s = %d;\n", value.Name, value.Number))
	}

	sb.WriteString("}\n```\n\n")

	sb.WriteString("---\n\n")
}

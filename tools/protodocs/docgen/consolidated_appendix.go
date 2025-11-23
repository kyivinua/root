package docgen

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// writeErrorCodes writes the error codes section
func (g *ConsolidatedDocGenerator) writeErrorCodes(sb *strings.Builder, doc *ServiceDocumentation) {
	emoji := ""
	if g.config.UseEmojis {
		emoji = "⚠️ "
	}

	sb.WriteString(fmt.Sprintf("## %sError Codes\n\n", emoji))

	if g.config.IncludeAnchors {
		sb.WriteString("<a name=\"error-codes\"></a>\n\n")
	}

	sb.WriteString("This service uses standard gRPC status codes:\n\n")

	errorCodes := []struct {
		Code        string
		HTTPStatus  int
		Description string
	}{
		{"OK", 200, "Success"},
		{"CANCELLED", 499, "Operation cancelled by client"},
		{"UNKNOWN", 500, "Unknown error"},
		{"INVALID_ARGUMENT", 400, "Client specified an invalid argument"},
		{"DEADLINE_EXCEEDED", 504, "Deadline expired before operation could complete"},
		{"NOT_FOUND", 404, "Requested entity not found"},
		{"ALREADY_EXISTS", 409, "Entity already exists"},
		{"PERMISSION_DENIED", 403, "Caller does not have permission"},
		{"RESOURCE_EXHAUSTED", 429, "Resource has been exhausted"},
		{"FAILED_PRECONDITION", 400, "Operation rejected because system is not in required state"},
		{"ABORTED", 409, "Operation aborted, typically due to concurrency issue"},
		{"OUT_OF_RANGE", 400, "Operation attempted past valid range"},
		{"UNIMPLEMENTED", 501, "Operation not implemented"},
		{"INTERNAL", 500, "Internal server error"},
		{"UNAVAILABLE", 503, "Service unavailable"},
		{"DATA_LOSS", 500, "Unrecoverable data loss or corruption"},
		{"UNAUTHENTICATED", 401, "Request does not have valid authentication credentials"},
	}

	sb.WriteString("| gRPC Code | HTTP Status | Description |\n")
	sb.WriteString("|-----------|-------------|-------------|\n")

	for _, ec := range errorCodes {
		sb.WriteString(fmt.Sprintf("| `%s` | %d | %s |\n", ec.Code, ec.HTTPStatus, ec.Description))
	}

	sb.WriteString("\n")

	// Error handling best practices
	sb.WriteString("### Error Handling Best Practices\n\n")
	sb.WriteString("1. **Check status codes**: Always check the gRPC status code before processing responses\n")
	sb.WriteString("2. **Implement retries**: Use exponential backoff for transient errors (`UNAVAILABLE`, `RESOURCE_EXHAUSTED`)\n")
	sb.WriteString("3. **Log errors**: Log error details with correlation IDs for troubleshooting\n")
	sb.WriteString("4. **Handle streaming errors**: Properly handle errors in streaming RPCs\n")
	sb.WriteString("5. **Validate inputs**: Validate inputs client-side to avoid `INVALID_ARGUMENT` errors\n\n")

	sb.WriteString("---\n\n")
}

// writeExamplesSection writes the examples section
func (g *ConsolidatedDocGenerator) writeExamplesSection(sb *strings.Builder, doc *ServiceDocumentation) {
	emoji := ""
	if g.config.UseEmojis {
		emoji = "💡 "
	}

	sb.WriteString(fmt.Sprintf("## %sExamples\n\n", emoji))

	if g.config.IncludeAnchors {
		sb.WriteString("<a name=\"examples\"></a>\n\n")
	}

	// Generate basic examples for common languages
	if g.config.IncludeGoExamples {
		g.writeGoExample(sb, doc)
	}
	if g.config.IncludePythonExamples {
		g.writePythonExample(sb, doc)
	}
	if g.config.IncludeTypeScriptExamples {
		g.writeTypeScriptExample(sb, doc)
	}

	sb.WriteString("---\n\n")
}

// writeGoExample writes a Go example
func (g *ConsolidatedDocGenerator) writeGoExample(sb *strings.Builder, doc *ServiceDocumentation) {
	sb.WriteString("### Go Example\n\n")

	sb.WriteString("```go\n")
	sb.WriteString("package main\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("    \"context\"\n")
	sb.WriteString("    \"log\"\n")
	sb.WriteString("    \"time\"\n\n")
	sb.WriteString("    \"google.golang.org/grpc\"\n")
	sb.WriteString("    \"google.golang.org/grpc/credentials/insecure\"\n\n")
	sb.WriteString(fmt.Sprintf("    pb \"%s\"\n", doc.Service.Package))
	sb.WriteString(")\n\n")

	sb.WriteString("func main() {\n")
	sb.WriteString("    // Connect to the service\n")
	sb.WriteString("    conn, err := grpc.Dial(\"localhost:50051\",\n")
	sb.WriteString("        grpc.WithTransportCredentials(insecure.NewCredentials()))\n")
	sb.WriteString("    if err != nil {\n")
	sb.WriteString("        log.Fatalf(\"Failed to connect: %v\", err)\n")
	sb.WriteString("    }\n")
	sb.WriteString("    defer conn.Close()\n\n")

	sb.WriteString(fmt.Sprintf("    client := pb.New%sClient(conn)\n\n", doc.Service.Name))

	// Example call for first method
	if len(doc.Methods) > 0 {
		method := doc.Methods[0]
		sb.WriteString("    // Example RPC call\n")
		sb.WriteString("    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)\n")
		sb.WriteString("    defer cancel()\n\n")

		inputType := getShortName(method.InputType)

		if !method.ClientStreaming && !method.ServerStreaming {
			// Unary call
			sb.WriteString(fmt.Sprintf("    req := &pb.%s{\n", inputType))
			sb.WriteString("        // Fill in request fields\n")
			sb.WriteString("    }\n\n")

			sb.WriteString(fmt.Sprintf("    resp, err := client.%s(ctx, req)\n", method.Name))
			sb.WriteString("    if err != nil {\n")
			sb.WriteString("        log.Fatalf(\"RPC failed: %v\", err)\n")
			sb.WriteString("    }\n\n")

			sb.WriteString("    log.Printf(\"Response: %v\", resp)\n")
		}
	}

	sb.WriteString("}\n")
	sb.WriteString("```\n\n")
}

// writePythonExample writes a Python example
func (g *ConsolidatedDocGenerator) writePythonExample(sb *strings.Builder, doc *ServiceDocumentation) {
	sb.WriteString("### Python Example\n\n")

	sb.WriteString("```python\n")
	sb.WriteString("import grpc\n")
	sb.WriteString(fmt.Sprintf("import %s_pb2\n", strings.ToLower(doc.Service.Name)))
	sb.WriteString(fmt.Sprintf("import %s_pb2_grpc\n\n", strings.ToLower(doc.Service.Name)))

	sb.WriteString("def main():\n")
	sb.WriteString("    # Connect to the service\n")
	sb.WriteString("    with grpc.insecure_channel('localhost:50051') as channel:\n")
	sb.WriteString(fmt.Sprintf("        stub = %s_pb2_grpc.%sStub(channel)\n\n",
		strings.ToLower(doc.Service.Name), doc.Service.Name))

	if len(doc.Methods) > 0 {
		method := doc.Methods[0]
		inputType := getShortName(method.InputType)

		sb.WriteString("        # Example RPC call\n")
		sb.WriteString(fmt.Sprintf("        request = %s_pb2.%s(\n",
			strings.ToLower(doc.Service.Name), inputType))
		sb.WriteString("            # Fill in request fields\n")
		sb.WriteString("        )\n\n")

		sb.WriteString("        try:\n")
		sb.WriteString(fmt.Sprintf("            response = stub.%s(request)\n", method.Name))
		sb.WriteString("            print(f'Response: {response}')\n")
		sb.WriteString("        except grpc.RpcError as e:\n")
		sb.WriteString("            print(f'RPC failed: {e.code()} - {e.details()}')\n\n")
	}

	sb.WriteString("if __name__ == '__main__':\n")
	sb.WriteString("    main()\n")
	sb.WriteString("```\n\n")
}

// writeTypeScriptExample writes a TypeScript example
func (g *ConsolidatedDocGenerator) writeTypeScriptExample(sb *strings.Builder, doc *ServiceDocumentation) {
	sb.WriteString("### TypeScript Example\n\n")

	sb.WriteString("```typescript\n")
	sb.WriteString("import * as grpc from '@grpc/grpc-js';\n")
	sb.WriteString("import * as protoLoader from '@grpc/proto-loader';\n")
	sb.WriteString(fmt.Sprintf("import { ProtoGrpcType } from './%s';\n", strings.TrimSuffix(doc.Service.ProtoFile, ".proto")))
	sb.WriteString(fmt.Sprintf("import { %sClient } from './%s/%s';\n\n", doc.Service.Name, doc.Service.Package, doc.Service.Name))

	sb.WriteString("// Load proto file\n")
	sb.WriteString("const packageDefinition = protoLoader.loadSync(\n")
	sb.WriteString(fmt.Sprintf("    '%s',\n", doc.Service.ProtoFile))
	sb.WriteString("    {\n")
	sb.WriteString("        keepCase: true,\n")
	sb.WriteString("        longs: String,\n")
	sb.WriteString("        enums: String,\n")
	sb.WriteString("        defaults: true,\n")
	sb.WriteString("        oneofs: true\n")
	sb.WriteString("    }\n")
	sb.WriteString(");\n\n")

	sb.WriteString("const proto = grpc.loadPackageDefinition(\n")
	sb.WriteString("    packageDefinition\n")
	sb.WriteString(") as unknown as ProtoGrpcType;\n\n")

	sb.WriteString("// Create client\n")
	packagePath := strings.ReplaceAll(doc.Service.Package, ".", ".")
	sb.WriteString(fmt.Sprintf("const client: %sClient = new proto.%s.%s(\n", doc.Service.Name, packagePath, doc.Service.Name))
	sb.WriteString("    'localhost:50051',\n")
	sb.WriteString("    grpc.credentials.createInsecure()\n")
	sb.WriteString(");\n\n")

	if len(doc.Methods) > 0 {
		method := doc.Methods[0]

		sb.WriteString("// Example RPC call\n")
		sb.WriteString("const request = {\n")
		sb.WriteString("    // Fill in request fields\n")
		sb.WriteString("};\n\n")

		sb.WriteString(fmt.Sprintf("client.%s(request, (error: grpc.ServiceError | null, response?: any) => {\n", method.Name))
		sb.WriteString("    if (error) {\n")
		sb.WriteString("        console.error('RPC failed:', error);\n")
		sb.WriteString("        return;\n")
		sb.WriteString("    }\n")
		sb.WriteString("    console.log('Response:', response);\n")
		sb.WriteString("});\n")
	}

	sb.WriteString("```\n\n")
}

// writeDiagramsAppendix writes all diagrams in an appendix
func (g *ConsolidatedDocGenerator) writeDiagramsAppendix(sb *strings.Builder, doc *ServiceDocumentation) {
	emoji := ""
	if g.config.UseEmojis {
		emoji = "📊 "
	}

	sb.WriteString(fmt.Sprintf("## %sDiagrams\n\n", emoji))

	if g.config.IncludeAnchors {
		sb.WriteString("<a name=\"diagrams\"></a>\n\n")
	}

	sb.WriteString("This section contains all diagrams for the service.\n\n")

	// List all available diagrams
	for diagramType, content := range doc.Diagrams {
		g.writeDiagramSection(sb, diagramType, content)
	}

	sb.WriteString("---\n\n")
}

// writeDiagramSection writes a single diagram section
func (g *ConsolidatedDocGenerator) writeDiagramSection(sb *strings.Builder, diagramType, content string) {
	title := formatDiagramTitle(diagramType)

	sb.WriteString(fmt.Sprintf("### %s\n\n", title))

	sb.WriteString("```mermaid\n")
	if g.config.DiagramTheme != "default" {
		sb.WriteString(fmt.Sprintf("%%{init: {'theme':'%s'}}%%\n", g.config.DiagramTheme))
	}
	sb.WriteString(content)
	sb.WriteString("\n```\n\n")
}

// formatDiagramTitle formats diagram type into a readable title
func formatDiagramTitle(diagramType string) string {
	switch diagramType {
	case "architecture":
		return "Service Architecture"
	case "sequence":
		return "Sequence Diagrams"
	case "message_graph":
		return "Message Relationships"
	case "data_flow":
		return "Data Flow"
	default:
		return cases.Title(language.English).String(strings.ReplaceAll(diagramType, "_", " "))
	}
}

// writeFooter writes the document footer
func (g *ConsolidatedDocGenerator) writeFooter(sb *strings.Builder, doc *ServiceDocumentation) {
	sb.WriteString("---\n\n")

	// Generation metadata
	sb.WriteString("<div align=\"center\">\n\n")

	sb.WriteString("**Generated Documentation**\n\n")

	sb.WriteString("| Attribute | Value |\n")
	sb.WriteString("|-----------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Generated At | %s |\n", doc.Metadata.Generated.Format("2006-01-02 15:04:05 MST")))
	sb.WriteString(fmt.Sprintf("| Generator Version | %s |\n", doc.Metadata.Version))

	if doc.Metadata.Author != "" {
		sb.WriteString(fmt.Sprintf("| Author | %s |\n", doc.Metadata.Author))
	}

	sb.WriteString("\n")

	// Footer message
	if g.config.UseEmojis {
		sb.WriteString("📚 **Documentation** | ")
		sb.WriteString("🔧 **ProtoDocs** | ")
		sb.WriteString("✨ **Auto-Generated**\n\n")
	} else {
		sb.WriteString("*This documentation was automatically generated from Protocol Buffer definitions*\n\n")
	}

	sb.WriteString("</div>\n")
}

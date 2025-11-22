package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kyivinua/docgen-tool/tools/protodocs/docgen"
)

func main() {
	// Command line flags
	var (
		protoDir   = flag.String("proto-dir", "proto", "Directory containing proto files")
		outputDir  = flag.String("output-dir", "docs/consolidated", "Output directory for generated docs")
		themeName  = flag.String("theme", "default", "Mermaid diagram theme (default, forest, dark, neutral)")
		noEmoji    = flag.Bool("no-emoji", false, "Disable emoji icons")
		noExamples = flag.Bool("no-examples", false, "Skip code examples")
		verbose    = flag.Bool("verbose", false, "Verbose output")
	)

	flag.Parse()

	if *verbose {
		fmt.Println("ProtoDocs Consolidated Documentation Generator")
		fmt.Println("===============================================")
		fmt.Printf("Proto Directory: %s\n", *protoDir)
		fmt.Printf("Output Directory: %s\n", *outputDir)
		fmt.Printf("Theme: %s\n", *themeName)
		fmt.Println()
	}

	// Ensure output directory exists
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Find all proto files
	protoFiles, err := findProtoFiles(*protoDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding proto files: %v\n", err)
		os.Exit(1)
	}

	if len(protoFiles) == 0 {
		fmt.Fprintf(os.Stderr, "No proto files found in %s\n", *protoDir)
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("Found %d proto files:\n", len(protoFiles))
		for _, f := range protoFiles {
			fmt.Printf("  - %s\n", f)
		}
		fmt.Println()
	}

	// Determine import paths
	importPaths := []string{
		"/usr/include",  // For google protobuf well-known types
		*protoDir,
		filepath.Join(*protoDir, "common"),
		filepath.Join(*protoDir, "users"),
		filepath.Join(*protoDir, "payments"),
		filepath.Join(*protoDir, "notifications"),
		filepath.Join(*protoDir, "analytics"),
	}

	// Create parser
	parser := docgen.NewProtoParser(protoFiles, importPaths)

	// Parse proto files
	fmt.Println("Parsing proto files...")
	docs, err := parser.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing proto files: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nNote: Make sure protoc is installed and in PATH\n")
		fmt.Fprintf(os.Stderr, "Install: https://grpc.io/docs/protoc-installation/\n")
		os.Exit(1)
	}

	if len(docs) == 0 {
		fmt.Fprintf(os.Stderr, "No services found in proto files\n")
		os.Exit(1)
	}

	fmt.Printf("Found %d services\n", len(docs))

	// Configure generator
	config := docgen.ConsolidatedConfig{
		IncludeTOC:            true,
		TOCDepth:              3,
		IncludeDiagrams:       true,
		DiagramPosition:       "section",
		IncludeCrossReferences: true,
		IncludeAnchors:        true,
		IncludeArchitecture:   true,
		IncludeSequence:       true,
		IncludeMessageGraph:   true,
		IncludeDataFlow:       false, // Can be large
		IncludeOverview:       true,
		IncludeAuthentication: false,
		IncludeExamples:       !*noExamples,
		IncludeErrorCodes:     true,
		IncludeChangelog:      false,
		IncludeGoExamples:         !*noExamples, // Go examples enabled by default
		IncludePythonExamples:     false,        // Python examples excluded
		IncludeJavaScriptExamples: !*noExamples, // JavaScript examples enabled by default
		UseEmojis:             !*noEmoji,
		CodeHighlighting:      "protobuf",
		DiagramTheme:          *themeName,
	}

	generator := docgen.NewConsolidatedDocGenerator(config)

	// Generate documentation for each service
	fmt.Println("\nGenerating consolidated documentation...")

	totalMethods := 0
	totalMessages := 0
	totalEnums := 0

	for _, doc := range docs {
		if *verbose {
			fmt.Printf("\n  Service: %s\n", doc.Service.Name)
			fmt.Printf("    Methods: %d\n", len(doc.Methods))
			fmt.Printf("    Messages: %d\n", len(doc.Messages))
			fmt.Printf("    Enums: %d\n", len(doc.Enums))
		}

		totalMethods += len(doc.Methods)
		totalMessages += len(doc.Messages)
		totalEnums += len(doc.Enums)

		// Generate markdown
		markdown := generator.GenerateConsolidatedDoc(doc)

		// Write to file
		filename := filepath.Join(*outputDir, fmt.Sprintf("%s.md", doc.Service.Name))
		if err := os.WriteFile(filename, []byte(markdown), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", filename, err)
			continue
		}

		fmt.Printf("  ✓ Generated: %s (%d KB)\n", filename, len(markdown)/1024)
	}

	// Generate index file
	fmt.Println("\nGenerating index...")
	indexContent := generateIndex(docs, config)
	indexFile := filepath.Join(*outputDir, "README.md")
	if err := os.WriteFile(indexFile, []byte(indexContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing index: %v\n", err)
	} else {
		fmt.Printf("  ✓ Generated: %s\n", indexFile)
	}

	// Summary
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("✅ Documentation Generation Complete!")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("\nStatistics:\n")
	fmt.Printf("  Services:    %d\n", len(docs))
	fmt.Printf("  Methods:     %d\n", totalMethods)
	fmt.Printf("  Messages:    %d\n", totalMessages)
	fmt.Printf("  Enumerations: %d\n", totalEnums)
	fmt.Printf("\nOutput:\n")
	fmt.Printf("  Directory:   %s\n", *outputDir)
	fmt.Printf("  Files:       %d service docs + 1 index\n", len(docs))
	fmt.Println()
}

// findProtoFiles recursively finds all .proto files in a directory
func findProtoFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".proto" {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// generateIndex generates an index/README file
func generateIndex(docs []*docgen.ServiceDocumentation, config docgen.ConsolidatedConfig) string {
	var content string

	emoji := ""
	if config.UseEmojis {
		emoji = "📚 "
	}

	content += fmt.Sprintf("# %sAPI Documentation\n\n", emoji)
	content += "Comprehensive, consolidated documentation for all services.\n\n"
	content += "---\n\n"

	content += "## Services\n\n"
	content += "| Service | Package | Methods | Messages | Enums |\n"
	content += "|---------|---------|---------|----------|-------|\n"

	for _, doc := range docs {
		content += fmt.Sprintf("| [%s](%s.md) | `%s` | %d | %d | %d |\n",
			doc.Service.Name,
			doc.Service.Name,
			doc.Service.Package,
			len(doc.Methods),
			len(doc.Messages),
			len(doc.Enums))
	}

	content += "\n## Documentation Features\n\n"
	content += "Each service documentation includes:\n\n"
	content += "- " + checkMark(config.UseEmojis) + " **Table of Contents** - Easy navigation\n"
	content += "- " + checkMark(config.UseEmojis) + " **Overview** - Service statistics and quick start\n"

	if config.IncludeArchitecture {
		content += "- " + checkMark(config.UseEmojis) + " **Architecture Diagrams** - Mermaid diagrams showing service structure\n"
	}

	if config.IncludeSequence {
		content += "- " + checkMark(config.UseEmojis) + " **Sequence Diagrams** - Per-method call flows\n"
	}

	content += "- " + checkMark(config.UseEmojis) + " **Method Details** - Complete RPC specifications\n"
	content += "- " + checkMark(config.UseEmojis) + " **Message Definitions** - Field tables and proto definitions\n"

	if config.IncludeExamples {
		content += "- " + checkMark(config.UseEmojis) + " **Code Examples** - Go, Python, JavaScript\n"
	}

	if config.IncludeErrorCodes {
		content += "- " + checkMark(config.UseEmojis) + " **Error Codes** - gRPC status codes and handling\n"
	}

	if config.IncludeCrossReferences {
		content += "- " + checkMark(config.UseEmojis) + " **Cross-References** - Links between related types\n"
	}

	content += "\n## Quick Links\n\n"

	for _, doc := range docs {
		content += fmt.Sprintf("### [%s](%s.md)\n\n", doc.Service.Name, doc.Service.Name)
		content += fmt.Sprintf("%s\n\n", doc.Service.Description)

		content += "**Methods:**\n"
		for i, method := range doc.Methods {
			if i >= 5 {
				content += fmt.Sprintf("- ... and %d more\n", len(doc.Methods)-5)
				break
			}
			content += fmt.Sprintf("- `%s` - %s\n", method.Name, method.Description)
		}
		content += "\n"
	}

	content += "---\n\n"
	content += "<div align=\"center\">\n\n"
	content += "*Auto-generated by ProtoDocs Consolidated Documentation Generator*\n\n"
	content += "</div>\n"

	return content
}

func checkMark(useEmoji bool) string {
	if useEmoji {
		return "✓"
	}
	return "*"
}

// Repeat is a helper since strings.Repeat might not be available
type stringRepeater string

func (s stringRepeater) Repeat(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += string(s)
	}
	return result
}

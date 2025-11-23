package diagrams

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DiagramExportFormat represents the export format for diagrams
type DiagramExportFormat string

const (
	FormatGraphML DiagramExportFormat = "graphml"
	FormatMermaid DiagramExportFormat = "mermaid"
	FormatSVG     DiagramExportFormat = "svg"
	FormatPNG     DiagramExportFormat = "png"
)

// ExportConfig holds configuration for diagram export
type ExportConfig struct {
	OutputDir         string
	Formats           []DiagramExportFormat
	CreateIndex       bool
	IncludeTimestamp  bool
	ServiceSubfolders bool // Create subfolder per service
}

// DiagramExport represents an exported diagram file
type DiagramExport struct {
	Type        string    `json:"type"`
	Format      string    `json:"format"`
	Filename    string    `json:"filename"`
	FilePath    string    `json:"file_path"`
	ServiceName string    `json:"service_name"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	GeneratedAt time.Time `json:"generated_at"`
	FileSize    int64     `json:"file_size"`
}

// DiagramManifest holds information about all exported diagrams
type DiagramManifest struct {
	GeneratedAt time.Time        `json:"generated_at"`
	Diagrams    []DiagramExport  `json:"diagrams"`
	Statistics  map[string]int64 `json:"statistics"`
}

// DiagramExporter handles exporting diagrams to various formats
type DiagramExporter struct {
	config *ExportConfig
}

// NewDiagramExporter creates a new diagram exporter
func NewDiagramExporter(config *ExportConfig) *DiagramExporter {
	if config == nil {
		config = &ExportConfig{
			OutputDir:         "./diagrams",
			Formats:           []DiagramExportFormat{FormatGraphML, FormatMermaid},
			CreateIndex:       true,
			IncludeTimestamp:  false,
			ServiceSubfolders: true,
		}
	}
	return &DiagramExporter{config: config}
}

// ExportServiceDiagrams exports all diagrams for a service
func (e *DiagramExporter) ExportServiceDiagrams(service *DocService, messages []DocMessage, enums []DocEnum) ([]DiagramExport, error) {
	exports := []DiagramExport{}

	// Create output directory
	outputDir := e.config.OutputDir
	if e.config.ServiceSubfolders {
		outputDir = filepath.Join(e.config.OutputDir, sanitizeFilename(service.Name))
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Export GraphML if requested
	for _, format := range e.config.Formats {
		switch format {
		case FormatGraphML:
			export, err := e.exportGraphML(service, messages, outputDir)
			if err != nil {
				return nil, fmt.Errorf("failed to export GraphML: %w", err)
			}
			exports = append(exports, *export)

		case FormatMermaid:
			mermaidExports, err := e.exportMermaidDiagrams(service, messages, enums, outputDir)
			if err != nil {
				return nil, fmt.Errorf("failed to export Mermaid diagrams: %w", err)
			}
			exports = append(exports, mermaidExports...)
		}
	}

	return exports, nil
}

// exportGraphML exports GraphML diagram for a service
func (e *DiagramExporter) exportGraphML(service *DocService, messages []DocMessage, outputDir string) (*DiagramExport, error) {
	generator := NewServiceGraphMLGenerator()
	graphml, err := generator.GenerateServiceGraphML(service, messages)
	if err != nil {
		return nil, err
	}

	xml, err := graphml.ToXML()
	if err != nil {
		return nil, err
	}

	filename := e.buildFilename(service.Name, "service-graph", "graphml")
	filepath := filepath.Join(outputDir, filename)

	if err := os.WriteFile(filepath, []byte(xml), 0644); err != nil {
		return nil, err
	}

	fileInfo, _ := os.Stat(filepath)

	return &DiagramExport{
		Type:        "graphml",
		Format:      "graphml",
		Filename:    filename,
		FilePath:    filepath,
		ServiceName: service.Name,
		Title:       fmt.Sprintf("%s Service Graph", service.Name),
		Description: "GraphML representation of service architecture",
		GeneratedAt: time.Now(),
		FileSize:    fileInfo.Size(),
	}, nil
}

// exportMermaidDiagrams exports all Mermaid diagrams for a service
func (e *DiagramExporter) exportMermaidDiagrams(service *DocService, messages []DocMessage, enums []DocEnum, outputDir string) ([]DiagramExport, error) {
	exports := []DiagramExport{}

	config := DefaultDiagramConfig()
	generator := NewEnhancedMermaidGenerator(config)

	// Generate all diagram types
	diagrams := generator.GenerateCompleteDiagramSet(service, messages, enums)

	// Define diagram metadata
	diagramMeta := map[string]struct {
		title       string
		description string
	}{
		"architecture":  {"Architecture Diagram", "Service architecture with methods and messages"},
		"sequence":      {"Sequence Diagram", "Request/response flow for all methods"},
		"class":         {"Class Diagram", "UML class diagram of message types"},
		"erd":           {"Entity Relationship Diagram", "Message relationships and cardinality"},
		"state_machine": {"State Machine Diagram", "Service lifecycle states"},
		"flowchart":     {"Method Flowchart", "Request processing flow"},
		"mindmap":       {"Service Mindmap", "Hierarchical service overview"},
		"c4_context":    {"C4 Context Diagram", "System context view"},
		"c4_container":  {"C4 Container Diagram", "Container-level architecture"},
		"data_flow":     {"Data Flow Diagram", "Data flow through service"},
	}

	// Export each diagram type
	for diagramType, content := range diagrams {
		meta, ok := diagramMeta[diagramType]
		if !ok {
			continue
		}

		filename := e.buildFilename(service.Name, diagramType, "mmd")
		filepath := filepath.Join(outputDir, filename)

		if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
			return nil, err
		}

		fileInfo, _ := os.Stat(filepath)

		exports = append(exports, DiagramExport{
			Type:        diagramType,
			Format:      "mermaid",
			Filename:    filename,
			FilePath:    filepath,
			ServiceName: service.Name,
			Title:       meta.title,
			Description: meta.description,
			GeneratedAt: time.Now(),
			FileSize:    fileInfo.Size(),
		})
	}

	return exports, nil
}

// ExportAllDiagrams exports diagrams for all services in the model
func (e *DiagramExporter) ExportAllDiagrams(model *ApiDocModel) (*DiagramManifest, error) {
	manifest := &DiagramManifest{
		GeneratedAt: time.Now(),
		Diagrams:    []DiagramExport{},
		Statistics:  make(map[string]int64),
	}

	// Export diagrams for each module's services
	for _, module := range model.Modules {
		for _, service := range module.Services {
			exports, err := e.ExportServiceDiagrams(&service, module.Messages, module.Enums)
			if err != nil {
				return nil, fmt.Errorf("failed to export diagrams for service %s: %w", service.Name, err)
			}
			manifest.Diagrams = append(manifest.Diagrams, exports...)
		}
	}

	// Calculate statistics
	manifest.Statistics["total_diagrams"] = int64(len(manifest.Diagrams))
	formatCounts := make(map[string]int64)
	typeCounts := make(map[string]int64)
	var totalSize int64

	for _, diagram := range manifest.Diagrams {
		formatCounts[diagram.Format]++
		typeCounts[diagram.Type]++
		totalSize += diagram.FileSize
	}

	manifest.Statistics["total_size_bytes"] = totalSize
	for format, count := range formatCounts {
		manifest.Statistics[fmt.Sprintf("count_%s", format)] = count
	}
	for diagType, count := range typeCounts {
		manifest.Statistics[fmt.Sprintf("type_%s", diagType)] = count
	}

	// Create index if requested
	if e.config.CreateIndex {
		if err := e.createIndex(manifest); err != nil {
			return nil, fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Save manifest
	if err := e.saveManifest(manifest); err != nil {
		return nil, fmt.Errorf("failed to save manifest: %w", err)
	}

	return manifest, nil
}

// createIndex creates an index.md file with links to all diagrams
func (e *DiagramExporter) createIndex(manifest *DiagramManifest) error {
	var sb strings.Builder

	sb.WriteString("# Generated Diagrams\n\n")
	sb.WriteString(fmt.Sprintf("Generated at: %s\n\n", manifest.GeneratedAt.Format(time.RFC3339)))

	// Statistics
	sb.WriteString("## Statistics\n\n")
	sb.WriteString(fmt.Sprintf("- Total Diagrams: %d\n", manifest.Statistics["total_diagrams"]))
	sb.WriteString(fmt.Sprintf("- Total Size: %d bytes (%.2f KB)\n",
		manifest.Statistics["total_size_bytes"],
		float64(manifest.Statistics["total_size_bytes"])/1024))
	sb.WriteString("\n")

	// Group by service
	serviceMap := make(map[string][]DiagramExport)
	for _, diagram := range manifest.Diagrams {
		serviceMap[diagram.ServiceName] = append(serviceMap[diagram.ServiceName], diagram)
	}

	sb.WriteString("## Diagrams by Service\n\n")
	for serviceName, diagrams := range serviceMap {
		sb.WriteString(fmt.Sprintf("### %s\n\n", serviceName))
		for _, diagram := range diagrams {
			relPath, _ := filepath.Rel(e.config.OutputDir, diagram.FilePath)
			sb.WriteString(fmt.Sprintf("- [%s](%s) - %s (%s, %.2f KB)\n",
				diagram.Title,
				relPath,
				diagram.Description,
				diagram.Format,
				float64(diagram.FileSize)/1024))
		}
		sb.WriteString("\n")
	}

	// Group by type
	typeMap := make(map[string][]DiagramExport)
	for _, diagram := range manifest.Diagrams {
		typeMap[diagram.Type] = append(typeMap[diagram.Type], diagram)
	}

	sb.WriteString("## Diagrams by Type\n\n")
	for diagType, diagrams := range typeMap {
		sb.WriteString(fmt.Sprintf("### %s (%d)\n\n", diagType, len(diagrams)))
		for _, diagram := range diagrams {
			relPath, _ := filepath.Rel(e.config.OutputDir, diagram.FilePath)
			sb.WriteString(fmt.Sprintf("- [%s - %s](%s)\n",
				diagram.ServiceName,
				diagram.Title,
				relPath))
		}
		sb.WriteString("\n")
	}

	indexPath := filepath.Join(e.config.OutputDir, "index.md")
	return os.WriteFile(indexPath, []byte(sb.String()), 0644)
}

// saveManifest saves the manifest as JSON
func (e *DiagramExporter) saveManifest(manifest *DiagramManifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	manifestPath := filepath.Join(e.config.OutputDir, "manifest.json")
	return os.WriteFile(manifestPath, data, 0644)
}

// buildFilename constructs a filename for a diagram
func (e *DiagramExporter) buildFilename(serviceName, diagramType, extension string) string {
	sanitized := sanitizeFilename(serviceName)
	filename := fmt.Sprintf("%s-%s.%s", sanitized, diagramType, extension)

	if e.config.IncludeTimestamp {
		timestamp := time.Now().Format("20060102-150405")
		filename = fmt.Sprintf("%s-%s-%s.%s", sanitized, diagramType, timestamp, extension)
	}

	return filename
}

// sanitizeFilename sanitizes a name for use in filenames
func sanitizeFilename(name string) string {
	// Replace non-alphanumeric characters with underscores
	result := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, name)

	// Convert to lowercase
	result = strings.ToLower(result)

	// Remove consecutive underscores
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}

	return strings.Trim(result, "_")
}

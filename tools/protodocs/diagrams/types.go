package diagrams

import (
	"time"
)

// DiagramType represents the type of diagram to generate
type DiagramType string

const (
	// Static diagrams (don't require ApiDocModel)
	DiagramTypePipeline   DiagramType = "pipeline"   // Pipeline architecture
	DiagramTypeEnricher   DiagramType = "enricher"   // Enricher orchestration
	DiagramTypeComponent  DiagramType = "component"  // Component interaction
	DiagramTypeTransform  DiagramType = "transform"  // Proto → Doc transformation
	DiagramTypeDeploy     DiagramType = "deploy"     // Deployment architecture

	// Dynamic diagrams (generated from ApiDocModel)
	DiagramTypeDataModel  DiagramType = "data_model" // Data model ER diagram
	DiagramTypeServiceMap DiagramType = "service_map" // Service relationship map
	DiagramTypeMessageHierarchy DiagramType = "message_hierarchy" // Message inheritance
)

// DiagramConfig holds configuration for diagram generation
type DiagramConfig struct {
	// Output settings
	OutputDir string `yaml:"output_dir"`

	// Enable/disable specific diagrams
	EnablePipeline          bool `yaml:"enable_pipeline"`
	EnableEnricher          bool `yaml:"enable_enricher"`
	EnableComponent         bool `yaml:"enable_component"`
	EnableTransform         bool `yaml:"enable_transform"`
	EnableDeploy            bool `yaml:"enable_deploy"`
	EnableDataModel         bool `yaml:"enable_data_model"`
	EnableServiceMap        bool `yaml:"enable_service_map"`
	EnableMessageHierarchy  bool `yaml:"enable_message_hierarchy"`

	// Generation settings
	GenerateIndex    bool `yaml:"generate_index"`     // Generate index.md linking all diagrams
	IncludeTimestamp bool `yaml:"include_timestamp"`  // Add timestamp to diagrams
	Theme            string `yaml:"theme"`            // Mermaid theme: default, forest, dark, neutral

	// Model-specific settings (for dynamic diagrams)
	MaxServicesPerDiagram  int `yaml:"max_services_per_diagram"`  // Split large diagrams
	MaxMessagesPerDiagram  int `yaml:"max_messages_per_diagram"`
	IncludePrivateTypes    bool `yaml:"include_private_types"`    // Include internal types
}

// DefaultDiagramConfig returns default configuration
func DefaultDiagramConfig() *DiagramConfig {
	return &DiagramConfig{
		OutputDir:              "./api-docs/diagrams",
		EnablePipeline:         true,
		EnableEnricher:         true,
		EnableComponent:        true,
		EnableTransform:        true,
		EnableDeploy:           true,
		EnableDataModel:        true,
		EnableServiceMap:       true,
		EnableMessageHierarchy: false, // Can be very large
		GenerateIndex:          true,
		IncludeTimestamp:       true,
		Theme:                  "default",
		MaxServicesPerDiagram:  20,
		MaxMessagesPerDiagram:  30,
		IncludePrivateTypes:    false,
	}
}

// DiagramMetadata holds metadata about a generated diagram
type DiagramMetadata struct {
	Type        DiagramType
	Title       string
	Description string
	Filename    string
	GeneratedAt time.Time
	MermaidType string // flowchart, erDiagram, graph, sequenceDiagram, etc.
}

// GenerationResult holds the result of diagram generation
type GenerationResult struct {
	Metadata DiagramMetadata
	Content  string // The Mermaid diagram content
	Error    error
}

// ApiDocModel is imported from pipeline package
// We define it here to avoid circular dependencies
type ApiDocModel struct {
	Modules      []DocModule        `json:"modules"`
	GeneratedAt  time.Time          `json:"generated_at"`
	SourceCommit string             `json:"source_commit"`
	Statistics   map[string]int64   `json:"statistics,omitempty"`
	Tools        map[string]string  `json:"tools,omitempty"`
}

type DocModule struct {
	Name        string       `json:"name"`
	Package     string       `json:"package"`
	Description string       `json:"description,omitempty"`
	FilePath    string       `json:"file_path"`
	Services    []DocService `json:"services,omitempty"`
	Messages    []DocMessage `json:"messages,omitempty"`
	Enums       []DocEnum    `json:"enums,omitempty"`
}

type DocService struct {
	Name        string      `json:"name"`
	FullName    string      `json:"full_name"`
	Description string      `json:"description,omitempty"`
	Methods     []DocMethod `json:"methods,omitempty"`
	Visibility  string      `json:"visibility,omitempty"`
}

type DocMethod struct {
	Name              string            `json:"name"`
	FullName          string            `json:"full_name"`
	Description       string            `json:"description,omitempty"`
	InputType         string            `json:"input_type"`
	OutputType        string            `json:"output_type"`
	ClientStreaming   bool              `json:"client_streaming"`
	ServerStreaming   bool              `json:"server_streaming"`
	HTTPMethods       []HTTPMethodInfo  `json:"http_methods,omitempty"`
	Visibility        string            `json:"visibility,omitempty"`
}

type HTTPMethodInfo struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type DocMessage struct {
	Name        string     `json:"name"`
	FullName    string     `json:"full_name"`
	Description string     `json:"description,omitempty"`
	Fields      []DocField `json:"fields,omitempty"`
	Visibility  string     `json:"visibility,omitempty"`
}

type DocField struct {
	Name        string `json:"name"`
	Number      int32  `json:"number"`
	Type        string `json:"type"`
	TypeName    string `json:"type_name,omitempty"` // For message/enum references
	Label       string `json:"label"` // optional, required, repeated
	Description string `json:"description,omitempty"`
	OneofGroup  string `json:"oneof_group,omitempty"`
}

type DocEnum struct {
	Name        string         `json:"name"`
	FullName    string         `json:"full_name"`
	Description string         `json:"description,omitempty"`
	Values      []DocEnumValue `json:"values,omitempty"`
	Visibility  string         `json:"visibility,omitempty"`
}

type DocEnumValue struct {
	Name        string `json:"name"`
	Number      int32  `json:"number"`
	Description string `json:"description,omitempty"`
}

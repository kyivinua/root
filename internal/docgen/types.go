// Package docgen provides core data models for the documentation generator.
package docgen

import "time"

// Service represents a Protocol Buffer service with its metadata.
type Service struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Package     string    `json:"package"`
	Methods     []Method  `json:"methods"`
	Messages    []Message `json:"messages"`
	Version     string    `json:"version"`
	ProtoFiles  []string  `json:"proto_files"`
	Generated   time.Time `json:"generated"`
}

// Method represents a service method/RPC endpoint.
type Method struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	InputType       string `json:"input_type"`
	OutputType      string `json:"output_type"`
	ClientStreaming bool   `json:"client_streaming"`
	ServerStreaming bool   `json:"server_streaming"`
	Deprecated      bool   `json:"deprecated"`
	Comments        string `json:"comments"`
}

// Message represents a Protocol Buffer message definition.
type Message struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Fields      []Field `json:"fields"`
	IsNested    bool    `json:"is_nested"`
	Comments    string  `json:"comments"`
}

// Field represents a message field.
type Field struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Number      int    `json:"number"`
	Label       string `json:"label"` // optional, required, repeated
	Default     string `json:"default"`
	Deprecated  bool   `json:"deprecated"`
	Comments    string `json:"comments"`
}

// Documentation represents the complete generated documentation.
type Documentation struct {
	Services   []Service         `json:"services"`
	Metadata   DocumentMetadata  `json:"metadata"`
	Quality    QualityReport     `json:"quality"`
	Diagrams   map[string]string `json:"diagrams"` // diagram name -> content
	GeneratedAt time.Time        `json:"generated_at"`
}

// DocumentMetadata contains metadata about the documentation.
type DocumentMetadata struct {
	ProjectName string            `json:"project_name"`
	Version     string            `json:"version"`
	Generated   time.Time         `json:"generated"`
	Generator   string            `json:"generator"`
	Config      map[string]string `json:"config"`
}

// QualityReport contains quality metrics for the documentation.
type QualityReport struct {
	CoverageScore          float64                `json:"coverage_score"`
	DescriptionQuality     float64                `json:"description_quality"`
	MethodCoverage         float64                `json:"method_coverage"`
	FieldCoverage          float64                `json:"field_coverage"`
	ServiceMetrics         map[string]ServiceQuality `json:"service_metrics"`
	Issues                 []QualityIssue         `json:"issues"`
	Passed                 bool                   `json:"passed"`
}

// ServiceQuality contains quality metrics for a single service.
type ServiceQuality struct {
	ServiceName        string  `json:"service_name"`
	Coverage           float64 `json:"coverage"`
	DescriptionQuality float64 `json:"description_quality"`
	MethodCount        int     `json:"method_count"`
	DocumentedMethods  int     `json:"documented_methods"`
	FieldCount         int     `json:"field_count"`
	DocumentedFields   int     `json:"documented_fields"`
}

// QualityIssue represents a quality validation issue.
type QualityIssue struct {
	Severity    string `json:"severity"` // error, warning, info
	Type        string `json:"type"`
	Message     string `json:"message"`
	Location    string `json:"location"`
	Suggestion  string `json:"suggestion"`
	AutoFixable bool   `json:"auto_fixable"`
}

// EnrichmentRequest represents a request for AI enrichment.
type EnrichmentRequest struct {
	ServiceName string    `json:"service_name"`
	Content     string    `json:"content"`
	Context     string    `json:"context"`
	Type        string    `json:"type"` // method, message, field, service
	Metadata    map[string]string `json:"metadata"`
}

// EnrichmentResponse represents the AI enrichment response.
type EnrichmentResponse struct {
	EnrichedContent string    `json:"enriched_content"`
	Confidence      float64   `json:"confidence"`
	Suggestions     []string  `json:"suggestions"`
	ProcessedAt     time.Time `json:"processed_at"`
	CacheHit        bool      `json:"cache_hit"`
}

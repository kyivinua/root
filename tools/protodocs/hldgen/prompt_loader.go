package hldgen

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"text/template"
)

// PromptTemplate represents a loaded prompt template
type PromptTemplate struct {
	Version      string `xml:"version,attr"`
	Role         string `xml:"role,attr"`
	System       string `xml:"system"`
	Context      string `xml:"context"`
	Thinking     string `xml:"thinking"`
	Instructions string `xml:"instructions"`
}

// PromptLoader loads and manages prompt templates
type PromptLoader struct {
	basePath string
	version  string
	cache    map[string]*PromptTemplate
	mu       sync.RWMutex
}

// NewPromptLoader creates a new prompt loader
func NewPromptLoader(basePath, version string) *PromptLoader {
	if basePath == "" {
		basePath = "prompts"
	}
	if version == "" {
		version = "v1"
	}
	return &PromptLoader{
		basePath: basePath,
		version:  version,
		cache:    make(map[string]*PromptTemplate),
	}
}

// Load loads a prompt template for the specified role
func (pl *PromptLoader) Load(role AgentRole) (*PromptTemplate, error) {
	roleStr := string(role)

	// Check cache first
	pl.mu.RLock()
	if cached, ok := pl.cache[roleStr]; ok {
		pl.mu.RUnlock()
		return cached, nil
	}
	pl.mu.RUnlock()

	// Load from file
	filePath := filepath.Join(pl.basePath, pl.version, fmt.Sprintf("%s.xml", roleStr))
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read prompt template %s: %w", filePath, err)
	}

	// Parse XML
	var pt PromptTemplate
	if err := xml.Unmarshal(data, &pt); err != nil {
		return nil, fmt.Errorf("failed to parse prompt template %s: %w", filePath, err)
	}

	// Cache the template
	pl.mu.Lock()
	pl.cache[roleStr] = &pt
	pl.mu.Unlock()

	return &pt, nil
}

// PromptData holds data for template rendering
type PromptData struct {
	// Basic metadata
	ModuleName    string
	ServicesCount int
	MessagesCount int
	MethodsCount  int
	SourceCommit  string
	Domain        string
	Complexity    string

	// Services info
	Services []ServiceTemplateData

	// Context info
	RAGContext     []RAGTemplateData
	PreviousDraft  string
	Criticism      []CriticismTemplateData
	Dependencies   []DependencyTemplateData
	CriticalPaths  []CriticalPathData
	ExpectedRPS    string
	CriticalPath   string

	// Security/Compliance
	DataClassification string
	PII                bool
	PaymentData        bool
}

// ServiceTemplateData holds service data for templates
type ServiceTemplateData struct {
	Name        string
	Description string
	MethodCount int
	Methods     []MethodTemplateData
}

// MethodTemplateData holds method data for templates
type MethodTemplateData struct {
	Name        string
	Description string
	InputType   string
	OutputType  string
}

// RAGTemplateData holds RAG context for templates
type RAGTemplateData struct {
	Source string
	Score  float64
}

// CriticismTemplateData holds criticism data
type CriticismTemplateData struct {
	Score       float64
	IssuesCount int
	Suggestion  string
}

// DependencyTemplateData holds dependency info
type DependencyTemplateData struct {
	Name string
	Type string
}

// CriticalPathData holds critical path info
type CriticalPathData struct {
	Name        string
	Description string
}

// Render renders a prompt template with the provided data
func (pl *PromptLoader) Render(role AgentRole, data PromptData) (string, error) {
	// Load template
	pt, err := pl.Load(role)
	if err != nil {
		return "", err
	}

	// Build full prompt
	var sb bytes.Buffer

	// Render system section
	if pt.System != "" {
		sb.WriteString("<system>\n")
		rendered, err := renderTemplate(pt.System, data)
		if err != nil {
			return "", fmt.Errorf("failed to render system section: %w", err)
		}
		sb.WriteString(rendered)
		sb.WriteString("\n</system>\n\n")
	}

	// Render context section
	if pt.Context != "" {
		sb.WriteString("<context>\n")
		rendered, err := renderTemplate(pt.Context, data)
		if err != nil {
			return "", fmt.Errorf("failed to render context section: %w", err)
		}
		sb.WriteString(rendered)
		sb.WriteString("\n</context>\n\n")
	}

	// Render thinking section
	if pt.Thinking != "" {
		sb.WriteString("<thinking>\n")
		rendered, err := renderTemplate(pt.Thinking, data)
		if err != nil {
			return "", fmt.Errorf("failed to render thinking section: %w", err)
		}
		sb.WriteString(rendered)
		sb.WriteString("\n</thinking>\n\n")
	}

	// Render instructions section
	if pt.Instructions != "" {
		sb.WriteString("<instructions>\n")
		rendered, err := renderTemplate(pt.Instructions, data)
		if err != nil {
			return "", fmt.Errorf("failed to render instructions section: %w", err)
		}
		sb.WriteString(rendered)
		sb.WriteString("\n</instructions>\n")
	}

	return sb.String(), nil
}

// renderTemplate renders a template string with data
func renderTemplate(tmplStr string, data PromptData) (string, error) {
	tmpl, err := template.New("prompt").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// ConvertAgentInputToPromptData converts AgentInput to PromptData
func ConvertAgentInputToPromptData(input *AgentInput) PromptData {
	data := PromptData{
		ModuleName:    input.Docs.ModuleName,
		ServicesCount: len(input.Docs.Services),
		MessagesCount: len(input.Docs.Messages),
		SourceCommit:  input.Docs.SourceCommit,
	}

	// Convert services
	totalMethods := 0
	for _, svc := range input.Docs.Services {
		totalMethods += len(svc.Methods)

		methods := make([]MethodTemplateData, len(svc.Methods))
		for i, method := range svc.Methods {
			methods[i] = MethodTemplateData{
				Name:        method.Name,
				Description: method.Description,
				InputType:   method.InputType,
				OutputType:  method.OutputType,
			}
		}

		data.Services = append(data.Services, ServiceTemplateData{
			Name:        svc.Name,
			Description: svc.Description,
			MethodCount: len(svc.Methods),
			Methods:     methods,
		})
	}
	data.MethodsCount = totalMethods

	// Convert RAG context
	if input.EnrichedCtx != nil && len(input.EnrichedCtx.RAGContext) > 0 {
		for _, rag := range input.EnrichedCtx.RAGContext {
			data.RAGContext = append(data.RAGContext, RAGTemplateData{
				Source: rag.Source,
				Score:  rag.Score,
			})
		}
	}

	// Previous draft
	if input.PreviousDraft != nil {
		data.PreviousDraft = input.PreviousDraft.Architecture
	}

	// Criticism
	for _, crit := range input.Criticism {
		data.Criticism = append(data.Criticism, CriticismTemplateData{
			Score:       crit.Score,
			IssuesCount: len(crit.Issues),
			Suggestion:  crit.Suggestion,
		})
	}

	return data
}

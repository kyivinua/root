package adapters

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"text/template"

	"github.com/kyivinua/docgen-tool/tools/protodocs/enricher"
)

// CoTTemplateEngine implements PromptTemplateEngine with Chain-of-Thought and XML formatting
type CoTTemplateEngine struct {
	mu        sync.RWMutex
	templates map[string]*template.Template
	useCoT    bool
	format    string // "xml", "plain", "cot"
}

// NewCoTTemplateEngine creates a new CoT template engine
func NewCoTTemplateEngine(useCoT bool, format string) (*CoTTemplateEngine, error) {
	engine := &CoTTemplateEngine{
		templates: make(map[string]*template.Template),
		useCoT:    useCoT,
		format:    format,
	}

	// Register default templates
	if err := engine.registerDefaultTemplates(); err != nil {
		return nil, fmt.Errorf("register default templates: %w", err)
	}

	return engine, nil
}

// Render renders a template with the given data
func (e *CoTTemplateEngine) Render(templateName string, data map[string]interface{}) (string, error) {
	e.mu.RLock()
	tmpl, ok := e.templates[templateName]
	e.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("template %s not found", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

// RegisterTemplate registers a new template
func (e *CoTTemplateEngine) RegisterTemplate(name string, templateStr string) error {
	tmpl, err := template.New(name).Parse(templateStr)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	e.mu.Lock()
	e.templates[name] = tmpl
	e.mu.Unlock()

	return nil
}

// registerDefaultTemplates registers built-in templates
func (e *CoTTemplateEngine) registerDefaultTemplates() error {
	templates := map[string]string{
		"enrich_method": e.getMethodTemplate(),
		"enrich_message": e.getMessageTemplate(),
		"enrich_field": e.getFieldTemplate(),
	}

	for name, tmplStr := range templates {
		if err := e.RegisterTemplate(name, tmplStr); err != nil {
			return fmt.Errorf("register %s: %w", name, err)
		}
	}

	return nil
}

// getMethodTemplate returns the method enrichment template
func (e *CoTTemplateEngine) getMethodTemplate() string {
	if e.format == "xml" && e.useCoT {
		return `You are a technical documentation expert specializing in gRPC/Protocol Buffer APIs. Your task is to enrich method documentation.

<task>
Enrich the documentation for the following gRPC method.
</task>

<context>
<service_name>{{ index .context "service_name" }}</service_name>
<method_name>{{ index .context "method_name" }}</method_name>
<package>{{ index .context "package_name" }}</package>
<input_type>{{ index .context "input_type" }}</input_type>
<output_type>{{ index .context "output_type" }}</output_type>
{{ if index .context "is_streaming" }}<streaming>true</streaming>{{ end }}
</context>

<current_documentation>
{{ .current_docs }}
</current_documentation>

{{ if .use_rag }}
<retrieved_context>
{{ .rag_context }}
</retrieved_context>
{{ end }}

<instructions>
Follow these steps to enrich the documentation:

<thinking>
1. Analyze the current documentation
2. Identify what information is missing or unclear
3. Consider the method's role in the service
4. Think about what developers need to know
{{ if .use_rag }}5. Incorporate relevant context from retrieved documentation{{ end }}
</thinking>

<requirements>
- Maintain technical accuracy
- Be concise but comprehensive
- Use clear, professional language
- Include purpose, behavior, and usage guidance
- Do NOT invent implementation details
- Do NOT include example code unless confident
- Do NOT add version-specific information without evidence
</requirements>
</instructions>

Provide the enriched documentation as a single paragraph or 2-3 short paragraphs. Do not include markdown formatting or section headers. Output only the documentation text.`
	}

	// Plain format
	return `Enrich the following gRPC method documentation:

Service: {{ index .context "service_name" }}
Method: {{ index .context "method_name" }}
Package: {{ index .context "package_name" }}
Input: {{ index .context "input_type" }}
Output: {{ index .context "output_type" }}

Current documentation:
{{ .current_docs }}

{{ if .use_rag }}
Additional context:
{{ .rag_context }}
{{ end }}

Provide enriched documentation that:
- Explains the method's purpose clearly
- Describes expected behavior
- Maintains technical accuracy
- Is concise and professional

Output only the enriched documentation text, without markdown formatting.`
}

// getMessageTemplate returns the message enrichment template
func (e *CoTTemplateEngine) getMessageTemplate() string {
	if e.format == "xml" && e.useCoT {
		return `You are a technical documentation expert specializing in Protocol Buffer message types.

<task>
Enrich the documentation for the following Protocol Buffer message.
</task>

<context>
<message_name>{{ index .context "message_name" }}</message_name>
<package>{{ index .context "package_name" }}</package>
<field_count>{{ index .context "field_count" }}</field_count>
<fields>{{ range index .context "fields" }}{{ . }}, {{ end }}</fields>
</context>

<current_documentation>
{{ .current_docs }}
</current_documentation>

{{ if .use_rag }}
<retrieved_context>
{{ .rag_context }}
</retrieved_context>
{{ end }}

<instructions>
<thinking>
1. Understand the message's role in the API
2. Consider relationships between fields
3. Think about usage patterns
{{ if .use_rag }}4. Incorporate relevant patterns from retrieved context{{ end }}
</thinking>

<requirements>
- Describe the message's purpose and usage
- Explain what the message represents
- Be factual and precise
- Do NOT describe individual fields (that's separate)
- Do NOT invent data validation rules
- Maintain consistency with API terminology
</requirements>
</instructions>

Provide the enriched documentation as a concise description. Output only the documentation text.`
	}

	return `Enrich the following Protocol Buffer message documentation:

Message: {{ index .context "message_name" }}
Package: {{ index .context "package_name" }}
Fields: {{ range index .context "fields" }}{{ . }}, {{ end }}

Current documentation:
{{ .current_docs }}

{{ if .use_rag }}
Additional context:
{{ .rag_context }}
{{ end }}

Provide enriched documentation that describes the message's purpose and usage. Be concise and accurate.`
}

// getFieldTemplate returns the field enrichment template
func (e *CoTTemplateEngine) getFieldTemplate() string {
	if e.format == "xml" && e.useCoT {
		return `You are a technical documentation expert specializing in API field documentation.

<task>
Enrich the documentation for a Protocol Buffer message field.
</task>

<context>
<field_name>{{ index .context "field_name" }}</field_name>
<field_type>{{ index .context "field_type" }}</field_type>
<message_name>{{ index .context "message_name" }}</message_name>
{{ if index .context "repeated" }}<repeated>true</repeated>{{ end }}
</context>

<current_documentation>
{{ .current_docs }}
</current_documentation>

<instructions>
<thinking>
1. What does this field represent?
2. What format or constraints apply?
3. Is it required or optional?
4. What are common use cases?
</thinking>

<requirements>
- Be specific about the field's purpose
- Include format constraints if applicable
- Mention if field is required/optional/deprecated
- Be concise (1-2 sentences max)
- Do NOT invent constraints without evidence
</requirements>
</instructions>

Provide the enriched field documentation. Output only the documentation text.`
	}

	return `Enrich this field documentation:

Field: {{ index .context "field_name" }} ({{ index .context "field_type" }})
Message: {{ index .context "message_name" }}

Current: {{ .current_docs }}

Provide clear, concise documentation for this field.`
}

// BuildCoTPrompt builds a Chain-of-Thought prompt manually (alternative to templates)
func BuildCoTPrompt(task string, context map[string]interface{}, currentDocs string, ragContext string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`You are a technical documentation expert. Task: %s

<thinking>
`, task))

	// CoT reasoning steps
	sb.WriteString("Step 1: Analyze the current documentation\n")
	sb.WriteString(fmt.Sprintf("Current: %s\n\n", currentDocs))

	sb.WriteString("Step 2: Identify gaps and improvement opportunities\n")
	sb.WriteString("- Is the purpose clear?\n")
	sb.WriteString("- Are there missing details?\n")
	sb.WriteString("- Is the language professional?\n\n")

	if ragContext != "" {
		sb.WriteString("Step 3: Review retrieved context\n")
		sb.WriteString(fmt.Sprintf("%s\n\n", ragContext))
	}

	sb.WriteString("Step 4: Synthesize enriched documentation\n")
	sb.WriteString("- Maintain technical accuracy\n")
	sb.WriteString("- Be concise\n")
	sb.WriteString("- Add value without hallucination\n")
	sb.WriteString("</thinking>\n\n")

	sb.WriteString("<output>\n")
	sb.WriteString("Provide the enriched documentation below:\n")
	sb.WriteString("</output>")

	return sb.String()
}

// NoOpTemplateEngine is a no-op implementation
type NoOpTemplateEngine struct{}

// NewNoOpTemplateEngine creates a no-op template engine
func NewNoOpTemplateEngine() *NoOpTemplateEngine {
	return &NoOpTemplateEngine{}
}

// Render returns a simple prompt
func (e *NoOpTemplateEngine) Render(templateName string, data map[string]interface{}) (string, error) {
	currentDocs, _ := data["current_docs"].(string)
	return fmt.Sprintf("Enrich this documentation: %s", currentDocs), nil
}

// RegisterTemplate does nothing
func (e *NoOpTemplateEngine) RegisterTemplate(name string, template string) error {
	return nil
}

// Package enricher provides advanced AI enrichment strategies.
package enricher

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyivinua/docgen-tool/internal/docgen"
)

// EnrichmentStrategy defines different approaches for enrichment.
type EnrichmentStrategy interface {
	Enrich(ctx context.Context, req *EnrichmentRequest) (*EnrichmentResult, error)
	Name() string
}

// EnrichmentRequest represents an advanced enrichment request.
type EnrichmentRequest struct {
	Type        string                 // service, method, message, field
	Name        string                 // Item name
	Content     string                 // Current content
	Context     *EnrichmentContext     // Rich context information
	Options     *EnrichmentOptions     // Enrichment options
}

// EnrichmentContext provides rich context for AI enrichment.
type EnrichmentContext struct {
	Service         *docgen.Service
	RelatedServices []*docgen.Service
	Method          *docgen.Method
	Message         *docgen.Message
	Field           *docgen.Field
	ProjectDomain   string
	Terminology     map[string]string
	Examples        []string
}

// EnrichmentOptions configures enrichment behavior.
type EnrichmentOptions struct {
	IncludeExamples      bool
	IncludeBestPractices bool
	CheckConsistency     bool
	GenerateCodeSamples  bool
	DetailLevel          string // minimal, standard, detailed, comprehensive
	Tone                 string // technical, conversational, formal
	MaxLength            int
}

// EnrichmentResult contains the enrichment output.
type EnrichmentResult struct {
	EnrichedContent  string
	Examples         []CodeExample
	BestPractices    []BestPractice
	Terminology      map[string]string
	QualityScore     float64
	Confidence       float64
	Suggestions      []string
	InconsistencyWarnings []string
}

// CodeExample represents a generated code example.
type CodeExample struct {
	Language    string
	Code        string
	Description string
}

// BestPractice represents an API design suggestion.
type BestPractice struct {
	Category    string
	Suggestion  string
	Rationale   string
	Priority    string // high, medium, low
}

// ContextAwareStrategy enriches content with full service context.
type ContextAwareStrategy struct {
	client AIClient
}

// NewContextAwareStrategy creates a new context-aware strategy.
func NewContextAwareStrategy(client AIClient) *ContextAwareStrategy {
	return &ContextAwareStrategy{client: client}
}

func (s *ContextAwareStrategy) Name() string {
	return "context-aware"
}

func (s *ContextAwareStrategy) Enrich(ctx context.Context, req *EnrichmentRequest) (*EnrichmentResult, error) {
	prompt := s.buildContextAwarePrompt(req)

	response, err := s.client.Complete(ctx, prompt, &CompletionOptions{
		MaxTokens:   2048,
		Temperature: 0.3,
	})
	if err != nil {
		return nil, fmt.Errorf("AI completion failed: %w", err)
	}

	result := s.parseResponse(response)
	return result, nil
}

func (s *ContextAwareStrategy) buildContextAwarePrompt(req *EnrichmentRequest) string {
	var sb strings.Builder

	sb.WriteString("You are a technical documentation expert specializing in API documentation.\n\n")

	// Add context
	if req.Context != nil {
		if req.Context.Service != nil {
			sb.WriteString(fmt.Sprintf("Service Context: %s (%s)\n",
				req.Context.Service.Name, req.Context.Service.Package))
			sb.WriteString(fmt.Sprintf("Service Description: %s\n\n", req.Context.Service.Description))
		}

		if req.Context.ProjectDomain != "" {
			sb.WriteString(fmt.Sprintf("Project Domain: %s\n\n", req.Context.ProjectDomain))
		}

		if len(req.Context.Terminology) > 0 {
			sb.WriteString("Project Terminology:\n")
			for term, def := range req.Context.Terminology {
				sb.WriteString(fmt.Sprintf("- %s: %s\n", term, def))
			}
			sb.WriteString("\n")
		}
	}

	// Add the main task
	sb.WriteString(fmt.Sprintf("Task: Generate comprehensive documentation for a %s\n", req.Type))
	sb.WriteString(fmt.Sprintf("Name: %s\n", req.Name))

	if req.Content != "" {
		sb.WriteString(fmt.Sprintf("Current Description: %s\n\n", req.Content))
	}

	// Add specific requirements based on type
	switch req.Type {
	case "service":
		sb.WriteString("Generate a description that:\n")
		sb.WriteString("1. Explains the service's primary purpose and responsibility\n")
		sb.WriteString("2. Describes what problems it solves\n")
		sb.WriteString("3. Mentions key features or capabilities\n")
		sb.WriteString("4. Indicates the target use cases\n")
	case "method":
		sb.WriteString("Generate a description that:\n")
		sb.WriteString("1. Explains what the method does in clear terms\n")
		sb.WriteString("2. Describes the expected inputs and outputs\n")
		sb.WriteString("3. Mentions any important behavior or side effects\n")
		sb.WriteString("4. Indicates when to use this method\n")
	case "message":
		sb.WriteString("Generate a description that:\n")
		sb.WriteString("1. Explains what data this message represents\n")
		sb.WriteString("2. Describes its role in the API\n")
		sb.WriteString("3. Mentions key fields and their purpose\n")
	case "field":
		sb.WriteString("Generate a description that:\n")
		sb.WriteString("1. Explains what this field represents\n")
		sb.WriteString("2. Describes valid values or constraints\n")
		sb.WriteString("3. Indicates if it's required or optional\n")
	}

	if req.Options != nil && req.Options.IncludeBestPractices {
		sb.WriteString("\nAlso provide any best practice recommendations.\n")
	}

	sb.WriteString("\nProvide a clear, professional description (2-4 sentences).\n")
	sb.WriteString("Return ONLY the description text, nothing else.")

	return sb.String()
}

func (s *ContextAwareStrategy) parseResponse(response string) *EnrichmentResult {
	return &EnrichmentResult{
		EnrichedContent: strings.TrimSpace(response),
		QualityScore:    0.85,
		Confidence:      0.9,
	}
}

// ExampleGenerationStrategy generates code examples for methods.
type ExampleGenerationStrategy struct {
	client AIClient
}

// NewExampleGenerationStrategy creates a new example generation strategy.
func NewExampleGenerationStrategy(client AIClient) *ExampleGenerationStrategy {
	return &ExampleGenerationStrategy{client: client}
}

func (s *ExampleGenerationStrategy) Name() string {
	return "example-generation"
}

func (s *ExampleGenerationStrategy) Enrich(ctx context.Context, req *EnrichmentRequest) (*EnrichmentResult, error) {
	if req.Type != "method" {
		return &EnrichmentResult{EnrichedContent: req.Content}, nil
	}

	prompt := s.buildExamplePrompt(req)

	response, err := s.client.Complete(ctx, prompt, &CompletionOptions{
		MaxTokens:   1500,
		Temperature: 0.4,
	})
	if err != nil {
		return nil, fmt.Errorf("example generation failed: %w", err)
	}

	result := s.parseExamples(response)
	return result, nil
}

func (s *ExampleGenerationStrategy) buildExamplePrompt(req *EnrichmentRequest) string {
	var sb strings.Builder

	sb.WriteString("Generate code examples for the following API method:\n\n")

	if req.Context != nil && req.Context.Method != nil {
		method := req.Context.Method
		sb.WriteString(fmt.Sprintf("Method: %s\n", method.Name))
		sb.WriteString(fmt.Sprintf("Input: %s\n", method.InputType))
		sb.WriteString(fmt.Sprintf("Output: %s\n", method.OutputType))

		if method.Description != "" {
			sb.WriteString(fmt.Sprintf("Description: %s\n", method.Description))
		}
	}

	sb.WriteString("\nGenerate code examples in the following format:\n")
	sb.WriteString("Language: Go\n")
	sb.WriteString("```go\n")
	sb.WriteString("// Example code here\n")
	sb.WriteString("```\n\n")
	sb.WriteString("Language: Python\n")
	sb.WriteString("```python\n")
	sb.WriteString("# Example code here\n")
	sb.WriteString("```\n")

	return sb.String()
}

func (s *ExampleGenerationStrategy) parseExamples(response string) *EnrichmentResult {
	examples := []CodeExample{}

	// Simple parsing - in production, this would be more sophisticated
	if strings.Contains(response, "```go") {
		examples = append(examples, CodeExample{
			Language:    "Go",
			Code:        extractCodeBlock(response, "go"),
			Description: "Go client example",
		})
	}

	if strings.Contains(response, "```python") {
		examples = append(examples, CodeExample{
			Language:    "Python",
			Code:        extractCodeBlock(response, "python"),
			Description: "Python client example",
		})
	}

	return &EnrichmentResult{
		Examples:   examples,
		Confidence: 0.85,
	}
}

// TerminologyStrategy extracts and maintains consistent terminology.
type TerminologyStrategy struct {
	client       AIClient
	terminology  map[string]string
}

// NewTerminologyStrategy creates a new terminology strategy.
func NewTerminologyStrategy(client AIClient) *TerminologyStrategy {
	return &TerminologyStrategy{
		client:      client,
		terminology: make(map[string]string),
	}
}

func (s *TerminologyStrategy) Name() string {
	return "terminology"
}

func (s *TerminologyStrategy) Enrich(ctx context.Context, req *EnrichmentRequest) (*EnrichmentResult, error) {
	prompt := s.buildTerminologyPrompt(req)

	response, err := s.client.Complete(ctx, prompt, &CompletionOptions{
		MaxTokens:   1024,
		Temperature: 0.2,
	})
	if err != nil {
		return nil, fmt.Errorf("terminology extraction failed: %w", err)
	}

	terminology := s.parseTerminology(response)

	// Check consistency
	warnings := s.checkConsistency(req, terminology)

	return &EnrichmentResult{
		EnrichedContent:       req.Content,
		Terminology:          terminology,
		InconsistencyWarnings: warnings,
		Confidence:           0.9,
	}, nil
}

func (s *TerminologyStrategy) buildTerminologyPrompt(req *EnrichmentRequest) string {
	var sb strings.Builder

	sb.WriteString("Extract key technical terminology from this API description:\n\n")
	sb.WriteString(req.Content)
	sb.WriteString("\n\nFor each technical term, provide a brief definition.\n")
	sb.WriteString("Format: TERM: definition\n")

	return sb.String()
}

func (s *TerminologyStrategy) parseTerminology(response string) map[string]string {
	terminology := make(map[string]string)

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		if parts := strings.SplitN(line, ":", 2); len(parts) == 2 {
			term := strings.TrimSpace(parts[0])
			def := strings.TrimSpace(parts[1])
			if term != "" && def != "" {
				terminology[term] = def
			}
		}
	}

	return terminology
}

func (s *TerminologyStrategy) checkConsistency(req *EnrichmentRequest, newTerms map[string]string) []string {
	var warnings []string

	for term, newDef := range newTerms {
		if existingDef, exists := s.terminology[term]; exists {
			if existingDef != newDef {
				warnings = append(warnings, fmt.Sprintf(
					"Inconsistent definition for '%s': '%s' vs '%s'",
					term, existingDef, newDef))
			}
		} else {
			s.terminology[term] = newDef
		}
	}

	return warnings
}

// BestPracticeStrategy analyzes and suggests API design improvements.
type BestPracticeStrategy struct {
	client AIClient
}

// NewBestPracticeStrategy creates a new best practice strategy.
func NewBestPracticeStrategy(client AIClient) *BestPracticeStrategy {
	return &BestPracticeStrategy{client: client}
}

func (s *BestPracticeStrategy) Name() string {
	return "best-practice"
}

func (s *BestPracticeStrategy) Enrich(ctx context.Context, req *EnrichmentRequest) (*EnrichmentResult, error) {
	prompt := s.buildBestPracticePrompt(req)

	response, err := s.client.Complete(ctx, prompt, &CompletionOptions{
		MaxTokens:   1024,
		Temperature: 0.3,
	})
	if err != nil {
		return nil, fmt.Errorf("best practice analysis failed: %w", err)
	}

	practices := s.parseBestPractices(response)

	return &EnrichmentResult{
		EnrichedContent: req.Content,
		BestPractices:   practices,
		Confidence:      0.8,
	}, nil
}

func (s *BestPracticeStrategy) buildBestPracticePrompt(req *EnrichmentRequest) string {
	var sb strings.Builder

	sb.WriteString("Analyze this API design and provide best practice recommendations:\n\n")

	if req.Context != nil {
		if req.Context.Service != nil {
			sb.WriteString(fmt.Sprintf("Service: %s\n", req.Context.Service.Name))
		}
		if req.Context.Method != nil {
			method := req.Context.Method
			sb.WriteString(fmt.Sprintf("Method: %s\n", method.Name))
			sb.WriteString(fmt.Sprintf("Input: %s, Output: %s\n", method.InputType, method.OutputType))
		}
	}

	sb.WriteString("\nFocus on:\n")
	sb.WriteString("1. Naming conventions\n")
	sb.WriteString("2. API design patterns\n")
	sb.WriteString("3. Error handling\n")
	sb.WriteString("4. Performance considerations\n")
	sb.WriteString("5. Security best practices\n")
	sb.WriteString("\nProvide actionable recommendations with rationale.\n")

	return sb.String()
}

func (s *BestPracticeStrategy) parseBestPractices(response string) []BestPractice {
	// Simple parsing - production would be more sophisticated
	practices := []BestPractice{}

	// Split by common patterns
	sections := strings.Split(response, "\n\n")
	for _, section := range sections {
		if strings.TrimSpace(section) != "" {
			practices = append(practices, BestPractice{
				Category:   "General",
				Suggestion: strings.TrimSpace(section),
				Rationale:  "AI-suggested best practice",
				Priority:   "medium",
			})
		}
	}

	return practices
}

// MultiPassStrategy performs iterative refinement.
type MultiPassStrategy struct {
	client      AIClient
	maxPasses   int
}

// NewMultiPassStrategy creates a new multi-pass strategy.
func NewMultiPassStrategy(client AIClient, maxPasses int) *MultiPassStrategy {
	return &MultiPassStrategy{
		client:    client,
		maxPasses: maxPasses,
	}
}

func (s *MultiPassStrategy) Name() string {
	return "multi-pass"
}

func (s *MultiPassStrategy) Enrich(ctx context.Context, req *EnrichmentRequest) (*EnrichmentResult, error) {
	currentContent := req.Content
	var qualityScore float64

	for i := 0; i < s.maxPasses; i++ {
		prompt := s.buildRefinementPrompt(currentContent, i)

		response, err := s.client.Complete(ctx, prompt, &CompletionOptions{
			MaxTokens:   1024,
			Temperature: 0.3,
		})
		if err != nil {
			return nil, fmt.Errorf("refinement pass %d failed: %w", i, err)
		}

		// Score quality
		newScore := s.scoreQuality(response)

		// Stop if quality isn't improving
		if i > 0 && newScore <= qualityScore {
			break
		}

		currentContent = response
		qualityScore = newScore
	}

	return &EnrichmentResult{
		EnrichedContent: currentContent,
		QualityScore:    qualityScore,
		Confidence:      0.9,
	}, nil
}

func (s *MultiPassStrategy) buildRefinementPrompt(content string, pass int) string {
	var sb strings.Builder

	if pass == 0 {
		sb.WriteString("Improve the following technical documentation:\n\n")
	} else {
		sb.WriteString("Further refine this technical documentation for clarity and completeness:\n\n")
	}

	sb.WriteString(content)
	sb.WriteString("\n\nMake it more:\n")
	sb.WriteString("- Clear and concise\n")
	sb.WriteString("- Technically accurate\n")
	sb.WriteString("- Easy to understand\n")
	sb.WriteString("- Complete and informative\n")

	return sb.String()
}

func (s *MultiPassStrategy) scoreQuality(content string) float64 {
	score := 50.0

	// Length scoring
	words := len(strings.Fields(content))
	if words >= 20 {
		score += 20
	} else if words >= 10 {
		score += 10
	}

	// Sentence structure
	sentences := strings.Count(content, ".") + strings.Count(content, "!") + strings.Count(content, "?")
	if sentences >= 3 {
		score += 15
	} else if sentences >= 2 {
		score += 10
	}

	// Technical terms (simple heuristic)
	if strings.Contains(content, "API") || strings.Contains(content, "service") {
		score += 10
	}

	// Examples or details
	if strings.Contains(content, "example") || strings.Contains(content, "e.g.") {
		score += 5
	}

	return min(score, 100.0)
}

// Helper functions

func extractCodeBlock(text, language string) string {
	pattern := "```" + language
	start := strings.Index(text, pattern)
	if start == -1 {
		return ""
	}

	start += len(pattern)
	end := strings.Index(text[start:], "```")
	if end == -1 {
		return ""
	}

	return strings.TrimSpace(text[start : start+end])
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

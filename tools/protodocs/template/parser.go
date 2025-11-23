package template

import (
	"fmt"
	"strings"
)

// Parser parses template strings into AST
type Parser struct {
	input    string
	position int
	errors   []error
}

// NewParser creates a new template parser
func NewParser(input string) *Parser {
	return &Parser{
		input:    input,
		position: 0,
		errors:   []error{},
	}
}

// Parse parses the template and returns an AST
func (p *Parser) Parse(name string) (*Template, error) {
	template := NewTemplate(name)

	for p.position < len(p.input) {
		node, err := p.parseNext()
		if err != nil {
			p.errors = append(p.errors, err)
			// Try to recover by skipping to next delimiter
			p.skipToNextDelimiter()
			continue
		}

		if node != nil {
			template.AddNode(node)
		}
	}

	if len(p.errors) > 0 {
		return template, fmt.Errorf("parsing errors: %v", p.errors)
	}

	return template, nil
}

// parseNext parses the next node in the template
func (p *Parser) parseNext() (Node, error) {
	// Check for different delimiter types
	if p.peek("{{") {
		return p.parseVariable()
	}

	if p.peek("{%") {
		return p.parseDirective()
	}

	if p.peek("{#") {
		return p.parseComment()
	}

	// Default: parse text until next delimiter
	return p.parseText()
}

// parseText parses static text content
func (p *Parser) parseText() (Node, error) {
	start := p.position

	// Find next delimiter
	for p.position < len(p.input) {
		if p.peek("{{") || p.peek("{%") || p.peek("{#") {
			break
		}
		p.position++
	}

	if start == p.position {
		return nil, nil
	}

	content := p.input[start:p.position]
	return NewTextNode(content), nil
}

// parseVariable parses {{ variable }}
func (p *Parser) parseVariable() (Node, error) {
	if !p.consume("{{") {
		return nil, fmt.Errorf("expected {{")
	}

	// Skip whitespace
	p.skipWhitespace()

	// Extract variable name and filters
	start := p.position
	for p.position < len(p.input) && !p.peek("}}") && !p.peek("|") {
		p.position++
	}

	if start == p.position {
		return nil, fmt.Errorf("empty variable name")
	}

	varName := strings.TrimSpace(p.input[start:p.position])
	node := NewVariableNode(varName)

	// Parse filters
	for p.peek("|") {
		p.consume("|")
		p.skipWhitespace()

		filterStart := p.position
		for p.position < len(p.input) && !p.peek("}}") && !p.peek("|") {
			p.position++
		}

		filter := strings.TrimSpace(p.input[filterStart:p.position])
		node.Filters = append(node.Filters, filter)
	}

	p.skipWhitespace()
	if !p.consume("}}") {
		return nil, fmt.Errorf("expected }}")
	}

	return node, nil
}

// parseDirective parses {% directive %}
func (p *Parser) parseDirective() (Node, error) {
	if !p.consume("{%") {
		return nil, fmt.Errorf("expected {%%")
	}

	p.skipWhitespace()

	// Extract directive name
	directiveName := p.parseIdentifier()

	switch directiveName {
	case "if":
		return p.parseConditional()
	case "for":
		return p.parseLoop()
	case "include":
		return p.parseInclude()
	case "enrich":
		return p.parseEnrich()
	case "section":
		return p.parseSection()
	case "endif", "endfor", "endsection":
		// These are handled by their respective parsing functions
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown directive: %s", directiveName)
	}
}

// parseConditional parses {% if condition %} ... {% endif %}
func (p *Parser) parseConditional() (Node, error) {
	p.skipWhitespace()

	// Parse condition
	conditionStart := p.position
	for p.position < len(p.input) && !p.peek("%}") {
		p.position++
	}

	condition := strings.TrimSpace(p.input[conditionStart:p.position])
	if !p.consume("%}") {
		return nil, fmt.Errorf("expected %%}")
	}

	node := NewConditionalNode(condition)

	// Parse then branch
	for {
		if p.peek("{% else") {
			break
		}
		if p.peek("{% endif") {
			break
		}
		if p.position >= len(p.input) {
			return nil, fmt.Errorf("unclosed if statement")
		}

		child, err := p.parseNext()
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.ThenBranch = append(node.ThenBranch, child)
		}
	}

	// Parse else branch if present
	if p.peek("{% else") {
		p.consume("{% else")
		p.skipWhitespace()
		p.consume("%}")

		for {
			if p.peek("{% endif") {
				break
			}
			if p.position >= len(p.input) {
				return nil, fmt.Errorf("unclosed else statement")
			}

			child, err := p.parseNext()
			if err != nil {
				return nil, err
			}
			if child != nil {
				node.ElseBranch = append(node.ElseBranch, child)
			}
		}
	}

	// Consume endif
	if !p.consume("{% endif") {
		return nil, fmt.Errorf("expected {%% endif")
	}
	p.skipWhitespace()
	p.consume("%}")

	return node, nil
}

// parseLoop parses {% for item in items %} ... {% endfor %}
func (p *Parser) parseLoop() (Node, error) {
	p.skipWhitespace()

	// Parse variable name
	variable := p.parseIdentifier()
	p.skipWhitespace()

	if !p.consume("in") {
		return nil, fmt.Errorf("expected 'in' keyword")
	}
	p.skipWhitespace()

	// Parse collection name
	collectionStart := p.position
	for p.position < len(p.input) && !p.peek("%}") {
		p.position++
	}

	collection := strings.TrimSpace(p.input[collectionStart:p.position])
	if !p.consume("%}") {
		return nil, fmt.Errorf("expected %%}")
	}

	node := NewLoopNode(variable, collection)

	// Parse loop body
	for {
		if p.peek("{% endfor") {
			break
		}
		if p.position >= len(p.input) {
			return nil, fmt.Errorf("unclosed for loop")
		}

		child, err := p.parseNext()
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.Body = append(node.Body, child)
		}
	}

	// Consume endfor
	if !p.consume("{% endfor") {
		return nil, fmt.Errorf("expected {%% endfor")
	}
	p.skipWhitespace()
	p.consume("%}")

	return node, nil
}

// parseInclude parses {% include "template.md" %}
func (p *Parser) parseInclude() (Node, error) {
	p.skipWhitespace()

	// Parse template name (quoted string)
	if !p.peek("\"") && !p.peek("'") {
		return nil, fmt.Errorf("expected quoted template name")
	}

	quote := p.input[p.position]
	p.position++ // Skip opening quote

	templateStart := p.position
	for p.position < len(p.input) && p.input[p.position] != quote {
		p.position++
	}

	templateName := p.input[templateStart:p.position]
	p.position++ // Skip closing quote

	p.skipWhitespace()
	if !p.consume("%}") {
		return nil, fmt.Errorf("expected %%}")
	}

	return NewIncludeNode(templateName), nil
}

// parseEnrich parses {% enrich type="description" prompt="..." %}
func (p *Parser) parseEnrich() (Node, error) {
	p.skipWhitespace()

	// Parse attributes
	attributes := make(map[string]string)
	for !p.peek("%}") && p.position < len(p.input) {
		// Parse attribute name
		attrName := p.parseIdentifier()
		p.skipWhitespace()

		if !p.consume("=") {
			return nil, fmt.Errorf("expected = after attribute name")
		}
		p.skipWhitespace()

		// Parse attribute value
		if !p.peek("\"") && !p.peek("'") {
			return nil, fmt.Errorf("expected quoted attribute value")
		}

		quote := p.input[p.position]
		p.position++

		valueStart := p.position
		for p.position < len(p.input) && p.input[p.position] != quote {
			p.position++
		}

		attrValue := p.input[valueStart:p.position]
		p.position++

		attributes[attrName] = attrValue
		p.skipWhitespace()
	}

	if !p.consume("%}") {
		return nil, fmt.Errorf("expected %%}")
	}

	enrichType := attributes["type"]
	if enrichType == "" {
		enrichType = "description"
	}

	node := NewEnrichNode(enrichType)
	node.Prompt = attributes["prompt"]

	// Parse max_tokens if present
	if maxTokensStr := attributes["max_tokens"]; maxTokensStr != "" {
		fmt.Sscanf(maxTokensStr, "%d", &node.MaxTokens)
	}

	// Parse cache setting
	if cacheStr := attributes["cache"]; cacheStr != "" {
		node.Cache = cacheStr == "true"
	}

	// Add remaining attributes to context
	for key, value := range attributes {
		if key != "type" && key != "prompt" && key != "max_tokens" && key != "cache" {
			node.Context[key] = value
		}
	}

	return node, nil
}

// parseSection parses {% section "name" %} ... {% endsection %}
func (p *Parser) parseSection() (Node, error) {
	p.skipWhitespace()

	// Parse section name
	if !p.peek("\"") && !p.peek("'") {
		return nil, fmt.Errorf("expected quoted section name")
	}

	quote := p.input[p.position]
	p.position++

	nameStart := p.position
	for p.position < len(p.input) && p.input[p.position] != quote {
		p.position++
	}

	sectionName := p.input[nameStart:p.position]
	p.position++

	p.skipWhitespace()
	if !p.consume("%}") {
		return nil, fmt.Errorf("expected %%}")
	}

	node := NewSectionNode(sectionName)

	// Parse section body
	for {
		if p.peek("{% endsection") {
			break
		}
		if p.position >= len(p.input) {
			return nil, fmt.Errorf("unclosed section")
		}

		child, err := p.parseNext()
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.Body = append(node.Body, child)
		}
	}

	// Consume endsection
	if !p.consume("{% endsection") {
		return nil, fmt.Errorf("expected {%% endsection")
	}
	p.skipWhitespace()
	p.consume("%}")

	return node, nil
}

// parseComment parses {# comment #}
func (p *Parser) parseComment() (Node, error) {
	if !p.consume("{#") {
		return nil, fmt.Errorf("expected {#")
	}

	start := p.position
	for p.position < len(p.input) && !p.peek("#}") {
		p.position++
	}

	comment := p.input[start:p.position]

	if !p.consume("#}") {
		return nil, fmt.Errorf("expected #}")
	}

	return NewCommentNode(comment), nil
}

// Helper methods

func (p *Parser) peek(s string) bool {
	if p.position+len(s) > len(p.input) {
		return false
	}
	return p.input[p.position:p.position+len(s)] == s
}

func (p *Parser) consume(s string) bool {
	if p.peek(s) {
		p.position += len(s)
		return true
	}
	return false
}

func (p *Parser) skipWhitespace() {
	for p.position < len(p.input) && isWhitespace(p.input[p.position]) {
		p.position++
	}
}

func (p *Parser) skipToNextDelimiter() {
	for p.position < len(p.input) {
		if p.peek("{{") || p.peek("{%") || p.peek("{#") {
			return
		}
		p.position++
	}
}

func (p *Parser) parseIdentifier() string {
	start := p.position
	for p.position < len(p.input) && isIdentifierChar(p.input[p.position]) {
		p.position++
	}
	return p.input[start:p.position]
}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func isIdentifierChar(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '_'
}

// ParseString is a convenience function for parsing a template string
func ParseString(name, input string) (*Template, error) {
	parser := NewParser(input)
	return parser.Parse(name)
}

// Validate validates a parsed template
func Validate(template *Template) error {
	errors := []error{}

	err := template.Walk(func(node Node) error {
		switch n := node.(type) {
		case *VariableNode:
			if n.Name == "" {
				errors = append(errors, fmt.Errorf("empty variable name"))
			}
		case *ConditionalNode:
			if n.Condition == "" {
				errors = append(errors, fmt.Errorf("empty conditional"))
			}
		case *LoopNode:
			if n.Variable == "" || n.Collection == "" {
				errors = append(errors, fmt.Errorf("invalid loop: variable=%s, collection=%s",
					n.Variable, n.Collection))
			}
		case *EnrichNode:
			if n.EnrichType == "" {
				errors = append(errors, fmt.Errorf("empty enrich type"))
			}
		}
		return nil
	})

	if err != nil {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %v", errors)
	}

	return nil
}

// OptimizeTemplate performs AST optimizations
func OptimizeTemplate(template *Template) *Template {
	// Merge consecutive text nodes
	optimized := NewTemplate(template.Name)
	optimized.Metadata = template.Metadata

	var pendingText strings.Builder

	for _, node := range template.Root {
		if textNode, ok := node.(*TextNode); ok {
			pendingText.WriteString(textNode.Content)
		} else {
			if pendingText.Len() > 0 {
				optimized.AddNode(NewTextNode(pendingText.String()))
				pendingText.Reset()
			}
			optimized.AddNode(node)
		}
	}

	if pendingText.Len() > 0 {
		optimized.AddNode(NewTextNode(pendingText.String()))
	}

	return optimized
}

// ExtractVariables extracts all variable names from the template
func ExtractVariables(template *Template) []string {
	variables := make(map[string]bool)

	template.Walk(func(node Node) error {
		if varNode, ok := node.(*VariableNode); ok {
			variables[varNode.Name] = true
		}
		return nil
	})

	result := make([]string, 0, len(variables))
	for v := range variables {
		result = append(result, v)
	}
	return result
}

// ExtractEnrichNodes extracts all enrichment nodes for processing
func ExtractEnrichNodes(template *Template) []*EnrichNode {
	enrichNodes := []*EnrichNode{}

	template.Walk(func(node Node) error {
		if enrichNode, ok := node.(*EnrichNode); ok {
			enrichNodes = append(enrichNodes, enrichNode)
		}
		return nil
	})

	return enrichNodes
}

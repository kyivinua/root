package template

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Renderer renders templates to strings
type Renderer struct {
	templates      map[string]*Template
	enricher       *Enricher
	filters        map[string]FilterFunc
	globals        map[string]interface{}
	enableEnrich   bool
}

// FilterFunc is a function that transforms a value
type FilterFunc func(value interface{}, args ...string) (interface{}, error)

// RenderContext holds the rendering context
type RenderContext struct {
	Variables        map[string]interface{}
	EnrichmentResults map[string]string  // chunk ID -> enriched content
	Parent           *RenderContext
}

// NewRenderer creates a new template renderer
func NewRenderer() *Renderer {
	r := &Renderer{
		templates:    make(map[string]*Template),
		filters:      make(map[string]FilterFunc),
		globals:      make(map[string]interface{}),
		enableEnrich: true,
	}

	// Register default filters
	r.RegisterDefaultFilters()

	return r
}

// SetEnricher sets the LLM enricher
func (r *Renderer) SetEnricher(enricher *Enricher) {
	r.enricher = enricher
}

// AddTemplate adds a template to the renderer
func (r *Renderer) AddTemplate(name string, template *Template) {
	r.templates[name] = template
}

// RegisterFilter registers a custom filter
func (r *Renderer) RegisterFilter(name string, fn FilterFunc) {
	r.filters[name] = fn
}

// RegisterDefaultFilters registers built-in filters
func (r *Renderer) RegisterDefaultFilters() {
	// String filters
	r.RegisterFilter("upper", func(value interface{}, args ...string) (interface{}, error) {
		return strings.ToUpper(fmt.Sprintf("%v", value)), nil
	})

	r.RegisterFilter("lower", func(value interface{}, args ...string) (interface{}, error) {
		return strings.ToLower(fmt.Sprintf("%v", value)), nil
	})

	r.RegisterFilter("title", func(value interface{}, args ...string) (interface{}, error) {
		return cases.Title(language.English).String(strings.ToLower(fmt.Sprintf("%v", value))), nil
	})

	r.RegisterFilter("trim", func(value interface{}, args ...string) (interface{}, error) {
		return strings.TrimSpace(fmt.Sprintf("%v", value)), nil
	})

	r.RegisterFilter("replace", func(value interface{}, args ...string) (interface{}, error) {
		if len(args) < 2 {
			return value, fmt.Errorf("replace filter requires 2 arguments")
		}
		str := fmt.Sprintf("%v", value)
		return strings.ReplaceAll(str, args[0], args[1]), nil
	})

	// Formatting filters
	r.RegisterFilter("markdown_code", func(value interface{}, args ...string) (interface{}, error) {
		lang := "text"
		if len(args) > 0 {
			lang = args[0]
		}
		return fmt.Sprintf("```%s\n%v\n```", lang, value), nil
	})

	r.RegisterFilter("escape_md", func(value interface{}, args ...string) (interface{}, error) {
		str := fmt.Sprintf("%v", value)
		replacer := strings.NewReplacer(
			"*", "\\*",
			"_", "\\_",
			"`", "\\`",
			"[", "\\[",
			"]", "\\]",
		)
		return replacer.Replace(str), nil
	})

	// Default value filter
	r.RegisterFilter("default", func(value interface{}, args ...string) (interface{}, error) {
		if value == nil || value == "" {
			if len(args) > 0 {
				return args[0], nil
			}
			return "", nil
		}
		return value, nil
	})
}

// Render renders a template with the given context
func (r *Renderer) Render(ctx context.Context, templateName string, variables map[string]interface{}) (string, error) {
	template, ok := r.templates[templateName]
	if !ok {
		return "", fmt.Errorf("template not found: %s", templateName)
	}

	// Create render context
	renderCtx := &RenderContext{
		Variables:         variables,
		EnrichmentResults: make(map[string]string),
	}

	// Perform enrichment if enabled
	if r.enableEnrich && r.enricher != nil {
		chunker := NewChunker()
		graph, err := chunker.ChunkTemplate(template)
		if err != nil {
			return "", fmt.Errorf("chunking failed: %w", err)
		}

		results, err := r.enricher.EnrichChunks(ctx, graph)
		if err != nil {
			// Log error but continue rendering
			fmt.Printf("Enrichment warning: %v\n", err)
		}

		// Store enrichment results
		for _, result := range results {
			if result.Error == nil && result.Enriched != "" {
				renderCtx.EnrichmentResults[result.ChunkID] = result.Enriched
			}
		}
	}

	// Render template
	var output strings.Builder
	if err := r.renderNodes(template.Root, renderCtx, &output); err != nil {
		return "", err
	}

	return output.String(), nil
}

// renderNodes renders a list of nodes
func (r *Renderer) renderNodes(nodes []Node, ctx *RenderContext, output *strings.Builder) error {
	for _, node := range nodes {
		if err := r.renderNode(node, ctx, output); err != nil {
			return err
		}
	}
	return nil
}

// renderNode renders a single node
func (r *Renderer) renderNode(node Node, ctx *RenderContext, output *strings.Builder) error {
	switch n := node.(type) {
	case *TextNode:
		output.WriteString(n.Content)

	case *VariableNode:
		value, err := r.resolveVariable(n.Name, ctx)
		if err != nil {
			return err
		}

		// Apply filters
		for _, filterStr := range n.Filters {
			// Parse filter name and arguments (space-separated or quoted)
			parts := strings.Split(filterStr, " ")
			filterName := strings.TrimSpace(parts[0])
			args := []string{}

			// Clean up arguments (remove quotes)
			for _, part := range parts[1:] {
				cleaned := strings.Trim(strings.TrimSpace(part), "\"'")
				if cleaned != "" {
					args = append(args, cleaned)
				}
			}

			filter, ok := r.filters[filterName]
			if !ok {
				return fmt.Errorf("unknown filter: %s", filterName)
			}

			value, err = filter(value, args...)
			if err != nil {
				return fmt.Errorf("filter %s failed: %w", filterName, err)
			}
		}

		// Use default value if variable is empty
		if value == nil || value == "" {
			value = n.DefaultVal
		}

		output.WriteString(fmt.Sprintf("%v", value))

	case *ConditionalNode:
		condition, err := r.evaluateCondition(n.Condition, ctx)
		if err != nil {
			return err
		}

		if condition {
			if err := r.renderNodes(n.ThenBranch, ctx, output); err != nil {
				return err
			}
		} else {
			if err := r.renderNodes(n.ElseBranch, ctx, output); err != nil {
				return err
			}
		}

	case *LoopNode:
		collection, err := r.resolveVariable(n.Collection, ctx)
		if err != nil {
			return err
		}

		// Convert to slice
		items, err := r.toSlice(collection)
		if err != nil {
			return err
		}

		// Render loop body for each item
		for i, item := range items {
			// Create new context with loop variable
			loopCtx := &RenderContext{
				Variables: make(map[string]interface{}),
				Parent:    ctx,
			}

			// Copy parent variables
			for k, v := range ctx.Variables {
				loopCtx.Variables[k] = v
			}

			// Set loop variable and special variables
			loopCtx.Variables[n.Variable] = item
			loopCtx.Variables["loop"] = map[string]interface{}{
				"index":  i,
				"index1": i + 1,
				"first":  i == 0,
				"last":   i == len(items)-1,
				"length": len(items),
			}

			if err := r.renderNodes(n.Body, loopCtx, output); err != nil {
				return err
			}
		}

	case *IncludeNode:
		// Render included template
		includedTemplate, ok := r.templates[n.TemplateName]
		if !ok {
			return fmt.Errorf("included template not found: %s", n.TemplateName)
		}

		// Create new context with include context
		includeCtx := &RenderContext{
			Variables: make(map[string]interface{}),
			Parent:    ctx,
		}

		// Copy parent variables
		for k, v := range ctx.Variables {
			includeCtx.Variables[k] = v
		}

		// Override with include context
		for k, v := range n.Context {
			includeCtx.Variables[k] = v
		}

		if err := r.renderNodes(includedTemplate.Root, includeCtx, output); err != nil {
			return err
		}

	case *EnrichNode:
		// Check if we have enriched content
		// In a real implementation, we would have a chunk ID to lookup
		// For now, render a placeholder
		if n.Prompt != "" {
			output.WriteString(fmt.Sprintf("[Enriched: %s]", n.EnrichType))
		}

	case *SectionNode:
		// Render section body
		if err := r.renderNodes(n.Body, ctx, output); err != nil {
			return err
		}

	case *CommentNode:
		// Comments don't render

	default:
		return fmt.Errorf("unknown node type: %T", node)
	}

	return nil
}

// resolveVariable resolves a variable from context, supporting dot notation
func (r *Renderer) resolveVariable(name string, ctx *RenderContext) (interface{}, error) {
	// Handle dot notation (e.g., "service.name" or "loop.index1")
	if strings.Contains(name, ".") {
		parts := strings.SplitN(name, ".", 2)
		base, err := r.resolveVariable(parts[0], ctx)
		if err != nil || base == nil {
			return nil, err
		}

		// Access nested field
		return r.accessField(base, parts[1])
	}

	// Check current context
	if value, ok := ctx.Variables[name]; ok {
		return value, nil
	}

	// Check parent context
	if ctx.Parent != nil {
		return r.resolveVariable(name, ctx.Parent)
	}

	// Check globals
	if value, ok := r.globals[name]; ok {
		return value, nil
	}

	return nil, nil // Return nil instead of error for missing variables
}

// accessField accesses a field from a map or struct
func (r *Renderer) accessField(obj interface{}, field string) (interface{}, error) {
	if obj == nil {
		return nil, nil
	}

	// Handle maps
	if m, ok := obj.(map[string]interface{}); ok {
		return m[field], nil
	}

	if m, ok := obj.(map[string]string); ok {
		return m[field], nil
	}

	// Use reflection for other types
	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Map {
		key := reflect.ValueOf(field)
		val := rv.MapIndex(key)
		if val.IsValid() {
			return val.Interface(), nil
		}
	}

	return nil, nil
}

// evaluateCondition evaluates a conditional expression
func (r *Renderer) evaluateCondition(condition string, ctx *RenderContext) (bool, error) {
	// Parse simple conditions
	condition = strings.TrimSpace(condition)

	// Handle negation
	if strings.HasPrefix(condition, "not ") {
		result, err := r.evaluateCondition(condition[4:], ctx)
		return !result, err
	}

	// Handle equality
	if strings.Contains(condition, "==") {
		parts := strings.SplitN(condition, "==", 2)
		left, _ := r.resolveVariable(strings.TrimSpace(parts[0]), ctx)
		right := strings.TrimSpace(parts[1])
		return fmt.Sprintf("%v", left) == strings.Trim(right, "\"'"), nil
	}

	// Handle inequality
	if strings.Contains(condition, "!=") {
		parts := strings.SplitN(condition, "!=", 2)
		left, _ := r.resolveVariable(strings.TrimSpace(parts[0]), ctx)
		right := strings.TrimSpace(parts[1])
		return fmt.Sprintf("%v", left) != strings.Trim(right, "\"'"), nil
	}

	// Simple boolean evaluation
	value, _ := r.resolveVariable(condition, ctx)
	return r.toBool(value), nil
}

// toBool converts a value to boolean
func (r *Renderer) toBool(value interface{}) bool {
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v != ""
	case int, int64, int32:
		return v != 0
	case float64, float32:
		return v != 0
	default:
		// Use reflection for slices/maps
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Slice, reflect.Map, reflect.Array:
			return rv.Len() > 0
		default:
			return true
		}
	}
}

// toSlice converts a value to a slice
func (r *Renderer) toSlice(value interface{}) ([]interface{}, error) {
	if value == nil {
		return []interface{}{}, nil
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		result := make([]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			result[i] = rv.Index(i).Interface()
		}
		return result, nil

	case reflect.Map:
		result := make([]interface{}, 0, rv.Len())
		for _, key := range rv.MapKeys() {
			result = append(result, map[string]interface{}{
				"key":   key.Interface(),
				"value": rv.MapIndex(key).Interface(),
			})
		}
		return result, nil

	default:
		return nil, fmt.Errorf("cannot iterate over %T", value)
	}
}

// SetGlobal sets a global variable
func (r *Renderer) SetGlobal(name string, value interface{}) {
	r.globals[name] = value
}

// EnableEnrichment enables/disables LLM enrichment
func (r *Renderer) EnableEnrichment(enable bool) {
	r.enableEnrich = enable
}

// RenderString is a convenience function for rendering a template string
func RenderString(templateStr string, variables map[string]interface{}) (string, error) {
	template, err := ParseString("inline", templateStr)
	if err != nil {
		return "", err
	}

	renderer := NewRenderer()
	renderer.AddTemplate("inline", template)

	return renderer.Render(context.Background(), "inline", variables)
}

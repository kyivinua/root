package template

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestBasicParsing tests basic template parsing
func TestBasicParsing(t *testing.T) {
	template := `
Hello {{name}}!

{% if greeting %}
{{greeting | upper}}
{% endif %}

{% for item in items %}
- {{item}}
{% endfor %}
`

	tmpl, err := ParseString("test", template)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(tmpl.Root) == 0 {
		t.Fatal("Expected parsed nodes")
	}

	t.Logf("Parsed template:\n%s", tmpl.String())
}

// TestRendering tests template rendering
func TestRendering(t *testing.T) {
	template := `# {{title | upper}}

Hello {{name | default "World"}}!

{% if show_items %}
## Items

{% for item in items %}
{{loop.index1}}. {{item}}
{% endfor %}
{% endif %}
`

	tmpl, err := ParseString("test", template)
	if err != nil{
		t.Fatalf("Parse failed: %v", err)
	}

	renderer := NewRenderer()
	renderer.AddTemplate("test", tmpl)

	variables := map[string]interface{}{
		"title":      "example",
		"name":       "User",
		"show_items": true,
		"items":      []string{"Apple", "Banana", "Cherry"},
	}

	result, err := renderer.Render(context.Background(), "test", variables)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	t.Logf("Rendered output:\n%s", result)

	if !strings.Contains(result, "EXAMPLE") {
		t.Error("Expected uppercase title")
	}

	if !strings.Contains(result, "1. Apple") {
		t.Error("Expected numbered list")
	}
}

// TestEnrichment tests LLM enrichment
func TestEnrichment(t *testing.T) {
	template := `# API Documentation

{% section "overview" %}
## Overview

{% enrich type="description" entity_type="API" entity_name="UserService" %}

{% endsection %}

## Methods

{% for method in methods %}
### {{method.name}}

{% enrich type="description" method="{{method.name}}" %}

{% endfor %}
`

	tmpl, err := ParseString("test", template)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Create mock LLM provider
	mockProvider := NewMockLLMProvider()
	mockProvider.SetResponse("description", "This is a comprehensive API service for managing users.")
	mockProvider.SetDelay(10 * time.Millisecond)

	// Create enricher
	config := DefaultEnricherConfig()
	config.MaxConcurrency = 2
	enricher := NewEnricher(mockProvider, config)

	// Create chunker and chunk template
	chunker := NewChunker()
	graph, err := chunker.ChunkTemplate(tmpl)
	if err != nil {
		t.Fatalf("Chunking failed: %v", err)
	}

	t.Logf("Chunk stats: %+v", graph.Stats())

	// Enrich chunks
	ctx := context.Background()
	results, err := enricher.EnrichChunks(ctx, graph)
	if err != nil {
		t.Logf("Enrichment warning: %v", err)
	}

	for _, result := range results {
		if result.Error == nil {
			t.Logf("Enriched chunk %s (%s): %s", result.ChunkID, result.EnrichType, result.Enriched)
		}
	}
}

// TestChunking tests template chunking
func TestChunking(t *testing.T) {
	template := `# Documentation

{% section "intro" %}
This is a long introduction paragraph that should be broken into chunks
for processing. It contains multiple sentences and provides context for
the entire document.
{% endsection %}

{% for item in items %}
- Item: {{item.name}}
  Description: {{item.desc}}
{% endfor %}

{% enrich type="summary" %}
`

	tmpl, err := ParseString("test", template)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	chunker := NewChunker()
	graph, err := chunker.ChunkTemplate(tmpl)
	if err != nil {
		t.Fatalf("Chunking failed: %v", err)
	}

	stats := graph.Stats()
	t.Logf("Total chunks: %d", stats["total_chunks"])
	t.Logf("Enrichable chunks: %d", stats["enrichable_chunks"])
	t.Logf("Independent chunks: %d", stats["independent_chunks"])

	// Test topological sort
	sorted, err := graph.TopologicalSort()
	if err != nil {
		t.Fatalf("Topological sort failed: %v", err)
	}

	t.Logf("Sorted chunks: %d", len(sorted))

	// Test parallelizable groups
	groups := graph.GetParallelizableGroups()
	t.Logf("Parallelizable groups: %d", len(groups))
	for i, group := range groups {
		t.Logf("  Group %d: %d chunks", i, len(group))
	}
}

// TestComplexTemplate tests a complex template with multiple features
func TestComplexTemplate(t *testing.T) {
	// Use a simpler template without backticks in raw string
	template := `# {{service.name | upper}} API Documentation

**Package**: {{service.package}}
**Version**: {{service.version}}

---

## Overview

{% section "overview" %}
{% enrich type="description" entity_type="Service" entity_name="{{service.name}}" %}
{% endsection %}

## Authentication

{% if service.requires_auth %}
This service requires authentication using:

{% for method in auth_methods %}
- **{{method.name}}**: {{method.description}}
{% endfor %}
{% else %}
No authentication required.
{% endif %}

## Methods

{% for method in methods %}
### {{method.name}}

{% enrich type="description" method="{{method.name}}" service="{{service.name}}" %}

**Input**: {{method.input}}
**Output**: {{method.output}}

{% if method.streaming %}
This is a streaming RPC.
{% endif %}

---

{% endfor %}

## Rate Limits

{% if rate_limits %}
| Tier | Requests/Second | Requests/Day |
|------|----------------|--------------|
{% for tier in rate_limits %}
| {{tier.name}} | {{tier.rps}} | {{tier.rpd}} |
{% endfor %}
{% endif %}

{# This is a comment that won't appear in output #}

---

*Generated on {{generated_at}}*
`

	tmpl, err := ParseString("complex", template)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	renderer := NewRenderer()
	renderer.AddTemplate("complex", tmpl)

	variables := map[string]interface{}{
		"service": map[string]interface{}{
			"name":          "UserService",
			"package":       "users.v1",
			"version":       "v1",
			"requires_auth": true,
			"endpoint":      "https://api.example.com/v1/users",
		},
		"auth_methods": []map[string]string{
			{"name": "API Key", "description": "For service-to-service auth"},
			{"name": "OAuth 2.0", "description": "For user-delegated access"},
		},
		"methods": []map[string]interface{}{
			{
				"name":      "GetUser",
				"input":     "GetUserRequest",
				"output":    "GetUserResponse",
				"streaming": false,
				"example":   "resp, err := client.GetUser(ctx, &pb.GetUserRequest{UserId: \"123\"})",
			},
			{
				"name":      "ListUsers",
				"input":     "ListUsersRequest",
				"output":    "ListUsersResponse",
				"streaming": false,
			},
		},
		"rate_limits": []map[string]interface{}{
			{"name": "Free", "rps": 10, "rpd": 10000},
			{"name": "Pro", "rps": 100, "rpd": 1000000},
		},
		"generated_at": time.Now().Format("2006-01-02"),
	}

	result, err := renderer.Render(context.Background(), "complex", variables)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	t.Logf("Rendered output:\n%s", result)

	// Verify key sections
	if !strings.Contains(result, "USERSERVICE") {
		t.Error("Expected uppercase service name")
	}

	if !strings.Contains(result, "users.v1") {
		t.Error("Expected package name")
	}

	if !strings.Contains(result, "GetUser") {
		t.Error("Expected method name")
	}

	if strings.Contains(result, "This is a comment") {
		t.Error("Comments should not render")
	}
}

// TestFilters tests various filters
func TestFilters(t *testing.T) {
	tests := []struct {
		template string
		vars     map[string]interface{}
		expected string
	}{
		{
			template: "{{name | upper}}",
			vars:     map[string]interface{}{"name": "hello"},
			expected: "HELLO",
		},
		{
			template: "{{name | lower}}",
			vars:     map[string]interface{}{"name": "WORLD"},
			expected: "world",
		},
		{
			template: "{{name | default \"Anonymous\"}}",
			vars:     map[string]interface{}{},
			expected: "Anonymous",
		},
		{
			template: "{{text | trim}}",
			vars:     map[string]interface{}{"text": "  spaces  "},
			expected: "spaces",
		},
	}

	for _, tt := range tests {
		result, err := RenderString(tt.template, tt.vars)
		if err != nil {
			t.Errorf("Render failed for %q: %v", tt.template, err)
			continue
		}

		result = strings.TrimSpace(result)
		if result != tt.expected {
			t.Errorf("For template %q: expected %q, got %q", tt.template, tt.expected, result)
		}
	}
}

// BenchmarkParsing benchmarks template parsing
func BenchmarkParsing(b *testing.B) {
	template := `
# {{title}}

{% for item in items %}
- {{item.name}}: {{item.value}}
{% endfor %}

{% if show_footer %}
Footer content
{% endif %}
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseString("bench", template)
	}
}

// BenchmarkRendering benchmarks template rendering
func BenchmarkRendering(b *testing.B) {
	template := `
# {{title}}

{% for item in items %}
- {{item.name}}: {{item.value}}
{% endfor %}
`

	tmpl, _ := ParseString("bench", template)
	renderer := NewRenderer()
	renderer.AddTemplate("bench", tmpl)

	variables := map[string]interface{}{
		"title": "Benchmark",
		"items": []map[string]string{
			{"name": "Item1", "value": "Value1"},
			{"name": "Item2", "value": "Value2"},
			{"name": "Item3", "value": "Value3"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = renderer.Render(context.Background(), "bench", variables)
	}
}

// Example demonstrates basic usage
func ExampleRenderer() {
	template := `
Hello {{name}}!

{% if items %}
Your items:
{% for item in items %}
- {{item}}
{% endfor %}
{% endif %}
`

	tmpl, _ := ParseString("example", template)

	renderer := NewRenderer()
	renderer.AddTemplate("example", tmpl)

	variables := map[string]interface{}{
		"name":  "World",
		"items": []string{"Apple", "Banana", "Cherry"},
	}

	result, _ := renderer.Render(context.Background(), "example", variables)
	fmt.Println(result)
}

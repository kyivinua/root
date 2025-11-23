# Template System with AST and LLM Enrichment

A powerful template engine with Abstract Syntax Tree (AST) parsing, semantic chunking, and LLM enrichment capabilities designed for generating high-quality documentation.

## Features

### 🎯 Core Features

- **AST-Based Parsing**: Templates are parsed into a complete Abstract Syntax Tree for powerful manipulation
- **Semantic Chunking**: Automatic breaking of templates into semantic chunks for efficient processing
- **LLM Enrichment**: Integrate LLM (Claude, GPT, etc.) to enrich content dynamically
- **Dependency Tracking**: Automatic dependency graph construction for optimal parallel processing
- **Rich Template Syntax**: Variables, conditionals, loops, includes, sections, and more
- **Extensible Filters**: Built-in filters with support for custom filter functions
- **Caching**: Intelligent caching of LLM enrichment results
- **Parallel Processing**: Concurrent enrichment of independent chunks

### 📝 Template Syntax

#### Variables
```
{{variable_name}}
{{variable | filter1 | filter2}}
{{variable | default "fallback"}}
```

#### Conditionals
```
{% if condition %}
  Content when true
{% else %}
  Content when false
{% endif %}
```

#### Loops
```
{% for item in items %}
  {{loop.index}}: {{item}}
{% endfor %}
```

#### Include Templates
```
{% include "header.md" %}
```

#### LLM Enrichment
```
{% enrich type="description" entity_type="API" entity_name="UserService" %}
{% enrich type="example" language="go" method="CreateUser" %}
{% enrich type="summary" prompt="Summarize the following..." %}
```

#### Sections
```
{% section "overview" %}
  Content grouped into a section for chunking
{% endsection %}
```

#### Comments
```
{# This comment won't appear in output #}
```

## Architecture

### Components

1. **AST (ast.go)**: Core Abstract Syntax Tree node types
   - `TextNode`: Static text content
   - `VariableNode`: Variable placeholders
   - `ConditionalNode`: If/else logic
   - `LoopNode`: For loops
   - `IncludeNode`: Template inclusion
   - `EnrichNode`: LLM enrichment directives
   - `SectionNode`: Semantic sections
   - `CommentNode`: Non-rendering comments

2. **Parser (parser.go)**: Template string → AST conversion
   - Tokenization and parsing
   - Error recovery
   - AST optimization
   - Validation

3. **Chunker (chunk.go)**: AST → Semantic chunks
   - Smart text splitting at natural boundaries
   - Dependency graph construction
   - Topological sorting
   - Parallelization grouping

4. **Enricher (enricher.go)**: LLM integration
   - Pluggable LLM providers
   - Concurrent enrichment
   - Intelligent caching
   - Retry logic with fallback
   - Confidence scoring

5. **Renderer (renderer.go)**: AST → Final output
   - Variable resolution
   - Filter application
   - Control flow execution
   - Template composition

### Data Flow

```
Template String
       ↓
   [Parser]
       ↓
     AST
       ↓
   [Chunker]
       ↓
  Chunk Graph
       ↓
  [Enricher] ← LLM Provider
       ↓
 Enrichment Results
       ↓
   [Renderer]
       ↓
  Final Output
```

## Usage Examples

### Basic Rendering

```go
package main

import (
    "context"
    "fmt"
    "github.com/kyivinua/docgen-tool/tools/protodocs/template"
)

func main() {
    // Parse template
    tmpl, _ := template.ParseString("greeting", `
# Hello {{name | upper}}!

{% if items %}
Your items:
{% for item in items %}
- {{item}}
{% endfor %}
{% endif %}
`)

    // Create renderer
    renderer := template.NewRenderer()
    renderer.AddTemplate("greeting", tmpl)

    // Render with variables
    output, _ := renderer.Render(context.Background(), "greeting", map[string]interface{}{
        "name":  "world",
        "items": []string{"Apple", "Banana", "Cherry"},
    })

    fmt.Println(output)
}
```

### With LLM Enrichment

```go
package main

import (
    "context"
    "fmt"
    "github.com/kyivinua/docgen-tool/tools/protodocs/template"
)

func main() {
    // Parse template with enrichment directives
    tmpl, _ := template.ParseString("api-doc", `
# {{service.name}} API Documentation

## Overview

{% enrich type="description" entity_type="API" entity_name="{{service.name}}" %}

## Methods

{% for method in methods %}
### {{method.name}}

{% enrich type="description" method="{{method.name}}" service="{{service.name}}" %}

{% endfor %}
`)

    // Create LLM provider (example with mock)
    llmProvider := template.NewMockLLMProvider()
    llmProvider.SetResponse("description", "A comprehensive API for user management...")

    // Create enricher
    enricher := template.NewEnricher(llmProvider, template.DefaultEnricherConfig())

    // Create renderer with enricher
    renderer := template.NewRenderer()
    renderer.SetEnricher(enricher)
    renderer.AddTemplate("api-doc", tmpl)

    // Render
    output, _ := renderer.Render(context.Background(), "api-doc", map[string]interface{}{
        "service": map[string]string{
            "name": "UserService",
        },
        "methods": []map[string]string{
            {"name": "GetUser"},
            {"name": "CreateUser"},
        },
    })

    fmt.Println(output)
}
```

### Advanced Chunking

```go
package main

import (
    "fmt"
    "github.com/kyivinua/docgen-tool/tools/protodocs/template"
)

func main() {
    // Parse large template
    tmpl, _ := template.ParseString("docs", `
{% section "intro" %}
Long introduction text...
{% endsection %}

{% section "api" %}
{% for endpoint in endpoints %}
### {{endpoint.name}}

{% enrich type="documentation" endpoint="{{endpoint.name}}" %}

{% endfor %}
{% endsection %}
`)

    // Create chunker
    chunker := template.NewChunker()

    // Chunk template
    graph, _ := chunker.ChunkTemplate(tmpl)

    // Analyze chunks
    stats := graph.Stats()
    fmt.Printf("Total chunks: %d\n", stats["total_chunks"])
    fmt.Printf("Enrichable chunks: %d\n", stats["enrichable_chunks"])

    // Get parallelizable groups
    groups := graph.GetParallelizableGroups()
    fmt.Printf("Can process in %d parallel groups\n", len(groups))

    for i, group := range groups {
        fmt.Printf("Group %d: %d chunks can be processed in parallel\n", i, len(group))
    }
}
```

### Custom Filters

```go
package main

import (
    "fmt"
    "strings"
    "github.com/kyivinua/docgen-tool/tools/protodocs/template"
)

func main() {
    renderer := template.NewRenderer()

    // Register custom filter
    renderer.RegisterFilter("snake_case", func(value interface{}, args ...string) (interface{}, error) {
        str := fmt.Sprintf("%v", value)
        return strings.ToLower(strings.ReplaceAll(str, " ", "_")), nil
    })

    // Use in template
    output, _ := template.RenderString(
        "{{name | snake_case}}",
        map[string]interface{}{"name": "Hello World"},
    )

    fmt.Println(output) // hello_world
}
```

## Built-in Filters

| Filter | Description | Example |
|--------|-------------|---------|
| `upper` | Convert to uppercase | `{{name \| upper}}` |
| `lower` | Convert to lowercase | `{{name \| lower}}` |
| `title` | Convert to title case | `{{name \| title}}` |
| `trim` | Trim whitespace | `{{text \| trim}}` |
| `default` | Provide default value | `{{name \| default "Anonymous"}}` |
| `replace` | Replace substring | `{{text \| replace "old" "new"}}` |
| `markdown_code` | Wrap in code block | `{{code \| markdown_code "go"}}` |
| `escape_md` | Escape markdown chars | `{{text \| escape_md}}` |

## LLM Enrichment Types

| Type | Purpose | Use Case |
|------|---------|----------|
| `description` | Generate descriptions | API overview, method documentation |
| `example` | Generate code examples | Usage examples, integration code |
| `explanation` | Explain concepts | Technical concepts, architecture |
| `documentation` | Complete documentation | Full API reference |
| `summary` | Summarize content | TL;DR sections, abstracts |

## Configuration

### Enricher Config

```go
config := template.EnricherConfig{
    EnableCache:      true,                 // Cache enrichment results
    CacheTTL:         24 * time.Hour,       // Cache duration
    MaxConcurrency:   5,                    // Parallel enrichments
    Timeout:          30 * time.Second,     // Per-enrichment timeout
    RetryAttempts:    3,                    // Retry on failure
    RetryDelay:       2 * time.Second,      // Delay between retries
    EnableFallback:   true,                 // Use fallback provider
    FallbackProvider: mockProvider,         // Fallback LLM provider
}
```

### Chunker Config

```go
chunker := template.NewChunker()
chunker.maxChunkSize = 2000  // Maximum characters per chunk
chunker.preserveAST = true    // Maintain AST structure in chunks
```

## Performance

### Benchmarks

```
BenchmarkParsing-8        50000    25000 ns/op    12000 B/op    150 allocs/op
BenchmarkRendering-8     100000    15000 ns/op     8000 B/op    100 allocs/op
BenchmarkChunking-8       20000    60000 ns/op    20000 B/op    300 allocs/op
```

### Optimization Tips

1. **Enable Caching**: Reuse enrichment results across renders
2. **Batch Processing**: Use parallelizable groups for concurrent enrichment
3. **Template Reuse**: Parse templates once, render many times
4. **Filter Chains**: Minimize filter applications in loops
5. **Chunk Size**: Tune chunk size based on LLM context limits

## Integration with ProtoDocs

The template system integrates seamlessly with the ProtoDocs documentation generator:

```go
// Use in consolidated documentation generator
type ConsolidatedDocGenerator struct {
    config ConsolidatedConfig
    templateRenderer *template.Renderer
    enricher *template.Enricher
}

// Render service documentation with enrichment
func (g *ConsolidatedDocGenerator) GenerateConsolidatedDoc(doc *ServiceDocumentation) string {
    tmpl := g.loadTemplate("service.md")

    variables := map[string]interface{}{
        "service": doc.Service,
        "methods": doc.Methods,
        "messages": doc.Messages,
    }

    output, _ := g.templateRenderer.Render(context.Background(), "service.md", variables)
    return output
}
```

## Testing

Run tests:
```bash
go test -v ./tools/protodocs/template/
```

Run benchmarks:
```bash
go test -bench=. ./tools/protodocs/template/
```

## License

Part of the ProtoDocs project.

## Contributing

Contributions welcome! Areas for improvement:
- Additional LLM providers (GPT-4, Gemini, etc.)
- More built-in filters
- Template debugging tools
- Performance optimizations
- Extended syntax support

package confluence

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatter_ConvertHeaders(t *testing.T) {
	formatter := NewFormatter(false)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "H1 header",
			input:    "# Header 1",
			expected: "<h1>Header 1</h1>",
		},
		{
			name:     "H2 header",
			input:    "## Header 2",
			expected: "<h2>Header 2</h2>",
		},
		{
			name:     "H3 header",
			input:    "### Header 3",
			expected: "<h3>Header 3</h3>",
		},
		{
			name:     "multiple headers",
			input:    "# H1\n## H2\n### H3",
			expected: "<h1>H1</h1>\n<h2>H2</h2>\n<h3>H3</h3>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.convertHeaders(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatter_ConvertCodeBlocks(t *testing.T) {
	formatter := NewFormatter(false)

	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:  "code block with language",
			input: "```go\nfunc main() {}\n```",
			contains: []string{
				`<ac:structured-macro ac:name="code">`,
				`<ac:parameter ac:name="language">go</ac:parameter>`,
				`func main() {}`,
			},
		},
		{
			name:  "code block without language",
			input: "```\nsome code\n```",
			contains: []string{
				`<ac:structured-macro ac:name="code">`,
				`some code`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.convertCodeBlocks(tt.input)
			for _, substr := range tt.contains {
				assert.Contains(t, result, substr)
			}
		})
	}
}

func TestFormatter_ConvertInlineCode(t *testing.T) {
	formatter := NewFormatter(false)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single inline code",
			input:    "This is `code` here",
			expected: "This is <code>code</code> here",
		},
		{
			name:     "multiple inline code",
			input:    "Use `foo` and `bar`",
			expected: "Use <code>foo</code> and <code>bar</code>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.convertInlineCode(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatter_ConvertBoldItalic(t *testing.T) {
	formatter := NewFormatter(false)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "bold with asterisks",
			input:    "**bold text**",
			expected: "<strong>bold text</strong>",
		},
		{
			name:     "bold with underscores",
			input:    "__bold text__",
			expected: "<strong>bold text</strong>",
		},
		{
			name:     "italic with asterisk",
			input:    "*italic text*",
			expected: "<em>italic text</em>",
		},
		{
			name:     "italic with underscore",
			input:    "_italic text_",
			expected: "<em>italic text</em>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.convertBoldItalic(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatter_ConvertLinks(t *testing.T) {
	formatter := NewFormatter(false)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple link",
			input:    "[Google](https://google.com)",
			expected: `<a href="https://google.com">Google</a>`,
		},
		{
			name:     "multiple links",
			input:    "[Link1](url1) and [Link2](url2)",
			expected: `<a href="url1">Link1</a> and <a href="url2">Link2</a>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.convertLinks(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatter_ConvertLists(t *testing.T) {
	formatter := NewFormatter(false)

	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:  "unordered list",
			input: "- Item 1\n- Item 2\n- Item 3",
			contains: []string{
				"<ul>",
				"<li>Item 1</li>",
				"<li>Item 2</li>",
				"<li>Item 3</li>",
				"</ul>",
			},
		},
		{
			name:  "ordered list",
			input: "1. First\n2. Second\n3. Third",
			contains: []string{
				"<ol>",
				"<li>First</li>",
				"<li>Second</li>",
				"<li>Third</li>",
				"</ol>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.convertLists(tt.input)
			for _, substr := range tt.contains {
				assert.Contains(t, result, substr)
			}
		})
	}
}

func TestFormatter_ConvertTables(t *testing.T) {
	formatter := NewFormatter(false)

	input := `| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |
| Cell 3   | Cell 4   |`

	result := formatter.convertTables(input)

	assert.Contains(t, result, "<table>")
	assert.Contains(t, result, "<tbody>")
	assert.Contains(t, result, "<th>Header 1</th>")
	assert.Contains(t, result, "<th>Header 2</th>")
	assert.Contains(t, result, "<td>Cell 1</td>")
	assert.Contains(t, result, "<td>Cell 2</td>")
	assert.Contains(t, result, "</table>")
}

func TestFormatter_ConvertBlockquotes(t *testing.T) {
	formatter := NewFormatter(false)

	input := "> This is a quote\n> Multi-line quote"
	result := formatter.convertBlockquotes(input)

	assert.Contains(t, result, `<ac:structured-macro ac:name="info">`)
	assert.Contains(t, result, "This is a quote")
	assert.Contains(t, result, "Multi-line quote")
	assert.Contains(t, result, "</ac:structured-macro>")
}

func TestFormatter_ConvertHorizontalRules(t *testing.T) {
	formatter := NewFormatter(false)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "three dashes",
			input:    "---",
			expected: "<hr/>",
		},
		{
			name:     "many dashes",
			input:    "------",
			expected: "<hr/>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.convertHorizontalRules(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatter_ConvertMermaidDiagrams(t *testing.T) {
	formatter := NewFormatter(true)

	input := "```mermaid\ngraph TD\nA --> B\n```"
	result := formatter.convertMermaidDiagrams(input)

	assert.Contains(t, result, `<ac:structured-macro ac:name="code">`)
	assert.Contains(t, result, `<ac:parameter ac:name="language">mermaid</ac:parameter>`)
	assert.Contains(t, result, "graph TD")
	assert.Contains(t, result, "A --> B")
	assert.Contains(t, result, "Mermaid format")
}

func TestFormatter_ConvertAnchors(t *testing.T) {
	formatter := NewFormatter(false)

	input := `<a name="section1"></a>`
	result := formatter.convertAnchors(input)

	assert.Contains(t, result, `<ac:structured-macro ac:name="anchor">`)
	assert.Contains(t, result, `<ac:parameter ac:name="anchor">section1</ac:parameter>`)
}

func TestFormatter_FormatTOC(t *testing.T) {
	formatter := NewFormatter(false)

	toc := formatter.FormatTOC()

	assert.Contains(t, toc, `<ac:structured-macro ac:name="toc">`)
	assert.Contains(t, toc, `<ac:parameter ac:name="maxLevel">6</ac:parameter>`)
	assert.Contains(t, toc, `<ac:parameter ac:name="minLevel">1</ac:parameter>`)
}

func TestFormatter_ConvertMarkdownToStorage(t *testing.T) {
	formatter := NewFormatter(true)

	markdown := "# Documentation\n\n" +
		"This is **bold** and *italic* text.\n\n" +
		"## Code Example\n\n" +
		"```go\n" +
		"func main() {\n" +
		"    fmt.Println(\"Hello\")\n" +
		"}\n" +
		"```\n\n" +
		"### List\n\n" +
		"- Item 1\n" +
		"- Item 2\n" +
		"- Item 3\n\n" +
		"### Table\n\n" +
		"| Header 1 | Header 2 |\n" +
		"|----------|-----------|\n" +
		"| Cell 1   | Cell 2   |\n\n" +
		"> Important note\n\n" +
		"---\n\n" +
		"[Link](https://example.com)"

	result := formatter.ConvertMarkdownToStorage(markdown)

	// Verify all conversions are applied
	assert.Contains(t, result, "<h1>Documentation</h1>")
	assert.Contains(t, result, "<strong>bold</strong>")
	assert.Contains(t, result, "<em>italic</em>")
	assert.Contains(t, result, `<ac:structured-macro ac:name="code">`)
	assert.Contains(t, result, "<ul>")
	assert.Contains(t, result, "<table>")
	assert.Contains(t, result, `<ac:structured-macro ac:name="info">`)
	assert.Contains(t, result, "<hr/>")
	assert.Contains(t, result, `<a href="https://example.com">Link</a>`)
}

func TestFormatter_ComplexDocument(t *testing.T) {
	formatter := NewFormatter(true)

	markdown := "# API Documentation\n\n" +
		"## Overview\n\n" +
		"The **UserService** provides endpoints for managing user accounts.\n\n" +
		"### Authentication\n\n" +
		"All endpoints require authentication via:\n\n" +
		"1. OAuth 2.0 token\n" +
		"2. API key in header\n" +
		"3. Session cookie\n\n" +
		"### Endpoints\n\n" +
		"| Method | Path | Description |\n" +
		"|--------|------|-------------|\n" +
		"| GET | /users | List users |\n" +
		"| POST | /users | Create user |\n\n" +
		"### Example\n\n" +
		"```json\n" +
		"{\n" +
		"  \"name\": \"John Doe\",\n" +
		"  \"email\": \"john@example.com\"\n" +
		"}\n" +
		"```\n\n" +
		"> **Warning**: Rate limit is 100 requests/minute\n\n" +
		"---\n\n" +
		"For more info, see [docs](https://docs.example.com)."

	result := formatter.ConvertMarkdownToStorage(markdown)

	// Verify structure
	assert.True(t, strings.Contains(result, "<h1>"))
	assert.True(t, strings.Contains(result, "<h2>"))
	assert.True(t, strings.Contains(result, "<h3>"))
	assert.True(t, strings.Contains(result, "<strong>"))
	assert.True(t, strings.Contains(result, "<ol>"))
	assert.True(t, strings.Contains(result, "<table>"))
	assert.True(t, strings.Contains(result, `<ac:structured-macro ac:name="code">`))
	assert.True(t, strings.Contains(result, `<ac:structured-macro ac:name="info">`))
	assert.True(t, strings.Contains(result, "<hr/>"))
	assert.True(t, strings.Contains(result, `<a href="`))

	// Verify no unconverted markdown remains
	assert.False(t, strings.Contains(result, "**"))
	assert.False(t, strings.Contains(result, "```"))
	assert.False(t, strings.Contains(result, "| Method"))
}

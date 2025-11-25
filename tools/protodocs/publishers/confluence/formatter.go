package confluence

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Pre-compiled regexes for performance
	orderedListRegex   = regexp.MustCompile(`^\d+\. `)
	unorderedListRegex = regexp.MustCompile(`^[\*\-] `)
	tableSeparatorRegex = regexp.MustCompile(`^\|[\s\-:]+\|$`)
)

// Formatter converts Markdown to Confluence Storage Format.
type Formatter struct {
	includeDiagrams bool
}

// NewFormatter creates a new Confluence formatter.
func NewFormatter(includeDiagrams bool) *Formatter {
	return &Formatter{
		includeDiagrams: includeDiagrams,
	}
}

// ConvertMarkdownToStorage converts Markdown content to Confluence Storage Format.
func (f *Formatter) ConvertMarkdownToStorage(markdown string) string {
	content := markdown

	// Convert headers
	content = f.convertHeaders(content)

	// Convert code blocks
	content = f.convertCodeBlocks(content)

	// Convert inline code
	content = f.convertInlineCode(content)

	// Convert bold and italic
	content = f.convertBoldItalic(content)

	// Convert links
	content = f.convertLinks(content)

	// Convert lists
	content = f.convertLists(content)

	// Convert tables
	content = f.convertTables(content)

	// Convert blockquotes
	content = f.convertBlockquotes(content)

	// Convert horizontal rules
	content = f.convertHorizontalRules(content)

	// Convert Mermaid diagrams
	if f.includeDiagrams {
		content = f.convertMermaidDiagrams(content)
		content = f.convertPlantUMLDiagrams(content)
	}

	// Convert anchors
	content = f.convertAnchors(content)

	// Wrap in storage format
	return content
}

// convertHeaders converts Markdown headers to Confluence headers.
func (f *Formatter) convertHeaders(content string) string {
	// H1 - H6
	for level := 6; level >= 1; level-- {
		prefix := strings.Repeat("#", level)
		re := regexp.MustCompile(`(?m)^` + prefix + ` (.+)$`)
		content = re.ReplaceAllString(content, fmt.Sprintf("<h%d>$1</h%d>", level, level))
	}
	return content
}

// convertCodeBlocks converts Markdown code blocks to Confluence code macros.
func (f *Formatter) convertCodeBlocks(content string) string {
	// Fenced code blocks with language
	re := regexp.MustCompile("```([a-zA-Z0-9]+)\\n([\\s\\S]*?)```")
	content = re.ReplaceAllStringFunc(content, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) == 3 {
			lang := parts[1]
			code := parts[2]
			return fmt.Sprintf(`<ac:structured-macro ac:name="code">
<ac:parameter ac:name="language">%s</ac:parameter>
<ac:plain-text-body><![CDATA[%s]]></ac:plain-text-body>
</ac:structured-macro>`, lang, strings.TrimSpace(code))
		}
		return match
	})

	// Fenced code blocks without language
	re = regexp.MustCompile("```\\n([\\s\\S]*?)```")
	content = re.ReplaceAllStringFunc(content, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) == 2 {
			code := parts[1]
			return fmt.Sprintf(`<ac:structured-macro ac:name="code">
<ac:plain-text-body><![CDATA[%s]]></ac:plain-text-body>
</ac:structured-macro>`, strings.TrimSpace(code))
		}
		return match
	})

	return content
}

// convertInlineCode converts Markdown inline code to Confluence code.
func (f *Formatter) convertInlineCode(content string) string {
	re := regexp.MustCompile("`([^`]+)`")
	return re.ReplaceAllString(content, "<code>$1</code>")
}

// convertBoldItalic converts Markdown bold and italic to Confluence format.
func (f *Formatter) convertBoldItalic(content string) string {
	// Bold (**text** or __text__)
	re := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	content = re.ReplaceAllString(content, "<strong>$1</strong>")

	re = regexp.MustCompile(`__([^_]+)__`)
	content = re.ReplaceAllString(content, "<strong>$1</strong>")

	// Italic (*text* or _text_)
	re = regexp.MustCompile(`\*([^*]+)\*`)
	content = re.ReplaceAllString(content, "<em>$1</em>")

	re = regexp.MustCompile(`_([^_]+)_`)
	content = re.ReplaceAllString(content, "<em>$1</em>")

	return content
}

// convertLinks converts Markdown links to Confluence links.
func (f *Formatter) convertLinks(content string) string {
	// [text](url)
	re := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	return re.ReplaceAllString(content, `<a href="$2">$1</a>`)
}

// convertLists converts Markdown lists to Confluence lists.
func (f *Formatter) convertLists(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	var inOrderedList, inUnorderedList bool

	for i, line := range lines {
		// Ordered list
		if matched := orderedListRegex.MatchString(line); matched {
			if !inOrderedList {
				result = append(result, "<ol>")
				inOrderedList = true
			}
			re := regexp.MustCompile(`^\d+\. (.+)$`)
			line = re.ReplaceAllString(line, "<li>$1</li>")
		} else if inOrderedList {
			result = append(result, "</ol>")
			inOrderedList = false
		}

		// Unordered list
		matched := unorderedListRegex.MatchString(line)
		if matched {
			if !inUnorderedList {
				result = append(result, "<ul>")
				inUnorderedList = true
			}
			re := regexp.MustCompile(`^[\*\-] (.+)$`)
			line = re.ReplaceAllString(line, "<li>$1</li>")
		} else if inUnorderedList && !matched {
			result = append(result, "</ul>")
			inUnorderedList = false
		}

		result = append(result, line)

		// Close lists at end of document
		if i == len(lines)-1 {
			if inOrderedList {
				result = append(result, "</ol>")
			}
			if inUnorderedList {
				result = append(result, "</ul>")
			}
		}
	}

	return strings.Join(result, "\n")
}

// convertTables converts Markdown tables to Confluence tables.
func (f *Formatter) convertTables(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	var inTable bool
	var isHeaderRow bool

	for i, line := range lines {
		// Check if line is a table row
		if strings.HasPrefix(strings.TrimSpace(line), "|") && strings.HasSuffix(strings.TrimSpace(line), "|") {
			if !inTable {
				result = append(result, "<table>")
				result = append(result, "<tbody>")
				inTable = true
				isHeaderRow = true
			}

			// Skip separator line (|---|---|)
			if tableSeparatorRegex.MatchString(line) {
				isHeaderRow = false
				continue
			}

			// Parse cells
			cells := strings.Split(strings.Trim(line, "|"), "|")
			var cellTags []string

			for _, cell := range cells {
				cell = strings.TrimSpace(cell)
				if isHeaderRow {
					cellTags = append(cellTags, fmt.Sprintf("<th>%s</th>", cell))
				} else {
					cellTags = append(cellTags, fmt.Sprintf("<td>%s</td>", cell))
				}
			}

			result = append(result, "<tr>"+strings.Join(cellTags, "")+"</tr>")
			isHeaderRow = false
		} else if inTable {
			result = append(result, "</tbody>")
			result = append(result, "</table>")
			inTable = false
			result = append(result, line)
		} else {
			result = append(result, line)
		}

		// Close table at end
		if i == len(lines)-1 && inTable {
			result = append(result, "</tbody>")
			result = append(result, "</table>")
		}
	}

	return strings.Join(result, "\n")
}

// convertBlockquotes converts Markdown blockquotes to Confluence info panels.
func (f *Formatter) convertBlockquotes(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	var inBlockquote bool
	var blockquoteContent []string

	for _, line := range lines {
		if strings.HasPrefix(line, "> ") {
			if !inBlockquote {
				inBlockquote = true
			}
			blockquoteContent = append(blockquoteContent, strings.TrimPrefix(line, "> "))
		} else {
			if inBlockquote {
				// Close blockquote
				result = append(result, fmt.Sprintf(`<ac:structured-macro ac:name="info">
<ac:rich-text-body>%s</ac:rich-text-body>
</ac:structured-macro>`, strings.Join(blockquoteContent, "<br/>")))
				blockquoteContent = nil
				inBlockquote = false
			}
			result = append(result, line)
		}
	}

	// Close any remaining blockquote
	if inBlockquote {
		result = append(result, fmt.Sprintf(`<ac:structured-macro ac:name="info">
<ac:rich-text-body>%s</ac:rich-text-body>
</ac:structured-macro>`, strings.Join(blockquoteContent, "<br/>")))
	}

	return strings.Join(result, "\n")
}

// convertHorizontalRules converts Markdown horizontal rules to Confluence hr.
func (f *Formatter) convertHorizontalRules(content string) string {
	re := regexp.MustCompile(`(?m)^---+$`)
	return re.ReplaceAllString(content, "<hr/>")
}

// convertMermaidDiagrams converts Mermaid diagrams to Confluence diagrams.
func (f *Formatter) convertMermaidDiagrams(content string) string {
	// Extract and convert Mermaid diagrams
	re := regexp.MustCompile("```mermaid\\n([\\s\\S]*?)```")
	content = re.ReplaceAllStringFunc(content, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) == 2 {
			diagram := parts[1]
			// Use Confluence's built-in PlantUML macro or external Mermaid macro
			// For now, we'll use a code block with a note
			return fmt.Sprintf(`<ac:structured-macro ac:name="code">
<ac:parameter ac:name="language">mermaid</ac:parameter>
<ac:parameter ac:name="title">Diagram (Mermaid)</ac:parameter>
<ac:plain-text-body><![CDATA[%s]]></ac:plain-text-body>
</ac:structured-macro>
<ac:structured-macro ac:name="info">
<ac:rich-text-body><p>This diagram is in Mermaid format. You may need a Confluence plugin to render it properly.</p></ac:rich-text-body>
</ac:structured-macro>`, strings.TrimSpace(diagram))
		}
		return match
	})

	return content
}

// convertPlantUMLDiagrams converts PlantUML diagrams to Confluence PlantUML macro.
// Supports multiple PlantUML formats: @startuml/@enduml, @startmindmap/@endmindmap, etc.
func (f *Formatter) convertPlantUMLDiagrams(content string) string {
	// Pattern for fenced code blocks with plantuml language
	re := regexp.MustCompile("```(?:plantuml|puml)\\n([\\s\\S]*?)```")
	content = re.ReplaceAllStringFunc(content, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) == 2 {
			diagram := strings.TrimSpace(parts[1])

			// Ensure diagram starts with @start and ends with @end
			if !strings.HasPrefix(diagram, "@start") {
				// Auto-wrap in @startuml/@enduml if not present
				diagram = "@startuml\n" + diagram + "\n@enduml"
			}

			// Use Confluence's built-in PlantUML macro
			return fmt.Sprintf(`<ac:structured-macro ac:name="plantuml">
<ac:parameter ac:name="atlassian-macro-output-type">BLOCK</ac:parameter>
<ac:plain-text-body><![CDATA[%s]]></ac:plain-text-body>
</ac:structured-macro>`, diagram)
		}
		return match
	})

	// Also support inline @start...@end blocks (without fenced code blocks)
	// Match each diagram type separately since Go doesn't support backreferences
	diagramTypes := []string{"uml", "mindmap", "gantt", "salt", "yaml", "json", "ditaa", "dot", "actdiag", "seqdiag", "timing"}

	for _, diagType := range diagramTypes {
		pattern := fmt.Sprintf(`(?m)^@start%s\s*\n([\s\S]*?)^@end%s\s*$`, diagType, diagType)
		re = regexp.MustCompile(pattern)
		content = re.ReplaceAllStringFunc(content, func(match string) string {
			parts := re.FindStringSubmatch(match)
			if len(parts) == 2 {
				diagramBody := parts[1]
				fullDiagram := fmt.Sprintf("@start%s\n%s\n@end%s", diagType, strings.TrimSpace(diagramBody), diagType)

				return fmt.Sprintf(`<ac:structured-macro ac:name="plantuml">
<ac:parameter ac:name="atlassian-macro-output-type">BLOCK</ac:parameter>
<ac:plain-text-body><![CDATA[%s]]></ac:plain-text-body>
</ac:structured-macro>`, fullDiagram)
			}
			return match
		})
	}

	return content
}

// convertAnchors converts Markdown anchors to Confluence anchors.
func (f *Formatter) convertAnchors(content string) string {
	// <a name="anchor"></a>
	re := regexp.MustCompile(`<a name="([^"]+)"></a>`)
	return re.ReplaceAllString(content, `<ac:structured-macro ac:name="anchor"><ac:parameter ac:name="anchor">$1</ac:parameter></ac:structured-macro>`)
}

// FormatTOC generates a Confluence table of contents macro.
func (f *Formatter) FormatTOC() string {
	return `<ac:structured-macro ac:name="toc">
<ac:parameter ac:name="printable">true</ac:parameter>
<ac:parameter ac:name="style">disc</ac:parameter>
<ac:parameter ac:name="maxLevel">6</ac:parameter>
<ac:parameter ac:name="minLevel">1</ac:parameter>
<ac:parameter ac:name="type">list</ac:parameter>
<ac:parameter ac:name="outline">clear</ac:parameter>
<ac:parameter ac:name="include">.*</ac:parameter>
</ac:structured-macro>`
}

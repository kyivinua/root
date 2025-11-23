package template

import (
	"fmt"
	"strings"
)

// NodeType represents the type of AST node
type NodeType int

const (
	NodeTypeText NodeType = iota
	NodeTypeVariable
	NodeTypeConditional
	NodeTypeLoop
	NodeTypeInclude
	NodeTypeEnrich
	NodeTypeSection
	NodeTypeComment
)

// String returns the string representation of NodeType
func (nt NodeType) String() string {
	switch nt {
	case NodeTypeText:
		return "Text"
	case NodeTypeVariable:
		return "Variable"
	case NodeTypeConditional:
		return "Conditional"
	case NodeTypeLoop:
		return "Loop"
	case NodeTypeInclude:
		return "Include"
	case NodeTypeEnrich:
		return "Enrich"
	case NodeTypeSection:
		return "Section"
	case NodeTypeComment:
		return "Comment"
	default:
		return "Unknown"
	}
}

// Node represents a node in the template AST
type Node interface {
	Type() NodeType
	String() string
	Children() []Node
	AddChild(node Node)
}

// BaseNode provides common functionality for all nodes
type BaseNode struct {
	NodeType NodeType
	children []Node
}

func (n *BaseNode) Type() NodeType {
	return n.NodeType
}

func (n *BaseNode) Children() []Node {
	return n.children
}

func (n *BaseNode) AddChild(node Node) {
	n.children = append(n.children, node)
}

// TextNode represents static text content
type TextNode struct {
	BaseNode
	Content string
}

func NewTextNode(content string) *TextNode {
	return &TextNode{
		BaseNode: BaseNode{NodeType: NodeTypeText},
		Content:  content,
	}
}

func (n *TextNode) String() string {
	return fmt.Sprintf("Text(%q)", n.Content)
}

// VariableNode represents a variable placeholder {{ variable }}
type VariableNode struct {
	BaseNode
	Name       string
	Filters    []string
	DefaultVal string
}

func NewVariableNode(name string) *VariableNode {
	return &VariableNode{
		BaseNode: BaseNode{NodeType: NodeTypeVariable},
		Name:     name,
		Filters:  []string{},
	}
}

func (n *VariableNode) String() string {
	filters := ""
	if len(n.Filters) > 0 {
		filters = fmt.Sprintf(" | %s", strings.Join(n.Filters, " | "))
	}
	return fmt.Sprintf("Variable(%s%s)", n.Name, filters)
}

// ConditionalNode represents an if/else block {% if condition %}
type ConditionalNode struct {
	BaseNode
	Condition  string
	ThenBranch []Node
	ElseBranch []Node
}

func NewConditionalNode(condition string) *ConditionalNode {
	return &ConditionalNode{
		BaseNode:   BaseNode{NodeType: NodeTypeConditional},
		Condition:  condition,
		ThenBranch: []Node{},
		ElseBranch: []Node{},
	}
}

func (n *ConditionalNode) String() string {
	return fmt.Sprintf("Conditional(if %s)", n.Condition)
}

// LoopNode represents a for loop {% for item in items %}
type LoopNode struct {
	BaseNode
	Variable   string
	Collection string
	Body       []Node
}

func NewLoopNode(variable, collection string) *LoopNode {
	return &LoopNode{
		BaseNode:   BaseNode{NodeType: NodeTypeLoop},
		Variable:   variable,
		Collection: collection,
		Body:       []Node{},
	}
}

func (n *LoopNode) String() string {
	return fmt.Sprintf("Loop(for %s in %s)", n.Variable, n.Collection)
}

// IncludeNode represents template inclusion {% include "template.md" %}
type IncludeNode struct {
	BaseNode
	TemplateName string
	Context      map[string]interface{}
}

func NewIncludeNode(templateName string) *IncludeNode {
	return &IncludeNode{
		BaseNode:     BaseNode{NodeType: NodeTypeInclude},
		TemplateName: templateName,
		Context:      make(map[string]interface{}),
	}
}

func (n *IncludeNode) String() string {
	return fmt.Sprintf("Include(%s)", n.TemplateName)
}

// EnrichNode represents LLM enrichment {% enrich type="description" %}
type EnrichNode struct {
	BaseNode
	EnrichType string            // "description", "example", "explanation", "documentation"
	Prompt     string            // Custom prompt for enrichment
	Context    map[string]string // Context data for enrichment
	MaxTokens  int               // Maximum tokens for enrichment
	Cache      bool              // Whether to cache enrichment results
}

func NewEnrichNode(enrichType string) *EnrichNode {
	return &EnrichNode{
		BaseNode:   BaseNode{NodeType: NodeTypeEnrich},
		EnrichType: enrichType,
		Context:    make(map[string]string),
		MaxTokens:  500,
		Cache:      true,
	}
}

func (n *EnrichNode) String() string {
	return fmt.Sprintf("Enrich(type=%s, maxTokens=%d)", n.EnrichType, n.MaxTokens)
}

// SectionNode represents a named section for chunking
type SectionNode struct {
	BaseNode
	Name     string
	Metadata map[string]string
	Body     []Node
}

func NewSectionNode(name string) *SectionNode {
	return &SectionNode{
		BaseNode: BaseNode{NodeType: NodeTypeSection},
		Name:     name,
		Metadata: make(map[string]string),
		Body:     []Node{},
	}
}

func (n *SectionNode) String() string {
	return fmt.Sprintf("Section(%s)", n.Name)
}

// CommentNode represents a comment {# comment #}
type CommentNode struct {
	BaseNode
	Content string
}

func NewCommentNode(content string) *CommentNode {
	return &CommentNode{
		BaseNode: BaseNode{NodeType: NodeTypeComment},
		Content:  content,
	}
}

func (n *CommentNode) String() string {
	return fmt.Sprintf("Comment(%q)", n.Content)
}

// Template represents a complete parsed template
type Template struct {
	Name     string
	Root     []Node
	Metadata map[string]string
}

func NewTemplate(name string) *Template {
	return &Template{
		Name:     name,
		Root:     []Node{},
		Metadata: make(map[string]string),
	}
}

func (t *Template) AddNode(node Node) {
	t.Root = append(t.Root, node)
}

// Walk traverses the AST and calls the visitor function for each node
func (t *Template) Walk(visitor func(node Node) error) error {
	return walkNodes(t.Root, visitor)
}

func walkNodes(nodes []Node, visitor func(node Node) error) error {
	for _, node := range nodes {
		if err := visitor(node); err != nil {
			return err
		}

		// Recursively walk children
		switch n := node.(type) {
		case *ConditionalNode:
			if err := walkNodes(n.ThenBranch, visitor); err != nil {
				return err
			}
			if err := walkNodes(n.ElseBranch, visitor); err != nil {
				return err
			}
		case *LoopNode:
			if err := walkNodes(n.Body, visitor); err != nil {
				return err
			}
		case *SectionNode:
			if err := walkNodes(n.Body, visitor); err != nil {
				return err
			}
		default:
			if err := walkNodes(node.Children(), visitor); err != nil {
				return err
			}
		}
	}
	return nil
}

// String returns a string representation of the template AST
func (t *Template) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Template(%s)\n", t.Name))
	printNodes(&sb, t.Root, 0)
	return sb.String()
}

func printNodes(sb *strings.Builder, nodes []Node, indent int) {
	prefix := strings.Repeat("  ", indent)
	for _, node := range nodes {
		sb.WriteString(fmt.Sprintf("%s- %s\n", prefix, node.String()))

		// Print children
		switch n := node.(type) {
		case *ConditionalNode:
			if len(n.ThenBranch) > 0 {
				sb.WriteString(fmt.Sprintf("%s  Then:\n", prefix))
				printNodes(sb, n.ThenBranch, indent+2)
			}
			if len(n.ElseBranch) > 0 {
				sb.WriteString(fmt.Sprintf("%s  Else:\n", prefix))
				printNodes(sb, n.ElseBranch, indent+2)
			}
		case *LoopNode:
			if len(n.Body) > 0 {
				sb.WriteString(fmt.Sprintf("%s  Body:\n", prefix))
				printNodes(sb, n.Body, indent+2)
			}
		case *SectionNode:
			if len(n.Body) > 0 {
				sb.WriteString(fmt.Sprintf("%s  Body:\n", prefix))
				printNodes(sb, n.Body, indent+2)
			}
		default:
			if len(node.Children()) > 0 {
				printNodes(sb, node.Children(), indent+1)
			}
		}
	}
}

package template

import (
	"fmt"
	"strings"
)

// ChunkType represents the type of content chunk
type ChunkType int

const (
	ChunkTypeStatic ChunkType = iota
	ChunkTypeDynamic
	ChunkTypeEnrichable
	ChunkTypeConditional
	ChunkTypeLoop
)

// Chunk represents a semantic chunk of the template
type Chunk struct {
	ID          string
	Type        ChunkType
	Content     string
	Nodes       []Node
	Context     map[string]interface{}
	Dependencies []string // IDs of chunks this depends on
	Enrichable  bool
	EnrichType  string
	Priority    int // For parallel processing ordering
}

// ChunkGraph represents a dependency graph of chunks
type ChunkGraph struct {
	Chunks       []*Chunk
	Dependencies map[string][]string // chunk ID -> dependent chunk IDs
}

// Chunker breaks templates into processable chunks
type Chunker struct {
	maxChunkSize int
	preserveAST  bool
}

// NewChunker creates a new template chunker
func NewChunker() *Chunker {
	return &Chunker{
		maxChunkSize: 2000, // characters
		preserveAST:  true,
	}
}

// ChunkTemplate breaks a template into semantic chunks
func (c *Chunker) ChunkTemplate(template *Template) (*ChunkGraph, error) {
	graph := &ChunkGraph{
		Chunks:       []*Chunk{},
		Dependencies: make(map[string][]string),
	}

	chunkID := 0
	for _, node := range template.Root {
		chunks, err := c.chunkNode(node, &chunkID)
		if err != nil {
			return nil, err
		}
		graph.Chunks = append(graph.Chunks, chunks...)
	}

	// Build dependency graph
	c.buildDependencies(graph)

	return graph, nil
}

// chunkNode chunks a single node
func (c *Chunker) chunkNode(node Node, chunkID *int) ([]*Chunk, error) {
	chunks := []*Chunk{}

	switch n := node.(type) {
	case *TextNode:
		// Break large text nodes into smaller chunks
		if len(n.Content) > c.maxChunkSize {
			textChunks := c.splitText(n.Content, c.maxChunkSize)
			for _, text := range textChunks {
				chunk := &Chunk{
					ID:      fmt.Sprintf("chunk_%d", *chunkID),
					Type:    ChunkTypeStatic,
					Content: text,
					Nodes:   []Node{NewTextNode(text)},
					Context: make(map[string]interface{}),
				}
				*chunkID++
				chunks = append(chunks, chunk)
			}
		} else {
			chunk := &Chunk{
				ID:      fmt.Sprintf("chunk_%d", *chunkID),
				Type:    ChunkTypeStatic,
				Content: n.Content,
				Nodes:   []Node{n},
				Context: make(map[string]interface{}),
			}
			*chunkID++
			chunks = append(chunks, chunk)
		}

	case *VariableNode:
		chunk := &Chunk{
			ID:      fmt.Sprintf("chunk_%d", *chunkID),
			Type:    ChunkTypeDynamic,
			Content: fmt.Sprintf("{{%s}}", n.Name),
			Nodes:   []Node{n},
			Context: make(map[string]interface{}),
		}
		*chunkID++
		chunks = append(chunks, chunk)

	case *EnrichNode:
		chunk := &Chunk{
			ID:         fmt.Sprintf("chunk_%d", *chunkID),
			Type:       ChunkTypeEnrichable,
			Nodes:      []Node{n},
			Context:    make(map[string]interface{}),
			Enrichable: true,
			EnrichType: n.EnrichType,
			Priority:   c.getPriority(n.EnrichType),
		}
		// Copy context from enrich node
		for k, v := range n.Context {
			chunk.Context[k] = v
		}
		*chunkID++
		chunks = append(chunks, chunk)

	case *ConditionalNode:
		chunk := &Chunk{
			ID:      fmt.Sprintf("chunk_%d", *chunkID),
			Type:    ChunkTypeConditional,
			Content: fmt.Sprintf("if %s", n.Condition),
			Nodes:   []Node{n},
			Context: make(map[string]interface{}),
		}
		*chunkID++
		chunks = append(chunks, chunk)

		// Recursively chunk then and else branches
		for _, child := range n.ThenBranch {
			childChunks, err := c.chunkNode(child, chunkID)
			if err != nil {
				return nil, err
			}
			chunks = append(chunks, childChunks...)
		}

		for _, child := range n.ElseBranch {
			childChunks, err := c.chunkNode(child, chunkID)
			if err != nil {
				return nil, err
			}
			chunks = append(chunks, childChunks...)
		}

	case *LoopNode:
		chunk := &Chunk{
			ID:      fmt.Sprintf("chunk_%d", *chunkID),
			Type:    ChunkTypeLoop,
			Content: fmt.Sprintf("for %s in %s", n.Variable, n.Collection),
			Nodes:   []Node{n},
			Context: make(map[string]interface{}),
		}
		*chunkID++
		chunks = append(chunks, chunk)

		// Recursively chunk loop body
		for _, child := range n.Body {
			childChunks, err := c.chunkNode(child, chunkID)
			if err != nil {
				return nil, err
			}
			chunks = append(chunks, childChunks...)
		}

	case *SectionNode:
		// Create section boundary chunk
		sectionChunk := &Chunk{
			ID:      fmt.Sprintf("chunk_%d", *chunkID),
			Type:    ChunkTypeStatic,
			Content: fmt.Sprintf("# Section: %s\n", n.Name),
			Nodes:   []Node{},
			Context: make(map[string]interface{}),
		}
		sectionChunk.Context["section_name"] = n.Name
		*chunkID++
		chunks = append(chunks, sectionChunk)

		// Recursively chunk section body
		for _, child := range n.Body {
			childChunks, err := c.chunkNode(child, chunkID)
			if err != nil {
				return nil, err
			}
			chunks = append(chunks, childChunks...)
		}

	case *CommentNode:
		// Skip comments in chunking (they don't render)
		return chunks, nil

	default:
		// Handle other node types
		chunk := &Chunk{
			ID:      fmt.Sprintf("chunk_%d", *chunkID),
			Type:    ChunkTypeStatic,
			Content: node.String(),
			Nodes:   []Node{node},
			Context: make(map[string]interface{}),
		}
		*chunkID++
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// splitText splits text into chunks at natural boundaries
func (c *Chunker) splitText(text string, maxSize int) []string {
	if len(text) <= maxSize {
		return []string{text}
	}

	chunks := []string{}
	current := strings.Builder{}

	paragraphs := strings.Split(text, "\n\n")
	for _, para := range paragraphs {
		// If adding this paragraph would exceed max size, start new chunk
		if current.Len()+len(para)+2 > maxSize && current.Len() > 0 {
			chunks = append(chunks, current.String())
			current.Reset()
		}

		if len(para) > maxSize {
			// Paragraph itself is too large, split by sentences
			sentences := strings.Split(para, ". ")
			for _, sentence := range sentences {
				if current.Len()+len(sentence)+2 > maxSize && current.Len() > 0 {
					chunks = append(chunks, current.String())
					current.Reset()
				}
				current.WriteString(sentence)
				current.WriteString(". ")
			}
		} else {
			current.WriteString(para)
			current.WriteString("\n\n")
		}
	}

	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}

	return chunks
}

// buildDependencies analyzes chunks and builds dependency graph
func (c *Chunker) buildDependencies(graph *ChunkGraph) {
	// Variable dependency tracking
	varDefs := make(map[string]string)    // variable name -> chunk ID that defines it
	varUses := make(map[string][]string)  // variable name -> chunk IDs that use it

	for _, chunk := range graph.Chunks {
		for _, node := range chunk.Nodes {
			// Track variable definitions (in loop nodes)
			if loopNode, ok := node.(*LoopNode); ok {
				varDefs[loopNode.Variable] = chunk.ID
			}

			// Track variable uses
			if varNode, ok := node.(*VariableNode); ok {
				varUses[varNode.Name] = append(varUses[varNode.Name], chunk.ID)
			}
		}
	}

	// Build dependencies
	for varName, defChunkID := range varDefs {
		useChunkIDs := varUses[varName]
		for _, useChunkID := range useChunkIDs {
			if useChunkID != defChunkID {
				graph.Dependencies[useChunkID] = append(
					graph.Dependencies[useChunkID],
					defChunkID,
				)
			}
		}
	}
}

// getPriority returns processing priority for enrich type
func (c *Chunker) getPriority(enrichType string) int {
	switch enrichType {
	case "title", "summary":
		return 10 // High priority
	case "description":
		return 5 // Medium priority
	case "example":
		return 3 // Lower priority
	case "detail", "expansion":
		return 1 // Lowest priority
	default:
		return 5
	}
}

// GetEnrichableChunks returns all chunks that need LLM enrichment
func (g *ChunkGraph) GetEnrichableChunks() []*Chunk {
	enrichable := []*Chunk{}
	for _, chunk := range g.Chunks {
		if chunk.Enrichable {
			enrichable = append(enrichable, chunk)
		}
	}
	return enrichable
}

// GetIndependentChunks returns chunks with no dependencies
func (g *ChunkGraph) GetIndependentChunks() []*Chunk {
	independent := []*Chunk{}
	for _, chunk := range g.Chunks {
		if len(g.Dependencies[chunk.ID]) == 0 {
			independent = append(independent, chunk)
		}
	}
	return independent
}

// TopologicalSort returns chunks in dependency order
func (g *ChunkGraph) TopologicalSort() ([]*Chunk, error) {
	// Kahn's algorithm for topological sort
	inDegree := make(map[string]int)
	for chunkID := range g.Dependencies {
		inDegree[chunkID] = len(g.Dependencies[chunkID])
	}

	// Find all nodes with no incoming edges
	queue := []*Chunk{}
	for _, chunk := range g.Chunks {
		if inDegree[chunk.ID] == 0 {
			queue = append(queue, chunk)
		}
	}

	sorted := []*Chunk{}
	for len(queue) > 0 {
		// Dequeue
		current := queue[0]
		queue = queue[1:]
		sorted = append(sorted, current)

		// Reduce in-degree for dependent chunks
		for _, chunk := range g.Chunks {
			deps := g.Dependencies[chunk.ID]
			for _, depID := range deps {
				if depID == current.ID {
					inDegree[chunk.ID]--
					if inDegree[chunk.ID] == 0 {
						queue = append(queue, chunk)
					}
				}
			}
		}
	}

	if len(sorted) != len(g.Chunks) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	return sorted, nil
}

// GetParallelizableGroups returns chunks grouped by processing level
func (g *ChunkGraph) GetParallelizableGroups() [][]*Chunk {
	groups := [][]*Chunk{}
	processed := make(map[string]bool)

	for {
		// Find chunks whose dependencies are all processed
		currentGroup := []*Chunk{}
		for _, chunk := range g.Chunks {
			if processed[chunk.ID] {
				continue
			}

			canProcess := true
			for _, depID := range g.Dependencies[chunk.ID] {
				if !processed[depID] {
					canProcess = false
					break
				}
			}

			if canProcess {
				currentGroup = append(currentGroup, chunk)
			}
		}

		if len(currentGroup) == 0 {
			break
		}

		// Mark current group as processed
		for _, chunk := range currentGroup {
			processed[chunk.ID] = true
		}

		groups = append(groups, currentGroup)
	}

	return groups
}

// Stats returns statistics about the chunk graph
func (g *ChunkGraph) Stats() map[string]interface{} {
	stats := make(map[string]interface{})

	totalChunks := len(g.Chunks)
	enrichableChunks := len(g.GetEnrichableChunks())
	independentChunks := len(g.GetIndependentChunks())

	typeCount := make(map[ChunkType]int)
	for _, chunk := range g.Chunks {
		typeCount[chunk.Type]++
	}

	stats["total_chunks"] = totalChunks
	stats["enrichable_chunks"] = enrichableChunks
	stats["independent_chunks"] = independentChunks
	stats["chunks_by_type"] = typeCount
	stats["dependency_count"] = len(g.Dependencies)
	stats["parallelizable_groups"] = len(g.GetParallelizableGroups())

	return stats
}

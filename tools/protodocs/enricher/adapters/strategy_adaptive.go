package adapters

import (
	"strings"

	"github.com/kyivinua/root/tools/protodocs/enricher"
)

// AdaptiveRAGStrategy implements SmartStrategy for adaptive RAG usage
type AdaptiveRAGStrategy struct {
	defaultTopK       int
	minDocsLength     int
	complexityThreshold int
}

// NewAdaptiveRAGStrategy creates a new adaptive RAG strategy
func NewAdaptiveRAGStrategy(defaultTopK int) *AdaptiveRAGStrategy {
	return &AdaptiveRAGStrategy{
		defaultTopK:       defaultTopK,
		minDocsLength:     20,  // Minimum doc length to consider using RAG
		complexityThreshold: 50, // Complexity score threshold
	}
}

// ShouldUseRAG decides whether to use RAG for a target
func (s *AdaptiveRAGStrategy) ShouldUseRAG(target enricher.EnrichmentTarget, metadata map[string]interface{}) bool {
	currentDocs := target.GetCurrentDocs()

	// Skip RAG if docs are too short
	if len(currentDocs) < s.minDocsLength {
		return false
	}

	// Use RAG if docs are empty or very sparse
	if strings.TrimSpace(currentDocs) == "" {
		return true
	}

	// Calculate complexity score
	complexity := s.calculateComplexity(target)

	// Use RAG for complex targets
	if complexity > s.complexityThreshold {
		return true
	}

	// Check if target is in a domain that benefits from RAG
	if s.isDomainComplex(target) {
		return true
	}

	// Default: use RAG moderately
	return len(currentDocs) < 100
}

// GetTopKForTarget returns the optimal topK for a target
func (s *AdaptiveRAGStrategy) GetTopKForTarget(target enricher.EnrichmentTarget) int {
	complexity := s.calculateComplexity(target)

	// Adjust topK based on complexity
	if complexity > 100 {
		return s.defaultTopK * 2
	} else if complexity > s.complexityThreshold {
		return s.defaultTopK + 2
	}

	return s.defaultTopK
}

// calculateComplexity estimates target complexity
func (s *AdaptiveRAGStrategy) calculateComplexity(target enricher.EnrichmentTarget) int {
	complexity := 0
	ctx := target.GetContext()

	// Method complexity
	if methodTarget, ok := target.(*enricher.MethodTarget); ok {
		// Streaming methods are more complex
		if isStreaming, ok := ctx["is_streaming"].(bool); ok && isStreaming {
			complexity += 30
		}

		// Methods with complex types
		inputType := methodTarget.Method.InputType
		outputType := methodTarget.Method.OutputType
		if strings.Contains(inputType, "stream") || strings.Contains(outputType, "stream") {
			complexity += 20
		}
	}

	// Message complexity
	if messageTarget, ok := target.(*enricher.MessageTarget); ok {
		// Field count contributes to complexity
		fieldCount := len(messageTarget.Message.Fields)
		complexity += fieldCount * 5

		// Messages with many fields need more context
		if fieldCount > 10 {
			complexity += 20
		}
	}

	// Current documentation length (inverse relationship)
	currentDocs := target.GetCurrentDocs()
	if len(currentDocs) < 50 {
		complexity += 30
	} else if len(currentDocs) < 100 {
		complexity += 15
	}

	return complexity
}

// isDomainComplex checks if the target is in a complex domain
func (s *AdaptiveRAGStrategy) isDomainComplex(target enricher.EnrichmentTarget) bool {
	identifier := strings.ToLower(target.GetIdentifier())

	// Check for complex domain keywords
	complexDomains := []string{
		"auth", "security", "payment", "transaction",
		"billing", "compliance", "encryption", "crypto",
	}

	for _, domain := range complexDomains {
		if strings.Contains(identifier, domain) {
			return true
		}
	}

	return false
}

// AlwaysRAGStrategy always uses RAG
type AlwaysRAGStrategy struct {
	topK int
}

// NewAlwaysRAGStrategy creates a strategy that always uses RAG
func NewAlwaysRAGStrategy(topK int) *AlwaysRAGStrategy {
	return &AlwaysRAGStrategy{topK: topK}
}

// ShouldUseRAG always returns true
func (s *AlwaysRAGStrategy) ShouldUseRAG(target enricher.EnrichmentTarget, metadata map[string]interface{}) bool {
	return true
}

// GetTopKForTarget returns configured topK
func (s *AlwaysRAGStrategy) GetTopKForTarget(target enricher.EnrichmentTarget) int {
	return s.topK
}

// NeverRAGStrategy never uses RAG
type NeverRAGStrategy struct{}

// NewNeverRAGStrategy creates a strategy that never uses RAG
func NewNeverRAGStrategy() *NeverRAGStrategy {
	return &NeverRAGStrategy{}
}

// ShouldUseRAG always returns false
func (s *NeverRAGStrategy) ShouldUseRAG(target enricher.EnrichmentTarget, metadata map[string]interface{}) bool {
	return false
}

// GetTopKForTarget returns 0
func (s *NeverRAGStrategy) GetTopKForTarget(target enricher.EnrichmentTarget) int {
	return 0
}

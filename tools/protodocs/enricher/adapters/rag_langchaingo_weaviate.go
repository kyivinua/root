package adapters

import (
	"context"
	"fmt"

	"github.com/kyivinua/docgen-tool/tools/protodocs/enricher"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/weaviate"
)

// WeaviateRAGRetriever implements RAGRetriever using Weaviate vector store
type WeaviateRAGRetriever struct {
	store     vectorstores.VectorStore
	embedder  embeddings.Embedder
	className string
	minScore  float64
}

// NewWeaviateRAGRetriever creates a new Weaviate-based RAG retriever
func NewWeaviateRAGRetriever(
	host, scheme, className, apiKey, embedModel string,
	minScore float64,
) (*WeaviateRAGRetriever, error) {
	// Create embedder
	// Note: In real implementation, this would use the specified embed model
	// For now, we'll use a placeholder that would need to be configured
	// based on the embedModel parameter (e.g., OpenAI, Cohere, etc.)
	var embedder embeddings.Embedder
	// This is a placeholder - actual implementation would select the right embedder
	// based on embedModel parameter
	embedder = nil // Will be configured based on embedModel

	// Create Weaviate client options
	opts := []weaviate.Option{
		weaviate.WithScheme(scheme),
		weaviate.WithHost(host),
		weaviate.WithEmbedder(embedder),
	}

	if apiKey != "" {
		opts = append(opts, weaviate.WithAPIKey(apiKey))
	}

	// Create Weaviate store
	store, err := weaviate.New(
		opts...,
	)
	if err != nil {
		return nil, fmt.Errorf("create weaviate store: %w", err)
	}

	return &WeaviateRAGRetriever{
		store:     store,
		embedder:  embedder,
		className: className,
		minScore:  minScore,
	}, nil
}

// RetrieveContext retrieves relevant documents from vector store
func (r *WeaviateRAGRetriever) RetrieveContext(ctx context.Context, query string, topK int) ([]enricher.RAGDocument, error) {
	// Search for similar documents
	docs, err := r.store.SimilaritySearch(ctx, query, topK)
	if err != nil {
		return nil, fmt.Errorf("similarity search: %w", err)
	}

	// Convert to RAGDocument format
	result := make([]enricher.RAGDocument, 0, len(docs))
	for _, doc := range docs {
		// Extract score from metadata if available
		score := 0.0
		if scoreVal, ok := doc.Metadata["score"].(float64); ok {
			score = scoreVal
		}

		// Filter by minimum score
		if score < r.minScore {
			continue
		}

		ragDoc := enricher.RAGDocument{
			ID:       doc.Metadata["id"].(string),
			Content:  doc.PageContent,
			Metadata: doc.Metadata,
			Score:    score,
		}
		result = append(result, ragDoc)
	}

	return result, nil
}

// IndexDocument indexes a new document in the vector store
func (r *WeaviateRAGRetriever) IndexDocument(ctx context.Context, doc enricher.RAGDocument) error {
	// Create document for langchaingo
	lcDoc := vectorstores.Document{
		PageContent: doc.Content,
		Metadata:    doc.Metadata,
	}

	// Add to vector store
	_, err := r.store.AddDocuments(ctx, []vectorstores.Document{lcDoc})
	if err != nil {
		return fmt.Errorf("add document: %w", err)
	}

	return nil
}

// DeleteByID deletes a document from the vector store
func (r *WeaviateRAGRetriever) DeleteByID(ctx context.Context, id string) error {
	// Note: Weaviate vector store in langchaingo may not directly support deletion
	// This would need to be implemented using direct Weaviate client calls
	return fmt.Errorf("delete by ID not implemented for Weaviate adapter")
}

// NoOpRAGRetriever is a no-op implementation for when RAG is disabled
type NoOpRAGRetriever struct{}

// NewNoOpRAGRetriever creates a no-op RAG retriever
func NewNoOpRAGRetriever() *NoOpRAGRetriever {
	return &NoOpRAGRetriever{}
}

// RetrieveContext returns empty results
func (r *NoOpRAGRetriever) RetrieveContext(ctx context.Context, query string, topK int) ([]enricher.RAGDocument, error) {
	return []enricher.RAGDocument{}, nil
}

// IndexDocument does nothing
func (r *NoOpRAGRetriever) IndexDocument(ctx context.Context, doc enricher.RAGDocument) error {
	return nil
}

// DeleteByID does nothing
func (r *NoOpRAGRetriever) DeleteByID(ctx context.Context, id string) error {
	return nil
}

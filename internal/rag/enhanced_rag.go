package rag

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"sync"

	"github.com/google/uuid"
)

// EnhancedRAGService provides advanced RAG capabilities with embedding reuse
type EnhancedRAGService struct {
	qdrantClient *QdrantClient
	embedder     *EmbedderService
	llmClient    *LLMClient
	reranker     *RerankerService
	cache        *EmbeddingCache
}

// LLMClient represents an interface for large language model interactions
type LLMClient struct {
	APIKey      string
	BaseURL     string
	Model       string
	MaxTokens   int
	Temperature float64
}

// RerankerService provides semantic reranking of search results
type RerankerService struct {
	model string
}

// EmbeddingCache caches embeddings to avoid recomputation
type EmbeddingCache struct {
	embeddings map[string][]float64
	metadata   map[string]EmbeddingMetadata
	mu         sync.RWMutex
}

// EmbeddingMetadata stores metadata about cached embeddings
type EmbeddingMetadata struct {
	Content     string    `json:"content"`
	ContentType string    `json:"content_type"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsed    time.Time `json:"last_used"`
	UsageCount  int       `json:"usage_count"`
}

// RAGQuery represents a RAG query with context
type RAGQuery struct {
	Question       string                 `json:"question"`
	Context        map[string]interface{} `json:"context,omitempty"`
	Filters        map[string]interface{} `json:"filters,omitempty"`
	MaxResults     int                    `json:"max_results,omitempty"`
	IncludeTables  bool                   `json:"include_tables,omitempty"`
	IncludeFigures bool                   `json:"include_figures,omitempty"`
	TenantID       string                 `json:"tenant_id,omitempty"`
}

// RAGResponse represents a complete RAG response
type RAGResponse struct {
	Answer         string                 `json:"answer"`
	Confidence     float64                `json:"confidence"`
	Citations      []Citation             `json:"citations"`
	Sources        []DocumentSource       `json:"sources"`
	ProcessingTime time.Duration          `json:"processing_time"`
	QueryID        string                 `json:"query_id"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// Citation represents a source citation
type Citation struct {
	DocumentID  string    `json:"document_id"`
	PageNum     int       `json:"page_num"`
	Content     string    `json:"content"`
	ContentType string    `json:"content_type"`
	BBox        []float64 `json:"bbox,omitempty"`
	Similarity  float64   `json:"similarity"`
	Relevance   float64   `json:"relevance"`
}

// DocumentSource represents a document source
type DocumentSource struct {
	DocumentID  string    `json:"document_id"`
	Filename    string    `json:"filename"`
	PageNum     int       `json:"page_num"`
	Content     string    `json:"content"`
	ContentType string    `json:"content_type"`
	UploadedAt  time.Time `json:"uploaded_at"`
	TenantID    string    `json:"tenant_id"`
}

// EnhancedSearchResult extends the base SearchResult with additional fields
type EnhancedSearchResult struct {
	ID              string    `json:"id"`
	DocumentID      string    `json:"document_id"`
	PageNum         int       `json:"page_num"`
	Content         string    `json:"content"`
	ContentType     string    `json:"content_type"`
	SimilarityScore float64   `json:"similarity_score"`
	BBox            []float64 `json:"bbox,omitempty"`
	Filename        string    `json:"filename"`
	TenantID        string    `json:"tenant_id"`
}

// NewEnhancedRAGService creates a new enhanced RAG service
func NewEnhancedRAGService(qdrantClient *QdrantClient, embedder *EmbedderService, llmClient *LLMClient) *EnhancedRAGService {
	return &EnhancedRAGService{
		qdrantClient: qdrantClient,
		embedder:     embedder,
		llmClient:    llmClient,
		reranker:     NewRerankerService(),
		cache:        NewEmbeddingCache(),
	}
}

// ProcessRAGQuery processes a RAG query and returns a comprehensive response
func (rag *EnhancedRAGService) ProcessRAGQuery(ctx context.Context, query RAGQuery) (*RAGResponse, error) {
	startTime := time.Now()
	queryID := generateQueryID()

	log.Printf("Processing RAG query: %s", queryID)

	// Step 1: Generate query embedding
	queryEmbedding, err := rag.generateOrRetrieveEmbedding(query.Question)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Step 2: Perform semantic search
	searchResults, err := rag.performSemanticSearch(ctx, queryEmbedding, query)
	if err != nil {
		return nil, fmt.Errorf("failed to perform semantic search: %w", err)
	}

	// Step 3: Rerank results
	rerankedResults, err := rag.reranker.Rerank(query.Question, searchResults)
	if err != nil {
		log.Printf("Warning: Reranking failed, using original results: %v", err)
		rerankedResults = searchResults
	}

	// Step 4: Compose context
	context := rag.composeContext(rerankedResults, query)

	// Step 5: Generate LLM response
	answer, err := rag.generateLLMResponse(ctx, query.Question, context)
	if err != nil {
		return nil, fmt.Errorf("failed to generate LLM response: %w", err)
	}

	// Step 6: Extract citations
	citations := rag.extractCitations(rerankedResults, answer)

	// Step 7: Calculate confidence
	confidence := rag.calculateConfidence(rerankedResults, citations)

	response := &RAGResponse{
		Answer:         answer,
		Confidence:     confidence,
		Citations:      citations,
		Sources:        rag.convertToSources(rerankedResults),
		ProcessingTime: time.Since(startTime),
		QueryID:        queryID,
		Metadata: map[string]interface{}{
			"total_sources":     len(rerankedResults),
			"reranking_applied": len(rerankedResults) != len(searchResults),
			"cache_hits":        rag.cache.GetStats()["hits"],
		},
	}

	log.Printf("RAG query completed in %v with confidence %.2f", response.ProcessingTime, confidence)
	return response, nil
}

// performSemanticSearch performs semantic search across multiple collections
func (rag *EnhancedRAGService) performSemanticSearch(ctx context.Context, queryEmbedding []float32, query RAGQuery) ([]EnhancedSearchResult, error) {
	var allResults []EnhancedSearchResult

	// Search in document chunks
	chunkRequest := SearchRequest{
		Vector:      queryEmbedding,
		Limit:       20,
		WithPayload: true,
		Filter: map[string]interface{}{
			"must": []map[string]interface{}{
				{"key": "tenant_id", "match": map[string]interface{}{"value": query.TenantID}},
			},
		},
	}

	chunkResults, err := rag.qdrantClient.Search(ctx, "jlrdi_document_chunks", chunkRequest)
	if err != nil {
		log.Printf("Warning: Document chunks search failed: %v", err)
	} else {
		allResults = append(allResults, rag.convertQdrantResults(chunkResults, "text_chunk")...)
	}

	// Search in table cells if requested
	if query.IncludeTables {
		tableRequest := SearchRequest{
			Vector:      queryEmbedding,
			Limit:       15,
			WithPayload: true,
			Filter: map[string]interface{}{
				"must": []map[string]interface{}{
					{"key": "tenant_id", "match": map[string]interface{}{"value": query.TenantID}},
				},
			},
		}

		tableResults, err := rag.qdrantClient.Search(ctx, "jlrdi_table_cells", tableRequest)
		if err != nil {
			log.Printf("Warning: Table cells search failed: %v", err)
		} else {
			allResults = append(allResults, rag.convertQdrantResults(tableResults, "table_cell")...)
		}
	}

	// Search in figures if requested
	if query.IncludeFigures {
		figureRequest := SearchRequest{
			Vector:      queryEmbedding,
			Limit:       10,
			WithPayload: true,
			Filter: map[string]interface{}{
				"must": []map[string]interface{}{
					{"key": "tenant_id", "match": map[string]interface{}{"value": query.TenantID}},
				},
			},
		}

		figureResults, err := rag.qdrantClient.Search(ctx, "jlrdi_figures", figureRequest)
		if err != nil {
			log.Printf("Warning: Figures search failed: %v", err)
		} else {
			allResults = append(allResults, rag.convertQdrantResults(figureResults, "figure")...)
		}
	}

	// Sort by similarity score
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].SimilarityScore > allResults[j].SimilarityScore
	})

	// Limit results
	if query.MaxResults > 0 && len(allResults) > query.MaxResults {
		allResults = allResults[:query.MaxResults]
	}

	return allResults, nil
}

// composeContext composes context from search results
func (rag *EnhancedRAGService) composeContext(results []EnhancedSearchResult, query RAGQuery) string {
	var contextBuilder strings.Builder

	contextBuilder.WriteString("Based on the following information:\n\n")

	for i, result := range results {
		contextBuilder.WriteString(fmt.Sprintf("Source %d:\n", i+1))
		contextBuilder.WriteString(fmt.Sprintf("Content: %s\n", result.Content))
		contextBuilder.WriteString(fmt.Sprintf("Type: %s\n", result.ContentType))
		contextBuilder.WriteString(fmt.Sprintf("Document: %s, Page: %d\n", result.DocumentID, result.PageNum))
		contextBuilder.WriteString(fmt.Sprintf("Relevance: %.2f\n\n", result.SimilarityScore))
	}

	contextBuilder.WriteString("Please answer the following question using only the information provided above:\n")
	contextBuilder.WriteString(query.Question)

	return contextBuilder.String()
}

// generateLLMResponse generates a response using the LLM
func (rag *EnhancedRAGService) generateLLMResponse(ctx context.Context, question, context string) (string, error) {
	// This is a placeholder - in production you'd integrate with OpenAI, Anthropic, etc.
	// For now, return a simulated response
	response := fmt.Sprintf("Based on the provided context, I can answer your question about '%s'. The information shows relevant details that address your query. Please refer to the specific citations for detailed information.", question)

	return response, nil
}

// extractCitations extracts citations from the LLM response
func (rag *EnhancedRAGService) extractCitations(results []EnhancedSearchResult, answer string) []Citation {
	var citations []Citation

	for _, result := range results {
		// Simple citation extraction - in production you'd use more sophisticated NLP
		if strings.Contains(strings.ToLower(answer), strings.ToLower(result.Content[:50])) {
			citation := Citation{
				DocumentID:  result.DocumentID,
				PageNum:     result.PageNum,
				Content:     result.Content,
				ContentType: result.ContentType,
				BBox:        result.BBox,
				Similarity:  result.SimilarityScore,
				Relevance:   result.SimilarityScore,
			}
			citations = append(citations, citation)
		}
	}

	return citations
}

// calculateConfidence calculates the overall confidence of the response
func (rag *EnhancedRAGService) calculateConfidence(results []EnhancedSearchResult, citations []Citation) float64 {
	if len(results) == 0 {
		return 0.0
	}

	// Calculate average similarity score
	var totalSimilarity float64
	for _, result := range results {
		totalSimilarity += result.SimilarityScore
	}
	avgSimilarity := totalSimilarity / float64(len(results))

	// Factor in citation coverage
	citationCoverage := float64(len(citations)) / float64(len(results))

	// Weighted confidence calculation
	confidence := (avgSimilarity * 0.7) + (citationCoverage * 0.3)

	return confidence
}

// convertToSources converts search results to document sources
func (rag *EnhancedRAGService) convertToSources(results []EnhancedSearchResult) []DocumentSource {
	var sources []DocumentSource

	for _, result := range results {
		source := DocumentSource{
			DocumentID:  result.DocumentID,
			Filename:    result.Filename,
			PageNum:     result.PageNum,
			Content:     result.Content,
			ContentType: result.ContentType,
			UploadedAt:  time.Now(), // This should come from the actual document metadata
			TenantID:    result.TenantID,
		}
		sources = append(sources, source)
	}

	return sources
}

// convertQdrantResults converts Qdrant search results to internal format
func (rag *EnhancedRAGService) convertQdrantResults(qdrantResults []SearchResult, contentType string) []EnhancedSearchResult {
	var results []EnhancedSearchResult

	for _, qr := range qdrantResults {
		// Extract payload fields safely
		docID, _ := qr.Payload["doc_id"].(string)
		pageNum, _ := qr.Payload["page_num"].(float64)
		content, _ := qr.Payload["content"].(string)
		filename, _ := qr.Payload["filename"].(string)
		tenantID, _ := qr.Payload["tenant_id"].(string)

		result := EnhancedSearchResult{
			ID:              qr.ID,
			DocumentID:      docID,
			PageNum:         int(pageNum),
			Content:         content,
			ContentType:     contentType,
			SimilarityScore: qr.Score,
			BBox:            parseBBox(qr.Payload["bbox"]),
			Filename:        filename,
			TenantID:        tenantID,
		}
		results = append(results, result)
	}

	return results
}

// generateOrRetrieveEmbedding generates a new embedding or retrieves from cache
func (rag *EnhancedRAGService) generateOrRetrieveEmbedding(content string) ([]float32, error) {
	// Check cache first
	if cached := rag.cache.Get(content); cached != nil {
		// Convert []float64 to []float32
		result := make([]float32, len(cached))
		for i, v := range cached {
			result[i] = float32(v)
		}
		return result, nil
	}

	// Generate new embedding
	embeddings, err := rag.embedder.Embed(context.Background(), []string{content})
	if err != nil {
		return nil, err
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding generated")
	}

	// Cache the embedding (convert to float64 for storage)
	embedding64 := make([]float64, len(embeddings[0]))
	for i, v := range embeddings[0] {
		embedding64[i] = float64(v)
	}
	rag.cache.Set(content, embedding64)

	return embeddings[0], nil
}

// Helper functions
func generateQueryID() string {
	return fmt.Sprintf("rag_%s", uuid.New().String()[:8])
}

func parseBBox(bbox interface{}) []float64 {
	if bbox == nil {
		return nil
	}

	// Handle different bbox formats
	switch v := bbox.(type) {
	case []float64:
		return v
	case []interface{}:
		result := make([]float64, len(v))
		for i, val := range v {
			if f, ok := val.(float64); ok {
				result[i] = f
			}
		}
		return result
	default:
		return nil
	}
}

// NewRerankerService creates a new reranker service
func NewRerankerService() *RerankerService {
	return &RerankerService{
		model: "cross-encoder/ms-marco-MiniLM-L-6-v2",
	}
}

// Rerank reranks search results based on relevance to the query
func (r *RerankerService) Rerank(query string, results []EnhancedSearchResult) ([]EnhancedSearchResult, error) {
	// This is a placeholder - in production you'd use a proper reranking model
	// For now, we'll just return the results as-is
	return results, nil
}

// NewEmbeddingCache creates a new embedding cache
func NewEmbeddingCache() *EmbeddingCache {
	return &EmbeddingCache{
		embeddings: make(map[string][]float64),
		metadata:   make(map[string]EmbeddingMetadata),
	}
}

// Get retrieves an embedding from cache
func (c *EmbeddingCache) Get(content string) []float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if embedding, exists := c.embeddings[content]; exists {
		// Update usage statistics
		if metadata, ok := c.metadata[content]; ok {
			metadata.LastUsed = time.Now()
			metadata.UsageCount++
			c.metadata[content] = metadata
		}
		return embedding
	}

	return nil
}

// Set stores an embedding in cache
func (c *EmbeddingCache) Set(content string, embedding []float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.embeddings[content] = embedding
	c.metadata[content] = EmbeddingMetadata{
		Content:     content,
		ContentType: "text",
		CreatedAt:   time.Now(),
		LastUsed:    time.Now(),
		UsageCount:  1,
	}
}

// GetStats returns cache statistics
func (c *EmbeddingCache) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var hits, misses int
	var totalUsage int

	for _, metadata := range c.metadata {
		totalUsage += metadata.UsageCount
	}

	return map[string]interface{}{
		"total_embeddings": len(c.embeddings),
		"total_usage":      totalUsage,
		"hits":             hits,
		"misses":           misses,
	}
}

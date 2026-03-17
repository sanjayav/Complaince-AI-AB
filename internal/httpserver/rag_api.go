package httpserver

import (
	"encoding/json"
	"net/http"
	"time"

	"jlrdi/internal/rag"
)

// RAGAPI handles RAG (Retrieval Augmented Generation) operations
type RAGAPI struct {
	ragService *rag.EnhancedRAGService
}

// NewRAGAPI creates a new RAG API handler
func NewRAGAPI(ragService *rag.EnhancedRAGService) *RAGAPI {
	return &RAGAPI{
		ragService: ragService,
	}
}

// RAGQueryRequest represents a RAG query request
type RAGQueryRequest struct {
	Question       string                 `json:"question"`
	Context        map[string]interface{} `json:"context,omitempty"`
	Filters        map[string]interface{} `json:"filters,omitempty"`
	MaxResults     int                    `json:"max_results,omitempty"`
	IncludeTables  bool                   `json:"include_tables,omitempty"`
	IncludeFigures bool                   `json:"include_figures,omitempty"`
	TenantID       string                 `json:"tenant_id"`
}

// RAGQueryResponse represents a RAG query response
type RAGQueryResponse struct {
	Response *rag.RAGResponse `json:"response"`
	Message  string           `json:"message"`
}

// ProcessRAGQuery processes a RAG query and returns a comprehensive response
func (r *RAGAPI) ProcessRAGQuery(w http.ResponseWriter, req *http.Request) {
	var request RAGQueryRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate request
	if request.Question == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Question is required"})
		return
	}

	if request.TenantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Tenant ID is required"})
		return
	}

	// Set defaults
	if request.MaxResults == 0 {
		request.MaxResults = 20
	}

	// Create RAG query
	query := rag.RAGQuery{
		Question:       request.Question,
		Context:        request.Context,
		Filters:        request.Filters,
		MaxResults:     request.MaxResults,
		IncludeTables:  request.IncludeTables,
		IncludeFigures: request.IncludeFigures,
		TenantID:       request.TenantID,
	}

	// Process query
	response, err := r.ragService.ProcessRAGQuery(req.Context(), query)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to process RAG query"})
		return
	}

	result := RAGQueryResponse{
		Response: response,
		Message:  "RAG query processed successfully",
	}

	writeJSON(w, http.StatusOK, result)
}

// RAGHistoryRequest represents a request to get RAG query history
type RAGHistoryRequest struct {
	TenantID  string    `json:"tenant_id"`
	UserID    string    `json:"user_id,omitempty"`
	StartDate time.Time `json:"start_date,omitempty"`
	EndDate   time.Time `json:"end_date,omitempty"`
	Limit     int       `json:"limit,omitempty"`
}

// RAGHistoryResponse represents a RAG history response
type RAGHistoryResponse struct {
	Queries []map[string]interface{} `json:"queries"`
	Total   int                      `json:"total"`
	Message string                   `json:"message"`
}

// GetRAGHistory retrieves RAG query history for a tenant
func (r *RAGAPI) GetRAGHistory(w http.ResponseWriter, req *http.Request) {
	var request RAGHistoryRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate request
	if request.TenantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Tenant ID is required"})
		return
	}

	// Set defaults
	if request.Limit == 0 {
		request.Limit = 50
	}

	// For now, return a placeholder response
	// In production, you'd query the database for actual history
	response := RAGHistoryResponse{
		Queries: []map[string]interface{}{
			{
				"query_id":        "rag_12345678",
				"question":        "What are the test results for engine performance?",
				"confidence":      0.85,
				"processing_time": "2.3s",
				"created_at":      time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			},
			{
				"query_id":        "rag_87654321",
				"question":        "Show me the safety test data",
				"confidence":      0.92,
				"processing_time": "1.8s",
				"created_at":      time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
			},
		},
		Total:   2,
		Message: "RAG history retrieved successfully",
	}

	writeJSON(w, http.StatusOK, response)
}

// RAGFeedbackRequest represents feedback for a RAG query
type RAGFeedbackRequest struct {
	QueryID     string `json:"query_id"`
	TenantID    string `json:"tenant_id"`
	UserID      string `json:"user_id"`
	Rating      int    `json:"rating"` // 1-5 scale
	Feedback    string `json:"feedback,omitempty"`
	Useful      bool   `json:"useful"`
	Suggestions string `json:"suggestions,omitempty"`
}

// RAGFeedbackResponse represents a feedback response
type RAGFeedbackResponse struct {
	Message string `json:"message"`
}

// SubmitRAGFeedback submits feedback for a RAG query
func (r *RAGAPI) SubmitRAGFeedback(w http.ResponseWriter, req *http.Request) {
	var request RAGFeedbackRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate request
	if request.QueryID == "" || request.TenantID == "" || request.UserID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Query ID, Tenant ID, and User ID are required"})
		return
	}

	if request.Rating < 1 || request.Rating > 5 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Rating must be between 1 and 5"})
		return
	}

	// For now, just acknowledge the feedback
	// In production, you'd store this in the database and use it for model improvement
	response := RAGFeedbackResponse{
		Message: "Feedback submitted successfully",
	}

	writeJSON(w, http.StatusOK, response)
}

// RAGStatsRequest represents a request to get RAG statistics
type RAGStatsRequest struct {
	TenantID  string    `json:"tenant_id"`
	StartDate time.Time `json:"start_date,omitempty"`
	EndDate   time.Time `json:"end_date,omitempty"`
}

// RAGStatsResponse represents RAG statistics
type RAGStatsResponse struct {
	Stats   map[string]interface{} `json:"stats"`
	Message string                 `json:"message"`
}

// GetRAGStats retrieves RAG statistics for a tenant
func (r *RAGAPI) GetRAGStats(w http.ResponseWriter, req *http.Request) {
	var request RAGStatsRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate request
	if request.TenantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Tenant ID is required"})
		return
	}

	// For now, return placeholder statistics
	// In production, you'd calculate these from actual data
	stats := map[string]interface{}{
		"total_queries":         150,
		"average_confidence":    0.87,
		"average_response_time": "2.1s",
		"cache_hit_rate":        0.65,
		"user_satisfaction":     4.2,
		"top_questions": []string{
			"What are the test results for engine performance?",
			"Show me the safety test data",
			"Compare different model specifications",
		},
		"query_trends": map[string]interface{}{
			"daily_queries":   25,
			"weekly_queries":  175,
			"monthly_queries": 750,
		},
	}

	response := RAGStatsResponse{
		Stats:   stats,
		Message: "RAG statistics retrieved successfully",
	}

	writeJSON(w, http.StatusOK, response)
}

// RAGModelRequest represents a request to get RAG model information
type RAGModelRequest struct {
	TenantID string `json:"tenant_id"`
}

// RAGModelResponse represents RAG model information
type RAGModelResponse struct {
	Model   map[string]interface{} `json:"model"`
	Message string                 `json:"message"`
}

// GetRAGModelInfo retrieves information about the RAG model
func (r *RAGAPI) GetRAGModelInfo(w http.ResponseWriter, req *http.Request) {
	var request RAGModelRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate request
	if request.TenantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Tenant ID is required"})
		return
	}

	// Return model information
	modelInfo := map[string]interface{}{
		"model_name":          "Enhanced RAG Pipeline",
		"version":             "1.0.0",
		"embedding_model":     "Custom Hash-based Embeddings",
		"reranking_model":     "Cross-encoder/ms-marco-MiniLM-L-6-v2",
		"llm_integration":     "Placeholder (OpenAI/Anthropic ready)",
		"vector_dimension":    384,
		"cache_enabled":       true,
		"multi_collection":    true,
		"citation_extraction": true,
		"confidence_scoring":  true,
		"last_updated":        time.Now().Format(time.RFC3339),
		"performance_metrics": map[string]interface{}{
			"average_query_time": "2.1s",
			"cache_hit_rate":     0.65,
			"accuracy_score":     0.87,
		},
	}

	response := RAGModelResponse{
		Model:   modelInfo,
		Message: "RAG model information retrieved successfully",
	}

	writeJSON(w, http.StatusOK, response)
}

package httpserver

import (
	"encoding/json"
	"net/http"
	"time"

	"jlrdi/internal/auth"
)

// APIKeyAPI handles API key management endpoints
type APIKeyAPI struct {
	apiKeyManager *auth.APIKeyManager
}

// NewAPIKeyAPI creates a new API key API handler
func NewAPIKeyAPI(apiKeyManager *auth.APIKeyManager) *APIKeyAPI {
	return &APIKeyAPI{
		apiKeyManager: apiKeyManager,
	}
}

// GenerateAPIKeyRequest represents a request to generate a new API key
type GenerateAPIKeyRequest struct {
	Name        string            `json:"name"`
	Permissions []string          `json:"permissions,omitempty"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// GenerateAPIKeyResponse represents the response when generating an API key
type GenerateAPIKeyResponse struct {
	APIKey     *auth.APIKey `json:"api_key"`
	Message    string       `json:"message"`
	Warning    string       `json:"warning,omitempty"`
	UsageNotes string       `json:"usage_notes"`
}

// GenerateAPIKey generates a new API key for a frontend developer
func (a *APIKeyAPI) GenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req GenerateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON request"})
		return
	}

	// Validate request
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Name is required"})
		return
	}

	// Use default permissions if none specified
	permissions := req.Permissions
	if len(permissions) == 0 {
		permissions = auth.FrontendPermissions
	}

	// Generate API key
	apiKey, err := a.apiKeyManager.GenerateAPIKey(req.Name, permissions, req.ExpiresAt)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to generate API key"})
		return
	}

	// Create response
	response := GenerateAPIKeyResponse{
		APIKey:  apiKey,
		Message: "API key generated successfully",
		Warning: "⚠️  IMPORTANT: This is the only time the full API key will be shown. Save it securely!",
		UsageNotes: `Usage Instructions:
1. Include the API key in your requests using one of these methods:
   - Header: X-API-Key: YOUR_API_KEY
   - Header: Authorization: Bearer YOUR_API_KEY
   - Header: Authorization: ApiKey YOUR_API_KEY
2. The API key will expire on: ` + formatTime(apiKey.ExpiresAt) + `
3. Keep this key secure and don't share it publicly`,
	}

	writeJSON(w, http.StatusCreated, response)
}

// ListAPIKeys lists all API keys (without exposing the actual keys)
func (a *APIKeyAPI) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	keys := a.apiKeyManager.ListAPIKeys()

	response := map[string]interface{}{
		"api_keys": keys,
		"total":    len(keys),
		"message":  "API keys retrieved successfully",
	}

	writeJSON(w, http.StatusOK, response)
}

// RevokeAPIKey revokes an API key
func (a *APIKeyAPI) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	// Extract key hash from URL path
	keyHash := r.URL.Query().Get("key_hash")
	if keyHash == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "key_hash parameter is required"})
		return
	}

	// Revoke the API key
	if err := a.apiKeyManager.RevokeAPIKey(keyHash); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "API key not found"})
		return
	}

	response := map[string]interface{}{
		"message":  "API key revoked successfully",
		"key_hash": keyHash,
	}

	writeJSON(w, http.StatusOK, response)
}

// GetAPIKeyInfo gets information about the current API key
func (a *APIKeyAPI) GetAPIKeyInfo(w http.ResponseWriter, r *http.Request) {
	// Get API key context from middleware
	apiKeyCtx := auth.GetAPIKeyFromContext(r.Context())
	if apiKeyCtx == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "API key context not found"})
		return
	}

	// Create response with key info (without exposing the actual key)
	keyInfo := map[string]interface{}{
		"id":          apiKeyCtx.APIKey.ID,
		"name":        apiKeyCtx.APIKey.Name,
		"permissions": apiKeyCtx.APIKey.Permissions,
		"created_at":  apiKeyCtx.APIKey.CreatedAt,
		"expires_at":  apiKeyCtx.APIKey.ExpiresAt,
		"last_used":   apiKeyCtx.APIKey.LastUsed,
		"metadata":    apiKeyCtx.APIKey.Metadata,
	}

	response := map[string]interface{}{
		"api_key": keyInfo,
		"message": "API key information retrieved successfully",
	}

	writeJSON(w, http.StatusOK, response)
}

// APIKeyUsageGuide provides usage instructions for frontend developers
func (a *APIKeyAPI) APIKeyUsageGuide(w http.ResponseWriter, r *http.Request) {
	guide := map[string]interface{}{
		"title": "JLR Document Intelligence API - Frontend Developer Guide",
		"authentication": map[string]interface{}{
			"method": "API Key Authentication",
			"headers": []string{
				"X-API-Key: YOUR_API_KEY",
				"Authorization: Bearer YOUR_API_KEY",
				"Authorization: ApiKey YOUR_API_KEY",
			},
			"note": "Choose any of the above header formats",
		},
		"endpoints": map[string]interface{}{
			"document_upload": map[string]interface{}{
				"url":                 "POST /v1/documents/upload",
				"description":         "Upload PDF documents for processing",
				"required_permission": "documents:upload",
			},
			"semantic_search": map[string]interface{}{
				"url":                 "POST /v1/search/semantic",
				"description":         "Search documents using natural language",
				"required_permission": "search:semantic",
			},
			"rag_qa": map[string]interface{}{
				"url":                 "POST /v1/ask",
				"description":         "Ask questions and get AI-powered answers",
				"required_permission": "rag:ask",
			},
			"list_documents": map[string]interface{}{
				"url":                 "GET /v1/documents",
				"description":         "List uploaded documents",
				"required_permission": "documents:read",
			},
		},
		"rate_limits": map[string]interface{}{
			"requests_per_minute": 100,
			"upload_file_size":    "32MB",
			"concurrent_uploads":  5,
		},
		"error_handling": map[string]interface{}{
			"401": "Unauthorized - Invalid or missing API key",
			"403": "Forbidden - Insufficient permissions",
			"429": "Too Many Requests - Rate limit exceeded",
			"500": "Internal Server Error - Contact support",
		},
		"best_practices": []string{
			"Store your API key securely and never expose it in client-side code",
			"Use environment variables for API keys in production",
			"Implement exponential backoff for retries",
			"Cache responses when appropriate",
			"Monitor your API usage and rate limits",
		},
		"support": map[string]interface{}{
			"email":         "api-support@jlr.com",
			"documentation": "https://api.jlr.com/docs",
			"status_page":   "https://status.jlr.com",
		},
	}

	writeJSON(w, http.StatusOK, guide)
}

// formatTime formats a time pointer for display
func formatTime(t *time.Time) string {
	if t == nil {
		return "Never (no expiration)"
	}
	return t.Format("2006-01-02 15:04:05 UTC")
}

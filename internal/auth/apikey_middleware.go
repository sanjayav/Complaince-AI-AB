package auth

import (
	"context"
	"net/http"
	"strings"
)

// APIKeyContextKey is the context key for API key information
type APIKeyContextKey struct{}

// APIKeyContext contains API key information for the request
type APIKeyContext struct {
	APIKey      *APIKey
	Permissions []string
}

// RequireAPIKey middleware validates API keys for frontend requests
func RequireAPIKey(apiKeyManager *APIKeyManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip API key check for health and public endpoints
			if isPublicEndpoint(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Extract API key from header
			apiKey := extractAPIKey(r)
			if apiKey == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"API key required"}`))
				return
			}

			// Validate API key
			keyInfo, err := apiKeyManager.ValidateAPIKey(apiKey)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"Invalid API key"}`))
				return
			}

			// Add API key context to request
			ctx := context.WithValue(r.Context(), APIKeyContextKey{}, &APIKeyContext{
				APIKey:      keyInfo,
				Permissions: keyInfo.Permissions,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission middleware checks if the API key has required permissions
func RequirePermission(requiredPermission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get API key context
			apiKeyCtx, ok := r.Context().Value(APIKeyContextKey{}).(*APIKeyContext)
			if !ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"API key context not found"}`))
				return
			}

			// Check if user has required permission
			if !hasPermission(apiKeyCtx.Permissions, requiredPermission) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"error":"Insufficient permissions"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractAPIKey extracts the API key from the request
func extractAPIKey(r *http.Request) string {
	// Check Authorization header (Bearer token)
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
		if strings.HasPrefix(authHeader, "ApiKey ") {
			return strings.TrimPrefix(authHeader, "ApiKey ")
		}
	}

	// Check X-API-Key header
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return apiKey
	}

	// Check query parameter (for testing purposes)
	if apiKey := r.URL.Query().Get("api_key"); apiKey != "" {
		return apiKey
	}

	return ""
}

// hasPermission checks if the user has the required permission
func hasPermission(userPermissions []string, requiredPermission string) bool {
	for _, permission := range userPermissions {
		if permission == requiredPermission {
			return true
		}
		// Check wildcard permissions
		if permission == "*" || permission == "admin:manage" {
			return true
		}
	}
	return false
}

// isPublicEndpoint checks if the endpoint is public (no API key required)
func isPublicEndpoint(path string) bool {
	publicPaths := []string{
		"/v1/health",
		"/v1/metrics",
		"/docs",
		"/swagger",
		"/v1/auth",
	}

	for _, publicPath := range publicPaths {
		if strings.HasPrefix(path, publicPath) {
			return true
		}
	}

	return false
}

// GetAPIKeyFromContext extracts API key information from the request context
func GetAPIKeyFromContext(ctx context.Context) *APIKeyContext {
	if apiKeyCtx, ok := ctx.Value(APIKeyContextKey{}).(*APIKeyContext); ok {
		return apiKeyCtx
	}
	return nil
}

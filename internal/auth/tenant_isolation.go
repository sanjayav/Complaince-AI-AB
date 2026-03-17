package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RequireTenantAccess ensures the user can only access their own tenant's data
func RequireTenantAccess() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from context
			user := UserFromContext(r.Context())
			if user.Subject == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Extract tenant ID from request (URL parameter, form, or body)
			requestedTenantID := extractTenantID(r)
			if requestedTenantID == "" {
				http.Error(w, "tenant ID required", http.StatusBadRequest)
				return
			}

			// Validate tenant access
			if user.TenantID != requestedTenantID {
				http.Error(w, "access denied: tenant mismatch", http.StatusForbidden)
				return
			}

			// Add tenant ID to context for downstream handlers
			ctx := context.WithValue(r.Context(), "requested_tenant_id", requestedTenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireTenantAccessWithParam ensures tenant access using a specific URL parameter
func RequireTenantAccessWithParam(paramName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from context
			user := UserFromContext(r.Context())
			if user.Subject == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Extract tenant ID from URL parameter
			requestedTenantID := chi.URLParam(r, paramName)
			if requestedTenantID == "" {
				http.Error(w, fmt.Sprintf("tenant ID parameter '%s' required", paramName), http.StatusBadRequest)
				return
			}

			// Validate tenant access
			if user.TenantID != requestedTenantID {
				http.Error(w, "access denied: tenant mismatch", http.StatusForbidden)
				return
			}

			// Add tenant ID to context for downstream handlers
			ctx := context.WithValue(r.Context(), "requested_tenant_id", requestedTenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireTenantAccessWithForm ensures tenant access using form data
func RequireTenantAccessWithForm() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from context
			user := UserFromContext(r.Context())
			if user.Subject == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Parse form if not already parsed
			if err := r.ParseForm(); err != nil {
				http.Error(w, "failed to parse form", http.StatusBadRequest)
				return
			}

			// Extract tenant ID from form
			requestedTenantID := r.FormValue("tenant_id")
			if requestedTenantID == "" {
				http.Error(w, "tenant_id form field required", http.StatusBadRequest)
				return
			}

			// Validate tenant access
			if user.TenantID != requestedTenantID {
				http.Error(w, "access denied: tenant mismatch", http.StatusForbidden)
				return
			}

			// Add tenant ID to context for downstream handlers
			ctx := context.WithValue(r.Context(), "requested_tenant_id", requestedTenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractTenantID extracts tenant ID from various request sources
func extractTenantID(r *http.Request) string {
	// Try URL parameter first
	if tenantID := chi.URLParam(r, "id"); tenantID != "" {
		return tenantID
	}

	// Try form data
	if tenantID := r.FormValue("tenant_id"); tenantID != "" {
		return tenantID
	}

	// Try header
	if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
		return tenantID
	}

	return ""
}

// GetRequestedTenantID gets the validated tenant ID from context
func GetRequestedTenantID(ctx context.Context) string {
	if tenantID, ok := ctx.Value("requested_tenant_id").(string); ok {
		return tenantID
	}
	return ""
}

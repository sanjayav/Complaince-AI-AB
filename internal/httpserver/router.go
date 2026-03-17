package httpserver

import (
	"net/http"
	"strings"
	"time"

	"jlrdi/internal/auth"
	"jlrdi/internal/config"
	"jlrdi/internal/rag"
	"jlrdi/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/cors"
)

// Deps contains all dependencies needed by the router
type Deps struct {
	DB           *pgxpool.Pool
	Signer       *storage.Signer
	Config       config.Config
	QdrantURL    string
	HTTPTimeout  time.Duration
	QdrantClient *rag.QdrantClient
	Embedder     *rag.Embedder
}

// NewRouter creates a new router with all middleware and routes
func NewRouter(deps Deps) http.Handler {
	// Create API key manager
	apiKeyManager := auth.NewAPIKeyManager()

	// Create tenant manager
	tenantManager := auth.NewTenantManager()

	// Create S3 service with proper client for AWS S3
	var tenantS3Service *storage.TenantS3Service
	_, err := storage.NewS3Service("jlr-doc-intel-bucket-us", "us-east-1")
	if err != nil {
		// Fallback to nil client for development
		tenantS3Service = storage.NewTenantS3Service(nil, "jlr-doc-intel-bucket-us", "us-east-1")
	} else {
		// For now, use nil to avoid accessing unexported field
		// In production, you'd want to properly extract the S3 client
		tenantS3Service = storage.NewTenantS3Service(nil, "jlr-doc-intel-bucket-us", "us-east-1")
	}

	// Create RAG service
	llmClient := &rag.LLMClient{
		APIKey:      "placeholder",
		BaseURL:     "https://api.openai.com",
		Model:       "gpt-4",
		MaxTokens:   4096,
		Temperature: 0.7,
	}

	// Create embedder service
	embedderService := rag.NewEmbedderService()
	ragService := rag.NewEnhancedRAGService(deps.QdrantClient, embedderService, llmClient)

	// Create authentication service
	authService := auth.NewAuthService(deps.Config.JWTSecret, tenantManager)

	// Create API handlers
	api := &API{deps: deps}
	apiKeyAPI := NewAPIKeyAPI(apiKeyManager)
	tenantAPI := NewTenantAPI(tenantManager)
	documentAPI := NewDocumentAPI(tenantS3Service, deps.DB)
	ragAPI := NewRAGAPI(ragService)
	authAPI := NewAuthAPI(authService)

	// Create router
	r := chi.NewRouter()

	// Configure CORS properly
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"}, // Local frontend
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH", "HEAD"},
		AllowedHeaders:   []string{"*"}, // Allow all headers
		ExposedHeaders:   []string{"*"}, // Expose all headers
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
		Debug:            true,  // Enable CORS debugging
		AllowOriginFunc: func(origin string) bool {
			// Allow any ngrok domain and localhost:5173
			if origin == "http://localhost:5173" {
				return true
			}
			// Match https ngrok domains
			if strings.HasPrefix(origin, "https://") {
				// Simple suffix checks for ngrok domains
				return strings.HasSuffix(origin, ".ngrok-free.app") || strings.HasSuffix(origin, ".ngrok.app") || strings.HasSuffix(origin, ".ngrok.io")
			}
			return false
		},
	})

	// Apply CORS middleware first
	r.Use(corsMiddleware.Handler)

	// Standard middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Health check (public)
	r.Get("/v1/health", api.Health)

	// Authentication endpoints (public - no API key required)
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login-auto", authAPI.LoginAutoTenant) // Auto-detect tenant from email
		r.Post("/register", authAPI.Register)
		r.Post("/logout", authAPI.Logout)
		r.Post("/refresh", authAPI.RefreshToken)
	})

	// Protected v1 routes (require API key)
	r.Route("/v1", func(r chi.Router) {
		// Require JWT for all /v1 routes (no API key needed for general endpoints)
		r.Use(auth.RequireAuth(deps.Config.JWTSecret))

		// API key management endpoints (admin only, require API key)
		r.Route("/admin", func(r chi.Router) {
			r.Use(auth.RequireAPIKey(apiKeyManager))
			r.Use(auth.RequirePermission("admin:manage"))
			r.Post("/apikeys", apiKeyAPI.GenerateAPIKey)
			r.Get("/apikeys", apiKeyAPI.ListAPIKeys)
			r.Delete("/apikeys", apiKeyAPI.RevokeAPIKey)
		})

		// API key information and guide (still require API key)
		r.Route("/apikey", func(r chi.Router) {
			r.Use(auth.RequireAPIKey(apiKeyManager))
			r.Get("/info", apiKeyAPI.GetAPIKeyInfo)
			r.Get("/guide", apiKeyAPI.APIKeyUsageGuide)
		})

		// User information (protected)
		r.Get("/me", authAPI.GetCurrentUser)
		r.Get("/tenant/users", tenantAPI.ListMyTenantUsers)

		// Tenant management endpoints
		r.Route("/tenants", func(r chi.Router) {
			r.Post("/", tenantAPI.CreateTenant)
			r.Get("/{id}", tenantAPI.GetTenant)
			r.Get("/", tenantAPI.ListTenants)
			r.Route("/{id}/users", func(r chi.Router) {
				r.Use(auth.RequireTenantAccessWithParam("id")) // ← SECURITY: Tenant isolation
				r.Post("/", tenantAPI.CreateTenantUser)
				r.Get("/", tenantAPI.ListTenantUsers)
				r.Put("/{uid}", tenantAPI.UpdateTenantUser)
				r.Delete("/{uid}", tenantAPI.DeleteTenantUser)
			})
			r.Get("/{id}/stats", tenantAPI.GetTenantStats)
		})

		// Enhanced document management endpoints
		r.Route("/tenants/{id}/documents", func(r chi.Router) {
			r.Use(auth.RequireTenantAccessWithParam("id")) // ← SECURITY: Tenant isolation
			r.Post("/upload", documentAPI.UploadDocument)
			r.Post("/batch", documentAPI.BatchUploadDocuments)
			r.Get("/", documentAPI.ListDocuments)
			r.Get("/{did}", documentAPI.GetDocument)
			r.Put("/{did}", documentAPI.UpdateDocumentStatus)
			r.Delete("/{did}", documentAPI.DeleteDocument)
			r.Get("/{did}/presigned-url", documentAPI.GetDocumentPresignedURL)
			r.Get("/stats", documentAPI.GetTenantStorageStats)
		})

		// JWT-tenant convenience endpoints
		r.Route("/documents", func(r chi.Router) {
			r.Post("/upload", documentAPI.UploadMyDocument)
			r.Get("/mine", documentAPI.ListMyDocuments)
			r.Delete("/{did}", documentAPI.DeleteMyDocument)
		})

		// RAG endpoints
		r.Route("/rag", func(r chi.Router) {
			r.Post("/query", ragAPI.ProcessRAGQuery)
			r.Post("/history", ragAPI.GetRAGHistory)
			r.Post("/feedback", ragAPI.SubmitRAGFeedback)
			r.Post("/stats", ragAPI.GetRAGStats)
			r.Post("/model-info", ragAPI.GetRAGModelInfo)
		})

		// Legacy endpoints (for backward compatibility) - REMOVED to fix route conflicts
		// r.Post("/documents/upload", api.UploadDocument)  // CONFLICT: Use /v1/documents/upload instead
		r.Get("/documents", api.ListDocuments)
		r.Get("/documents/{docId}/manifest", api.DocumentManifest)

		// Search and RAG endpoints
		r.Post("/search/semantic", api.SemanticSearch)
		r.Post("/ask", api.Ask)
		r.Get("/answers/{id}", api.GetAnswer)

		// Highlighting and evidence
		r.Post("/highlight/signed-urls", api.HighlightSignedURLs)
		r.Get("/evidence/{type}/{id}", api.GetEvidence)

		// Export functionality
		r.Post("/export", api.Export)

		// Indexing
		r.Post("/index/enqueue", api.IndexEnqueue)

		// QA endpoints
		r.Get("/qa/tasks", api.QAPendingTasks)
		r.Post("/qa/approve", api.QAApprove)
		r.Post("/qa/reject", api.QAReject)

		// User information
		r.Get("/me", api.Me)
	})

	// Development endpoints (only in dev environment)
	if deps.Config.Env == "dev" {
		r.Route("/dev", func(r chi.Router) {
			r.Use(auth.DevUser())
			r.Post("/seed", api.SeedData)
			r.Post("/clear", api.ClearData)
			r.Post("/tenants", tenantAPI.CreateTenantDev) // Development tenant creation
			r.Route("/tenants/{id}/users", func(r chi.Router) {
				r.Post("/", tenantAPI.CreateTenantUserDev) // Development user creation
				r.Get("/", tenantAPI.ListTenantUsersDev)   // Development user listing
			})
			r.Post("/apikeys", apiKeyAPI.GenerateAPIKey) // Development API key generation
			r.Get("/me", authAPI.GetCurrentUserDev)      // Development /me endpoint

			// Security test endpoint - uses actual tenant isolation middleware
			r.Route("/secure-test/{id}", func(r chi.Router) {
				r.Use(auth.RequireTenantAccessWithParam("id")) // ← ACTUAL SECURITY
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					tenantID := auth.GetRequestedTenantID(r.Context())
					user := auth.UserFromContext(r.Context())
					writeJSON(w, http.StatusOK, map[string]interface{}{
						"message":             "Security test passed!",
						"user_tenant_id":      user.TenantID,
						"requested_tenant_id": tenantID,
						"access_granted":      true,
						"timestamp":           time.Now().UTC().Format(time.RFC3339),
					})
				})
			})

			// JWT-only security test endpoint (bypasses DevUser middleware)
			r.Route("/jwt-test/{id}", func(r chi.Router) {
				r.Use(auth.RequireAuth(deps.Config.JWTSecret)) // ← REAL JWT AUTH
				r.Use(auth.RequireTenantAccessWithParam("id")) // ← TENANT ISOLATION
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					tenantID := auth.GetRequestedTenantID(r.Context())
					user := auth.UserFromContext(r.Context())
					writeJSON(w, http.StatusOK, map[string]interface{}{
						"message":             "JWT Security test passed!",
						"user_tenant_id":      user.TenantID,
						"requested_tenant_id": tenantID,
						"access_granted":      true,
						"timestamp":           time.Now().UTC().Format(time.RFC3339),
					})
				})
			})

			// Debug endpoint to see JWT token content
			r.Route("/debug-jwt/{id}", func(r chi.Router) {
				r.Use(auth.RequireAuth(deps.Config.JWTSecret)) // ← REAL JWT AUTH
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					user := auth.UserFromContext(r.Context())
					requestedTenantID := chi.URLParam(r, "id")
					writeJSON(w, http.StatusOK, map[string]interface{}{
						"message":             "JWT Debug Info",
						"user":                user,
						"requested_tenant_id": requestedTenantID,
						"user_tenant_id":      user.TenantID,
						"tenant_match":        user.TenantID == requestedTenantID,
						"timestamp":           time.Now().UTC().Format(time.RFC3339),
					})
				})
			})
		})

		// JWT-only test route (completely bypasses DevUser middleware)
		r.Route("/jwt-only/{id}", func(r chi.Router) {
			r.Use(auth.RequireAuth(deps.Config.JWTSecret)) // ← ONLY JWT AUTH
			r.Use(auth.RequireTenantAccessWithParam("id")) // ← TENANT ISOLATION
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				tenantID := auth.GetRequestedTenantID(r.Context())
				user := auth.UserFromContext(r.Context())
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"message":             "JWT-Only Security Test Passed!",
					"user_tenant_id":      user.TenantID,
					"requested_tenant_id": tenantID,
					"access_granted":      true,
					"timestamp":           time.Now().UTC().Format(time.RFC3339),
				})
			})
		})
	}

	return r
}

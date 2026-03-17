package httpserver

import (
	"encoding/json"
	"net/http"
	"time"

	"jlrdi/internal/auth"
)

// AuthAPI handles authentication-related endpoints
type AuthAPI struct {
	authService *auth.AuthService
}

// NewAuthAPI creates a new authentication API handler
func NewAuthAPI(authService *auth.AuthService) *AuthAPI {
	return &AuthAPI{
		authService: authService,
	}
}

// Login handles user authentication
func (api *AuthAPI) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}

	var response *auth.LoginResponse
	var err error

	// If tenant_id is provided, use traditional login
	if req.TenantID != "" {
		response, err = api.authService.AuthenticateUser(req.Email, req.Password, req.TenantID)
	} else {
		// Auto-detect tenant from email domain
		response, err = api.authService.AuthenticateUserAutoTenant(req.Email, req.Password)
	}

	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	// Return successful login response
	writeJSON(w, http.StatusOK, response)
}

// LoginAutoTenant handles user authentication with automatic tenant detection
func (api *AuthAPI) LoginAutoTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}

	// Auto-detect tenant from email domain
	response, err := api.authService.AuthenticateUserAutoTenant(req.Email, req.Password)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	// Return successful login response
	writeJSON(w, http.StatusOK, response)
}

// Logout handles user logout
func (api *AuthAPI) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.SessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "session_id is required"})
		return
	}

	// Logout user
	if err := api.authService.LogoutUser(req.SessionID); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "successfully logged out"})
}

// Register handles user registration
func (api *AuthAPI) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req struct {
		TenantID  string `json:"tenant_id"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Validate required fields
	if req.TenantID == "" || req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "tenant_id, email, and password are required"})
		return
	}

	// Validate password strength
	if len(req.Password) < 8 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters long"})
		return
	}

	// Register user
	user, err := api.authService.RegisterUser(req.TenantID, req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Return successful registration response
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "user registered successfully",
		"user": map[string]interface{}{
			"user_id":   user.UserID,
			"email":     user.Email,
			"tenant_id": user.TenantID,
			"status":    user.Status,
		},
	})
}

// GetCurrentUser returns information about the currently authenticated user
func (api *AuthAPI) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	// Get user from context (set by auth middleware)
	user := auth.UserFromContext(r.Context())
	if user.Subject == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user": user,
	})
}

// GetCurrentUserDev returns information about the currently authenticated user in development mode
func (api *AuthAPI) GetCurrentUserDev(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	// Get user from context (set by auth middleware)
	user := auth.UserFromContext(r.Context())
	if user.Subject == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		return
	}

	// For development, return enhanced user information
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user":        user,
		"message":     "User information retrieved successfully (development mode)",
		"environment": "development",
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	})
}

// RefreshToken refreshes an expired JWT token
func (api *AuthAPI) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.SessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "session_id is required"})
		return
	}

	// For now, return a simple response
	// In a real implementation, you would validate the session and generate a new token
	writeJSON(w, http.StatusOK, map[string]string{"message": "token refresh endpoint - implementation pending"})
}

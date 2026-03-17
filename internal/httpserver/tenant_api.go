package httpserver

import (
	"encoding/json"
	"net/http"
	"time"

	"jlrdi/internal/auth"

	"github.com/go-chi/chi/v5"
)

// TenantAPI handles tenant management operations
type TenantAPI struct {
	tenantManager *auth.TenantManager
}

// NewTenantAPI creates a new tenant API handler
func NewTenantAPI(tenantManager *auth.TenantManager) *TenantAPI {
	return &TenantAPI{
		tenantManager: tenantManager,
	}
}

// CreateTenantRequest represents a request to create a new tenant
type CreateTenantRequest struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
	Plan   string `json:"plan"`
}

// CreateTenantResponse represents the response when creating a tenant
type CreateTenantResponse struct {
	Tenant  *auth.Tenant `json:"tenant"`
	Message string       `json:"message"`
}

// CreateTenant creates a new tenant
func (t *TenantAPI) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var req CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate request
	if req.Name == "" || req.Domain == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Name and domain are required"})
		return
	}

	if req.Plan == "" {
		req.Plan = "basic" // Default plan
	}

	// Validate plan
	validPlans := map[string]bool{"basic": true, "professional": true, "enterprise": true}
	if !validPlans[req.Plan] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid plan. Must be basic, professional, or enterprise"})
		return
	}

	// Create tenant
	tenant, err := t.tenantManager.CreateTenant(req.Name, req.Domain, req.Plan)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response := CreateTenantResponse{
		Tenant:  tenant,
		Message: "Tenant created successfully",
	}

	writeJSON(w, http.StatusCreated, response)
}

// CreateTenantDev creates a new tenant in development mode (bypasses auth)
func (t *TenantAPI) CreateTenantDev(w http.ResponseWriter, r *http.Request) {
	var req CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate request
	if req.Name == "" || req.Domain == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Name and domain are required"})
		return
	}

	if req.Plan == "" {
		req.Plan = "basic" // Default plan
	}

	// Validate plan
	validPlans := map[string]bool{"basic": true, "professional": true, "enterprise": true}
	if !validPlans[req.Plan] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid plan. Must be basic, professional, or enterprise"})
		return
	}

	// Create tenant
	tenant, err := t.tenantManager.CreateTenant(req.Name, req.Domain, req.Plan)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response := CreateTenantResponse{
		Tenant:  tenant,
		Message: "Tenant created successfully (development mode)",
	}

	writeJSON(w, http.StatusCreated, response)
}

// GetTenant retrieves tenant information
func (t *TenantAPI) GetTenant(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")
	if tenantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Tenant ID is required"})
		return
	}

	// For now, we'll get tenant by domain (you can enhance this)
	// In production, you'd want to get by ID from the database
	tenant, err := t.tenantManager.GetTenantByDomain(tenantID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Tenant not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tenant":  tenant,
		"message": "Tenant retrieved successfully",
	})
}

// ListTenants lists all tenants (admin only)
func (t *TenantAPI) ListTenants(w http.ResponseWriter, r *http.Request) {
	// This would typically require admin permissions
	// For now, we'll return a simple response
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Tenant listing requires admin permissions",
		"tenants": []interface{}{},
	})
}

// CreateTenantUserRequest represents a request to create a user within a tenant
type CreateTenantUserRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}

// CreateTenantUserResponse represents the response when creating a tenant user
type CreateTenantUserResponse struct {
	User    *auth.TenantUser `json:"user"`
	Message string           `json:"message"`
}

// CreateTenantUser creates a new user within a tenant
func (t *TenantAPI) CreateTenantUser(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")
	if tenantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Tenant ID is required"})
		return
	}

	var req CreateTenantUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate request
	if req.Email == "" || req.FirstName == "" || req.LastName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Email, first name, and last name are required"})
		return
	}

	if req.Role == "" {
		req.Role = "viewer" // Default role
	}

	// Validate role
	validRoles := map[string]bool{"owner": true, "admin": true, "manager": true, "analyst": true, "viewer": true}
	if !validRoles[req.Role] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid role"})
		return
	}

	// Create user
	user, err := t.tenantManager.CreateTenantUser(tenantID, req.Email, req.FirstName, req.LastName, req.Role)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response := CreateTenantUserResponse{
		User:    user,
		Message: "User created successfully",
	}

	writeJSON(w, http.StatusCreated, response)
}

// ListTenantUsersResponse represents the response when listing tenant users
type ListTenantUsersResponse struct {
	Users   []*auth.TenantUser `json:"users"`
	Total   int                `json:"total"`
	Message string             `json:"message"`
}

// ListTenantUsers lists all users within a tenant
func (t *TenantAPI) ListTenantUsers(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")
	if tenantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Tenant ID is required"})
		return
	}

	users, err := t.tenantManager.ListTenantUsers(tenantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response := ListTenantUsersResponse{
		Users:   users,
		Total:   len(users),
		Message: "Users retrieved successfully",
	}

	writeJSON(w, http.StatusOK, response)
}

// ListMyTenantUsers returns minimal user info for the tenant from JWT
func (t *TenantAPI) ListMyTenantUsers(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user.Subject == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		return
	}
	if user.TenantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "tenant_id missing in token"})
		return
	}

	users, err := t.tenantManager.ListTenantUsers(user.TenantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	type briefUser struct {
		Name      string    `json:"name"`
		Email     string    `json:"email"`
		Status    string    `json:"status"`
		Role      string    `json:"role"`
		CreatedAt time.Time `json:"created_at"`
	}
	brief := make([]briefUser, 0, len(users))
	for _, u := range users {
		name := u.FirstName
		if u.LastName != "" {
			if name != "" {
				name = name + " " + u.LastName
			} else {
				name = u.LastName
			}
		}
		brief = append(brief, briefUser{Name: name, Email: u.Email, Status: u.Status, Role: u.Role, CreatedAt: u.CreatedAt})
	}

	// Fallback: if no tenant users are stored yet, include the authenticated user
	if len(brief) == 0 {
		role := "user"
		if len(user.Roles) > 0 {
			role = user.Roles[0]
		}
		brief = append(brief, briefUser{
			Name:      user.Email,
			Email:     user.Email,
			Status:    "active",
			Role:      role,
			CreatedAt: time.Now().UTC(),
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"users": brief, "total": len(brief)})
}

// CreateTenantUserDev creates a new user within a tenant in development mode
func (t *TenantAPI) CreateTenantUserDev(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")
	if tenantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Tenant ID is required"})
		return
	}

	var req CreateTenantUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate request
	if req.Email == "" || req.FirstName == "" || req.LastName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Email, first name, and last name are required"})
		return
	}

	if req.Role == "" {
		req.Role = "viewer" // Default role
	}

	// Validate role
	validRoles := map[string]bool{"owner": true, "admin": true, "manager": true, "analyst": true, "viewer": true}
	if !validRoles[req.Role] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid role. Must be owner, admin, manager, analyst, or viewer"})
		return
	}

	// Create tenant user
	user, err := t.tenantManager.CreateTenantUser(tenantID, req.Email, req.FirstName, req.LastName, req.Role)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response := CreateTenantUserResponse{
		User:    user,
		Message: "Tenant user created successfully (development mode)",
	}

	writeJSON(w, http.StatusCreated, response)
}

// ListTenantUsersDev lists all users within a tenant in development mode
func (t *TenantAPI) ListTenantUsersDev(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")
	if tenantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Tenant ID is required"})
		return
	}

	// Get tenant users
	users, err := t.tenantManager.ListTenantUsers(tenantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response := ListTenantUsersResponse{
		Users:   users,
		Total:   len(users),
		Message: "Tenant users retrieved successfully (development mode)",
	}

	writeJSON(w, http.StatusOK, response)
}

// GetTenantStats returns statistics for a tenant
func (t *TenantAPI) GetTenantStats(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")
	if tenantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Tenant ID is required"})
		return
	}

	stats, err := t.tenantManager.GetTenantStats(tenantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"stats":   stats,
		"message": "Tenant statistics retrieved successfully",
	})
}

// UpdateTenantUserRequest represents a request to update a user
type UpdateTenantUserRequest struct {
	Role      string `json:"role,omitempty"`
	Status    string `json:"status,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// UpdateTenantUser updates a user's information
func (t *TenantAPI) UpdateTenantUser(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")
	userID := chi.URLParam(r, "uid")

	if tenantID == "" || userID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Tenant ID and User ID are required"})
		return
	}

	var req UpdateTenantUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.FirstName != "" {
		updates["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		updates["last_name"] = req.LastName
	}

	if len(updates) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "No updates provided"})
		return
	}

	// Update user
	err := t.tenantManager.UpdateTenantUser(userID, updates)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "User updated successfully",
	})
}

// DeleteTenantUser deletes a user from a tenant
func (t *TenantAPI) DeleteTenantUser(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")
	userID := chi.URLParam(r, "uid")

	if tenantID == "" || userID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Tenant ID and User ID are required"})
		return
	}

	// Delete user
	err := t.tenantManager.DeleteTenantUser(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "User deleted successfully",
	})
}

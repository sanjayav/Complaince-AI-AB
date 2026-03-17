package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// Tenant represents a multi-tenant organization
type Tenant struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Domain    string            `json:"domain"`
	Status    string            `json:"status"` // active, suspended, deleted
	Plan      string            `json:"plan"`   // basic, professional, enterprise
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Limits    TenantLimits      `json:"limits"`
	Settings  TenantSettings    `json:"settings"`
}

// TenantLimits defines usage limits for a tenant
type TenantLimits struct {
	MaxDocuments   int `json:"max_documents"`
	MaxStorageGB   int `json:"max_storage_gb"`
	MaxUsers       int `json:"max_users"`
	MaxAPIRequests int `json:"max_api_requests_per_minute"`
	MaxConcurrent  int `json:"max_concurrent_uploads"`
}

// TenantSettings defines tenant-specific configurations
type TenantSettings struct {
	AllowedFileTypes []string          `json:"allowed_file_types"`
	MaxFileSizeMB    int               `json:"max_file_size_mb"`
	RetentionDays    int               `json:"retention_days"`
	CustomFields     map[string]string `json:"custom_fields,omitempty"`
}

// TenantUser represents a user within a tenant
type TenantUser struct {
	ID          string            `json:"id"`
	TenantID    string            `json:"tenant_id"`
	Email       string            `json:"email"`
	FirstName   string            `json:"first_name"`
	LastName    string            `json:"last_name"`
	Role        string            `json:"role"`   // owner, admin, manager, analyst, viewer
	Status      string            `json:"status"` // active, inactive, suspended
	Permissions []string          `json:"permissions"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	LastLoginAt *time.Time        `json:"last_login_at,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// TenantManagerInterface defines the contract for tenant operations
type TenantManagerInterface interface {
	GetTenantByEmail(email string) (*Tenant, error)
	GetTenantByDomain(domain string) (*Tenant, error)
	GetTenantByID(id string) (*Tenant, error)
	CreateTenant(name, domain, plan string) (*Tenant, error)
	ListTenants() []*Tenant
}

// TenantManager manages multi-tenant operations
type TenantManager struct {
	tenants map[string]*Tenant
	users   map[string]*TenantUser
	mu      sync.RWMutex
}

// NewTenantManager creates a new tenant manager
func NewTenantManager() *TenantManager {
	return &TenantManager{
		tenants: make(map[string]*Tenant),
		users:   make(map[string]*TenantUser),
	}
}

// CreateTenant creates a new tenant
func (tm *TenantManager) CreateTenant(name, domain, plan string) (*Tenant, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Check if domain already exists
	for _, tenant := range tm.tenants {
		if tenant.Domain == domain {
			return nil, fmt.Errorf("domain %s already exists", domain)
		}
	}

	tenantID := generateTenantID()
	tenant := &Tenant{
		ID:        tenantID,
		Name:      name,
		Domain:    domain,
		Status:    "active",
		Plan:      plan,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  make(map[string]string),
		Limits:    getDefaultLimits(plan),
		Settings:  getDefaultSettings(plan),
	}

	tm.tenants[tenantID] = tenant
	log.Printf("Created tenant: %s (%s) with plan: %s", name, domain, plan)

	return tenant, nil
}

// CreateTenantUser creates a new user within a tenant
func (tm *TenantManager) CreateTenantUser(tenantID, email, firstName, lastName, role string) (*TenantUser, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Check if tenant exists
	tenant, exists := tm.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("tenant %s not found", tenantID)
	}

	// Check if user already exists in this tenant
	for _, user := range tm.users {
		if user.TenantID == tenantID && user.Email == email {
			return nil, fmt.Errorf("user %s already exists in tenant %s", email, tenantID)
		}
	}

	userID := generateUserID()
	user := &TenantUser{
		ID:          userID,
		TenantID:    tenantID,
		Email:       email,
		FirstName:   firstName,
		LastName:    lastName,
		Role:        role,
		Status:      "active",
		Permissions: getRolePermissions(role),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    make(map[string]string),
	}

	tm.users[userID] = user
	log.Printf("Created user %s in tenant %s with role %s", email, tenant.Name, role)

	return user, nil
}

// GetTenantByDomain finds a tenant by domain
func (tm *TenantManager) GetTenantByDomain(domain string) (*Tenant, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	for _, tenant := range tm.tenants {
		if tenant.Domain == domain {
			return tenant, nil
		}
	}

	return nil, fmt.Errorf("tenant with domain %s not found", domain)
}

// GetTenantByEmailDomain automatically detects tenant from email domain
func (tm *TenantManager) GetTenantByEmailDomain(email string) (*Tenant, error) {
	// Extract domain from email (e.g., "user@jlr.com" -> "jlr.com")
	atIndex := strings.Index(email, "@")
	if atIndex == -1 {
		return nil, fmt.Errorf("invalid email format")
	}

	domain := email[atIndex+1:]
	return tm.GetTenantByDomain(domain)
}

// GetTenantByEmail automatically detects tenant from email
func (tm *TenantManager) GetTenantByEmail(email string) (*Tenant, error) {
	return tm.GetTenantByEmailDomain(email)
}

// GetTenantByID retrieves a tenant by ID
func (tm *TenantManager) GetTenantByID(id string) (*Tenant, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tenant, exists := tm.tenants[id]
	if !exists {
		return nil, fmt.Errorf("tenant with ID %s not found", id)
	}

	return tenant, nil
}

// GetTenantUser retrieves a user by ID
func (tm *TenantManager) GetTenantUser(userID string) (*TenantUser, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	user, exists := tm.users[userID]
	if !exists {
		return nil, fmt.Errorf("user %s not found", userID)
	}

	return user, nil
}

// ValidateTenantAccess validates if a user has access to a tenant
func (tm *TenantManager) ValidateTenantAccess(userID, tenantID string) error {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	user, exists := tm.users[userID]
	if !exists {
		return fmt.Errorf("user %s not found", userID)
	}

	if user.TenantID != tenantID {
		return fmt.Errorf("user %s does not have access to tenant %s", userID, tenantID)
	}

	if user.Status != "active" {
		return fmt.Errorf("user %s is not active", userID)
	}

	return nil
}

// CheckPermission checks if a user has a specific permission
func (tm *TenantManager) CheckPermission(userID, permission string) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	user, exists := tm.users[userID]
	if !exists || user.Status != "active" {
		return false
	}

	for _, userPerm := range user.Permissions {
		if userPerm == permission || userPerm == "*" {
			return true
		}
	}

	return false
}

// ListTenantUsers lists all users in a tenant
func (tm *TenantManager) ListTenantUsers(tenantID string) ([]*TenantUser, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var users []*TenantUser
	for _, user := range tm.users {
		if user.TenantID == tenantID {
			users = append(users, user)
		}
	}

	return users, nil
}

// UpdateTenantUser updates a user's information
func (tm *TenantManager) UpdateTenantUser(userID string, updates map[string]interface{}) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	user, exists := tm.users[userID]
	if !exists {
		return fmt.Errorf("user %s not found", userID)
	}

	// Update allowed fields
	if role, ok := updates["role"].(string); ok {
		user.Role = role
		user.Permissions = getRolePermissions(role)
	}

	if status, ok := updates["status"].(string); ok {
		user.Status = status
	}

	if firstName, ok := updates["first_name"].(string); ok {
		user.FirstName = firstName
	}

	if lastName, ok := updates["last_name"].(string); ok {
		user.LastName = lastName
	}

	user.UpdatedAt = time.Now()
	return nil
}

// DeleteTenantUser deletes a user from a tenant
func (tm *TenantManager) DeleteTenantUser(userID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.users[userID]; !exists {
		return fmt.Errorf("user %s not found", userID)
	}

	delete(tm.users, userID)
	log.Printf("Deleted user: %s", userID)
	return nil
}

// GetTenantStats returns statistics for a tenant
func (tm *TenantManager) GetTenantStats(tenantID string) (map[string]interface{}, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tenant, exists := tm.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("tenant %s not found", tenantID)
	}

	// Count active users
	activeUsers := 0
	for _, user := range tm.users {
		if user.TenantID == tenantID && user.Status == "active" {
			activeUsers++
		}
	}

	stats := map[string]interface{}{
		"tenant_id":    tenantID,
		"tenant_name":  tenant.Name,
		"plan":         tenant.Plan,
		"status":       tenant.Status,
		"active_users": activeUsers,
		"limits":       tenant.Limits,
		"created_at":   tenant.CreatedAt,
	}

	return stats, nil
}

// ListTenants returns all tenants
func (tm *TenantManager) ListTenants() []*Tenant {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tenants := make([]*Tenant, 0, len(tm.tenants))
	for _, tenant := range tm.tenants {
		tenants = append(tenants, tenant)
	}

	return tenants
}

// Helper functions
func generateTenantID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return "tenant_" + hex.EncodeToString(bytes)
}

func generateUserID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return "user_" + hex.EncodeToString(bytes)
}

func getDefaultLimits(plan string) TenantLimits {
	switch plan {
	case "enterprise":
		return TenantLimits{
			MaxDocuments:   100000,
			MaxStorageGB:   1000,
			MaxUsers:       1000,
			MaxAPIRequests: 1000,
			MaxConcurrent:  50,
		}
	case "professional":
		return TenantLimits{
			MaxDocuments:   10000,
			MaxStorageGB:   100,
			MaxUsers:       100,
			MaxAPIRequests: 500,
			MaxConcurrent:  20,
		}
	default: // basic
		return TenantLimits{
			MaxDocuments:   1000,
			MaxStorageGB:   10,
			MaxUsers:       10,
			MaxAPIRequests: 100,
			MaxConcurrent:  5,
		}
	}
}

func getDefaultSettings(plan string) TenantSettings {
	switch plan {
	case "enterprise":
		return TenantSettings{
			AllowedFileTypes: []string{".pdf", ".doc", ".docx", ".txt", ".csv", ".xlsx"},
			MaxFileSizeMB:    100,
			RetentionDays:    3650, // 10 years
		}
	case "professional":
		return TenantSettings{
			AllowedFileTypes: []string{".pdf", ".doc", ".docx", ".txt", ".csv"},
			MaxFileSizeMB:    50,
			RetentionDays:    1095, // 3 years
		}
	default: // basic
		return TenantSettings{
			AllowedFileTypes: []string{".pdf", ".txt", ".csv"},
			MaxFileSizeMB:    25,
			RetentionDays:    365, // 1 year
		}
	}
}

func getRolePermissions(role string) []string {
	switch role {
	case "owner":
		return []string{"*"} // All permissions
	case "admin":
		return []string{
			"tenant:manage", "users:manage", "documents:manage",
			"search:all", "analytics:all", "settings:manage",
		}
	case "manager":
		return []string{
			"users:view", "documents:manage", "search:all",
			"analytics:view", "reports:manage",
		}
	case "analyst":
		return []string{
			"documents:read", "documents:upload", "search:all",
			"analytics:view", "reports:view",
		}
	case "viewer":
		return []string{
			"documents:read", "search:basic", "reports:view",
		}
	default:
		return []string{"documents:read"}
	}
}

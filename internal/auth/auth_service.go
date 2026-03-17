package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthService handles user authentication and session management
type AuthService struct {
	users         map[string]*UserCredentials
	sessions      map[string]*Session
	tenantManager TenantManagerInterface
	mu            sync.RWMutex
	jwtSecret     string
}

// UserCredentials stores user authentication information
type UserCredentials struct {
	UserID       string     `json:"user_id"`
	TenantID     string     `json:"tenant_id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Salt         string     `json:"-"`
	Status       string     `json:"status"` // active, inactive, suspended
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
}

// Session represents an active user session
type Session struct {
	SessionID  string    `json:"session_id"`
	UserID     string    `json:"user_id"`
	TenantID   string    `json:"tenant_id"`
	Token      string    `json:"token"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
	IPAddress  string    `json:"ip_address,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
}

// LoginRequest represents a login attempt
type LoginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	TenantID   string `json:"tenant_id"`
	RememberMe bool   `json:"remember_me"`
}

// LoginResponse represents a successful login response
type LoginResponse struct {
	Token       string    `json:"token"`
	User        *User     `json:"user"`
	Tenant      *Tenant   `json:"tenant"`
	ExpiresAt   time.Time `json:"expires_at"`
	SessionID   string    `json:"session_id"`
	Permissions []string  `json:"permissions"`
}

// NewAuthService creates a new authentication service
func NewAuthService(jwtSecret string, tenantManager TenantManagerInterface) *AuthService {
	return &AuthService{
		users:         make(map[string]*UserCredentials),
		sessions:      make(map[string]*Session),
		tenantManager: tenantManager,
		jwtSecret:     jwtSecret,
	}
}

// RegisterUser registers a new user with password authentication
func (as *AuthService) RegisterUser(tenantID, email, password, firstName, lastName string) (*UserCredentials, error) {
	as.mu.Lock()
	defer as.mu.Unlock()

	// Check if user already exists
	for _, user := range as.users {
		if user.Email == email && user.TenantID == tenantID {
			return nil, fmt.Errorf("user with email %s already exists in tenant %s", email, tenantID)
		}
	}

	// Generate salt and hash password
	salt := generateSalt()
	passwordHash := hashPassword(password, salt)

	userID := generateUserID()
	user := &UserCredentials{
		UserID:       userID,
		TenantID:     tenantID,
		Email:        email,
		PasswordHash: passwordHash,
		Salt:         salt,
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	as.users[userID] = user
	return user, nil
}

// AuthenticateUser validates user credentials and creates a session
func (as *AuthService) AuthenticateUser(email, password, tenantID string) (*LoginResponse, error) {
	// If tenantID is not provided, try to auto-detect it from email
	if tenantID == "" {
		// This would require access to tenant manager
		// For now, we'll require tenantID but provide a helper method
		return nil, fmt.Errorf("tenant_id is required. Use AuthenticateUserAutoTenant for automatic tenant detection")
	}

	// Find user by email and tenant (read lock)
	var user *UserCredentials
	as.mu.RLock()
	for _, u := range as.users {
		if u.Email == email && u.TenantID == tenantID {
			user = u
			break
		}
	}
	as.mu.RUnlock()

	if user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if user.Status != "active" {
		return nil, fmt.Errorf("user account is %s", user.Status)
	}

	// Validate password
	if !validatePassword(password, user.PasswordHash, user.Salt) {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Update last login (write lock)
	now := time.Now()
	as.mu.Lock()
	user.LastLoginAt = &now
	as.mu.Unlock()

	// Create JWT token
	token, expiresAt, err := as.createJWTToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	// Create session (write lock)
	sessionID := generateSessionID()
	session := &Session{
		SessionID:  sessionID,
		UserID:     user.UserID,
		TenantID:   user.TenantID,
		Token:      token,
		ExpiresAt:  expiresAt,
		CreatedAt:  now,
		LastUsedAt: now,
	}

	as.mu.Lock()
	as.sessions[sessionID] = session
	as.mu.Unlock()

	// Get user permissions (this would typically come from tenant manager)
	permissions := getDefaultRolePermissions("analyst") // Default role

	// Create response
	response := &LoginResponse{
		Token:       token,
		User:        as.convertToUser(user),
		ExpiresAt:   expiresAt,
		SessionID:   sessionID,
		Permissions: permissions,
	}

	return response, nil
}

// AuthenticateUserAutoTenant automatically detects tenant from email domain
func (as *AuthService) AuthenticateUserAutoTenant(email, password string) (*LoginResponse, error) {
	// Auto-detect tenant from email domain
	tenant, err := as.tenantManager.GetTenantByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("could not determine tenant from email %s: %w", email, err)
	}

	// Use the detected tenant ID
	return as.AuthenticateUser(email, password, tenant.ID)
}

// ValidateToken validates a JWT token and returns user information
func (as *AuthService) ValidateToken(tokenString string) (*User, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(as.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Find user by subject
	as.mu.RLock()
	defer as.mu.RUnlock()

	user, exists := as.users[claims.Subject]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	if user.Status != "active" {
		return nil, fmt.Errorf("user account is %s", user.Status)
	}

	return as.convertToUser(user), nil
}

// LogoutUser invalidates a user session
func (as *AuthService) LogoutUser(sessionID string) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	if _, exists := as.sessions[sessionID]; !exists {
		return fmt.Errorf("session not found")
	}

	delete(as.sessions, sessionID)
	return nil
}

// GetUserByEmail finds a user by email and tenant
func (as *AuthService) GetUserByEmail(email, tenantID string) (*UserCredentials, error) {
	as.mu.RLock()
	defer as.mu.RUnlock()

	for _, user := range as.users {
		if user.Email == email && user.TenantID == tenantID {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

// createJWTToken creates a JWT token for a user
func (as *AuthService) createJWTToken(user *UserCredentials) (string, time.Time, error) {
	expiresAt := time.Now().Add(24 * time.Hour) // 24 hour expiration

	claims := &Claims{
		Email:    user.Email,
		TenantID: user.TenantID,
		Roles:    []string{"user"}, // Default role
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.UserID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(as.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// convertToUser converts UserCredentials to User for API responses
func (as *AuthService) convertToUser(creds *UserCredentials) *User {
	return &User{
		Subject: creds.UserID,
		Email:   creds.Email,
		Roles:   []string{"user"}, // Default role
	}
}

// Helper functions
func generateSalt() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func hashPassword(password, salt string) string {
	hash := sha256.Sum256([]byte(password + salt))
	return hex.EncodeToString(hash[:])
}

func validatePassword(password, hash, salt string) bool {
	expectedHash := hashPassword(password, salt)
	return hash == expectedHash
}

func generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return "session_" + hex.EncodeToString(bytes)
}

func getDefaultRolePermissions(role string) []string {
	switch role {
	case "owner":
		return []string{"*"}
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

package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"
)

// APIKey represents an API key for frontend access
type APIKey struct {
	ID          string            `json:"id"`
	Key         string            `json:"key"`         // Full key (only shown once)
	KeyHash     string            `json:"-"`           // Hashed key for storage
	Name        string            `json:"name"`        // Developer name/description
	Permissions []string          `json:"permissions"` // Allowed operations
	CreatedAt   time.Time         `json:"created_at"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
	LastUsed    *time.Time        `json:"last_used,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// APIKeyManager manages API keys
type APIKeyManager struct {
	keys map[string]*APIKey // key_hash -> APIKey
	mu   sync.RWMutex
}

// NewAPIKeyManager creates a new API key manager
func NewAPIKeyManager() *APIKeyManager {
	return &APIKeyManager{
		keys: make(map[string]*APIKey),
	}
}

// GenerateAPIKey generates a new API key for a frontend developer
func (m *APIKeyManager) GenerateAPIKey(name string, permissions []string, expiresAt *time.Time) (*APIKey, error) {
	// Generate random key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	key := hex.EncodeToString(keyBytes)
	keyHash := hashKey(key)

	apiKey := &APIKey{
		ID:          generateID(),
		Key:         key,
		KeyHash:     keyHash,
		Name:        name,
		Permissions: permissions,
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
		Metadata: map[string]string{
			"created_by": "system",
			"purpose":    "frontend_development",
		},
	}

	m.mu.Lock()
	m.keys[keyHash] = apiKey
	m.mu.Unlock()

	log.Printf("Generated API key for %s: %s", name, key)
	return apiKey, nil
}

// ValidateAPIKey validates an API key and returns the associated permissions
func (m *APIKeyManager) ValidateAPIKey(key string) (*APIKey, error) {
	keyHash := hashKey(key)

	m.mu.RLock()
	apiKey, exists := m.keys[keyHash]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("invalid API key")
	}

	// Check if key is expired
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, fmt.Errorf("API key has expired")
	}

	// Update last used timestamp
	now := time.Now()
	apiKey.LastUsed = &now

	return apiKey, nil
}

// ListAPIKeys returns all API keys (without the actual key values)
func (m *APIKeyManager) ListAPIKeys() []*APIKey {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var keys []*APIKey
	for _, key := range m.keys {
		// Don't expose the actual key
		keyCopy := *key
		keyCopy.Key = "***hidden***"
		keys = append(keys, &keyCopy)
	}

	return keys
}

// RevokeAPIKey revokes an API key
func (m *APIKeyManager) RevokeAPIKey(keyHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.keys[keyHash]; !exists {
		return fmt.Errorf("API key not found")
	}

	delete(m.keys, keyHash)
	log.Printf("Revoked API key: %s", keyHash)
	return nil
}

// hashKey creates a SHA256 hash of the API key
func hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// generateID generates a unique ID for the API key
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Default permissions for frontend developers
var FrontendPermissions = []string{
	"documents:read",
	"documents:upload",
	"search:semantic",
	"rag:ask",
	"answers:read",
	"health:read",
}

// AdminPermissions includes all permissions
var AdminPermissions = []string{
	"documents:read",
	"documents:upload",
	"documents:delete",
	"search:semantic",
	"rag:ask",
	"answers:read",
	"answers:write",
	"health:read",
	"metrics:read",
	"admin:manage",
}

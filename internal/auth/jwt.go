package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userKey contextKey = "user"

type User struct {
	Subject  string   `json:"sub"`
	Email    string   `json:"email"`
	TenantID string   `json:"tenant_id"`
	Roles    []string `json:"roles"`
}

type Claims struct {
	Email    string   `json:"email"`
	TenantID string   `json:"tenant_id"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

func RequireAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// First check if a user is already in the context (from DevUser injection)
			if user := UserFromContext(r.Context()); user.Subject != "" {
				next.ServeHTTP(w, r)
				return
			}

			// Otherwise, require a valid bearer token
			authz := r.Header.Get("Authorization")
			if !strings.HasPrefix(strings.ToLower(authz), "bearer ") {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}
			tokenString := strings.TrimSpace(authz[len("Bearer "):])
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			user := User{Subject: claims.Subject, Email: claims.Email, TenantID: claims.TenantID, Roles: claims.Roles}
			// Debug logging
			if claims.TenantID == "" {
				// Try to extract from email domain as fallback
				if strings.Contains(claims.Email, "@") {
					parts := strings.Split(claims.Email, "@")
					if len(parts) == 2 {
						user.TenantID = parts[1] // Use domain as fallback tenant ID
					}
				}
			}
			ctx := context.WithValue(r.Context(), userKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	set := map[string]struct{}{}
	for _, r := range roles {
		set[r] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := UserFromContext(r.Context())
			for _, role := range user.Roles {
				if _, ok := set[role]; ok {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, "forbidden", http.StatusForbidden)
		})
	}
}

func UserFromContext(ctx context.Context) User {
	if u, ok := ctx.Value(userKey).(User); ok {
		return u
	}
	return User{}
}

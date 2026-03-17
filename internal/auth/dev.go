package auth

import (
	"context"
	"log"
	"net/http"
)

// DevUser injects a default user in dev environments when no auth is present.
func DevUser() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("DevUser middleware executing for path: %s", r.URL.Path)

			if _, ok := r.Context().Value(userKey).(User); !ok {
				log.Printf("No user in context, injecting dev user")
				u := User{Subject: "dev-user", Email: "dev@example.com", Roles: []string{"admin", "qa_lead", "analyst", "viewer"}}
				ctx := context.WithValue(r.Context(), userKey, u)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			log.Printf("User already in context, proceeding")
			next.ServeHTTP(w, r)
		})
	}
}

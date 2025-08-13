package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/AnjuRKrishnan/fleet-tracker/internal/auth"
)

type contextKey string

const UserIDContextKey = contextKey("userID")

// JWTAuthenticator is a middleware to validate JWT tokens.
func JWTAuthenticator(jwtAuth *auth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				http.Error(w, "Invalid token format", http.StatusUnauthorized)
				return
			}

			claims, err := jwtAuth.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Store user claims in context
			ctx := context.WithValue(r.Context(), UserIDContextKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

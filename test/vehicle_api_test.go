package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AnjuRKrishnan/fleet-tracker/internal/middleware"
	"github.com/golang-jwt/jwt/v5"
)

func TestJWTMiddleware(t *testing.T) {
	secret := "test-secret"
	m := middleware.NewJWTMiddleware(secret)

	// helper to create token
	makeToken := func(valid bool) string {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "user1"})
		if valid {
			tok, _ := token.SignedString([]byte(secret))
			return tok
		}
		tok, _ := token.SignedString([]byte("wrong"))
		return tok
	}

	tests := []struct {
		name     string
		token    string
		wantCode int
	}{
		{"no token", "", http.StatusUnauthorized},
		{"valid token", makeToken(true), http.StatusOK},
		{"invalid token", makeToken(false), http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest("GET", "/api/vehicle/status", nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tc.wantCode {
				t.Fatalf("expected %d got %d", tc.wantCode, rec.Code)
			}
		})
	}
}

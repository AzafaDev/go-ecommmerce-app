package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-ecommerce-app/internal/middleware"
	"go-ecommerce-app/pkg/security"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {
	const secret = "test-jwt-secret"

	validToken, err := security.GenerateToken(secret, time.Minute, uuid.New(), "customer")
	require.NoError(t, err)

	expiredToken, err := security.GenerateToken(secret, -time.Minute, uuid.New(), "customer")
	require.NoError(t, err)

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{name: "valid token is accepted", authHeader: "Bearer " + validToken, wantStatus: http.StatusOK},
		{name: "expired token is rejected", authHeader: "Bearer " + expiredToken, wantStatus: http.StatusUnauthorized},
		{name: "malformed token is rejected", authHeader: "Bearer not-a-real-token", wantStatus: http.StatusUnauthorized},
		{name: "missing header is rejected", authHeader: "", wantStatus: http.StatusUnauthorized},
		{name: "missing Bearer prefix is rejected", authHeader: validToken, wantStatus: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			handler := middleware.AuthMiddleware(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantStatus == http.StatusOK, called)
		})
	}
}

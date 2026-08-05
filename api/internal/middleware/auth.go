package middleware

import (
	"context"
	"fmt"
	"go-ecommerce-app/pkg/response"
	"go-ecommerce-app/pkg/security"
	"net/http"
	"strings"
)

type contextKey string

const claimsContextKey contextKey = "claims"

func AuthMiddleware(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.WriteErrorJSON("unauthorized", http.StatusUnauthorized, w)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.WriteErrorJSON("unauthorized", http.StatusUnauthorized, w)
				return
			}

			claims, err := security.ParseToken(secretKey, parts[1])
			if err != nil {
				response.WriteErrorJSON("invalid or expired token", http.StatusUnauthorized, w)
				return
			}

			ctx := context.WithValue(r.Context(), claimsContextKey, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := GetClaims(r.Context())
			if err != nil {
				response.WriteErrorJSON("forbidden", http.StatusForbidden, w)
				return
			}

			if claims.Role != role {
				response.WriteErrorJSON("forbidden", http.StatusForbidden, w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetClaims(ctx context.Context) (*security.Claims, error) {
	claims, ok := ctx.Value(claimsContextKey).(*security.Claims)
	if !ok {
		return nil, fmt.Errorf("error in getting claims from context value")
	}
	return claims, nil
}

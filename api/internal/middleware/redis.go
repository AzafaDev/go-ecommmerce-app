package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-ecommerce-app/pkg/response"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func RateLimit(rdb *redis.Client, limit int, window time.Duration, keyFunc func(*http.Request) (string, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identifier, err := keyFunc(r)
			if err != nil || identifier == "" {
				slog.Error("rate limit", "error", err)
				response.WriteErrorJSON("unable to identify client", http.StatusBadRequest, w)
				return
			}

			redisKey := fmt.Sprintf("rate_limit:%s", identifier)

			pipe := rdb.Pipeline()
			incrCmd := pipe.Incr(r.Context(), redisKey)
			ttlCmd := pipe.TTL(r.Context(), redisKey)

			_, err = pipe.Exec(r.Context())
			if err != nil && err != redis.Nil {
				slog.Error("rate limit", "error", err)
				next.ServeHTTP(w, r)
				return
			}

			reqCount := incrCmd.Val()

			if ttlCmd.Val() < 0 {
				rdb.Expire(r.Context(), redisKey, window)
			}

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			remaining := limit - int(reqCount)
			if remaining < 0 {
				remaining = 0
			}
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

			if int(reqCount) > limit {
				w.Header().Set("Retry-After", strconv.Itoa(int(window.Seconds())))
				slog.Warn("rate limit exceeded", "identifier", identifier)
				response.WriteErrorJSON("too many request", http.StatusTooManyRequests, w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func IPKeyFunc(endpoint string) func(*http.Request) (string, error) {
	return func(r *http.Request) (string, error) {
		return fmt.Sprintf("%s:%s", endpoint, clientIP(r)), nil
	}
}

func EmailKeyFunc(endpoint string) func(*http.Request) (string, error) {
	return func(r *http.Request) (string, error) {
		email, err := peekEmail(r)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s:%s", endpoint, email), nil
	}
}

func UserIDKeyFunc(endpoint string) func(*http.Request) (string, error) {
	return func(r *http.Request) (string, error) {
		claims, err := GetClaims(r.Context())
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s:%s", endpoint, claims.ID), nil
	}
}

func IPAndEmailKeyFunc(endpoint string) func(*http.Request) (string, error) {
	return func(r *http.Request) (string, error) {
		email, err := peekEmail(r)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s:%s:%s", endpoint, clientIP(r), email), nil
	}
}

// peekEmail baca "email" dari body JSON tanpa menghabiskan body, jadi
// handler di belakang middleware ini masih bisa decode body-nya sendiri.
func peekEmail(r *http.Request) (string, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(body))

	var payload struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}

	email := strings.ToLower(strings.TrimSpace(payload.Email))
	if email == "" {
		return "", fmt.Errorf("email is required")
	}

	return email, nil
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

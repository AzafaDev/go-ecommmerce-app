package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-ecommerce-app/internal/middleware"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { client.Close() })

	return client
}

func newPassthroughHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestRateLimit(t *testing.T) {
	t.Run("request under the limit passes through", func(t *testing.T) {
		rdb := newTestRedis(t)
		handler := middleware.RateLimit(rdb, 3, time.Minute, middleware.IPKeyFunc("test"))(newPassthroughHandler())

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("request past the limit is rejected with 429 and Retry-After", func(t *testing.T) {
		rdb := newTestRedis(t)
		handler := middleware.RateLimit(rdb, 2, time.Minute, middleware.IPKeyFunc("test"))(newPassthroughHandler())

		for i := 0; i < 2; i++ {
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			require.Equal(t, http.StatusOK, rec.Code)
		}

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusTooManyRequests, rec.Code)
		assert.NotEmpty(t, rec.Header().Get("Retry-After"))
	})

	t.Run("X-RateLimit-Remaining decreases on each successive request", func(t *testing.T) {
		rdb := newTestRedis(t)
		handler := middleware.RateLimit(rdb, 5, time.Minute, middleware.IPKeyFunc("test"))(newPassthroughHandler())

		for _, want := range []string{"4", "3", "2"} {
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			assert.Equal(t, want, rec.Header().Get("X-RateLimit-Remaining"))
		}
	})

	t.Run("TTL is set once at the start of the window, not reset on every request", func(t *testing.T) {
		mr, err := miniredis.Run()
		require.NoError(t, err)
		t.Cleanup(mr.Close)

		rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		t.Cleanup(func() { rdb.Close() })

		const window = time.Minute
		handler := middleware.RateLimit(rdb, 5, window, middleware.IPKeyFunc("test"))(newPassthroughHandler())
		key := "rate_limit:test:192.0.2.1"

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)
		firstTTL := mr.TTL(key)
		require.Greater(t, firstTTL, time.Duration(0))

		// Move the fake clock forward without a real sleep. If the code
		// wrongly re-armed the TTL on every request instead of only when
		// the key has no expiry yet, the second read would jump back up
		// to ~window instead of continuing to count down.
		mr.FastForward(10 * time.Second)

		req = httptest.NewRequest(http.MethodPost, "/", nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)
		secondTTL := mr.TTL(key)

		assert.Less(t, secondTTL, firstTTL)
		assert.InDelta(t, (firstTTL - 10*time.Second).Seconds(), secondTTL.Seconds(), 1)
	})
}

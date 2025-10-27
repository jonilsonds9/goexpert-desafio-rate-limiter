package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jonilsonds9/goexpert-desafio-rate-limiter/configs"
	"github.com/jonilsonds9/goexpert-desafio-rate-limiter/internal/limiter"
	"github.com/jonilsonds9/goexpert-desafio-rate-limiter/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiterMiddleware_IPBased(t *testing.T) {
	// Setup
	cfg := &configs.Config{
		RateLimitIP:          5,
		RateLimitIPBlockTime: 1,
	}
	store := storage.NewMemoryStorage()
	rateLimiter := limiter.NewRateLimiter(store)
	middleware := RateLimiterMiddleware(cfg, rateLimiter)

	// Create a simple handler
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))

	// Test requests within limit
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}

	// Test request over limit
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code, "Request over limit should be blocked")

	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "you have reached the maximum number of requests or actions allowed within a certain time frame", response["error"])
}

func TestRateLimiterMiddleware_TokenBased(t *testing.T) {
	// Setup
	cfg := &configs.Config{
		RateLimitIP:             10,
		RateLimitIPBlockTime:    1,
		RateLimitToken:          3,
		RateLimitTokenBlockTime: 1,
		TokenConfigs:            make(map[string]configs.TokenConfig),
	}
	store := storage.NewMemoryStorage()
	rateLimiter := limiter.NewRateLimiter(store)
	middleware := RateLimiterMiddleware(cfg, rateLimiter)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))

	// Test requests with token within limit
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("API_KEY", "test-token")
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}

	// Test request over limit
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("API_KEY", "test-token")
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code, "Request over limit should be blocked")
}

func TestRateLimiterMiddleware_TokenOverridesIP(t *testing.T) {
	// Setup: IP limit is 2, Token limit is 5
	cfg := &configs.Config{
		RateLimitIP:             2,
		RateLimitIPBlockTime:    1,
		RateLimitToken:          5,
		RateLimitTokenBlockTime: 1,
		TokenConfigs:            make(map[string]configs.TokenConfig),
	}
	store := storage.NewMemoryStorage()
	rateLimiter := limiter.NewRateLimiter(store)
	middleware := RateLimiterMiddleware(cfg, rateLimiter)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))

	// Make 5 requests with token (should all succeed even though IP limit is 2)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("API_KEY", "test-token")
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Request %d with token should succeed", i+1)
	}

	// 6th request should be blocked
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("API_KEY", "test-token")
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code, "Request over token limit should be blocked")
}

func TestRateLimiterMiddleware_SpecificTokenConfig(t *testing.T) {
	// Setup with specific token configuration
	cfg := &configs.Config{
		RateLimitIP:             5,
		RateLimitIPBlockTime:    1,
		RateLimitToken:          10,
		RateLimitTokenBlockTime: 1,
		TokenConfigs: map[string]configs.TokenConfig{
			"abc123": {
				Limit:     3,
				BlockTime: 2,
			},
		},
	}
	store := storage.NewMemoryStorage()
	rateLimiter := limiter.NewRateLimiter(store)
	middleware := RateLimiterMiddleware(cfg, rateLimiter)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))

	// Make 3 requests with specific token
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("API_KEY", "abc123")
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}

	// 4th request should be blocked (limit is 3)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("API_KEY", "abc123")
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code, "Request over token limit should be blocked")
}

func TestRateLimiterMiddleware_DifferentIPs(t *testing.T) {
	// Setup
	cfg := &configs.Config{
		RateLimitIP:          2,
		RateLimitIPBlockTime: 1,
	}
	store := storage.NewMemoryStorage()
	rateLimiter := limiter.NewRateLimiter(store)
	middleware := RateLimiterMiddleware(cfg, rateLimiter)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))

	// IP1: Make 2 requests
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// IP1: 3rd request should be blocked
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// IP2: Should still be allowed
	req = httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.2:1234"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimiterMiddleware_BlockDuration(t *testing.T) {
	// Setup with short block time
	cfg := &configs.Config{
		RateLimitIP:          2,
		RateLimitIPBlockTime: 1, // 1 second
	}
	store := storage.NewMemoryStorage()
	rateLimiter := limiter.NewRateLimiter(store)
	middleware := RateLimiterMiddleware(cfg, rateLimiter)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))

	// Exceed limit
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	// Should be blocked
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// Wait for block to expire
	time.Sleep(1200 * time.Millisecond)

	// Should be allowed again
	req = httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

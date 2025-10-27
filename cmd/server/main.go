package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/jonilsonds9/goexpert-desafio-rate-limiter/configs"
	"github.com/jonilsonds9/goexpert-desafio-rate-limiter/internal/limiter"
	"github.com/jonilsonds9/goexpert-desafio-rate-limiter/internal/middleware"
	"github.com/jonilsonds9/goexpert-desafio-rate-limiter/internal/storage"
)

func main() {
	cfg, err := configs.LoadConfig()
	if err != nil {
		panic(err)
	}

	var store storage.Storage

	redisStore, err := storage.NewRedisStorage(cfg.RedisHost, cfg.RedisPort, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		log.Println("Falling back to in-memory storage")
		store = storage.NewMemoryStorage()
	} else {
		log.Println("Connected to Redis successfully")
		store = redisStore
	}

	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("Error closing storage: %v", err)
		}
	}()

	rateLimiter := limiter.NewRateLimiter(store)

	mux := http.NewServeMux()

	// Health check endpoint (no rate limiting)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
		})
	})

	// Main endpoint with rate limiting
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Welcome to Rate Limiter API",
			"status":  "ok",
		})
	})

	handler := applyMiddleware(mux, cfg, rateLimiter)

	addr := fmt.Sprintf(":%s", cfg.ServerPort)

	log.Printf("Starting server on port %s", cfg.ServerPort)
	log.Printf("Rate Limit IP: %d req/s (block time: %ds)", cfg.RateLimitIP, cfg.RateLimitIPBlockTime)
	log.Printf("Rate Limit Token (default): %d req/s (block time: %ds)", cfg.RateLimitToken, cfg.RateLimitTokenBlockTime)

	if len(cfg.TokenConfigs) > 0 {
		log.Println("Token-specific configurations:")
		for token, tc := range cfg.TokenConfigs {
			log.Printf("  - %s: %d req/s (block time: %ds)", token, tc.Limit, tc.BlockTime)
		}
	}

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func applyMiddleware(mux *http.ServeMux, cfg *configs.Config, rateLimiter *limiter.RateLimiter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			mux.ServeHTTP(w, r)
			return
		}

		rateLimiterMiddleware := middleware.RateLimiterMiddleware(cfg, rateLimiter)
		rateLimiterMiddleware(mux).ServeHTTP(w, r)
	})
}

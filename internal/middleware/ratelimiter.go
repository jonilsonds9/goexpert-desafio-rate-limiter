package middleware

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/jonilsonds9/goexpert-desafio-rate-limiter/configs"
	"github.com/jonilsonds9/goexpert-desafio-rate-limiter/internal/limiter"
)

func RateLimiterMiddleware(cfg *configs.Config, limiter *limiter.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			token := r.Header.Get("API_KEY")

			var key string
			var limit int
			var blockDuration time.Duration

			if token != "" {
				// Use token-based limiting (priority over IP)
				key = fmt.Sprintf("token:%s", token)

				// Get token-specific configuration
				tokenConfig, exists := cfg.GetTokenConfig(token)
				if exists {
					limit = tokenConfig.Limit
					blockDuration = time.Duration(tokenConfig.BlockTime) * time.Second
				} else {
					// Use default token limits
					limit = cfg.RateLimitToken
					blockDuration = time.Duration(cfg.RateLimitTokenBlockTime) * time.Second
				}
			} else {
				// Use IP-based limiting
				ip := getClientIP(r)
				key = fmt.Sprintf("ip:%s", ip)
				limit = cfg.RateLimitIP
				blockDuration = time.Duration(cfg.RateLimitIPBlockTime) * time.Second
			}

			allowed, err := limiter.AllowRequest(ctx, key, limit, blockDuration)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if !allowed {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "you have reached the maximum number of requests or actions allowed within a certain time frame",
				})
				return
			}

			next.ServeHTTP(w, r) // Request is allowed, continue to next handler
		})
	}
}

func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

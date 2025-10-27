package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/jonilsonds9/goexpert-desafio-rate-limiter/internal/storage"
)

type RateLimiter struct {
	storage storage.Storage
}

func NewRateLimiter(storage storage.Storage) *RateLimiter {
	return &RateLimiter{
		storage: storage,
	}
}

// AllowRequest checks if a request should be allowed
func (rl *RateLimiter) AllowRequest(ctx context.Context, key string, limit int, blockDuration time.Duration) (bool, error) {
	// Check if already blocked
	blocked, err := rl.storage.IsBlocked(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to check block status: %w", err)
	}

	if blocked {
		return false, nil
	}

	// Increment counter with 1-second window
	count, err := rl.storage.Increment(ctx, key, 1*time.Second)
	if err != nil {
		return false, fmt.Errorf("failed to increment counter: %w", err)
	}

	// Check if limit exceeded
	if count > int64(limit) {
		// Block the key
		if err := rl.storage.SetBlock(ctx, key, blockDuration); err != nil {
			return false, fmt.Errorf("failed to set block: %w", err)
		}
		return false, nil
	}

	return true, nil
}

func (rl *RateLimiter) IsBlocked(ctx context.Context, key string) (bool, error) {
	return rl.storage.IsBlocked(ctx, key)
}

func (rl *RateLimiter) GetCurrentCount(ctx context.Context, key string) (int64, error) {
	return rl.storage.Get(ctx, key)
}

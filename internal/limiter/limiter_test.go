package limiter

import (
	"context"
	"testing"
	"time"

	"github.com/jonilsonds9/goexpert-desafio-rate-limiter/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_AllowRequest(t *testing.T) {
	store := storage.NewMemoryStorage()
	limiter := NewRateLimiter(store)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		allowed, err := limiter.AllowRequest(ctx, "test-key", 5, 1*time.Second)
		assert.NoError(t, err)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	allowed, err := limiter.AllowRequest(ctx, "test-key", 5, 1*time.Second)
	assert.NoError(t, err)
	assert.False(t, allowed, "Request over limit should be blocked")
}

func TestRateLimiter_BlockDuration(t *testing.T) {
	store := storage.NewMemoryStorage()
	limiter := NewRateLimiter(store)
	ctx := context.Background()

	for i := 0; i <= 5; i++ {
		limiter.AllowRequest(ctx, "test-key", 5, 100*time.Millisecond)
	}

	allowed, err := limiter.AllowRequest(ctx, "test-key", 5, 100*time.Millisecond)
	assert.NoError(t, err)
	assert.False(t, allowed)

	time.Sleep(1200 * time.Millisecond)

	allowed, err = limiter.AllowRequest(ctx, "test-key", 5, 100*time.Millisecond)
	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestRateLimiter_IsBlocked(t *testing.T) {
	store := storage.NewMemoryStorage()
	limiter := NewRateLimiter(store)
	ctx := context.Background()

	blocked, err := limiter.IsBlocked(ctx, "test-key")
	assert.NoError(t, err)
	assert.False(t, blocked)

	for i := 0; i <= 5; i++ {
		limiter.AllowRequest(ctx, "test-key", 5, 1*time.Second)
	}

	blocked, err = limiter.IsBlocked(ctx, "test-key")
	assert.NoError(t, err)
	assert.True(t, blocked)
}

func TestRateLimiter_GetCurrentCount(t *testing.T) {
	store := storage.NewMemoryStorage()
	limiter := NewRateLimiter(store)
	ctx := context.Background()

	count, err := limiter.GetCurrentCount(ctx, "test-key")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	limiter.AllowRequest(ctx, "test-key", 10, 1*time.Second)
	limiter.AllowRequest(ctx, "test-key", 10, 1*time.Second)
	limiter.AllowRequest(ctx, "test-key", 10, 1*time.Second)

	count, err = limiter.GetCurrentCount(ctx, "test-key")
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestRateLimiter_DifferentKeys(t *testing.T) {
	store := storage.NewMemoryStorage()
	limiter := NewRateLimiter(store)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		allowed, err := limiter.AllowRequest(ctx, "key1", 5, 1*time.Second)
		assert.NoError(t, err)
		assert.True(t, allowed)
	}

	allowed, err := limiter.AllowRequest(ctx, "key1", 5, 1*time.Second)
	assert.NoError(t, err)
	assert.False(t, allowed)

	allowed, err = limiter.AllowRequest(ctx, "key2", 5, 1*time.Second)
	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestRateLimiter_WindowReset(t *testing.T) {
	store := storage.NewMemoryStorage()
	limiter := NewRateLimiter(store)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		limiter.AllowRequest(ctx, "test-key", 5, 200*time.Millisecond)
	}

	allowed, err := limiter.AllowRequest(ctx, "test-key", 5, 200*time.Millisecond)
	assert.NoError(t, err)
	assert.False(t, allowed)

	time.Sleep(1100 * time.Millisecond)

	allowed, err = limiter.AllowRequest(ctx, "test-key", 5, 200*time.Millisecond)
	assert.NoError(t, err)
	assert.True(t, allowed)
}

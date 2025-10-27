package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryStorage_Increment(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()

	// Test first increment
	count, err := storage.Increment(ctx, "test-key", 1*time.Second)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Test second increment
	count, err = storage.Increment(ctx, "test-key", 1*time.Second)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Test third increment
	count, err = storage.Increment(ctx, "test-key", 1*time.Second)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestMemoryStorage_IncrementExpiration(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()

	// Increment with short expiration
	count, err := storage.Increment(ctx, "test-key", 100*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should reset to 1
	count, err = storage.Increment(ctx, "test-key", 1*time.Second)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestMemoryStorage_Get(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()

	// Get non-existent key
	count, err := storage.Get(ctx, "non-existent")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Increment and get
	_, err = storage.Increment(ctx, "test-key", 1*time.Second)
	assert.NoError(t, err)

	count, err = storage.Get(ctx, "test-key")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestMemoryStorage_Block(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()

	// Check not blocked
	blocked, err := storage.IsBlocked(ctx, "test-key")
	assert.NoError(t, err)
	assert.False(t, blocked)

	// Set block
	err = storage.SetBlock(ctx, "test-key", 100*time.Millisecond)
	assert.NoError(t, err)

	// Check blocked
	blocked, err = storage.IsBlocked(ctx, "test-key")
	assert.NoError(t, err)
	assert.True(t, blocked)

	// Wait for block to expire
	time.Sleep(150 * time.Millisecond)

	// Check not blocked anymore
	blocked, err = storage.IsBlocked(ctx, "test-key")
	assert.NoError(t, err)
	assert.False(t, blocked)
}

func TestMemoryStorage_Concurrent(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()

	// Concurrent increments
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := storage.Increment(ctx, "test-key", 1*time.Second)
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check final count
	count, err := storage.Get(ctx, "test-key")
	assert.NoError(t, err)
	assert.Equal(t, int64(10), count)
}

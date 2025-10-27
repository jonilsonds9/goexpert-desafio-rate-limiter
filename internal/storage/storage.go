package storage

import (
	"context"
	"time"
)

type Storage interface {
	// Increment increments the counter for a given key and returns the current count
	Increment(ctx context.Context, key string, expiration time.Duration) (int64, error)

	// Get retrieves the current count for a given key
	Get(ctx context.Context, key string) (int64, error)

	// SetBlock blocks a key for a specified duration
	SetBlock(ctx context.Context, key string, duration time.Duration) error

	// IsBlocked checks if a key is currently blocked
	IsBlocked(ctx context.Context, key string) (bool, error)

	// Close closes the storage connection
	Close() error
}

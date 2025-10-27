package storage

import (
	"context"
	"sync"
	"time"
)

type MemoryStorage struct {
	counters map[string]counterEntry
	blocks   map[string]time.Time
	mu       sync.RWMutex
}

type counterEntry struct {
	count      int64
	expiration time.Time
}

func NewMemoryStorage() *MemoryStorage {
	storage := &MemoryStorage{
		counters: make(map[string]counterEntry),
		blocks:   make(map[string]time.Time),
	}

	// Start cleanup goroutine
	go storage.cleanup()

	return storage
}

func (m *MemoryStorage) Increment(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	entry, exists := m.counters[key]

	if !exists || now.After(entry.expiration) {
		// Create new entry
		m.counters[key] = counterEntry{
			count:      1,
			expiration: now.Add(expiration),
		}
		return 1, nil
	}

	// Increment existing entry
	entry.count++
	m.counters[key] = entry
	return entry.count, nil
}

func (m *MemoryStorage) Get(ctx context.Context, key string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.counters[key]
	if !exists {
		return 0, nil
	}

	if time.Now().After(entry.expiration) {
		return 0, nil
	}

	return entry.count, nil
}

func (m *MemoryStorage) SetBlock(ctx context.Context, key string, duration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.blocks[key] = time.Now().Add(duration)
	return nil
}

func (m *MemoryStorage) IsBlocked(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	blockUntil, exists := m.blocks[key]
	if !exists {
		return false, nil
	}

	if time.Now().After(blockUntil) {
		return false, nil
	}

	return true, nil
}

func (m *MemoryStorage) Close() error {
	return nil
}

// cleanup removes expired entries periodically
func (m *MemoryStorage) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()

		// Clean up expired counters
		for key, entry := range m.counters {
			if now.After(entry.expiration) {
				delete(m.counters, key)
			}
		}

		// Clean up expired blocks
		for key, blockUntil := range m.blocks {
			if now.After(blockUntil) {
				delete(m.blocks, key)
			}
		}

		m.mu.Unlock()
	}
}

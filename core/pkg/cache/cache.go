package cache

import (
	"context"
	"time"
)

// Cache is the interface that all cache implementations must satisfy
type Cache interface {
	// Get retrieves a value from the cache
	Get(ctx context.Context, key string) (interface{}, error)
	
	// Set stores a value in the cache with TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	
	// Delete removes a value from the cache
	Delete(ctx context.Context, key string) error
	
	// Exists checks if a key exists in the cache
	Exists(ctx context.Context, key string) (bool, error)
	
	// Clear removes all values from the cache
	Clear(ctx context.Context) error
	
	// Keys returns all keys matching the pattern
	Keys(ctx context.Context, pattern string) ([]string, error)
	
	// TTL returns the remaining time to live for a key
	TTL(ctx context.Context, key string) (time.Duration, error)
	
	// Expire sets a new TTL for a key
	Expire(ctx context.Context, key string, ttl time.Duration) error
	
	// Increment atomically increments a counter
	Increment(ctx context.Context, key string, delta int64) (int64, error)
	
	// Decrement atomically decrements a counter
	Decrement(ctx context.Context, key string, delta int64) (int64, error)
	
	// GetMulti retrieves multiple values
	GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error)
	
	// SetMulti stores multiple values
	SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error
	
	// DeleteMulti removes multiple values
	DeleteMulti(ctx context.Context, keys []string) error
	
	// Close closes the cache connection
	Close() error
}

// Stats represents cache statistics
type Stats struct {
	Hits        uint64
	Misses      uint64
	Keys        uint64
	Memory      uint64 // bytes
	Evictions   uint64
	Connections uint64
}

// StatsProvider provides cache statistics
type StatsProvider interface {
	Stats(ctx context.Context) (*Stats, error)
}

// CacheError represents a cache error
type CacheError struct {
	Op  string
	Key string
	Err error
}

func (e *CacheError) Error() string {
	if e.Key != "" {
		return "cache." + e.Op + " " + e.Key + ": " + e.Err.Error()
	}
	return "cache." + e.Op + ": " + e.Err.Error()
}

func (e *CacheError) Unwrap() error {
	return e.Err
}

// Common errors
var (
	ErrKeyNotFound = &CacheError{Op: "get", Err: ErrNotFound}
	ErrNotFound    = &CacheError{Err: nil}
	ErrTimeout     = &CacheError{Op: "timeout", Err: nil}
	ErrClosed      = &CacheError{Op: "closed", Err: nil}
)

// Config is the base configuration for all caches
type Config struct {
	// DefaultTTL is the default time-to-live for cache entries
	DefaultTTL time.Duration
	
	// MaxRetries is the maximum number of retries for failed operations
	MaxRetries int
	
	// RetryDelay is the delay between retries
	RetryDelay time.Duration
	
	// Timeout is the operation timeout
	Timeout time.Duration
}

// DefaultConfig returns the default cache configuration
func DefaultConfig() Config {
	return Config{
		DefaultTTL: 5 * time.Minute,
		MaxRetries: 3,
		RetryDelay: 100 * time.Millisecond,
		Timeout:    5 * time.Second,
	}
}

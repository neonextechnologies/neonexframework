package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache is a Redis-based cache implementation
type RedisCache struct {
	client *redis.Client
	config Config
	stats  Stats
}

// RedisCacheConfig configures the Redis cache
type RedisCacheConfig struct {
	Config
	Addr         string // Redis address (host:port)
	Password     string // Redis password
	DB           int    // Redis database number
	PoolSize     int    // Connection pool size
	MinIdleConns int    // Minimum idle connections
	MaxRetries   int    // Maximum number of retries
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// DefaultRedisCacheConfig returns the default Redis cache configuration
func DefaultRedisCacheConfig() RedisCacheConfig {
	return RedisCacheConfig{
		Config:       DefaultConfig(),
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 2,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(config RedisCacheConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, &CacheError{Op: "connect", Err: err}
	}

	return &RedisCache{
		client: client,
		config: config.Config,
	}, nil
}

// Get retrieves a value from the cache
func (rc *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	val, err := rc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		rc.stats.Misses++
		return nil, ErrKeyNotFound
	}
	if err != nil {
		return nil, &CacheError{Op: "get", Key: key, Err: err}
	}

	rc.stats.Hits++

	// Try to unmarshal as JSON
	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		// Return as string if not JSON
		return val, nil
	}

	return result, nil
}

// Set stores a value in the cache with TTL
func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = rc.config.DefaultTTL
	}

	// Serialize value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return &CacheError{Op: "set", Key: key, Err: err}
	}

	if err := rc.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return &CacheError{Op: "set", Key: key, Err: err}
	}

	return nil
}

// Delete removes a value from the cache
func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	if err := rc.client.Del(ctx, key).Err(); err != nil {
		return &CacheError{Op: "delete", Key: key, Err: err}
	}
	return nil
}

// Exists checks if a key exists in the cache
func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := rc.client.Exists(ctx, key).Result()
	if err != nil {
		return false, &CacheError{Op: "exists", Key: key, Err: err}
	}
	return n > 0, nil
}

// Clear removes all values from the cache
func (rc *RedisCache) Clear(ctx context.Context) error {
	if err := rc.client.FlushDB(ctx).Err(); err != nil {
		return &CacheError{Op: "clear", Err: err}
	}
	return nil
}

// Keys returns all keys matching the pattern
func (rc *RedisCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	keys, err := rc.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, &CacheError{Op: "keys", Err: err}
	}
	return keys, nil
}

// TTL returns the remaining time to live for a key
func (rc *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := rc.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, &CacheError{Op: "ttl", Key: key, Err: err}
	}
	if ttl < 0 {
		return 0, nil
	}
	return ttl, nil
}

// Expire sets a new TTL for a key
func (rc *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := rc.client.Expire(ctx, key, ttl).Err(); err != nil {
		return &CacheError{Op: "expire", Key: key, Err: err}
	}
	return nil
}

// Increment atomically increments a counter
func (rc *RedisCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	val, err := rc.client.IncrBy(ctx, key, delta).Result()
	if err != nil {
		return 0, &CacheError{Op: "increment", Key: key, Err: err}
	}
	return val, nil
}

// Decrement atomically decrements a counter
func (rc *RedisCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	val, err := rc.client.DecrBy(ctx, key, delta).Result()
	if err != nil {
		return 0, &CacheError{Op: "decrement", Key: key, Err: err}
	}
	return val, nil
}

// GetMulti retrieves multiple values
func (rc *RedisCache) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return map[string]interface{}{}, nil
	}

	vals, err := rc.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, &CacheError{Op: "mget", Err: err}
	}

	result := make(map[string]interface{})
	for i, val := range vals {
		if val != nil {
			// Try to unmarshal as JSON
			var v interface{}
			if err := json.Unmarshal([]byte(val.(string)), &v); err == nil {
				result[keys[i]] = v
			} else {
				result[keys[i]] = val
			}
		}
	}

	return result, nil
}

// SetMulti stores multiple values
func (rc *RedisCache) SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = rc.config.DefaultTTL
	}

	pipe := rc.client.Pipeline()

	for key, value := range items {
		data, err := json.Marshal(value)
		if err != nil {
			return &CacheError{Op: "mset", Key: key, Err: err}
		}
		pipe.Set(ctx, key, data, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return &CacheError{Op: "mset", Err: err}
	}

	return nil
}

// DeleteMulti removes multiple values
func (rc *RedisCache) DeleteMulti(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	if err := rc.client.Del(ctx, keys...).Err(); err != nil {
		return &CacheError{Op: "mdel", Err: err}
	}

	return nil
}

// Stats returns cache statistics
func (rc *RedisCache) Stats(ctx context.Context) (*Stats, error) {
	info, err := rc.client.Info(ctx, "stats", "memory").Result()
	if err != nil {
		return nil, &CacheError{Op: "stats", Err: err}
	}

	stats := &Stats{
		Hits:   rc.stats.Hits,
		Misses: rc.stats.Misses,
	}

	// Parse Redis INFO output
	// This is a simplified version - in production, parse all relevant fields
	// stats.Memory = parseMemoryUsage(info)
	// stats.Evictions = parseEvictions(info)
	// stats.Connections = parseConnections(info)

	_ = info // Use info to parse stats

	// Get key count
	dbSize, err := rc.client.DBSize(ctx).Result()
	if err == nil {
		stats.Keys = uint64(dbSize)
	}

	return stats, nil
}

// Close closes the cache connection
func (rc *RedisCache) Close() error {
	return rc.client.Close()
}

// Pipeline returns a Redis pipeline for batch operations
func (rc *RedisCache) Pipeline() redis.Pipeliner {
	return rc.client.Pipeline()
}

// Client returns the underlying Redis client
func (rc *RedisCache) Client() *redis.Client {
	return rc.client
}

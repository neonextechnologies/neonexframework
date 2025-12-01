package ai

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// InferenceCache caches inference results
type InferenceCache struct {
	cache      map[string]*cacheEntry
	maxSize    int
	ttl        time.Duration
	mu         sync.RWMutex
	hits       int64
	misses     int64
	evictions  int64
}

type cacheEntry struct {
	output    *InferenceOutput
	expiresAt time.Time
}

// NewInferenceCache creates a new inference cache
func NewInferenceCache(maxSize int, ttl time.Duration) *InferenceCache {
	cache := &InferenceCache{
		cache:   make(map[string]*cacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
	}

	// Start cleanup goroutine
	go cache.cleanupLoop()

	return cache
}

// Get gets a cached result
func (c *InferenceCache) Get(input *InferenceInput) *InferenceOutput {
	key := c.generateKey(input)

	c.mu.RLock()
	entry, exists := c.cache[key]
	c.mu.RUnlock()

	if !exists {
		c.mu.Lock()
		c.misses++
		c.mu.Unlock()
		return nil
	}

	// Check expiration
	if time.Now().After(entry.expiresAt) {
		c.mu.Lock()
		delete(c.cache, key)
		c.misses++
		c.mu.Unlock()
		return nil
	}

	c.mu.Lock()
	c.hits++
	c.mu.Unlock()

	return entry.output
}

// Set sets a cached result
func (c *InferenceCache) Set(input *InferenceInput, output *InferenceOutput) {
	key := c.generateKey(input)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict if cache is full
	if len(c.cache) >= c.maxSize {
		c.evictOldest()
	}

	c.cache[key] = &cacheEntry{
		output:    output,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Clear clears the cache
func (c *InferenceCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]*cacheEntry)
}

// GetStats returns cache statistics
func (c *InferenceCache) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return map[string]interface{}{
		"size":       len(c.cache),
		"max_size":   c.maxSize,
		"hits":       c.hits,
		"misses":     c.misses,
		"hit_rate":   hitRate,
		"evictions":  c.evictions,
	}
}

// generateKey generates cache key from input
func (c *InferenceCache) generateKey(input *InferenceInput) string {
	data, _ := json.Marshal(map[string]interface{}{
		"model_id":   input.ModelID,
		"data":       input.Data,
		"parameters": input.Parameters,
	})
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// evictOldest evicts the oldest entry
func (c *InferenceCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.cache {
		if oldestKey == "" || entry.expiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.expiresAt
		}
	}

	if oldestKey != "" {
		delete(c.cache, oldestKey)
		c.evictions++
	}
}

// cleanupLoop periodically cleans up expired entries
func (c *InferenceCache) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup removes expired entries
func (c *InferenceCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.cache {
		if now.After(entry.expiresAt) {
			delete(c.cache, key)
		}
	}
}

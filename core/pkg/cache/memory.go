package cache

import (
	"container/list"
	"context"
	"sync"
	"time"
)

// MemoryCache is an in-memory LRU cache implementation
type MemoryCache struct {
	mu        sync.RWMutex
	items     map[string]*list.Element
	lru       *list.List
	maxSize   int
	stats     Stats
	config    Config
	closed    bool
	closeChan chan struct{}
}

// cacheItem represents an item in the cache
type cacheItem struct {
	key       string
	value     interface{}
	expiresAt time.Time
}

// MemoryCacheConfig configures the memory cache
type MemoryCacheConfig struct {
	Config
	MaxSize         int           // Maximum number of items
	CleanupInterval time.Duration // Interval for cleanup of expired items
}

// DefaultMemoryCacheConfig returns the default memory cache configuration
func DefaultMemoryCacheConfig() MemoryCacheConfig {
	return MemoryCacheConfig{
		Config:          DefaultConfig(),
		MaxSize:         10000,
		CleanupInterval: 1 * time.Minute,
	}
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(config MemoryCacheConfig) *MemoryCache {
	mc := &MemoryCache{
		items:     make(map[string]*list.Element),
		lru:       list.New(),
		maxSize:   config.MaxSize,
		config:    config.Config,
		closeChan: make(chan struct{}),
	}
	
	// Start cleanup goroutine
	if config.CleanupInterval > 0 {
		go mc.cleanupLoop(config.CleanupInterval)
	}
	
	return mc
}

// Get retrieves a value from the cache
func (mc *MemoryCache) Get(ctx context.Context, key string) (interface{}, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	if mc.closed {
		return nil, ErrClosed
	}
	
	elem, found := mc.items[key]
	if !found {
		mc.stats.Misses++
		return nil, ErrKeyNotFound
	}
	
	item := elem.Value.(*cacheItem)
	
	// Check if expired
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		mc.removeElement(elem)
		mc.stats.Misses++
		return nil, ErrKeyNotFound
	}
	
	// Move to front (most recently used)
	mc.lru.MoveToFront(elem)
	mc.stats.Hits++
	
	return item.value, nil
}

// Set stores a value in the cache with TTL
func (mc *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	if mc.closed {
		return ErrClosed
	}
	
	if ttl == 0 {
		ttl = mc.config.DefaultTTL
	}
	
	// Calculate expiration time
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}
	
	// Update existing item
	if elem, found := mc.items[key]; found {
		item := elem.Value.(*cacheItem)
		item.value = value
		item.expiresAt = expiresAt
		mc.lru.MoveToFront(elem)
		return nil
	}
	
	// Add new item
	item := &cacheItem{
		key:       key,
		value:     value,
		expiresAt: expiresAt,
	}
	
	elem := mc.lru.PushFront(item)
	mc.items[key] = elem
	mc.stats.Keys++
	
	// Evict if necessary
	if mc.lru.Len() > mc.maxSize {
		mc.evict()
	}
	
	return nil
}

// Delete removes a value from the cache
func (mc *MemoryCache) Delete(ctx context.Context, key string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	if mc.closed {
		return ErrClosed
	}
	
	elem, found := mc.items[key]
	if !found {
		return nil
	}
	
	mc.removeElement(elem)
	return nil
}

// Exists checks if a key exists in the cache
func (mc *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	if mc.closed {
		return false, ErrClosed
	}
	
	elem, found := mc.items[key]
	if !found {
		return false, nil
	}
	
	item := elem.Value.(*cacheItem)
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		return false, nil
	}
	
	return true, nil
}

// Clear removes all values from the cache
func (mc *MemoryCache) Clear(ctx context.Context) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	if mc.closed {
		return ErrClosed
	}
	
	mc.items = make(map[string]*list.Element)
	mc.lru.Init()
	mc.stats.Keys = 0
	
	return nil
}

// Keys returns all keys matching the pattern
func (mc *MemoryCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	if mc.closed {
		return nil, ErrClosed
	}
	
	keys := make([]string, 0, len(mc.items))
	now := time.Now()
	
	for key, elem := range mc.items {
		item := elem.Value.(*cacheItem)
		
		// Skip expired items
		if !item.expiresAt.IsZero() && now.After(item.expiresAt) {
			continue
		}
		
		// Simple pattern matching (supports * wildcard)
		if pattern == "*" || matchPattern(key, pattern) {
			keys = append(keys, key)
		}
	}
	
	return keys, nil
}

// TTL returns the remaining time to live for a key
func (mc *MemoryCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	if mc.closed {
		return 0, ErrClosed
	}
	
	elem, found := mc.items[key]
	if !found {
		return 0, ErrKeyNotFound
	}
	
	item := elem.Value.(*cacheItem)
	if item.expiresAt.IsZero() {
		return 0, nil // No expiration
	}
	
	ttl := time.Until(item.expiresAt)
	if ttl < 0 {
		return 0, nil
	}
	
	return ttl, nil
}

// Expire sets a new TTL for a key
func (mc *MemoryCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	if mc.closed {
		return ErrClosed
	}
	
	elem, found := mc.items[key]
	if !found {
		return ErrKeyNotFound
	}
	
	item := elem.Value.(*cacheItem)
	if ttl > 0 {
		item.expiresAt = time.Now().Add(ttl)
	} else {
		item.expiresAt = time.Time{}
	}
	
	return nil
}

// Increment atomically increments a counter
func (mc *MemoryCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	if mc.closed {
		return 0, ErrClosed
	}
	
	elem, found := mc.items[key]
	if !found {
		// Create new counter
		item := &cacheItem{
			key:   key,
			value: delta,
		}
		elem = mc.lru.PushFront(item)
		mc.items[key] = elem
		mc.stats.Keys++
		return delta, nil
	}
	
	item := elem.Value.(*cacheItem)
	val, ok := item.value.(int64)
	if !ok {
		return 0, &CacheError{Op: "increment", Key: key, Err: ErrNotFound}
	}
	
	val += delta
	item.value = val
	mc.lru.MoveToFront(elem)
	
	return val, nil
}

// Decrement atomically decrements a counter
func (mc *MemoryCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return mc.Increment(ctx, key, -delta)
}

// GetMulti retrieves multiple values
func (mc *MemoryCache) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	for _, key := range keys {
		if value, err := mc.Get(ctx, key); err == nil {
			result[key] = value
		}
	}
	
	return result, nil
}

// SetMulti stores multiple values
func (mc *MemoryCache) SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	for key, value := range items {
		if err := mc.Set(ctx, key, value, ttl); err != nil {
			return err
		}
	}
	return nil
}

// DeleteMulti removes multiple values
func (mc *MemoryCache) DeleteMulti(ctx context.Context, keys []string) error {
	for _, key := range keys {
		if err := mc.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

// Stats returns cache statistics
func (mc *MemoryCache) Stats(ctx context.Context) (*Stats, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	statsCopy := mc.stats
	statsCopy.Keys = uint64(len(mc.items))
	
	return &statsCopy, nil
}

// Close closes the cache
func (mc *MemoryCache) Close() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	if mc.closed {
		return nil
	}
	
	mc.closed = true
	close(mc.closeChan)
	mc.items = nil
	mc.lru = nil
	
	return nil
}

// removeElement removes an element from the cache
func (mc *MemoryCache) removeElement(elem *list.Element) {
	item := elem.Value.(*cacheItem)
	delete(mc.items, item.key)
	mc.lru.Remove(elem)
	mc.stats.Keys--
}

// evict removes the least recently used item
func (mc *MemoryCache) evict() {
	elem := mc.lru.Back()
	if elem != nil {
		mc.removeElement(elem)
		mc.stats.Evictions++
	}
}

// cleanupLoop periodically removes expired items
func (mc *MemoryCache) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			mc.cleanup()
		case <-mc.closeChan:
			return
		}
	}
}

// cleanup removes all expired items
func (mc *MemoryCache) cleanup() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	if mc.closed {
		return
	}
	
	now := time.Now()
	toRemove := []*list.Element{}
	
	for elem := mc.lru.Front(); elem != nil; elem = elem.Next() {
		item := elem.Value.(*cacheItem)
		if !item.expiresAt.IsZero() && now.After(item.expiresAt) {
			toRemove = append(toRemove, elem)
		}
	}
	
	for _, elem := range toRemove {
		mc.removeElement(elem)
	}
}

// matchPattern performs simple pattern matching
func matchPattern(str, pattern string) bool {
	// Simple implementation - in production use proper glob matching
	if pattern == "*" {
		return true
	}
	// TODO: Implement proper pattern matching
	return str == pattern
}

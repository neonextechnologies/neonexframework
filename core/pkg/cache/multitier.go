package cache

import (
	"context"
	"sync"
	"time"
)

// TierLevel represents cache tier priority
type TierLevel int

const (
	TierL1 TierLevel = 1 // Fastest (memory)
	TierL2 TierLevel = 2 // Fast (Redis)
	TierL3 TierLevel = 3 // Slower (disk/remote)
)

// MultiTierCache implements a multi-tier caching strategy
type MultiTierCache struct {
	tiers      []cacheWithLevel
	mu         sync.RWMutex
	config     Config
	promoteL1  bool // Promote hits to L1 cache
	writeThru  bool // Write-through to all tiers
	writeBack  bool // Write-back strategy
	stats      Stats
}

type cacheWithLevel struct {
	cache Cache
	level TierLevel
}

// MultiTierConfig configures the multi-tier cache
type MultiTierConfig struct {
	Config
	PromoteL1 bool // Promote cache hits to L1
	WriteThru bool // Write to all tiers immediately
	WriteBack bool // Write to lower tiers asynchronously
}

// DefaultMultiTierConfig returns default configuration
func DefaultMultiTierConfig() MultiTierConfig {
	return MultiTierConfig{
		Config:    DefaultConfig(),
		PromoteL1: true,
		WriteThru: true,
		WriteBack: false,
	}
}

// NewMultiTierCache creates a new multi-tier cache
func NewMultiTierCache(config MultiTierConfig) *MultiTierCache {
	return &MultiTierCache{
		tiers:     []cacheWithLevel{},
		config:    config.Config,
		promoteL1: config.PromoteL1,
		writeThru: config.WriteThru,
		writeBack: config.WriteBack,
	}
}

// AddTier adds a cache tier
func (mtc *MultiTierCache) AddTier(cache Cache, level TierLevel) {
	mtc.mu.Lock()
	defer mtc.mu.Unlock()

	mtc.tiers = append(mtc.tiers, cacheWithLevel{
		cache: cache,
		level: level,
	})

	// Sort tiers by level (L1 first)
	mtc.sortTiers()
}

// Get retrieves a value from the cache (tries L1, L2, L3 in order)
func (mtc *MultiTierCache) Get(ctx context.Context, key string) (interface{}, error) {
	mtc.mu.RLock()
	defer mtc.mu.RUnlock()

	for i, tier := range mtc.tiers {
		value, err := tier.cache.Get(ctx, key)
		if err == nil {
			mtc.stats.Hits++

			// Promote to higher tiers
			if mtc.promoteL1 && i > 0 {
				go mtc.promoteToHigherTiers(key, value, i)
			}

			return value, nil
		}

		// Continue to next tier if not found
		if err == ErrKeyNotFound {
			continue
		}

		// Return other errors immediately
		return nil, err
	}

	mtc.stats.Misses++
	return nil, ErrKeyNotFound
}

// Set stores a value in all cache tiers
func (mtc *MultiTierCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	mtc.mu.RLock()
	defer mtc.mu.RUnlock()

	if ttl == 0 {
		ttl = mtc.config.DefaultTTL
	}

	if mtc.writeThru {
		// Write to all tiers synchronously
		for _, tier := range mtc.tiers {
			if err := tier.cache.Set(ctx, key, value, ttl); err != nil {
				return err
			}
		}
		return nil
	}

	// Write to L1 only, write-back to others asynchronously
	if len(mtc.tiers) > 0 {
		if err := mtc.tiers[0].cache.Set(ctx, key, value, ttl); err != nil {
			return err
		}

		if mtc.writeBack && len(mtc.tiers) > 1 {
			go mtc.writeToLowerTiers(key, value, ttl)
		}
	}

	return nil
}

// Delete removes a value from all cache tiers
func (mtc *MultiTierCache) Delete(ctx context.Context, key string) error {
	mtc.mu.RLock()
	defer mtc.mu.RUnlock()

	var lastErr error
	for _, tier := range mtc.tiers {
		if err := tier.cache.Delete(ctx, key); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// Exists checks if a key exists in any tier
func (mtc *MultiTierCache) Exists(ctx context.Context, key string) (bool, error) {
	mtc.mu.RLock()
	defer mtc.mu.RUnlock()

	for _, tier := range mtc.tiers {
		exists, err := tier.cache.Exists(ctx, key)
		if err != nil {
			continue
		}
		if exists {
			return true, nil
		}
	}

	return false, nil
}

// Clear removes all values from all tiers
func (mtc *MultiTierCache) Clear(ctx context.Context) error {
	mtc.mu.RLock()
	defer mtc.mu.RUnlock()

	var lastErr error
	for _, tier := range mtc.tiers {
		if err := tier.cache.Clear(ctx); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// Keys returns all keys from L1 cache
func (mtc *MultiTierCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	mtc.mu.RLock()
	defer mtc.mu.RUnlock()

	if len(mtc.tiers) == 0 {
		return []string{}, nil
	}

	return mtc.tiers[0].cache.Keys(ctx, pattern)
}

// TTL returns the TTL from the first tier that has the key
func (mtc *MultiTierCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	mtc.mu.RLock()
	defer mtc.mu.RUnlock()

	for _, tier := range mtc.tiers {
		ttl, err := tier.cache.TTL(ctx, key)
		if err == nil {
			return ttl, nil
		}
	}

	return 0, ErrKeyNotFound
}

// Expire sets TTL on all tiers
func (mtc *MultiTierCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	mtc.mu.RLock()
	defer mtc.mu.RUnlock()

	var lastErr error
	for _, tier := range mtc.tiers {
		if err := tier.cache.Expire(ctx, key, ttl); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// Increment increments on L1 and propagates
func (mtc *MultiTierCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	mtc.mu.RLock()
	defer mtc.mu.RUnlock()

	if len(mtc.tiers) == 0 {
		return 0, &CacheError{Op: "increment", Err: ErrClosed}
	}

	// Increment on L1
	val, err := mtc.tiers[0].cache.Increment(ctx, key, delta)
	if err != nil {
		return 0, err
	}

	// Propagate to other tiers
	if len(mtc.tiers) > 1 {
		go func() {
			for i := 1; i < len(mtc.tiers); i++ {
				mtc.tiers[i].cache.Set(context.Background(), key, val, mtc.config.DefaultTTL)
			}
		}()
	}

	return val, nil
}

// Decrement decrements on L1 and propagates
func (mtc *MultiTierCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return mtc.Increment(ctx, key, -delta)
}

// GetMulti retrieves multiple values
func (mtc *MultiTierCache) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	remaining := keys

	mtc.mu.RLock()
	defer mtc.mu.RUnlock()

	// Try each tier
	for i, tier := range mtc.tiers {
		if len(remaining) == 0 {
			break
		}

		values, err := tier.cache.GetMulti(ctx, remaining)
		if err != nil {
			continue
		}

		// Collect found values
		for key, value := range values {
			result[key] = value
		}

		// Promote to higher tiers if needed
		if mtc.promoteL1 && i > 0 && len(values) > 0 {
			go mtc.promoteMultiToHigherTiers(values, i)
		}

		// Update remaining keys
		newRemaining := []string{}
		for _, key := range remaining {
			if _, found := values[key]; !found {
				newRemaining = append(newRemaining, key)
			}
		}
		remaining = newRemaining
	}

	return result, nil
}

// SetMulti stores multiple values
func (mtc *MultiTierCache) SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	mtc.mu.RLock()
	defer mtc.mu.RUnlock()

	if ttl == 0 {
		ttl = mtc.config.DefaultTTL
	}

	if mtc.writeThru {
		// Write to all tiers
		for _, tier := range mtc.tiers {
			if err := tier.cache.SetMulti(ctx, items, ttl); err != nil {
				return err
			}
		}
		return nil
	}

	// Write to L1 only
	if len(mtc.tiers) > 0 {
		if err := mtc.tiers[0].cache.SetMulti(ctx, items, ttl); err != nil {
			return err
		}

		if mtc.writeBack && len(mtc.tiers) > 1 {
			go mtc.setMultiToLowerTiers(items, ttl)
		}
	}

	return nil
}

// DeleteMulti removes multiple values
func (mtc *MultiTierCache) DeleteMulti(ctx context.Context, keys []string) error {
	mtc.mu.RLock()
	defer mtc.mu.RUnlock()

	var lastErr error
	for _, tier := range mtc.tiers {
		if err := tier.cache.DeleteMulti(ctx, keys); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// Stats returns combined statistics from all tiers
func (mtc *MultiTierCache) Stats(ctx context.Context) (*Stats, error) {
	mtc.mu.RLock()
	defer mtc.mu.RUnlock()

	combined := &Stats{
		Hits:   mtc.stats.Hits,
		Misses: mtc.stats.Misses,
	}

	for _, tier := range mtc.tiers {
		if sp, ok := tier.cache.(StatsProvider); ok {
			stats, err := sp.Stats(ctx)
			if err == nil {
				combined.Keys += stats.Keys
				combined.Memory += stats.Memory
				combined.Evictions += stats.Evictions
			}
		}
	}

	return combined, nil
}

// Close closes all cache tiers
func (mtc *MultiTierCache) Close() error {
	mtc.mu.Lock()
	defer mtc.mu.Unlock()

	var lastErr error
	for _, tier := range mtc.tiers {
		if err := tier.cache.Close(); err != nil {
			lastErr = err
		}
	}

	mtc.tiers = nil
	return lastErr
}

// Helper methods

func (mtc *MultiTierCache) sortTiers() {
	// Simple bubble sort by tier level
	for i := 0; i < len(mtc.tiers); i++ {
		for j := i + 1; j < len(mtc.tiers); j++ {
			if mtc.tiers[i].level > mtc.tiers[j].level {
				mtc.tiers[i], mtc.tiers[j] = mtc.tiers[j], mtc.tiers[i]
			}
		}
	}
}

func (mtc *MultiTierCache) promoteToHigherTiers(key string, value interface{}, fromTier int) {
	ctx := context.Background()
	for i := 0; i < fromTier; i++ {
		mtc.tiers[i].cache.Set(ctx, key, value, mtc.config.DefaultTTL)
	}
}

func (mtc *MultiTierCache) promoteMultiToHigherTiers(values map[string]interface{}, fromTier int) {
	ctx := context.Background()
	for i := 0; i < fromTier; i++ {
		mtc.tiers[i].cache.SetMulti(ctx, values, mtc.config.DefaultTTL)
	}
}

func (mtc *MultiTierCache) writeToLowerTiers(key string, value interface{}, ttl time.Duration) {
	ctx := context.Background()
	for i := 1; i < len(mtc.tiers); i++ {
		mtc.tiers[i].cache.Set(ctx, key, value, ttl)
	}
}

func (mtc *MultiTierCache) setMultiToLowerTiers(items map[string]interface{}, ttl time.Duration) {
	ctx := context.Background()
	for i := 1; i < len(mtc.tiers); i++ {
		mtc.tiers[i].cache.SetMulti(ctx, items, ttl)
	}
}

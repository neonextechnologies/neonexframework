# Cache Management

Master high-performance caching with NeonEx Framework's powerful Redis-based cache system. Learn cache patterns, TTL management, and optimization strategies for production applications.

## Table of Contents

- [Introduction](#introduction)
- [Quick Start](#quick-start)
- [Cache Configuration](#cache-configuration)
- [Basic Operations](#basic-operations)
- [Advanced Patterns](#advanced-patterns)
- [Multi-Tier Caching](#multi-tier-caching)
- [Cache Invalidation](#cache-invalidation)
- [Performance Optimization](#performance-optimization)
- [Integration Examples](#integration-examples)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Introduction

NeonEx provides a flexible caching system with multiple backend support (Redis, Memory, Multi-tier). The cache system offers:

- **Multiple Backends**: Redis, in-memory, and multi-tier caching
- **TTL Management**: Automatic expiration and refresh
- **Batch Operations**: Get/Set/Delete multiple keys efficiently
- **Atomic Operations**: Counters and atomic increments
- **Cache Patterns**: Cache-aside, write-through, write-behind
- **Statistics**: Monitor cache hits, misses, and performance
- **Type Safety**: Generic support for type-safe caching

## Quick Start

### Basic Setup

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "neonex/core/pkg/cache"
)

func main() {
    // Create Redis cache
    config := cache.RedisCacheConfig{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    }
    
    cacheManager, err := cache.NewRedisCache(config)
    if err != nil {
        panic(err)
    }
    defer cacheManager.Close()
    
    ctx := context.Background()
    
    // Set a value
    err = cacheManager.Set(ctx, "user:1", map[string]interface{}{
        "id":    1,
        "name":  "John Doe",
        "email": "john@example.com",
    }, 5*time.Minute)
    
    // Get the value
    value, err := cacheManager.Get(ctx, "user:1")
    if err != nil {
        fmt.Println("Cache miss:", err)
    } else {
        fmt.Println("Cache hit:", value)
    }
}
```

### In-Memory Cache

```go
import "neonex/core/pkg/cache"

// Create in-memory cache (for single-instance apps)
memCache := cache.NewMemoryCache(cache.DefaultConfig())

ctx := context.Background()

// Store data
memCache.Set(ctx, "session:abc123", sessionData, 30*time.Minute)

// Retrieve data
data, err := memCache.Get(ctx, "session:abc123")
```

## Cache Configuration

### Redis Configuration

```go
config := cache.RedisCacheConfig{
    Config: cache.Config{
        DefaultTTL: 10 * time.Minute,
        MaxRetries: 3,
        RetryDelay: 100 * time.Millisecond,
        Timeout:    5 * time.Second,
    },
    
    // Redis connection
    Addr:         "localhost:6379",
    Password:     "your-redis-password",
    DB:           0,
    
    // Connection pool
    PoolSize:     10,
    MinIdleConns: 2,
    
    // Timeouts
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
}

cacheManager, err := cache.NewRedisCache(config)
```

### Environment-Based Configuration

```go
import (
    "os"
    "strconv"
    "time"
)

func NewCacheFromEnv() (cache.Cache, error) {
    poolSize, _ := strconv.Atoi(os.Getenv("REDIS_POOL_SIZE"))
    if poolSize == 0 {
        poolSize = 10
    }
    
    config := cache.RedisCacheConfig{
        Addr:         os.Getenv("REDIS_ADDR"),
        Password:     os.Getenv("REDIS_PASSWORD"),
        DB:           0,
        PoolSize:     poolSize,
        MinIdleConns: 2,
    }
    
    return cache.NewRedisCache(config)
}
```

### Configuration File

```yaml
# config/cache.yaml
cache:
  driver: redis
  redis:
    addr: localhost:6379
    password: ""
    db: 0
    pool_size: 10
    min_idle_conns: 2
  default_ttl: 600  # 10 minutes in seconds
  timeout: 5s
```

## Basic Operations

### Set and Get

```go
ctx := context.Background()

// Set with default TTL
err := cacheManager.Set(ctx, "key", "value", 0)

// Set with specific TTL
err = cacheManager.Set(ctx, "user:1:profile", userData, 15*time.Minute)

// Get value
value, err := cacheManager.Get(ctx, "user:1:profile")
if err == cache.ErrKeyNotFound {
    // Handle cache miss
    value = fetchFromDatabase()
    cacheManager.Set(ctx, "user:1:profile", value, 15*time.Minute)
}
```

### Check Existence

```go
exists, err := cacheManager.Exists(ctx, "user:1:profile")
if exists {
    fmt.Println("Key exists in cache")
}
```

### Delete

```go
// Delete single key
err := cacheManager.Delete(ctx, "user:1:profile")

// Delete multiple keys
err = cacheManager.DeleteMulti(ctx, []string{
    "user:1:profile",
    "user:1:preferences",
    "user:1:settings",
})
```

### TTL Management

```go
// Get remaining TTL
ttl, err := cacheManager.TTL(ctx, "user:1:session")
fmt.Printf("Key expires in: %v\n", ttl)

// Extend TTL
err = cacheManager.Expire(ctx, "user:1:session", 1*time.Hour)

// Remove expiration (make key persistent)
err = cacheManager.Expire(ctx, "user:1:session", -1)
```

### Clear Cache

```go
// Clear all keys (use with caution!)
err := cacheManager.Clear(ctx)
```

## Advanced Patterns

### Cache-Aside Pattern

The most common caching pattern - application checks cache first, then database:

```go
type UserRepository struct {
    db    *gorm.DB
    cache cache.Cache
}

func (r *UserRepository) GetUser(ctx context.Context, userID int) (*User, error) {
    cacheKey := fmt.Sprintf("user:%d", userID)
    
    // Try cache first
    cachedData, err := r.cache.Get(ctx, cacheKey)
    if err == nil {
        var user User
        // Convert cached data to user struct
        json.Unmarshal(cachedData.([]byte), &user)
        return &user, nil
    }
    
    // Cache miss - fetch from database
    var user User
    if err := r.db.WithContext(ctx).First(&user, userID).Error; err != nil {
        return nil, err
    }
    
    // Store in cache for next time
    userData, _ := json.Marshal(user)
    r.cache.Set(ctx, cacheKey, userData, 15*time.Minute)
    
    return &user, nil
}
```

### Write-Through Pattern

Write to cache and database simultaneously:

```go
func (r *UserRepository) UpdateUser(ctx context.Context, user *User) error {
    // Update database
    if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
        return err
    }
    
    // Update cache immediately
    cacheKey := fmt.Sprintf("user:%d", user.ID)
    userData, _ := json.Marshal(user)
    r.cache.Set(ctx, cacheKey, userData, 15*time.Minute)
    
    return nil
}
```

### Write-Behind Pattern

Write to cache immediately, sync to database asynchronously:

```go
type WriteBuffer struct {
    cache   cache.Cache
    db      *gorm.DB
    queue   chan *WriteOperation
}

type WriteOperation struct {
    Key   string
    Value interface{}
}

func (wb *WriteBuffer) UpdateAsync(ctx context.Context, key string, value interface{}) error {
    // Write to cache immediately
    if err := wb.cache.Set(ctx, key, value, 0); err != nil {
        return err
    }
    
    // Queue for async database write
    select {
    case wb.queue <- &WriteOperation{Key: key, Value: value}:
        return nil
    default:
        return fmt.Errorf("write queue full")
    }
}

func (wb *WriteBuffer) ProcessWrites(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case op := <-wb.queue:
            // Write to database
            wb.writeToDatabase(op)
        }
    }
}
```

### Cache Warming

Pre-populate cache with frequently accessed data:

```go
func (r *UserRepository) WarmCache(ctx context.Context) error {
    // Get top 100 active users
    var users []User
    err := r.db.WithContext(ctx).
        Order("last_login_at DESC").
        Limit(100).
        Find(&users).Error
    
    if err != nil {
        return err
    }
    
    // Populate cache
    items := make(map[string]interface{})
    for _, user := range users {
        cacheKey := fmt.Sprintf("user:%d", user.ID)
        userData, _ := json.Marshal(user)
        items[cacheKey] = userData
    }
    
    return r.cache.SetMulti(ctx, items, 1*time.Hour)
}
```

## Multi-Tier Caching

Combine memory and Redis caching for optimal performance:

```go
import "neonex/core/pkg/cache"

type MultiTierCache struct {
    l1 cache.Cache // Memory cache (fast, small)
    l2 cache.Cache // Redis cache (slower, large)
}

func NewMultiTierCache(memCache, redisCache cache.Cache) *MultiTierCache {
    return &MultiTierCache{
        l1: memCache,
        l2: redisCache,
    }
}

func (m *MultiTierCache) Get(ctx context.Context, key string) (interface{}, error) {
    // Check L1 (memory) first
    value, err := m.l1.Get(ctx, key)
    if err == nil {
        return value, nil
    }
    
    // Check L2 (Redis)
    value, err = m.l2.Get(ctx, key)
    if err == nil {
        // Populate L1 for next access
        m.l1.Set(ctx, key, value, 1*time.Minute)
        return value, nil
    }
    
    return nil, cache.ErrKeyNotFound
}

func (m *MultiTierCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    // Write to both tiers
    m.l1.Set(ctx, key, value, ttl)
    return m.l2.Set(ctx, key, value, ttl)
}

func (m *MultiTierCache) Delete(ctx context.Context, key string) error {
    // Delete from both tiers
    m.l1.Delete(ctx, key)
    return m.l2.Delete(ctx, key)
}
```

## Cache Invalidation

### Pattern Matching

```go
// Delete all user-related caches
pattern := "user:*"
keys, err := cacheManager.Keys(ctx, pattern)
if err != nil {
    return err
}

// Delete all matching keys
err = cacheManager.DeleteMulti(ctx, keys)
```

### Tag-Based Invalidation

```go
type TaggedCache struct {
    cache cache.Cache
}

func (tc *TaggedCache) SetWithTags(ctx context.Context, key string, value interface{}, tags []string, ttl time.Duration) error {
    // Store the value
    if err := tc.cache.Set(ctx, key, value, ttl); err != nil {
        return err
    }
    
    // Store key in each tag set
    for _, tag := range tags {
        tagKey := fmt.Sprintf("tag:%s:keys", tag)
        tc.cache.Set(ctx, tagKey, key, ttl)
    }
    
    return nil
}

func (tc *TaggedCache) InvalidateTag(ctx context.Context, tag string) error {
    // Get all keys with this tag
    tagKey := fmt.Sprintf("tag:%s:keys", tag)
    keys, err := tc.cache.Keys(ctx, tagKey)
    if err != nil {
        return err
    }
    
    // Delete all keys
    return tc.cache.DeleteMulti(ctx, keys)
}
```

### Event-Based Invalidation

```go
import "neonex/core/pkg/events"

func SetupCacheInvalidation(cache cache.Cache, dispatcher *events.EventDispatcher) {
    // Invalidate user cache on update
    dispatcher.Register(events.EventUserUpdated, func(ctx context.Context, event events.Event) error {
        userID := event.Data.(map[string]interface{})["user_id"]
        cacheKey := fmt.Sprintf("user:%v", userID)
        return cache.Delete(ctx, cacheKey)
    })
    
    // Invalidate product cache on update
    dispatcher.Register("product.updated", func(ctx context.Context, event events.Event) error {
        productID := event.Data.(map[string]interface{})["product_id"]
        cacheKey := fmt.Sprintf("product:%v", productID)
        return cache.Delete(ctx, cacheKey)
    })
}
```

## Performance Optimization

### Batch Operations

```go
// Get multiple keys efficiently
keys := []string{
    "user:1:profile",
    "user:1:settings",
    "user:1:preferences",
}

results, err := cacheManager.GetMulti(ctx, keys)
for key, value := range results {
    fmt.Printf("%s: %v\n", key, value)
}

// Set multiple keys efficiently
items := map[string]interface{}{
    "product:1": productData1,
    "product:2": productData2,
    "product:3": productData3,
}

err = cacheManager.SetMulti(ctx, items, 10*time.Minute)
```

### Atomic Counters

```go
// Increment view count
viewCount, err := cacheManager.Increment(ctx, "post:123:views", 1)
fmt.Printf("Post viewed %d times\n", viewCount)

// Decrement inventory
stock, err := cacheManager.Decrement(ctx, "product:456:stock", 1)
if stock < 0 {
    // Out of stock
}

// Rate limiting
attempts, err := cacheManager.Increment(ctx, "login:attempts:user123", 1)
if attempts > 5 {
    return errors.New("too many login attempts")
}
```

### Pipeline Operations

```go
// Use Redis pipeline for multiple operations
redisCache := cacheManager.(*cache.RedisCache)
pipe := redisCache.Pipeline()

// Queue multiple operations
for i := 1; i <= 100; i++ {
    key := fmt.Sprintf("key:%d", i)
    pipe.Get(ctx, key)
}

// Execute all at once
results, err := pipe.Exec(ctx)
```

### Compression

```go
import (
    "bytes"
    "compress/gzip"
    "io"
)

type CompressedCache struct {
    cache cache.Cache
}

func (cc *CompressedCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    // Serialize value
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    
    // Compress if large
    if len(data) > 1024 {
        var buf bytes.Buffer
        gz := gzip.NewWriter(&buf)
        gz.Write(data)
        gz.Close()
        data = buf.Bytes()
    }
    
    return cc.cache.Set(ctx, key, data, ttl)
}
```

## Integration Examples

### Complete Repository Pattern

```go
type ProductRepository struct {
    db    *gorm.DB
    cache cache.Cache
}

func NewProductRepository(db *gorm.DB, cache cache.Cache) *ProductRepository {
    return &ProductRepository{
        db:    db,
        cache: cache,
    }
}

func (r *ProductRepository) GetProduct(ctx context.Context, id int) (*Product, error) {
    cacheKey := fmt.Sprintf("product:%d", id)
    
    // Try cache
    if cached, err := r.cache.Get(ctx, cacheKey); err == nil {
        var product Product
        json.Unmarshal(cached.([]byte), &product)
        return &product, nil
    }
    
    // Fetch from database
    var product Product
    if err := r.db.WithContext(ctx).First(&product, id).Error; err != nil {
        return nil, err
    }
    
    // Cache result
    data, _ := json.Marshal(product)
    r.cache.Set(ctx, cacheKey, data, 30*time.Minute)
    
    return &product, nil
}

func (r *ProductRepository) UpdateProduct(ctx context.Context, product *Product) error {
    // Update database
    if err := r.db.WithContext(ctx).Save(product).Error; err != nil {
        return err
    }
    
    // Invalidate cache
    cacheKey := fmt.Sprintf("product:%d", product.ID)
    r.cache.Delete(ctx, cacheKey)
    
    // Also invalidate list caches
    r.cache.Delete(ctx, "products:list")
    
    return nil
}

func (r *ProductRepository) GetProducts(ctx context.Context, limit, offset int) ([]Product, error) {
    cacheKey := fmt.Sprintf("products:list:%d:%d", limit, offset)
    
    // Try cache
    if cached, err := r.cache.Get(ctx, cacheKey); err == nil {
        var products []Product
        json.Unmarshal(cached.([]byte), &products)
        return products, nil
    }
    
    // Fetch from database
    var products []Product
    err := r.db.WithContext(ctx).
        Limit(limit).
        Offset(offset).
        Find(&products).Error
    
    if err != nil {
        return nil, err
    }
    
    // Cache result
    data, _ := json.Marshal(products)
    r.cache.Set(ctx, cacheKey, data, 5*time.Minute)
    
    return products, nil
}
```

### Session Management

```go
type SessionManager struct {
    cache cache.Cache
}

func (sm *SessionManager) CreateSession(ctx context.Context, userID int) (string, error) {
    // Generate session ID
    sessionID := generateSessionID()
    
    sessionData := map[string]interface{}{
        "user_id":    userID,
        "created_at": time.Now(),
        "ip_address": getIPFromContext(ctx),
    }
    
    // Store session
    key := fmt.Sprintf("session:%s", sessionID)
    err := sm.cache.Set(ctx, key, sessionData, 24*time.Hour)
    
    return sessionID, err
}

func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (map[string]interface{}, error) {
    key := fmt.Sprintf("session:%s", sessionID)
    return sm.cache.Get(ctx, key)
}

func (sm *SessionManager) RefreshSession(ctx context.Context, sessionID string) error {
    key := fmt.Sprintf("session:%s", sessionID)
    return sm.cache.Expire(ctx, key, 24*time.Hour)
}

func (sm *SessionManager) DestroySession(ctx context.Context, sessionID string) error {
    key := fmt.Sprintf("session:%s", sessionID)
    return sm.cache.Delete(ctx, key)
}
```

### API Rate Limiting

```go
type RateLimiter struct {
    cache cache.Cache
}

func (rl *RateLimiter) AllowRequest(ctx context.Context, clientID string, limit int, window time.Duration) (bool, error) {
    key := fmt.Sprintf("ratelimit:%s", clientID)
    
    // Get current count
    count, err := rl.cache.Get(ctx, key)
    if err == cache.ErrKeyNotFound {
        // First request in window
        rl.cache.Set(ctx, key, 1, window)
        return true, nil
    }
    
    currentCount := int(count.(float64))
    if currentCount >= limit {
        return false, nil
    }
    
    // Increment counter
    rl.cache.Increment(ctx, key, 1)
    return true, nil
}
```

## Best Practices

### 1. Cache Key Naming

```go
// Use consistent, hierarchical naming
const (
    KeyUser          = "user:%d"                    // user:123
    KeyUserProfile   = "user:%d:profile"           // user:123:profile
    KeyUserSettings  = "user:%d:settings"          // user:123:settings
    KeyProductList   = "products:list:%s:%d:%d"    // products:list:category:limit:offset
    KeySession       = "session:%s"                 // session:abc123
)

// Helper function
func UserCacheKey(userID int) string {
    return fmt.Sprintf(KeyUser, userID)
}
```

### 2. Appropriate TTL Values

```go
const (
    TTLVeryShort = 1 * time.Minute   // Frequently changing data
    TTLShort     = 5 * time.Minute   // Semi-static data
    TTLMedium    = 15 * time.Minute  // User profiles, settings
    TTLLong      = 1 * time.Hour     // Static content
    TTLVeryLong  = 24 * time.Hour    // Rarely changing data
)

// Apply based on data volatility
cache.Set(ctx, "realtime:stock:price", price, TTLVeryShort)
cache.Set(ctx, "user:profile", profile, TTLMedium)
cache.Set(ctx, "static:categories", categories, TTLVeryLong)
```

### 3. Error Handling

```go
func GetWithFallback(ctx context.Context, cache cache.Cache, key string, fetcher func() (interface{}, error)) (interface{}, error) {
    // Try cache
    value, err := cache.Get(ctx, key)
    if err == nil {
        return value, nil
    }
    
    // Log cache errors (but don't fail)
    if err != cache.ErrKeyNotFound {
        log.Warn("Cache error", logger.Fields{
            "key":   key,
            "error": err,
        })
    }
    
    // Fetch from source
    value, err = fetcher()
    if err != nil {
        return nil, err
    }
    
    // Try to cache (but don't fail if caching fails)
    if err := cache.Set(ctx, key, value, 10*time.Minute); err != nil {
        log.Warn("Failed to cache value", logger.Fields{
            "key":   key,
            "error": err,
        })
    }
    
    return value, nil
}
```

### 4. Cache Statistics

```go
// Monitor cache performance
if statsProvider, ok := cacheManager.(cache.StatsProvider); ok {
    stats, err := statsProvider.Stats(ctx)
    if err == nil {
        hitRate := float64(stats.Hits) / float64(stats.Hits+stats.Misses) * 100
        
        log.Info("Cache statistics", logger.Fields{
            "hits":      stats.Hits,
            "misses":    stats.Misses,
            "hit_rate":  fmt.Sprintf("%.2f%%", hitRate),
            "keys":      stats.Keys,
            "memory_mb": stats.Memory / 1024 / 1024,
        })
    }
}
```

### 5. Graceful Degradation

```go
type ResilientCache struct {
    cache   cache.Cache
    enabled bool
}

func (rc *ResilientCache) Get(ctx context.Context, key string) (interface{}, error) {
    if !rc.enabled {
        return nil, cache.ErrKeyNotFound
    }
    
    return rc.cache.Get(ctx, key)
}

func (rc *ResilientCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    if !rc.enabled {
        return nil // Silently skip
    }
    
    return rc.cache.Set(ctx, key, value, ttl)
}

// Disable cache if Redis is down
func (rc *ResilientCache) HealthCheck(ctx context.Context) {
    if err := rc.cache.Exists(ctx, "healthcheck"); err != nil {
        rc.enabled = false
        log.Warn("Cache disabled due to health check failure")
    } else {
        rc.enabled = true
    }
}
```

## Troubleshooting

### Cache Not Working

```go
// Test cache connectivity
func TestCache(cache cache.Cache) error {
    ctx := context.Background()
    
    // Try to set a test key
    if err := cache.Set(ctx, "test:key", "test value", 1*time.Minute); err != nil {
        return fmt.Errorf("failed to set: %w", err)
    }
    
    // Try to get it back
    value, err := cache.Get(ctx, "test:key")
    if err != nil {
        return fmt.Errorf("failed to get: %w", err)
    }
    
    if value != "test value" {
        return fmt.Errorf("value mismatch")
    }
    
    // Clean up
    cache.Delete(ctx, "test:key")
    
    return nil
}
```

### Memory Issues

```go
// Monitor memory usage
stats, _ := cacheManager.(*cache.RedisCache).Stats(ctx)
fmt.Printf("Memory: %d MB\n", stats.Memory/1024/1024)

// If memory is high, implement eviction
if stats.Memory > 1024*1024*1024 { // 1GB
    // Delete old keys
    keys, _ := cacheManager.Keys(ctx, "*")
    for _, key := range keys {
        ttl, _ := cacheManager.TTL(ctx, key)
        if ttl < 0 {
            // Key has no expiration
            cacheManager.Delete(ctx, key)
        }
    }
}
```

### Connection Issues

```go
// Implement retry logic
func GetWithRetry(cache cache.Cache, ctx context.Context, key string, maxRetries int) (interface{}, error) {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        value, err := cache.Get(ctx, key)
        if err == nil {
            return value, nil
        }
        
        lastErr = err
        time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
    }
    
    return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
```

### Debug Logging

```go
type DebugCache struct {
    cache  cache.Cache
    logger logger.Logger
}

func (dc *DebugCache) Get(ctx context.Context, key string) (interface{}, error) {
    start := time.Now()
    value, err := dc.cache.Get(ctx, key)
    duration := time.Since(start)
    
    dc.logger.Debug("Cache GET", logger.Fields{
        "key":      key,
        "duration": duration,
        "hit":      err == nil,
    })
    
    return value, err
}

func (dc *DebugCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    start := time.Now()
    err := dc.cache.Set(ctx, key, value, ttl)
    duration := time.Since(start)
    
    dc.logger.Debug("Cache SET", logger.Fields{
        "key":      key,
        "ttl":      ttl,
        "duration": duration,
        "success":  err == nil,
    })
    
    return err
}
```

---

**Next Steps:**
- Learn about [Events](events.md) for event-driven cache invalidation
- Explore [Metrics](metrics.md) for cache monitoring
- See [Performance](../deployment/performance.md) for optimization strategies

**Related Topics:**
- [Database Repository](../database/repository.md)
- [Session Management](../security/authentication.md)
- [API Rate Limiting](../core-concepts/middleware.md)

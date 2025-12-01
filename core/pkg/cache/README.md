# Cache Package

Advanced multi-tier caching system with memory, Redis, and distributed cache support for NeonexCore.

## Features

- ✅ **Multi-Tier Caching** - L1 (Memory) → L2 (Redis) → L3 (Remote)
- ✅ **Memory Cache** - LRU eviction with auto-cleanup
- ✅ **Redis Cache** - Distributed caching with persistence
- ✅ **Write Strategies** - Write-through, Write-back, Write-around
- ✅ **Cache Promotion** - Auto-promote hot data to faster tiers
- ✅ **Atomic Operations** - Increment/Decrement counters
- ✅ **Batch Operations** - GetMulti, SetMulti, DeleteMulti
- ✅ **TTL Management** - Per-key expiration
- ✅ **Statistics** - Hits, misses, evictions tracking
- ✅ **Pattern Matching** - Wildcard key search
- ✅ **Thread-Safe** - Concurrent access support

## Architecture

```
pkg/cache/
├── cache.go      - Cache interface and base types
├── memory.go     - In-memory LRU cache
├── redis.go      - Redis cache implementation
└── multitier.go  - Multi-tier cache orchestration
```

## Quick Start

### 1. Memory Cache (L1)

```go
import "neonexcore/pkg/cache"

// Create memory cache
config := cache.DefaultMemoryCacheConfig()
config.MaxSize = 10000
config.CleanupInterval = 1 * time.Minute

memCache := cache.NewMemoryCache(config)
defer memCache.Close()

// Set value
memCache.Set(ctx, "user:123", user, 5*time.Minute)

// Get value
value, err := memCache.Get(ctx, "user:123")
```

### 2. Redis Cache (L2)

```go
// Create Redis cache
config := cache.DefaultRedisCacheConfig()
config.Addr = "localhost:6379"
config.Password = ""
config.DB = 0

redisCache, err := cache.NewRedisCache(config)
if err != nil {
    log.Fatal(err)
}
defer redisCache.Close()

// Use like memory cache
redisCache.Set(ctx, "session:abc", session, 30*time.Minute)
value, err := redisCache.Get(ctx, "session:abc")
```

### 3. Multi-Tier Cache (L1 + L2)

```go
// Create multi-tier cache
config := cache.DefaultMultiTierConfig()
config.PromoteL1 = true // Promote hits to L1
config.WriteThru = true // Write to all tiers

multiCache := cache.NewMultiTierCache(config)

// Add tiers
multiCache.AddTier(memCache, cache.TierL1)    // Fast
multiCache.AddTier(redisCache, cache.TierL2)  // Distributed

// Use seamlessly
multiCache.Set(ctx, "user:123", user, 5*time.Minute)

// Auto-searches L1 → L2, promotes to L1 on hit
value, err := multiCache.Get(ctx, "user:123")
```

## Cache Interface

All cache implementations satisfy the `Cache` interface:

```go
type Cache interface {
    Get(ctx, key) (interface{}, error)
    Set(ctx, key, value, ttl) error
    Delete(ctx, key) error
    Exists(ctx, key) (bool, error)
    Clear(ctx) error
    Keys(ctx, pattern) ([]string, error)
    TTL(ctx, key) (time.Duration, error)
    Expire(ctx, key, ttl) error
    Increment(ctx, key, delta) (int64, error)
    Decrement(ctx, key, delta) (int64, error)
    GetMulti(ctx, keys) (map[string]interface{}, error)
    SetMulti(ctx, items, ttl) error
    DeleteMulti(ctx, keys) error
    Close() error
}
```

## Usage Patterns

### Basic Operations

```go
// Set with TTL
cache.Set(ctx, "key", "value", 10*time.Minute)

// Get
value, err := cache.Get(ctx, "key")
if err == cache.ErrKeyNotFound {
    // Cache miss
}

// Delete
cache.Delete(ctx, "key")

// Check existence
exists, _ := cache.Exists(ctx, "key")

// Clear all
cache.Clear(ctx)
```

### TTL Management

```go
// Get remaining TTL
ttl, _ := cache.TTL(ctx, "key")
fmt.Printf("Expires in: %v\n", ttl)

// Update TTL
cache.Expire(ctx, "key", 5*time.Minute)

// Set without expiration (0 = default TTL)
cache.Set(ctx, "key", "value", 0)

// Infinite TTL (negative duration)
cache.Set(ctx, "key", "value", -1)
```

### Atomic Counters

```go
// Initialize counter
cache.Set(ctx, "views:post:123", 0, 24*time.Hour)

// Increment
newValue, _ := cache.Increment(ctx, "views:post:123", 1)
fmt.Printf("Views: %d\n", newValue)

// Decrement
cache.Decrement(ctx, "stock:item:456", 1)

// Increment by custom amount
cache.Increment(ctx, "score:user:789", 10)
```

### Batch Operations

```go
// Get multiple keys
keys := []string{"user:1", "user:2", "user:3"}
values, _ := cache.GetMulti(ctx, keys)

for key, value := range values {
    fmt.Printf("%s = %v\n", key, value)
}

// Set multiple keys
items := map[string]interface{}{
    "config:timeout": 30,
    "config:retries": 3,
    "config:enabled": true,
}
cache.SetMulti(ctx, items, 1*time.Hour)

// Delete multiple keys
cache.DeleteMulti(ctx, []string{"temp:1", "temp:2", "temp:3"})
```

### Pattern Matching

```go
// Get all user keys
userKeys, _ := cache.Keys(ctx, "user:*")

// Get specific pattern
sessionKeys, _ := cache.Keys(ctx, "session:*")

// Get all keys
allKeys, _ := cache.Keys(ctx, "*")

// Delete by pattern
keys, _ := cache.Keys(ctx, "temp:*")
cache.DeleteMulti(ctx, keys)
```

### Cache Statistics

```go
if sp, ok := cache.(cache.StatsProvider); ok {
    stats, _ := sp.Stats(ctx)
    
    fmt.Printf("Hits: %d\n", stats.Hits)
    fmt.Printf("Misses: %d\n", stats.Misses)
    fmt.Printf("Hit Rate: %.2f%%\n", 
        float64(stats.Hits)/(float64(stats.Hits+stats.Misses))*100)
    fmt.Printf("Keys: %d\n", stats.Keys)
    fmt.Printf("Memory: %d bytes\n", stats.Memory)
    fmt.Printf("Evictions: %d\n", stats.Evictions)
}
```

## Multi-Tier Strategies

### Write-Through (Default)

Writes to all tiers synchronously:

```go
config := cache.DefaultMultiTierConfig()
config.WriteThru = true
config.WriteBack = false

// Set writes to L1 and L2 immediately
multiCache.Set(ctx, "key", "value", ttl)
```

**Pros:** Data consistency across tiers  
**Cons:** Slower writes

### Write-Back

Writes to L1 immediately, L2 asynchronously:

```go
config := cache.DefaultMultiTierConfig()
config.WriteThru = false
config.WriteBack = true

// Set writes to L1 immediately, L2 in background
multiCache.Set(ctx, "key", "value", ttl)
```

**Pros:** Faster writes  
**Cons:** Potential data loss if L1 fails

### Cache Promotion

Automatically promotes frequently accessed data to faster tiers:

```go
config := cache.DefaultMultiTierConfig()
config.PromoteL1 = true

// First access: Cache miss, loads from L2
value, _ := multiCache.Get(ctx, "key") // L2 hit

// Second access: Cache hit from L1 (promoted)
value, _ := multiCache.Get(ctx, "key") // L1 hit (faster!)
```

## Configuration

### Memory Cache

```go
config := cache.MemoryCacheConfig{
    MaxSize:         10000,              // Max items
    CleanupInterval: 1 * time.Minute,    // Cleanup frequency
    DefaultTTL:      5 * time.Minute,    // Default expiration
}
```

### Redis Cache

```go
config := cache.RedisCacheConfig{
    Addr:         "localhost:6379",
    Password:     "",
    DB:           0,
    PoolSize:     10,                   // Connection pool
    MinIdleConns: 2,
    MaxRetries:   3,
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
    DefaultTTL:   5 * time.Minute,
}
```

### Multi-Tier Cache

```go
config := cache.MultiTierConfig{
    PromoteL1:  true,                   // Promote to L1
    WriteThru:  true,                   // Write-through
    WriteBack:  false,                  // Write-back
    DefaultTTL: 5 * time.Minute,
}
```

## Integration Examples

### HTTP Response Caching

```go
func GetUser(c *fiber.Ctx) error {
    userID := c.Params("id")
    cacheKey := "user:" + userID
    
    // Try cache first
    if cached, err := cache.Get(c.Context(), cacheKey); err == nil {
        return c.JSON(cached)
    }
    
    // Cache miss - fetch from database
    user, err := db.GetUser(userID)
    if err != nil {
        return err
    }
    
    // Store in cache
    cache.Set(c.Context(), cacheKey, user, 10*time.Minute)
    
    return c.JSON(user)
}
```

### Session Management

```go
func SaveSession(c *fiber.Ctx) error {
    sessionID := c.Cookies("session_id")
    sessionData := map[string]interface{}{
        "user_id": 123,
        "role":    "admin",
        "expires": time.Now().Add(24 * time.Hour),
    }
    
    // Store session in cache
    cache.Set(c.Context(), "session:"+sessionID, sessionData, 24*time.Hour)
    
    return c.SendStatus(200)
}

func GetSession(c *fiber.Ctx) error {
    sessionID := c.Cookies("session_id")
    
    // Get session from cache
    session, err := cache.Get(c.Context(), "session:"+sessionID)
    if err != nil {
        return c.Status(401).SendString("Unauthorized")
    }
    
    return c.JSON(session)
}
```

### Rate Limiting

```go
func RateLimitMiddleware(limit int64, window time.Duration) fiber.Handler {
    return func(c *fiber.Ctx) error {
        ip := c.IP()
        key := "ratelimit:" + ip
        
        // Increment request count
        count, _ := cache.Increment(c.Context(), key, 1)
        
        // Set expiry on first request
        if count == 1 {
            cache.Expire(c.Context(), key, window)
        }
        
        // Check limit
        if count > limit {
            return c.Status(429).SendString("Too Many Requests")
        }
        
        return c.Next()
    }
}
```

### Query Result Caching

```go
func GetUsers(page, limit int) ([]User, error) {
    cacheKey := fmt.Sprintf("users:page:%d:limit:%d", page, limit)
    
    // Try cache
    if cached, err := cache.Get(ctx, cacheKey); err == nil {
        return cached.([]User), nil
    }
    
    // Query database
    users, err := db.Query("SELECT * FROM users LIMIT ? OFFSET ?", 
        limit, (page-1)*limit)
    if err != nil {
        return nil, err
    }
    
    // Cache for 1 minute
    cache.Set(ctx, cacheKey, users, 1*time.Minute)
    
    return users, nil
}
```

### Cache Invalidation

```go
func UpdateUser(userID string, data User) error {
    // Update database
    if err := db.Update(userID, data); err != nil {
        return err
    }
    
    // Invalidate cache
    cache.Delete(ctx, "user:"+userID)
    
    // Invalidate related caches
    cache.DeleteMulti(ctx, []string{
        "user:"+userID+":profile",
        "user:"+userID+":permissions",
        "users:list",
    })
    
    return nil
}
```

## Performance

### Memory Cache
- **Speed**: < 1μs per operation
- **Capacity**: 10,000+ items with 10MB memory
- **Eviction**: LRU (Least Recently Used)
- **Concurrency**: Lock-based, high throughput

### Redis Cache
- **Speed**: 1-5ms per operation (network)
- **Capacity**: Millions of items (limited by RAM)
- **Persistence**: Optional RDB/AOF
- **Concurrency**: Single-threaded, pipelining support

### Multi-Tier Cache
- **L1 Hit**: < 1μs (memory)
- **L2 Hit**: 1-5ms (Redis)
- **Cache Miss**: Database query time
- **Promotion**: Async, no blocking

## Best Practices

1. **Use appropriate TTLs** - Short for volatile data, long for static
2. **Implement cache warming** - Pre-populate on startup
3. **Monitor hit rates** - Aim for > 80% hit rate
4. **Use multi-tier for hot data** - Reduce Redis calls
5. **Invalidate proactively** - Delete on updates
6. **Batch operations** - Use GetMulti/SetMulti for efficiency
7. **Handle cache failures gracefully** - Always have fallback
8. **Use cache keys consistently** - Namespace by resource type
9. **Avoid cache stampede** - Use locking for expensive operations
10. **Monitor memory usage** - Set appropriate max sizes

## Cache Key Patterns

```go
// User data
"user:{id}"                    // User object
"user:{id}:profile"           // User profile
"user:{id}:permissions"       // User permissions

// Session data
"session:{token}"             // Session data
"session:{token}:activity"   // User activity

// API responses
"api:users:list:page:{n}"     // Paginated list
"api:post:{id}"               // Single post

// Rate limiting
"ratelimit:{ip}:{endpoint}"   // Per-IP rate limit
"ratelimit:user:{id}"         // Per-user rate limit

// Counters
"views:post:{id}"             // Post view count
"likes:post:{id}"             // Post like count

// Temporary data
"temp:upload:{id}"            // Temporary upload
"temp:token:{token}"          // Temporary token
```

## Error Handling

```go
value, err := cache.Get(ctx, "key")
if err != nil {
    switch {
    case errors.Is(err, cache.ErrKeyNotFound):
        // Cache miss - fetch from source
    case errors.Is(err, cache.ErrTimeout):
        // Timeout - retry or fallback
    case errors.Is(err, cache.ErrClosed):
        // Cache closed - reinitialize
    default:
        // Unknown error - log and fallback
    }
}
```

## Future Enhancements

- [ ] Distributed locking (Redis-based)
- [ ] Cache stampede protection
- [ ] Cache warming strategies
- [ ] Compression support
- [ ] Cache tags for group invalidation
- [ ] Probabilistic early expiration
- [ ] Cache metrics export (Prometheus)
- [ ] Cache replication across regions
- [ ] Hot key detection and handling
- [ ] Cache versioning for schema changes

## License

MIT License - Part of NeonexCore Framework

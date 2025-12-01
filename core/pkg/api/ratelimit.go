package api

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	tokens         map[string]*TokenBucket
	mu             sync.RWMutex
	maxRequests    int
	windowDuration time.Duration
	cleanup        time.Duration
}

// TokenBucket represents a token bucket for rate limiting
type TokenBucket struct {
	tokens     int
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxRequests int, windowDuration time.Duration) *RateLimiter {
	rl := &RateLimiter{
		tokens:         make(map[string]*TokenBucket),
		maxRequests:    maxRequests,
		windowDuration: windowDuration,
		cleanup:        time.Hour, // Cleanup old entries every hour
	}

	// Start cleanup goroutine
	go rl.cleanupRoutine()

	return rl
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	bucket, exists := rl.tokens[key]
	if !exists {
		bucket = &TokenBucket{
			tokens:     rl.maxRequests,
			lastRefill: time.Now(),
		}
		rl.tokens[key] = bucket
	}
	rl.mu.Unlock()

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill)
	if elapsed >= rl.windowDuration {
		bucket.tokens = rl.maxRequests
		bucket.lastRefill = now
	}

	// Check if request is allowed
	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

// GetRemaining returns remaining tokens for a key
func (rl *RateLimiter) GetRemaining(key string) int {
	rl.mu.RLock()
	bucket, exists := rl.tokens[key]
	rl.mu.RUnlock()

	if !exists {
		return rl.maxRequests
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Refill if needed
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill)
	if elapsed >= rl.windowDuration {
		bucket.tokens = rl.maxRequests
		bucket.lastRefill = now
	}

	return bucket.tokens
}

// GetResetTime returns when the rate limit will reset
func (rl *RateLimiter) GetResetTime(key string) time.Time {
	rl.mu.RLock()
	bucket, exists := rl.tokens[key]
	rl.mu.RUnlock()

	if !exists {
		return time.Now().Add(rl.windowDuration)
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	return bucket.lastRefill.Add(rl.windowDuration)
}

// cleanupRoutine periodically removes old entries
func (rl *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, bucket := range rl.tokens {
			bucket.mu.Lock()
			if now.Sub(bucket.lastRefill) > rl.windowDuration*2 {
				delete(rl.tokens, key)
			}
			bucket.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// RateLimitConfig represents rate limit configuration
type RateLimitConfig struct {
	MaxRequests    int           // Maximum requests per window
	WindowDuration time.Duration // Time window duration
	KeyGenerator   func(*fiber.Ctx) string // Function to generate rate limit key
	SkipFunc       func(*fiber.Ctx) bool   // Function to skip rate limiting
	Handler        fiber.Handler           // Custom handler when limit exceeded
}

// DefaultRateLimitConfig returns default rate limit configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		MaxRequests:    100,
		WindowDuration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		SkipFunc: nil,
		Handler: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(Response{
				Success: false,
				Message: "Too many requests. Please try again later.",
				Timestamp: time.Now().Unix(),
			})
		},
	}
}

// RateLimitMiddleware creates rate limiting middleware
func RateLimitMiddleware(config ...RateLimitConfig) fiber.Handler {
	cfg := DefaultRateLimitConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	limiter := NewRateLimiter(cfg.MaxRequests, cfg.WindowDuration)

	return func(c *fiber.Ctx) error {
		// Skip rate limiting if configured
		if cfg.SkipFunc != nil && cfg.SkipFunc(c) {
			return c.Next()
		}

		// Generate key for this request
		key := cfg.KeyGenerator(c)

		// Check if request is allowed
		if !limiter.Allow(key) {
			// Add rate limit headers
			c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.MaxRequests))
			c.Set("X-RateLimit-Remaining", "0")
			resetTime := limiter.GetResetTime(key)
			c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))
			c.Set("Retry-After", fmt.Sprintf("%d", int(time.Until(resetTime).Seconds())))

			return cfg.Handler(c)
		}

		// Add rate limit headers
		remaining := limiter.GetRemaining(key)
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.MaxRequests))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		resetTime := limiter.GetResetTime(key)
		c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

		return c.Next()
	}
}

// IPRateLimitMiddleware creates IP-based rate limiting middleware
func IPRateLimitMiddleware(maxRequests int, window time.Duration) fiber.Handler {
	return RateLimitMiddleware(RateLimitConfig{
		MaxRequests:    maxRequests,
		WindowDuration: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	})
}

// UserRateLimitMiddleware creates user-based rate limiting middleware
func UserRateLimitMiddleware(maxRequests int, window time.Duration) fiber.Handler {
	return RateLimitMiddleware(RateLimitConfig{
		MaxRequests:    maxRequests,
		WindowDuration: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Get user ID from context (set by auth middleware)
			if userID := c.Locals("user_id"); userID != nil {
				return fmt.Sprintf("user:%v", userID)
			}
			// Fallback to IP if user not authenticated
			return c.IP()
		},
	})
}

// EndpointRateLimitMiddleware creates endpoint-specific rate limiting
func EndpointRateLimitMiddleware(maxRequests int, window time.Duration) fiber.Handler {
	return RateLimitMiddleware(RateLimitConfig{
		MaxRequests:    maxRequests,
		WindowDuration: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			return fmt.Sprintf("%s:%s", c.IP(), c.Path())
		},
	})
}

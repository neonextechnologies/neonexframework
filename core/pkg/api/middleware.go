package api

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	ExposeHeaders    []string
	MaxAge           int
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodPatch,
			fiber.MethodDelete,
			fiber.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"API-Version",
		},
		AllowCredentials: true,
		ExposeHeaders: []string{
			"Content-Length",
			"API-Version",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
		},
		MaxAge: 3600,
	}
}

// ProductionCORSConfig returns production CORS configuration
func ProductionCORSConfig(allowedOrigins []string) CORSConfig {
	return CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodPatch,
			fiber.MethodDelete,
			fiber.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"API-Version",
		},
		AllowCredentials: true,
		ExposeHeaders: []string{
			"Content-Length",
			"API-Version",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
		},
		MaxAge: 86400, // 24 hours
	}
}

// CORSMiddleware creates CORS middleware
func CORSMiddleware(config ...CORSConfig) fiber.Handler {
	cfg := DefaultCORSConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return cors.New(cors.Config{
		AllowOrigins:     joinStrings(cfg.AllowOrigins, ", "),
		AllowMethods:     joinStrings(cfg.AllowMethods, ", "),
		AllowHeaders:     joinStrings(cfg.AllowHeaders, ", "),
		AllowCredentials: cfg.AllowCredentials,
		ExposeHeaders:    joinStrings(cfg.ExposeHeaders, ", "),
		MaxAge:           cfg.MaxAge,
	})
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Prevent MIME type sniffing
		c.Set("X-Content-Type-Options", "nosniff")

		// Enable XSS protection
		c.Set("X-XSS-Protection", "1; mode=block")

		// Prevent clickjacking
		c.Set("X-Frame-Options", "DENY")

		// Force HTTPS
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Referrer policy
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy
		c.Set("Content-Security-Policy", "default-src 'self'")

		// Permissions Policy
		c.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		return c.Next()
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set("X-Request-ID", requestID)
		c.Locals("request_id", requestID)

		return c.Next()
	}
}

// LoggerMiddleware creates a simple request logger
func LoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Log request
		duration := time.Since(start)
		status := c.Response().StatusCode()

		// You can integrate with your logger here
		_ = duration
		_ = status

		return err
	}
}

// CompressionMiddleware enables response compression
func CompressionMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if client accepts compression
		acceptEncoding := c.Get("Accept-Encoding")
		if acceptEncoding == "" {
			return c.Next()
		}

		// Fiber has built-in compression, this is a placeholder
		// In production, use fiber.Compress() middleware
		return c.Next()
	}
}

// Helper functions
func joinStrings(slice []string, sep string) string {
	if len(slice) == 0 {
		return ""
	}
	result := slice[0]
	for i := 1; i < len(slice); i++ {
		result += sep + slice[i]
	}
	return result
}

func generateRequestID() string {
	// Simple request ID generation
	// In production, use UUID or similar
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

package logger

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// HTTPMiddleware creates a Fiber middleware for request logging
func HTTPMiddleware(logger Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get status code
		status := c.Response().StatusCode()

		// Prepare log fields
		fields := Fields{
			"method":     c.Method(),
			"path":       c.Path(),
			"status":     status,
			"duration":   duration.Milliseconds(),
			"ip":         c.IP(),
			"user_agent": c.Get("User-Agent"),
		}

		// Add query params if any
		if len(c.Context().QueryArgs().String()) > 0 {
			fields["query"] = c.Context().QueryArgs().String()
		}

		// Log based on status code
		msg := "HTTP Request"
		if err != nil {
			fields["error"] = err.Error()
			logger.Error(msg, fields)
		} else if status >= 500 {
			logger.Error(msg, fields)
		} else if status >= 400 {
			logger.Warn(msg, fields)
		} else {
			logger.Info(msg, fields)
		}

		return err
	}
}

// RequestIDMiddleware adds a request ID to each request
func RequestIDMiddleware(logger Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
			c.Set("X-Request-ID", requestID)
		}

		// Store request ID in context for later use
		c.Locals("request_id", requestID)
		c.Locals("logger", logger.With(Fields{"request_id": requestID}))

		return c.Next()
	}
}

// GetLogger retrieves the logger from Fiber context
func GetLogger(c *fiber.Ctx) Logger {
	if logger, ok := c.Locals("logger").(Logger); ok {
		return logger
	}
	return defaultLogger
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randString(8)
}

func randString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[time.Now().UnixNano()%int64(len(letterBytes))]
	}
	return string(b)
}

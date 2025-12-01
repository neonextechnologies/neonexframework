package errors

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"neonexcore/pkg/logger"
)

// ErrorResponse represents error response structure
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Code    ErrorCode              `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ErrorHandler creates global error handler middleware
func ErrorHandler(log logger.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Default to 500 Internal Server Error
		code := fiber.StatusInternalServerError
		response := ErrorResponse{
			Error:   "internal_server_error",
			Message: "An unexpected error occurred",
		}

		// Check if it's a Fiber error
		if fiberErr, ok := err.(*fiber.Error); ok {
			code = fiberErr.Code
			response.Message = fiberErr.Message
			response.Error = http.StatusText(code)
		}

		// Check if it's our AppError
		if appErr, ok := err.(*AppError); ok {
			code = appErr.StatusCode
			response.Code = appErr.Code
			response.Message = appErr.Message
			response.Error = string(appErr.Code)
			response.Details = appErr.Details

			// Log error with details
			log.Error("Application error", logger.Fields{
				"code":        appErr.Code,
				"message":     appErr.Message,
				"status_code": appErr.StatusCode,
				"path":        c.Path(),
				"method":      c.Method(),
			})

			// Log underlying error if exists
			if appErr.Err != nil {
				log.Error("Underlying error", logger.Fields{
					"error": appErr.Err.Error(),
				})
			}
		} else {
			// Log unexpected errors
			log.Error("Unexpected error", logger.Fields{
				"error":  err.Error(),
				"path":   c.Path(),
				"method": c.Method(),
			})
		}

		// Send error response
		return c.Status(code).JSON(response)
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(log logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				log.Error("Panic recovered", logger.Fields{
					"panic":  r,
					"path":   c.Path(),
					"method": c.Method(),
				})

				err := NewInternal("Internal server error")
				c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
					Error:   string(err.Code),
					Message: err.Message,
					Code:    err.Code,
				})
			}
		}()
		return c.Next()
	}
}

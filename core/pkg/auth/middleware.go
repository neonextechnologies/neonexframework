package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware creates authentication middleware
func AuthMiddleware(jwtManager *JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "missing authorization header",
			})
		}

		// Extract Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "invalid authorization format",
			})
		}

		token := parts[1]

		// Validate token
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": err.Error(),
			})
		}

		// Store claims in context
		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)
		c.Locals("permissions", claims.Permissions)
		c.Locals("claims", claims)

		return c.Next()
	}
}

// OptionalAuthMiddleware makes authentication optional
func OptionalAuthMiddleware(jwtManager *JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			claims, err := jwtManager.ValidateToken(parts[1])
			if err == nil {
				c.Locals("user_id", claims.UserID)
				c.Locals("email", claims.Email)
				c.Locals("role", claims.Role)
				c.Locals("permissions", claims.Permissions)
				c.Locals("claims", claims)
			}
		}

		return c.Next()
	}
}

// GetUserID gets user ID from context
func GetUserID(c *fiber.Ctx) (uint, bool) {
	userID, ok := c.Locals("user_id").(uint)
	return userID, ok
}

// GetUserEmail gets user email from context
func GetUserEmail(c *fiber.Ctx) (string, bool) {
	email, ok := c.Locals("email").(string)
	return email, ok
}

// GetUserRole gets user role from context
func GetUserRole(c *fiber.Ctx) (string, bool) {
	role, ok := c.Locals("role").(string)
	return role, ok
}

// GetUserPermissions gets user permissions from context
func GetUserPermissions(c *fiber.Ctx) ([]string, bool) {
	permissions, ok := c.Locals("permissions").([]string)
	return permissions, ok
}

// GetClaims gets full claims from context
func GetClaims(c *fiber.Ctx) (*Claims, bool) {
	claims, ok := c.Locals("claims").(*Claims)
	return claims, ok
}

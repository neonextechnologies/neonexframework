package rbac

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

// RequirePermission creates middleware that checks for required permission
func RequirePermission(manager *Manager, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(uint)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "user not authenticated",
			})
		}

		ctx := context.Background()
		hasPermission, err := manager.HasPermission(ctx, userID, permission)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "internal_error",
				"message": "failed to check permission",
			})
		}

		if !hasPermission {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "forbidden",
				"message": "insufficient permissions",
			})
		}

		return c.Next()
	}
}

// RequireRole creates middleware that checks for required role
func RequireRole(manager *Manager, role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(uint)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "user not authenticated",
			})
		}

		ctx := context.Background()
		hasRole, err := manager.HasRole(ctx, userID, role)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "internal_error",
				"message": "failed to check role",
			})
		}

		if !hasRole {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "forbidden",
				"message": "insufficient role",
			})
		}

		return c.Next()
	}
}

// RequireAnyPermission checks if user has any of the given permissions
func RequireAnyPermission(manager *Manager, permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(uint)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "user not authenticated",
			})
		}

		ctx := context.Background()
		hasAny, err := manager.HasAnyPermission(ctx, userID, permissions)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "internal_error",
				"message": "failed to check permissions",
			})
		}

		if !hasAny {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "forbidden",
				"message": "insufficient permissions",
			})
		}

		return c.Next()
	}
}

// RequireAllPermissions checks if user has all of the given permissions
func RequireAllPermissions(manager *Manager, permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(uint)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "user not authenticated",
			})
		}

		ctx := context.Background()
		hasAll, err := manager.HasAllPermissions(ctx, userID, permissions)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "internal_error",
				"message": "failed to check permissions",
			})
		}

		if !hasAll {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "forbidden",
				"message": "insufficient permissions",
			})
		}

		return c.Next()
	}
}

package user

import (
	"neonexcore/internal/core"
	"neonexcore/pkg/auth"
	"neonexcore/pkg/rbac"

	"github.com/gofiber/fiber/v2"
)

func (m *UserModule) Routes(app *fiber.App, c *core.Container) {
	// Resolve controllers from DI container
	authCtrl := core.Resolve[*AuthController](c)
	userCtrl := core.Resolve[*UserController](c)
	
	// Resolve middleware dependencies
	jwtManager := core.Resolve[*auth.JWTManager](c)
	rbacManager := core.Resolve[*rbac.Manager](c)

	// API v1 group
	api := app.Group("/api/v1")

	// ==================== Authentication Routes (Public) ====================
	authGroup := api.Group("/auth")
	{
		// Public auth endpoints
		authGroup.Post("/login", authCtrl.Login)
		authGroup.Post("/register", authCtrl.Register)
		authGroup.Post("/refresh", authCtrl.RefreshToken)
		authGroup.Post("/forgot-password", authCtrl.ForgotPassword)
		authGroup.Post("/reset-password", authCtrl.ResetPassword)
		authGroup.Get("/verify-email/:token", authCtrl.VerifyEmail)

		// Protected auth endpoints (require authentication)
		authProtected := authGroup.Group("", auth.AuthMiddleware(jwtManager))
		authProtected.Post("/logout", authCtrl.Logout)
		authProtected.Get("/profile", authCtrl.GetProfile)
		authProtected.Put("/profile", authCtrl.UpdateProfile)
		authProtected.Post("/change-password", authCtrl.ChangePassword)
		authProtected.Post("/api-key", authCtrl.GenerateAPIKey)
	}

	// ==================== User Management Routes ====================
	usersGroup := api.Group("/users")
	{
		// Public/Optional auth endpoints
		usersGroup.Get("/search", userCtrl.Search)

		// Protected endpoints (require authentication)
		usersProtected := usersGroup.Group("", auth.AuthMiddleware(jwtManager))
		{
			// Read operations (require 'users.read' permission)
			usersProtected.Get("/", 
				rbac.RequirePermission(rbacManager, "users.read"),
				userCtrl.GetAll,
			)
			usersProtected.Get("/:id", 
				rbac.RequirePermission(rbacManager, "users.read"),
				userCtrl.GetByID,
			)

			// Write operations (require 'users.create' permission)
			usersProtected.Post("/", 
				rbac.RequirePermission(rbacManager, "users.create"),
				userCtrl.Create,
			)

			// Update operations (require 'users.update' permission)
			usersProtected.Put("/:id", 
				rbac.RequirePermission(rbacManager, "users.update"),
				userCtrl.Update,
			)

			// Delete operations (require 'users.delete' permission)
			usersProtected.Delete("/:id", 
				rbac.RequirePermission(rbacManager, "users.delete"),
				userCtrl.Delete,
			)

			// Role management (require 'users.manage-roles' permission)
			usersProtected.Get("/:id/roles",
				rbac.RequirePermission(rbacManager, "users.manage-roles"),
				userCtrl.GetUserRoles,
			)
			usersProtected.Post("/:id/roles",
				rbac.RequirePermission(rbacManager, "users.manage-roles"),
				userCtrl.AssignRole,
			)
			usersProtected.Delete("/:id/roles/:roleId",
				rbac.RequirePermission(rbacManager, "users.manage-roles"),
				userCtrl.RemoveRole,
			)

			// Permission management (require 'users.manage-permissions' permission)
			usersProtected.Get("/:id/permissions",
				rbac.RequirePermission(rbacManager, "users.manage-permissions"),
				userCtrl.GetUserPermissions,
			)
		}
	}

	// ==================== Legacy Routes (backward compatibility) ====================
	// Keep old /user routes for backward compatibility
	legacyGroup := app.Group("/user")
	{
		legacyGroup.Get("/search", userCtrl.Search)
		
		legacyProtected := legacyGroup.Group("", auth.AuthMiddleware(jwtManager))
		{
			legacyProtected.Get("/", userCtrl.GetAll)
			legacyProtected.Get("/:id", userCtrl.GetByID)
			legacyProtected.Post("/", userCtrl.Create)
			legacyProtected.Put("/:id", userCtrl.Update)
			legacyProtected.Delete("/:id", userCtrl.Delete)
		}
	}
}

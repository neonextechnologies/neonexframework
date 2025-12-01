package admin

import (
	"neonexcore/internal/core"
	"neonexcore/pkg/rbac"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(router fiber.Router, container *core.Container) {
	// Get dependencies
	controller := container.GetSingleton("admin.controller").(*Controller)
	rbacManager := container.GetSingleton("rbac.manager").(*rbac.Manager)

	// Create admin routes group
	admin := router.Group("/admin")

	// Apply authentication middleware (assuming it exists)
	// admin.Use(auth.Middleware())

	// Dashboard routes (require admin.dashboard.view permission)
	admin.Get("/dashboard", 
		rbac.RequirePermission(rbacManager, "admin.dashboard.view"),
		controller.GetDashboard,
	)

	// Statistics routes (require admin.system.view permission)
	admin.Get("/stats", 
		rbac.RequirePermission(rbacManager, "admin.system.view"),
		controller.GetStats,
	)
	admin.Get("/stats/users", 
		rbac.RequirePermission(rbacManager, "admin.system.view"),
		controller.GetUserStats,
	)
	admin.Get("/stats/modules", 
		rbac.RequirePermission(rbacManager, "admin.system.view"),
		controller.GetModuleStats,
	)

	// System health route
	admin.Get("/health", 
		rbac.RequirePermission(rbacManager, "admin.system.view"),
		controller.GetSystemHealth,
	)

	// Audit logs routes (require admin.logs.view permission)
	admin.Get("/audit-logs", 
		rbac.RequirePermission(rbacManager, "admin.logs.view"),
		controller.GetAuditLogs,
	)
	admin.Get("/activity", 
		rbac.RequirePermission(rbacManager, "admin.logs.view"),
		controller.GetActivitySummary,
	)

	// Settings routes (require admin.settings.manage permission)
	settingsGroup := admin.Group("/settings")
	settingsGroup.Use(rbac.RequirePermission(rbacManager, "admin.settings.manage"))
	
	settingsGroup.Get("/", controller.GetSettings)
	settingsGroup.Get("/:key", controller.GetSetting)
	settingsGroup.Post("/", controller.CreateSetting)
	settingsGroup.Put("/:key", controller.UpdateSetting)
	settingsGroup.Delete("/:key", controller.DeleteSetting)
}

package main

import (
	"context"
	"fmt"
	"log"

	"neonexcore/internal/config"
	"neonexcore/internal/core"
	coreAdmin "neonexcore/modules/admin"
	coreUser "neonexcore/modules/user"
	"neonexcore/pkg/database"
	"neonexcore/pkg/logger"
	"neonexcore/pkg/module"
	"neonexcore/pkg/rbac"
	
	"neonexframework/modules/frontend"
	"neonexframework/modules/web"
)

func main() {
	fmt.Println("=====================================")
	fmt.Println("NeonEx Framework v0.2.0")
	fmt.Println("Full-Stack Go Framework")
	fmt.Println("=====================================")
	fmt.Println()

	// Register core modules
	core.ModuleMap["user"] = func() core.Module { return coreUser.New() }
	core.ModuleMap["admin"] = func() core.Module { return coreAdmin.New() }
	
	// Register framework modules
	core.ModuleMap["frontend"] = func() core.Module { return frontend.New() }
	// Initialize Logger
	loggerConfig := logger.LoadConfig()
	if err := app.InitLogger(loggerConfig); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	app.Logger.Info("✓ Logger initialized")

	// Initialize Database
	if err := app.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	app.Logger.Info("✓ Database connected")
	if err := app.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	app.Logger.Info("Database connected successfully")

	// Register core models for auto-migration
	app.RegisterModels(
		// Core User Models
		&coreUser.User{},

		// RBAC Models
		&rbac.Role{},
		&rbac.Permission{},
		&rbac.UserRole{},
		&rbac.UserPermission{},

		// Module System Models
		&module.Module{},
		&module.ModuleDependency{},
		&module.ModuleMigration{},

		// Admin Models
		&coreAdmin.AuditLog{},
		&coreAdmin.SystemSettings{},
		&coreAdmin.BackupInfo{},
	)

	// Run auto-migration
	app.Logger.Info("Running database migrations...")
	if err := app.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	app.Logger.Info("✓ Migrations completed")

	// Seed RBAC data
	ctx := context.Background()
	rbacManager := rbac.NewManager(config.DB.GetDB())
	
	app.Logger.Info("Seeding default roles and permissions...")
	if err := rbacManager.SeedDefaultRoles(ctx); err != nil {
		log.Printf("Warning: Failed to seed roles: %v", err)
	}

	if err := seedUserPermissions(ctx, rbacManager); err != nil {
		log.Printf("Warning: Failed to seed permissions: %v", err)
	}
	app.Logger.Info("RBAC data seeded successfully")

	// Seed database (optional)
	app.Logger.Info("Running database seeders...")
	seeder := database.NewSeederManager(config.DB.GetDB())
	seeder.Register(coreUser.NewUserSeeder(config.DB.GetDB()))
	seeder.Register(coreAdmin.NewAdminSeeder(config.DB.GetDB()))
	if err := seeder.Run(context.Background()); err != nil {
		log.Printf("Warning: Seeding failed: %v", err)
	}
	app.Logger.Info("Database seeding completed")

	// Load modules
	app.Logger.Info("Loading framework modules...")
	app.Registry.AutoDiscover()
	app.Boot()
	app.Registry.Load()
	app.Logger.Info("All modules loaded successfully")	// Display startup information
	fmt.Println()
	fmt.Println("=====================================")
	fmt.Println("🚀 Server starting...")
	fmt.Println("📍 Admin Panel: http://localhost:8080/admin")
	fmt.Println("📍 API: http://localhost:8080/api/v1")
	fmt.Println("📍 Health: http://localhost:8080/health")
	fmt.Println("=====================================")
	fmt.Println()

	// Start HTTP server
	app.StartHTTP()
	// Start HTTP server
	app.StartHTTP()
}/ seedUserPermissions seeds default user module permissions
func seedUserPermissions(ctx context.Context, rbacManager *rbac.Manager) error {
	permissions := []rbac.Permission{
		{
			Name:        "Read Users",
			Slug:        "users.read",
			Description: "View user list and details",
			Module:      "user",
			Category:    "users",
		},
		{
			Name:        "Create Users",
			Slug:        "users.create",
			Description: "Create new users",
			Module:      "user",
			Category:    "users",
		},
		{
			Name:        "Update Users",
			Slug:        "users.update",
			Description: "Update existing users",
			Module:      "user",
			Category:    "users",
		},
		{
			Name:        "Delete Users",
			Slug:        "users.delete",
			Description: "Delete users",
			Module:      "user",
			Category:    "users",
		},
	}

	for _, perm := range permissions {
		existing, _ := rbacManager.GetPermissionBySlug(ctx, perm.Slug)
		if existing == nil {
			if err := rbacManager.CreatePermission(ctx, &perm); err != nil {
				return fmt.Errorf("failed to create permission %s: %w", perm.Slug, err)
			}
		}
	}

	// Assign all permissions to super-admin role
	superAdminRole, _ := rbacManager.GetRoleBySlug(ctx, "super-admin")
	if superAdminRole != nil {
		var permIDs []uint
		for _, perm := range permissions {
			p, _ := rbacManager.GetPermissionBySlug(ctx, perm.Slug)
			if p != nil {
				permIDs = append(permIDs, p.ID)
			}
		}
		if len(permIDs) > 0 {
			rbacManager.SyncRolePermissions(ctx, superAdminRole.ID, permIDs)
		}
	}

	return nil
}

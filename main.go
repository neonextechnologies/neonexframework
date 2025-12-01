package main

import (
	"context"
	"fmt"
	"log"

	"neonexframework/pkg/app"
	"neonexframework/pkg/config"
	
	coreAdmin "neonexcore/modules/admin"
	coreUser "neonexcore/modules/user"
	"neonexcore/pkg/database"
	"neonexcore/pkg/logger"
	"neonexcore/pkg/module"
	"neonexcore/pkg/rbac"
)

func main() {
	fmt.Println("=====================================")
	fmt.Println("NeonEx Framework v0.1.0")
	fmt.Println("Full-Stack Go Framework")
	fmt.Println("=====================================")
	fmt.Println()

	// Register core module factories
	app.ModuleMap["user"] = func() app.Module { return coreUser.New() }
	app.ModuleMap["admin"] = func() app.Module { return coreAdmin.New() }

	appInstance := app.NewApp()

	// Initialize Logger
	loggerConfig := logger.LoadConfig()
	if err := appInstance.InitLogger(loggerConfig); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	appInstance.Logger.Info("Logger initialized successfully")

	// Initialize Database
	if err := appInstance.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	appInstance.Logger.Info("Database connected successfully")

	// Register core models for auto-migration
	appInstance.RegisterModels(
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
	appInstance.Logger.Info("Running database migrations...")
	if err := appInstance.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	appInstance.Logger.Info("Migrations completed successfully")

	// Seed RBAC data
	ctx := context.Background()
	rbacManager := rbac.NewManager(config.GetDB())
	
	appInstance.Logger.Info("Seeding default roles and permissions...")
	if err := rbacManager.SeedDefaultRoles(ctx); err != nil {
		log.Printf("Warning: Failed to seed roles: %v", err)
	}

	if err := seedUserPermissions(ctx, rbacManager); err != nil {
		log.Printf("Warning: Failed to seed permissions: %v", err)
	}
	appInstance.Logger.Info("RBAC data seeded successfully")

	// Seed database (optional)
	appInstance.Logger.Info("Running database seeders...")
	seeder := database.NewSeederManager(config.GetDB())
	seeder.Register(coreUser.NewUserSeeder(config.GetDB()))
	seeder.Register(coreAdmin.NewAdminSeeder(config.GetDB()))
	if err := seeder.Run(context.Background()); err != nil {
		log.Printf("Warning: Seeding failed: %v", err)
	}
	appInstance.Logger.Info("Database seeding completed")

	// Load modules
	appInstance.Logger.Info("Loading framework modules...")
	appInstance.Registry.AutoDiscover()
	appInstance.Boot()
	appInstance.Registry.Load()
	appInstance.Logger.Info("All modules loaded successfully")	// Display startup information
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
	appInstance.StartHTTP()
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
		{
			Name:        "Manage User Roles",
			Slug:        "users.manage-roles",
			Description: "Assign and remove roles from users",
			Module:      "user",
			Category:    "users",
		},
		{
			Name:        "Manage User Permissions",
			Slug:        "users.manage-permissions",
			Description: "Assign and remove permissions from users",
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

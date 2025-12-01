package main

import (
	"context"
	"fmt"
	"log"

	"neonexcore/internal/config"
	"neonexcore/internal/core"
	"neonexcore/modules/admin"
	"neonexcore/modules/user"
	"neonexcore/pkg/api"
	"neonexcore/pkg/database"
	"neonexcore/pkg/logger"
	"neonexcore/pkg/module"
	"neonexcore/pkg/rbac"
)

func main() {
	fmt.Println("Neonex Core v0.1 starting...")

	// Register module factories
	core.ModuleMap["user"] = func() core.Module { return user.New() }
	core.ModuleMap["admin"] = func() core.Module { return admin.New() }

	app := core.NewApp()

	// Initialize Logger
	loggerConfig := logger.LoadConfig()
	if err := app.InitLogger(loggerConfig); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Initialize Database
	if err := app.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Register models for auto-migration
	app.RegisterModels(
		&user.User{},
		&rbac.Role{},
		&rbac.Permission{},
		&rbac.UserRole{},
		&rbac.UserPermission{},
		&module.Module{},
		&module.ModuleDependency{},
		&module.ModuleMigration{},
		&admin.AuditLog{},
		&admin.SystemSettings{},
		&admin.BackupInfo{},
	)

	// Run auto-migration
	if err := app.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed RBAC data (roles and permissions)
	ctx := context.Background()
	rbacManager := rbac.NewManager(config.DB.GetDB())
	
	app.Logger.Info("Seeding default roles...")
	if err := rbacManager.SeedDefaultRoles(ctx); err != nil {
		log.Printf("Warning: Failed to seed roles: %v", err)
	}

	app.Logger.Info("Seeding user permissions...")
	if err := seedUserPermissions(ctx, rbacManager); err != nil {
		log.Printf("Warning: Failed to seed permissions: %v", err)
	}

	// Seed database (optional)
	seeder := database.NewSeederManager(config.DB.GetDB())
	seeder.Register(user.NewUserSeeder(config.DB.GetDB()))
	seeder.Register(admin.NewAdminSeeder(config.DB.GetDB()))
	if err := seeder.Run(context.Background()); err != nil {
		log.Printf("Warning: Seeding failed: %v", err)
	}

	// Load modules
	app.Registry.AutoDiscover()
	app.Boot()
	app.Registry.Load()

	// Start HTTP server
	app.StartHTTP()
}

// seedUserPermissions seeds default user module permissions
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

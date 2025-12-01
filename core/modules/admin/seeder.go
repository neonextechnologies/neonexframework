package admin

import (
	"context"
	"fmt"

	"neonexcore/modules/user"
	"neonexcore/pkg/rbac"

	"gorm.io/gorm"
)

type AdminSeeder struct {
	db *gorm.DB
}

func NewAdminSeeder(db *gorm.DB) *AdminSeeder {
	return &AdminSeeder{db: db}
}

func (s *AdminSeeder) Name() string {
	return "AdminSeeder"
}

func (s *AdminSeeder) Run(ctx context.Context) error {
	fmt.Println("Seeding admin data...")

	// Create admin permissions
	if err := s.seedPermissions(ctx); err != nil {
		return fmt.Errorf("failed to seed admin permissions: %w", err)
	}

	// Create default settings
	if err := s.seedSettings(ctx); err != nil {
		return fmt.Errorf("failed to seed settings: %w", err)
	}

	fmt.Println("Admin data seeded successfully")
	return nil
}

func (s *AdminSeeder) seedPermissions(ctx context.Context) error {
	rbacManager := rbac.NewManager(s.db)

	permissions := []rbac.Permission{
		{
			Name:        "View Dashboard",
			Slug:        "admin.dashboard.view",
			Description: "Access admin dashboard",
			Module:      "admin",
			Category:    "admin",
		},
		{
			Name:        "View System Stats",
			Slug:        "admin.system.view",
			Description: "View system statistics and health",
			Module:      "admin",
			Category:    "admin",
		},
		{
			Name:        "Manage Users (Admin)",
			Slug:        "admin.users.manage",
			Description: "Full admin access to user management",
			Module:      "admin",
			Category:    "admin",
		},
		{
			Name:        "Manage Modules (Admin)",
			Slug:        "admin.modules.manage",
			Description: "Full admin access to module management",
			Module:      "admin",
			Category:    "admin",
		},
		{
			Name:        "Manage Roles (Admin)",
			Slug:        "admin.roles.manage",
			Description: "Full admin access to role management",
			Module:      "admin",
			Category:    "admin",
		},
		{
			Name:        "Manage Settings",
			Slug:        "admin.settings.manage",
			Description: "Manage system settings",
			Module:      "admin",
			Category:    "admin",
		},
		{
			Name:        "View Audit Logs",
			Slug:        "admin.logs.view",
			Description: "View system audit logs",
			Module:      "admin",
			Category:    "admin",
		},
	}

	for _, perm := range permissions {
		existing, _ := rbacManager.GetPermissionBySlug(ctx, perm.Slug)
		if existing == nil {
			if err := rbacManager.CreatePermission(ctx, &perm); err != nil {
				return fmt.Errorf("failed to create permission %s: %w", perm.Slug, err)
			}
			fmt.Printf("  ✓ Created permission: %s\n", perm.Slug)
		}
	}

	// Assign all admin permissions to super-admin role
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
			fmt.Printf("  ✓ Assigned %d admin permissions to super-admin role\n", len(permIDs))
		}
	}

	return nil
}

func (s *AdminSeeder) seedSettings(ctx context.Context) error {
	settings := []SystemSettings{
		{
			Key:         "site.name",
			Value:       "Neonex Core",
			Type:        "string",
			Category:    "general",
			Description: "Site name displayed in UI",
			IsPublic:    true,
		},
		{
			Key:         "site.description",
			Value:       "Modular Backend Framework",
			Type:        "string",
			Category:    "general",
			Description: "Site description",
			IsPublic:    true,
		},
		{
			Key:         "maintenance.mode",
			Value:       "false",
			Type:        "bool",
			Category:    "system",
			Description: "Enable maintenance mode",
			IsPublic:    false,
		},
		{
			Key:         "registration.enabled",
			Value:       "true",
			Type:        "bool",
			Category:    "auth",
			Description: "Allow user registration",
			IsPublic:    true,
		},
		{
			Key:         "password.min_length",
			Value:       "8",
			Type:        "int",
			Category:    "auth",
			Description: "Minimum password length",
			IsPublic:    true,
		},
		{
			Key:         "session.timeout",
			Value:       "3600",
			Type:        "int",
			Category:    "auth",
			Description: "Session timeout in seconds",
			IsPublic:    false,
		},
		{
			Key:         "api.rate_limit",
			Value:       "100",
			Type:        "int",
			Category:    "api",
			Description: "API rate limit per minute",
			IsPublic:    false,
		},
		{
			Key:         "logs.retention_days",
			Value:       "30",
			Type:        "int",
			Category:    "system",
			Description: "Number of days to keep audit logs",
			IsPublic:    false,
		},
	}

	for _, setting := range settings {
		var existing SystemSettings
		result := s.db.WithContext(ctx).Where("key = ?", setting.Key).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			if err := s.db.WithContext(ctx).Create(&setting).Error; err != nil {
				return fmt.Errorf("failed to create setting %s: %w", setting.Key, err)
			}
			fmt.Printf("  ✓ Created setting: %s\n", setting.Key)
		}
	}

	return nil
}

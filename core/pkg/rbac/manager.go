package rbac

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// Manager handles RBAC operations
type Manager struct {
	db *gorm.DB
}

// NewManager creates a new RBAC manager
func NewManager(db *gorm.DB) *Manager {
	return &Manager{db: db}
}

// AssignRole assigns a role to a user
func (m *Manager) AssignRole(ctx context.Context, userID, roleID uint) error {
	userRole := &UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	return m.db.WithContext(ctx).Create(userRole).Error
}

// RemoveRole removes a role from a user
func (m *Manager) RemoveRole(ctx context.Context, userID, roleID uint) error {
	return m.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&UserRole{}).Error
}

// AssignPermission assigns a permission to a user
func (m *Manager) AssignPermission(ctx context.Context, userID, permissionID uint) error {
	userPermission := &UserPermission{
		UserID:       userID,
		PermissionID: permissionID,
	}
	return m.db.WithContext(ctx).Create(userPermission).Error
}

// RemovePermission removes a permission from a user
func (m *Manager) RemovePermission(ctx context.Context, userID, permissionID uint) error {
	return m.db.WithContext(ctx).
		Where("user_id = ? AND permission_id = ?", userID, permissionID).
		Delete(&UserPermission{}).Error
}

// GetUserRoles gets all roles for a user
func (m *Manager) GetUserRoles(ctx context.Context, userID uint) ([]Role, error) {
	var roles []Role
	err := m.db.WithContext(ctx).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error
	return roles, err
}

// GetUserPermissions gets all permissions for a user (from roles + direct)
func (m *Manager) GetUserPermissions(ctx context.Context, userID uint) ([]Permission, error) {
	var permissions []Permission

	// Get permissions from roles
	err := m.db.WithContext(ctx).
		Distinct().
		Table("permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Find(&permissions).Error

	if err != nil {
		return nil, err
	}

	// Get direct permissions
	var directPermissions []Permission
	err = m.db.WithContext(ctx).
		Joins("JOIN user_permissions ON user_permissions.permission_id = permissions.id").
		Where("user_permissions.user_id = ?", userID).
		Find(&directPermissions).Error

	if err != nil {
		return nil, err
	}

	// Merge permissions (avoid duplicates)
	permMap := make(map[uint]Permission)
	for _, p := range permissions {
		permMap[p.ID] = p
	}
	for _, p := range directPermissions {
		permMap[p.ID] = p
	}

	result := make([]Permission, 0, len(permMap))
	for _, p := range permMap {
		result = append(result, p)
	}

	return result, nil
}

// HasRole checks if user has a specific role
func (m *Manager) HasRole(ctx context.Context, userID uint, roleSlug string) (bool, error) {
	var count int64
	err := m.db.WithContext(ctx).
		Table("user_roles").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.slug = ?", userID, roleSlug).
		Count(&count).Error
	return count > 0, err
}

// HasPermission checks if user has a specific permission
func (m *Manager) HasPermission(ctx context.Context, userID uint, permissionSlug string) (bool, error) {
	var count int64

	// Check from roles
	err := m.db.WithContext(ctx).
		Table("permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ? AND permissions.slug = ?", userID, permissionSlug).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	// Check direct permissions
	err = m.db.WithContext(ctx).
		Table("user_permissions").
		Joins("JOIN permissions ON permissions.id = user_permissions.permission_id").
		Where("user_permissions.user_id = ? AND permissions.slug = ?", userID, permissionSlug).
		Count(&count).Error

	return count > 0, err
}

// HasAnyPermission checks if user has any of the given permissions
func (m *Manager) HasAnyPermission(ctx context.Context, userID uint, permissionSlugs []string) (bool, error) {
	for _, slug := range permissionSlugs {
		has, err := m.HasPermission(ctx, userID, slug)
		if err != nil {
			return false, err
		}
		if has {
			return true, nil
		}
	}
	return false, nil
}

// HasAllPermissions checks if user has all of the given permissions
func (m *Manager) HasAllPermissions(ctx context.Context, userID uint, permissionSlugs []string) (bool, error) {
	for _, slug := range permissionSlugs {
		has, err := m.HasPermission(ctx, userID, slug)
		if err != nil {
			return false, err
		}
		if !has {
			return false, nil
		}
	}
	return true, nil
}

// CreateRole creates a new role
func (m *Manager) CreateRole(ctx context.Context, role *Role) error {
	return m.db.WithContext(ctx).Create(role).Error
}

// CreatePermission creates a new permission
func (m *Manager) CreatePermission(ctx context.Context, permission *Permission) error {
	return m.db.WithContext(ctx).Create(permission).Error
}

// AttachPermissionToRole attaches a permission to a role
func (m *Manager) AttachPermissionToRole(ctx context.Context, roleID, permissionID uint) error {
	return m.db.WithContext(ctx).
		Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)", roleID, permissionID).
		Error
}

// DetachPermissionFromRole detaches a permission from a role
func (m *Manager) DetachPermissionFromRole(ctx context.Context, roleID, permissionID uint) error {
	return m.db.WithContext(ctx).
		Exec("DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?", roleID, permissionID).
		Error
}

// SyncRolePermissions syncs permissions for a role
func (m *Manager) SyncRolePermissions(ctx context.Context, roleID uint, permissionIDs []uint) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing
		if err := tx.Exec("DELETE FROM role_permissions WHERE role_id = ?", roleID).Error; err != nil {
			return err
		}

		// Insert new
		for _, permID := range permissionIDs {
			if err := tx.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)", roleID, permID).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetPermissionsByModule gets permissions by module
func (m *Manager) GetPermissionsByModule(ctx context.Context, module string) ([]Permission, error) {
	var permissions []Permission
	err := m.db.WithContext(ctx).Where("module = ?", module).Find(&permissions).Error
	return permissions, err
}

// GetRoleBySlug gets a role by slug
func (m *Manager) GetRoleBySlug(ctx context.Context, slug string) (*Role, error) {
	var role Role
	err := m.db.WithContext(ctx).
		Preload("Permissions").
		Where("slug = ?", slug).
		First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetPermissionBySlug gets a permission by slug
func (m *Manager) GetPermissionBySlug(ctx context.Context, slug string) (*Permission, error) {
	var permission Permission
	err := m.db.WithContext(ctx).Where("slug = ?", slug).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// SeedDefaultRoles seeds default system roles
func (m *Manager) SeedDefaultRoles(ctx context.Context) error {
	roles := []Role{
		{
			Name:        "Super Admin",
			Slug:        "super-admin",
			Description: "Full system access",
			IsSystem:    true,
		},
		{
			Name:        "Admin",
			Slug:        "admin",
			Description: "Administrative access",
			IsSystem:    true,
		},
		{
			Name:        "User",
			Slug:        "user",
			Description: "Regular user access",
			IsSystem:    true,
		},
	}

	for _, role := range roles {
		var existing Role
		err := m.db.WithContext(ctx).Where("slug = ?", role.Slug).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := m.db.WithContext(ctx).Create(&role).Error; err != nil {
				return fmt.Errorf("failed to create role %s: %w", role.Slug, err)
			}
		}
	}

	return nil
}

# Authorization

NeonEx Framework provides a powerful Role-Based Access Control (RBAC) system for managing user permissions and access control throughout your application.

## Table of Contents

- [Overview](#overview)
- [RBAC Manager](#rbac-manager)
- [Roles](#roles)
- [Permissions](#permissions)
- [User Assignments](#user-assignments)
- [Middleware](#middleware)
- [Checking Permissions](#checking-permissions)
- [Best Practices](#best-practices)
- [Advanced Patterns](#advanced-patterns)
- [Troubleshooting](#troubleshooting)

## Overview

NeonEx uses a flexible RBAC system where:
- **Users** can have multiple **Roles**
- **Roles** contain multiple **Permissions**
- **Users** can also have direct **Permissions** (overrides)
- **Permissions** are organized by **Module**

### Default Roles

- `super-admin` - Full system access
- `admin` - Administrative access
- `editor` - Content management
- `user` - Regular user access

## RBAC Manager

### Initialization

```go
import (
    "neonexcore/pkg/rbac"
    "gorm.io/gorm"
)

func InitializeRBAC(db *gorm.DB) *rbac.Manager {
    return rbac.NewManager(db)
}
```

### Manager Structure

```go
// core/pkg/rbac/manager.go
type Manager struct {
    db *gorm.DB
}

func NewManager(db *gorm.DB) *Manager {
    return &Manager{db: db}
}
```

## Roles

### Role Model

```go
type Role struct {
    ID          uint           `gorm:"primarykey" json:"id"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
    
    Name        string       `gorm:"size:100;not null" json:"name"`
    Slug        string       `gorm:"size:100;unique;not null" json:"slug"`
    Description string       `gorm:"type:text" json:"description"`
    IsSystem    bool         `gorm:"default:false" json:"is_system"`
    Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}
```

### Creating Roles

```go
func (m *Manager) CreateRole(ctx context.Context, role *Role) error {
    return m.db.WithContext(ctx).Create(role).Error
}

// Example usage
role := &rbac.Role{
    Name:        "Manager",
    Slug:        "manager",
    Description: "Can manage team resources",
    IsSystem:    false,
}
err := rbacManager.CreateRole(ctx, role)
```

### Getting Roles

```go
// Get role by slug
func (m *Manager) GetRoleBySlug(ctx context.Context, slug string) (*Role, error) {
    var role Role
    err := m.db.WithContext(ctx).
        Preload("Permissions").
        Where("slug = ?", slug).
        First(&role).Error
    return &role, err
}

// Example
adminRole, err := rbacManager.GetRoleBySlug(ctx, "admin")
```

### Assigning Roles to Users

```go
// Assign role to user
func (m *Manager) AssignRole(ctx context.Context, userID, roleID uint) error {
    userRole := &UserRole{
        UserID: userID,
        RoleID: roleID,
    }
    return m.db.WithContext(ctx).Create(userRole).Error
}

// Example
err := rbacManager.AssignRole(ctx, user.ID, adminRole.ID)
```

### Removing Roles

```go
func (m *Manager) RemoveRole(ctx context.Context, userID, roleID uint) error {
    return m.db.WithContext(ctx).
        Where("user_id = ? AND role_id = ?", userID, roleID).
        Delete(&UserRole{}).Error
}
```

### Getting User Roles

```go
func (m *Manager) GetUserRoles(ctx context.Context, userID uint) ([]Role, error) {
    var roles []Role
    err := m.db.WithContext(ctx).
        Joins("JOIN user_roles ON user_roles.role_id = roles.id").
        Where("user_roles.user_id = ?", userID).
        Find(&roles).Error
    return roles, err
}

// Example
roles, err := rbacManager.GetUserRoles(ctx, user.ID)
for _, role := range roles {
    fmt.Printf("User has role: %s\n", role.Name)
}
```

## Permissions

### Permission Model

```go
type Permission struct {
    ID          uint           `gorm:"primarykey" json:"id"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
    
    Module      string `gorm:"size:50;not null;index" json:"module"`
    Name        string `gorm:"size:100;not null" json:"name"`
    Slug        string `gorm:"size:100;unique;not null" json:"slug"`
    Description string `gorm:"type:text" json:"description"`
}
```

### Common Permission Patterns

```go
// CRUD permissions
users.read    - View users
users.create  - Create new users
users.update  - Edit existing users
users.delete  - Delete users

// Module permissions
products.read
products.create
products.update
products.delete
products.publish

// Admin permissions
admin.dashboard
admin.settings
admin.logs
admin.users.impersonate
```

### Creating Permissions

```go
func (m *Manager) CreatePermission(ctx context.Context, permission *Permission) error {
    return m.db.WithContext(ctx).Create(permission).Error
}

// Example
permission := &rbac.Permission{
    Module:      "users",
    Name:        "View Users",
    Slug:        "users.read",
    Description: "Permission to view user list",
}
err := rbacManager.CreatePermission(ctx, permission)
```

### Getting Permissions

```go
// Get permission by slug
func (m *Manager) GetPermissionBySlug(ctx context.Context, slug string) (*Permission, error) {
    var permission Permission
    err := m.db.WithContext(ctx).Where("slug = ?", slug).First(&permission).Error
    return &permission, err
}

// Get permissions by module
func (m *Manager) GetPermissionsByModule(ctx context.Context, module string) ([]Permission, error) {
    var permissions []Permission
    err := m.db.WithContext(ctx).Where("module = ?", module).Find(&permissions).Error
    return permissions, err
}
```

### Attaching Permissions to Roles

```go
func (m *Manager) AttachPermissionToRole(ctx context.Context, roleID, permissionID uint) error {
    return m.db.WithContext(ctx).
        Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)", 
             roleID, permissionID).
        Error
}

// Example
adminRole, _ := rbacManager.GetRoleBySlug(ctx, "admin")
viewUsersPermission, _ := rbacManager.GetPermissionBySlug(ctx, "users.read")
err := rbacManager.AttachPermissionToRole(ctx, adminRole.ID, viewUsersPermission.ID)
```

### Syncing Role Permissions

```go
func (m *Manager) SyncRolePermissions(ctx context.Context, roleID uint, permissionIDs []uint) error {
    return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // Delete existing permissions
        if err := tx.Exec("DELETE FROM role_permissions WHERE role_id = ?", roleID).Error; err != nil {
            return err
        }
        
        // Insert new permissions
        for _, permID := range permissionIDs {
            if err := tx.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)", 
                             roleID, permID).Error; err != nil {
                return err
            }
        }
        
        return nil
    })
}

// Example: Assign specific permissions to editor role
editorPermissions := []uint{1, 2, 5, 8} // Permission IDs
err := rbacManager.SyncRolePermissions(ctx, editorRole.ID, editorPermissions)
```

## User Assignments

### Direct Permission Assignment

```go
// Assign permission directly to user (bypasses roles)
func (m *Manager) AssignPermission(ctx context.Context, userID, permissionID uint) error {
    userPermission := &UserPermission{
        UserID:       userID,
        PermissionID: permissionID,
    }
    return m.db.WithContext(ctx).Create(userPermission).Error
}

// Remove direct permission
func (m *Manager) RemovePermission(ctx context.Context, userID, permissionID uint) error {
    return m.db.WithContext(ctx).
        Where("user_id = ? AND permission_id = ?", userID, permissionID).
        Delete(&UserPermission{}).Error
}
```

### Getting User Permissions

```go
// Get all user permissions (from roles + direct)
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
    
    // Merge (avoiding duplicates)
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
```

## Checking Permissions

### HasRole

```go
func (m *Manager) HasRole(ctx context.Context, userID uint, roleSlug string) (bool, error) {
    var count int64
    err := m.db.WithContext(ctx).
        Table("user_roles").
        Joins("JOIN roles ON roles.id = user_roles.role_id").
        Where("user_roles.user_id = ? AND roles.slug = ?", userID, roleSlug).
        Count(&count).Error
    return count > 0, err
}

// Example
isAdmin, err := rbacManager.HasRole(ctx, user.ID, "admin")
if isAdmin {
    // User is an admin
}
```

### HasPermission

```go
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

// Example
canDelete, err := rbacManager.HasPermission(ctx, user.ID, "users.delete")
if !canDelete {
    return errors.NewForbidden("You don't have permission to delete users")
}
```

### HasAnyPermission

```go
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

// Example: User needs either read or update permission
hasAccess, _ := rbacManager.HasAnyPermission(ctx, user.ID, []string{
    "products.read",
    "products.update",
})
```

### HasAllPermissions

```go
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
```

## Middleware

### RequirePermission Middleware

```go
// pkg/http/middleware/permission.go
package middleware

func RequirePermission(rbacManager *rbac.Manager, permission string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        userID := c.Locals("user_id").(uint)
        
        hasPermission, err := rbacManager.HasPermission(c.Context(), userID, permission)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{
                "error": "Failed to check permissions",
            })
        }
        
        if !hasPermission {
            return c.Status(403).JSON(fiber.Map{
                "error": "Insufficient permissions",
            })
        }
        
        return c.Next()
    }
}
```

### RequireRole Middleware

```go
func RequireRole(rbacManager *rbac.Manager, role string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        userID := c.Locals("user_id").(uint)
        
        hasRole, err := rbacManager.HasRole(c.Context(), userID, role)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{
                "error": "Failed to check role",
            })
        }
        
        if !hasRole {
            return c.Status(403).JSON(fiber.Map{
                "error": "Insufficient role",
            })
        }
        
        return c.Next()
    }
}
```

### Usage in Routes

```go
// modules/user/routes.go
func (m *UserModule) RegisterRoutes(router fiber.Router, rbacMgr *rbac.Manager) {
    users := router.Group("/users")
    
    // Require authentication
    users.Use(middleware.RequireAuth())
    
    // Public routes (authenticated users)
    users.Get("/", middleware.RequirePermission(rbacMgr, "users.read"), m.controller.List)
    users.Get("/:id", middleware.RequirePermission(rbacMgr, "users.read"), m.controller.Get)
    
    // Admin routes
    users.Post("/", 
        middleware.RequirePermission(rbacMgr, "users.create"),
        m.controller.Create,
    )
    users.Put("/:id",
        middleware.RequirePermission(rbacMgr, "users.update"),
        m.controller.Update,
    )
    users.Delete("/:id",
        middleware.RequirePermission(rbacMgr, "users.delete"),
        m.controller.Delete,
    )
    
    // Super admin only
    users.Post("/:id/impersonate",
        middleware.RequireRole(rbacMgr, "super-admin"),
        m.controller.Impersonate,
    )
}
```

## Best Practices

### 1. Use Permission-Based Authorization

```go
// ✅ Good: Check permissions, not roles
if rbacManager.HasPermission(ctx, userID, "users.delete") {
    // Delete user
}

// ❌ Bad: Checking roles directly
if user.Role == "admin" {
    // Too rigid, hard to maintain
}
```

### 2. Follow Naming Conventions

```go
// Pattern: module.action
users.read
users.create
users.update
users.delete

products.read
products.create
products.publish
products.archive

admin.dashboard
admin.settings
admin.logs
```

### 3. Group Related Permissions

```go
// Create permission sets
userManagementPermissions := []string{
    "users.read",
    "users.create",
    "users.update",
    "users.delete",
}

hasUserManagement, _ := rbacManager.HasAllPermissions(ctx, userID, userManagementPermissions)
```

### 4. Cache Permission Checks

```go
type PermissionCache struct {
    cache cache.Cache
    rbac  *rbac.Manager
}

func (pc *PermissionCache) HasPermission(ctx context.Context, userID uint, permission string) (bool, error) {
    cacheKey := fmt.Sprintf("perm:%d:%s", userID, permission)
    
    // Check cache
    cached, err := pc.cache.Get(ctx, cacheKey)
    if err == nil && cached != nil {
        return cached.(bool), nil
    }
    
    // Check database
    has, err := pc.rbac.HasPermission(ctx, userID, permission)
    if err != nil {
        return false, err
    }
    
    // Cache result for 5 minutes
    pc.cache.Set(ctx, cacheKey, has, 5*time.Minute)
    
    return has, nil
}
```

### 5. Implement Permission Inheritance

```go
// Super admin automatically has all permissions
func (m *Manager) HasPermission(ctx context.Context, userID uint, permissionSlug string) (bool, error) {
    // Check if user is super admin
    isSuperAdmin, _ := m.HasRole(ctx, userID, "super-admin")
    if isSuperAdmin {
        return true, nil // Super admin has all permissions
    }
    
    // Regular permission check
    return m.checkPermission(ctx, userID, permissionSlug)
}
```

## Advanced Patterns

### Dynamic Permissions

```go
// Generate permissions dynamically based on ownership
func (s *ProductService) CanEditProduct(ctx context.Context, userID, productID uint) (bool, error) {
    // Check global permission
    hasGlobalEdit, _ := s.rbacManager.HasPermission(ctx, userID, "products.update")
    if hasGlobalEdit {
        return true, nil
    }
    
    // Check ownership
    product, err := s.productRepo.FindByID(ctx, productID)
    if err != nil {
        return false, err
    }
    
    return product.CreatedBy == userID, nil
}
```

### Permission Scoping

```go
type ScopedPermission struct {
    Permission string
    Scope      string // "own", "team", "all"
}

func (s *UserService) GetUsers(ctx context.Context, userID uint) ([]User, error) {
    hasAllAccess, _ := s.rbacManager.HasPermission(ctx, userID, "users.read.all")
    if hasAllAccess {
        // Return all users
        return s.userRepo.FindAll(ctx)
    }
    
    hasTeamAccess, _ := s.rbacManager.HasPermission(ctx, userID, "users.read.team")
    if hasTeamAccess {
        // Return team users only
        return s.userRepo.FindByTeam(ctx, user.TeamID)
    }
    
    // Return only own user
    user, err := s.userRepo.FindByID(ctx, userID)
    return []User{*user}, err
}
```

### Hierarchical Roles

```go
var roleHierarchy = map[string]int{
    "super-admin": 100,
    "admin":       80,
    "manager":     60,
    "editor":      40,
    "user":        20,
}

func (m *Manager) CanManageUser(ctx context.Context, managerID, targetUserID uint) (bool, error) {
    managerRoles, _ := m.GetUserRoles(ctx, managerID)
    targetRoles, _ := m.GetUserRoles(ctx, targetUserID)
    
    managerLevel := 0
    for _, role := range managerRoles {
        if level, ok := roleHierarchy[role.Slug]; ok && level > managerLevel {
            managerLevel = level
        }
    }
    
    targetLevel := 0
    for _, role := range targetRoles {
        if level, ok := roleHierarchy[role.Slug]; ok && level > targetLevel {
            targetLevel = level
        }
    }
    
    return managerLevel > targetLevel, nil
}
```

## Troubleshooting

### Common Issues

**Issue: Permission check is slow**

```go
// Solution: Add indexes
// migrations/add_rbac_indexes.sql
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_permissions_slug ON permissions(slug);
```

**Issue: Permissions not updating**

```go
// Solution: Clear cache after permission changes
func (s *AdminService) UpdateUserRoles(ctx context.Context, userID uint, roleIDs []uint) error {
    // Update roles
    err := s.updateRoles(ctx, userID, roleIDs)
    if err != nil {
        return err
    }
    
    // Clear permission cache
    cacheKey := fmt.Sprintf("permissions:user:%d", userID)
    s.cache.Delete(ctx, cacheKey)
    
    return nil
}
```

## Summary

NeonEx Framework authorization provides:

✅ **Flexible RBAC** - Roles and permissions  
✅ **Multiple assignments** - Users can have multiple roles  
✅ **Direct permissions** - Override role permissions  
✅ **Module organization** - Permissions grouped by module  
✅ **Middleware support** - Easy route protection  
✅ **Production-ready** - Optimized for performance

For more information:
- [Authentication](authentication.md)
- [JWT Security](jwt.md)
- [Middleware](../core-concepts/middleware.md)

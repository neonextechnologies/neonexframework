package rbac

import (
	"time"

	"gorm.io/gorm"
)

// Role represents a user role
type Role struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Slug        string         `gorm:"size:50;uniqueIndex;not null" json:"slug"`
	Description string         `gorm:"size:255" json:"description"`
	IsSystem    bool           `gorm:"default:false" json:"is_system"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

// Permission represents a permission
type Permission struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Slug        string         `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	Description string         `gorm:"size:255" json:"description"`
	Module      string         `gorm:"size:50;index" json:"module"`
	Category    string         `gorm:"size:50" json:"category"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Roles []Role `gorm:"many2many:role_permissions;" json:"roles,omitempty"`
}

// UserRole represents user-role relationship
type UserRole struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	RoleID    uint      `gorm:"index;not null" json:"role_id"`
	CreatedAt time.Time `json:"created_at"`

	Role Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

// UserPermission represents direct user permissions
type UserPermission struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	UserID       uint      `gorm:"index;not null" json:"user_id"`
	PermissionID uint      `gorm:"index;not null" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`

	Permission Permission `gorm:"foreignKey:PermissionID" json:"permission,omitempty"`
}

// TableName specifies the table name for Role
func (Role) TableName() string {
	return "roles"
}

// TableName specifies the table name for Permission
func (Permission) TableName() string {
	return "permissions"
}

// TableName specifies the table name for UserRole
func (UserRole) TableName() string {
	return "user_roles"
}

// TableName specifies the table name for UserPermission
func (UserPermission) TableName() string {
	return "user_permissions"
}

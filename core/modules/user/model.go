package user

import (
	"time"

	"neonexcore/pkg/rbac"

	"gorm.io/gorm"
)

// User model represents a user in the database
type User struct {
	ID                  uint           `gorm:"primarykey" json:"id"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
	Name                string         `gorm:"size:255;not null" json:"name"`
	Email               string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Username            string         `gorm:"size:50;uniqueIndex" json:"username"`
	Password            string         `gorm:"size:255;not null" json:"-"`
	Age                 int            `gorm:"default:0" json:"age"`
	Active              bool           `gorm:"default:true" json:"active"`
	IsActive            bool           `gorm:"default:true" json:"is_active"`
	IsEmailVerified     bool           `gorm:"default:false" json:"is_email_verified"`
	EmailVerifiedAt     *time.Time     `json:"email_verified_at,omitempty"`
	LastLoginAt         *time.Time     `json:"last_login_at,omitempty"`
	PasswordResetToken  *string        `gorm:"size:255" json:"-"`
	PasswordResetExpiry *time.Time     `json:"-"`
	APIKey              *string        `gorm:"size:255;uniqueIndex" json:"-"`

	// Relations
	Roles       []rbac.UserRole       `gorm:"foreignKey:UserID" json:"roles,omitempty"`
	Permissions []rbac.UserPermission `gorm:"foreignKey:UserID" json:"permissions,omitempty"`
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}

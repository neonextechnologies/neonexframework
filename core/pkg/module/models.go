package module

import (
	"time"

	"gorm.io/gorm"
)

// ModuleStatus represents the status of a module
type ModuleStatus string

const (
	ModuleStatusInstalled   ModuleStatus = "installed"
	ModuleStatusActive      ModuleStatus = "active"
	ModuleStatusInactive    ModuleStatus = "inactive"
	ModuleStatusUninstalled ModuleStatus = "uninstalled"
	ModuleStatusError       ModuleStatus = "error"
)

// Module represents a module record in database
type Module struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"uniqueIndex;not null" json:"name"`
	DisplayName string         `gorm:"not null" json:"display_name"`
	Description string         `json:"description"`
	Version     string         `gorm:"not null" json:"version"`
	Author      string         `json:"author"`
	Homepage    string         `json:"homepage"`
	Status      ModuleStatus   `gorm:"default:'installed'" json:"status"`
	Priority    int            `gorm:"default:100" json:"priority"`
	Path        string         `gorm:"not null" json:"path"`
	Config      string         `gorm:"type:text" json:"config"` // JSON string
	InstalledAt time.Time      `json:"installed_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Module model
func (Module) TableName() string {
	return "modules"
}

// ModuleDependency represents module dependency relationship
type ModuleDependency struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	ModuleID        uint           `gorm:"not null;index" json:"module_id"`
	DependsOnModule string         `gorm:"not null" json:"depends_on_module"`
	Version         string         `json:"version"` // Semantic version constraint (e.g., ">=1.0.0", "^2.0.0")
	Required        bool           `gorm:"default:true" json:"required"`
	CreatedAt       time.Time      `json:"created_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	Module Module `gorm:"foreignKey:ModuleID" json:"-"`
}

// TableName specifies the table name for ModuleDependency model
func (ModuleDependency) TableName() string {
	return "module_dependencies"
}

// ModuleMigration tracks module migration history
type ModuleMigration struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	ModuleID  uint      `gorm:"not null;index" json:"module_id"`
	Migration string    `gorm:"not null" json:"migration"`
	Batch     int       `gorm:"not null" json:"batch"`
	CreatedAt time.Time `json:"created_at"`

	Module Module `gorm:"foreignKey:ModuleID" json:"-"`
}

// TableName specifies the table name for ModuleMigration model
func (ModuleMigration) TableName() string {
	return "module_migrations"
}

// ModuleMetadata represents module.json structure
type ModuleMetadata struct {
	Name         string              `json:"name" validate:"required"`
	DisplayName  string              `json:"display_name" validate:"required"`
	Description  string              `json:"description"`
	Version      string              `json:"version" validate:"required,semver"`
	Author       string              `json:"author"`
	Homepage     string              `json:"homepage,omitempty"`
	License      string              `json:"license,omitempty"`
	Priority     int                 `json:"priority"`
	Dependencies []ModuleDependencyInfo `json:"dependencies,omitempty"`
	Routes       bool                `json:"routes"`
	Migrations   bool                `json:"migrations"`
	Seeders      bool                `json:"seeders"`
	Config       map[string]interface{} `json:"config,omitempty"`
}

// ModuleDependencyInfo represents dependency information in module.json
type ModuleDependencyInfo struct {
	Name     string `json:"name" validate:"required"`
	Version  string `json:"version" validate:"required"`
	Required bool   `json:"required"`
}

// ModuleInfo represents module information for API responses
type ModuleInfo struct {
	ID           uint                   `json:"id"`
	Name         string                 `json:"name"`
	DisplayName  string                 `json:"display_name"`
	Description  string                 `json:"description"`
	Version      string                 `json:"version"`
	Author       string                 `json:"author"`
	Homepage     string                 `json:"homepage"`
	Status       ModuleStatus           `json:"status"`
	Priority     int                    `json:"priority"`
	Path         string                 `json:"path"`
	Dependencies []ModuleDependencyInfo `json:"dependencies"`
	InstalledAt  time.Time              `json:"installed_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// ModuleListFilter represents filter options for listing modules
type ModuleListFilter struct {
	Status   ModuleStatus `json:"status"`
	Search   string       `json:"search"`
	Page     int          `json:"page"`
	Limit    int          `json:"limit"`
	OrderBy  string       `json:"order_by"`
	OrderDir string       `json:"order_dir"`
}

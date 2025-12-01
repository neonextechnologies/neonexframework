package admin

import (
	"time"
)

// DashboardStats represents overall system statistics
type DashboardStats struct {
	TotalUsers       int64     `json:"total_users"`
	ActiveUsers      int64     `json:"active_users"`
	TotalModules     int64     `json:"total_modules"`
	ActiveModules    int64     `json:"active_modules"`
	TotalRoles       int64     `json:"total_roles"`
	TotalPermissions int64     `json:"total_permissions"`
	SystemUptime     float64   `json:"system_uptime_seconds"`
	LastBackup       time.Time `json:"last_backup,omitempty"`
}

// UserStatistics represents user-related statistics
type UserStatistics struct {
	TotalUsers          int64              `json:"total_users"`
	ActiveUsers         int64              `json:"active_users"`
	InactiveUsers       int64              `json:"inactive_users"`
	NewUsersToday       int64              `json:"new_users_today"`
	NewUsersThisWeek    int64              `json:"new_users_this_week"`
	NewUsersThisMonth   int64              `json:"new_users_this_month"`
	UsersByRole         map[string]int64   `json:"users_by_role"`
	RecentLogins        []RecentLoginInfo  `json:"recent_logins"`
}

// RecentLoginInfo represents recent login information
type RecentLoginInfo struct {
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	LoginAt   time.Time `json:"login_at"`
	IPAddress string    `json:"ip_address"`
}

// ModuleStatistics represents module-related statistics
type ModuleStatistics struct {
	TotalModules     int64            `json:"total_modules"`
	ActiveModules    int64            `json:"active_modules"`
	InactiveModules  int64            `json:"inactive_modules"`
	ModulesByStatus  map[string]int64 `json:"modules_by_status"`
	RecentlyUpdated  []ModuleInfo     `json:"recently_updated"`
}

// ModuleInfo represents basic module information
type ModuleInfo struct {
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Version     string    `json:"version"`
	Status      string    `json:"status"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SystemHealth represents system health metrics
type SystemHealth struct {
	Status           string                 `json:"status"`
	DatabaseStatus   string                 `json:"database_status"`
	MemoryUsageMB    float64                `json:"memory_usage_mb"`
	GoroutineCount   int                    `json:"goroutine_count"`
	CPUUsage         float64                `json:"cpu_usage_percent"`
	DiskUsagePercent float64                `json:"disk_usage_percent"`
	UptimeSeconds    float64                `json:"uptime_seconds"`
	Details          map[string]interface{} `json:"details,omitempty"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	UserID      uint      `json:"user_id" gorm:"index"`
	Username    string    `json:"username"`
	Action      string    `json:"action" gorm:"index"`
	Resource    string    `json:"resource"`
	ResourceID  string    `json:"resource_id"`
	Description string    `json:"description"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	Status      string    `json:"status"`
	ErrorMsg    string    `json:"error_message,omitempty"`
	Metadata    string    `json:"metadata,omitempty" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at" gorm:"index"`
}

// ActivitySummary represents activity summary
type ActivitySummary struct {
	TotalActions      int64            `json:"total_actions"`
	ActionsByType     map[string]int64 `json:"actions_by_type"`
	ActionsByUser     map[string]int64 `json:"actions_by_user"`
	RecentActivities  []AuditLog       `json:"recent_activities"`
}

// SystemSettings represents global system settings
type SystemSettings struct {
	ID                uint      `json:"id" gorm:"primarykey"`
	Key               string    `json:"key" gorm:"uniqueIndex"`
	Value             string    `json:"value" gorm:"type:text"`
	Type              string    `json:"type"` // string, int, bool, json
	Category          string    `json:"category" gorm:"index"`
	Description       string    `json:"description"`
	IsPublic          bool      `json:"is_public"` // Can be accessed without admin rights
	UpdatedBy         uint      `json:"updated_by"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// BackupInfo represents backup information
type BackupInfo struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	Filename    string    `json:"filename"`
	Size        int64     `json:"size"`
	Type        string    `json:"type"` // full, incremental
	Status      string    `json:"status"` // success, failed, in_progress
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
	CreatedBy   uint      `json:"created_by"`
}

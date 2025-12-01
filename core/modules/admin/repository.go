package admin

import (
	"context"
	"time"

	"neonexcore/modules/user"
	"neonexcore/pkg/module"
	"neonexcore/pkg/rbac"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// GetDashboardStats retrieves overall dashboard statistics
func (r *Repository) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// Count total users
	r.db.WithContext(ctx).Model(&user.User{}).Count(&stats.TotalUsers)

	// Count active users
	r.db.WithContext(ctx).Model(&user.User{}).Where("is_active = ?", true).Count(&stats.ActiveUsers)

	// Count total modules
	r.db.WithContext(ctx).Model(&module.Module{}).Count(&stats.TotalModules)

	// Count active modules
	r.db.WithContext(ctx).Model(&module.Module{}).Where("is_active = ?", true).Count(&stats.ActiveModules)

	// Count roles
	r.db.WithContext(ctx).Model(&rbac.Role{}).Count(&stats.TotalRoles)

	// Count permissions
	r.db.WithContext(ctx).Model(&rbac.Permission{}).Count(&stats.TotalPermissions)

	return stats, nil
}

// GetUserStatistics retrieves detailed user statistics
func (r *Repository) GetUserStatistics(ctx context.Context) (*UserStatistics, error) {
	stats := &UserStatistics{
		UsersByRole: make(map[string]int64),
	}

	// Count total users
	r.db.WithContext(ctx).Model(&user.User{}).Count(&stats.TotalUsers)

	// Count active users
	r.db.WithContext(ctx).Model(&user.User{}).Where("is_active = ?", true).Count(&stats.ActiveUsers)

	// Count inactive users
	stats.InactiveUsers = stats.TotalUsers - stats.ActiveUsers

	// Count new users today
	today := time.Now().Truncate(24 * time.Hour)
	r.db.WithContext(ctx).Model(&user.User{}).
		Where("created_at >= ?", today).
		Count(&stats.NewUsersToday)

	// Count new users this week
	weekStart := time.Now().AddDate(0, 0, -7)
	r.db.WithContext(ctx).Model(&user.User{}).
		Where("created_at >= ?", weekStart).
		Count(&stats.NewUsersThisWeek)

	// Count new users this month
	monthStart := time.Now().AddDate(0, -1, 0)
	r.db.WithContext(ctx).Model(&user.User{}).
		Where("created_at >= ?", monthStart).
		Count(&stats.NewUsersThisMonth)

	// Get users by role
	var roleCounts []struct {
		Slug  string
		Count int64
	}
	r.db.WithContext(ctx).
		Table("user_roles").
		Select("roles.slug, COUNT(*) as count").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Group("roles.slug").
		Scan(&roleCounts)

	for _, rc := range roleCounts {
		stats.UsersByRole[rc.Slug] = rc.Count
	}

	// Get recent logins (last 10)
	var recentUsers []user.User
	r.db.WithContext(ctx).
		Model(&user.User{}).
		Where("last_login_at IS NOT NULL").
		Order("last_login_at DESC").
		Limit(10).
		Find(&recentUsers)

	for _, u := range recentUsers {
		if u.LastLoginAt != nil {
			stats.RecentLogins = append(stats.RecentLogins, RecentLoginInfo{
				UserID:   u.ID,
				Username: u.Username,
				Email:    u.Email,
				LoginAt:  *u.LastLoginAt,
			})
		}
	}

	return stats, nil
}

// GetModuleStatistics retrieves detailed module statistics
func (r *Repository) GetModuleStatistics(ctx context.Context) (*ModuleStatistics, error) {
	stats := &ModuleStatistics{
		ModulesByStatus: make(map[string]int64),
	}

	// Count total modules
	r.db.WithContext(ctx).Model(&module.Module{}).Count(&stats.TotalModules)

	// Count active modules
	r.db.WithContext(ctx).Model(&module.Module{}).
		Where("is_active = ?", true).
		Count(&stats.ActiveModules)

	// Count inactive modules
	stats.InactiveModules = stats.TotalModules - stats.ActiveModules

	// Get modules by status
	var statusCounts []struct {
		Status string
		Count  int64
	}
	r.db.WithContext(ctx).
		Model(&module.Module{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts)

	for _, sc := range statusCounts {
		stats.ModulesByStatus[sc.Status] = sc.Count
	}

	// Get recently updated modules
	var modules []module.Module
	r.db.WithContext(ctx).
		Order("updated_at DESC").
		Limit(5).
		Find(&modules)

	for _, m := range modules {
		stats.RecentlyUpdated = append(stats.RecentlyUpdated, ModuleInfo{
			Name:        m.Name,
			DisplayName: m.DisplayName,
			Version:     m.Version,
			Status:      m.Status,
			UpdatedAt:   m.UpdatedAt,
		})
	}

	return stats, nil
}

// CreateAuditLog creates a new audit log entry
func (r *Repository) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetAuditLogs retrieves audit logs with pagination
func (r *Repository) GetAuditLogs(ctx context.Context, page, limit int, filters map[string]interface{}) ([]AuditLog, int64, error) {
	var logs []AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&AuditLog{})

	// Apply filters
	if userID, ok := filters["user_id"].(uint); ok {
		query = query.Where("user_id = ?", userID)
	}
	if action, ok := filters["action"].(string); ok {
		query = query.Where("action = ?", action)
	}
	if resource, ok := filters["resource"].(string); ok {
		query = query.Where("resource = ?", resource)
	}
	if startDate, ok := filters["start_date"].(time.Time); ok {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate, ok := filters["end_date"].(time.Time); ok {
		query = query.Where("created_at <= ?", endDate)
	}

	// Count total
	query.Count(&total)

	// Get paginated results
	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error

	return logs, total, err
}

// GetActivitySummary retrieves activity summary
func (r *Repository) GetActivitySummary(ctx context.Context, days int) (*ActivitySummary, error) {
	summary := &ActivitySummary{
		ActionsByType: make(map[string]int64),
		ActionsByUser: make(map[string]int64),
	}

	startDate := time.Now().AddDate(0, 0, -days)

	// Count total actions
	r.db.WithContext(ctx).
		Model(&AuditLog{}).
		Where("created_at >= ?", startDate).
		Count(&summary.TotalActions)

	// Count actions by type
	var actionCounts []struct {
		Action string
		Count  int64
	}
	r.db.WithContext(ctx).
		Model(&AuditLog{}).
		Select("action, COUNT(*) as count").
		Where("created_at >= ?", startDate).
		Group("action").
		Scan(&actionCounts)

	for _, ac := range actionCounts {
		summary.ActionsByType[ac.Action] = ac.Count
	}

	// Count actions by user
	var userCounts []struct {
		Username string
		Count    int64
	}
	r.db.WithContext(ctx).
		Model(&AuditLog{}).
		Select("username, COUNT(*) as count").
		Where("created_at >= ?", startDate).
		Group("username").
		Scan(&userCounts)

	for _, uc := range userCounts {
		summary.ActionsByUser[uc.Username] = uc.Count
	}

	// Get recent activities
	r.db.WithContext(ctx).
		Model(&AuditLog{}).
		Order("created_at DESC").
		Limit(20).
		Find(&summary.RecentActivities)

	return summary, nil
}

// Settings operations
func (r *Repository) GetSetting(ctx context.Context, key string) (*SystemSettings, error) {
	var setting SystemSettings
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *Repository) GetSettingsByCategory(ctx context.Context, category string) ([]SystemSettings, error) {
	var settings []SystemSettings
	err := r.db.WithContext(ctx).Where("category = ?", category).Find(&settings).Error
	return settings, err
}

func (r *Repository) GetAllSettings(ctx context.Context) ([]SystemSettings, error) {
	var settings []SystemSettings
	err := r.db.WithContext(ctx).Find(&settings).Error
	return settings, err
}

func (r *Repository) CreateSetting(ctx context.Context, setting *SystemSettings) error {
	return r.db.WithContext(ctx).Create(setting).Error
}

func (r *Repository) UpdateSetting(ctx context.Context, key, value string, updatedBy uint) error {
	return r.db.WithContext(ctx).
		Model(&SystemSettings{}).
		Where("key = ?", key).
		Updates(map[string]interface{}{
			"value":      value,
			"updated_by": updatedBy,
			"updated_at": time.Now(),
		}).Error
}

func (r *Repository) DeleteSetting(ctx context.Context, key string) error {
	return r.db.WithContext(ctx).Where("key = ?", key).Delete(&SystemSettings{}).Error
}

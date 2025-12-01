package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"neonexcore/pkg/errors"

	"gorm.io/gorm"
)

type Service struct {
	repo      *Repository
	startTime time.Time
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo:      repo,
		startTime: time.Now(),
	}
}

// GetDashboard retrieves complete dashboard data
func (s *Service) GetDashboard(ctx context.Context) (map[string]interface{}, error) {
	stats, err := s.repo.GetDashboardStats(ctx)
	if err != nil {
		return nil, errors.NewAppError(errors.ErrCodeInternalError, "Failed to retrieve dashboard stats", err)
	}

	// Add system uptime
	stats.SystemUptime = time.Since(s.startTime).Seconds()

	// Get system health
	health := s.GetSystemHealth()

	// Get recent activity
	activity, err := s.repo.GetActivitySummary(ctx, 7) // Last 7 days
	if err != nil {
		// Don't fail the whole request if activity fails
		activity = &ActivitySummary{}
	}

	return map[string]interface{}{
		"stats":    stats,
		"health":   health,
		"activity": activity,
	}, nil
}

// GetStats retrieves detailed system statistics
func (s *Service) GetStats(ctx context.Context) (map[string]interface{}, error) {
	userStats, err := s.repo.GetUserStatistics(ctx)
	if err != nil {
		return nil, errors.NewAppError(errors.ErrCodeInternalError, "Failed to retrieve user statistics", err)
	}

	moduleStats, err := s.repo.GetModuleStatistics(ctx)
	if err != nil {
		return nil, errors.NewAppError(errors.ErrCodeInternalError, "Failed to retrieve module statistics", err)
	}

	return map[string]interface{}{
		"users":   userStats,
		"modules": moduleStats,
		"system":  s.GetSystemHealth(),
	}, nil
}

// GetSystemHealth retrieves current system health metrics
func (s *Service) GetSystemHealth() *SystemHealth {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	health := &SystemHealth{
		Status:         "healthy",
		DatabaseStatus: "connected",
		MemoryUsageMB:  float64(m.Alloc) / 1024 / 1024,
		GoroutineCount: runtime.NumGoroutine(),
		UptimeSeconds:  time.Since(s.startTime).Seconds(),
		Details:        make(map[string]interface{}),
	}

	// Add detailed memory stats
	health.Details["sys_mb"] = float64(m.Sys) / 1024 / 1024
	health.Details["num_gc"] = m.NumGC
	health.Details["go_version"] = runtime.Version()
	health.Details["num_cpu"] = runtime.NumCPU()

	// Determine overall health status
	if health.MemoryUsageMB > 1000 || health.GoroutineCount > 1000 {
		health.Status = "degraded"
	}
	if health.MemoryUsageMB > 2000 || health.GoroutineCount > 10000 {
		health.Status = "critical"
	}

	return health
}

// LogActivity creates an audit log entry
func (s *Service) LogActivity(ctx context.Context, log *AuditLog) error {
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	if log.Status == "" {
		log.Status = "success"
	}

	return s.repo.CreateAuditLog(ctx, log)
}

// GetAuditLogs retrieves audit logs with pagination and filters
func (s *Service) GetAuditLogs(ctx context.Context, page, limit int, filters map[string]interface{}) ([]AuditLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	return s.repo.GetAuditLogs(ctx, page, limit, filters)
}

// GetActivitySummary retrieves activity summary for specified days
func (s *Service) GetActivitySummary(ctx context.Context, days int) (*ActivitySummary, error) {
	if days < 1 {
		days = 7
	}
	if days > 365 {
		days = 365
	}

	return s.repo.GetActivitySummary(ctx, days)
}

// Settings management
func (s *Service) GetSetting(ctx context.Context, key string) (*SystemSettings, error) {
	setting, err := s.repo.GetSetting(ctx, key)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewAppError(errors.ErrCodeNotFound, "Setting not found", err)
		}
		return nil, errors.NewAppError(errors.ErrCodeInternalError, "Failed to retrieve setting", err)
	}
	return setting, nil
}

func (s *Service) GetSettingsByCategory(ctx context.Context, category string) ([]SystemSettings, error) {
	settings, err := s.repo.GetSettingsByCategory(ctx, category)
	if err != nil {
		return nil, errors.NewAppError(errors.ErrCodeInternalError, "Failed to retrieve settings", err)
	}
	return settings, nil
}

func (s *Service) GetAllSettings(ctx context.Context, includePrivate bool) ([]SystemSettings, error) {
	settings, err := s.repo.GetAllSettings(ctx)
	if err != nil {
		return nil, errors.NewAppError(errors.ErrCodeInternalError, "Failed to retrieve settings", err)
	}

	// Filter out private settings if requested
	if !includePrivate {
		var publicSettings []SystemSettings
		for _, setting := range settings {
			if setting.IsPublic {
				publicSettings = append(publicSettings, setting)
			}
		}
		return publicSettings, nil
	}

	return settings, nil
}

func (s *Service) CreateSetting(ctx context.Context, setting *SystemSettings) error {
	// Check if setting already exists
	existing, _ := s.repo.GetSetting(ctx, setting.Key)
	if existing != nil {
		return errors.NewAppError(errors.ErrCodeConflict, "Setting already exists", nil)
	}

	if err := s.repo.CreateSetting(ctx, setting); err != nil {
		return errors.NewAppError(errors.ErrCodeInternalError, "Failed to create setting", err)
	}

	return nil
}

func (s *Service) UpdateSetting(ctx context.Context, key, value string, updatedBy uint) error {
	// Verify setting exists
	_, err := s.repo.GetSetting(ctx, key)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewAppError(errors.ErrCodeNotFound, "Setting not found", err)
		}
		return errors.NewAppError(errors.ErrCodeInternalError, "Failed to retrieve setting", err)
	}

	if err := s.repo.UpdateSetting(ctx, key, value, updatedBy); err != nil {
		return errors.NewAppError(errors.ErrCodeInternalError, "Failed to update setting", err)
	}

	return nil
}

func (s *Service) DeleteSetting(ctx context.Context, key string) error {
	if err := s.repo.DeleteSetting(ctx, key); err != nil {
		return errors.NewAppError(errors.ErrCodeInternalError, "Failed to delete setting", err)
	}
	return nil
}

// GetSettingValue retrieves and parses setting value by type
func (s *Service) GetSettingValue(ctx context.Context, key string) (interface{}, error) {
	setting, err := s.GetSetting(ctx, key)
	if err != nil {
		return nil, err
	}

	switch setting.Type {
	case "int":
		var value int
		if err := json.Unmarshal([]byte(setting.Value), &value); err != nil {
			return setting.Value, nil // Return as string if parsing fails
		}
		return value, nil
	case "bool":
		var value bool
		if err := json.Unmarshal([]byte(setting.Value), &value); err != nil {
			return setting.Value, nil
		}
		return value, nil
	case "json":
		var value interface{}
		if err := json.Unmarshal([]byte(setting.Value), &value); err != nil {
			return setting.Value, nil
		}
		return value, nil
	default:
		return setting.Value, nil
	}
}

// SetSettingValue sets a setting value with automatic type conversion
func (s *Service) SetSettingValue(ctx context.Context, key string, value interface{}, updatedBy uint) error {
	var stringValue string

	switch v := value.(type) {
	case string:
		stringValue = v
	case int, int64, float64, bool:
		bytes, err := json.Marshal(v)
		if err != nil {
			return errors.NewAppError(errors.ErrCodeInvalidInput, "Failed to marshal value", err)
		}
		stringValue = string(bytes)
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return errors.NewAppError(errors.ErrCodeInvalidInput, "Failed to marshal value", err)
		}
		stringValue = string(bytes)
	}

	return s.UpdateSetting(ctx, key, stringValue, updatedBy)
}

// Helper to log admin actions
func (s *Service) LogAdminAction(ctx context.Context, userID uint, username, action, resource, resourceID string, status string, metadata interface{}) error {
	metadataStr := ""
	if metadata != nil {
		bytes, _ := json.Marshal(metadata)
		metadataStr = string(bytes)
	}

	log := &AuditLog{
		UserID:      userID,
		Username:    username,
		Action:      action,
		Resource:    resource,
		ResourceID:  fmt.Sprintf("%v", resourceID),
		Description: fmt.Sprintf("%s %s %s", username, action, resource),
		Status:      status,
		Metadata:    metadataStr,
		CreatedAt:   time.Now(),
	}

	return s.LogActivity(ctx, log)
}

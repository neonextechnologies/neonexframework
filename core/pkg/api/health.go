package api

import (
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthCheck represents a health check response
type HealthCheck struct {
	Status    HealthStatus           `json:"status"`
	Version   string                 `json:"version"`
	Timestamp int64                  `json:"timestamp"`
	Uptime    float64                `json:"uptime_seconds"`
	Checks    map[string]CheckResult `json:"checks"`
}

// CheckResult represents the result of a single health check
type CheckResult struct {
	Status  HealthStatus `json:"status"`
	Message string       `json:"message,omitempty"`
	Details interface{}  `json:"details,omitempty"`
}

// HealthChecker performs health checks
type HealthChecker struct {
	startTime time.Time
	version   string
	db        *gorm.DB
	checks    map[string]func() CheckResult
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(version string, db *gorm.DB) *HealthChecker {
	hc := &HealthChecker{
		startTime: time.Now(),
		version:   version,
		db:        db,
		checks:    make(map[string]func() CheckResult),
	}

	// Register default checks
	hc.RegisterCheck("database", hc.checkDatabase)
	hc.RegisterCheck("memory", hc.checkMemory)
	hc.RegisterCheck("goroutines", hc.checkGoroutines)

	return hc
}

// RegisterCheck registers a custom health check
func (hc *HealthChecker) RegisterCheck(name string, check func() CheckResult) {
	hc.checks[name] = check
}

// Check performs all health checks
func (hc *HealthChecker) Check() HealthCheck {
	checks := make(map[string]CheckResult)
	overallStatus := HealthStatusHealthy

	for name, check := range hc.checks {
		result := check()
		checks[name] = result

		// Update overall status
		if result.Status == HealthStatusUnhealthy {
			overallStatus = HealthStatusUnhealthy
		} else if result.Status == HealthStatusDegraded && overallStatus == HealthStatusHealthy {
			overallStatus = HealthStatusDegraded
		}
	}

	return HealthCheck{
		Status:    overallStatus,
		Version:   hc.version,
		Timestamp: time.Now().Unix(),
		Uptime:    time.Since(hc.startTime).Seconds(),
		Checks:    checks,
	}
}

// checkDatabase checks database connectivity
func (hc *HealthChecker) checkDatabase() CheckResult {
	if hc.db == nil {
		return CheckResult{
			Status:  HealthStatusUnhealthy,
			Message: "Database not configured",
		}
	}

	sqlDB, err := hc.db.DB()
	if err != nil {
		return CheckResult{
			Status:  HealthStatusUnhealthy,
			Message: "Failed to get database connection",
		}
	}

	if err := sqlDB.Ping(); err != nil {
		return CheckResult{
			Status:  HealthStatusUnhealthy,
			Message: "Database ping failed",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		}
	}

	stats := sqlDB.Stats()
	return CheckResult{
		Status:  HealthStatusHealthy,
		Message: "Database is healthy",
		Details: map[string]interface{}{
			"open_connections": stats.OpenConnections,
			"in_use":           stats.InUse,
			"idle":             stats.Idle,
		},
	}
}

// checkMemory checks memory usage
func (hc *HealthChecker) checkMemory() CheckResult {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	allocMB := float64(m.Alloc) / 1024 / 1024
	sysMB := float64(m.Sys) / 1024 / 1024

	status := HealthStatusHealthy
	message := "Memory usage is normal"

	if allocMB > 1000 { // More than 1GB
		status = HealthStatusDegraded
		message = "Memory usage is high"
	}

	if allocMB > 2000 { // More than 2GB
		status = HealthStatusUnhealthy
		message = "Memory usage is critical"
	}

	return CheckResult{
		Status:  status,
		Message: message,
		Details: map[string]interface{}{
			"alloc_mb":      allocMB,
			"sys_mb":        sysMB,
			"num_gc":        m.NumGC,
			"goroutines":    runtime.NumGoroutine(),
		},
	}
}

// checkGoroutines checks goroutine count
func (hc *HealthChecker) checkGoroutines() CheckResult {
	count := runtime.NumGoroutine()

	status := HealthStatusHealthy
	message := "Goroutine count is normal"

	if count > 1000 {
		status = HealthStatusDegraded
		message = "Goroutine count is high"
	}

	if count > 10000 {
		status = HealthStatusUnhealthy
		message = "Goroutine count is critical"
	}

	return CheckResult{
		Status:  status,
		Message: message,
		Details: map[string]interface{}{
			"count": count,
		},
	}
}

// HealthCheckHandler creates a health check endpoint handler
func HealthCheckHandler(checker *HealthChecker) fiber.Handler {
	return func(c *fiber.Ctx) error {
		health := checker.Check()

		statusCode := fiber.StatusOK
		if health.Status == HealthStatusUnhealthy {
			statusCode = fiber.StatusServiceUnavailable
		} else if health.Status == HealthStatusDegraded {
			statusCode = fiber.StatusOK // Still return 200 for degraded
		}

		return c.Status(statusCode).JSON(health)
	}
}

// ReadinessHandler creates a readiness check endpoint (simpler than health)
func ReadinessHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check database connectivity
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"ready": false,
			})
		}

		return c.JSON(fiber.Map{
			"ready": true,
		})
	}
}

// LivenessHandler creates a liveness check endpoint (always returns OK if app is running)
func LivenessHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"alive": true,
		})
	}
}

// SetupHealthRoutes sets up health check routes
func SetupHealthRoutes(app *fiber.App, checker *HealthChecker, db *gorm.DB) {
	app.Get("/health", HealthCheckHandler(checker))
	app.Get("/health/ready", ReadinessHandler(db))
	app.Get("/health/live", LivenessHandler())
}

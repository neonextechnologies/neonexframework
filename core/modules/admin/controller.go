package admin

import (
	"strconv"
	"time"

	"neonexcore/pkg/api"

	"github.com/gofiber/fiber/v2"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

// GetDashboard retrieves complete dashboard data
// @Summary Get admin dashboard
// @Description Get complete admin dashboard with stats, health, and activity
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} api.Response{data=map[string]interface{}}
// @Failure 401 {object} api.Response
// @Failure 403 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /admin/dashboard [get]
func (c *Controller) GetDashboard(ctx *fiber.Ctx) error {
	dashboard, err := c.service.GetDashboard(ctx.Context())
	if err != nil {
		return api.InternalError(ctx, err.Error())
	}

	return api.Success(ctx, dashboard)
}

// GetStats retrieves detailed system statistics
// @Summary Get system statistics
// @Description Get detailed statistics about users, modules, and system
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} api.Response{data=map[string]interface{}}
// @Failure 401 {object} api.Response
// @Failure 403 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /admin/stats [get]
func (c *Controller) GetStats(ctx *fiber.Ctx) error {
	stats, err := c.service.GetStats(ctx.Context())
	if err != nil {
		return api.InternalError(ctx, err.Error())
	}

	return api.Success(ctx, stats)
}

// GetUserStats retrieves user statistics
// @Summary Get user statistics
// @Description Get detailed statistics about users
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} api.Response{data=UserStatistics}
// @Failure 500 {object} api.Response
// @Router /admin/stats/users [get]
func (c *Controller) GetUserStats(ctx *fiber.Ctx) error {
	stats, err := c.service.repo.GetUserStatistics(ctx.Context())
	if err != nil {
		return api.InternalError(ctx, err.Error())
	}

	return api.Success(ctx, stats)
}

// GetModuleStats retrieves module statistics
// @Summary Get module statistics
// @Description Get detailed statistics about modules
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} api.Response{data=ModuleStatistics}
// @Failure 500 {object} api.Response
// @Router /admin/stats/modules [get]
func (c *Controller) GetModuleStats(ctx *fiber.Ctx) error {
	stats, err := c.service.repo.GetModuleStatistics(ctx.Context())
	if err != nil {
		return api.InternalError(ctx, err.Error())
	}

	return api.Success(ctx, stats)
}

// GetSystemHealth retrieves system health metrics
// @Summary Get system health
// @Description Get current system health and performance metrics
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} api.Response{data=SystemHealth}
// @Failure 500 {object} api.Response
// @Router /admin/health [get]
func (c *Controller) GetSystemHealth(ctx *fiber.Ctx) error {
	health := c.service.GetSystemHealth()
	return api.Success(ctx, health)
}

// GetAuditLogs retrieves audit logs with pagination
// @Summary Get audit logs
// @Description Get audit logs with pagination and filters
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param user_id query int false "Filter by user ID"
// @Param action query string false "Filter by action"
// @Param resource query string false "Filter by resource"
// @Success 200 {object} api.Response{data=[]AuditLog}
// @Failure 500 {object} api.Response
// @Router /admin/audit-logs [get]
func (c *Controller) GetAuditLogs(ctx *fiber.Ctx) error {
	pagination := api.GetPagination(ctx)
	
	// Build filters
	filters := make(map[string]interface{})
	if userID := ctx.QueryInt("user_id", 0); userID > 0 {
		filters["user_id"] = uint(userID)
	}
	if action := ctx.Query("action"); action != "" {
		filters["action"] = action
	}
	if resource := ctx.Query("resource"); resource != "" {
		filters["resource"] = resource
	}

	logs, total, err := c.service.GetAuditLogs(ctx.Context(), pagination.Page, pagination.Limit, filters)
	if err != nil {
		return api.InternalError(ctx, err.Error())
	}

	meta := api.CalculateMeta(pagination.Page, pagination.Limit, int(total))
	return api.Paginated(ctx, logs, meta)
}

// GetActivitySummary retrieves activity summary
// @Summary Get activity summary
// @Description Get activity summary for specified number of days
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param days query int false "Number of days" default(7)
// @Success 200 {object} api.Response{data=ActivitySummary}
// @Failure 500 {object} api.Response
// @Router /admin/activity [get]
func (c *Controller) GetActivitySummary(ctx *fiber.Ctx) error {
	days := ctx.QueryInt("days", 7)
	
	summary, err := c.service.GetActivitySummary(ctx.Context(), days)
	if err != nil {
		return api.InternalError(ctx, err.Error())
	}

	return api.Success(ctx, summary)
}

// GetSettings retrieves all settings or by category
// @Summary Get system settings
// @Description Get all system settings or filter by category
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param category query string false "Filter by category"
// @Success 200 {object} api.Response{data=[]SystemSettings}
// @Failure 500 {object} api.Response
// @Router /admin/settings [get]
func (c *Controller) GetSettings(ctx *fiber.Ctx) error {
	category := ctx.Query("category")
	
	var settings []SystemSettings
	var err error

	if category != "" {
		settings, err = c.service.GetSettingsByCategory(ctx.Context(), category)
	} else {
		settings, err = c.service.GetAllSettings(ctx.Context(), true)
	}

	if err != nil {
		return api.InternalError(ctx, err.Error())
	}

	return api.Success(ctx, settings)
}

// GetSetting retrieves a specific setting
// @Summary Get a setting
// @Description Get a specific system setting by key
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param key path string true "Setting key"
// @Success 200 {object} api.Response{data=SystemSettings}
// @Failure 404 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /admin/settings/{key} [get]
func (c *Controller) GetSetting(ctx *fiber.Ctx) error {
	key := ctx.Params("key")
	
	setting, err := c.service.GetSetting(ctx.Context(), key)
	if err != nil {
		return api.NotFound(ctx, "Setting not found")
	}

	return api.Success(ctx, setting)
}

// CreateSetting creates a new setting
// @Summary Create a setting
// @Description Create a new system setting
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param setting body SystemSettings true "Setting data"
// @Success 201 {object} api.Response{data=SystemSettings}
// @Failure 400 {object} api.Response
// @Failure 409 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /admin/settings [post]
func (c *Controller) CreateSetting(ctx *fiber.Ctx) error {
	var setting SystemSettings
	if err := ctx.BodyParser(&setting); err != nil {
		return api.BadRequest(ctx, "Invalid request body")
	}

	// Get current user ID
	userID := ctx.Locals("user_id")
	if userID != nil {
		if uid, ok := userID.(uint); ok {
			setting.UpdatedBy = uid
		}
	}

	if err := c.service.CreateSetting(ctx.Context(), &setting); err != nil {
		return api.InternalError(ctx, err.Error())
	}

	return api.Created(ctx, setting)
}

// UpdateSetting updates an existing setting
// @Summary Update a setting
// @Description Update an existing system setting
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param key path string true "Setting key"
// @Param setting body map[string]string true "Setting value"
// @Success 200 {object} api.Response{data=SystemSettings}
// @Failure 400 {object} api.Response
// @Failure 404 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /admin/settings/{key} [put]
func (c *Controller) UpdateSetting(ctx *fiber.Ctx) error {
	key := ctx.Params("key")
	
	var body struct {
		Value string `json:"value"`
	}
	if err := ctx.BodyParser(&body); err != nil {
		return api.BadRequest(ctx, "Invalid request body")
	}

	// Get current user ID
	var userID uint
	if uid := ctx.Locals("user_id"); uid != nil {
		if id, ok := uid.(uint); ok {
			userID = id
		}
	}

	if err := c.service.UpdateSetting(ctx.Context(), key, body.Value, userID); err != nil {
		return api.InternalError(ctx, err.Error())
	}

	// Return updated setting
	setting, _ := c.service.GetSetting(ctx.Context(), key)
	return api.Success(ctx, setting)
}

// DeleteSetting deletes a setting
// @Summary Delete a setting
// @Description Delete a system setting
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param key path string true "Setting key"
// @Success 204 "No Content"
// @Failure 404 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /admin/settings/{key} [delete]
func (c *Controller) DeleteSetting(ctx *fiber.Ctx) error {
	key := ctx.Params("key")
	
	if err := c.service.DeleteSetting(ctx.Context(), key); err != nil {
		return api.InternalError(ctx, err.Error())
	}

	return api.NoContent(ctx)
}

// Helper to get user info from context
func getUserInfo(ctx *fiber.Ctx) (uint, string) {
	var userID uint
	var username string

	if uid := ctx.Locals("user_id"); uid != nil {
		if id, ok := uid.(uint); ok {
			userID = id
		}
	}
	if uname := ctx.Locals("username"); uname != nil {
		if name, ok := uname.(string); ok {
			username = name
		}
	}

	return userID, username
}

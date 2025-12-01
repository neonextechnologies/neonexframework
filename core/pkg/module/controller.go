package module

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"neonexcore/pkg/errors"
)

// ModuleController handles module management HTTP endpoints
type ModuleController struct {
	manager *ModuleManager
}

// NewModuleController creates a new module controller
func NewModuleController(manager *ModuleManager) *ModuleController {
	return &ModuleController{
		manager: manager,
	}
}

// ListModules handles GET /api/v1/modules
func (c *ModuleController) ListModules(ctx *fiber.Ctx) error {
	filter := ModuleListFilter{
		Status:   ModuleStatus(ctx.Query("status")),
		Search:   ctx.Query("search"),
		Page:     ctx.QueryInt("page", 1),
		Limit:    ctx.QueryInt("limit", 10),
		OrderBy:  ctx.Query("order_by", "priority"),
		OrderDir: ctx.Query("order_dir", "ASC"),
	}

	modules, total, err := c.manager.ListModules(ctx.Context(), filter)
	if err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    modules,
		"meta": fiber.Map{
			"total":       total,
			"page":        filter.Page,
			"limit":       filter.Limit,
			"total_pages": (total + int64(filter.Limit) - 1) / int64(filter.Limit),
		},
	})
}

// GetModule handles GET /api/v1/modules/:name
func (c *ModuleController) GetModule(ctx *fiber.Ctx) error {
	name := ctx.Params("name")

	module, err := c.manager.GetModule(ctx.Context(), name)
	if err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    module,
	})
}

// InstallModule handles POST /api/v1/modules/install
func (c *ModuleController) InstallModule(ctx *fiber.Ctx) error {
	var req struct {
		Path string `json:"path" validate:"required"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return errors.NewBadRequest("Invalid request body")
	}

	module, err := c.manager.Install(ctx.Context(), req.Path)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Module installed successfully",
		"data":    module,
	})
}

// UninstallModule handles DELETE /api/v1/modules/:name
func (c *ModuleController) UninstallModule(ctx *fiber.Ctx) error {
	name := ctx.Params("name")
	force := ctx.QueryBool("force", false)

	if err := c.manager.Uninstall(ctx.Context(), name, force); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "Module uninstalled successfully",
	})
}

// ActivateModule handles POST /api/v1/modules/:name/activate
func (c *ModuleController) ActivateModule(ctx *fiber.Ctx) error {
	name := ctx.Params("name")

	if err := c.manager.Activate(ctx.Context(), name); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "Module activated successfully",
	})
}

// DeactivateModule handles POST /api/v1/modules/:name/deactivate
func (c *ModuleController) DeactivateModule(ctx *fiber.Ctx) error {
	name := ctx.Params("name")

	if err := c.manager.Deactivate(ctx.Context(), name); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "Module deactivated successfully",
	})
}

// UpdateModule handles PUT /api/v1/modules/:name
func (c *ModuleController) UpdateModule(ctx *fiber.Ctx) error {
	name := ctx.Params("name")

	var req struct {
		Path string `json:"path" validate:"required"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return errors.NewBadRequest("Invalid request body")
	}

	if err := c.manager.Update(ctx.Context(), name, req.Path); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "Module updated successfully",
	})
}

// GetModuleConfig handles GET /api/v1/modules/:name/config
func (c *ModuleController) GetModuleConfig(ctx *fiber.Ctx) error {
	name := ctx.Params("name")

	module, err := c.manager.GetModule(ctx.Context(), name)
	if err != nil {
		return err
	}

	// Parse config from JSON string
	config, err := c.manager.repo.ParseConfig(module.Path)
	if err != nil {
		return errors.NewInternal("Failed to parse module config")
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    config,
	})
}

// UpdateModuleConfig handles PUT /api/v1/modules/:name/config
func (c *ModuleController) UpdateModuleConfig(ctx *fiber.Ctx) error {
	name := ctx.Params("name")

	module, err := c.manager.GetModule(ctx.Context(), name)
	if err != nil {
		return err
	}

	var config map[string]interface{}
	if err := ctx.BodyParser(&config); err != nil {
		return errors.NewBadRequest("Invalid request body")
	}

	if err := c.manager.repo.SaveConfig(ctx.Context(), module.ID, config); err != nil {
		return errors.NewInternal("Failed to save module config")
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "Module config updated successfully",
		"data":    config,
	})
}

// GetModuleStats handles GET /api/v1/modules/stats
func (c *ModuleController) GetModuleStats(ctx *fiber.Ctx) error {
	// Get counts by status
	installed, _, _ := c.manager.repo.List(ctx.Context(), ModuleListFilter{
		Status: ModuleStatusInstalled,
		Limit:  1,
	})
	active, _, _ := c.manager.repo.List(ctx.Context(), ModuleListFilter{
		Status: ModuleStatusActive,
		Limit:  1,
	})
	inactive, _, _ := c.manager.repo.List(ctx.Context(), ModuleListFilter{
		Status: ModuleStatusInactive,
		Limit:  1,
	})

	return ctx.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"total":    len(installed) + len(active) + len(inactive),
			"active":   len(active),
			"inactive": len(inactive),
			"error":    0, // TODO: implement error status
		},
	})
}

// RegisterRoutes registers module routes
func (c *ModuleController) RegisterRoutes(router fiber.Router) {
	modules := router.Group("/modules")

	// List and stats
	modules.Get("/", c.ListModules)
	modules.Get("/stats", c.GetModuleStats)

	// Single module operations
	modules.Get("/:name", c.GetModule)
	modules.Post("/install", c.InstallModule)
	modules.Delete("/:name", c.UninstallModule)
	modules.Put("/:name", c.UpdateModule)

	// Module status
	modules.Post("/:name/activate", c.ActivateModule)
	modules.Post("/:name/deactivate", c.DeactivateModule)

	// Module config
	modules.Get("/:name/config", c.GetModuleConfig)
	modules.Put("/:name/config", c.UpdateModuleConfig)
}

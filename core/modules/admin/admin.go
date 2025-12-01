package admin

import (
	"neonexcore/internal/config"
	"neonexcore/internal/core"

	"github.com/gofiber/fiber/v2"
)

type AdminModule struct{}

func New() *AdminModule {
	return &AdminModule{}
}

func (m *AdminModule) Name() string {
	return "admin"
}

func (m *AdminModule) Init() {}

func (m *AdminModule) RegisterServices(c *core.Container) {
	RegisterDependencies(c, config.DB.GetDB())
}

func (m *AdminModule) Routes(router fiber.Router, c *core.Container) {
	SetupRoutes(router, c)
}

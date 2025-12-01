package product

import (
	"github.com/gofiber/fiber/v2"
	"neonexcore/internal/core"
)

type Module struct{}

func New() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "product"
}

func (m *Module) Init() {
	// Module initialization logic
}

func (m *Module) Routes(app *fiber.App, c *core.Container) {
	RegisterRoutes(app, c)
}

func (m *Module) RegisterServices(c *core.Container) {
	RegisterDependencies(c)
}

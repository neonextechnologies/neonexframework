package user

import (
	"neonexcore/internal/config"
	"neonexcore/internal/core"

	"github.com/gofiber/fiber/v2"
)

type UserModule struct{}

func New() *UserModule {
	return &UserModule{}
}

func (m *UserModule) Name() string {
	return "user"
}

func (m *UserModule) Init() {}

func (m *UserModule) RegisterServices(c *core.Container) {
	RegisterDependencies(c, config.DB.GetDB())
}

func (m *UserModule) Routes(app fiber.Router, c *core.Container) {
	SetupRoutes(app, c)
}

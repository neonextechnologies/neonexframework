package product

import (
	"neonexcore/internal/config"
	"neonexcore/internal/core"
)

func RegisterDependencies(container *core.Container) {
	// Register Repository
	container.Singleton("product.Repository", func(c *core.Container) interface{} {
		return NewRepository(config.DB.GetDB())
	})

	// Register Service
	container.Singleton("product.Service", func(c *core.Container) interface{} {
		repo := c.Resolve("product.Repository").(*Repository)
		return NewService(repo)
	})

	// Register Controller
	container.Singleton("product.Controller", func(c *core.Container) interface{} {
		service := c.Resolve("product.Service").(*Service)
		return NewController(service)
	})
}

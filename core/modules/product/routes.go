package product

import (
	"neonexcore/internal/core"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App, container *core.Container) {
	ctrl := container.Resolve("product.Controller").(*Controller)

	group := app.Group("/product")
	group.Get("/", ctrl.GetAll)
	group.Get("/:id", ctrl.GetByID)
	group.Post("/", ctrl.Create)
	group.Put("/:id", ctrl.Update)
	group.Delete("/:id", ctrl.Delete)
	group.Get("/search", ctrl.Search)
}

package product

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) GetAll(ctx *fiber.Ctx) error {
	entities, err := c.service.GetAll(ctx.Context())
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(entities)
}

func (c *Controller) GetByID(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	entity, err := c.service.GetByID(ctx.Context(), uint(id))
	if err != nil {
		return ctx.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.JSON(entity)
}

func (c *Controller) Create(ctx *fiber.Ctx) error {
	var entity Product
	if err := ctx.BodyParser(&entity); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := c.service.Create(ctx.Context(), &entity); err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(201).JSON(entity)
}

func (c *Controller) Update(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var entity Product
	if err := ctx.BodyParser(&entity); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := c.service.Update(ctx.Context(), uint(id), &entity); err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.JSON(entity)
}

func (c *Controller) Delete(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	if err := c.service.Delete(ctx.Context(), uint(id)); err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.SendStatus(204)
}

func (c *Controller) Search(ctx *fiber.Ctx) error {
	query := ctx.Query("q")
	entities, err := c.service.Search(ctx.Context(), query)
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(entities)
}

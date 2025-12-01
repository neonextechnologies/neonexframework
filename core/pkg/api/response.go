package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// Response represents a standard API response
type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Errors    interface{} `json:"errors,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// Meta represents metadata for paginated responses
type Meta struct {
	Page         int   `json:"page,omitempty"`
	Limit        int   `json:"limit,omitempty"`
	Total        int64 `json:"total,omitempty"`
	TotalPages   int   `json:"total_pages,omitempty"`
	HasNextPage  bool  `json:"has_next_page,omitempty"`
	HasPrevPage  bool  `json:"has_prev_page,omitempty"`
	NextPage     *int  `json:"next_page,omitempty"`
	PrevPage     *int  `json:"prev_page,omitempty"`
}

// PaginationParams represents pagination query parameters
type PaginationParams struct {
	Page  int `query:"page" validate:"omitempty,min=1"`
	Limit int `query:"limit" validate:"omitempty,min=1,max=100"`
}

// GetPagination extracts pagination params from request
func GetPagination(c *fiber.Ctx) PaginationParams {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return PaginationParams{
		Page:  page,
		Limit: limit,
	}
}

// CalculateMeta calculates pagination metadata
func CalculateMeta(page, limit int, total int64) *Meta {
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	hasNextPage := page < totalPages
	hasPrevPage := page > 1

	meta := &Meta{
		Page:        page,
		Limit:       limit,
		Total:       total,
		TotalPages:  totalPages,
		HasNextPage: hasNextPage,
		HasPrevPage: hasPrevPage,
	}

	if hasNextPage {
		nextPage := page + 1
		meta.NextPage = &nextPage
	}

	if hasPrevPage {
		prevPage := page - 1
		meta.PrevPage = &prevPage
	}

	return meta
}

// Success sends a successful response
func Success(c *fiber.Ctx, data interface{}) error {
	return c.JSON(Response{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// SuccessWithMessage sends a successful response with a message
func SuccessWithMessage(c *fiber.Ctx, message string, data interface{}) error {
	return c.JSON(Response{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// Created sends a 201 Created response
func Created(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(Response{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// NoContent sends a 204 No Content response
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// Paginated sends a paginated response
func Paginated(c *fiber.Ctx, data interface{}, page, limit int, total int64) error {
	meta := CalculateMeta(page, limit, total)
	return c.JSON(Response{
		Success:   true,
		Data:      data,
		Meta:      meta,
		Timestamp: time.Now().Unix(),
	})
}

// Error sends an error response
func Error(c *fiber.Ctx, statusCode int, message string, errors interface{}) error {
	return c.Status(statusCode).JSON(Response{
		Success:   false,
		Message:   message,
		Errors:    errors,
		Timestamp: time.Now().Unix(),
	})
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *fiber.Ctx, message string, errors interface{}) error {
	return Error(c, fiber.StatusBadRequest, message, errors)
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusUnauthorized, message, nil)
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusForbidden, message, nil)
}

// NotFound sends a 404 Not Found response
func NotFound(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusNotFound, message, nil)
}

// Conflict sends a 409 Conflict response
func Conflict(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusConflict, message, nil)
}

// ValidationError sends a 422 Unprocessable Entity response
func ValidationError(c *fiber.Ctx, errors interface{}) error {
	return Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errors)
}

// InternalError sends a 500 Internal Server Error response
func InternalError(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusInternalServerError, message, nil)
}

// ServiceUnavailable sends a 503 Service Unavailable response
func ServiceUnavailable(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusServiceUnavailable, message, nil)
}

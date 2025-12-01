package validation

import (
	"github.com/gofiber/fiber/v2"
	"neonexcore/pkg/errors"
)

// ValidateBody validates request body and binds to struct
func ValidateBody(c *fiber.Ctx, data interface{}) error {
	// Parse body
	if err := c.BodyParser(data); err != nil {
		return errors.NewBadRequest("Invalid request body")
	}

	// Validate
	validator := NewValidator()
	if errs := validator.Validate(data); errs != nil {
		details := make(map[string]interface{})
		for field, message := range errs {
			details[field] = message
		}
		return errors.NewValidationError("Validation failed", details)
	}

	return nil
}

// ValidateQuery validates query parameters
func ValidateQuery(c *fiber.Ctx, data interface{}) error {
	if err := c.QueryParser(data); err != nil {
		return errors.NewBadRequest("Invalid query parameters")
	}

	validator := NewValidator()
	if errs := validator.Validate(data); errs != nil {
		details := make(map[string]interface{})
		for field, message := range errs {
			details[field] = message
		}
		return errors.NewValidationError("Validation failed", details)
	}

	return nil
}

// ValidateParams validates URL parameters
func ValidateParams(c *fiber.Ctx, data interface{}) error {
	if err := c.ParamsParser(data); err != nil {
		return errors.NewBadRequest("Invalid URL parameters")
	}

	validator := NewValidator()
	if errs := validator.Validate(data); errs != nil {
		details := make(map[string]interface{})
		for field, message := range errs {
			details[field] = message
		}
		return errors.NewValidationError("Validation failed", details)
	}

	return nil
}

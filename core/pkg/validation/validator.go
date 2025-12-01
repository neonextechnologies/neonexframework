package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps go-playground/validator
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	v := validator.New()
	
	// Register custom validators
	v.RegisterValidation("slug", validateSlug)
	v.RegisterValidation("username", validateUsername)
	v.RegisterValidation("semver", validateSemver)
	
	// Use JSON tag names in error messages
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	
	return &Validator{validate: v}
}

// Validate validates a struct
func (v *Validator) Validate(data interface{}) map[string]string {
	err := v.validate.Struct(data)
	if err == nil {
		return nil
	}

	errors := make(map[string]string)
	for _, err := range err.(validator.ValidationErrors) {
		field := err.Field()
		errors[field] = formatError(err)
	}

	return errors
}

// ValidateVar validates a single variable
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	return v.validate.Var(field, tag)
}

// formatError formats validation error message
func formatError(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, err.Param())
	case "max":
		return fmt.Sprintf("%s must not exceed %s characters", field, err.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, err.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, err.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, err.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, err.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, err.Param())
	case "eq":
		return fmt.Sprintf("%s must be equal to %s", field, err.Param())
	case "ne":
		return fmt.Sprintf("%s must not be equal to %s", field, err.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of [%s]", field, err.Param())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uri":
		return fmt.Sprintf("%s must be a valid URI", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", field)
	case "numeric":
		return fmt.Sprintf("%s must be numeric", field)
	case "number":
		return fmt.Sprintf("%s must be a number", field)
	case "slug":
		return fmt.Sprintf("%s must be a valid slug (lowercase letters, numbers, and hyphens)", field)
	case "username":
		return fmt.Sprintf("%s must be a valid username (3-20 alphanumeric characters or underscore)", field)
	case "semver":
		return fmt.Sprintf("%s must be a valid semantic version (e.g., 1.0.0)", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "uuid4":
		return fmt.Sprintf("%s must be a valid UUID v4", field)
	case "datetime":
		return fmt.Sprintf("%s must be a valid datetime", field)
	case "e164":
		return fmt.Sprintf("%s must be a valid E.164 phone number", field)
	case "ip":
		return fmt.Sprintf("%s must be a valid IP address", field)
	case "ipv4":
		return fmt.Sprintf("%s must be a valid IPv4 address", field)
	case "ipv6":
		return fmt.Sprintf("%s must be a valid IPv6 address", field)
	case "mac":
		return fmt.Sprintf("%s must be a valid MAC address", field)
	default:
		return fmt.Sprintf("%s failed validation: %s", field, tag)
	}
}

// Custom validators

// validateSlug validates slug format (lowercase, alphanumeric, hyphens)
func validateSlug(fl validator.FieldLevel) bool {
	slug := fl.Field().String()
	match, _ := regexp.MatchString(`^[a-z0-9]+(?:-[a-z0-9]+)*$`, slug)
	return match
}

// validateUsername validates username format
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	if len(username) < 3 || len(username) > 20 {
		return false
	}
	match, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, username)
	return match
}

// validateSemver validates semantic versioning format (e.g., 1.0.0, 1.2.3-alpha)
func validateSemver(fl validator.FieldLevel) bool {
	version := fl.Field().String()
	// Basic semver regex: major.minor.patch with optional pre-release/build metadata
	match, _ := regexp.MatchString(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`, version)
	return match
}

// Common validation rules
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,username"`
	Password string `json:"password" validate:"required,min=8,max=100"`
}

type UpdateProfileRequest struct {
	Name  string `json:"name" validate:"omitempty,min=2,max=100"`
	Email string `json:"email" validate:"omitempty,email"`
	Bio   string `json:"bio" validate:"omitempty,max=500"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=100"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

type CreateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=50"`
	Slug        string `json:"slug" validate:"required,slug"`
	Description string `json:"description" validate:"omitempty,max=255"`
}

type CreatePermissionRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Slug        string `json:"slug" validate:"required,slug"`
	Module      string `json:"module" validate:"required,min=2,max=50"`
	Description string `json:"description" validate:"omitempty,max=255"`
}

# Input Validation

NeonEx Framework provides robust input validation using the go-playground/validator package with custom validators and user-friendly error messages. This ensures data integrity and provides clear feedback to API clients.

## Table of Contents

- [Overview](#overview)
- [Basic Validation](#basic-validation)
- [Struct Validation](#struct-validation)
- [Custom Validators](#custom-validators)
- [Error Messages](#error-messages)
- [Common Validation Rules](#common-validation-rules)
- [Advanced Validation](#advanced-validation)
- [Best Practices](#best-practices)

## Overview

The validation package wraps go-playground/validator with additional features:
- Custom validation rules (slug, username, semver)
- Friendly error messages
- JSON field name support
- Easy integration with request handlers

```go
import "neonexcore/pkg/validation"

// Create validator instance
validator := validation.NewValidator()

// Validate struct
errors := validator.Validate(data)
if errors != nil {
    return api.ValidationError(c, errors)
}
```

## Basic Validation

### Simple Struct Validation

```go
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"required,gte=18,lte=120"`
    Username string `json:"username" validate:"required,username"`
}

func createUser(c *fiber.Ctx) error {
    var req CreateUserRequest
    
    // Parse request
    if err := c.BodyParser(&req); err != nil {
        return api.BadRequest(c, "Invalid request body", nil)
    }
    
    // Validate
    validator := validation.NewValidator()
    if errors := validator.Validate(req); errors != nil {
        return api.ValidationError(c, errors)
    }
    
    // Process valid request
    user, err := userService.Create(req)
    if err != nil {
        return api.InternalError(c, err.Error())
    }
    
    return api.Created(c, "User created successfully", user)
}
```

### Validation Response Format

When validation fails, the response includes field-specific errors:

```json
{
  "success": false,
  "message": "Validation failed",
  "errors": {
    "name": "name must be at least 2 characters",
    "email": "email must be a valid email address",
    "age": "age must be greater than or equal to 18"
  },
  "timestamp": 1638360000
}
```

## Struct Validation

### Login Request

```go
type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
}

func login(c *fiber.Ctx) error {
    var req LoginRequest
    if err := c.BodyParser(&req); err != nil {
        return api.BadRequest(c, "Invalid request", nil)
    }
    
    if errors := validator.Validate(req); errors != nil {
        return api.ValidationError(c, errors)
    }
    
    token, err := authService.Login(req.Email, req.Password)
    if err != nil {
        return api.Unauthorized(c, "Invalid credentials")
    }
    
    return api.Success(c, fiber.Map{"token": token})
}
```

### Registration Request

```go
type RegisterRequest struct {
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Email    string `json:"email" validate:"required,email"`
    Username string `json:"username" validate:"required,username"`
    Password string `json:"password" validate:"required,min=8,max=100"`
}

func register(c *fiber.Ctx) error {
    var req RegisterRequest
    if err := c.BodyParser(&req); err != nil {
        return api.BadRequest(c, "Invalid request", nil)
    }
    
    if errors := validator.Validate(req); errors != nil {
        return api.ValidationError(c, errors)
    }
    
    user, err := authService.Register(req)
    if err != nil {
        return api.InternalError(c, err.Error())
    }
    
    return api.Created(c, "Registration successful", user)
}
```

### Update Profile Request

```go
type UpdateProfileRequest struct {
    Name  string `json:"name" validate:"omitempty,min=2,max=100"`
    Email string `json:"email" validate:"omitempty,email"`
    Bio   string `json:"bio" validate:"omitempty,max=500"`
}

// Note: Use 'omitempty' for optional fields
```

### Change Password Request

```go
type ChangePasswordRequest struct {
    CurrentPassword string `json:"current_password" validate:"required"`
    NewPassword     string `json:"new_password" validate:"required,min=8,max=100"`
    ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

func changePassword(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(uint)
    
    var req ChangePasswordRequest
    if err := c.BodyParser(&req); err != nil {
        return api.BadRequest(c, "Invalid request", nil)
    }
    
    if errors := validator.Validate(req); errors != nil {
        return api.ValidationError(c, errors)
    }
    
    err := userService.ChangePassword(userID, req.CurrentPassword, req.NewPassword)
    if err != nil {
        return api.BadRequest(c, err.Error(), nil)
    }
    
    return api.Success(c, fiber.Map{"message": "Password changed successfully"})
}
```

## Custom Validators

NeonEx includes custom validators for common patterns:

### Slug Validator

Validates URL-friendly slugs (lowercase, alphanumeric, hyphens):

```go
type CreatePostRequest struct {
    Title   string `json:"title" validate:"required,min=3,max=200"`
    Slug    string `json:"slug" validate:"required,slug"`
    Content string `json:"content" validate:"required,min=10"`
}

// Valid slugs:
// - "my-post"
// - "hello-world-123"
// - "post-title"

// Invalid slugs:
// - "My Post" (spaces, uppercase)
// - "my_post" (underscores)
// - "my--post" (consecutive hyphens)
```

### Username Validator

Validates usernames (3-20 alphanumeric characters or underscore):

```go
type CreateUserRequest struct {
    Username string `json:"username" validate:"required,username"`
}

// Valid usernames:
// - "john_doe"
// - "user123"
// - "alice"

// Invalid usernames:
// - "ab" (too short)
// - "user@name" (special characters)
// - "very_long_username_here" (too long)
```

### Semantic Version Validator

Validates semantic versioning format:

```go
type CreateModuleRequest struct {
    Name    string `json:"name" validate:"required"`
    Version string `json:"version" validate:"required,semver"`
}

// Valid versions:
// - "1.0.0"
// - "2.1.3"
// - "1.0.0-alpha"
// - "1.0.0-beta.1"
// - "1.0.0+build.123"

// Invalid versions:
// - "1.0" (incomplete)
// - "v1.0.0" (prefix not allowed)
// - "1.0.0.0" (too many parts)
```

### Creating Custom Validators

Add your own validation rules:

```go
package validation

import "github.com/go-playground/validator/v10"

// Register custom validator
func (v *Validator) RegisterCustom() {
    v.validate.RegisterValidation("phone", validatePhone)
    v.validate.RegisterValidation("zipcode", validateZipCode)
    v.validate.RegisterValidation("color", validateHexColor)
}

// Phone number validator
func validatePhone(fl validator.FieldLevel) bool {
    phone := fl.Field().String()
    match, _ := regexp.MatchString(`^\+?[1-9]\d{1,14}$`, phone)
    return match
}

// Zip code validator (US format)
func validateZipCode(fl validator.FieldLevel) bool {
    zip := fl.Field().String()
    match, _ := regexp.MatchString(`^\d{5}(-\d{4})?$`, zip)
    return match
}

// Hex color validator
func validateHexColor(fl validator.FieldLevel) bool {
    color := fl.Field().String()
    match, _ := regexp.MatchString(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`, color)
    return match
}
```

Usage:

```go
type ProfileRequest struct {
    Phone string `json:"phone" validate:"required,phone"`
    Zip   string `json:"zip" validate:"required,zipcode"`
    Color string `json:"color" validate:"omitempty,color"`
}
```

## Error Messages

The validator provides user-friendly error messages:

```go
// Required field
"name is required"

// Email validation
"email must be a valid email address"

// Min/Max length
"password must be at least 8 characters"
"name must not exceed 100 characters"

// Numeric ranges
"age must be greater than or equal to 18"
"age must be less than or equal to 120"

// Pattern matching
"username must be a valid username (3-20 alphanumeric characters or underscore)"
"slug must be a valid slug (lowercase letters, numbers, and hyphens)"

// Field comparison
"confirm_password must be equal to new_password"
```

### Custom Error Messages

Override default messages:

```go
func customFormatError(err validator.FieldError) string {
    field := err.Field()
    tag := err.Tag()
    
    switch tag {
    case "required":
        return fmt.Sprintf("The %s field is required", field)
    case "email":
        return fmt.Sprintf("Please provide a valid email address")
    case "min":
        return fmt.Sprintf("%s must have at least %s characters", field, err.Param())
    case "password_strength":
        return fmt.Sprintf("Password must contain uppercase, lowercase, number, and special character")
    default:
        return fmt.Sprintf("Validation failed for %s", field)
    }
}
```

## Common Validation Rules

### String Validation

```go
type Example struct {
    // Required
    Name string `validate:"required"`
    
    // Length constraints
    Title    string `validate:"min=3,max=100"`
    Username string `validate:"len=10"` // Exactly 10 characters
    
    // Format validation
    Email string `validate:"email"`
    URL   string `validate:"url"`
    UUID  string `validate:"uuid"`
    
    // Pattern matching
    Alpha    string `validate:"alpha"`      // Letters only
    AlphaNum string `validate:"alphanum"`   // Letters and numbers
    Numeric  string `validate:"numeric"`    // Numbers only
    
    // Custom patterns
    Slug string `validate:"slug"`
    
    // One of allowed values
    Status string `validate:"oneof=active inactive pending"`
}
```

### Numeric Validation

```go
type Example struct {
    // Required
    Age int `validate:"required"`
    
    // Range constraints
    Age     int     `validate:"gte=18,lte=120"`
    Price   float64 `validate:"gt=0"`
    Qty     int     `validate:"min=1,max=100"`
    
    // Number validation
    Number int `validate:"number"`
    
    // Equality
    Rating int `validate:"eq=5"`
    Count  int `validate:"ne=0"` // Not equal
}
```

### Boolean Validation

```go
type Example struct {
    Active  bool `validate:"required"`
    Premium bool `validate:"omitempty"` // Optional
}
```

### Array/Slice Validation

```go
type Example struct {
    // Array length
    Tags []string `validate:"min=1,max=10"`
    IDs  []int    `validate:"required,dive,gt=0"`
    
    // Dive validates each element
    Emails []string `validate:"dive,email"`
    URLs   []string `validate:"dive,url"`
}
```

### Nested Struct Validation

```go
type Address struct {
    Street  string `validate:"required"`
    City    string `validate:"required"`
    ZipCode string `validate:"required,zipcode"`
}

type User struct {
    Name    string   `validate:"required"`
    Email   string   `validate:"required,email"`
    Address Address  `validate:"required"` // Validates nested struct
}
```

### Conditional Validation

```go
type Example struct {
    // Required if other field is present
    Email        string `validate:"omitempty,email"`
    EmailConfirm string `validate:"required_with=Email,eqfield=Email"`
    
    // Required if other field has specific value
    Premium      bool   `validate:"omitempty"`
    BillingEmail string `validate:"required_if=Premium true,email"`
}
```

## Advanced Validation

### Cross-Field Validation

```go
type DateRange struct {
    StartDate time.Time `validate:"required"`
    EndDate   time.Time `validate:"required,gtfield=StartDate"`
}

type PriceRange struct {
    MinPrice float64 `validate:"required,gte=0"`
    MaxPrice float64 `validate:"required,gtfield=MinPrice"`
}
```

### Custom Validation Logic

```go
func validateCreatePost(req CreatePostRequest) error {
    // Use built-in validation first
    if errors := validator.Validate(req); errors != nil {
        return errors
    }
    
    // Custom business logic validation
    if len(req.Tags) > 0 {
        for _, tag := range req.Tags {
            if len(tag) < 2 || len(tag) > 30 {
                return fmt.Errorf("tag length must be between 2 and 30 characters")
            }
        }
    }
    
    // Check unique constraints
    exists, err := postRepo.SlugExists(req.Slug)
    if err != nil {
        return err
    }
    if exists {
        return fmt.Errorf("slug already exists")
    }
    
    return nil
}
```

### Variable Validation

Validate individual values without struct:

```go
func handler(c *fiber.Ctx) error {
    email := c.Query("email")
    
    // Validate single variable
    err := validator.ValidateVar(email, "required,email")
    if err != nil {
        return api.BadRequest(c, "Invalid email", nil)
    }
    
    return api.Success(c, fiber.Map{"email": email})
}
```

### Validation Middleware

Create reusable validation middleware:

```go
func ValidateRequest[T any]() fiber.Handler {
    return func(c *fiber.Ctx) error {
        var req T
        
        if err := c.BodyParser(&req); err != nil {
            return api.BadRequest(c, "Invalid request body", nil)
        }
        
        validator := validation.NewValidator()
        if errors := validator.Validate(req); errors != nil {
            return api.ValidationError(c, errors)
        }
        
        // Store validated request in context
        c.Locals("validated_request", req)
        return c.Next()
    }
}

// Usage
app.Post("/users", 
    ValidateRequest[CreateUserRequest](),
    createUserHandler,
)

func createUserHandler(c *fiber.Ctx) error {
    req := c.Locals("validated_request").(CreateUserRequest)
    // Request is already validated
    user, err := userService.Create(req)
    // ...
}
```

## Best Practices

### 1. Always Validate User Input

```go
func handler(c *fiber.Ctx) error {
    var req Request
    
    // Parse
    if err := c.BodyParser(&req); err != nil {
        return api.BadRequest(c, "Invalid request", nil)
    }
    
    // Validate
    if errors := validator.Validate(req); errors != nil {
        return api.ValidationError(c, errors)
    }
    
    // Process (data is safe to use)
    return processRequest(req)
}
```

### 2. Use Descriptive Validation Tags

```go
// Good: Clear validation rules
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Email    string `json:"email" validate:"required,email"`
    Username string `json:"username" validate:"required,username"`
}

// Bad: Unclear or missing validation
type CreateUserRequest struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Username string `json:"username"`
}
```

### 3. Validate at API Boundary

Validate as early as possible (in controllers/handlers):

```go
// Good: Validate in handler
func createUser(c *fiber.Ctx) error {
    var req CreateUserRequest
    if err := c.BodyParser(&req); err != nil {
        return api.BadRequest(c, "Invalid request", nil)
    }
    
    // Validate immediately
    if errors := validator.Validate(req); errors != nil {
        return api.ValidationError(c, errors)
    }
    
    return userService.Create(req)
}

// Bad: Validation deep in service layer
func (s *UserService) Create(req CreateUserRequest) error {
    // Too late to validate here
    if req.Name == "" {
        return errors.New("name required")
    }
    // ...
}
```

### 4. Use Custom Validators for Domain Logic

```go
// Register domain-specific validators
validator.RegisterValidation("product_code", validateProductCode)
validator.RegisterValidation("order_status", validateOrderStatus)

type CreateOrderRequest struct {
    ProductCode string `validate:"required,product_code"`
    Quantity    int    `validate:"required,min=1,max=1000"`
    Status      string `validate:"required,order_status"`
}
```

### 5. Provide Clear Error Messages

```go
// Return validation errors with clear messages
if errors := validator.Validate(req); errors != nil {
    return c.Status(422).JSON(fiber.Map{
        "success": false,
        "message": "Validation failed",
        "errors":  errors,
    })
}

// Response:
// {
//   "success": false,
//   "message": "Validation failed",
//   "errors": {
//     "email": "email must be a valid email address",
//     "password": "password must be at least 8 characters"
//   }
// }
```

### 6. Test Validation Rules

```go
func TestCreateUserValidation(t *testing.T) {
    validator := validation.NewValidator()
    
    tests := []struct {
        name    string
        request CreateUserRequest
        wantErr bool
    }{
        {
            name: "valid request",
            request: CreateUserRequest{
                Name:     "John Doe",
                Email:    "john@example.com",
                Username: "johndoe",
            },
            wantErr: false,
        },
        {
            name: "invalid email",
            request: CreateUserRequest{
                Name:     "John Doe",
                Email:    "invalid-email",
                Username: "johndoe",
            },
            wantErr: true,
        },
        {
            name: "short username",
            request: CreateUserRequest{
                Name:     "John Doe",
                Email:    "john@example.com",
                Username: "ab", // Too short
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            errors := validator.Validate(tt.request)
            if (errors != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", errors, tt.wantErr)
            }
        })
    }
}
```

### 7. Separate Validation from Business Logic

```go
// Good: Separate concerns
func createUser(c *fiber.Ctx) error {
    var req CreateUserRequest
    
    // 1. Parse
    if err := c.BodyParser(&req); err != nil {
        return api.BadRequest(c, "Invalid request", nil)
    }
    
    // 2. Validate format
    if errors := validator.Validate(req); errors != nil {
        return api.ValidationError(c, errors)
    }
    
    // 3. Business logic (uniqueness checks, etc.)
    if exists := userService.EmailExists(req.Email); exists {
        return api.Conflict(c, "Email already exists")
    }
    
    // 4. Create user
    user, err := userService.Create(req)
    if err != nil {
        return api.InternalError(c, err.Error())
    }
    
    return api.Created(c, "User created", user)
}
```

This comprehensive validation system ensures data integrity and provides excellent developer experience in NeonEx Framework!

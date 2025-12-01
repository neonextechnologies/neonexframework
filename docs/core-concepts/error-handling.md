# Error Handling

NeonEx Framework provides a comprehensive error handling system with typed errors, standardized responses, logging integration, and recovery middleware. This ensures consistent error handling across your application.

## Table of Contents

- [Error Types](#error-types)
- [Error Responses](#error-responses)
- [Creating Errors](#creating-errors)
- [Error Logging](#error-logging)
- [Recovery Middleware](#recovery-middleware)
- [Custom Error Handlers](#custom-error-handlers)
- [Best Practices](#best-practices)

## Error Types

NeonEx defines structured error types with codes, messages, and HTTP status codes:

```go
import "neonexcore/pkg/errors"

// AppError represents application error
type AppError struct {
    Code       ErrorCode              // Error code
    Message    string                 // User-friendly message
    StatusCode int                    // HTTP status code
    Details    map[string]interface{} // Additional details
    Err        error                  // Underlying error
}
```

### Error Codes

```go
const (
    // General errors
    ErrCodeInternal        = "INTERNAL_ERROR"
    ErrCodeNotFound        = "NOT_FOUND"
    ErrCodeBadRequest      = "BAD_REQUEST"
    ErrCodeUnauthorized    = "UNAUTHORIZED"
    ErrCodeForbidden       = "FORBIDDEN"
    ErrCodeConflict        = "CONFLICT"
    ErrCodeValidation      = "VALIDATION_ERROR"
    ErrCodeTooManyRequests = "TOO_MANY_REQUESTS"
    
    // Authentication errors
    ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
    ErrCodeTokenExpired       = "TOKEN_EXPIRED"
    ErrCodeTokenInvalid       = "TOKEN_INVALID"
    ErrCodeAccountLocked      = "ACCOUNT_LOCKED"
    ErrCodeAccountDisabled    = "ACCOUNT_DISABLED"
    
    // Database errors
    ErrCodeDatabaseConnection  = "DATABASE_CONNECTION"
    ErrCodeRecordNotFound      = "RECORD_NOT_FOUND"
    ErrCodeDuplicateEntry      = "DUPLICATE_ENTRY"
    ErrCodeConstraintViolation = "CONSTRAINT_VIOLATION"
    
    // Module errors
    ErrCodeModuleNotFound    = "MODULE_NOT_FOUND"
    ErrCodeModuleDisabled    = "MODULE_DISABLED"
    ErrCodeModuleInstallFail = "MODULE_INSTALL_FAIL"
    
    // Permission errors
    ErrCodeInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"
    ErrCodeInvalidRole             = "INVALID_ROLE"
)
```

## Error Responses

### Standard Error Response Format

```json
{
  "success": false,
  "code": "NOT_FOUND",
  "message": "User not found",
  "details": {
    "user_id": 123
  },
  "timestamp": 1638360000
}
```

### Using Error Responses

```go
func getUser(c *fiber.Ctx) error {
    id, _ := strconv.Atoi(c.Params("id"))
    
    user, err := userService.GetByID(uint(id))
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return api.NotFound(c, "User not found")
        }
        return api.InternalError(c, "Failed to fetch user")
    }
    
    return api.Success(c, user)
}
```

## Creating Errors

### Quick Error Constructors

```go
// Bad Request (400)
err := errors.NewBadRequest("Invalid input")

// Unauthorized (401)
err := errors.NewUnauthorized("Authentication required")

// Forbidden (403)
err := errors.NewForbidden("Access denied")

// Not Found (404)
err := errors.NewNotFound("Resource not found")

// Conflict (409)
err := errors.NewConflict("Email already exists")

// Internal Server Error (500)
err := errors.NewInternal("Something went wrong")

// Validation Error (422)
err := errors.NewValidationError("Validation failed", map[string]interface{}{
    "email": "Invalid email format",
    "password": "Password too short",
})
```

### Custom Errors

```go
// Create custom error
err := errors.New(
    errors.ErrCodeInvalidCredentials,
    "Invalid email or password",
    http.StatusUnauthorized,
)

// Add details
err = err.WithDetails(map[string]interface{}{
    "attempt_count": 3,
    "locked_until": time.Now().Add(15 * time.Minute),
})

// Add underlying error
err = err.WithError(originalErr)
```

### Complete Error Example

```go
func login(c *fiber.Ctx) error {
    var req LoginRequest
    if err := c.BodyParser(&req); err != nil {
        return api.BadRequest(c, "Invalid request body", nil)
    }
    
    // Attempt login
    user, err := userService.Login(req.Email, req.Password)
    if err != nil {
        // Check error type
        switch {
        case errors.Is(err, ErrInvalidCredentials):
            return c.Status(401).JSON(fiber.Map{
                "success": false,
                "code":    errors.ErrCodeInvalidCredentials,
                "message": "Invalid email or password",
            })
            
        case errors.Is(err, ErrAccountLocked):
            return c.Status(403).JSON(fiber.Map{
                "success": false,
                "code":    errors.ErrCodeAccountLocked,
                "message": "Account locked due to multiple failed login attempts",
                "details": fiber.Map{
                    "locked_until": user.LockedUntil,
                },
            })
            
        default:
            logger.Error("Login failed", logger.Fields{
                "error": err.Error(),
                "email": req.Email,
            })
            return api.InternalError(c, "Login failed")
        }
    }
    
    return api.Success(c, user)
}
```

## Error Logging

### Basic Error Logging

```go
import "neonexcore/pkg/logger"

func handler(c *fiber.Ctx) error {
    user, err := userService.GetByID(userID)
    if err != nil {
        // Log error with context
        logger.Error("Failed to fetch user", logger.Fields{
            "error":   err.Error(),
            "user_id": userID,
            "request_id": c.Locals("request_id"),
        })
        
        return api.InternalError(c, "Failed to fetch user")
    }
    
    return api.Success(c, user)
}
```

### Structured Error Logging

```go
func processOrder(c *fiber.Ctx) error {
    var req CreateOrderRequest
    if err := c.BodyParser(&req); err != nil {
        logger.Warn("Invalid request body", logger.Fields{
            "error": err.Error(),
            "ip":    c.IP(),
        })
        return api.BadRequest(c, "Invalid request", nil)
    }
    
    order, err := orderService.Create(req)
    if err != nil {
        logger.Error("Order creation failed", logger.Fields{
            "error":      err.Error(),
            "user_id":    c.Locals("user_id"),
            "product_id": req.ProductID,
            "quantity":   req.Quantity,
        })
        return api.InternalError(c, "Failed to create order")
    }
    
    logger.Info("Order created", logger.Fields{
        "order_id": order.ID,
        "user_id":  c.Locals("user_id"),
        "total":    order.Total,
    })
    
    return api.Created(c, "Order created", order)
}
```

### Error Logging Middleware

```go
func ErrorLoggingMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        err := c.Next()
        
        if err != nil {
            // Log all errors
            logger.Error("Request error", logger.Fields{
                "error":      err.Error(),
                "method":     c.Method(),
                "path":       c.Path(),
                "ip":         c.IP(),
                "user_agent": c.Get("User-Agent"),
                "request_id": c.Locals("request_id"),
            })
            
            // Check if it's an AppError
            if appErr, ok := err.(*errors.AppError); ok {
                return c.Status(appErr.StatusCode).JSON(fiber.Map{
                    "success": false,
                    "code":    appErr.Code,
                    "message": appErr.Message,
                    "details": appErr.Details,
                })
            }
            
            // Generic error
            return c.Status(500).JSON(fiber.Map{
                "success": false,
                "message": "Internal server error",
            })
        }
        
        return nil
    }
}
```

## Recovery Middleware

### Basic Recovery

```go
func RecoveryMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        defer func() {
            if r := recover(); r != nil {
                // Log panic
                logger.Error("Panic recovered", logger.Fields{
                    "panic":      fmt.Sprintf("%v", r),
                    "stack":      string(debug.Stack()),
                    "method":     c.Method(),
                    "path":       c.Path(),
                    "request_id": c.Locals("request_id"),
                })
                
                // Return error response
                c.Status(500).JSON(fiber.Map{
                    "success": false,
                    "message": "Internal server error",
                })
            }
        }()
        
        return c.Next()
    }
}
```

### Advanced Recovery with Notification

```go
func AdvancedRecoveryMiddleware(notifier Notifier) fiber.Handler {
    return func(c *fiber.Ctx) error {
        defer func() {
            if r := recover(); r != nil {
                panicErr := fmt.Sprintf("%v", r)
                stack := string(debug.Stack())
                
                // Log panic
                logger.Error("PANIC", logger.Fields{
                    "panic":      panicErr,
                    "stack":      stack,
                    "method":     c.Method(),
                    "path":       c.Path(),
                    "ip":         c.IP(),
                    "user_id":    c.Locals("user_id"),
                    "request_id": c.Locals("request_id"),
                })
                
                // Send notification (Slack, email, etc.)
                notifier.SendAlert("Panic in production", fiber.Map{
                    "panic":   panicErr,
                    "path":    c.Path(),
                    "method":  c.Method(),
                    "user_id": c.Locals("user_id"),
                })
                
                // Return error response
                c.Status(500).JSON(fiber.Map{
                    "success": false,
                    "message": "Internal server error",
                })
            }
        }()
        
        return c.Next()
    }
}
```

## Custom Error Handlers

### Global Error Handler

```go
func CustomErrorHandler(c *fiber.Ctx, err error) error {
    // Default status code
    code := fiber.StatusInternalServerError
    message := "Internal Server Error"
    
    // Check error type
    if e, ok := err.(*fiber.Error); ok {
        code = e.Code
        message = e.Message
    }
    
    // Check AppError
    if appErr, ok := err.(*errors.AppError); ok {
        return c.Status(appErr.StatusCode).JSON(fiber.Map{
            "success":   false,
            "code":      appErr.Code,
            "message":   appErr.Message,
            "details":   appErr.Details,
            "timestamp": time.Now().Unix(),
        })
    }
    
    // Log error
    logger.Error("Request failed", logger.Fields{
        "error":      err.Error(),
        "method":     c.Method(),
        "path":       c.Path(),
        "status":     code,
        "request_id": c.Locals("request_id"),
    })
    
    // Return error response
    return c.Status(code).JSON(fiber.Map{
        "success":   false,
        "message":   message,
        "timestamp": time.Now().Unix(),
    })
}

// Use in Fiber config
app := fiber.New(fiber.Config{
    ErrorHandler: CustomErrorHandler,
})
```

### Domain-Specific Error Handlers

```go
type UserError struct {
    Type    string
    Message string
    UserID  uint
}

func (e *UserError) Error() string {
    return fmt.Sprintf("[User Error] %s: %s (UserID: %d)", e.Type, e.Message, e.UserID)
}

func HandleUserError(c *fiber.Ctx, err *UserError) error {
    logger.Error("User error", logger.Fields{
        "type":    err.Type,
        "message": err.Message,
        "user_id": err.UserID,
    })
    
    switch err.Type {
    case "not_found":
        return api.NotFound(c, err.Message)
    case "already_exists":
        return api.Conflict(c, err.Message)
    case "invalid_state":
        return api.BadRequest(c, err.Message, nil)
    default:
        return api.InternalError(c, err.Message)
    }
}
```

## Best Practices

### 1. Use Appropriate Error Types

```go
// Good: Specific error types
func getUser(c *fiber.Ctx) error {
    user, err := userService.GetByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return api.NotFound(c, "User not found")
        }
        return api.InternalError(c, "Database error")
    }
    return api.Success(c, user)
}

// Bad: Generic errors
func getUser(c *fiber.Ctx) error {
    user, err := userService.GetByID(id)
    if err != nil {
        return c.Status(500).SendString("Error")
    }
    return c.JSON(user)
}
```

### 2. Always Log Errors

```go
func handler(c *fiber.Ctx) error {
    result, err := service.DoSomething()
    if err != nil {
        // Always log errors before returning
        logger.Error("Operation failed", logger.Fields{
            "error":      err.Error(),
            "user_id":    c.Locals("user_id"),
            "request_id": c.Locals("request_id"),
        })
        return api.InternalError(c, "Operation failed")
    }
    return api.Success(c, result)
}
```

### 3. Don't Expose Internal Details

```go
// Good: Safe error message
func handler(c *fiber.Ctx) error {
    err := db.Query(dangerousSQL)
    if err != nil {
        logger.Error("Database error", logger.Fields{
            "error": err.Error(), // Log full error
        })
        return api.InternalError(c, "Database operation failed") // Generic message
    }
    return api.Success(c, result)
}

// Bad: Exposes internal details
func handler(c *fiber.Ctx) error {
    err := db.Query(dangerousSQL)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": err.Error(), // Could expose SQL, file paths, etc.
        })
    }
    return c.JSON(result)
}
```

### 4. Use Error Wrapping

```go
func (s *UserService) Create(req CreateUserRequest) (*User, error) {
    user := &User{
        Name:  req.Name,
        Email: req.Email,
    }
    
    if err := s.repo.Create(user); err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    return user, nil
}

// Handler can check wrapped errors
func createUser(c *fiber.Ctx) error {
    user, err := userService.Create(req)
    if err != nil {
        if errors.Is(err, gorm.ErrDuplicatedKey) {
            return api.Conflict(c, "Email already exists")
        }
        logger.Error("Failed to create user", logger.Fields{
            "error": err.Error(),
        })
        return api.InternalError(c, "Failed to create user")
    }
    return api.Created(c, "User created", user)
}
```

### 5. Provide Helpful Error Details

```go
// Good: Detailed validation errors
func handler(c *fiber.Ctx) error {
    if errors := validator.Validate(req); errors != nil {
        return api.ValidationError(c, errors)
    }
    // ...
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

### 6. Use Recovery Middleware

```go
func main() {
    app := fiber.New(fiber.Config{
        ErrorHandler: CustomErrorHandler,
    })
    
    // Recovery middleware should be first
    app.Use(RecoveryMiddleware())
    
    // Other middleware...
    app.Use(logger.LoggerMiddleware())
    app.Use(auth.AuthMiddleware(jwtManager))
    
    // Routes...
}
```

### 7. Test Error Handling

```go
func TestGetUserNotFound(t *testing.T) {
    app := fiber.New()
    app.Get("/users/:id", getUser)
    
    req := httptest.NewRequest("GET", "/users/999", nil)
    resp, _ := app.Test(req)
    
    assert.Equal(t, 404, resp.StatusCode)
    
    var body map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&body)
    
    assert.Equal(t, false, body["success"])
    assert.Equal(t, "User not found", body["message"])
}

func TestGetUserInternalError(t *testing.T) {
    // Mock service to return error
    mockService := &MockUserService{
        GetByIDFunc: func(id uint) (*User, error) {
            return nil, errors.New("database error")
        },
    }
    
    app := fiber.New()
    app.Get("/users/:id", getUser)
    
    req := httptest.NewRequest("GET", "/users/1", nil)
    resp, _ := app.Test(req)
    
    assert.Equal(t, 500, resp.StatusCode)
}
```

## Complete Example

```go
package main

import (
    "neonexcore/pkg/api"
    "neonexcore/pkg/errors"
    "neonexcore/pkg/logger"
    "github.com/gofiber/fiber/v2"
)

// Custom error handler
func errorHandler(c *fiber.Ctx, err error) error {
    code := fiber.StatusInternalServerError
    
    if e, ok := err.(*fiber.Error); ok {
        code = e.Code
    }
    
    if appErr, ok := err.(*errors.AppError); ok {
        logger.Error("Request failed", logger.Fields{
            "code":       appErr.Code,
            "message":    appErr.Message,
            "path":       c.Path(),
            "request_id": c.Locals("request_id"),
        })
        
        return c.Status(appErr.StatusCode).JSON(fiber.Map{
            "success": false,
            "code":    appErr.Code,
            "message": appErr.Message,
            "details": appErr.Details,
        })
    }
    
    logger.Error("Unhandled error", logger.Fields{
        "error":      err.Error(),
        "path":       c.Path(),
        "request_id": c.Locals("request_id"),
    })
    
    return c.Status(code).JSON(fiber.Map{
        "success": false,
        "message": "Internal server error",
    })
}

// Recovery middleware
func recoveryMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        defer func() {
            if r := recover(); r != nil {
                logger.Error("Panic recovered", logger.Fields{
                    "panic":      fmt.Sprintf("%v", r),
                    "stack":      string(debug.Stack()),
                    "path":       c.Path(),
                    "request_id": c.Locals("request_id"),
                })
                
                c.Status(500).JSON(fiber.Map{
                    "success": false,
                    "message": "Internal server error",
                })
            }
        }()
        return c.Next()
    }
}

func main() {
    app := fiber.New(fiber.Config{
        ErrorHandler: errorHandler,
    })
    
    // Apply recovery middleware first
    app.Use(recoveryMiddleware())
    
    // Routes
    app.Post("/users", createUser)
    app.Get("/users/:id", getUser)
    
    app.Listen(":3000")
}

func getUser(c *fiber.Ctx) error {
    id, _ := strconv.Atoi(c.Params("id"))
    
    user, err := userService.GetByID(uint(id))
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return api.NotFound(c, "User not found")
        }
        logger.Error("Failed to fetch user", logger.Fields{
            "error":   err.Error(),
            "user_id": id,
        })
        return api.InternalError(c, "Failed to fetch user")
    }
    
    return api.Success(c, user)
}
```

This comprehensive error handling system ensures robust and maintainable applications with NeonEx Framework!

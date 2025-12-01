# Middleware

Middleware in NeonEx Framework provides a powerful mechanism to filter, inspect, and modify HTTP requests and responses before they reach your route handlers. Built on top of Fiber's middleware system, NeonEx provides both built-in middleware and an intuitive API for creating custom middleware.

## Table of Contents

- [Overview](#overview)
- [Built-in Middleware](#built-in-middleware)
- [Authentication Middleware](#authentication-middleware)
- [RBAC Middleware](#rbac-middleware)
- [Custom Middleware](#custom-middleware)
- [Error Handling Middleware](#error-handling-middleware)
- [Middleware Ordering](#middleware-ordering)
- [Best Practices](#best-practices)

## Overview

Middleware functions are executed in the order they are registered and can perform operations before and after route handlers. They have access to the request context and can:

- Execute code before the handler
- Modify the request or response
- End the request-response cycle
- Call the next middleware in the stack

```go
func MyMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Before handler
        start := time.Now()
        
        // Call next middleware/handler
        err := c.Next()
        
        // After handler
        duration := time.Since(start)
        log.Printf("Request took %v", duration)
        
        return err
    }
}
```

## Built-in Middleware

### CORS Middleware

Configure Cross-Origin Resource Sharing (CORS) for your API:

```go
package main

import (
    "neonexcore/pkg/api"
    "github.com/gofiber/fiber/v2"
)

func setupRoutes(app *fiber.App) {
    // Default CORS (allows all origins)
    app.Use(api.CORSMiddleware())
    
    // Custom CORS configuration
    app.Use(api.CORSMiddleware(api.CORSConfig{
        AllowOrigins: []string{
            "https://yourdomain.com",
            "https://app.yourdomain.com",
        },
        AllowMethods: []string{
            fiber.MethodGet,
            fiber.MethodPost,
            fiber.MethodPut,
            fiber.MethodDelete,
        },
        AllowHeaders: []string{
            "Origin",
            "Content-Type",
            "Authorization",
        },
        AllowCredentials: true,
        MaxAge: 3600,
    }))
}
```

**Production CORS Setup:**

```go
func setupProductionCORS(app *fiber.App) {
    corsConfig := api.ProductionCORSConfig([]string{
        "https://yourdomain.com",
        "https://api.yourdomain.com",
    })
    
    app.Use(api.CORSMiddleware(corsConfig))
}
```

### Security Headers Middleware

Add essential security headers to all responses:

```go
app.Use(api.SecurityHeadersMiddleware())

// Adds the following headers:
// X-Content-Type-Options: nosniff
// X-XSS-Protection: 1; mode=block
// X-Frame-Options: DENY
// Strict-Transport-Security: max-age=31536000; includeSubDomains
// Referrer-Policy: strict-origin-when-cross-origin
// Content-Security-Policy: default-src 'self'
// Permissions-Policy: geolocation=(), microphone=(), camera=()
```

### Request ID Middleware

Track requests with unique IDs:

```go
app.Use(api.RequestIDMiddleware())

// In your handler
func MyHandler(c *fiber.Ctx) error {
    requestID := c.Locals("request_id").(string)
    log.Printf("Processing request: %s", requestID)
    return c.SendString("OK")
}
```

### Logger Middleware

Log all incoming requests:

```go
app.Use(api.LoggerMiddleware())
```

### Rate Limiting

Protect your API from abuse with rate limiting:

```go
import "neonexcore/pkg/api"

// Basic rate limiting (10 requests per minute)
app.Use(api.RateLimitMiddleware(api.RateLimitConfig{
    Max:      10,
    Duration: time.Minute,
}))

// Advanced rate limiting with custom key
app.Use(api.RateLimitMiddleware(api.RateLimitConfig{
    Max:      100,
    Duration: time.Minute,
    KeyGenerator: func(c *fiber.Ctx) string {
        // Rate limit by user ID if authenticated
        if userID, ok := c.Locals("user_id").(uint); ok {
            return fmt.Sprintf("user:%d", userID)
        }
        // Otherwise by IP
        return c.IP()
    },
}))
```

## Authentication Middleware

### JWT Authentication

Protect routes with JWT authentication:

```go
import (
    "neonexcore/pkg/auth"
    "github.com/gofiber/fiber/v2"
)

func setupAuthRoutes(app *fiber.App, jwtManager *auth.JWTManager) {
    // Public routes
    app.Post("/login", loginHandler)
    app.Post("/register", registerHandler)
    
    // Protected routes - require authentication
    api := app.Group("/api")
    api.Use(auth.AuthMiddleware(jwtManager))
    
    api.Get("/profile", getProfile)
    api.Put("/profile", updateProfile)
    api.Post("/posts", createPost)
}

// Access user info in handlers
func getProfile(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(uint)
    email := c.Locals("email").(string)
    role := c.Locals("role").(string)
    
    // Fetch user data
    user, err := userService.GetByID(userID)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch user"})
    }
    
    return c.JSON(user)
}
```

### Optional Authentication

Allow both authenticated and guest access:

```go
app.Use(auth.OptionalAuthMiddleware(jwtManager))

func myHandler(c *fiber.Ctx) error {
    // Check if user is authenticated
    userID, isAuthenticated := auth.GetUserID(c)
    
    if isAuthenticated {
        return c.JSON(fiber.Map{
            "message": "Welcome back!",
            "user_id": userID,
        })
    }
    
    return c.JSON(fiber.Map{
        "message": "Welcome, guest!",
    })
}
```

### Helper Functions

Extract authentication data from context:

```go
// Get user ID
userID, ok := auth.GetUserID(c)

// Get user email
email, ok := auth.GetUserEmail(c)

// Get user role
role, ok := auth.GetUserRole(c)

// Get user permissions
permissions, ok := auth.GetUserPermissions(c)

// Get full claims
claims, ok := auth.GetClaims(c)
```

## RBAC Middleware

### Permission-Based Access

Require specific permissions:

```go
import "neonexcore/pkg/rbac"

func setupAdminRoutes(app *fiber.App, rbacManager *rbac.Manager) {
    admin := app.Group("/admin")
    admin.Use(auth.AuthMiddleware(jwtManager))
    
    // Require 'users.manage' permission
    admin.Get("/users", 
        rbac.RequirePermission(rbacManager, "users.manage"),
        listUsers,
    )
    
    // Require 'posts.delete' permission
    admin.Delete("/posts/:id",
        rbac.RequirePermission(rbacManager, "posts.delete"),
        deletePost,
    )
}
```

### Role-Based Access

Require specific roles:

```go
// Only admins can access
admin := app.Group("/admin")
admin.Use(auth.AuthMiddleware(jwtManager))
admin.Use(rbac.RequireRole(rbacManager, "admin"))

admin.Get("/dashboard", adminDashboard)
admin.Get("/settings", adminSettings)
```

### Multiple Permissions

Require any one of multiple permissions:

```go
// User needs either 'posts.create' OR 'posts.manage'
app.Post("/posts",
    auth.AuthMiddleware(jwtManager),
    rbac.RequireAnyPermission(rbacManager, "posts.create", "posts.manage"),
    createPost,
)
```

Require all permissions:

```go
// User needs both 'users.manage' AND 'roles.assign'
app.Post("/users/:id/roles",
    auth.AuthMiddleware(jwtManager),
    rbac.RequireAllPermissions(rbacManager, "users.manage", "roles.assign"),
    assignRole,
)
```

## Custom Middleware

### Simple Custom Middleware

Create basic middleware:

```go
func TimingMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        start := time.Now()
        
        err := c.Next()
        
        duration := time.Since(start)
        c.Set("X-Response-Time", duration.String())
        
        return err
    }
}

// Usage
app.Use(TimingMiddleware())
```

### Middleware with Configuration

Create configurable middleware:

```go
type LogConfig struct {
    SkipPaths []string
    LogBody   bool
}

func LoggingMiddleware(config LogConfig) fiber.Handler {
    // Create skip path map for fast lookup
    skipPaths := make(map[string]bool)
    for _, path := range config.SkipPaths {
        skipPaths[path] = true
    }
    
    return func(c *fiber.Ctx) error {
        // Skip logging for certain paths
        if skipPaths[c.Path()] {
            return c.Next()
        }
        
        // Log request
        log.Printf("[%s] %s", c.Method(), c.Path())
        
        if config.LogBody {
            log.Printf("Body: %s", c.Body())
        }
        
        return c.Next()
    }
}

// Usage
app.Use(LoggingMiddleware(LogConfig{
    SkipPaths: []string{"/health", "/metrics"},
    LogBody:   false,
}))
```

### Middleware with Dependencies

Inject dependencies into middleware:

```go
type APIKeyMiddleware struct {
    db *gorm.DB
}

func NewAPIKeyMiddleware(db *gorm.DB) *APIKeyMiddleware {
    return &APIKeyMiddleware{db: db}
}

func (m *APIKeyMiddleware) Validate() fiber.Handler {
    return func(c *fiber.Ctx) error {
        apiKey := c.Get("X-API-Key")
        if apiKey == "" {
            return c.Status(401).JSON(fiber.Map{
                "error": "API key required",
            })
        }
        
        // Validate API key in database
        var user User
        if err := m.db.Where("api_key = ?", apiKey).First(&user).Error; err != nil {
            return c.Status(401).JSON(fiber.Map{
                "error": "Invalid API key",
            })
        }
        
        c.Locals("user_id", user.ID)
        return c.Next()
    }
}

// Usage
apiKeyMW := NewAPIKeyMiddleware(db)
app.Use(apiKeyMW.Validate())
```

### Conditional Middleware

Apply middleware based on conditions:

```go
func ConditionalMiddleware(condition func(*fiber.Ctx) bool, middleware fiber.Handler) fiber.Handler {
    return func(c *fiber.Ctx) error {
        if condition(c) {
            return middleware(c)
        }
        return c.Next()
    }
}

// Usage: Only apply rate limiting to API routes
app.Use(ConditionalMiddleware(
    func(c *fiber.Ctx) bool {
        return strings.HasPrefix(c.Path(), "/api/")
    },
    api.RateLimitMiddleware(api.RateLimitConfig{
        Max:      100,
        Duration: time.Minute,
    }),
))
```

## Error Handling Middleware

### Global Error Handler

Catch and format all errors:

```go
func ErrorHandlerMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        err := c.Next()
        
        if err == nil {
            return nil
        }
        
        // Log error
        log.Printf("Error: %v", err)
        
        // Check if it's a Fiber error
        if fiberErr, ok := err.(*fiber.Error); ok {
            return c.Status(fiberErr.Code).JSON(fiber.Map{
                "success": false,
                "error":   fiberErr.Message,
            })
        }
        
        // Check if it's an AppError
        if appErr, ok := err.(*errors.AppError); ok {
            return c.Status(appErr.StatusCode).JSON(fiber.Map{
                "success": false,
                "code":    appErr.Code,
                "message": appErr.Message,
                "details": appErr.Details,
            })
        }
        
        // Unknown error - return 500
        return c.Status(500).JSON(fiber.Map{
            "success": false,
            "error":   "Internal server error",
        })
    }
}

// Apply globally
app.Use(ErrorHandlerMiddleware())
```

### Recovery Middleware

Recover from panics:

```go
func RecoveryMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        defer func() {
            if r := recover(); r != nil {
                log.Printf("Panic recovered: %v", r)
                log.Printf("Stack trace: %s", debug.Stack())
                
                c.Status(500).JSON(fiber.Map{
                    "success": false,
                    "error":   "Internal server error",
                })
            }
        }()
        
        return c.Next()
    }
}

// Apply as first middleware
app.Use(RecoveryMiddleware())
```

## Middleware Ordering

Order matters! Apply middleware in this sequence:

```go
func setupMiddleware(app *fiber.App) {
    // 1. Recovery (catch panics first)
    app.Use(RecoveryMiddleware())
    
    // 2. Request ID (for logging/tracing)
    app.Use(api.RequestIDMiddleware())
    
    // 3. Logger (log all requests)
    app.Use(api.LoggerMiddleware())
    
    // 4. CORS (handle preflight requests)
    app.Use(api.CORSMiddleware())
    
    // 5. Security headers
    app.Use(api.SecurityHeadersMiddleware())
    
    // 6. Compression
    app.Use(api.CompressionMiddleware())
    
    // 7. Rate limiting
    app.Use(api.RateLimitMiddleware(api.RateLimitConfig{
        Max:      100,
        Duration: time.Minute,
    }))
    
    // 8. Authentication (for protected routes)
    // Applied per route group
    
    // 9. Error handler (catch errors last)
    app.Use(ErrorHandlerMiddleware())
}
```

## Best Practices

### 1. Keep Middleware Focused

Each middleware should have a single responsibility:

```go
// Good: Focused middleware
func LoggingMiddleware() fiber.Handler { /* ... */ }
func AuthMiddleware() fiber.Handler { /* ... */ }

// Bad: Does too much
func LoggingAndAuthMiddleware() fiber.Handler { /* ... */ }
```

### 2. Use Locals for Data Passing

Store data in context for downstream handlers:

```go
func UserMiddleware(db *gorm.DB) fiber.Handler {
    return func(c *fiber.Ctx) error {
        userID := c.Locals("user_id").(uint)
        
        var user User
        if err := db.First(&user, userID).Error; err != nil {
            return c.Status(404).JSON(fiber.Map{"error": "User not found"})
        }
        
        c.Locals("user", user) // Pass to handler
        return c.Next()
    }
}
```

### 3. Handle Errors Properly

Always check and handle errors:

```go
func ValidateMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        if err := validateRequest(c); err != nil {
            return c.Status(400).JSON(fiber.Map{
                "error": err.Error(),
            })
        }
        return c.Next()
    }
}
```

### 4. Use Middleware Groups

Group related middleware:

```go
func AuthGroup(jwtManager *auth.JWTManager, rbacManager *rbac.Manager) []fiber.Handler {
    return []fiber.Handler{
        auth.AuthMiddleware(jwtManager),
        rbac.RequireRole(rbacManager, "user"),
    }
}

// Usage
api := app.Group("/api", AuthGroup(jwtManager, rbacManager)...)
```

### 5. Make Middleware Testable

Write testable middleware with dependency injection:

```go
type MiddlewareConfig struct {
    DB     *gorm.DB
    Cache  cache.Cache
    Logger logger.Logger
}

func NewMiddleware(config MiddlewareConfig) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Use injected dependencies
        return c.Next()
    }
}

// Easy to test with mocks
func TestMiddleware(t *testing.T) {
    mockDB := &MockDB{}
    mockCache := &MockCache{}
    
    mw := NewMiddleware(MiddlewareConfig{
        DB:    mockDB,
        Cache: mockCache,
    })
    
    // Test middleware
}
```

### 6. Document Middleware Behavior

Always document what your middleware does:

```go
// RateLimitMiddleware limits the number of requests per client
// based on IP address or API key. When the limit is exceeded,
// it returns a 429 Too Many Requests response with a Retry-After header.
//
// Example:
//   app.Use(RateLimitMiddleware(RateLimitConfig{
//       Max:      100,
//       Duration: time.Minute,
//   }))
func RateLimitMiddleware(config RateLimitConfig) fiber.Handler {
    // ...
}
```

### 7. Performance Considerations

- Avoid heavy operations in middleware (use goroutines if needed)
- Cache compiled regexes and other resources
- Use sync.Pool for frequently allocated objects
- Consider middleware overhead in hot paths

```go
// Use sync.Pool for request validation
var validatorPool = sync.Pool{
    New: func() interface{} {
        return validator.New()
    },
}

func ValidationMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        v := validatorPool.Get().(*validator.Validate)
        defer validatorPool.Put(v)
        
        // Use validator
        return c.Next()
    }
}
```

## Complete Example

Here's a complete example combining multiple middleware:

```go
package main

import (
    "neonexcore/pkg/api"
    "neonexcore/pkg/auth"
    "neonexcore/pkg/rbac"
    "github.com/gofiber/fiber/v2"
    "time"
)

func setupApplication() *fiber.App {
    app := fiber.New(fiber.Config{
        ErrorHandler: customErrorHandler,
    })
    
    // Global middleware
    app.Use(RecoveryMiddleware())
    app.Use(api.RequestIDMiddleware())
    app.Use(api.LoggerMiddleware())
    app.Use(api.CORSMiddleware())
    app.Use(api.SecurityHeadersMiddleware())
    
    // Public routes
    app.Post("/login", loginHandler)
    app.Post("/register", registerHandler)
    
    // API routes with rate limiting
    apiGroup := app.Group("/api")
    apiGroup.Use(api.RateLimitMiddleware(api.RateLimitConfig{
        Max:      100,
        Duration: time.Minute,
    }))
    
    // Protected routes
    protected := apiGroup.Group("")
    protected.Use(auth.AuthMiddleware(jwtManager))
    protected.Get("/profile", getProfile)
    protected.Put("/profile", updateProfile)
    
    // Admin routes
    admin := apiGroup.Group("/admin")
    admin.Use(auth.AuthMiddleware(jwtManager))
    admin.Use(rbac.RequireRole(rbacManager, "admin"))
    admin.Get("/users", listUsers)
    admin.Post("/users", createUser)
    
    return app
}
```

This comprehensive middleware system provides the foundation for building secure, scalable, and maintainable APIs with NeonEx Framework.

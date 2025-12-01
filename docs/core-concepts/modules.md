# Module System

## Overview

NeonEx Framework uses a modular architecture where each feature is encapsulated in a self-contained module. Modules are the building blocks of your application, promoting separation of concerns, reusability, and maintainability.

## Module Interface

Every module must implement the `Module` interface:

```go
type Module interface {
    Name() string
    RegisterServices(*Container) error
    RegisterRoutes(fiber.Router) error
    Boot() error
}
```

### Interface Methods

- **`Name()`** - Returns the unique module name
- **`RegisterServices()`** - Registers services in the DI container
- **`RegisterRoutes()`** - Defines HTTP routes
- **`Boot()`** - Runs initialization logic after all modules are loaded

## Creating a Module

### Basic Module Structure

```go
// modules/blog/module.go
package blog

import (
    "github.com/gofiber/fiber/v2"
    "neonexcore/internal/core"
)

type BlogModule struct{}

func New() *BlogModule {
    return &BlogModule{}
}

func (m *BlogModule) Name() string {
    return "blog"
}

func (m *BlogModule) RegisterServices(c *core.Container) error {
    // Register dependencies in order
    c.Provide(func() *BlogRepository {
        db := config.DB.GetDB()
        return NewBlogRepository(db)
    }, core.Singleton)
    
    c.Provide(func() *BlogService {
        repo := core.Resolve[*BlogRepository](c)
        return NewBlogService(repo)
    }, core.Singleton)
    
    c.Provide(func() *BlogController {
        service := core.Resolve[*BlogService](c)
        return NewBlogController(service)
    }, core.Transient)
    
    return nil
}

func (m *BlogModule) RegisterRoutes(router fiber.Router) error {
    api := router.Group("/api/v1/blog")
    
    ctrl := core.Resolve[*BlogController]()
    
    api.Get("/posts", ctrl.List)
    api.Get("/posts/:id", ctrl.Get)
    api.Post("/posts", ctrl.Create)
    api.Put("/posts/:id", ctrl.Update)
    api.Delete("/posts/:id", ctrl.Delete)
    
    return nil
}

func (m *BlogModule) Boot() error {
    // Run any initialization logic
    log.Println("Blog module initialized")
    return nil
}
```

## Module Registration

Register your module in `main.go`:

```go
import (
    "neonexcore/internal/core"
    "neonexframework/modules/blog"
)

func main() {
    // Register module factory
    core.ModuleMap["blog"] = func() core.Module { 
        return blog.New() 
    }
    
    // Framework will auto-discover and load modules
    app := core.NewApp()
    app.Registry.AutoDiscover()
    app.Boot()
    app.Registry.Load()
    
    app.StartHTTP()
}
```

## Module Structure

Recommended module directory structure:

```
modules/blog/
├── module.go           # Module definition
├── models.go           # Data models
├── repository.go       # Database operations
├── service.go          # Business logic
├── controller.go       # HTTP handlers
├── dto.go              # Data Transfer Objects
├── validators.go       # Input validation
└── routes.go           # Route definitions (optional)
```

## Built-in Core Modules

### User Module

Provides user management and authentication:

```go
import coreUser "neonexcore/modules/user"

core.ModuleMap["user"] = func() core.Module { 
    return coreUser.New() 
}
```

**Features:**
- User registration and login
- Profile management
- Password reset
- Email verification
- API key generation

### Admin Module

Provides administrative features:

```go
import coreAdmin "neonexcore/modules/admin"

core.ModuleMap["admin"] = func() core.Module { 
    return coreAdmin.New() 
}
```

**Features:**
- Dashboard with statistics
- User management
- Module management
- Audit logging
- System health monitoring

## Module Lifecycle

1. **Registration** - Module registered in `ModuleMap`
2. **Discovery** - `AutoDiscover()` finds all registered modules
3. **Service Registration** - `RegisterServices()` called for each module
4. **Boot** - `Boot()` called for initialization
5. **Route Registration** - `RegisterRoutes()` sets up HTTP endpoints
6. **Load** - `Load()` completes module loading

## Module Dependencies

### Declaring Dependencies

```go
func (m *BlogModule) RegisterServices(c *core.Container) error {
    // Depend on User module services
    userService := core.Resolve[*user.UserService](c)
    
    c.Provide(func() *BlogService {
        repo := core.Resolve[*BlogRepository](c)
        return NewBlogService(repo, userService)
    }, core.Singleton)
    
    return nil
}
```

### Module Loading Order

Modules are loaded in the order they're registered. If your module depends on another, register the dependency first:

```go
// Register dependencies first
core.ModuleMap["user"] = func() core.Module { return coreUser.New() }
core.ModuleMap["admin"] = func() core.Module { return coreAdmin.New() }

// Then your modules
core.ModuleMap["blog"] = func() core.Module { return blog.New() }
```

## Best Practices

### 1. Single Responsibility

Each module should handle one domain:

```go
// ✅ Good - Single responsibility
modules/blog/
modules/comment/
modules/tag/

// ❌ Bad - Too many responsibilities
modules/content/  // handles blog, comments, tags, etc.
```

### 2. Explicit Dependencies

Always declare dependencies in `RegisterServices()`:

```go
// ✅ Good - Explicit dependency injection
c.Provide(func() *BlogService {
    repo := core.Resolve[*BlogRepository](c)
    cache := core.Resolve[*cache.CacheManager](c)
    return NewBlogService(repo, cache)
}, core.Singleton)

// ❌ Bad - Hidden dependency
c.Provide(func() *BlogService {
    // Gets cache from global variable
    return NewBlogService(repo, globalCache)
}, core.Singleton)
```

### 3. Use DTOs for API

Define Data Transfer Objects for API requests/responses:

```go
// dto.go
type CreatePostRequest struct {
    Title   string `json:"title" validate:"required,min=3,max=200"`
    Content string `json:"content" validate:"required,min=10"`
    Tags    []string `json:"tags"`
}

type PostResponse struct {
    ID        uint      `json:"id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    Author    string    `json:"author"`
    CreatedAt time.Time `json:"created_at"`
}
```

### 4. Error Handling

Return meaningful errors:

```go
func (s *BlogService) CreatePost(ctx context.Context, req *CreatePostRequest) (*Post, error) {
    if err := s.validator.Validate(req); err != nil {
        return nil, errors.NewValidationError("Invalid post data", err)
    }
    
    post := &Post{
        Title:   req.Title,
        Content: req.Content,
    }
    
    if err := s.repo.Create(ctx, post); err != nil {
        return nil, errors.NewDatabaseError("Failed to create post", err)
    }
    
    return post, nil
}
```

### 5. Use Middleware

Protect routes with middleware:

```go
func (m *BlogModule) RegisterRoutes(router fiber.Router) error {
    api := router.Group("/api/v1/blog")
    ctrl := core.Resolve[*BlogController]()
    
    // Public routes
    api.Get("/posts", ctrl.List)
    api.Get("/posts/:id", ctrl.Get)
    
    // Protected routes
    jwtManager := core.Resolve[*auth.JWTManager]()
    rbacManager := core.Resolve[*rbac.Manager]()
    
    protected := api.Group("", auth.AuthMiddleware(jwtManager))
    {
        protected.Post("/posts", 
            rbac.RequirePermission(rbacManager, "blog.create"),
            ctrl.Create,
        )
        protected.Put("/posts/:id", 
            rbac.RequirePermission(rbacManager, "blog.update"),
            ctrl.Update,
        )
        protected.Delete("/posts/:id", 
            rbac.RequirePermission(rbacManager, "blog.delete"),
            ctrl.Delete,
        )
    }
    
    return nil
}
```

## Advanced Features

### Dynamic Module Loading

Load modules conditionally:

```go
func main() {
    // Always load core modules
    core.ModuleMap["user"] = func() core.Module { return coreUser.New() }
    core.ModuleMap["admin"] = func() core.Module { return coreAdmin.New() }
    
    // Load optional modules based on config
    if config.IsFeatureEnabled("blog") {
        core.ModuleMap["blog"] = func() core.Module { return blog.New() }
    }
    
    if config.IsFeatureEnabled("ecommerce") {
        core.ModuleMap["product"] = func() core.Module { return product.New() }
        core.ModuleMap["cart"] = func() core.Module { return cart.New() }
        core.ModuleMap["order"] = func() core.Module { return order.New() }
    }
    
    app := core.NewApp()
    app.Registry.AutoDiscover()
    app.Boot()
    app.Registry.Load()
    app.StartHTTP()
}
```

### Module Configuration

Pass configuration to modules:

```go
type BlogModule struct {
    config *BlogConfig
}

type BlogConfig struct {
    PostsPerPage     int
    EnableComments   bool
    EnableDrafts     bool
    MaxUploadSize    int64
}

func NewWithConfig(config *BlogConfig) *BlogModule {
    return &BlogModule{config: config}
}

func (m *BlogModule) RegisterServices(c *core.Container) error {
    c.Provide(func() *BlogService {
        repo := core.Resolve[*BlogRepository](c)
        return NewBlogService(repo, m.config)
    }, core.Singleton)
    
    return nil
}
```

## Testing Modules

```go
func TestBlogModule(t *testing.T) {
    // Create test container
    container := core.NewContainer()
    
    // Create test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    // Create module
    module := blog.New()
    
    // Register services
    err := module.RegisterServices(container)
    assert.NoError(t, err)
    
    // Test service resolution
    service := core.Resolve[*blog.BlogService](container)
    assert.NotNil(t, service)
    
    // Test business logic
    post, err := service.CreatePost(context.Background(), &blog.CreatePostRequest{
        Title:   "Test Post",
        Content: "Test content",
    })
    assert.NoError(t, err)
    assert.NotNil(t, post)
}
```

## Common Patterns

### Repository + Service + Controller

```go
// repository.go - Data access
type BlogRepository struct {
    *database.BaseRepository[Post]
}

// service.go - Business logic
type BlogService struct {
    repo *BlogRepository
}

// controller.go - HTTP handlers
type BlogController struct {
    service *BlogService
}
```

### Event-Driven Modules

```go
func (s *BlogService) CreatePost(ctx context.Context, req *CreatePostRequest) (*Post, error) {
    post, err := s.repo.Create(ctx, &Post{...})
    if err != nil {
        return nil, err
    }
    
    // Emit event
    s.eventBus.Publish("post.created", post)
    
    return post, nil
}
```

### Module Health Checks

```go
func (m *BlogModule) Boot() error {
    // Check database connection
    if err := m.repo.Ping(); err != nil {
        return fmt.Errorf("blog module: database connection failed: %w", err)
    }
    
    // Run migrations
    if err := m.runMigrations(); err != nil {
        return fmt.Errorf("blog module: migration failed: %w", err)
    }
    
    log.Println("✓ Blog module initialized successfully")
    return nil
}
```

## Next Steps

- Learn about [Dependency Injection](dependency-injection.md)
- Explore [Routing](routing.md)
- Understand [Middleware](middleware.md)

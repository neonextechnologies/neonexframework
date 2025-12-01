# Routing

## Overview

NeonEx Framework uses [Fiber v2](https://gofiber.io/) for HTTP routing, providing Express-like routing with high performance. Routes are defined in your module's `RegisterRoutes()` method.

## Basic Routing

### Defining Routes

```go
func (m *BlogModule) RegisterRoutes(router fiber.Router) error {
    api := router.Group("/api/v1/blog")
    ctrl := core.Resolve[*BlogController]()
    
    // GET /api/v1/blog/posts
    api.Get("/posts", ctrl.List)
    
    // GET /api/v1/blog/posts/:id
    api.Get("/posts/:id", ctrl.Get)
    
    // POST /api/v1/blog/posts
    api.Post("/posts", ctrl.Create)
    
    // PUT /api/v1/blog/posts/:id
    api.Put("/posts/:id", ctrl.Update)
    
    // DELETE /api/v1/blog/posts/:id
    api.Delete("/posts/:id", ctrl.Delete)
    
    return nil
}
```

### HTTP Methods

```go
// GET - Retrieve resources
api.Get("/posts", ctrl.List)

// POST - Create resources
api.Post("/posts", ctrl.Create)

// PUT - Update entire resource
api.Put("/posts/:id", ctrl.Update)

// PATCH - Partial update
api.Patch("/posts/:id", ctrl.PartialUpdate)

// DELETE - Remove resource
api.Delete("/posts/:id", ctrl.Delete)

// HEAD - Get headers only
api.Head("/posts/:id", ctrl.Head)

// OPTIONS - Get supported methods
api.Options("/posts", ctrl.Options)
```

## Route Parameters

### Path Parameters

```go
// Route with parameter
api.Get("/posts/:id", ctrl.Get)
api.Get("/posts/:id/comments/:commentId", ctrl.GetComment)

// Controller
func (c *BlogController) Get(ctx *fiber.Ctx) error {
    id := ctx.Params("id")
    commentId := ctx.Params("commentId")
    
    // Convert to uint
    postID, err := strconv.ParseUint(id, 10, 32)
    if err != nil {
        return ctx.Status(400).JSON(fiber.Map{
            "error": "Invalid post ID",
        })
    }
    
    post, err := c.service.GetPost(ctx.Context(), uint(postID))
    if err != nil {
        return ctx.Status(404).JSON(fiber.Map{
            "error": "Post not found",
        })
    }
    
    return ctx.JSON(post)
}
```

### Query Parameters

```go
// GET /posts?page=1&limit=10&sort=created_at&order=desc
api.Get("/posts", ctrl.List)

// Controller
func (c *BlogController) List(ctx *fiber.Ctx) error {
    // Get query parameters with defaults
    page := ctx.QueryInt("page", 1)
    limit := ctx.QueryInt("limit", 10)
    sort := ctx.Query("sort", "created_at")
    order := ctx.Query("order", "desc")
    
    // Get search query
    search := ctx.Query("search", "")
    
    posts, total, err := c.service.ListPosts(ctx.Context(), &ListParams{
        Page:   page,
        Limit:  limit,
        Sort:   sort,
        Order:  order,
        Search: search,
    })
    
    if err != nil {
        return ctx.Status(500).JSON(fiber.Map{
            "error": "Failed to fetch posts",
        })
    }
    
    return ctx.JSON(fiber.Map{
        "data":  posts,
        "total": total,
        "page":  page,
        "limit": limit,
    })
}
```

### Optional Parameters

```go
// Match both /posts and /posts/:id
api.Get("/posts/:id?", ctrl.GetOrList)

func (c *BlogController) GetOrList(ctx *fiber.Ctx) error {
    id := ctx.Params("id")
    
    if id == "" {
        // List all posts
        return c.List(ctx)
    }
    
    // Get single post
    return c.Get(ctx)
}
```

### Wildcard Parameters

```go
// Match /files/documents/report.pdf
api.Get("/files/*", ctrl.ServeFile)

func (c *FileController) ServeFile(ctx *fiber.Ctx) error {
    // Get everything after /files/
    path := ctx.Params("*")  // "documents/report.pdf"
    
    return ctx.SendFile(filepath.Join("./storage", path))
}
```

## Route Groups

### Basic Grouping

```go
func (m *BlogModule) RegisterRoutes(router fiber.Router) error {
    // API v1 group
    v1 := router.Group("/api/v1")
    
    // Blog group
    blog := v1.Group("/blog")
    ctrl := core.Resolve[*BlogController]()
    
    blog.Get("/posts", ctrl.List)
    blog.Get("/posts/:id", ctrl.Get)
    blog.Post("/posts", ctrl.Create)
    
    // Comments group
    comments := blog.Group("/comments")
    commentCtrl := core.Resolve[*CommentController]()
    
    comments.Get("/", commentCtrl.List)
    comments.Post("/", commentCtrl.Create)
    
    return nil
}
```

### Nested Groups

```go
api := router.Group("/api")
{
    v1 := api.Group("/v1")
    {
        users := v1.Group("/users")
        {
            users.Get("/", userCtrl.List)
            users.Post("/", userCtrl.Create)
            
            user := users.Group("/:id")
            {
                user.Get("/", userCtrl.Get)
                user.Put("/", userCtrl.Update)
                user.Delete("/", userCtrl.Delete)
                
                user.Get("/posts", userCtrl.GetPosts)
                user.Get("/comments", userCtrl.GetComments)
            }
        }
    }
}
```

## Middleware

### Group Middleware

```go
func (m *BlogModule) RegisterRoutes(router fiber.Router) error {
    api := router.Group("/api/v1/blog")
    ctrl := core.Resolve[*BlogController]()
    jwtManager := core.Resolve[*auth.JWTManager]()
    rbacManager := core.Resolve[*rbac.Manager]()
    
    // Public routes (no middleware)
    api.Get("/posts", ctrl.List)
    api.Get("/posts/:id", ctrl.Get)
    
    // Protected routes (with middleware)
    protected := api.Group("", auth.AuthMiddleware(jwtManager))
    {
        // Create requires authentication
        protected.Post("/posts", ctrl.Create)
        
        // Update requires authentication + permission
        protected.Put("/posts/:id",
            rbac.RequirePermission(rbacManager, "blog.update"),
            ctrl.Update,
        )
        
        // Delete requires authentication + permission
        protected.Delete("/posts/:id",
            rbac.RequirePermission(rbacManager, "blog.delete"),
            ctrl.Delete,
        )
    }
    
    return nil
}
```

### Route-Specific Middleware

```go
// Single middleware
api.Get("/posts", 
    middleware.RateLimiter(),
    ctrl.List,
)

// Multiple middleware
api.Post("/posts",
    auth.AuthMiddleware(jwtManager),
    rbac.RequirePermission(rbacManager, "blog.create"),
    middleware.ValidateRequest(&CreatePostRequest{}),
    ctrl.Create,
)
```

## Real-World Examples

### User Module Routes

```go
func (m *UserModule) Routes(app *fiber.App, c *core.Container) {
    authCtrl := core.Resolve[*AuthController](c)
    userCtrl := core.Resolve[*UserController](c)
    jwtManager := core.Resolve[*auth.JWTManager](c)
    rbacManager := core.Resolve[*rbac.Manager](c)

    api := app.Group("/api/v1")

    // ==================== Authentication Routes (Public) ====================
    authGroup := api.Group("/auth")
    {
        // Public endpoints
        authGroup.Post("/login", authCtrl.Login)
        authGroup.Post("/register", authCtrl.Register)
        authGroup.Post("/refresh", authCtrl.RefreshToken)
        authGroup.Post("/forgot-password", authCtrl.ForgotPassword)
        authGroup.Post("/reset-password", authCtrl.ResetPassword)
        authGroup.Get("/verify-email/:token", authCtrl.VerifyEmail)

        // Protected endpoints
        authProtected := authGroup.Group("", auth.AuthMiddleware(jwtManager))
        {
            authProtected.Post("/logout", authCtrl.Logout)
            authProtected.Get("/profile", authCtrl.GetProfile)
            authProtected.Put("/profile", authCtrl.UpdateProfile)
            authProtected.Post("/change-password", authCtrl.ChangePassword)
            authProtected.Post("/api-key", authCtrl.GenerateAPIKey)
        }
    }

    // ==================== User Management Routes ====================
    usersGroup := api.Group("/users")
    {
        // Public search
        usersGroup.Get("/search", userCtrl.Search)

        // Protected endpoints
        usersProtected := usersGroup.Group("", auth.AuthMiddleware(jwtManager))
        {
            // Read operations
            usersProtected.Get("/", 
                rbac.RequirePermission(rbacManager, "users.read"),
                userCtrl.GetAll,
            )
            usersProtected.Get("/:id", 
                rbac.RequirePermission(rbacManager, "users.read"),
                userCtrl.GetByID,
            )

            // Write operations
            usersProtected.Post("/", 
                rbac.RequirePermission(rbacManager, "users.create"),
                userCtrl.Create,
            )
            usersProtected.Put("/:id", 
                rbac.RequirePermission(rbacManager, "users.update"),
                userCtrl.Update,
            )
            usersProtected.Delete("/:id", 
                rbac.RequirePermission(rbacManager, "users.delete"),
                userCtrl.Delete,
            )
        }
    }
}
```

### Admin Module Routes

```go
func SetupRoutes(router fiber.Router, container *core.Container) {
    controller := container.GetSingleton("admin.controller").(*Controller)
    rbacManager := container.GetSingleton("rbac.manager").(*rbac.Manager)
    jwtManager := container.GetSingleton("jwt.manager").(*auth.JWTManager)

    admin := router.Group("/admin", auth.AuthMiddleware(jwtManager))

    // Dashboard
    admin.Get("/dashboard", 
        rbac.RequirePermission(rbacManager, "admin.dashboard.view"),
        controller.GetDashboard,
    )

    // Statistics
    admin.Get("/stats", 
        rbac.RequirePermission(rbacManager, "admin.system.view"),
        controller.GetStats,
    )
    admin.Get("/stats/users", 
        rbac.RequirePermission(rbacManager, "admin.system.view"),
        controller.GetUserStats,
    )
    admin.Get("/stats/modules", 
        rbac.RequirePermission(rbacManager, "admin.system.view"),
        controller.GetModuleStats,
    )

    // System health
    admin.Get("/health", controller.GetHealth)

    // Audit logs
    admin.Get("/audit-logs", 
        rbac.RequirePermission(rbacManager, "admin.audit.view"),
        controller.GetAuditLogs,
    )
    admin.Post("/audit-logs", 
        rbac.RequirePermission(rbacManager, "admin.audit.create"),
        controller.CreateAuditLog,
    )
}
```

## Route Naming

```go
// Name routes for URL generation
api.Get("/posts/:id", ctrl.Get).Name("posts.show")
api.Post("/posts", ctrl.Create).Name("posts.create")

// Generate URL
url, err := app.GetRoute("posts.show").URL(fiber.Map{"id": "123"})
// Returns: /api/v1/blog/posts/123
```

## Route Constraints

```go
// Only match numeric IDs
api.Get("/posts/:id<int>", ctrl.Get)

// Only match alphanumeric slugs
api.Get("/posts/:slug<alpha>", ctrl.GetBySlug)

// Custom regex
api.Get("/posts/:year<regex(\\d{4})>/:month<regex(\\d{2})>", ctrl.GetArchive)
```

## RESTful Routes

### Standard REST API

```go
func (m *BlogModule) RegisterRoutes(router fiber.Router) error {
    api := router.Group("/api/v1")
    posts := api.Group("/posts")
    ctrl := core.Resolve[*BlogController]()
    
    // GET /posts - List all
    posts.Get("/", ctrl.Index)
    
    // POST /posts - Create new
    posts.Post("/", ctrl.Store)
    
    // GET /posts/:id - Show single
    posts.Get("/:id", ctrl.Show)
    
    // PUT /posts/:id - Update entire
    posts.Put("/:id", ctrl.Update)
    
    // PATCH /posts/:id - Partial update
    posts.Patch("/:id", ctrl.Patch)
    
    // DELETE /posts/:id - Delete
    posts.Delete("/:id", ctrl.Destroy)
    
    return nil
}
```

### Nested Resources

```go
// Posts have many comments
posts := api.Group("/posts")
{
    posts.Get("/", postCtrl.Index)
    posts.Post("/", postCtrl.Store)
    posts.Get("/:id", postCtrl.Show)
    
    // Nested comments
    comments := posts.Group("/:postId/comments")
    {
        // GET /posts/:postId/comments
        comments.Get("/", commentCtrl.Index)
        
        // POST /posts/:postId/comments
        comments.Post("/", commentCtrl.Store)
        
        // GET /posts/:postId/comments/:id
        comments.Get("/:id", commentCtrl.Show)
        
        // PUT /posts/:postId/comments/:id
        comments.Put("/:id", commentCtrl.Update)
        
        // DELETE /posts/:postId/comments/:id
        comments.Delete("/:id", commentCtrl.Destroy)
    }
}
```

## API Versioning

### URL Versioning

```go
// Version 1
v1 := router.Group("/api/v1")
{
    v1.Get("/users", v1UserCtrl.List)
}

// Version 2
v2 := router.Group("/api/v2")
{
    v2.Get("/users", v2UserCtrl.List)  // Different implementation
}
```

### Header Versioning

```go
api := router.Group("/api")

api.Use(func(c *fiber.Ctx) error {
    version := c.Get("API-Version", "v1")
    c.Locals("api_version", version)
    return c.Next()
})

api.Get("/users", func(c *fiber.Ctx) error {
    version := c.Locals("api_version").(string)
    
    switch version {
    case "v1":
        return v1UserCtrl.List(c)
    case "v2":
        return v2UserCtrl.List(c)
    default:
        return c.Status(400).JSON(fiber.Map{
            "error": "Unsupported API version",
        })
    }
})
```

## Error Handling in Routes

```go
func (c *BlogController) Get(ctx *fiber.Ctx) error {
    id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
    if err != nil {
        return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid post ID",
        })
    }
    
    post, err := c.service.GetPost(ctx.Context(), uint(id))
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
                "error": "Post not found",
            })
        }
        return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Internal server error",
        })
    }
    
    return ctx.JSON(post)
}
```

## Route Testing

```go
func TestBlogRoutes(t *testing.T) {
    // Setup
    app := fiber.New()
    module := blog.New()
    container := setupTestContainer(t)
    
    // Register routes
    module.RegisterRoutes(app)
    
    // Test GET /posts
    req := httptest.NewRequest("GET", "/api/v1/blog/posts", nil)
    resp, err := app.Test(req)
    
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
    
    // Test POST /posts
    body := `{"title":"Test Post","content":"Test content"}`
    req = httptest.NewRequest("POST", "/api/v1/blog/posts", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    resp, err = app.Test(req)
    
    assert.NoError(t, err)
    assert.Equal(t, 201, resp.StatusCode)
}
```

## Best Practices

### 1. Use Route Groups

```go
// ✅ Good - Organized with groups
api := router.Group("/api/v1")
blog := api.Group("/blog")
blog.Get("/posts", ctrl.List)

// ❌ Bad - Repetitive paths
router.Get("/api/v1/blog/posts", ctrl.List)
router.Post("/api/v1/blog/posts", ctrl.Create)
```

### 2. Apply Middleware at Group Level

```go
// ✅ Good - Middleware on group
protected := api.Group("", auth.AuthMiddleware(jwt))
protected.Post("/posts", ctrl.Create)
protected.Put("/posts/:id", ctrl.Update)

// ❌ Bad - Repeat middleware
api.Post("/posts", auth.AuthMiddleware(jwt), ctrl.Create)
api.Put("/posts/:id", auth.AuthMiddleware(jwt), ctrl.Update)
```

### 3. Use Consistent Naming

```go
// ✅ Good - RESTful naming
GET    /posts          - List
POST   /posts          - Create
GET    /posts/:id      - Show
PUT    /posts/:id      - Update
DELETE /posts/:id      - Delete

// ❌ Bad - Inconsistent
GET /getPosts
POST /createPost
GET /showPost/:id
```

### 4. Version Your APIs

```go
// ✅ Good - Versioned
/api/v1/users
/api/v2/users

// ❌ Bad - No version
/api/users
```

## Next Steps

- Learn about [Middleware](middleware.md)
- Explore [Request & Response](request-response.md)
- Understand [Validation](validation.md)

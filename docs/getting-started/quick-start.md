# Quick Start

This guide will get you up and running with NeonEx Framework in under 10 minutes.

---

## Step 1: Create New Project

```bash
# Clone the framework
git clone https://github.com/neonextechnologies/neonexframework.git my-app
cd my-app

# Install dependencies
go mod download
```

---

## Step 2: Configure Environment

```bash
# Copy environment template
cp .env.example .env
```

Edit `.env`:

```env
APP_NAME=MyApp
APP_PORT=8080
APP_DEBUG=true

DB_CONNECTION=postgres
DB_HOST=localhost
DB_PORT=5432
DB_DATABASE=myappdb
DB_USERNAME=postgres
DB_PASSWORD=secret

JWT_SECRET=change-this-to-random-string
```

---

## Step 3: Create Database

```bash
# PostgreSQL
createdb myappdb

# Or using psql
psql -U postgres
CREATE DATABASE myappdb;
\q
```

---

## Step 4: Run Application

```bash
# Start the server
go run main.go
```

You should see:

```
✓ Database connected
✓ Redis connected (optional)
✓ Modules loaded: admin, user, frontend, web
✓ Server started on :8080
```

---

## Step 5: Test the API

### Check Health

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy",
  "database": "connected"
}
```

### Register a User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "secret123"
  }'
```

Response:
```json
{
  "success": true,
  "data": {
    "user": {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com"
    },
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "secret123"
  }'
```

### Access Protected Route

```bash
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

---

## Step 6: Create Your First Module

Let's create a simple blog module.

### Generate Module Structure

```bash
# Create module directory
mkdir -p modules/blog/pkg
```

### Create Module File

`modules/blog/module.go`:

```go
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
    c.Provide(NewPostService)
    c.Provide(NewPostController)
    return nil
}

func (m *BlogModule) RegisterRoutes(router fiber.Router) error {
    api := router.Group("/api/v1/blog")
    
    ctrl := core.Resolve[*PostController]()
    
    api.Get("/posts", ctrl.List)
    api.Get("/posts/:id", ctrl.Get)
    api.Post("/posts", ctrl.Create)
    
    return nil
}

func (m *BlogModule) Boot() error {
    return nil
}
```

### Create Model

`modules/blog/pkg/post.go`:

```go
package blog

import "gorm.io/gorm"

type Post struct {
    gorm.Model
    Title   string `json:"title" gorm:"not null"`
    Content string `json:"content" gorm:"type:text"`
    Author  string `json:"author"`
}
```

### Create Service

`modules/blog/pkg/service.go`:

```go
package blog

import (
    "context"
    "gorm.io/gorm"
)

type PostService struct {
    db *gorm.DB
}

func NewPostService(db *gorm.DB) *PostService {
    return &PostService{db: db}
}

func (s *PostService) FindAll(ctx context.Context) ([]Post, error) {
    var posts []Post
    err := s.db.WithContext(ctx).Find(&posts).Error
    return posts, err
}

func (s *PostService) Create(ctx context.Context, post *Post) error {
    return s.db.WithContext(ctx).Create(post).Error
}

func (s *PostService) FindByID(ctx context.Context, id uint) (*Post, error) {
    var post Post
    err := s.db.WithContext(ctx).First(&post, id).Error
    return &post, err
}
```

### Create Controller

`modules/blog/pkg/controller.go`:

```go
package blog

import (
    "github.com/gofiber/fiber/v2"
)

type PostController struct {
    service *PostService
}

func NewPostController(service *PostService) *PostController {
    return &PostController{service: service}
}

func (c *PostController) List(ctx *fiber.Ctx) error {
    posts, err := c.service.FindAll(ctx.Context())
    if err != nil {
        return ctx.Status(500).JSON(fiber.Map{
            "error": err.Error(),
        })
    }
    
    return ctx.JSON(fiber.Map{
        "success": true,
        "data":    posts,
    })
}

func (c *PostController) Get(ctx *fiber.Ctx) error {
    id, _ := ctx.ParamsInt("id")
    
    post, err := c.service.FindByID(ctx.Context(), uint(id))
    if err != nil {
        return ctx.Status(404).JSON(fiber.Map{
            "error": "Post not found",
        })
    }
    
    return ctx.JSON(fiber.Map{
        "success": true,
        "data":    post,
    })
}

func (c *PostController) Create(ctx *fiber.Ctx) error {
    var post Post
    
    if err := ctx.BodyParser(&post); err != nil {
        return ctx.Status(400).JSON(fiber.Map{
            "error": "Invalid request",
        })
    }
    
    if err := c.service.Create(ctx.Context(), &post); err != nil {
        return ctx.Status(500).JSON(fiber.Map{
            "error": err.Error(),
        })
    }
    
    return ctx.Status(201).JSON(fiber.Map{
        "success": true,
        "data":    post,
    })
}
```

### Register Module

Update `main.go`:

```go
package main

import (
    "neonexcore/internal/core"
    "your-app/modules/blog"  // Add this import
)

func main() {
    // Initialize app
    app := core.NewApp()
    
    // Register custom modules
    core.ModuleMap["blog"] = blog.New()
    
    // Run app
    app.Run()
}
```

---

## Step 7: Test Your Module

Restart the application:

```bash
go run main.go
```

### Create a Post

```bash
curl -X POST http://localhost:8080/api/v1/blog/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My First Post",
    "content": "Hello, NeonEx!",
    "author": "John Doe"
  }'
```

### Get All Posts

```bash
curl http://localhost:8080/api/v1/blog/posts
```

### Get Single Post

```bash
curl http://localhost:8080/api/v1/blog/posts/1
```

---

## Step 8: Add Authentication

Protect routes with authentication:

```go
func (m *BlogModule) RegisterRoutes(router fiber.Router) error {
    api := router.Group("/api/v1/blog")
    
    ctrl := core.Resolve[*PostController]()
    
    // Public routes
    api.Get("/posts", ctrl.List)
    api.Get("/posts/:id", ctrl.Get)
    
    // Protected routes
    authenticated := api.Group("", middleware.Auth())
    authenticated.Post("/posts", ctrl.Create)
    authenticated.Put("/posts/:id", ctrl.Update)
    authenticated.Delete("/posts/:id", ctrl.Delete)
    
    return nil
}
```

Now creating posts requires authentication:

```bash
curl -X POST http://localhost:8080/api/v1/blog/posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "title": "Authenticated Post",
    "content": "This is protected!"
  }'
```

---

## Step 9: Add Validation

Add validation to your controller:

```go
type CreatePostRequest struct {
    Title   string `json:"title" validate:"required,min=3,max=200"`
    Content string `json:"content" validate:"required,min=10"`
    Author  string `json:"author" validate:"required"`
}

func (c *PostController) Create(ctx *fiber.Ctx) error {
    var req CreatePostRequest
    
    if err := ctx.BodyParser(&req); err != nil {
        return ctx.Status(400).JSON(fiber.Map{
            "error": "Invalid request",
        })
    }
    
    // Validate
    if err := validator.Validate(req); err != nil {
        return ctx.Status(422).JSON(fiber.Map{
            "error": "Validation failed",
            "details": err,
        })
    }
    
    post := &Post{
        Title:   req.Title,
        Content: req.Content,
        Author:  req.Author,
    }
    
    if err := c.service.Create(ctx.Context(), post); err != nil {
        return ctx.Status(500).JSON(fiber.Map{
            "error": err.Error(),
        })
    }
    
    return ctx.Status(201).JSON(fiber.Map{
        "success": true,
        "data":    post,
    })
}
```

---

## What's Next?

Congratulations! You've built your first NeonEx application with:
- ✅ Database connection
- ✅ RESTful API
- ✅ Custom module
- ✅ Authentication
- ✅ Validation

### Continue Learning

- 📖 [Configuration Guide](./configuration.md) - Configure your app
- 🏗️ [Project Structure](./project-structure.md) - Understand the architecture
- 📚 [Core Concepts](../core-concepts/modules.md) - Deep dive into modules
- 🔐 [Authentication](../security/authentication.md) - Advanced auth features
- 💾 [Database](../database/configuration.md) - Database best practices

### Build Something Real

- 🎓 [Building a Blog](../tutorials/building-blog.md)
- 🛒 [E-commerce API](../tutorials/rest-api.md)
- 💬 [Real-time Chat](../tutorials/realtime-chat.md)
- 📊 [Admin Dashboard](../tutorials/admin-dashboard.md)

---

## Need Help?

- 💬 [GitHub Discussions](https://github.com/neonextechnologies/neonexframework/discussions)
- 🐛 [Report Issues](https://github.com/neonextechnologies/neonexframework/issues)
- 📧 Email: support@neonexframework.dev

# First Application

Build a complete REST API for a todo application from scratch.

---

## What We'll Build

A todo REST API with:
- ✅ CRUD operations
- ✅ User authentication
- ✅ Authorization
- ✅ Validation
- ✅ Database persistence

**Endpoints:**
- `POST /api/v1/todos` - Create todo
- `GET /api/v1/todos` - List todos
- `GET /api/v1/todos/:id` - Get single todo
- `PUT /api/v1/todos/:id` - Update todo
- `DELETE /api/v1/todos/:id` - Delete todo
- `POST /api/v1/todos/:id/complete` - Mark as complete

---

## Step 1: Setup Project

```bash
# Clone or create project
git clone https://github.com/neonextechnologies/neonexframework.git todo-api
cd todo-api

# Install dependencies
go mod download

# Setup environment
cp .env.example .env

# Create database
createdb tododb
```

Edit `.env`:
```env
APP_NAME=TodoAPI
DB_DATABASE=tododb
DB_USERNAME=postgres
DB_PASSWORD=secret
```

---

## Step 2: Create Todo Module

### Directory Structure

```bash
mkdir -p modules/todo/pkg
```

### Create Model

`modules/todo/pkg/todo.go`:

```go
package todo

import (
    "time"
    "gorm.io/gorm"
)

type Todo struct {
    ID          uint           `json:"id" gorm:"primaryKey"`
    Title       string         `json:"title" gorm:"not null"`
    Description string         `json:"description"`
    Completed   bool           `json:"completed" gorm:"default:false"`
    UserID      uint           `json:"user_id" gorm:"not null"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies table name
func (Todo) TableName() string {
    return "todos"
}
```

### Create DTOs

`modules/todo/pkg/dto.go`:

```go
package todo

type CreateTodoDTO struct {
    Title       string `json:"title" validate:"required,min=3,max=200"`
    Description string `json:"description" validate:"max=1000"`
}

type UpdateTodoDTO struct {
    Title       *string `json:"title" validate:"omitempty,min=3,max=200"`
    Description *string `json:"description" validate:"omitempty,max=1000"`
    Completed   *bool   `json:"completed"`
}

type TodoResponse struct {
    ID          uint      `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Completed   bool      `json:"completed"`
    CreatedAt   time.Time `json:"created_at"`
}

func (t *Todo) ToResponse() *TodoResponse {
    return &TodoResponse{
        ID:          t.ID,
        Title:       t.Title,
        Description: t.Description,
        Completed:   t.Completed,
        CreatedAt:   t.CreatedAt,
    }
}
```

---

## Step 3: Create Repository

`modules/todo/pkg/repository.go`:

```go
package todo

import (
    "context"
    "gorm.io/gorm"
)

type TodoRepository struct {
    db *gorm.DB
}

func NewTodoRepository(db *gorm.DB) *TodoRepository {
    return &TodoRepository{db: db}
}

func (r *TodoRepository) Create(ctx context.Context, todo *Todo) error {
    return r.db.WithContext(ctx).Create(todo).Error
}

func (r *TodoRepository) FindAll(ctx context.Context, userID uint) ([]Todo, error) {
    var todos []Todo
    err := r.db.WithContext(ctx).
        Where("user_id = ?", userID).
        Order("created_at DESC").
        Find(&todos).Error
    return todos, err
}

func (r *TodoRepository) FindByID(ctx context.Context, id, userID uint) (*Todo, error) {
    var todo Todo
    err := r.db.WithContext(ctx).
        Where("id = ? AND user_id = ?", id, userID).
        First(&todo).Error
    return &todo, err
}

func (r *TodoRepository) Update(ctx context.Context, todo *Todo) error {
    return r.db.WithContext(ctx).Save(todo).Error
}

func (r *TodoRepository) Delete(ctx context.Context, id, userID uint) error {
    return r.db.WithContext(ctx).
        Where("id = ? AND user_id = ?", id, userID).
        Delete(&Todo{}).Error
}
```

---

## Step 4: Create Service

`modules/todo/pkg/service.go`:

```go
package todo

import (
    "context"
    "errors"
)

type TodoService struct {
    repo *TodoRepository
}

func NewTodoService(repo *TodoRepository) *TodoService {
    return &TodoService{repo: repo}
}

func (s *TodoService) Create(ctx context.Context, dto *CreateTodoDTO, userID uint) (*Todo, error) {
    todo := &Todo{
        Title:       dto.Title,
        Description: dto.Description,
        UserID:      userID,
    }
    
    if err := s.repo.Create(ctx, todo); err != nil {
        return nil, err
    }
    
    return todo, nil
}

func (s *TodoService) List(ctx context.Context, userID uint) ([]Todo, error) {
    return s.repo.FindAll(ctx, userID)
}

func (s *TodoService) Get(ctx context.Context, id, userID uint) (*Todo, error) {
    return s.repo.FindByID(ctx, id, userID)
}

func (s *TodoService) Update(ctx context.Context, id uint, dto *UpdateTodoDTO, userID uint) (*Todo, error) {
    todo, err := s.repo.FindByID(ctx, id, userID)
    if err != nil {
        return nil, err
    }
    
    if dto.Title != nil {
        todo.Title = *dto.Title
    }
    if dto.Description != nil {
        todo.Description = *dto.Description
    }
    if dto.Completed != nil {
        todo.Completed = *dto.Completed
    }
    
    if err := s.repo.Update(ctx, todo); err != nil {
        return nil, err
    }
    
    return todo, nil
}

func (s *TodoService) Delete(ctx context.Context, id, userID uint) error {
    return s.repo.Delete(ctx, id, userID)
}

func (s *TodoService) Complete(ctx context.Context, id, userID uint) (*Todo, error) {
    todo, err := s.repo.FindByID(ctx, id, userID)
    if err != nil {
        return nil, err
    }
    
    if todo.Completed {
        return nil, errors.New("todo already completed")
    }
    
    todo.Completed = true
    
    if err := s.repo.Update(ctx, todo); err != nil {
        return nil, err
    }
    
    return todo, nil
}
```

---

## Step 5: Create Controller

`modules/todo/pkg/controller.go`:

```go
package todo

import (
    "github.com/gofiber/fiber/v2"
    "neonexcore/pkg/validator"
)

type TodoController struct {
    service *TodoService
}

func NewTodoController(service *TodoService) *TodoController {
    return &TodoController{service: service}
}

func (c *TodoController) Create(ctx *fiber.Ctx) error {
    var dto CreateTodoDTO
    
    if err := ctx.BodyParser(&dto); err != nil {
        return ctx.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }
    
    if err := validator.Validate(&dto); err != nil {
        return ctx.Status(422).JSON(fiber.Map{
            "error":   "Validation failed",
            "details": err,
        })
    }
    
    // Get user ID from context (set by auth middleware)
    userID := ctx.Locals("user_id").(uint)
    
    todo, err := c.service.Create(ctx.Context(), &dto, userID)
    if err != nil {
        return ctx.Status(500).JSON(fiber.Map{
            "error": "Failed to create todo",
        })
    }
    
    return ctx.Status(201).JSON(fiber.Map{
        "success": true,
        "data":    todo.ToResponse(),
    })
}

func (c *TodoController) List(ctx *fiber.Ctx) error {
    userID := ctx.Locals("user_id").(uint)
    
    todos, err := c.service.List(ctx.Context(), userID)
    if err != nil {
        return ctx.Status(500).JSON(fiber.Map{
            "error": "Failed to fetch todos",
        })
    }
    
    responses := make([]*TodoResponse, len(todos))
    for i, todo := range todos {
        responses[i] = todo.ToResponse()
    }
    
    return ctx.JSON(fiber.Map{
        "success": true,
        "data":    responses,
    })
}

func (c *TodoController) Get(ctx *fiber.Ctx) error {
    id, err := ctx.ParamsInt("id")
    if err != nil {
        return ctx.Status(400).JSON(fiber.Map{
            "error": "Invalid todo ID",
        })
    }
    
    userID := ctx.Locals("user_id").(uint)
    
    todo, err := c.service.Get(ctx.Context(), uint(id), userID)
    if err != nil {
        return ctx.Status(404).JSON(fiber.Map{
            "error": "Todo not found",
        })
    }
    
    return ctx.JSON(fiber.Map{
        "success": true,
        "data":    todo.ToResponse(),
    })
}

func (c *TodoController) Update(ctx *fiber.Ctx) error {
    id, err := ctx.ParamsInt("id")
    if err != nil {
        return ctx.Status(400).JSON(fiber.Map{
            "error": "Invalid todo ID",
        })
    }
    
    var dto UpdateTodoDTO
    if err := ctx.BodyParser(&dto); err != nil {
        return ctx.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }
    
    if err := validator.Validate(&dto); err != nil {
        return ctx.Status(422).JSON(fiber.Map{
            "error":   "Validation failed",
            "details": err,
        })
    }
    
    userID := ctx.Locals("user_id").(uint)
    
    todo, err := c.service.Update(ctx.Context(), uint(id), &dto, userID)
    if err != nil {
        return ctx.Status(500).JSON(fiber.Map{
            "error": "Failed to update todo",
        })
    }
    
    return ctx.JSON(fiber.Map{
        "success": true,
        "data":    todo.ToResponse(),
    })
}

func (c *TodoController) Delete(ctx *fiber.Ctx) error {
    id, err := ctx.ParamsInt("id")
    if err != nil {
        return ctx.Status(400).JSON(fiber.Map{
            "error": "Invalid todo ID",
        })
    }
    
    userID := ctx.Locals("user_id").(uint)
    
    if err := c.service.Delete(ctx.Context(), uint(id), userID); err != nil {
        return ctx.Status(500).JSON(fiber.Map{
            "error": "Failed to delete todo",
        })
    }
    
    return ctx.JSON(fiber.Map{
        "success": true,
        "message": "Todo deleted successfully",
    })
}

func (c *TodoController) Complete(ctx *fiber.Ctx) error {
    id, err := ctx.ParamsInt("id")
    if err != nil {
        return ctx.Status(400).JSON(fiber.Map{
            "error": "Invalid todo ID",
        })
    }
    
    userID := ctx.Locals("user_id").(uint)
    
    todo, err := c.service.Complete(ctx.Context(), uint(id), userID)
    if err != nil {
        return ctx.Status(400).JSON(fiber.Map{
            "error": err.Error(),
        })
    }
    
    return ctx.JSON(fiber.Map{
        "success": true,
        "data":    todo.ToResponse(),
    })
}
```

---

## Step 6: Create Module Definition

`modules/todo/module.go`:

```go
package todo

import (
    "github.com/gofiber/fiber/v2"
    "neonexcore/internal/core"
    "neonexcore/internal/middleware"
    "modules/todo/pkg"
)

type TodoModule struct{}

func New() *TodoModule {
    return &TodoModule{}
}

func (m *TodoModule) Name() string {
    return "todo"
}

func (m *TodoModule) RegisterServices(c *core.Container) error {
    c.Provide(todo.NewTodoRepository)
    c.Provide(todo.NewTodoService)
    c.Provide(todo.NewTodoController)
    return nil
}

func (m *TodoModule) RegisterRoutes(router fiber.Router) error {
    api := router.Group("/api/v1/todos")
    
    // Apply authentication middleware
    api.Use(middleware.Auth())
    
    ctrl := core.Resolve[*todo.TodoController]()
    
    api.Post("/", ctrl.Create)
    api.Get("/", ctrl.List)
    api.Get("/:id", ctrl.Get)
    api.Put("/:id", ctrl.Update)
    api.Delete("/:id", ctrl.Delete)
    api.Post("/:id/complete", ctrl.Complete)
    
    return nil
}

func (m *TodoModule) Boot() error {
    // Run migrations
    db := core.ResolveDB()
    return db.AutoMigrate(&todo.Todo{})
}
```

---

## Step 7: Register Module

Update `main.go`:

```go
package main

import (
    "neonexcore/internal/core"
    "your-app/modules/todo"
)

func main() {
    app := core.NewApp()
    
    // Register todo module
    core.ModuleMap["todo"] = todo.New()
    
    app.Run()
}
```

---

## Step 8: Test the API

### Start Server

```bash
go run main.go
```

### Register User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "secret123"
  }'
```

Save the token from response!

### Create Todo

```bash
curl -X POST http://localhost:8080/api/v1/todos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "title": "Buy groceries",
    "description": "Milk, eggs, bread"
  }'
```

### List Todos

```bash
curl -X GET http://localhost:8080/api/v1/todos \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Get Single Todo

```bash
curl -X GET http://localhost:8080/api/v1/todos/1 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Update Todo

```bash
curl -X PUT http://localhost:8080/api/v1/todos/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "title": "Buy groceries (Updated)",
    "completed": false
  }'
```

### Mark as Complete

```bash
curl -X POST http://localhost:8080/api/v1/todos/1/complete \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Delete Todo

```bash
curl -X DELETE http://localhost:8080/api/v1/todos/1 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## Step 9: Add Tests

`modules/todo/pkg/service_test.go`:

```go
package todo_test

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "modules/todo/pkg"
)

type MockTodoRepository struct {
    mock.Mock
}

func (m *MockTodoRepository) Create(ctx context.Context, todo *todo.Todo) error {
    args := m.Called(ctx, todo)
    return args.Error(0)
}

func TestTodoService_Create(t *testing.T) {
    mockRepo := new(MockTodoRepository)
    service := todo.NewTodoService(mockRepo)
    
    dto := &todo.CreateTodoDTO{
        Title:       "Test Todo",
        Description: "Test Description",
    }
    
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
    
    result, err := service.Create(context.Background(), dto, 1)
    
    assert.NoError(t, err)
    assert.Equal(t, "Test Todo", result.Title)
    mockRepo.AssertExpectations(t)
}
```

Run tests:

```bash
go test ./modules/todo/...
```

---

## Congratulations! 🎉

You've built a complete REST API with:
- ✅ CRUD operations
- ✅ Authentication
- ✅ Authorization (user can only access their todos)
- ✅ Validation
- ✅ Error handling
- ✅ Tests

---

## Next Steps

- [Testing Guide](../testing/introduction.md) - Write more tests
- [API Documentation](../api-reference/core-api.md) - Learn more APIs
- [Deployment](../deployment/overview.md) - Deploy your app
- [Advanced Topics](../advanced/websockets.md) - WebSockets, GraphQL, etc.

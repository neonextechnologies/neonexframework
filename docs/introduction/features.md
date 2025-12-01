# Features

NeonEx Framework provides a comprehensive set of features for building modern web applications.

---

## Core Features

### 🏗️ Modular Architecture

**Self-Contained Modules**
- Auto-discovery of modules
- Dependency injection
- Lifecycle hooks
- Service registration
- Route registration

```go
type BlogModule struct{}

func (m *BlogModule) Name() string { return "blog" }
func (m *BlogModule) RegisterServices(c *core.Container) error { ... }
func (m *BlogModule) RegisterRoutes(r fiber.Router) error { ... }
func (m *BlogModule) Boot() error { ... }
```

**Benefits:**
- Clean code organization
- Easy to add/remove features
- Reusable across projects
- Independent testing

---

### 💉 Dependency Injection

**Type-Safe DI Container**
- Constructor injection
- Auto-resolution
- Circular dependency detection
- Scoped lifetimes

```go
// Register
container.Provide(NewUserService)
container.Provide(NewUserRepository)

// Resolve with dependencies injected
service := container.Resolve[*UserService]()
```

---

### 🚀 High Performance

**Built for Speed**
- 10,000+ requests/second
- Sub-millisecond latency
- Efficient memory usage
- Goroutine-based concurrency

**Performance Features:**
- Connection pooling
- Query caching
- Asset minification
- Gzip compression

---

### 🔐 Authentication & Authorization

**Complete Security System**
- JWT token authentication
- Role-Based Access Control (RBAC)
- Permission system
- Password hashing (bcrypt)
- API key support

```go
// Protect routes with authentication
api.Use(middleware.Auth())

// Check permissions
api.Use(middleware.RequirePermission("posts.create"))

// Role-based access
api.Use(middleware.RequireRole("admin"))
```

---

### 💾 Database & ORM

**Powered by GORM**
- Multiple database support (PostgreSQL, MySQL, SQLite)
- Auto-migrations
- Repository pattern
- Query builder
- Relationships (One-to-One, One-to-Many, Many-to-Many)
- Transactions

```go
// Repository pattern
repo := database.NewBaseRepository[User](db)

users, _ := repo.FindAll(ctx)
user, _ := repo.FindByID(ctx, 1)
repo.Create(ctx, &newUser)
repo.Update(ctx, &user)
repo.Delete(ctx, 1)

// With conditions
users, _ := repo.FindWhere(ctx, "active = ?", true)

// Pagination
users, total, _ := repo.Paginate(ctx, page, perPage)
```

---

### 🌐 HTTP & Routing

**Built on Fiber v2**
- Fast routing
- Route groups
- Parameter binding
- Query parameters
- Path parameters
- Request validation
- File uploads

```go
api := app.Group("/api/v1")

// RESTful routes
api.Get("/users", controller.List)
api.Get("/users/:id", controller.Get)
api.Post("/users", controller.Create)
api.Put("/users/:id", controller.Update)
api.Delete("/users/:id", controller.Delete)

// Route with middleware
api.Get("/admin/users", 
    middleware.Auth(),
    middleware.RequireRole("admin"),
    controller.AdminList,
)
```

---

### 🎨 Template Engine

**HTML Template Rendering**
- Layout system
- Partial templates
- Helper functions
- Asset integration
- Theme support

```go
// Render template
ctx.Render("pages/home", fiber.Map{
    "Title": "Welcome",
    "User":  currentUser,
})
```

**Template Example:**
```html
{{define "pages/home"}}
{{template "layouts/main" .}}
{{define "content"}}
    <h1>{{.Title}}</h1>
    <p>Welcome, {{.User.Name}}!</p>
{{end}}
{{end}}
```

---

### 📦 Asset Management

**Asset Pipeline**
- CSS/JS bundling
- Minification
- Cache busting
- CDN support
- Source maps

```go
// In templates
{{asset "css/app.css"}}        // → /assets/css/app.min.css?v=123
{{asset "js/app.js"}}           // → /assets/js/app.min.js?v=456
```

---

### 🎨 Theme System

**Multi-Theme Support**
- Switch themes dynamically
- Theme inheritance
- Custom layouts per theme
- Theme-specific assets

```go
// Set theme
themeService.SetTheme("dark")

// Theme-specific template
{{theme "components/header.html"}}
```

---

### 🔄 Middleware

**Extensive Middleware Support**
- CORS
- Rate limiting
- Compression
- Logging
- Recovery
- Request ID
- Custom middleware

```go
// Global middleware
app.Use(middleware.Logger())
app.Use(middleware.Recover())
app.Use(middleware.CORS())

// Route-specific middleware
api.Use(middleware.RateLimit(100)) // 100 req/min

// Custom middleware
app.Use(func(c *fiber.Ctx) error {
    // Custom logic
    return c.Next()
})
```

---

## Advanced Features

### 🌐 WebSocket Support

**Real-Time Communication**
- WebSocket connections
- Broadcasting
- Room management
- Authentication

```go
// WebSocket route
app.Get("/ws", websocket.New(func(c *websocket.Conn) {
    for {
        // Read message
        _, msg, _ := c.ReadMessage()
        
        // Echo back
        c.WriteMessage(websocket.TextMessage, msg)
    }
}))
```

---

### 📡 GraphQL

**GraphQL API**
- Schema-first approach
- Resolvers
- Subscriptions
- DataLoader pattern

```go
// GraphQL schema
type Query {
    user(id: ID!): User
    users: [User]
}

type Mutation {
    createUser(input: UserInput!): User
}
```

---

### 🔄 gRPC & Microservices

**Service Communication**
- gRPC support
- Protocol Buffers
- Service discovery
- Load balancing

```protobuf
// user.proto
service UserService {
    rpc GetUser (UserRequest) returns (UserResponse);
    rpc ListUsers (Empty) returns (UserList);
}
```

---

### 🗄️ Caching

**Multi-Level Caching**
- In-memory cache
- Redis integration
- Cache strategies (TTL, LRU)
- Cache tags

```go
// Cache data
cache.Set("user:123", user, 5*time.Minute)

// Get from cache
user, _ := cache.Get("user:123")

// Cache with tags
cache.Tags("users", "active").Set("user:123", user)

// Invalidate by tag
cache.Tags("users").Flush()
```

---

### 📬 Queue & Jobs

**Background Processing**
- Job queues
- Delayed jobs
- Job retry
- Priority queues

```go
// Dispatch job
queue.Dispatch(&SendEmailJob{
    To:      "user@example.com",
    Subject: "Welcome",
})

// Delayed job
queue.Dispatch(&SendReminderJob{...}).Delay(1 * time.Hour)

// Priority
queue.Dispatch(&UrgentJob{...}).Priority(HIGH)
```

---

### 📧 Email

**Email System**
- SMTP integration
- Email templates
- Attachments
- Queue integration

```go
// Send email
email.Send(&mail.Message{
    To:      []string{"user@example.com"},
    Subject: "Welcome",
    Body:    "Welcome to our platform!",
})

// With template
email.SendTemplate("welcome", user.Email, fiber.Map{
    "Name": user.Name,
})
```

---

### 📝 Logging

**Structured Logging**
- Multiple log levels
- Contextual logging
- Log rotation
- Multiple outputs

```go
// Structured logging
logger.Info("User registered",
    zap.String("email", user.Email),
    zap.Uint("user_id", user.ID),
)

// With context
logger.WithContext(ctx).Error("Database error", zap.Error(err))
```

---

### 📊 Metrics & Monitoring

**Observability**
- Prometheus metrics
- Health checks
- Request tracing
- Performance profiling

```go
// Health check endpoint
app.Get("/health", func(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{
        "status": "healthy",
        "uptime": time.Since(startTime),
    })
})
```

---

### 🔍 Validation

**Request Validation**
- Struct validation
- Custom validators
- Error messages
- Localization

```go
type RegisterRequest struct {
    Name     string `json:"name" validate:"required,min=3"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}

// Validate
if err := validator.Validate(req); err != nil {
    return c.Status(400).JSON(err)
}
```

---

### 📁 File Storage

**File Management**
- Local storage
- Cloud storage (S3, GCS)
- Image processing
- File validation

```go
// Upload file
file, _ := c.FormFile("file")
path, _ := storage.Put(file)

// Get URL
url := storage.URL(path)

// Delete
storage.Delete(path)
```

---

### 🔐 Security Features

**Built-in Security**
- CSRF protection
- XSS prevention
- SQL injection prevention
- Rate limiting
- Secure headers
- Input sanitization

```go
// CSRF protection
app.Use(middleware.CSRF())

// Secure headers
app.Use(middleware.SecureHeaders())

// Rate limiting
app.Use(middleware.RateLimit(100)) // 100 req/min
```

---

## Testing Features

### 🧪 Testing Utilities

**Comprehensive Testing**
- Unit testing helpers
- Integration testing
- HTTP testing
- Database testing
- Mocking support

```go
func TestUserService(t *testing.T) {
    // Setup
    service := NewUserService(mockRepo)
    
    // Test
    user, err := service.Create(ctx, &UserDTO{
        Name: "John",
    })
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "John", user.Name)
}
```

---

## CLI Features

### 🛠️ Command-Line Tools

**Code Generation**
```bash
# Generate module
neonex make:module blog

# Generate components
neonex make:model Post
neonex make:service PostService
neonex make:controller PostController
neonex make:repository PostRepository

# Generate migration
neonex make:migration create_posts_table
```

**Database Commands**
```bash
# Run migrations
neonex migrate:up

# Rollback
neonex migrate:down

# Seed database
neonex db:seed
```

---

## Deployment Features

### 🚢 Production Ready

**Deployment Options**
- Single binary
- Docker containers
- Kubernetes
- Cloud platforms

**Features:**
- Graceful shutdown
- Zero-downtime deploys
- Health checks
- Auto-scaling support

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main .

FROM alpine:latest
COPY --from=builder /app/main .
CMD ["./main"]
```

---

## Summary

NeonEx Framework provides:

- ✅ **20+ Core Features** - Everything you need
- ✅ **Production Ready** - Used by companies
- ✅ **Well Documented** - Comprehensive docs
- ✅ **Actively Maintained** - Regular updates
- ✅ **Community Support** - Active community
- ✅ **MIT Licensed** - Use freely

---

## Next Steps

- [Installation](../getting-started/installation.md)
- [Quick Start](../getting-started/quick-start.md)
- [Architecture](./architecture.md)

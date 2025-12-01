# Architecture Overview

## System Architecture

NeonEx Framework follows a layered, modular architecture designed for scalability, maintainability, and testability.

```
┌───────────────────────────────────────────────────────────┐
│                    Application Layer                       │
│            (Your Business Logic & Custom Modules)         │
└───────────────────────────────────────────────────────────┘
                            ↓
┌───────────────────────────────────────────────────────────┐
│                   Framework Layer                          │
│         (NeonEx Framework - Extended Features)            │
│  ┌─────────────┬──────────────┬─────────────────────┐   │
│  │  Frontend   │     Web      │   Theme System      │   │
│  │  Module     │   Utilities  │   Asset Pipeline    │   │
│  └─────────────┴──────────────┴─────────────────────┘   │
└───────────────────────────────────────────────────────────┘
                            ↓
┌───────────────────────────────────────────────────────────┐
│                     Core Layer                             │
│              (NeonEx Core - Foundation)                    │
│  ┌────────────┬────────────┬──────────┬─────────────┐   │
│  │  Module    │     DI     │   Auth   │  Database   │   │
│  │  System    │ Container  │  System  │    ORM      │   │
│  ├────────────┼────────────┼──────────┼─────────────┤   │
│  │  Routing   │ Middleware │  Config  │   Logging   │   │
│  └────────────┴────────────┴──────────┴─────────────┘   │
└───────────────────────────────────────────────────────────┘
                            ↓
┌───────────────────────────────────────────────────────────┐
│                 Infrastructure Layer                       │
│          (Third-Party Libraries & Drivers)                │
│  ┌────────┬────────┬────────┬────────┬────────────┐     │
│  │ Fiber  │  GORM  │  Zap   │ Redis  │  Postgres  │     │
│  └────────┴────────┴────────┴────────┴────────────┘     │
└───────────────────────────────────────────────────────────┘
```

---

## Architectural Patterns

### 1. Modular Architecture

Every feature is encapsulated in a self-contained module:

```go
type Module interface {
    Name() string
    RegisterServices(*Container) error
    RegisterRoutes(fiber.Router) error
    Boot() error
}
```

**Benefits:**
- **Separation of Concerns** - Each module handles one domain
- **Reusability** - Modules can be shared across projects
- **Testability** - Test modules in isolation
- **Maintainability** - Changes are localized

### 2. Dependency Injection

Type-safe dependency injection for loose coupling:

```go
// Register dependencies
container.Provide(NewDatabase)
container.Provide(NewUserRepository)
container.Provide(NewUserService)

// Auto-resolve with dependencies injected
service := container.Resolve[*UserService]()
```

**Benefits:**
- **Loose Coupling** - Components don't know about concrete implementations
- **Testability** - Easy to mock dependencies
- **Flexibility** - Swap implementations without changing code
- **Type Safety** - Compile-time checking

### 3. Repository Pattern

Abstraction layer between business logic and data access:

```go
type UserRepository interface {
    FindByID(ctx context.Context, id uint) (*User, error)
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id uint) error
}
```

**Benefits:**
- **Database Agnostic** - Switch databases without changing business logic
- **Testability** - Mock repositories for testing
- **Query Optimization** - Centralized query logic
- **Consistency** - Standardized data access patterns

### 4. Service Layer

Business logic separated from HTTP handlers:

```go
type UserService struct {
    repo UserRepository
}

func (s *UserService) Register(ctx context.Context, data *RegisterDTO) (*User, error) {
    // Business logic here
    // Validation, password hashing, etc.
}
```

**Benefits:**
- **Reusability** - Services used by multiple controllers
- **Testability** - Test business logic without HTTP
- **Separation** - HTTP concerns separated from business logic
- **Maintainability** - Logic changes don't affect controllers

### 5. MVC Pattern (Optional)

For web applications with views:

```
Model      → Database entities and business logic
View       → Templates and presentation
Controller → HTTP handlers and request/response
```

---

## Component Architecture

### Module System

```
Module
├── Models        (Entities & DTOs)
├── Repositories  (Data access)
├── Services      (Business logic)
├── Controllers   (HTTP handlers)
├── Routes        (Route definitions)
├── Middleware    (Request processing)
└── Config        (Module configuration)
```

**Example Module Structure:**

```go
// modules/blog/module.go
package blog

type BlogModule struct{}

func New() *BlogModule {
    return &BlogModule{}
}

func (m *BlogModule) Name() string {
    return "blog"
}

func (m *BlogModule) RegisterServices(c *core.Container) error {
    // Register in order of dependencies
    c.Provide(NewPostRepository)
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
    api.Put("/posts/:id", ctrl.Update)
    api.Delete("/posts/:id", ctrl.Delete)
    
    return nil
}

func (m *BlogModule) Boot() error {
    // Module initialization
    return nil
}
```

---

## Request Lifecycle

Understanding how a request flows through NeonEx:

```
1. HTTP Request
        ↓
2. Fiber Router
        ↓
3. Global Middleware
        ↓
4. Route Matching
        ↓
5. Route Middleware
        ↓
6. Controller Handler
        ↓
7. Service Layer
        ↓
8. Repository Layer
        ↓
9. Database Query
        ↓
10. Response Assembly
        ↓
11. Response Middleware
        ↓
12. HTTP Response
```

**Detailed Flow:**

```go
// 1. Request arrives
GET /api/v1/users/123

// 2. Fiber routes to handler
router.Get("/api/v1/users/:id", controller.GetUser)

// 3. Middleware executes
- Logger: Log request
- Auth: Verify JWT token
- CORS: Check origin
- RateLimit: Check limits

// 4. Controller receives request
func (c *UserController) GetUser(ctx *fiber.Ctx) error {
    id := ctx.Params("id")
    
    // 5. Call service
    user, err := c.service.GetByID(ctx.Context(), id)
    
    // 6. Return response
    return ctx.JSON(user)
}

// 7. Service handles business logic
func (s *UserService) GetByID(ctx context.Context, id string) (*User, error) {
    // Validate
    // Check permissions
    
    // 8. Query repository
    return s.repo.FindByID(ctx, id)
}

// 9. Repository queries database
func (r *UserRepository) FindByID(ctx context.Context, id string) (*User, error) {
    var user User
    err := r.db.Where("id = ?", id).First(&user).Error
    return &user, err
}
```

---

## Data Flow Architecture

### Write Operations

```
Controller → Validation → Service → Repository → Database
                ↓
           Event Dispatch
                ↓
        Cache Invalidation
```

### Read Operations

```
Controller → Service → Cache Check
                         ├── Hit: Return
                         └── Miss: Repository → Database
                                        ↓
                                   Cache Store
```

---

## Scalability Architecture

### Horizontal Scaling

```
                Load Balancer
                      ↓
        ┌─────────────┼─────────────┐
        ↓             ↓             ↓
    Instance 1    Instance 2    Instance 3
        ↓             ↓             ↓
        └─────────────┼─────────────┘
                      ↓
                  Database
                      ↓
                Redis Cache
```

**Key Points:**
- **Stateless** - No session data stored locally
- **Shared Cache** - Redis for distributed caching
- **Database Connection Pool** - Efficient resource usage
- **Load Balancer** - Distribute traffic evenly

### Vertical Scaling

```go
// Configuration for vertical scaling
config := &core.Config{
    MaxCPU:         runtime.NumCPU(), // Use all cores
    MaxConnections: 10000,            // Handle more connections
    WorkerPool:     1000,             // More workers
}
```

---

## Database Architecture

### Connection Management

```
Application
    ↓
Connection Pool (50 connections)
    ↓
┌───┴───┬───────┬───────┬───────┐
│ Conn1 │ Conn2 │ Conn3 │ ...   │
└───┬───┴───┬───┴───┬───┴───────┘
    └───────┼───────┘
            ↓
        Database
```

### Migration Strategy

```
Development → Staging → Production
     ↓            ↓          ↓
  Auto-Migrate  Manual   Manual
               Review    Review
```

---

## Security Architecture

### Multi-Layer Security

```
1. Transport Layer (TLS/HTTPS)
2. Application Layer (JWT, RBAC)
3. Business Layer (Validation, Authorization)
4. Data Layer (Encryption, SQL Injection Prevention)
```

### Authentication Flow

```
1. User Login
        ↓
2. Validate Credentials
        ↓
3. Generate JWT Token
        ↓
4. Return Token to Client
        ↓
5. Client Includes Token in Requests
        ↓
6. Server Validates Token
        ↓
7. Extract User Info
        ↓
8. Check Permissions (RBAC)
        ↓
9. Allow/Deny Access
```

---

## Caching Architecture

### Multi-Level Caching

```
Request
   ↓
Memory Cache (Local)
   ├── Hit → Return
   └── Miss
        ↓
   Redis Cache (Distributed)
        ├── Hit → Store in Memory → Return
        └── Miss
             ↓
        Database Query
             ↓
        Store in Redis + Memory → Return
```

---

## Microservices Architecture

NeonEx can be used in microservices:

```
                API Gateway
                     ↓
        ┌────────────┼────────────┐
        ↓            ↓            ↓
   User Service  Order Service  Payment Service
        │            │            │
        └────────────┼────────────┘
                     ↓
              Message Queue (RabbitMQ)
                     ↓
              Event Handlers
```

Each service is a NeonEx application with specific modules.

---

## Best Practices

### 1. Module Organization
- One module per domain/feature
- Keep modules small and focused
- Avoid circular dependencies

### 2. Dependency Management
- Register dependencies in order
- Use interfaces for flexibility
- Inject dependencies, don't instantiate

### 3. Error Handling
- Use custom error types
- Handle errors at appropriate layers
- Log errors with context

### 4. Testing
- Unit test services and repositories
- Integration test APIs
- Mock external dependencies

### 5. Performance
- Use connection pooling
- Implement caching strategically
- Profile and optimize hot paths

---

## Architecture Decisions

### Why Fiber?
- **Performance** - One of the fastest Go web frameworks
- **Express-like API** - Familiar to many developers
- **Rich Middleware** - Extensive ecosystem
- **Active Development** - Well maintained

### Why GORM?
- **Maturity** - Battle-tested ORM
- **Features** - Comprehensive feature set
- **Community** - Large user base
- **Documentation** - Well documented

### Why Dependency Injection?
- **Testability** - Easy to mock dependencies
- **Flexibility** - Swap implementations easily
- **Maintainability** - Loose coupling
- **Type Safety** - Compile-time checking

---

## Further Reading

- [Modules System](../core-concepts/modules.md)
- [Dependency Injection](../core-concepts/dependency-injection.md)
- [Repository Pattern](../database/repository-pattern.md)
- [Microservices](../advanced/microservices.md)

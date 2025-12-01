# Dependency Injection

## Overview

NeonEx Framework provides a powerful, type-safe Dependency Injection (DI) container that manages service lifecycles and automatically resolves dependencies. This promotes loose coupling, testability, and clean architecture.

## Why Dependency Injection?

### Without DI (❌ Tight Coupling)

```go
type UserService struct {
    repo *UserRepository
}

func NewUserService() *UserService {
    // Directly creates dependency - tight coupling
    db := connectDatabase()
    repo := NewUserRepository(db)
    return &UserService{repo: repo}
}
```

### With DI (✅ Loose Coupling)

```go
type UserService struct {
    repo *UserRepository
}

// Dependencies injected - loose coupling
func NewUserService(repo *UserRepository) *UserService {
    return &UserService{repo: repo}
}
```

## Container Basics

### Providing Services

Register services in your module's `RegisterServices()` method:

```go
func (m *BlogModule) RegisterServices(c *core.Container) error {
    // Register with Singleton lifecycle
    c.Provide(func() *BlogRepository {
        db := config.DB.GetDB()
        return NewBlogRepository(db)
    }, core.Singleton)
    
    // Register with dependencies
    c.Provide(func() *BlogService {
        repo := core.Resolve[*BlogRepository](c)
        cache := core.Resolve[*cache.CacheManager](c)
        return NewBlogService(repo, cache)
    }, core.Singleton)
    
    // Register controller (Transient - new instance per request)
    c.Provide(func() *BlogController {
        service := core.Resolve[*BlogService](c)
        return NewBlogController(service)
    }, core.Transient)
    
    return nil
}
```

### Resolving Services

```go
// In your code, resolve services from the container
service := core.Resolve[*BlogService](container)

// Or without container reference (uses global container)
service := core.Resolve[*BlogService]()
```

## Service Lifecycles

### Singleton

Created once and reused for the entire application lifetime:

```go
c.Provide(func() *DatabaseConnection {
    return NewDatabaseConnection()
}, core.Singleton)
```

**Use for:**
- Database connections
- Configuration
- Caches
- Shared resources

### Transient

New instance created every time it's resolved:

```go
c.Provide(func() *RequestHandler {
    return NewRequestHandler()
}, core.Transient)
```

**Use for:**
- Controllers (per-request handlers)
- Short-lived operations
- Stateless services

### Scoped (Advanced)

Lives for the duration of a request scope:

```go
c.Provide(func() *RequestContext {
    return NewRequestContext()
}, core.Scoped)
```

## Practical Examples

### Example 1: User Module Dependencies

```go
func (m *UserModule) RegisterServices(c *core.Container) error {
    // Infrastructure
    c.Provide(func() *database.TxManager {
        db := config.DB.GetDB()
        return database.NewTxManager(db)
    }, core.Singleton)
    
    // Security
    c.Provide(func() *auth.JWTManager {
        return auth.NewJWTManager(&auth.JWTConfig{
            SecretKey:     config.Get("JWT_SECRET"),
            AccessExpiry:  15 * time.Minute,
            RefreshExpiry: 7 * 24 * time.Hour,
        })
    }, core.Singleton)
    
    c.Provide(func() *auth.PasswordHasher {
        return auth.NewPasswordHasher(12)
    }, core.Singleton)
    
    // RBAC
    c.Provide(func() *rbac.Manager {
        db := config.DB.GetDB()
        return rbac.NewManager(db)
    }, core.Singleton)
    
    // Repository
    c.Provide(func() *UserRepository {
        db := config.DB.GetDB()
        return NewUserRepository(db)
    }, core.Singleton)
    
    // Service with multiple dependencies
    c.Provide(func() *UserService {
        repo := core.Resolve[*UserRepository](c)
        txManager := core.Resolve[*database.TxManager](c)
        return NewUserService(repo, txManager)
    }, core.Singleton)
    
    c.Provide(func() *AuthService {
        userRepo := core.Resolve[*UserRepository](c)
        jwtManager := core.Resolve[*auth.JWTManager](c)
        hasher := core.Resolve[*auth.PasswordHasher](c)
        rbacManager := core.Resolve[*rbac.Manager](c)
        return NewAuthService(userRepo, jwtManager, hasher, rbacManager)
    }, core.Singleton)
    
    // Controllers
    c.Provide(func() *AuthController {
        authService := core.Resolve[*AuthService](c)
        return NewAuthController(authService)
    }, core.Transient)
    
    c.Provide(func() *UserController {
        service := core.Resolve[*UserService](c)
        rbacManager := core.Resolve[*rbac.Manager](c)
        return NewUserController(service, rbacManager)
    }, core.Transient)
    
    return nil
}
```

### Example 2: E-commerce Module

```go
func (m *OrderModule) RegisterServices(c *core.Container) error {
    // Repositories
    c.Provide(func() *OrderRepository {
        db := config.DB.GetDB()
        return NewOrderRepository(db)
    }, core.Singleton)
    
    c.Provide(func() *PaymentRepository {
        db := config.DB.GetDB()
        return NewPaymentRepository(db)
    }, core.Singleton)
    
    // External services
    c.Provide(func() *PaymentGateway {
        apiKey := config.Get("STRIPE_API_KEY")
        return NewStripeGateway(apiKey)
    }, core.Singleton)
    
    c.Provide(func() *EmailService {
        smtp := core.Resolve[*notification.SMTPClient](c)
        return NewEmailService(smtp)
    }, core.Singleton)
    
    // Business logic
    c.Provide(func() *OrderService {
        orderRepo := core.Resolve[*OrderRepository](c)
        paymentRepo := core.Resolve[*PaymentRepository](c)
        paymentGateway := core.Resolve[*PaymentGateway](c)
        emailService := core.Resolve[*EmailService](c)
        eventBus := core.Resolve[*events.EventBus](c)
        
        return NewOrderService(
            orderRepo,
            paymentRepo,
            paymentGateway,
            emailService,
            eventBus,
        )
    }, core.Singleton)
    
    // Controller
    c.Provide(func() *OrderController {
        service := core.Resolve[*OrderService](c)
        return NewOrderController(service)
    }, core.Transient)
    
    return nil
}
```

## Constructor Injection Pattern

### Service with Dependencies

```go
// service.go
type BlogService struct {
    repo        *BlogRepository
    cache       *cache.CacheManager
    eventBus    *events.EventBus
    validator   *validation.Validator
}

func NewBlogService(
    repo *BlogRepository,
    cache *cache.CacheManager,
    eventBus *events.EventBus,
    validator *validation.Validator,
) *BlogService {
    return &BlogService{
        repo:      repo,
        cache:     cache,
        eventBus:  eventBus,
        validator: validator,
    }
}

func (s *BlogService) CreatePost(ctx context.Context, req *CreatePostRequest) (*Post, error) {
    // Validate using injected validator
    if err := s.validator.Validate(req); err != nil {
        return nil, err
    }
    
    // Create post
    post := &Post{
        Title:   req.Title,
        Content: req.Content,
    }
    
    if err := s.repo.Create(ctx, post); err != nil {
        return nil, err
    }
    
    // Publish event using injected event bus
    s.eventBus.Publish("post.created", post)
    
    // Invalidate cache using injected cache manager
    s.cache.Delete(ctx, "posts:list")
    
    return post, nil
}
```

### Registration

```go
c.Provide(func() *BlogService {
    repo := core.Resolve[*BlogRepository](c)
    cache := core.Resolve[*cache.CacheManager](c)
    eventBus := core.Resolve[*events.EventBus](c)
    validator := core.Resolve[*validation.Validator](c)
    
    return NewBlogService(repo, cache, eventBus, validator)
}, core.Singleton)
```

## Type Safety

The container is fully type-safe using Go generics:

```go
// ✅ Type-safe resolution
service := core.Resolve[*BlogService](c)
service.CreatePost(...)  // Compile-time type checking

// ❌ Compile error if type doesn't exist
wrong := core.Resolve[*NonExistentService](c)  // Compile error
```

## Common Patterns

### 1. Repository Pattern with DI

```go
// Define interface
type IUserRepository interface {
    FindByID(ctx context.Context, id uint) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
    Create(ctx context.Context, user *User) error
}

// Implement interface
type UserRepository struct {
    *database.BaseRepository[User]
}

func NewUserRepository(db *gorm.DB) IUserRepository {
    return &UserRepository{
        BaseRepository: database.NewBaseRepository[User](db),
    }
}

// Register interface, not concrete type
c.Provide(func() IUserRepository {
    db := config.DB.GetDB()
    return NewUserRepository(db)
}, core.Singleton)

// Depend on interface
type UserService struct {
    repo IUserRepository  // Interface, not concrete type
}

func NewUserService(repo IUserRepository) *UserService {
    return &UserService{repo: repo}
}
```

### 2. Factory Pattern

```go
// Factory function
type ServiceFactory func() *SomeService

// Register factory
c.Provide(func() ServiceFactory {
    return func() *SomeService {
        return NewSomeService()
    }
}, core.Singleton)

// Use factory
factory := core.Resolve[ServiceFactory](c)
service1 := factory()
service2 := factory()  // Create multiple instances
```

### 3. Conditional Registration

```go
func (m *MyModule) RegisterServices(c *core.Container) error {
    // Register different implementations based on config
    if config.Get("CACHE_DRIVER") == "redis" {
        c.Provide(func() cache.CacheManager {
            return cache.NewRedisCache()
        }, core.Singleton)
    } else {
        c.Provide(func() cache.CacheManager {
            return cache.NewMemoryCache()
        }, core.Singleton)
    }
    
    return nil
}
```

## Testing with DI

### Mocking Dependencies

```go
// Mock repository
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uint) (*User, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*User), args.Error(1)
}

// Test with mock
func TestUserService_GetUser(t *testing.T) {
    // Create mock
    mockRepo := new(MockUserRepository)
    mockRepo.On("FindByID", mock.Anything, uint(1)).
        Return(&User{ID: 1, Name: "John"}, nil)
    
    // Create service with mock
    service := NewUserService(mockRepo)
    
    // Test
    user, err := service.GetUser(context.Background(), 1)
    
    assert.NoError(t, err)
    assert.Equal(t, "John", user.Name)
    mockRepo.AssertExpectations(t)
}
```

### Test Container

```go
func setupTestContainer(t *testing.T) *core.Container {
    c := core.NewContainer()
    
    // Register test database
    c.Provide(func() *gorm.DB {
        return setupTestDB(t)
    }, core.Singleton)
    
    // Register real services
    c.Provide(func() *UserRepository {
        db := core.Resolve[*gorm.DB](c)
        return NewUserRepository(db)
    }, core.Singleton)
    
    c.Provide(func() *UserService {
        repo := core.Resolve[*UserRepository](c)
        return NewUserService(repo)
    }, core.Singleton)
    
    return c
}

func TestIntegration(t *testing.T) {
    container := setupTestContainer(t)
    service := core.Resolve[*UserService](container)
    
    // Test with real database
    user, err := service.CreateUser(context.Background(), &CreateUserRequest{
        Name:  "Test User",
        Email: "test@example.com",
    })
    
    assert.NoError(t, err)
    assert.NotNil(t, user)
}
```

## Best Practices

### 1. Register in Correct Order

Register dependencies before dependents:

```go
// ✅ Good - Repository before Service
c.Provide(NewUserRepository, core.Singleton)
c.Provide(NewUserService, core.Singleton)  // Depends on Repository

// ❌ Bad - Service before Repository
c.Provide(NewUserService, core.Singleton)   // Will fail - Repository not registered
c.Provide(NewUserRepository, core.Singleton)
```

### 2. Use Interfaces for Flexibility

```go
// ✅ Good - Depend on interface
type UserService struct {
    repo IUserRepository
    cache ICacheManager
}

// ❌ Bad - Depend on concrete types
type UserService struct {
    repo *UserRepository
    cache *RedisCache
}
```

### 3. Avoid Circular Dependencies

```go
// ❌ Bad - Circular dependency
type ServiceA struct {
    serviceB *ServiceB
}

type ServiceB struct {
    serviceA *ServiceA  // Circular!
}

// ✅ Good - Use event bus or mediator
type ServiceA struct {
    eventBus *EventBus
}

type ServiceB struct {
    eventBus *EventBus
}
```

### 4. Keep Constructors Simple

```go
// ✅ Good - Simple constructor
func NewUserService(repo *UserRepository) *UserService {
    return &UserService{repo: repo}
}

// ❌ Bad - Complex logic in constructor
func NewUserService(repo *UserRepository) *UserService {
    service := &UserService{repo: repo}
    service.loadCache()        // Side effect!
    service.startWorker()      // Background process!
    return service
}
```

### 5. Use Lifecycle Methods

```go
// Put initialization logic in Boot()
func (m *MyModule) Boot() error {
    service := core.Resolve[*MyService]()
    if err := service.Initialize(); err != nil {
        return err
    }
    return nil
}
```

## Advanced Features

### Service Decorators

```go
// Base service
type LoggingUserService struct {
    inner  *UserService
    logger *logger.Logger
}

func NewLoggingUserService(inner *UserService, logger *logger.Logger) *LoggingUserService {
    return &LoggingUserService{inner: inner, logger: logger}
}

func (s *LoggingUserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    s.logger.Info("Creating user", "email", req.Email)
    
    user, err := s.inner.CreateUser(ctx, req)
    
    if err != nil {
        s.logger.Error("Failed to create user", "error", err)
    } else {
        s.logger.Info("User created successfully", "id", user.ID)
    }
    
    return user, err
}

// Register decorator
c.Provide(func() *UserService {
    repo := core.Resolve[*UserRepository](c)
    baseService := NewUserService(repo)
    
    logger := core.Resolve[*logger.Logger](c)
    return NewLoggingUserService(baseService, logger)
}, core.Singleton)
```

## Troubleshooting

### Service Not Found

```go
// Error: service not registered
service := core.Resolve[*MyService](c)  // Panic: service not found

// Solution: Register the service first
c.Provide(func() *MyService {
    return NewMyService()
}, core.Singleton)
```

### Dependency Resolution Failed

```go
// Error: dependency not found
c.Provide(func() *ServiceA {
    serviceB := core.Resolve[*ServiceB](c)  // Panic: ServiceB not registered
    return NewServiceA(serviceB)
}, core.Singleton)

// Solution: Register ServiceB first
c.Provide(NewServiceB, core.Singleton)
c.Provide(NewServiceA, core.Singleton)
```

## Next Steps

- Learn about [Routing](routing.md)
- Explore [Middleware](middleware.md)
- Understand [Module System](modules.md)

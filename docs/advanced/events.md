# Event System

Build event-driven applications with NeonEx Framework's powerful publish-subscribe system. Learn to decouple components, implement async workflows, and create reactive applications.

## Table of Contents

- [Introduction](#introduction)
- [Quick Start](#quick-start)
- [Event Dispatcher](#event-dispatcher)
- [Event Handlers](#event-handlers)
- [Async Events](#async-events)
- [Built-in Events](#built-in-events)
- [Custom Events](#custom-events)
- [Integration Patterns](#integration-patterns)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Introduction

NeonEx provides a flexible event system based on the publish-subscribe pattern. The event system enables:

- **Loose Coupling**: Decouple components through events
- **Async Processing**: Handle events synchronously or asynchronously
- **Event Listeners**: Multiple handlers per event
- **Type Safety**: Strongly-typed event data
- **Error Handling**: Comprehensive error propagation
- **Performance**: Efficient concurrent event processing

## Quick Start

### Basic Event Usage

```go
package main

import (
    "context"
    "fmt"
    
    "neonex/core/pkg/events"
)

func main() {
    // Create event dispatcher
    dispatcher := events.NewEventDispatcher()
    
    // Register event handler
    dispatcher.Register("user.created", func(ctx context.Context, event events.Event) error {
        userData := event.Data.(map[string]interface{})
        fmt.Printf("User created: %v\n", userData["name"])
        return nil
    })
    
    // Dispatch event
    ctx := context.Background()
    dispatcher.Dispatch(ctx, events.Event{
        Name: "user.created",
        Data: map[string]interface{}{
            "id":    1,
            "name":  "John Doe",
            "email": "john@example.com",
        },
    })
}
```

### Global Event Dispatcher

```go
import "neonex/core/pkg/events"

func init() {
    // Register global event handlers
    events.Register("user.created", SendWelcomeEmail)
    events.Register("user.created", CreateUserProfile)
    events.Register("user.created", NotifyAdmins)
}

func SendWelcomeEmail(ctx context.Context, event events.Event) error {
    userData := event.Data.(map[string]interface{})
    email := userData["email"].(string)
    
    // Send welcome email
    return emailService.Send(email, "Welcome!", welcomeTemplate)
}
```

## Event Dispatcher

### Creating Dispatcher

```go
// Create new dispatcher
dispatcher := events.NewEventDispatcher()

// Use in module
type UserModule struct {
    dispatcher *events.EventDispatcher
}

func NewUserModule(dispatcher *events.EventDispatcher) *UserModule {
    return &UserModule{
        dispatcher: dispatcher,
    }
}
```

### Registering Handlers

```go
// Single handler
dispatcher.Register("user.created", func(ctx context.Context, event events.Event) error {
    // Handle event
    return nil
})

// Multiple handlers for same event
dispatcher.Register("user.created", SendWelcomeEmail)
dispatcher.Register("user.created", CreateUserProfile)
dispatcher.Register("user.created", LogUserCreation)
dispatcher.Register("user.created", UpdateStatistics)
```

### Checking Handler Existence

```go
if dispatcher.HasHandlers("user.created") {
    fmt.Println("User creation handlers registered")
}
```

### Dispatching Events

```go
ctx := context.Background()

// Synchronous dispatch
err := dispatcher.Dispatch(ctx, events.Event{
    Name: "user.created",
    Data: userData,
})

if err != nil {
    log.Error("Event dispatch failed", logger.Fields{"error": err})
}

// Asynchronous dispatch (fire and forget)
dispatcher.DispatchAsync(ctx, events.Event{
    Name: "user.created",
    Data: userData,
})
```

## Event Handlers

### Simple Handler

```go
func UserCreatedHandler(ctx context.Context, event events.Event) error {
    userData := event.Data.(map[string]interface{})
    
    log.Info("User created", logger.Fields{
        "user_id": userData["id"],
        "email":   userData["email"],
    })
    
    return nil
}

// Register
dispatcher.Register(events.EventUserCreated, UserCreatedHandler)
```

### Typed Handler

```go
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func TypedUserCreatedHandler(ctx context.Context, event events.Event) error {
    // Type assertion
    user, ok := event.Data.(*User)
    if !ok {
        return fmt.Errorf("invalid event data type")
    }
    
    fmt.Printf("User %s created with email %s\n", user.Name, user.Email)
    return nil
}
```

### Handler with Dependencies

```go
type UserEventHandlers struct {
    db          *gorm.DB
    cache       cache.Cache
    emailSender *notification.Manager
}

func NewUserEventHandlers(db *gorm.DB, cache cache.Cache, email *notification.Manager) *UserEventHandlers {
    return &UserEventHandlers{
        db:          db,
        cache:       cache,
        emailSender: email,
    }
}

func (h *UserEventHandlers) OnUserCreated(ctx context.Context, event events.Event) error {
    userData := event.Data.(map[string]interface{})
    
    // Create user profile
    profile := &UserProfile{
        UserID:    int(userData["id"].(float64)),
        CreatedAt: time.Now(),
    }
    
    if err := h.db.Create(profile).Error; err != nil {
        return err
    }
    
    // Send welcome email
    email := userData["email"].(string)
    return h.emailSender.SendEmail(ctx, email, "Welcome!", "Welcome to our platform!")
}

// Register
func (h *UserEventHandlers) Register(dispatcher *events.EventDispatcher) {
    dispatcher.Register(events.EventUserCreated, h.OnUserCreated)
    dispatcher.Register(events.EventUserUpdated, h.OnUserUpdated)
    dispatcher.Register(events.EventUserDeleted, h.OnUserDeleted)
}
```

### Error Handling in Handlers

```go
func ResilientHandler(ctx context.Context, event events.Event) error {
    // Implement retry logic
    maxRetries := 3
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        err := processEvent(ctx, event)
        if err == nil {
            return nil
        }
        
        lastErr = err
        log.Warn("Handler retry", logger.Fields{
            "attempt": i + 1,
            "error":   err,
        })
        
        // Exponential backoff
        time.Sleep(time.Duration(i+1) * time.Second)
    }
    
    return fmt.Errorf("handler failed after %d retries: %w", maxRetries, lastErr)
}
```

## Async Events

### Fire and Forget

```go
// Dispatch asynchronously - doesn't wait for completion
dispatcher.DispatchAsync(ctx, events.Event{
    Name: "user.logged_in",
    Data: map[string]interface{}{
        "user_id": userID,
        "ip":      ipAddress,
        "time":    time.Now(),
    },
})

// Continue without waiting
return c.JSON(http.StatusOK, response)
```

### Background Processing

```go
type AsyncEventProcessor struct {
    dispatcher *events.EventDispatcher
    queue      chan events.Event
    workers    int
}

func NewAsyncEventProcessor(dispatcher *events.EventDispatcher, workers int) *AsyncEventProcessor {
    return &AsyncEventProcessor{
        dispatcher: dispatcher,
        queue:      make(chan events.Event, 1000),
        workers:    workers,
    }
}

func (p *AsyncEventProcessor) Start(ctx context.Context) {
    for i := 0; i < p.workers; i++ {
        go p.worker(ctx)
    }
}

func (p *AsyncEventProcessor) worker(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case event := <-p.queue:
            if err := p.dispatcher.Dispatch(ctx, event); err != nil {
                log.Error("Async event processing failed", logger.Fields{
                    "event": event.Name,
                    "error": err,
                })
            }
        }
    }
}

func (p *AsyncEventProcessor) Enqueue(event events.Event) error {
    select {
    case p.queue <- event:
        return nil
    default:
        return fmt.Errorf("event queue full")
    }
}
```

### Event Batching

```go
type EventBatcher struct {
    dispatcher  *events.EventDispatcher
    events      []events.Event
    batchSize   int
    flushTime   time.Duration
    mu          sync.Mutex
}

func NewEventBatcher(dispatcher *events.EventDispatcher, batchSize int, flushTime time.Duration) *EventBatcher {
    b := &EventBatcher{
        dispatcher: dispatcher,
        events:     make([]events.Event, 0, batchSize),
        batchSize:  batchSize,
        flushTime:  flushTime,
    }
    
    go b.autoFlush()
    return b
}

func (b *EventBatcher) Add(event events.Event) {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    b.events = append(b.events, event)
    
    if len(b.events) >= b.batchSize {
        b.flush()
    }
}

func (b *EventBatcher) flush() {
    if len(b.events) == 0 {
        return
    }
    
    ctx := context.Background()
    for _, event := range b.events {
        b.dispatcher.DispatchAsync(ctx, event)
    }
    
    b.events = b.events[:0]
}

func (b *EventBatcher) autoFlush() {
    ticker := time.NewTicker(b.flushTime)
    defer ticker.Stop()
    
    for range ticker.C {
        b.mu.Lock()
        b.flush()
        b.mu.Unlock()
    }
}
```

## Built-in Events

### User Events

```go
const (
    EventUserCreated       = "user.created"
    EventUserUpdated       = "user.updated"
    EventUserDeleted       = "user.deleted"
    EventUserLoggedIn      = "user.logged_in"
    EventUserLoggedOut     = "user.logged_out"
    EventUserPasswordReset = "user.password_reset"
)

// Usage
events.Dispatch(ctx, events.Event{
    Name: events.EventUserCreated,
    Data: map[string]interface{}{
        "user_id": user.ID,
        "email":   user.Email,
        "name":    user.Name,
    },
})
```

### Module Events

```go
const (
    EventModuleInstalled   = "module.installed"
    EventModuleUninstalled = "module.uninstalled"
    EventModuleActivated   = "module.activated"
    EventModuleDeactivated = "module.deactivated"
    EventModuleUpdated     = "module.updated"
)
```

### System Events

```go
const (
    EventSystemStarted  = "system.started"
    EventSystemShutdown = "system.shutdown"
)

// Dispatch on application startup
func (app *App) Start() {
    // ... startup logic ...
    
    events.Dispatch(context.Background(), events.Event{
        Name: events.EventSystemStarted,
        Data: map[string]interface{}{
            "version":    app.Version,
            "start_time": time.Now(),
        },
    })
}
```

## Custom Events

### Defining Custom Events

```go
package product

const (
    EventProductCreated     = "product.created"
    EventProductUpdated     = "product.updated"
    EventProductDeleted     = "product.deleted"
    EventProductOutOfStock  = "product.out_of_stock"
    EventProductPriceChanged = "product.price_changed"
)

// Typed event data
type ProductCreatedEvent struct {
    ProductID   int
    Name        string
    Price       float64
    CategoryID  int
    CreatedBy   int
    CreatedAt   time.Time
}

type ProductPriceChangedEvent struct {
    ProductID int
    OldPrice  float64
    NewPrice  float64
    ChangedBy int
    ChangedAt time.Time
}
```

### Dispatching Custom Events

```go
func (s *ProductService) CreateProduct(ctx context.Context, req *CreateProductRequest) (*Product, error) {
    product := &Product{
        Name:       req.Name,
        Price:      req.Price,
        CategoryID: req.CategoryID,
        CreatedBy:  getUserID(ctx),
    }
    
    if err := s.db.Create(product).Error; err != nil {
        return nil, err
    }
    
    // Dispatch event
    s.dispatcher.DispatchAsync(ctx, events.Event{
        Name: EventProductCreated,
        Data: &ProductCreatedEvent{
            ProductID:  product.ID,
            Name:       product.Name,
            Price:      product.Price,
            CategoryID: product.CategoryID,
            CreatedBy:  product.CreatedBy,
            CreatedAt:  time.Now(),
        },
    })
    
    return product, nil
}
```

### Handling Custom Events

```go
type ProductEventHandlers struct {
    notifier   *notification.Manager
    analytics  *AnalyticsService
    cache      cache.Cache
}

func (h *ProductEventHandlers) OnProductCreated(ctx context.Context, event events.Event) error {
    data := event.Data.(*ProductCreatedEvent)
    
    // Send notification to admins
    h.notifier.SendEmail(ctx, "admin@example.com", "New Product", 
        fmt.Sprintf("Product %s created", data.Name))
    
    // Track in analytics
    h.analytics.Track("product_created", map[string]interface{}{
        "product_id": data.ProductID,
        "category":   data.CategoryID,
    })
    
    // Invalidate cache
    h.cache.Delete(ctx, "products:list")
    
    return nil
}

func (h *ProductEventHandlers) OnPriceChanged(ctx context.Context, event events.Event) error {
    data := event.Data.(*ProductPriceChangedEvent)
    
    // Alert if significant price change
    changePercent := (data.NewPrice - data.OldPrice) / data.OldPrice * 100
    if math.Abs(changePercent) > 20 {
        h.notifier.SendEmail(ctx, "pricing@example.com", 
            "Significant Price Change",
            fmt.Sprintf("Product %d price changed by %.2f%%", data.ProductID, changePercent))
    }
    
    return nil
}
```

## Integration Patterns

### Event-Driven Cache Invalidation

```go
func SetupCacheInvalidation(dispatcher *events.EventDispatcher, cache cache.Cache) {
    // User cache invalidation
    dispatcher.Register(events.EventUserUpdated, func(ctx context.Context, event events.Event) error {
        data := event.Data.(map[string]interface{})
        userID := int(data["user_id"].(float64))
        
        keys := []string{
            fmt.Sprintf("user:%d", userID),
            fmt.Sprintf("user:%d:profile", userID),
            fmt.Sprintf("user:%d:settings", userID),
        }
        
        return cache.DeleteMulti(ctx, keys)
    })
    
    // Product cache invalidation
    dispatcher.Register("product.updated", func(ctx context.Context, event events.Event) error {
        data := event.Data.(map[string]interface{})
        productID := int(data["product_id"].(float64))
        
        // Invalidate product cache
        cache.Delete(ctx, fmt.Sprintf("product:%d", productID))
        
        // Invalidate list caches
        pattern := "products:list:*"
        keys, _ := cache.Keys(ctx, pattern)
        cache.DeleteMulti(ctx, keys)
        
        return nil
    })
}
```

### Event-Driven Notifications

```go
func SetupNotificationHandlers(dispatcher *events.EventDispatcher, notifier *notification.Manager) {
    // Welcome email on user creation
    dispatcher.Register(events.EventUserCreated, func(ctx context.Context, event events.Event) error {
        userData := event.Data.(map[string]interface{})
        email := userData["email"].(string)
        name := userData["name"].(string)
        
        return notifier.SendEmail(ctx, email, "Welcome!",
            fmt.Sprintf("Hello %s, welcome to our platform!", name))
    })
    
    // Password reset notification
    dispatcher.Register(events.EventUserPasswordReset, func(ctx context.Context, event events.Event) error {
        userData := event.Data.(map[string]interface{})
        email := userData["email"].(string)
        token := userData["reset_token"].(string)
        
        resetURL := fmt.Sprintf("https://example.com/reset-password?token=%s", token)
        return notifier.SendEmail(ctx, email, "Password Reset",
            fmt.Sprintf("Click here to reset your password: %s", resetURL))
    })
    
    // Order confirmation
    dispatcher.Register("order.created", func(ctx context.Context, event events.Event) error {
        orderData := event.Data.(map[string]interface{})
        email := orderData["customer_email"].(string)
        orderID := orderData["order_id"]
        
        return notifier.SendEmail(ctx, email, "Order Confirmation",
            fmt.Sprintf("Your order #%v has been confirmed!", orderID))
    })
}
```

### Event-Driven Analytics

```go
type AnalyticsTracker struct {
    analytics *AnalyticsService
}

func (at *AnalyticsTracker) RegisterHandlers(dispatcher *events.EventDispatcher) {
    // Track user events
    dispatcher.Register(events.EventUserCreated, at.trackUserCreated)
    dispatcher.Register(events.EventUserLoggedIn, at.trackUserLogin)
    
    // Track product events
    dispatcher.Register("product.viewed", at.trackProductView)
    dispatcher.Register("product.purchased", at.trackProductPurchase)
}

func (at *AnalyticsTracker) trackUserCreated(ctx context.Context, event events.Event) error {
    userData := event.Data.(map[string]interface{})
    
    return at.analytics.Track("user_signup", map[string]interface{}{
        "user_id":    userData["user_id"],
        "source":     userData["signup_source"],
        "timestamp":  time.Now(),
    })
}

func (at *AnalyticsTracker) trackProductView(ctx context.Context, event events.Event) error {
    data := event.Data.(map[string]interface{})
    
    return at.analytics.Track("product_view", map[string]interface{}{
        "product_id": data["product_id"],
        "user_id":    data["user_id"],
        "session_id": data["session_id"],
    })
}
```

### Event-Driven Workflow

```go
type OrderWorkflow struct {
    dispatcher  *events.EventDispatcher
    inventory   *InventoryService
    payment     *PaymentService
    shipping    *ShippingService
}

func (w *OrderWorkflow) RegisterHandlers() {
    // Step 1: Validate order
    w.dispatcher.Register("order.created", w.validateOrder)
    
    // Step 2: Reserve inventory
    w.dispatcher.Register("order.validated", w.reserveInventory)
    
    // Step 3: Process payment
    w.dispatcher.Register("inventory.reserved", w.processPayment)
    
    // Step 4: Arrange shipping
    w.dispatcher.Register("payment.completed", w.arrangeShipping)
    
    // Step 5: Confirm order
    w.dispatcher.Register("shipping.arranged", w.confirmOrder)
    
    // Error handling
    w.dispatcher.Register("order.failed", w.handleOrderFailure)
}

func (w *OrderWorkflow) validateOrder(ctx context.Context, event events.Event) error {
    orderData := event.Data.(map[string]interface{})
    orderID := int(orderData["order_id"].(float64))
    
    // Validate order
    if err := w.validateOrderData(orderID); err != nil {
        w.dispatcher.DispatchAsync(ctx, events.Event{
            Name: "order.failed",
            Data: map[string]interface{}{
                "order_id": orderID,
                "reason":   err.Error(),
            },
        })
        return err
    }
    
    // Dispatch next step
    w.dispatcher.DispatchAsync(ctx, events.Event{
        Name: "order.validated",
        Data: orderData,
    })
    
    return nil
}
```

### Event Logging

```go
type EventLogger struct {
    logger logger.Logger
}

func (el *EventLogger) RegisterHandlers(dispatcher *events.EventDispatcher) {
    // Log all user events
    dispatcher.Register(events.EventUserCreated, el.logEvent)
    dispatcher.Register(events.EventUserUpdated, el.logEvent)
    dispatcher.Register(events.EventUserDeleted, el.logEvent)
    dispatcher.Register(events.EventUserLoggedIn, el.logEvent)
    
    // Log critical system events
    dispatcher.Register(events.EventSystemStarted, el.logSystemEvent)
    dispatcher.Register(events.EventSystemShutdown, el.logSystemEvent)
}

func (el *EventLogger) logEvent(ctx context.Context, event events.Event) error {
    el.logger.Info("Event dispatched", logger.Fields{
        "event": event.Name,
        "data":  event.Data,
        "time":  time.Now(),
    })
    return nil
}

func (el *EventLogger) logSystemEvent(ctx context.Context, event events.Event) error {
    el.logger.Info("System event", logger.Fields{
        "event": event.Name,
        "data":  event.Data,
    })
    return nil
}
```

## Best Practices

### 1. Event Naming Convention

```go
// Use hierarchical naming: domain.action
const (
    // Good
    EventUserCreated      = "user.created"
    EventUserUpdated      = "user.updated"
    EventOrderPlaced      = "order.placed"
    EventPaymentCompleted = "payment.completed"
    
    // Avoid
    // "newUser" (unclear, no namespace)
    // "user_created" (use dots, not underscores)
)
```

### 2. Event Data Structure

```go
// Define typed event data
type UserCreatedEvent struct {
    UserID    int       `json:"user_id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
    IPAddress string    `json:"ip_address"`
}

// Use struct instead of map[string]interface{} when possible
dispatcher.Dispatch(ctx, events.Event{
    Name: EventUserCreated,
    Data: &UserCreatedEvent{
        UserID:    user.ID,
        Email:     user.Email,
        Name:      user.Name,
        CreatedAt: time.Now(),
        IPAddress: getIP(ctx),
    },
})
```

### 3. Idempotent Handlers

```go
// Make handlers idempotent (safe to run multiple times)
func IdempotentHandler(ctx context.Context, event events.Event) error {
    data := event.Data.(map[string]interface{})
    userID := int(data["user_id"].(float64))
    
    // Check if already processed
    processed, _ := cache.Get(ctx, fmt.Sprintf("event:processed:%s:%d", event.Name, userID))
    if processed != nil {
        return nil // Already processed
    }
    
    // Process event
    if err := processEvent(data); err != nil {
        return err
    }
    
    // Mark as processed
    cache.Set(ctx, fmt.Sprintf("event:processed:%s:%d", event.Name, userID), true, 1*time.Hour)
    
    return nil
}
```

### 4. Error Handling

```go
// Don't let one handler failure stop others
func SafeDispatch(dispatcher *events.EventDispatcher, ctx context.Context, event events.Event) {
    if err := dispatcher.Dispatch(ctx, event); err != nil {
        log.Error("Event dispatch failed", logger.Fields{
            "event": event.Name,
            "error": err,
        })
        
        // Optionally retry or queue for later
    }
}
```

### 5. Context Propagation

```go
// Pass context with timeout
func DispatchWithTimeout(dispatcher *events.EventDispatcher, event events.Event, timeout time.Duration) error {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    return dispatcher.Dispatch(ctx, event)
}

// Pass context with values
func DispatchWithUser(dispatcher *events.EventDispatcher, event events.Event, userID int) error {
    ctx := context.WithValue(context.Background(), "user_id", userID)
    return dispatcher.Dispatch(ctx, event)
}
```

## Troubleshooting

### Events Not Firing

```go
// Check if handlers are registered
if !dispatcher.HasHandlers("user.created") {
    log.Warn("No handlers registered for user.created")
}

// Debug event dispatching
func DebugDispatch(dispatcher *events.EventDispatcher, ctx context.Context, event events.Event) error {
    log.Debug("Dispatching event", logger.Fields{
        "event": event.Name,
        "data":  event.Data,
    })
    
    err := dispatcher.Dispatch(ctx, event)
    
    if err != nil {
        log.Error("Event dispatch error", logger.Fields{
            "event": event.Name,
            "error": err,
        })
    } else {
        log.Debug("Event dispatched successfully")
    }
    
    return err
}
```

### Handler Failures

```go
// Wrap handlers with error recovery
func RecoverableHandler(handler events.Handler) events.Handler {
    return func(ctx context.Context, event events.Event) (err error) {
        defer func() {
            if r := recover(); r != nil {
                err = fmt.Errorf("handler panic: %v", r)
                log.Error("Handler panic", logger.Fields{
                    "event": event.Name,
                    "panic": r,
                })
            }
        }()
        
        return handler(ctx, event)
    }
}

// Register with recovery
dispatcher.Register("user.created", RecoverableHandler(MyHandler))
```

### Performance Issues

```go
// Use async dispatch for non-critical events
dispatcher.DispatchAsync(ctx, event)

// Batch events
type EventQueue struct {
    events []events.Event
    mu     sync.Mutex
}

func (eq *EventQueue) Enqueue(event events.Event) {
    eq.mu.Lock()
    defer eq.mu.Unlock()
    eq.events = append(eq.events, event)
}

func (eq *EventQueue) Flush(dispatcher *events.EventDispatcher, ctx context.Context) {
    eq.mu.Lock()
    events := eq.events
    eq.events = nil
    eq.mu.Unlock()
    
    for _, event := range events {
        dispatcher.DispatchAsync(ctx, event)
    }
}
```

---

**Next Steps:**
- Learn about [Queue System](queue.md) for reliable event processing
- Explore [Notifications](email.md) for event-driven alerts
- See [Cache Invalidation](cache.md#cache-invalidation) with events

**Related Topics:**
- [Middleware](../core-concepts/middleware.md)
- [Background Jobs](queue.md)
- [Webhooks](../api-reference/webhooks.md)

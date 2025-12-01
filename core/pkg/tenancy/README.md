# Multi-tenancy Package

Complete multi-tenancy solution with database isolation, tenant management, and SaaS features for NeonexCore.

## Features

- ✅ **Multiple Isolation Strategies** - Shared DB, Separate DB, Shared Schema
- ✅ **Tenant Management** - CRUD operations with caching
- ✅ **Automatic Resolution** - Subdomain, domain, header, query parameter
- ✅ **Database Scoping** - Automatic tenant_id filtering
- ✅ **Middleware** - Fiber middleware for tenant detection
- ✅ **Quota Management** - Users, storage, rate limiting per tenant
- ✅ **Status Management** - Active, suspended, expired, deleted
- ✅ **Context Integration** - Tenant information in context
- ✅ **GORM Integration** - Automatic tenant_id injection

## Architecture

```
pkg/tenancy/
├── tenant.go      - Tenant model and manager
├── resolver.go    - Database resolution strategies
├── middleware.go  - Fiber middleware
├── store.go       - GORM persistence layer
└── README.md      - Documentation
```

## Quick Start

### 1. Initialize Tenancy

```go
import (
    "neonexcore/pkg/tenancy"
    "gorm.io/gorm"
)

// Create store
store := tenancy.NewGormTenantStore(db)
store.Migrate()

// Create manager
manager := tenancy.NewTenantManager(store)

// Create resolver with strategy
resolverConfig := tenancy.ResolverConfig{
    Strategy: tenancy.StrategySharedDatabase,
    SharedDB: db,
}
resolver := tenancy.NewResolver(resolverConfig, manager)
```

### 2. Add Middleware

```go
app := fiber.New()

// Tenant middleware
app.Use(tenancy.Middleware(manager, resolver))

// Your routes
app.Get("/api/users", handleGetUsers)
```

### 3. Use Tenant in Handlers

```go
func handleGetUsers(c *fiber.Ctx) error {
    // Get tenant from context
    tenant, _ := tenancy.GetTenantFromLocals(c)
    
    // Get tenant-scoped database
    db, _ := tenancy.GetTenantDBFromLocals(c)
    
    // Your logic
    return c.JSON(fiber.Map{"tenant": tenant.Name})
}
```

## Isolation Strategies

### Shared Database (Recommended)

All tenants share one database with `tenant_id` column:

```go
resolverConfig := tenancy.ResolverConfig{
    Strategy: tenancy.StrategySharedDatabase,
    SharedDB: db,
}
```

**Pros:**
- Simple setup
- Easy maintenance
- Cost-effective
- Good for most use cases

**Cons:**
- Data must be carefully isolated
- Limited customization per tenant

### Separate Database

Each tenant has its own database:

```go
resolverConfig := tenancy.ResolverConfig{
    Strategy: tenancy.StrategySeparateDatabase,
}

// Set database URL per tenant
tenant.DatabaseURL = "postgres://user:pass@host/tenant_db"
```

**Pros:**
- Complete isolation
- Easy to backup per tenant
- Can customize per tenant

**Cons:**
- More expensive
- Complex maintenance
- Connection pooling challenges

### Shared Schema

All tenants share database but have separate schemas:

```go
resolverConfig := tenancy.ResolverConfig{
    Strategy: tenancy.StrategySharedSchema,
    SharedDB: db,
}
```

**Pros:**
- Good isolation
- Easier than separate databases
- Can use same connection pool

**Cons:**
- PostgreSQL only
- Schema switching overhead

## Tenant Management

### Create Tenant

```go
tenant := &tenancy.Tenant{
    ID:         "tenant-123",
    Name:       "Acme Corp",
    Domain:     "acme",
    Plan:       "premium",
    MaxUsers:   100,
    MaxStorage: 10 * 1024 * 1024 * 1024, // 10GB
    Settings: map[string]interface{}{
        "theme": "dark",
        "logo":  "https://example.com/logo.png",
    },
}

manager.Create(ctx, tenant)
```

### Get Tenant

```go
// By ID
tenant, _ := manager.Get(ctx, "tenant-123")

// By domain
tenant, _ := manager.GetByDomain(ctx, "acme")
```

### Update Tenant

```go
tenant.MaxUsers = 200
manager.Update(ctx, tenant)
```

### Suspend/Activate

```go
// Suspend
manager.Suspend(ctx, "tenant-123")

// Activate
manager.Activate(ctx, "tenant-123")
```

## Tenant Detection

Middleware tries multiple methods in order:

### 1. Subdomain
```
https://acme.example.com/api/users
→ Tenant domain: "acme"
```

### 2. Custom Domain
```
https://custom-domain.com/api/users
→ Tenant domain: "custom-domain.com"
```

### 3. Header
```
GET /api/users
X-Tenant-ID: tenant-123
→ Tenant ID: "tenant-123"
```

### 4. Query Parameter
```
GET /api/users?tenant_id=tenant-123
→ Tenant ID: "tenant-123"
```

### 5. Path Parameter
```
GET /api/:tenant_id/users
→ Tenant ID from URL
```

## Database Scoping

### Automatic Scoping

Models with `TenantModel` embedded get automatic scoping:

```go
type User struct {
    tenancy.TenantModel  // Adds tenant_id field
    ID       uint
    Username string
    Email    string
}

// Queries are automatically scoped
db.Find(&users) // WHERE tenant_id = 'current-tenant'
```

### Manual Scoping

```go
// Get tenant from context
tenant, _ := tenancy.GetTenant(ctx)

// Apply scope
db.Scopes(tenancy.TenantScope(tenant.ID)).Find(&users)

// Or from context
db.Scopes(tenancy.TenantScopeFromContext(ctx)).Find(&users)
```

### Scoped Database

```go
scopedDB := tenancy.NewScopedDB(db, tenant.ID)

// All queries are automatically scoped
scopedDB.DB().Find(&users)
scopedDB.Create(&user)
scopedDB.Model(&User{}).Where("email = ?", email).First(&user)
```

## Quota Management

```go
// Middleware with quota check
app.Use(tenancy.QuotaMiddleware(func(tenant *tenancy.Tenant) error {
    // Check user count
    var userCount int64
    db.Model(&User{}).
        Scopes(tenancy.TenantScope(tenant.ID)).
        Count(&userCount)
    
    if userCount >= int64(tenant.MaxUsers) {
        return tenancy.ErrMaxUsersReached
    }
    
    // Check storage
    var storageUsed int64
    // ... calculate storage
    
    if storageUsed >= tenant.MaxStorage {
        return tenancy.ErrMaxStorageReached
    }
    
    return nil
}))
```

## Rate Limiting

```go
// Per-tenant rate limiter
limiter := make(map[string]*time.Time)

app.Use(tenancy.RateLimitMiddleware(func(tenantID string) bool {
    // Implement your rate limiting logic
    return true // Allow request
}))
```

## Context Usage

```go
// Store tenant in context
ctx := tenancy.WithTenant(ctx, tenant)

// Retrieve tenant
tenant, _ := tenancy.GetTenant(ctx)

// Must get (panics if not found)
tenant := tenancy.MustGetTenant(ctx)
```

## Admin Operations

```go
// Admin-only routes
admin := app.Group("/admin", tenancy.AdminMiddleware())

admin.Post("/tenants", createTenant)
admin.Get("/tenants", listTenants)
admin.Put("/tenants/:id", updateTenant)
admin.Delete("/tenants/:id", deleteTenant)
admin.Put("/tenants/:id/suspend", suspendTenant)
admin.Put("/tenants/:id/activate", activateTenant)
```

## Migration

```go
// Migrate tenant table
store.Migrate()

// Migrate tenant data tables (shared database)
tenancy.MigrateTenantTables(db, tenancy.StrategySharedDatabase,
    &User{},
    &Product{},
    &Order{},
)

// Create schema for tenant (shared schema strategy)
tenancy.CreateTenantSchema(db, tenant.ID)

// Drop schema
tenancy.DropTenantSchema(db, tenant.ID)
```

## Best Practices

1. **Choose Right Strategy**
   - Start with Shared Database for simplicity
   - Use Separate Database for high-value customers
   - Use Shared Schema for medium isolation

2. **Always Validate Tenant**
   ```go
   if err := manager.Validate(ctx, tenant); err != nil {
       // Handle suspended/expired tenants
   }
   ```

3. **Use Context Consistently**
   ```go
   ctx = tenancy.WithTenant(ctx, tenant)
   // Pass ctx to all functions
   ```

4. **Implement Quotas**
   - Max users per tenant
   - Max storage per tenant
   - Rate limits per tenant
   - Feature flags per plan

5. **Monitor Tenant Health**
   - Track usage metrics
   - Alert on quota limits
   - Monitor database performance

6. **Backup Strategy**
   - Regular backups per tenant
   - Point-in-time recovery
   - Tenant isolation in backups

## License

MIT License - Part of NeonexCore Framework

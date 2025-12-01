# Database Seeding

NeonEx Framework provides a powerful seeding system through the **SeederManager** to populate your database with initial or test data.

## Table of Contents

- [Overview](#overview)
- [Seeder Manager](#seeder-manager)
- [Creating Seeders](#creating-seeders)
- [Registering Seeders](#registering-seeders)
- [Running Seeders](#running-seeders)
- [Default Data Seeding](#default-data-seeding)
- [Module Seeders](#module-seeders)
- [Best Practices](#best-practices)
- [Advanced Patterns](#advanced-patterns)
- [Troubleshooting](#troubleshooting)

## Overview

Database seeding is the process of populating your database with initial data. Common use cases include:

- **Default system data** (roles, permissions, settings)
- **Development data** (test users, sample products)
- **Production defaults** (admin accounts, configurations)
- **Demo data** (for showcasing features)

## Seeder Manager

### Structure

```go
// core/pkg/database/seeder.go
package database

type Seeder interface {
    Name() string
    Run(ctx context.Context) error
}

type SeederManager struct {
    db      *gorm.DB
    seeders []Seeder
}

func NewSeederManager(db *gorm.DB) *SeederManager {
    return &SeederManager{
        db:      db,
        seeders: make([]Seeder, 0),
    }
}
```

### Seeder Interface

Every seeder must implement:
- `Name()` - Returns seeder name for logging
- `Run(ctx)` - Executes seeding logic

## Creating Seeders

### Basic Seeder Structure

```go
package user

import (
    "context"
    "fmt"
    "neonexcore/pkg/auth"
    "gorm.io/gorm"
)

type UserSeeder struct {
    db *gorm.DB
}

func NewUserSeeder(db *gorm.DB) *UserSeeder {
    return &UserSeeder{db: db}
}

func (s *UserSeeder) Name() string {
    return "UserSeeder"
}

func (s *UserSeeder) Run(ctx context.Context) error {
    // Check if data already exists
    var count int64
    if err := s.db.Model(&User{}).Count(&count).Error; err != nil {
        return err
    }
    
    if count > 0 {
        fmt.Println("  ⏭️  Users already seeded, skipping...")
        return nil
    }
    
    // Create seed data
    hasher := auth.NewPasswordHasher(auth.DefaultCost)
    adminPass, _ := hasher.Hash("admin123")
    
    users := []User{
        {
            Name:     "Admin User",
            Email:    "admin@neonex.local",
            Username: "admin",
            Password: adminPass,
            IsActive: true,
        },
        {
            Name:     "Test User",
            Email:    "user@neonex.local",
            Username: "testuser",
            Password: adminPass,
            IsActive: true,
        },
    }
    
    result := s.db.Create(&users)
    if result.Error != nil {
        return result.Error
    }
    
    fmt.Printf("  ✓ Seeded %d users\n", len(users))
    return nil
}
```

### Seeder with Dependencies

```go
type AdminSeeder struct {
    db          *gorm.DB
    hasher      *auth.PasswordHasher
    rbacManager *rbac.Manager
}

func NewAdminSeeder(db *gorm.DB, hasher *auth.PasswordHasher, rbac *rbac.Manager) *AdminSeeder {
    return &AdminSeeder{
        db:          db,
        hasher:      hasher,
        rbacManager: rbac,
    }
}

func (s *AdminSeeder) Name() string {
    return "AdminSeeder"
}

func (s *AdminSeeder) Run(ctx context.Context) error {
    var count int64
    if err := s.db.Model(&User{}).Where("email = ?", "superadmin@neonex.local").Count(&count).Error; err != nil {
        return err
    }
    
    if count > 0 {
        fmt.Println("  ⏭️  Super admin already exists, skipping...")
        return nil
    }
    
    // Create super admin
    hashedPassword, err := s.hasher.Hash("SuperSecure123!")
    if err != nil {
        return fmt.Errorf("failed to hash password: %w", err)
    }
    
    admin := &User{
        Name:     "Super Admin",
        Email:    "superadmin@neonex.local",
        Username: "superadmin",
        Password: hashedPassword,
        IsActive: true,
    }
    
    if err := s.db.Create(admin).Error; err != nil {
        return fmt.Errorf("failed to create admin: %w", err)
    }
    
    // Assign super-admin role
    superAdminRole, err := s.rbacManager.GetRoleBySlug(ctx, "super-admin")
    if err != nil {
        return fmt.Errorf("super-admin role not found: %w", err)
    }
    
    if err := s.rbacManager.AssignRole(ctx, admin.ID, superAdminRole.ID); err != nil {
        return fmt.Errorf("failed to assign role: %w", err)
    }
    
    fmt.Println("  ✓ Seeded super admin user")
    return nil
}
```

## Registering Seeders

### Register Method

```go
// Register a seeder
manager := database.NewSeederManager(db)
manager.Register(NewUserSeeder(db))
manager.Register(NewProductSeeder(db))
manager.Register(NewRoleSeeder(db))
```

### Module-Based Registration

```go
// modules/user/di.go
package user

func RegisterSeeders(manager *database.SeederManager, db *gorm.DB) {
    manager.Register(NewUserSeeder(db))
}
```

### Application-Wide Registration

```go
// internal/database/seeders.go
package database

import (
    "neonexcore/modules/user"
    "neonexcore/modules/product"
    "neonexcore/modules/admin"
    "neonexcore/pkg/database"
    "neonexcore/pkg/rbac"
)

func RegisterAllSeeders(manager *database.SeederManager, db *gorm.DB, rbacMgr *rbac.Manager) {
    // System seeders (run first)
    manager.Register(rbac.NewRoleSeeder(db))
    manager.Register(rbac.NewPermissionSeeder(db))
    
    // Module seeders
    manager.Register(user.NewUserSeeder(db))
    manager.Register(admin.NewAdminSeeder(db, auth.NewPasswordHasher(12), rbacMgr))
    manager.Register(product.NewProductSeeder(db))
    manager.Register(product.NewCategorySeeder(db))
}
```

## Running Seeders

### Run All Seeders

```go
func main() {
    // Initialize database
    db := database.NewConnection(config)
    
    // Create seeder manager
    seederManager := database.NewSeederManager(db)
    
    // Register seeders
    database.RegisterAllSeeders(seederManager, db, rbacManager)
    
    // Run all seeders
    ctx := context.Background()
    if err := seederManager.Run(ctx); err != nil {
        panic(fmt.Sprintf("Seeding failed: %v", err))
    }
}
```

### CLI Command for Seeding

```go
// cmd/seed/main.go
package main

import (
    "context"
    "flag"
    "fmt"
    "neonexcore/internal/database"
    "neonexcore/pkg/config"
)

func main() {
    // Parse flags
    fresh := flag.Bool("fresh", false, "Drop all tables and reseed")
    flag.Parse()
    
    // Load config
    cfg := config.Load()
    
    // Connect to database
    db := database.NewConnection(cfg.Database)
    
    if *fresh {
        fmt.Println("🗑️  Dropping all tables...")
        // Drop tables (implement as needed)
        database.DropAllTables(db)
        
        fmt.Println("📦 Running migrations...")
        database.RegisterAllModels(db)
    }
    
    // Run seeders
    fmt.Println("🌱 Seeding database...")
    seederManager := database.NewSeederManager(db)
    database.RegisterAllSeeders(seederManager, db, nil)
    
    ctx := context.Background()
    if err := seederManager.Run(ctx); err != nil {
        panic(fmt.Sprintf("Seeding failed: %v", err))
    }
    
    fmt.Println("✅ Database seeded successfully!")
}
```

Usage:
```bash
# Regular seeding (checks if data exists)
go run cmd/seed/main.go

# Fresh seeding (drops and recreates)
go run cmd/seed/main.go --fresh
```

## Default Data Seeding

### Role Seeder

```go
package rbac

import (
    "context"
    "fmt"
    "gorm.io/gorm"
)

type RoleSeeder struct {
    db *gorm.DB
}

func NewRoleSeeder(db *gorm.DB) *RoleSeeder {
    return &RoleSeeder{db: db}
}

func (s *RoleSeeder) Name() string {
    return "RoleSeeder"
}

func (s *RoleSeeder) Run(ctx context.Context) error {
    roles := []Role{
        {
            Name:        "Super Admin",
            Slug:        "super-admin",
            Description: "Full system access with all permissions",
            IsSystem:    true,
        },
        {
            Name:        "Admin",
            Slug:        "admin",
            Description: "Administrative access to manage users and content",
            IsSystem:    true,
        },
        {
            Name:        "Editor",
            Slug:        "editor",
            Description: "Can create and edit content",
            IsSystem:    true,
        },
        {
            Name:        "User",
            Slug:        "user",
            Description: "Regular user access",
            IsSystem:    true,
        },
    }
    
    for _, role := range roles {
        var existing Role
        err := s.db.Where("slug = ?", role.Slug).First(&existing).Error
        
        if err == gorm.ErrRecordNotFound {
            if err := s.db.Create(&role).Error; err != nil {
                return fmt.Errorf("failed to create role %s: %w", role.Slug, err)
            }
            fmt.Printf("  ✓ Created role: %s\n", role.Name)
        }
    }
    
    return nil
}
```

### Permission Seeder

```go
type PermissionSeeder struct {
    db *gorm.DB
}

func NewPermissionSeeder(db *gorm.DB) *PermissionSeeder {
    return &PermissionSeeder{db: db}
}

func (s *PermissionSeeder) Name() string {
    return "PermissionSeeder"
}

func (s *PermissionSeeder) Run(ctx context.Context) error {
    permissions := []Permission{
        // User permissions
        {Module: "users", Name: "View Users", Slug: "users.read", Description: "View user list"},
        {Module: "users", Name: "Create Users", Slug: "users.create", Description: "Create new users"},
        {Module: "users", Name: "Edit Users", Slug: "users.update", Description: "Edit existing users"},
        {Module: "users", Name: "Delete Users", Slug: "users.delete", Description: "Delete users"},
        
        // Product permissions
        {Module: "products", Name: "View Products", Slug: "products.read", Description: "View product list"},
        {Module: "products", Name: "Create Products", Slug: "products.create", Description: "Create new products"},
        {Module: "products", Name: "Edit Products", Slug: "products.update", Description: "Edit existing products"},
        {Module: "products", Name: "Delete Products", Slug: "products.delete", Description: "Delete products"},
        
        // Admin permissions
        {Module: "admin", Name: "Access Dashboard", Slug: "admin.dashboard", Description: "Access admin dashboard"},
        {Module: "admin", Name: "Manage Settings", Slug: "admin.settings", Description: "Manage system settings"},
        {Module: "admin", Name: "View Logs", Slug: "admin.logs", Description: "View system logs"},
    }
    
    for _, perm := range permissions {
        var existing Permission
        err := s.db.Where("slug = ?", perm.Slug).First(&existing).Error
        
        if err == gorm.ErrRecordNotFound {
            if err := s.db.Create(&perm).Error; err != nil {
                return fmt.Errorf("failed to create permission %s: %w", perm.Slug, err)
            }
        }
    }
    
    fmt.Printf("  ✓ Seeded %d permissions\n", len(permissions))
    return nil
}
```

### Role-Permission Assignment

```go
type RolePermissionSeeder struct {
    db          *gorm.DB
    rbacManager *rbac.Manager
}

func NewRolePermissionSeeder(db *gorm.DB, rbacMgr *rbac.Manager) *RolePermissionSeeder {
    return &RolePermissionSeeder{
        db:          db,
        rbacManager: rbacMgr,
    }
}

func (s *RolePermissionSeeder) Name() string {
    return "RolePermissionSeeder"
}

func (s *RolePermissionSeeder) Run(ctx context.Context) error {
    // Super Admin gets all permissions
    superAdminRole, _ := s.rbacManager.GetRoleBySlug(ctx, "super-admin")
    var allPermissions []Permission
    s.db.Find(&allPermissions)
    
    var permissionIDs []uint
    for _, perm := range allPermissions {
        permissionIDs = append(permissionIDs, perm.ID)
    }
    
    if err := s.rbacManager.SyncRolePermissions(ctx, superAdminRole.ID, permissionIDs); err != nil {
        return err
    }
    
    // Admin gets most permissions
    adminRole, _ := s.rbacManager.GetRoleBySlug(ctx, "admin")
    adminPermissions := []string{
        "users.read", "users.create", "users.update",
        "products.read", "products.create", "products.update",
        "admin.dashboard",
    }
    
    for _, slug := range adminPermissions {
        perm, _ := s.rbacManager.GetPermissionBySlug(ctx, slug)
        if perm != nil {
            s.rbacManager.AttachPermissionToRole(ctx, adminRole.ID, perm.ID)
        }
    }
    
    fmt.Println("  ✓ Assigned permissions to roles")
    return nil
}
```

## Module Seeders

### Product Module Seeder

```go
package product

import (
    "context"
    "fmt"
    "gorm.io/gorm"
)

type ProductSeeder struct {
    db *gorm.DB
}

func NewProductSeeder(db *gorm.DB) *ProductSeeder {
    return &ProductSeeder{db: db}
}

func (s *ProductSeeder) Name() string {
    return "ProductSeeder"
}

func (s *ProductSeeder) Run(ctx context.Context) error {
    var count int64
    if err := s.db.Model(&Product{}).Count(&count).Error; err != nil {
        return err
    }
    
    if count > 0 {
        fmt.Println("  ⏭️  Products already seeded, skipping...")
        return nil
    }
    
    // Get or create categories first
    electronics := &Category{Name: "Electronics", Slug: "electronics"}
    s.db.FirstOrCreate(electronics, Category{Slug: "electronics"})
    
    books := &Category{Name: "Books", Slug: "books"}
    s.db.FirstOrCreate(books, Category{Slug: "books"})
    
    products := []Product{
        {
            Name:        "Laptop Pro",
            Slug:        "laptop-pro",
            Description: "High-performance laptop for professionals",
            Price:       1299.99,
            Stock:       50,
            CategoryID:  electronics.ID,
            IsActive:    true,
        },
        {
            Name:        "Wireless Mouse",
            Slug:        "wireless-mouse",
            Description: "Ergonomic wireless mouse",
            Price:       29.99,
            Stock:       200,
            CategoryID:  electronics.ID,
            IsActive:    true,
        },
        {
            Name:        "Go Programming Book",
            Slug:        "go-programming-book",
            Description: "Complete guide to Go programming",
            Price:       49.99,
            Stock:       100,
            CategoryID:  books.ID,
            IsActive:    true,
        },
    }
    
    result := s.db.Create(&products)
    if result.Error != nil {
        return result.Error
    }
    
    fmt.Printf("  ✓ Seeded %d products\n", len(products))
    return nil
}
```

## Best Practices

### 1. Check Before Insert

Always check if data exists to make seeders idempotent:

```go
func (s *UserSeeder) Run(ctx context.Context) error {
    // Check if data exists
    var count int64
    s.db.Model(&User{}).Where("email = ?", "admin@example.com").Count(&count)
    
    if count > 0 {
        fmt.Println("  ⏭️  Admin user already exists, skipping...")
        return nil
    }
    
    // Seed data
    return s.db.Create(&User{...}).Error
}
```

### 2. Use FirstOrCreate

```go
func (s *CategorySeeder) Run(ctx context.Context) error {
    categories := []Category{
        {Name: "Electronics", Slug: "electronics"},
        {Name: "Books", Slug: "books"},
    }
    
    for _, cat := range categories {
        var category Category
        s.db.FirstOrCreate(&category, Category{Slug: cat.Slug})
        fmt.Printf("  ✓ Ensured category: %s\n", cat.Name)
    }
    
    return nil
}
```

### 3. Seed in Correct Order

```go
// Seed in dependency order
func RegisterAllSeeders(manager *database.SeederManager, db *gorm.DB) {
    // 1. Foundation data
    manager.Register(NewRoleSeeder(db))
    manager.Register(NewPermissionSeeder(db))
    
    // 2. User data (depends on roles)
    manager.Register(NewUserSeeder(db))
    
    // 3. Content data
    manager.Register(NewCategorySeeder(db))
    manager.Register(NewProductSeeder(db)) // Depends on categories
    
    // 4. Relationships
    manager.Register(NewRolePermissionSeeder(db))
}
```

### 4. Use Transactions for Related Data

```go
func (s *OrderSeeder) Run(ctx context.Context) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // Create order
        order := &Order{UserID: 1, Total: 100}
        if err := tx.Create(order).Error; err != nil {
            return err
        }
        
        // Create order items
        items := []OrderItem{
            {OrderID: order.ID, ProductID: 1, Quantity: 2},
            {OrderID: order.ID, ProductID: 2, Quantity: 1},
        }
        return tx.Create(&items).Error
    })
}
```

### 5. Environment-Specific Seeding

```go
func (s *UserSeeder) Run(ctx context.Context) error {
    env := os.Getenv("APP_ENV")
    
    if env == "production" {
        // Only seed essential data in production
        return s.seedProductionUsers()
    }
    
    // Seed test data in development
    return s.seedDevelopmentUsers()
}

func (s *UserSeeder) seedProductionUsers() error {
    // Only admin user
    admin := &User{
        Email:    "admin@company.com",
        Password: os.Getenv("ADMIN_PASSWORD"),
        IsActive: true,
    }
    return s.db.FirstOrCreate(admin, User{Email: admin.Email}).Error
}

func (s *UserSeeder) seedDevelopmentUsers() error {
    // Multiple test users
    users := []User{
        {Email: "admin@test.com", Password: "test123"},
        {Email: "user1@test.com", Password: "test123"},
        {Email: "user2@test.com", Password: "test123"},
    }
    return s.db.Create(&users).Error
}
```

## Advanced Patterns

### Seeder with Faker Data

```go
import "github.com/brianvoe/gofakeit/v6"

type ProductSeeder struct {
    db    *gorm.DB
    count int
}

func NewProductSeeder(db *gorm.DB, count int) *ProductSeeder {
    return &ProductSeeder{db: db, count: count}
}

func (s *ProductSeeder) Run(ctx context.Context) error {
    products := make([]Product, s.count)
    
    for i := 0; i < s.count; i++ {
        products[i] = Product{
            Name:        gofakeit.ProductName(),
            Description: gofakeit.Sentence(10),
            Price:       gofakeit.Float64Range(10, 1000),
            Stock:       gofakeit.Number(0, 500),
            IsActive:    true,
        }
    }
    
    return s.db.CreateInBatches(products, 100).Error
}
```

### CSV Data Seeder

```go
import (
    "encoding/csv"
    "os"
)

type CSVProductSeeder struct {
    db       *gorm.DB
    filePath string
}

func (s *CSVProductSeeder) Run(ctx context.Context) error {
    file, err := os.Open(s.filePath)
    if err != nil {
        return err
    }
    defer file.Close()
    
    reader := csv.NewReader(file)
    records, err := reader.ReadAll()
    if err != nil {
        return err
    }
    
    for _, record := range records[1:] { // Skip header
        product := &Product{
            Name:  record[0],
            Price: parseFloat(record[1]),
            Stock: parseInt(record[2]),
        }
        s.db.FirstOrCreate(product, Product{Name: product.Name})
    }
    
    return nil
}
```

## Troubleshooting

### Common Issues

**Issue: Foreign key constraint fails**

```go
// Solution: Seed in correct order
func RegisterSeeders(manager *database.SeederManager, db *gorm.DB) {
    // Seed parent entities first
    manager.Register(NewCategorySeeder(db))
    manager.Register(NewUserSeeder(db))
    
    // Then child entities
    manager.Register(NewProductSeeder(db)) // References Category
    manager.Register(NewOrderSeeder(db))   // References User
}
```

**Issue: Duplicate entry errors**

```go
// Solution: Use FirstOrCreate or check existence
var user User
s.db.FirstOrCreate(&user, User{Email: "admin@example.com"})
```

**Issue: Seeder runs every time**

```go
// Solution: Check before creating
var count int64
s.db.Model(&User{}).Count(&count)
if count > 0 {
    return nil // Skip if data exists
}
```

## Summary

NeonEx Framework's seeding system provides:

✅ **Simple interface** with Name() and Run() methods  
✅ **Idempotent seeders** that check before inserting  
✅ **Module-based organization** for maintainability  
✅ **Default data support** for roles, permissions, and admin users  
✅ **Production-ready** with environment-specific seeding

For more information:
- [Database Configuration](configuration.md)
- [Migrations](migrations.md)
- [Models](models.md)

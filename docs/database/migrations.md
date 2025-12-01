# Database Migrations

NeonEx Framework uses **GORM AutoMigrate** for database schema management, providing automatic table creation and schema updates based on your Go models.

## Table of Contents

- [Overview](#overview)
- [Auto-Migration Basics](#auto-migration-basics)
- [Registering Models](#registering-models)
- [Migration Patterns](#migration-patterns)
- [Manual Migrations](#manual-migrations)
- [Best Practices](#best-practices)
- [Advanced Usage](#advanced-usage)
- [Troubleshooting](#troubleshooting)

## Overview

NeonEx uses GORM's AutoMigrate feature which:
- Automatically creates tables based on struct definitions
- Adds missing columns when models are updated
- Does NOT delete existing columns (safe by default)
- Does NOT modify existing column types
- Creates indexes and foreign keys

### When Migrations Run

Migrations typically run during application startup after database connection is established.

## Auto-Migration Basics

### Simple Auto-Migration

```go
package main

import (
    "neonexcore/pkg/database"
    "gorm.io/gorm"
)

type User struct {
    gorm.Model
    Name     string `gorm:"size:100;not null"`
    Email    string `gorm:"size:255;unique;not null"`
    Password string `gorm:"size:255;not null"`
    IsActive bool   `gorm:"default:true"`
}

func main() {
    // Initialize database connection
    db := database.NewConnection(config)
    
    // Run auto-migration
    err := db.AutoMigrate(&User{})
    if err != nil {
        panic("Failed to migrate database: " + err.Error())
    }
}
```

### Multiple Models

```go
// Migrate multiple models at once
err := db.AutoMigrate(
    &User{},
    &Product{},
    &Order{},
    &OrderItem{},
)
```

## Registering Models

### Module-Based Registration

NeonEx modules can register their models through the module initialization:

```go
// modules/user/model.go
package user

import (
    "time"
    "gorm.io/gorm"
)

type User struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
    
    Name         string     `gorm:"size:100;not null" json:"name"`
    Email        string     `gorm:"size:255;unique;not null" json:"email"`
    Username     string     `gorm:"size:50;unique;not null" json:"username"`
    Password     string     `gorm:"size:255;not null" json:"-"`
    IsActive     bool       `gorm:"default:true" json:"is_active"`
    LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
    APIKey       *string    `gorm:"size:100;unique" json:"api_key,omitempty"`
}

// TableName specifies the table name
func (User) TableName() string {
    return "users"
}
```

```go
// modules/user/di.go
package user

import (
    "neonexcore/internal/core"
    "gorm.io/gorm"
)

func RegisterModels(db *gorm.DB) error {
    // Register models for migration
    return db.AutoMigrate(
        &User{},
    )
}

func SetupDependencies(container *core.Container) error {
    // Register models during module initialization
    db := container.GetDatabase()
    return RegisterModels(db)
}
```

### Centralized Model Registration

```go
// internal/database/models.go
package database

import (
    "neonexcore/modules/user"
    "neonexcore/modules/product"
    "neonexcore/modules/admin"
    "neonexcore/pkg/rbac"
    "gorm.io/gorm"
)

// RegisterAllModels registers all application models
func RegisterAllModels(db *gorm.DB) error {
    models := []interface{}{
        // User module
        &user.User{},
        
        // Product module
        &product.Product{},
        &product.Category{},
        
        // Admin module
        &admin.Admin{},
        
        // RBAC
        &rbac.Role{},
        &rbac.Permission{},
        &rbac.UserRole{},
        &rbac.UserPermission{},
        &rbac.RolePermission{},
    }
    
    return db.AutoMigrate(models...)
}
```

## Migration Patterns

### Basic Field Types

```go
type User struct {
    gorm.Model
    
    // String fields
    Name     string `gorm:"size:100;not null"`
    Email    string `gorm:"size:255;unique;not null"`
    Bio      string `gorm:"type:text"`
    
    // Numeric fields
    Age      int    `gorm:"default:0"`
    Balance  float64
    
    // Boolean fields
    IsActive bool   `gorm:"default:true"`
    IsAdmin  bool   `gorm:"default:false"`
    
    // Time fields
    BirthDate    time.Time
    LastLoginAt  *time.Time
    
    // JSON fields
    Metadata datatypes.JSON `gorm:"type:json"`
}
```

### Indexes

```go
type User struct {
    gorm.Model
    
    // Single column index
    Email    string `gorm:"uniqueIndex"`
    Username string `gorm:"index"`
    
    // Named index
    Status   string `gorm:"index:idx_status"`
    
    // Composite index (defined at struct level)
    FirstName string `gorm:"index:idx_name"`
    LastName  string `gorm:"index:idx_name"`
}

// Alternative: Composite index using struct tags
type User struct {
    gorm.Model
    FirstName string `gorm:"index:idx_full_name,priority:1"`
    LastName  string `gorm:"index:idx_full_name,priority:2"`
}
```

### Foreign Keys and Relations

```go
// One-to-One
type User struct {
    gorm.Model
    Profile Profile `gorm:"foreignKey:UserID"`
}

type Profile struct {
    gorm.Model
    UserID uint
    Bio    string
}

// One-to-Many
type User struct {
    gorm.Model
    Orders []Order `gorm:"foreignKey:UserID"`
}

type Order struct {
    gorm.Model
    UserID uint
    Total  float64
}

// Many-to-Many
type User struct {
    gorm.Model
    Roles []Role `gorm:"many2many:user_roles;"`
}

type Role struct {
    gorm.Model
    Name  string
    Users []User `gorm:"many2many:user_roles;"`
}
```

### Polymorphic Relations

```go
type Comment struct {
    gorm.Model
    CommentableID   uint
    CommentableType string
    Content         string
}

type Post struct {
    gorm.Model
    Title    string
    Comments []Comment `gorm:"polymorphic:Commentable;"`
}

type Video struct {
    gorm.Model
    Title    string
    Comments []Comment `gorm:"polymorphic:Commentable;"`
}
```

## Manual Migrations

### Checking Migration Status

```go
// Check if table exists
if db.Migrator().HasTable(&User{}) {
    fmt.Println("Users table exists")
}

// Check if column exists
if db.Migrator().HasColumn(&User{}, "email") {
    fmt.Println("Email column exists")
}

// Check if index exists
if db.Migrator().HasIndex(&User{}, "idx_email") {
    fmt.Println("Email index exists")
}
```

### Adding Columns

```go
// Add a new column
if !db.Migrator().HasColumn(&User{}, "phone") {
    err := db.Migrator().AddColumn(&User{}, "phone")
    if err != nil {
        return fmt.Errorf("failed to add phone column: %w", err)
    }
}
```

### Renaming Columns

```go
// Rename a column
err := db.Migrator().RenameColumn(&User{}, "name", "full_name")
if err != nil {
    return fmt.Errorf("failed to rename column: %w", err)
}
```

### Dropping Columns

```go
// Drop a column (use with caution)
err := db.Migrator().DropColumn(&User{}, "deprecated_field")
if err != nil {
    return fmt.Errorf("failed to drop column: %w", err)
}
```

### Managing Indexes

```go
// Create index
err := db.Migrator().CreateIndex(&User{}, "Email")

// Drop index
err := db.Migrator().DropIndex(&User{}, "idx_email")

// Rename index
err := db.Migrator().RenameIndex(&User{}, "old_idx", "new_idx")
```

## Best Practices

### 1. Version Control Models

Always keep model definitions in version control alongside your code.

```go
// Good: Models evolve with code
type User struct {
    gorm.Model
    Name     string `gorm:"size:100;not null"` // v1.0
    Email    string `gorm:"size:255;unique;not null"` // v1.0
    Phone    string `gorm:"size:20"` // Added in v1.1
}
```

### 2. Use Soft Deletes

```go
type User struct {
    ID        uint           `gorm:"primarykey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"` // Enables soft delete
    
    Name  string
    Email string
}

// Records are marked as deleted, not removed
db.Delete(&user) // Sets DeletedAt

// Query excludes soft-deleted records
db.Find(&users) // Excludes deleted

// Include soft-deleted records
db.Unscoped().Find(&users)
```

### 3. Add Default Values

```go
type User struct {
    gorm.Model
    IsActive  bool   `gorm:"default:true"`
    Role      string `gorm:"default:user"`
    Status    string `gorm:"default:pending"`
    CreatedBy uint   `gorm:"default:0"`
}
```

### 4. Validate Before Migration

```go
func ValidateAndMigrate(db *gorm.DB) error {
    // Test database connection
    sqlDB, err := db.DB()
    if err != nil {
        return fmt.Errorf("database connection error: %w", err)
    }
    
    if err := sqlDB.Ping(); err != nil {
        return fmt.Errorf("database ping failed: %w", err)
    }
    
    // Run migrations
    models := []interface{}{
        &User{},
        &Product{},
    }
    
    if err := db.AutoMigrate(models...); err != nil {
        return fmt.Errorf("migration failed: %w", err)
    }
    
    fmt.Println("✅ Database migration completed successfully")
    return nil
}
```

### 5. Separate Migration from Application

```go
// cmd/migrate/main.go
package main

import (
    "fmt"
    "neonexcore/internal/database"
    "neonexcore/pkg/config"
)

func main() {
    // Load configuration
    cfg := config.Load()
    
    // Connect to database
    db := database.NewConnection(cfg.Database)
    
    // Run migrations
    if err := database.RegisterAllModels(db); err != nil {
        panic(fmt.Sprintf("Migration failed: %v", err))
    }
    
    fmt.Println("✅ Migration completed successfully")
}
```

Run migrations separately:
```bash
go run cmd/migrate/main.go
```

## Advanced Usage

### Migration with Transaction

```go
func MigrateWithTransaction(db *gorm.DB) error {
    return db.Transaction(func(tx *gorm.DB) error {
        // Migrate all models in transaction
        if err := tx.AutoMigrate(&User{}, &Product{}); err != nil {
            return err
        }
        
        // Run additional SQL
        if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_custom ON users(email, username)").Error; err != nil {
            return err
        }
        
        return nil
    })
}
```

### Conditional Migrations

```go
func ConditionalMigrate(db *gorm.DB) error {
    // Only migrate if table doesn't exist
    if !db.Migrator().HasTable(&User{}) {
        fmt.Println("Creating users table...")
        if err := db.AutoMigrate(&User{}); err != nil {
            return err
        }
    }
    
    // Add column if missing
    if !db.Migrator().HasColumn(&User{}, "phone") {
        fmt.Println("Adding phone column...")
        if err := db.Migrator().AddColumn(&User{}, "phone"); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Custom SQL Migrations

```go
func CustomMigration(db *gorm.DB) error {
    // Create extension
    if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
        return err
    }
    
    // Create custom type
    if err := db.Exec(`
        DO $$ BEGIN
            CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended');
        EXCEPTION
            WHEN duplicate_object THEN null;
        END $$;
    `).Error; err != nil {
        return err
    }
    
    return nil
}
```

### Migration Logging

```go
import (
    "neonexcore/pkg/logger"
)

func MigrateWithLogging(db *gorm.DB) error {
    log := logger.NewLogger()
    
    log.Info("Starting database migration...")
    
    models := []interface{}{
        &User{},
        &Product{},
    }
    
    for _, model := range models {
        modelName := fmt.Sprintf("%T", model)
        log.Info("Migrating model", logger.Fields{"model": modelName})
        
        if err := db.AutoMigrate(model); err != nil {
            log.Error("Migration failed", logger.Fields{
                "model": modelName,
                "error": err.Error(),
            })
            return err
        }
    }
    
    log.Info("Database migration completed successfully")
    return nil
}
```

## Rollback Strategies

### Manual Rollback

Since GORM AutoMigrate doesn't support automatic rollbacks, implement manual rollback:

```go
func RollbackMigration(db *gorm.DB) error {
    // Drop tables in reverse order (to respect foreign keys)
    return db.Migrator().DropTable(
        &OrderItem{},
        &Order{},
        &Product{},
        &User{},
    )
}
```

### Backup Before Migration

```go
func MigrateWithBackup(db *gorm.DB) error {
    // Create backup
    timestamp := time.Now().Format("20060102_150405")
    backupName := fmt.Sprintf("backup_%s", timestamp)
    
    if err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", backupName)).Error; err != nil {
        return fmt.Errorf("backup failed: %w", err)
    }
    
    // Run migration
    if err := db.AutoMigrate(&User{}); err != nil {
        return fmt.Errorf("migration failed: %w", err)
    }
    
    return nil
}
```

### Version Tracking

```go
type Migration struct {
    gorm.Model
    Version     string `gorm:"unique;not null"`
    Description string
    AppliedAt   time.Time
}

func TrackMigration(db *gorm.DB, version, description string) error {
    migration := &Migration{
        Version:     version,
        Description: description,
        AppliedAt:   time.Now(),
    }
    return db.Create(migration).Error
}

func IsMigrationApplied(db *gorm.DB, version string) bool {
    var count int64
    db.Model(&Migration{}).Where("version = ?", version).Count(&count)
    return count > 0
}
```

## Troubleshooting

### Common Issues

**Issue: Migration doesn't add new column**

```go
// Problem: AutoMigrate may not detect changes
type User struct {
    gorm.Model
    Name  string
    Email string
    Phone string // Newly added
}

// Solution 1: Explicitly add column
db.Migrator().AddColumn(&User{}, "Phone")

// Solution 2: Drop and recreate (loses data)
db.Migrator().DropTable(&User{})
db.AutoMigrate(&User{})
```

**Issue: Foreign key constraint fails**

```go
// Problem: Order of migration matters
db.AutoMigrate(&Order{}, &User{}) // ❌ Fails if Order references User

// Solution: Migrate in correct order
db.AutoMigrate(&User{}, &Order{}) // ✅ User first, then Order
```

**Issue: Duplicate column/index names**

```go
// Problem: Naming collision
type User struct {
    Email string `gorm:"index:idx_email"`
}

type Admin struct {
    Email string `gorm:"index:idx_email"` // Same index name!
}

// Solution: Use unique index names
type User struct {
    Email string `gorm:"index:idx_user_email"`
}

type Admin struct {
    Email string `gorm:"index:idx_admin_email"`
}
```

### Debug Migration Issues

```go
import (
    "gorm.io/gorm/logger"
)

func DebugMigration(db *gorm.DB) error {
    // Enable detailed logging
    db = db.Debug()
    
    // Set log level
    db.Logger = logger.Default.LogMode(logger.Info)
    
    // Run migration with logging
    return db.AutoMigrate(&User{})
}
```

### Performance Tips

```go
// 1. Disable foreign key checks during bulk migration (MySQL)
db.Exec("SET FOREIGN_KEY_CHECKS=0")
db.AutoMigrate(&User{}, &Order{}, &Product{})
db.Exec("SET FOREIGN_KEY_CHECKS=1")

// 2. Create indexes after data insertion
db.Migrator().DropIndex(&User{}, "idx_email")
// ... insert large amounts of data ...
db.Migrator().CreateIndex(&User{}, "Email")
```

## Summary

NeonEx Framework's migration system provides:

✅ **Automatic schema management** with GORM AutoMigrate  
✅ **Safe migrations** that add but don't delete  
✅ **Module-based model registration**  
✅ **Flexible manual migration support**  
✅ **Production-ready patterns** for versioning and rollback

For more information:
- [Database Configuration](configuration.md)
- [Models](models.md)
- [Repository Pattern](repository.md)

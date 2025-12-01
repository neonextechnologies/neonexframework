# Database Configuration

NeonEx Framework provides flexible database configuration with support for multiple database drivers, connection pooling, and easy setup. Built on GORM, it offers a powerful ORM with excellent performance.

## Table of Contents

- [Supported Databases](#supported-databases)
- [Configuration](#configuration)
- [Connection Setup](#connection-setup)
- [Connection Pooling](#connection-pooling)
- [Multiple Databases](#multiple-databases)
- [Environment Variables](#environment-variables)
- [Best Practices](#best-practices)

## Supported Databases

NeonEx supports multiple database drivers:

- **SQLite** - Lightweight, file-based database
- **MySQL** - Popular relational database
- **PostgreSQL** - Advanced open-source database
- **Turso** - Distributed SQLite (libSQL)

## Configuration

### Database Config Structure

```go
type DatabaseConfig struct {
    Driver          string        // Database driver (sqlite, mysql, postgres)
    Host            string        // Database host
    Port            string        // Database port
    Username        string        // Database username
    Password        string        // Database password
    Database        string        // Database name
    Charset         string        // Character set (MySQL)
    ParseTime       string        // Parse time (MySQL)
    Loc             string        // Location/timezone
    MaxIdleConns    int           // Maximum idle connections
    MaxOpenConns    int           // Maximum open connections
    ConnMaxLifetime time.Duration // Connection maximum lifetime
    LogLevel        logger.LogLevel // GORM log level
}
```

### Loading Configuration

```go
import "neonexcore/internal/config"

// Load from environment variables
dbConfig := config.LoadDatabaseConfig()

// Initialize database
dbManager, err := config.InitDatabase(dbConfig)
if err != nil {
    log.Fatal("Failed to connect to database:", err)
}

// Get database instance
db := dbManager.GetDB()
```

## Connection Setup

### SQLite

```go
// .env
DB_DRIVER=sqlite
DB_DATABASE=neonex.db

// Initialize
dbConfig := &config.DatabaseConfig{
    Driver:   "sqlite",
    Database: "neonex.db",
    LogLevel: logger.Info,
}

dbManager, err := config.InitDatabase(dbConfig)
```

### MySQL

```go
// .env
DB_DRIVER=mysql
DB_HOST=localhost
DB_PORT=3306
DB_USERNAME=root
DB_PASSWORD=secret
DB_DATABASE=neonex
DB_CHARSET=utf8mb4
DB_PARSE_TIME=True
DB_LOC=Local

// Initialize
dbConfig := &config.DatabaseConfig{
    Driver:   "mysql",
    Host:     "localhost",
    Port:     "3306",
    Username: "root",
    Password: "secret",
    Database: "neonex",
    Charset:  "utf8mb4",
    ParseTime: "True",
    Loc:      "Local",
    MaxIdleConns: 10,
    MaxOpenConns: 100,
    ConnMaxLifetime: time.Hour,
    LogLevel: logger.Info,
}

dbManager, err := config.InitDatabase(dbConfig)
```

### PostgreSQL

```go
// .env
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=secret
DB_DATABASE=neonex

// Initialize
dbConfig := &config.DatabaseConfig{
    Driver:   "postgres",
    Host:     "localhost",
    Port:     "5432",
    Username: "postgres",
    Password: "secret",
    Database: "neonex",
    MaxIdleConns: 10,
    MaxOpenConns: 100,
    ConnMaxLifetime: time.Hour,
    LogLevel: logger.Info,
}

dbManager, err := config.InitDatabase(dbConfig)
```

### Turso (Distributed SQLite)

```go
// .env
DB_DRIVER=turso
DB_DATABASE=libsql://your-db-org.turso.io?authToken=your-token

// Initialize
dbConfig := &config.DatabaseConfig{
    Driver:   "turso",
    Database: "libsql://your-db-org.turso.io?authToken=your-token",
    LogLevel: logger.Info,
}

dbManager, err := config.InitDatabase(dbConfig)
```

## Connection Pooling

### Pool Configuration

```go
dbConfig := &config.DatabaseConfig{
    Driver:   "mysql",
    // ... other settings
    
    // Connection pool settings
    MaxIdleConns:    10,              // Maximum idle connections
    MaxOpenConns:    100,             // Maximum open connections
    ConnMaxLifetime: time.Hour,       // Connection lifetime
}

dbManager, err := config.InitDatabase(dbConfig)
```

### Adjusting Pool Settings

```go
// Get underlying SQL database
sqlDB, err := db.DB()
if err != nil {
    log.Fatal(err)
}

// Configure connection pool
sqlDB.SetMaxIdleConns(25)
sqlDB.SetMaxOpenConns(200)
sqlDB.SetConnMaxLifetime(time.Hour * 2)
sqlDB.SetConnMaxIdleTime(time.Minute * 10)
```

### Checking Pool Statistics

```go
func getPoolStats(db *gorm.DB) {
    sqlDB, _ := db.DB()
    stats := sqlDB.Stats()
    
    log.Printf("Max Open Connections: %d", stats.MaxOpenConnections)
    log.Printf("Open Connections: %d", stats.OpenConnections)
    log.Printf("In Use: %d", stats.InUse)
    log.Printf("Idle: %d", stats.Idle)
    log.Printf("Wait Count: %d", stats.WaitCount)
    log.Printf("Wait Duration: %v", stats.WaitDuration)
}
```

## Multiple Databases

### Primary and Secondary Databases

```go
package main

type DatabaseConnections struct {
    Primary   *gorm.DB
    Secondary *gorm.DB
    Analytics *gorm.DB
}

func initDatabases() (*DatabaseConnections, error) {
    // Primary database (MySQL)
    primaryConfig := &config.DatabaseConfig{
        Driver:   "mysql",
        Host:     "primary-db.example.com",
        Port:     "3306",
        Username: "app_user",
        Password: "secret",
        Database: "app_production",
    }
    
    primaryManager, err := config.InitDatabase(primaryConfig)
    if err != nil {
        return nil, fmt.Errorf("primary db: %w", err)
    }
    
    // Secondary database (PostgreSQL for analytics)
    secondaryConfig := &config.DatabaseConfig{
        Driver:   "postgres",
        Host:     "analytics-db.example.com",
        Port:     "5432",
        Username: "analytics_user",
        Password: "secret",
        Database: "analytics",
    }
    
    secondaryManager, err := config.InitDatabase(secondaryConfig)
    if err != nil {
        return nil, fmt.Errorf("secondary db: %w", err)
    }
    
    return &DatabaseConnections{
        Primary:   primaryManager.GetDB(),
        Secondary: secondaryManager.GetDB(),
    }, nil
}
```

### Using Multiple Databases

```go
type UserService struct {
    primaryDB   *gorm.DB
    analyticsDB *gorm.DB
}

func (s *UserService) CreateUser(user *User) error {
    // Create in primary database
    if err := s.primaryDB.Create(user).Error; err != nil {
        return err
    }
    
    // Log to analytics database
    analyticsRecord := &UserAnalytics{
        UserID:    user.ID,
        Action:    "user_created",
        CreatedAt: time.Now(),
    }
    
    s.analyticsDB.Create(analyticsRecord)
    
    return nil
}
```

### Database Routing

```go
type DatabaseRouter struct {
    primary  *gorm.DB
    replicas []*gorm.DB
    index    int
}

func (r *DatabaseRouter) Write() *gorm.DB {
    return r.primary
}

func (r *DatabaseRouter) Read() *gorm.DB {
    if len(r.replicas) == 0 {
        return r.primary
    }
    
    // Round-robin load balancing
    r.index = (r.index + 1) % len(r.replicas)
    return r.replicas[r.index]
}

// Usage
func (s *UserService) GetUser(id uint) (*User, error) {
    var user User
    err := dbRouter.Read().First(&user, id).Error
    return &user, err
}

func (s *UserService) UpdateUser(user *User) error {
    return dbRouter.Write().Save(user).Error
}
```

## Environment Variables

### Default Configuration

```env
# Database Driver (sqlite, mysql, postgres, turso)
DB_DRIVER=sqlite

# SQLite Configuration
DB_DATABASE=neonex.db

# MySQL/PostgreSQL Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USERNAME=root
DB_PASSWORD=
DB_DATABASE=neonex

# MySQL Specific
DB_CHARSET=utf8mb4
DB_PARSE_TIME=True
DB_LOC=Local

# Connection Pool (optional)
DB_MAX_IDLE_CONNS=10
DB_MAX_OPEN_CONNS=100
DB_CONN_MAX_LIFETIME=3600  # seconds
```

### Development Configuration

```env
DB_DRIVER=sqlite
DB_DATABASE=dev.db
DB_LOG_LEVEL=debug
```

### Production Configuration

```env
DB_DRIVER=mysql
DB_HOST=prod-db.example.com
DB_PORT=3306
DB_USERNAME=prod_user
DB_PASSWORD=strong_password_here
DB_DATABASE=app_production
DB_CHARSET=utf8mb4
DB_PARSE_TIME=True
DB_LOC=UTC

# Production pool settings
DB_MAX_IDLE_CONNS=25
DB_MAX_OPEN_CONNS=200
DB_CONN_MAX_LIFETIME=7200

# Minimal logging in production
DB_LOG_LEVEL=error
```

### Docker Configuration

```env
DB_DRIVER=postgres
DB_HOST=postgres
DB_PORT=5432
DB_USERNAME=neonex
DB_PASSWORD=neonex_password
DB_DATABASE=neonex_db
```

## Best Practices

### 1. Use Environment Variables

```go
// Good: Load from environment
dbConfig := config.LoadDatabaseConfig()
dbManager, err := config.InitDatabase(dbConfig)

// Bad: Hardcoded credentials
dbConfig := &config.DatabaseConfig{
    Username: "root",
    Password: "password123", // Never hardcode credentials!
}
```

### 2. Configure Appropriate Pool Sizes

```go
// For web applications
dbConfig := &config.DatabaseConfig{
    MaxIdleConns:    10,  // Minimum connections to keep
    MaxOpenConns:    100, // Maximum concurrent connections
    ConnMaxLifetime: time.Hour,
}

// For background workers
dbConfig := &config.DatabaseConfig{
    MaxIdleConns:    2,   // Fewer idle connections
    MaxOpenConns:    20,  // Lower max connections
    ConnMaxLifetime: time.Hour * 2,
}

// For high-traffic applications
dbConfig := &config.DatabaseConfig{
    MaxIdleConns:    50,  // More idle connections
    MaxOpenConns:    300, // Higher max connections
    ConnMaxLifetime: time.Minute * 30,
}
```

### 3. Set Appropriate Log Levels

```go
// Development
dbConfig.LogLevel = logger.Info

// Production
dbConfig.LogLevel = logger.Error

// Debugging
dbConfig.LogLevel = logger.Silent // No logs
```

### 4. Handle Connection Errors Gracefully

```go
func connectWithRetry(config *config.DatabaseConfig, maxRetries int) (*config.DatabaseManager, error) {
    var dbManager *config.DatabaseManager
    var err error
    
    for i := 0; i < maxRetries; i++ {
        dbManager, err = config.InitDatabase(config)
        if err == nil {
            return dbManager, nil
        }
        
        log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
        time.Sleep(time.Second * time.Duration(i+1))
    }
    
    return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, err)
}
```

### 5. Check Database Health

```go
func checkDatabaseHealth(dbManager *config.DatabaseManager) error {
    // Ping database
    if err := dbManager.Ping(); err != nil {
        return fmt.Errorf("database ping failed: %w", err)
    }
    
    // Check pool stats
    sqlDB, err := dbManager.GetDB().DB()
    if err != nil {
        return fmt.Errorf("failed to get sql.DB: %w", err)
    }
    
    stats := sqlDB.Stats()
    if stats.OpenConnections >= stats.MaxOpenConnections {
        log.Warn("Database connection pool is at maximum capacity")
    }
    
    return nil
}

// Health check endpoint
func healthCheck(c *fiber.Ctx) error {
    if err := checkDatabaseHealth(dbManager); err != nil {
        return c.Status(503).JSON(fiber.Map{
            "status": "unhealthy",
            "error":  err.Error(),
        })
    }
    
    return c.JSON(fiber.Map{
        "status": "healthy",
        "database": "connected",
    })
}
```

### 6. Close Connections Properly

```go
func main() {
    dbManager, err := config.InitDatabase(dbConfig)
    if err != nil {
        log.Fatal(err)
    }
    
    // Ensure database connection is closed
    defer func() {
        if err := dbManager.Close(); err != nil {
            log.Printf("Error closing database: %v", err)
        }
    }()
    
    // Application code...
}
```

### 7. Use Database Transactions for Critical Operations

```go
func (s *UserService) CreateUserWithProfile(user *User, profile *Profile) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // Create user
        if err := tx.Create(user).Error; err != nil {
            return err
        }
        
        // Create profile
        profile.UserID = user.ID
        if err := tx.Create(profile).Error; err != nil {
            return err
        }
        
        return nil
    })
}
```

## Complete Example

```go
package main

import (
    "log"
    "neonexcore/internal/config"
    "neonexcore/pkg/logger"
    "github.com/gofiber/fiber/v2"
)

func main() {
    // Load database configuration
    dbConfig := config.LoadDatabaseConfig()
    
    // Initialize database with retry
    dbManager, err := connectWithRetry(dbConfig, 5)
    if err != nil {
        log.Fatal("Failed to initialize database:", err)
    }
    defer dbManager.Close()
    
    // Get database instance
    db := dbManager.GetDB()
    
    // Auto-migrate models
    if err := db.AutoMigrate(&User{}, &Post{}, &Comment{}); err != nil {
        log.Fatal("Failed to migrate database:", err)
    }
    
    // Initialize Fiber app
    app := fiber.New()
    
    // Health check endpoint
    app.Get("/health", func(c *fiber.Ctx) error {
        if err := checkDatabaseHealth(dbManager); err != nil {
            return c.Status(503).JSON(fiber.Map{
                "status": "unhealthy",
                "error":  err.Error(),
            })
        }
        return c.JSON(fiber.Map{"status": "healthy"})
    })
    
    // Start server
    log.Fatal(app.Listen(":3000"))
}

func connectWithRetry(config *config.DatabaseConfig, maxRetries int) (*config.DatabaseManager, error) {
    var dbManager *config.DatabaseManager
    var err error
    
    for i := 0; i < maxRetries; i++ {
        dbManager, err = config.InitDatabase(config)
        if err == nil {
            log.Println("✅ Database connected successfully")
            return dbManager, nil
        }
        
        log.Printf("❌ Failed to connect (attempt %d/%d): %v", i+1, maxRetries, err)
        time.Sleep(time.Second * time.Duration(i+1))
    }
    
    return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, err)
}
```

This comprehensive guide covers everything you need to configure and manage databases in NeonEx Framework!

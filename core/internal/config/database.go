package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseConfig struct {
	Driver          string
	Host            string
	Port            string
	Username        string
	Password        string
	Database        string
	Charset         string
	ParseTime       string
	Loc             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	LogLevel        logger.LogLevel
}

type DatabaseManager struct {
	db     *gorm.DB
	config *DatabaseConfig
}

var DB *DatabaseManager

// LoadDatabaseConfig loads database configuration from environment
func LoadDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Driver:          getEnv("DB_DRIVER", "sqlite"),
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnv("DB_PORT", "3306"),
		Username:        getEnv("DB_USERNAME", "root"),
		Password:        getEnv("DB_PASSWORD", ""),
		Database:        getEnv("DB_DATABASE", "neonex.db"),
		Charset:         getEnv("DB_CHARSET", "utf8mb4"),
		ParseTime:       getEnv("DB_PARSE_TIME", "True"),
		Loc:             getEnv("DB_LOC", "Local"),
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		LogLevel:        logger.Info,
	}
}

// InitDatabase initializes database connection
func InitDatabase(config *DatabaseConfig) (*DatabaseManager, error) {
	var dialector gorm.Dialector

	switch config.Driver {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%s&loc=%s",
			config.Username,
			config.Password,
			config.Host,
			config.Port,
			config.Database,
			config.Charset,
			config.ParseTime,
			config.Loc,
		)
		dialector = mysql.Open(dsn)

	case "postgres", "postgresql":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Bangkok",
			config.Host,
			config.Username,
			config.Password,
			config.Database,
			config.Port,
		)
		dialector = postgres.Open(dsn)

	case "sqlite":
		dialector = sqlite.Open(config.Database)

	case "turso":
		// Turso uses libsql URL format: libsql://[name]-[org].turso.io?authToken=xxx
		dialector = sqlite.Open(config.Database)

	default:
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}

	// Configure GORM logger
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  config.LogLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB to set connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	manager := &DatabaseManager{
		db:     db,
		config: config,
	}

	DB = manager

	fmt.Printf("✅ Database connected: %s\n", config.Driver)
	return manager, nil
}

// GetDB returns the database instance
func (dm *DatabaseManager) GetDB() *gorm.DB {
	return dm.db
}

// Close closes the database connection
func (dm *DatabaseManager) Close() error {
	sqlDB, err := dm.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping checks database connectivity
func (dm *DatabaseManager) Ping() error {
	sqlDB, err := dm.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

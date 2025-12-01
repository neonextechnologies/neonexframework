package config

import (
	"neonexcore/internal/config"

	"gorm.io/gorm"
)

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return config.DB.GetDB()
}

// DatabaseConfig wraps the core database config
type DatabaseConfig = config.DatabaseConfig

// DB is the global database instance
var DB = config.DB

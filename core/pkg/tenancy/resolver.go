package tenancy

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// Strategy represents tenant isolation strategy
type Strategy string

const (
	// StrategySharedDatabase - All tenants share the same database with tenant_id column
	StrategySharedDatabase Strategy = "shared_database"

	// StrategySeparateDatabase - Each tenant has its own database
	StrategySeparateDatabase Strategy = "separate_database"

	// StrategySharedSchema - All tenants share database but have separate schemas
	StrategySharedSchema Strategy = "shared_schema"
)

// Resolver resolves database connection based on tenant and strategy
type Resolver struct {
	strategy       Strategy
	sharedDB       *gorm.DB
	tenantDBs      map[string]*gorm.DB
	manager        *TenantManager
	schemaTemplate string
}

// ResolverConfig holds resolver configuration
type ResolverConfig struct {
	Strategy       Strategy
	SharedDB       *gorm.DB
	SchemaTemplate string // For separate schema strategy
}

// NewResolver creates a new tenant resolver
func NewResolver(config ResolverConfig, manager *TenantManager) *Resolver {
	return &Resolver{
		strategy:       config.Strategy,
		sharedDB:       config.SharedDB,
		tenantDBs:      make(map[string]*gorm.DB),
		manager:        manager,
		schemaTemplate: config.SchemaTemplate,
	}
}

// Resolve resolves database connection for tenant
func (r *Resolver) Resolve(ctx context.Context, tenant *Tenant) (*gorm.DB, error) {
	switch r.strategy {
	case StrategySharedDatabase:
		return r.resolveSharedDatabase(ctx, tenant)
	case StrategySeparateDatabase:
		return r.resolveSeparateDatabase(ctx, tenant)
	case StrategySharedSchema:
		return r.resolveSharedSchema(ctx, tenant)
	default:
		return nil, fmt.Errorf("unknown strategy: %s", r.strategy)
	}
}

// resolveSharedDatabase returns DB with tenant_id filter
func (r *Resolver) resolveSharedDatabase(ctx context.Context, tenant *Tenant) (*gorm.DB, error) {
	if r.sharedDB == nil {
		return nil, fmt.Errorf("shared database not configured")
	}

	// Add tenant_id to WHERE clause for all queries
	return r.sharedDB.Where("tenant_id = ?", tenant.ID), nil
}

// resolveSeparateDatabase returns tenant-specific database
func (r *Resolver) resolveSeparateDatabase(ctx context.Context, tenant *Tenant) (*gorm.DB, error) {
	// Check if already connected
	if db, exists := r.tenantDBs[tenant.ID]; exists {
		return db, nil
	}

	// Connect to tenant database
	if tenant.DatabaseURL == "" {
		return nil, fmt.Errorf("tenant database URL not configured")
	}

	db, err := gorm.Open(nil, &gorm.Config{}) // TODO: Parse DatabaseURL
	if err != nil {
		return nil, fmt.Errorf("failed to connect to tenant database: %w", err)
	}

	// Cache connection
	r.tenantDBs[tenant.ID] = db

	return db, nil
}

// resolveSharedSchema returns DB with schema prefix
func (r *Resolver) resolveSharedSchema(ctx context.Context, tenant *Tenant) (*gorm.DB, error) {
	if r.sharedDB == nil {
		return nil, fmt.Errorf("shared database not configured")
	}

	schemaName := fmt.Sprintf("tenant_%s", tenant.ID)

	// Set search_path for PostgreSQL or USE for MySQL
	return r.sharedDB.Exec(fmt.Sprintf("SET search_path TO %s", schemaName)), nil
}

// Close closes all tenant database connections
func (r *Resolver) Close() error {
	for _, db := range r.tenantDBs {
		sqlDB, err := db.DB()
		if err != nil {
			continue
		}
		sqlDB.Close()
	}
	return nil
}

// ScopedDB adds tenant scope to database queries
type ScopedDB struct {
	db       *gorm.DB
	tenantID string
}

// NewScopedDB creates a new scoped database
func NewScopedDB(db *gorm.DB, tenantID string) *ScopedDB {
	return &ScopedDB{
		db:       db,
		tenantID: tenantID,
	}
}

// DB returns database with tenant scope
func (s *ScopedDB) DB() *gorm.DB {
	return s.db.Where("tenant_id = ?", s.tenantID)
}

// Create creates a record with tenant_id
func (s *ScopedDB) Create(value interface{}) *gorm.DB {
	// TODO: Set tenant_id field
	return s.DB().Create(value)
}

// Model specifies the model with tenant scope
func (s *ScopedDB) Model(value interface{}) *gorm.DB {
	return s.DB().Model(value)
}

// Table specifies the table with tenant scope
func (s *ScopedDB) Table(name string) *gorm.DB {
	return s.DB().Table(name)
}

// TenantModel is a base model with tenant_id
type TenantModel struct {
	TenantID string `json:"tenant_id" gorm:"index:idx_tenant_id;not null"`
}

// BeforeCreate hook to set tenant_id
func (tm *TenantModel) BeforeCreate(tx *gorm.DB) error {
	if tm.TenantID == "" {
		// Try to get tenant from context
		if tenant, err := GetTenant(tx.Statement.Context); err == nil {
			tm.TenantID = tenant.ID
		}
	}
	return nil
}

// TenantScope adds tenant_id scope to queries
func TenantScope(tenantID string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ?", tenantID)
	}
}

// TenantScopeFromContext adds tenant_id scope from context
func TenantScopeFromContext(ctx context.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		tenant, err := GetTenant(ctx)
		if err != nil {
			return db
		}
		return db.Where("tenant_id = ?", tenant.ID)
	}
}

// MigrateTenantTables creates tables with tenant isolation
func MigrateTenantTables(db *gorm.DB, strategy Strategy, models ...interface{}) error {
	switch strategy {
	case StrategySharedDatabase:
		// Add tenant_id column to all tables
		for _, model := range models {
			if err := db.AutoMigrate(model); err != nil {
				return err
			}
		}
		return nil

	case StrategySeparateDatabase:
		// Each tenant will have their own database
		// Migration should be run per tenant
		return nil

	case StrategySharedSchema:
		// Create schema per tenant
		return nil

	default:
		return fmt.Errorf("unknown strategy: %s", strategy)
	}
}

// CreateTenantSchema creates a new schema for a tenant
func CreateTenantSchema(db *gorm.DB, tenantID string) error {
	schemaName := fmt.Sprintf("tenant_%s", tenantID)
	
	// PostgreSQL
	if strings.Contains(db.Dialector.Name(), "postgres") {
		return db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)).Error
	}
	
	// MySQL doesn't have schemas, use separate database
	if strings.Contains(db.Dialector.Name(), "mysql") {
		return db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", schemaName)).Error
	}
	
	return nil
}

// DropTenantSchema drops a tenant schema
func DropTenantSchema(db *gorm.DB, tenantID string) error {
	schemaName := fmt.Sprintf("tenant_%s", tenantID)
	
	// PostgreSQL
	if strings.Contains(db.Dialector.Name(), "postgres") {
		return db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName)).Error
	}
	
	// MySQL
	if strings.Contains(db.Dialector.Name(), "mysql") {
		return db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", schemaName)).Error
	}
	
	return nil
}

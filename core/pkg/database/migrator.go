package database

import (
	"fmt"

	"gorm.io/gorm"
)

// Migrator handles database migrations
type Migrator struct {
	db     *gorm.DB
	models []interface{}
}

// NewMigrator creates a new migrator
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{
		db:     db,
		models: make([]interface{}, 0),
	}
}

// RegisterModels registers models for migration
func (m *Migrator) RegisterModels(models ...interface{}) {
	m.models = append(m.models, models...)
}

// AutoMigrate runs auto migration for all registered models
func (m *Migrator) AutoMigrate() error {
	if len(m.models) == 0 {
		fmt.Println("⚠️  No models registered for migration")
		return nil
	}

	fmt.Printf("🔄 Running auto-migration for %d models...\n", len(m.models))

	for _, model := range m.models {
		if err := m.db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate model: %w", err)
		}
	}

	fmt.Println("✅ Database migration completed")
	return nil
}

// DropTables drops all registered model tables
func (m *Migrator) DropTables() error {
	if len(m.models) == 0 {
		return nil
	}

	fmt.Println("⚠️  Dropping tables...")

	for _, model := range m.models {
		if err := m.db.Migrator().DropTable(model); err != nil {
			return fmt.Errorf("failed to drop table: %w", err)
		}
	}

	fmt.Println("✅ Tables dropped")
	return nil
}

// Reset drops and recreates all tables
func (m *Migrator) Reset() error {
	if err := m.DropTables(); err != nil {
		return err
	}
	return m.AutoMigrate()
}

// HasTable checks if a table exists
func (m *Migrator) HasTable(model interface{}) bool {
	return m.db.Migrator().HasTable(model)
}

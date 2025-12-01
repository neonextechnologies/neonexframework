package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// Seeder interface for database seeding
type Seeder interface {
	Name() string
	Run(ctx context.Context) error
}

// SeederManager manages database seeders
type SeederManager struct {
	db      *gorm.DB
	seeders []Seeder
}

// NewSeederManager creates a new seeder manager
func NewSeederManager(db *gorm.DB) *SeederManager {
	return &SeederManager{
		db:      db,
		seeders: make([]Seeder, 0),
	}
}

// Register registers a seeder
func (sm *SeederManager) Register(seeder Seeder) {
	sm.seeders = append(sm.seeders, seeder)
}

// Run runs all registered seeders
func (sm *SeederManager) Run(ctx context.Context) error {
	if len(sm.seeders) == 0 {
		fmt.Println("⚠️  No seeders registered")
		return nil
	}

	fmt.Printf("🌱 Running %d seeders...\n", len(sm.seeders))

	for _, seeder := range sm.seeders {
		fmt.Printf("Running %s...\n", seeder.Name())
		if err := seeder.Run(ctx); err != nil {
			return fmt.Errorf("seeder %s failed: %w", seeder.Name(), err)
		}
	}

	fmt.Println("✅ Database seeding completed")
	return nil
}

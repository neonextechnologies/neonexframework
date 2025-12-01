package product

import (
	"context"

	"gorm.io/gorm"
)

type ProductSeeder struct{}

func (s *ProductSeeder) Seed(ctx context.Context, db *gorm.DB) error {
	var count int64
	db.Model(&Product{}).Count(&count)
	if count > 0 {
		return nil // Already seeded
	}

	samples := []Product{
		{Name: "Sample 1", Description: "First sample product", IsActive: true},
		{Name: "Sample 2", Description: "Second sample product", IsActive: true},
		{Name: "Sample 3", Description: "Third sample product", IsActive: false},
	}

	return db.Create(&samples).Error
}

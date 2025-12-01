package product

import (
	"neonexcore/pkg/database"

	"gorm.io/gorm"
)

type Repository struct {
	*database.Repository[Product]
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		Repository: database.NewRepository[Product](db),
	}
}

// Add custom repository methods here
func (r *Repository) FindByName(name string) (*Product, error) {
	var entity Product
	err := r.DB.Where("name = ?", name).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *Repository) FindActive() ([]Product, error) {
	var entities []Product
	err := r.DB.Where("is_active = ?", true).Find(&entities).Error
	return entities, err
}

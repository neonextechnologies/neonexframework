package database

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// Repository interface defines common CRUD operations
type Repository[T any] interface {
	Create(ctx context.Context, entity *T) error
	CreateBatch(ctx context.Context, entities []*T) error
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id interface{}) error
	FindByID(ctx context.Context, id interface{}) (*T, error)
	FindAll(ctx context.Context) ([]*T, error)
	FindByCondition(ctx context.Context, condition interface{}, args ...interface{}) ([]*T, error)
	FindOne(ctx context.Context, condition interface{}, args ...interface{}) (*T, error)
	Count(ctx context.Context, condition interface{}, args ...interface{}) (int64, error)
	Paginate(ctx context.Context, page, pageSize int) ([]*T, int64, error)
}

// BaseRepository implements the Repository interface
type BaseRepository[T any] struct {
	db *gorm.DB
}

// NewBaseRepository creates a new base repository
func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

// GetDB returns the database instance
func (r *BaseRepository[T]) GetDB() *gorm.DB {
	return r.db
}

// WithTx returns a repository with a transaction
func (r *BaseRepository[T]) WithTx(tx *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: tx}
}

// Create creates a new entity
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// CreateBatch creates multiple entities
func (r *BaseRepository[T]) CreateBatch(ctx context.Context, entities []*T) error {
	return r.db.WithContext(ctx).CreateInBatches(entities, 100).Error
}

// Update updates an entity
func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// Delete deletes an entity by ID
func (r *BaseRepository[T]) Delete(ctx context.Context, id interface{}) error {
	var entity T
	return r.db.WithContext(ctx).Delete(&entity, id).Error
}

// FindByID finds an entity by ID
func (r *BaseRepository[T]) FindByID(ctx context.Context, id interface{}) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).First(&entity, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// FindAll finds all entities
func (r *BaseRepository[T]) FindAll(ctx context.Context) ([]*T, error) {
	var entities []*T
	err := r.db.WithContext(ctx).Find(&entities).Error
	return entities, err
}

// FindByCondition finds entities by condition
func (r *BaseRepository[T]) FindByCondition(ctx context.Context, condition interface{}, args ...interface{}) ([]*T, error) {
	var entities []*T
	err := r.db.WithContext(ctx).Where(condition, args...).Find(&entities).Error
	return entities, err
}

// FindOne finds one entity by condition
func (r *BaseRepository[T]) FindOne(ctx context.Context, condition interface{}, args ...interface{}) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).Where(condition, args...).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// Count counts entities by condition
func (r *BaseRepository[T]) Count(ctx context.Context, condition interface{}, args ...interface{}) (int64, error) {
	var count int64
	var entity T
	err := r.db.WithContext(ctx).Model(&entity).Where(condition, args...).Count(&count).Error
	return count, err
}

// Paginate returns paginated results
func (r *BaseRepository[T]) Paginate(ctx context.Context, page, pageSize int) ([]*T, int64, error) {
	var entities []*T
	var total int64

	offset := (page - 1) * pageSize

	var entity T
	if err := r.db.WithContext(ctx).Model(&entity).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Find(&entities).Error
	return entities, total, err
}

// Query returns a query builder
func (r *BaseRepository[T]) Query(ctx context.Context) *gorm.DB {
	var entity T
	return r.db.WithContext(ctx).Model(&entity)
}

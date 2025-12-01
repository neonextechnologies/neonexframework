package module

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"
	"neonexcore/pkg/database"
)

// ModuleRepository handles database operations for modules
type ModuleRepository struct {
	*database.BaseRepository[Module]
	db *gorm.DB
}

// NewModuleRepository creates a new module repository
func NewModuleRepository(db *gorm.DB) *ModuleRepository {
	return &ModuleRepository{
		BaseRepository: database.NewBaseRepository[Module](db),
		db:             db,
	}
}

// FindByName finds a module by name
func (r *ModuleRepository) FindByName(ctx context.Context, name string) (*Module, error) {
	var module Module
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&module).Error
	if err != nil {
		return nil, err
	}
	return &module, nil
}

// FindByStatus finds modules by status
func (r *ModuleRepository) FindByStatus(ctx context.Context, status ModuleStatus) ([]Module, error) {
	var modules []Module
	err := r.db.WithContext(ctx).Where("status = ?", status).Order("priority ASC, name ASC").Find(&modules).Error
	return modules, err
}

// GetActiveModules gets all active modules ordered by priority
func (r *ModuleRepository) GetActiveModules(ctx context.Context) ([]Module, error) {
	return r.FindByStatus(ctx, ModuleStatusActive)
}

// UpdateStatus updates module status
func (r *ModuleRepository) UpdateStatus(ctx context.Context, moduleID uint, status ModuleStatus) error {
	return r.db.WithContext(ctx).Model(&Module{}).Where("id = ?", moduleID).Update("status", status).Error
}

// Search searches modules by name or description
func (r *ModuleRepository) Search(ctx context.Context, query string) ([]Module, error) {
	var modules []Module
	searchPattern := "%" + query + "%"
	err := r.db.WithContext(ctx).
		Where("name LIKE ? OR display_name LIKE ? OR description LIKE ?", searchPattern, searchPattern, searchPattern).
		Order("priority ASC, name ASC").
		Find(&modules).Error
	return modules, err
}

// List lists modules with filters and pagination
func (r *ModuleRepository) List(ctx context.Context, filter ModuleListFilter) ([]Module, int64, error) {
	query := r.db.WithContext(ctx).Model(&Module{})

	// Apply filters
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("name LIKE ? OR display_name LIKE ? OR description LIKE ?", 
			searchPattern, searchPattern, searchPattern)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply ordering
	orderBy := "priority"
	if filter.OrderBy != "" {
		orderBy = filter.OrderBy
	}
	orderDir := "ASC"
	if filter.OrderDir != "" {
		orderDir = filter.OrderDir
	}
	query = query.Order(orderBy + " " + orderDir)

	// Apply pagination
	page := 1
	if filter.Page > 0 {
		page = filter.Page
	}
	limit := 10
	if filter.Limit > 0 {
		limit = filter.Limit
	}
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	var modules []Module
	err := query.Find(&modules).Error
	return modules, total, err
}

// GetDependencies gets module dependencies
func (r *ModuleRepository) GetDependencies(ctx context.Context, moduleID uint) ([]ModuleDependency, error) {
	var deps []ModuleDependency
	err := r.db.WithContext(ctx).Where("module_id = ?", moduleID).Find(&deps).Error
	return deps, err
}

// CreateDependency creates a module dependency
func (r *ModuleRepository) CreateDependency(ctx context.Context, dep *ModuleDependency) error {
	return r.db.WithContext(ctx).Create(dep).Error
}

// DeleteDependencies deletes all dependencies for a module
func (r *ModuleRepository) DeleteDependencies(ctx context.Context, moduleID uint) error {
	return r.db.WithContext(ctx).Where("module_id = ?", moduleID).Delete(&ModuleDependency{}).Error
}

// GetMigrations gets module migration history
func (r *ModuleRepository) GetMigrations(ctx context.Context, moduleID uint) ([]ModuleMigration, error) {
	var migrations []ModuleMigration
	err := r.db.WithContext(ctx).Where("module_id = ?", moduleID).Order("batch ASC, created_at ASC").Find(&migrations).Error
	return migrations, err
}

// CreateMigration creates a migration record
func (r *ModuleRepository) CreateMigration(ctx context.Context, migration *ModuleMigration) error {
	return r.db.WithContext(ctx).Create(migration).Error
}

// GetLastBatch gets the last migration batch number
func (r *ModuleRepository) GetLastBatch(ctx context.Context) (int, error) {
	var batch int
	err := r.db.WithContext(ctx).Model(&ModuleMigration{}).Select("COALESCE(MAX(batch), 0)").Scan(&batch).Error
	return batch, err
}

// GetModuleWithDependencies gets module with its dependencies
func (r *ModuleRepository) GetModuleWithDependencies(ctx context.Context, moduleID uint) (*Module, []ModuleDependency, error) {
	module, err := r.FindByID(ctx, moduleID)
	if err != nil {
		return nil, nil, err
	}

	deps, err := r.GetDependencies(ctx, moduleID)
	if err != nil {
		return nil, nil, err
	}

	return module, deps, nil
}

// ParseConfig parses module config JSON string
func (r *ModuleRepository) ParseConfig(configJSON string) (map[string]interface{}, error) {
	var config map[string]interface{}
	if configJSON == "" {
		return config, nil
	}
	err := json.Unmarshal([]byte(configJSON), &config)
	return config, err
}

// SaveConfig saves module config as JSON string
func (r *ModuleRepository) SaveConfig(ctx context.Context, moduleID uint, config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&Module{}).Where("id = ?", moduleID).Update("config", string(configJSON)).Error
}

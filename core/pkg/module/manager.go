package module

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"
	"neonexcore/pkg/database"
	"neonexcore/pkg/errors"
	"neonexcore/pkg/events"
	"neonexcore/pkg/logger"
	"neonexcore/pkg/validation"
)

// Module lifecycle events
const (
	EventModuleInstalling   = "module.installing"
	EventModuleInstalled    = "module.installed"
	EventModuleUninstalling = "module.uninstalling"
	EventModuleUninstalled  = "module.uninstalled"
	EventModuleActivating   = "module.activating"
	EventModuleActivated    = "module.activated"
	EventModuleDeactivating = "module.deactivating"
	EventModuleDeactivated  = "module.deactivated"
	EventModuleUpdating     = "module.updating"
	EventModuleUpdated      = "module.updated"
)

// ModuleManager manages module lifecycle operations
type ModuleManager struct {
	repo       *ModuleRepository
	db         *gorm.DB
	txManager  *database.TransactionManager
	events     *events.EventDispatcher
	logger     logger.Logger
	validator  *validation.Validator
	modulesDir string
}

// NewModuleManager creates a new module manager
func NewModuleManager(
	repo *ModuleRepository,
	db *gorm.DB,
	txManager *database.TransactionManager,
	events *events.EventDispatcher,
	logger logger.Logger,
	validator *validation.Validator,
	modulesDir string,
) *ModuleManager {
	return &ModuleManager{
		repo:       repo,
		db:         db,
		txManager:  txManager,
		events:     events,
		logger:     logger,
		validator:  validator,
		modulesDir: modulesDir,
	}
}

// Install installs a module
func (m *ModuleManager) Install(ctx context.Context, modulePath string) (*Module, error) {
	m.logger.Info("Installing module", logger.Fields{"path": modulePath})

	// Dispatch installing event
	m.events.Dispatch(ctx, EventModuleInstalling, map[string]interface{}{
		"path": modulePath,
	})

	// Load and validate module metadata
	metadata, err := m.LoadMetadata(modulePath)
	if err != nil {
		return nil, errors.NewBadRequest(fmt.Sprintf("Invalid module metadata: %v", err))
	}

	// Validate metadata
	if err := m.validator.Validate(metadata); err != nil {
		return nil, errors.NewValidation("Invalid module metadata", map[string]interface{}{
			"errors": err,
		})
	}

	// Check if module already exists
	existing, _ := m.repo.FindByName(ctx, metadata.Name)
	if existing != nil {
		return nil, errors.NewConflict("Module already installed")
	}

	// Check dependencies
	if err := m.CheckDependencies(ctx, metadata.Dependencies); err != nil {
		return nil, err
	}

	// Create module in transaction
	var module *Module
	err = m.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Create module record
		configJSON, _ := json.Marshal(metadata.Config)
		module = &Module{
			Name:        metadata.Name,
			DisplayName: metadata.DisplayName,
			Description: metadata.Description,
			Version:     metadata.Version,
			Author:      metadata.Author,
			Homepage:    metadata.Homepage,
			Status:      ModuleStatusInstalled,
			Priority:    metadata.Priority,
			Path:        modulePath,
			Config:      string(configJSON),
			InstalledAt: time.Now(),
		}

		if err := m.repo.Create(txCtx, module); err != nil {
			return errors.NewInternal(fmt.Sprintf("Failed to create module record: %v", err))
		}

		// Create dependencies
		for _, dep := range metadata.Dependencies {
			dependency := &ModuleDependency{
				ModuleID:        module.ID,
				DependsOnModule: dep.Name,
				Version:         dep.Version,
				Required:        dep.Required,
			}
			if err := m.repo.CreateDependency(txCtx, dependency); err != nil {
				return errors.NewInternal(fmt.Sprintf("Failed to create dependency: %v", err))
			}
		}

		// Run migrations if exists
		if metadata.Migrations {
			if err := m.RunMigrations(txCtx, module); err != nil {
				return errors.NewInternal(fmt.Sprintf("Failed to run migrations: %v", err))
			}
		}

		// Run seeders if exists
		if metadata.Seeders {
			if err := m.RunSeeders(txCtx, module); err != nil {
				m.logger.Warn("Failed to run seeders", logger.Fields{
					"module": module.Name,
					"error":  err.Error(),
				})
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	m.logger.Info("Module installed successfully", logger.Fields{
		"module":  module.Name,
		"version": module.Version,
	})

	// Dispatch installed event
	m.events.Dispatch(ctx, EventModuleInstalled, map[string]interface{}{
		"module_id": module.ID,
		"module":    module.Name,
		"version":   module.Version,
	})

	return module, nil
}

// Uninstall uninstalls a module
func (m *ModuleManager) Uninstall(ctx context.Context, moduleName string, force bool) error {
	m.logger.Info("Uninstalling module", logger.Fields{"module": moduleName, "force": force})

	module, err := m.repo.FindByName(ctx, moduleName)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewNotFound("Module not found")
		}
		return errors.NewInternal(fmt.Sprintf("Failed to find module: %v", err))
	}

	// Dispatch uninstalling event
	m.events.Dispatch(ctx, EventModuleUninstalling, map[string]interface{}{
		"module_id": module.ID,
		"module":    module.Name,
	})

	// Check if other modules depend on this
	if !force {
		if err := m.CheckDependents(ctx, moduleName); err != nil {
			return err
		}
	}

	// Deactivate if active
	if module.Status == ModuleStatusActive {
		if err := m.Deactivate(ctx, moduleName); err != nil {
			return err
		}
	}

	// Uninstall in transaction
	err = m.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Rollback migrations
		if err := m.RollbackMigrations(txCtx, module); err != nil {
			m.logger.Warn("Failed to rollback migrations", logger.Fields{
				"module": module.Name,
				"error":  err.Error(),
			})
		}

		// Delete dependencies
		if err := m.repo.DeleteDependencies(txCtx, module.ID); err != nil {
			return errors.NewInternal(fmt.Sprintf("Failed to delete dependencies: %v", err))
		}

		// Delete module
		if err := m.repo.Delete(txCtx, module.ID); err != nil {
			return errors.NewInternal(fmt.Sprintf("Failed to delete module: %v", err))
		}

		return nil
	})

	if err != nil {
		return err
	}

	m.logger.Info("Module uninstalled successfully", logger.Fields{"module": moduleName})

	// Dispatch uninstalled event
	m.events.Dispatch(ctx, EventModuleUninstalled, map[string]interface{}{
		"module": moduleName,
	})

	return nil
}

// Activate activates a module
func (m *ModuleManager) Activate(ctx context.Context, moduleName string) error {
	m.logger.Info("Activating module", logger.Fields{"module": moduleName})

	module, err := m.repo.FindByName(ctx, moduleName)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewNotFound("Module not found")
		}
		return errors.NewInternal(fmt.Sprintf("Failed to find module: %v", err))
	}

	if module.Status == ModuleStatusActive {
		return errors.NewBadRequest("Module is already active")
	}

	// Dispatch activating event
	m.events.Dispatch(ctx, EventModuleActivating, map[string]interface{}{
		"module_id": module.ID,
		"module":    module.Name,
	})

	// Check dependencies are active
	deps, err := m.repo.GetDependencies(ctx, module.ID)
	if err != nil {
		return errors.NewInternal(fmt.Sprintf("Failed to get dependencies: %v", err))
	}

	for _, dep := range deps {
		if dep.Required {
			depModule, err := m.repo.FindByName(ctx, dep.DependsOnModule)
			if err != nil || depModule.Status != ModuleStatusActive {
				return errors.NewBadRequest(fmt.Sprintf("Required dependency '%s' is not active", dep.DependsOnModule))
			}
		}
	}

	// Update status
	if err := m.repo.UpdateStatus(ctx, module.ID, ModuleStatusActive); err != nil {
		return errors.NewInternal(fmt.Sprintf("Failed to activate module: %v", err))
	}

	m.logger.Info("Module activated successfully", logger.Fields{"module": moduleName})

	// Dispatch activated event
	m.events.Dispatch(ctx, EventModuleActivated, map[string]interface{}{
		"module_id": module.ID,
		"module":    module.Name,
	})

	return nil
}

// Deactivate deactivates a module
func (m *ModuleManager) Deactivate(ctx context.Context, moduleName string) error {
	m.logger.Info("Deactivating module", logger.Fields{"module": moduleName})

	module, err := m.repo.FindByName(ctx, moduleName)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewNotFound("Module not found")
		}
		return errors.NewInternal(fmt.Sprintf("Failed to find module: %v", err))
	}

	if module.Status != ModuleStatusActive {
		return errors.NewBadRequest("Module is not active")
	}

	// Dispatch deactivating event
	m.events.Dispatch(ctx, EventModuleDeactivating, map[string]interface{}{
		"module_id": module.ID,
		"module":    module.Name,
	})

	// Update status
	if err := m.repo.UpdateStatus(ctx, module.ID, ModuleStatusInactive); err != nil {
		return errors.NewInternal(fmt.Sprintf("Failed to deactivate module: %v", err))
	}

	m.logger.Info("Module deactivated successfully", logger.Fields{"module": moduleName})

	// Dispatch deactivated event
	m.events.Dispatch(ctx, EventModuleDeactivated, map[string]interface{}{
		"module_id": module.ID,
		"module":    module.Name,
	})

	return nil
}

// Update updates a module to a new version
func (m *ModuleManager) Update(ctx context.Context, moduleName string, newPath string) error {
	m.logger.Info("Updating module", logger.Fields{"module": moduleName, "path": newPath})

	module, err := m.repo.FindByName(ctx, moduleName)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewNotFound("Module not found")
		}
		return errors.NewInternal(fmt.Sprintf("Failed to find module: %v", err))
	}

	// Load new metadata
	metadata, err := m.LoadMetadata(newPath)
	if err != nil {
		return errors.NewBadRequest(fmt.Sprintf("Invalid module metadata: %v", err))
	}

	if metadata.Name != moduleName {
		return errors.NewBadRequest("Module name mismatch")
	}

	// Dispatch updating event
	m.events.Dispatch(ctx, EventModuleUpdating, map[string]interface{}{
		"module_id":   module.ID,
		"module":      module.Name,
		"old_version": module.Version,
		"new_version": metadata.Version,
	})

	// Update in transaction
	err = m.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Update module record
		configJSON, _ := json.Marshal(metadata.Config)
		updates := map[string]interface{}{
			"display_name": metadata.DisplayName,
			"description":  metadata.Description,
			"version":      metadata.Version,
			"author":       metadata.Author,
			"homepage":     metadata.Homepage,
			"priority":     metadata.Priority,
			"path":         newPath,
			"config":       string(configJSON),
		}

		if err := m.repo.Update(txCtx, module.ID, updates); err != nil {
			return errors.NewInternal(fmt.Sprintf("Failed to update module: %v", err))
		}

		// Update dependencies
		if err := m.repo.DeleteDependencies(txCtx, module.ID); err != nil {
			return errors.NewInternal(fmt.Sprintf("Failed to delete old dependencies: %v", err))
		}

		for _, dep := range metadata.Dependencies {
			dependency := &ModuleDependency{
				ModuleID:        module.ID,
				DependsOnModule: dep.Name,
				Version:         dep.Version,
				Required:        dep.Required,
			}
			if err := m.repo.CreateDependency(txCtx, dependency); err != nil {
				return errors.NewInternal(fmt.Sprintf("Failed to create dependency: %v", err))
			}
		}

		// Run new migrations
		if metadata.Migrations {
			if err := m.RunMigrations(txCtx, module); err != nil {
				return errors.NewInternal(fmt.Sprintf("Failed to run migrations: %v", err))
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	m.logger.Info("Module updated successfully", logger.Fields{
		"module":      moduleName,
		"old_version": module.Version,
		"new_version": metadata.Version,
	})

	// Dispatch updated event
	m.events.Dispatch(ctx, EventModuleUpdated, map[string]interface{}{
		"module_id":   module.ID,
		"module":      module.Name,
		"old_version": module.Version,
		"new_version": metadata.Version,
	})

	return nil
}

// LoadMetadata loads module.json from module path
func (m *ModuleManager) LoadMetadata(modulePath string) (*ModuleMetadata, error) {
	metadataPath := filepath.Join(modulePath, "module.json")
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read module.json: %w", err)
	}

	var metadata ModuleMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse module.json: %w", err)
	}

	return &metadata, nil
}

// CheckDependencies checks if all required dependencies are installed and active
func (m *ModuleManager) CheckDependencies(ctx context.Context, deps []ModuleDependencyInfo) error {
	for _, dep := range deps {
		if !dep.Required {
			continue
		}

		module, err := m.repo.FindByName(ctx, dep.Name)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.NewBadRequest(fmt.Sprintf("Required dependency '%s' is not installed", dep.Name))
			}
			return errors.NewInternal(fmt.Sprintf("Failed to check dependency: %v", err))
		}

		if module.Status != ModuleStatusActive {
			return errors.NewBadRequest(fmt.Sprintf("Required dependency '%s' is not active", dep.Name))
		}

		// TODO: Check version compatibility using semantic versioning
	}

	return nil
}

// CheckDependents checks if other modules depend on this module
func (m *ModuleManager) CheckDependents(ctx context.Context, moduleName string) error {
	var count int64
	err := m.db.Model(&ModuleDependency{}).
		Where("depends_on_module = ? AND required = ?", moduleName, true).
		Count(&count).Error

	if err != nil {
		return errors.NewInternal(fmt.Sprintf("Failed to check dependents: %v", err))
	}

	if count > 0 {
		return errors.NewBadRequest(fmt.Sprintf("Cannot uninstall: %d module(s) depend on this module", count))
	}

	return nil
}

// RunMigrations runs module migrations (placeholder)
func (m *ModuleManager) RunMigrations(ctx context.Context, module *Module) error {
	// TODO: Implement migration runner
	m.logger.Info("Running migrations", logger.Fields{"module": module.Name})
	return nil
}

// RollbackMigrations rollbacks module migrations (placeholder)
func (m *ModuleManager) RollbackMigrations(ctx context.Context, module *Module) error {
	// TODO: Implement migration rollback
	m.logger.Info("Rolling back migrations", logger.Fields{"module": module.Name})
	return nil
}

// RunSeeders runs module seeders (placeholder)
func (m *ModuleManager) RunSeeders(ctx context.Context, module *Module) error {
	// TODO: Implement seeder runner
	m.logger.Info("Running seeders", logger.Fields{"module": module.Name})
	return nil
}

// GetModule gets module by name
func (m *ModuleManager) GetModule(ctx context.Context, moduleName string) (*ModuleInfo, error) {
	module, err := m.repo.FindByName(ctx, moduleName)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFound("Module not found")
		}
		return nil, errors.NewInternal(fmt.Sprintf("Failed to get module: %v", err))
	}

	deps, err := m.repo.GetDependencies(ctx, module.ID)
	if err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("Failed to get dependencies: %v", err))
	}

	depInfos := make([]ModuleDependencyInfo, len(deps))
	for i, dep := range deps {
		depInfos[i] = ModuleDependencyInfo{
			Name:     dep.DependsOnModule,
			Version:  dep.Version,
			Required: dep.Required,
		}
	}

	return &ModuleInfo{
		ID:           module.ID,
		Name:         module.Name,
		DisplayName:  module.DisplayName,
		Description:  module.Description,
		Version:      module.Version,
		Author:       module.Author,
		Homepage:     module.Homepage,
		Status:       module.Status,
		Priority:     module.Priority,
		Path:         module.Path,
		Dependencies: depInfos,
		InstalledAt:  module.InstalledAt,
		UpdatedAt:    module.UpdatedAt,
	}, nil
}

// ListModules lists all modules with filters
func (m *ModuleManager) ListModules(ctx context.Context, filter ModuleListFilter) ([]ModuleInfo, int64, error) {
	modules, total, err := m.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, errors.NewInternal(fmt.Sprintf("Failed to list modules: %v", err))
	}

	infos := make([]ModuleInfo, len(modules))
	for i, module := range modules {
		deps, _ := m.repo.GetDependencies(ctx, module.ID)
		depInfos := make([]ModuleDependencyInfo, len(deps))
		for j, dep := range deps {
			depInfos[j] = ModuleDependencyInfo{
				Name:     dep.DependsOnModule,
				Version:  dep.Version,
				Required: dep.Required,
			}
		}

		infos[i] = ModuleInfo{
			ID:           module.ID,
			Name:         module.Name,
			DisplayName:  module.DisplayName,
			Description:  module.Description,
			Version:      module.Version,
			Author:       module.Author,
			Homepage:     module.Homepage,
			Status:       module.Status,
			Priority:     module.Priority,
			Path:         module.Path,
			Dependencies: depInfos,
			InstalledAt:  module.InstalledAt,
			UpdatedAt:    module.UpdatedAt,
		}
	}

	return infos, total, nil
}

package tenancy

import (
	"context"
	"sync"

	"gorm.io/gorm"
)

// GormTenantStore implements TenantStore using GORM
type GormTenantStore struct {
	db *gorm.DB
	mu sync.RWMutex
}

// NewGormTenantStore creates a new GORM tenant store
func NewGormTenantStore(db *gorm.DB) *GormTenantStore {
	return &GormTenantStore{
		db: db,
	}
}

// Get retrieves a tenant by ID
func (s *GormTenantStore) Get(ctx context.Context, id string) (*Tenant, error) {
	var tenant Tenant
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&tenant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}
	return &tenant, nil
}

// GetByDomain retrieves a tenant by domain
func (s *GormTenantStore) GetByDomain(ctx context.Context, domain string) (*Tenant, error) {
	var tenant Tenant
	if err := s.db.WithContext(ctx).Where("domain = ?", domain).First(&tenant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}
	return &tenant, nil
}

// Create creates a new tenant
func (s *GormTenantStore) Create(ctx context.Context, tenant *Tenant) error {
	return s.db.WithContext(ctx).Create(tenant).Error
}

// Update updates a tenant
func (s *GormTenantStore) Update(ctx context.Context, tenant *Tenant) error {
	return s.db.WithContext(ctx).Save(tenant).Error
}

// Delete deletes a tenant
func (s *GormTenantStore) Delete(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Where("id = ?", id).Delete(&Tenant{}).Error
}

// List lists tenants with filter
func (s *GormTenantStore) List(ctx context.Context, filter TenantFilter) ([]*Tenant, error) {
	var tenants []*Tenant
	query := s.db.WithContext(ctx)

	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	if filter.Plan != nil {
		query = query.Where("plan = ?", *filter.Plan)
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	if err := query.Find(&tenants).Error; err != nil {
		return nil, err
	}

	return tenants, nil
}

// Migrate runs tenant table migration
func (s *GormTenantStore) Migrate() error {
	return s.db.AutoMigrate(&Tenant{})
}

package tenancy

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Tenant represents a tenant in the system
type Tenant struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	Name        string                 `json:"name" gorm:"index"`
	Domain      string                 `json:"domain" gorm:"uniqueIndex"`
	Status      TenantStatus           `json:"status" gorm:"index"`
	Plan        string                 `json:"plan" gorm:"index"`
	MaxUsers    int                    `json:"max_users"`
	MaxStorage  int64                  `json:"max_storage"` // bytes
	DatabaseURL string                 `json:"database_url,omitempty"`
	Settings    map[string]interface{} `json:"settings" gorm:"serializer:json"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"serializer:json"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// TenantStatus represents tenant status
type TenantStatus string

const (
	TenantActive    TenantStatus = "active"
	TenantSuspended TenantStatus = "suspended"
	TenantExpired   TenantStatus = "expired"
	TenantDeleted   TenantStatus = "deleted"
)

// TenantContext holds tenant information in context
type TenantContext struct {
	Tenant *Tenant
	UserID string
	Permissions []string
}

// ContextKey is the key for tenant context
type contextKey string

const TenantContextKey contextKey = "tenant"

// Common errors
var (
	ErrTenantNotFound     = errors.New("tenant not found")
	ErrTenantSuspended    = errors.New("tenant is suspended")
	ErrTenantExpired      = errors.New("tenant has expired")
	ErrInvalidTenant      = errors.New("invalid tenant")
	ErrTenantExists       = errors.New("tenant already exists")
	ErrMaxUsersReached    = errors.New("maximum users reached")
	ErrMaxStorageReached  = errors.New("maximum storage reached")
	ErrInvalidDomain      = errors.New("invalid domain")
)

// TenantManager manages tenants
type TenantManager struct {
	tenants map[string]*Tenant
	domains map[string]string // domain -> tenant ID
	mu      sync.RWMutex
	store   TenantStore
}

// TenantStore interface for persistence
type TenantStore interface {
	Get(ctx context.Context, id string) (*Tenant, error)
	GetByDomain(ctx context.Context, domain string) (*Tenant, error)
	Create(ctx context.Context, tenant *Tenant) error
	Update(ctx context.Context, tenant *Tenant) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter TenantFilter) ([]*Tenant, error)
}

// TenantFilter for querying tenants
type TenantFilter struct {
	Status *TenantStatus
	Plan   *string
	Limit  int
	Offset int
}

// NewTenantManager creates a new tenant manager
func NewTenantManager(store TenantStore) *TenantManager {
	return &TenantManager{
		tenants: make(map[string]*Tenant),
		domains: make(map[string]string),
		store:   store,
	}
}

// Get retrieves a tenant by ID
func (tm *TenantManager) Get(ctx context.Context, id string) (*Tenant, error) {
	// Check cache first
	tm.mu.RLock()
	if tenant, exists := tm.tenants[id]; exists {
		tm.mu.RUnlock()
		return tenant, nil
	}
	tm.mu.RUnlock()

	// Load from store
	tenant, err := tm.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache it
	tm.mu.Lock()
	tm.tenants[id] = tenant
	tm.domains[tenant.Domain] = tenant.ID
	tm.mu.Unlock()

	return tenant, nil
}

// GetByDomain retrieves a tenant by domain
func (tm *TenantManager) GetByDomain(ctx context.Context, domain string) (*Tenant, error) {
	// Check cache first
	tm.mu.RLock()
	if tenantID, exists := tm.domains[domain]; exists {
		tenant := tm.tenants[tenantID]
		tm.mu.RUnlock()
		return tenant, nil
	}
	tm.mu.RUnlock()

	// Load from store
	tenant, err := tm.store.GetByDomain(ctx, domain)
	if err != nil {
		return nil, err
	}

	// Cache it
	tm.mu.Lock()
	tm.tenants[tenant.ID] = tenant
	tm.domains[tenant.Domain] = tenant.ID
	tm.mu.Unlock()

	return tenant, nil
}

// Create creates a new tenant
func (tm *TenantManager) Create(ctx context.Context, tenant *Tenant) error {
	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()
	tenant.Status = TenantActive

	if err := tm.store.Create(ctx, tenant); err != nil {
		return err
	}

	// Cache it
	tm.mu.Lock()
	tm.tenants[tenant.ID] = tenant
	tm.domains[tenant.Domain] = tenant.ID
	tm.mu.Unlock()

	return nil
}

// Update updates a tenant
func (tm *TenantManager) Update(ctx context.Context, tenant *Tenant) error {
	tenant.UpdatedAt = time.Now()

	if err := tm.store.Update(ctx, tenant); err != nil {
		return err
	}

	// Update cache
	tm.mu.Lock()
	tm.tenants[tenant.ID] = tenant
	tm.domains[tenant.Domain] = tenant.ID
	tm.mu.Unlock()

	return nil
}

// Delete deletes a tenant
func (tm *TenantManager) Delete(ctx context.Context, id string) error {
	tenant, err := tm.Get(ctx, id)
	if err != nil {
		return err
	}

	if err := tm.store.Delete(ctx, id); err != nil {
		return err
	}

	// Remove from cache
	tm.mu.Lock()
	delete(tm.tenants, id)
	delete(tm.domains, tenant.Domain)
	tm.mu.Unlock()

	return nil
}

// List lists tenants
func (tm *TenantManager) List(ctx context.Context, filter TenantFilter) ([]*Tenant, error) {
	return tm.store.List(ctx, filter)
}

// Suspend suspends a tenant
func (tm *TenantManager) Suspend(ctx context.Context, id string) error {
	tenant, err := tm.Get(ctx, id)
	if err != nil {
		return err
	}

	tenant.Status = TenantSuspended
	return tm.Update(ctx, tenant)
}

// Activate activates a tenant
func (tm *TenantManager) Activate(ctx context.Context, id string) error {
	tenant, err := tm.Get(ctx, id)
	if err != nil {
		return err
	}

	tenant.Status = TenantActive
	return tm.Update(ctx, tenant)
}

// Validate validates tenant access
func (tm *TenantManager) Validate(ctx context.Context, tenant *Tenant) error {
	if tenant == nil {
		return ErrInvalidTenant
	}

	switch tenant.Status {
	case TenantSuspended:
		return ErrTenantSuspended
	case TenantExpired:
		return ErrTenantExpired
	case TenantDeleted:
		return ErrTenantNotFound
	}

	// Check expiration
	if tenant.ExpiresAt != nil && time.Now().After(*tenant.ExpiresAt) {
		tenant.Status = TenantExpired
		tm.Update(ctx, tenant)
		return ErrTenantExpired
	}

	return nil
}

// Context helpers

// WithTenant adds tenant to context
func WithTenant(ctx context.Context, tenant *Tenant) context.Context {
	return context.WithValue(ctx, TenantContextKey, &TenantContext{
		Tenant: tenant,
	})
}

// WithTenantContext adds tenant context
func WithTenantContext(ctx context.Context, tc *TenantContext) context.Context {
	return context.WithValue(ctx, TenantContextKey, tc)
}

// GetTenant retrieves tenant from context
func GetTenant(ctx context.Context) (*Tenant, error) {
	tc, ok := ctx.Value(TenantContextKey).(*TenantContext)
	if !ok || tc == nil || tc.Tenant == nil {
		return nil, ErrTenantNotFound
	}
	return tc.Tenant, nil
}

// GetTenantContext retrieves tenant context
func GetTenantContext(ctx context.Context) (*TenantContext, error) {
	tc, ok := ctx.Value(TenantContextKey).(*TenantContext)
	if !ok || tc == nil {
		return nil, ErrTenantNotFound
	}
	return tc, nil
}

// MustGetTenant retrieves tenant from context or panics
func MustGetTenant(ctx context.Context) *Tenant {
	tenant, err := GetTenant(ctx)
	if err != nil {
		panic(err)
	}
	return tenant
}

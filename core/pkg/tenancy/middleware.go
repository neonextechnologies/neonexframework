package tenancy

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Middleware creates a Fiber middleware for multi-tenancy
func Middleware(manager *TenantManager, resolver *Resolver) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract tenant from request
		tenant, err := extractTenant(c, manager)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid or missing tenant",
			})
		}

		// Validate tenant
		if err := manager.Validate(c.Context(), tenant); err != nil {
			switch err {
			case ErrTenantSuspended:
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Tenant is suspended",
				})
			case ErrTenantExpired:
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Tenant subscription has expired",
				})
			default:
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Tenant access denied",
				})
			}
		}

		// Resolve database for tenant
		db, err := resolver.Resolve(c.Context(), tenant)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to resolve tenant database",
			})
		}

		// Store tenant in context
		ctx := WithTenant(c.Context(), tenant)
		c.SetUserContext(ctx)

		// Store database in locals
		c.Locals("tenant", tenant)
		c.Locals("tenant_db", db)

		// Add tenant header to response
		c.Set("X-Tenant-ID", tenant.ID)

		return c.Next()
	}
}

// extractTenant extracts tenant from request
func extractTenant(c *fiber.Ctx, manager *TenantManager) (*Tenant, error) {
	ctx := c.Context()

	// 1. Try subdomain (tenant.example.com)
	host := c.Hostname()
	parts := strings.Split(host, ".")
	if len(parts) >= 2 {
		subdomain := parts[0]
		if subdomain != "www" && subdomain != "api" {
			tenant, err := manager.GetByDomain(ctx, subdomain)
			if err == nil {
				return tenant, nil
			}
		}
	}

	// 2. Try custom domain (custom-domain.com)
	tenant, err := manager.GetByDomain(ctx, host)
	if err == nil {
		return tenant, nil
	}

	// 3. Try X-Tenant-ID header
	tenantID := c.Get("X-Tenant-ID")
	if tenantID != "" {
		return manager.Get(ctx, tenantID)
	}

	// 4. Try query parameter
	tenantID = c.Query("tenant_id")
	if tenantID != "" {
		return manager.Get(ctx, tenantID)
	}

	// 5. Try path parameter
	tenantID = c.Params("tenant_id")
	if tenantID != "" {
		return manager.Get(ctx, tenantID)
	}

	return nil, ErrTenantNotFound
}

// GetTenantFromLocals retrieves tenant from fiber locals
func GetTenantFromLocals(c *fiber.Ctx) (*Tenant, error) {
	tenant, ok := c.Locals("tenant").(*Tenant)
	if !ok || tenant == nil {
		return nil, ErrTenantNotFound
	}
	return tenant, nil
}

// GetTenantDBFromLocals retrieves tenant database from fiber locals
func GetTenantDBFromLocals(c *fiber.Ctx) (interface{}, error) {
	db := c.Locals("tenant_db")
	if db == nil {
		return nil, ErrTenantNotFound
	}
	return db, nil
}

// OptionalMiddleware creates optional tenant middleware (doesn't fail if tenant not found)
func OptionalMiddleware(manager *TenantManager, resolver *Resolver) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenant, err := extractTenant(c, manager)
		if err != nil {
			// Continue without tenant
			return c.Next()
		}

		// Validate tenant
		if err := manager.Validate(c.Context(), tenant); err != nil {
			// Continue without tenant
			return c.Next()
		}

		// Resolve database
		db, err := resolver.Resolve(c.Context(), tenant)
		if err != nil {
			return c.Next()
		}

		// Store in context and locals
		ctx := WithTenant(c.Context(), tenant)
		c.SetUserContext(ctx)
		c.Locals("tenant", tenant)
		c.Locals("tenant_db", db)
		c.Set("X-Tenant-ID", tenant.ID)

		return c.Next()
	}
}

// AdminMiddleware allows access only for admin operations
func AdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if request has admin privileges
		// This is a placeholder - implement your own admin check
		isAdmin := c.Get("X-Admin-Token") != ""
		
		if !isAdmin {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Admin access required",
			})
		}

		return c.Next()
	}
}

// RateLimitMiddleware creates per-tenant rate limiting
func RateLimitMiddleware(limiter func(tenantID string) bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenant, err := GetTenantFromLocals(c)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Tenant not found",
			})
		}

		if !limiter(tenant.ID) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded for tenant",
			})
		}

		return c.Next()
	}
}

// QuotaMiddleware checks tenant quotas
func QuotaMiddleware(checker func(tenant *Tenant) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenant, err := GetTenantFromLocals(c)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Tenant not found",
			})
		}

		if err := checker(tenant); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Next()
	}
}

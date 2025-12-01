# API Keys

NeonEx Framework supports API key authentication for machine-to-machine communication and third-party integrations.

## Table of Contents

- [Overview](#overview)
- [Generating API Keys](#generating-api-keys)
- [Storing API Keys](#storing-api-keys)
- [Validating API Keys](#validating-api-keys)
- [Usage in Requests](#usage-in-requests)
- [Middleware](#middleware)
- [Best Practices](#best-practices)
- [Advanced Patterns](#advanced-patterns)
- [Troubleshooting](#troubleshooting)

## Overview

API keys provide a simple authentication method for:
- **Third-party integrations** - External services accessing your API
- **Machine-to-machine** - Server-to-server communication
- **Mobile apps** - Alternative to session-based auth
- **Webhooks** - Authenticating callback requests

### API Key vs JWT

| Feature | API Key | JWT |
|---------|---------|-----|
| **Lifespan** | Long-lived (until revoked) | Short-lived (15 min access) |
| **Use Case** | M2M, integrations | User sessions |
| **Revocation** | Delete from database | Cannot revoke (must expire) |
| **Payload** | Opaque token | Contains claims |
| **Best For** | Server-to-server | Web/mobile apps |

## Generating API Keys

### Generation Function

```go
// core/pkg/auth/password.go
func GenerateAPIKey() (string, error) {
    return GenerateRandomToken(32)
}

func GenerateRandomToken(length int) (string, error) {
    bytes := make([]byte, length)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(bytes), nil
}
```

### Generate for User

```go
func (s *AuthService) GenerateAPIKey(ctx context.Context, userID uint) (string, error) {
    // Find user
    user, err := s.userRepo.FindByID(ctx, userID)
    if err != nil || user == nil {
        return "", errors.NewNotFound("User not found")
    }
    
    // Generate API key
    apiKey, err := auth.GenerateAPIKey()
    if err != nil {
        return "", errors.NewInternal("Failed to generate API key")
    }
    
    // Hash API key before storing
    hashedKey, err := s.hasher.Hash(apiKey)
    if err != nil {
        return "", errors.NewInternal("Failed to hash API key")
    }
    
    // Save hashed key to database
    user.APIKey = &hashedKey
    user.APIKeyCreatedAt = &time.Now()
    
    if err := s.userRepo.Update(ctx, user); err != nil {
        return "", errors.NewInternal("Failed to save API key")
    }
    
    // Return plain API key (only time it's visible)
    return apiKey, nil
}
```

### Controller Endpoint

```go
func (c *AuthController) GenerateAPIKey(ctx *fiber.Ctx) error {
    userID := ctx.Locals("user_id").(uint)
    
    apiKey, err := c.authService.GenerateAPIKey(ctx.Context(), userID)
    if err != nil {
        return ctx.Status(err.Code).JSON(fiber.Map{"error": err.Message})
    }
    
    return ctx.JSON(fiber.Map{
        "api_key": apiKey,
        "message": "Save this API key securely. It will not be shown again.",
    })
}
```

## Storing API Keys

### User Model with API Key

```go
type User struct {
    gorm.Model
    Name            string
    Email           string
    Password        string
    APIKey          *string    `gorm:"size:100;unique"`
    APIKeyCreatedAt *time.Time
    IsActive        bool
}
```

### Dedicated API Key Model

For multiple keys per user:

```go
type APIKey struct {
    ID        uint           `gorm:"primarykey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
    
    UserID      uint       `gorm:"not null;index"`
    Name        string     `gorm:"size:100"` // "Production Server", "Mobile App"
    KeyHash     string     `gorm:"size:255;unique;not null"`
    LastUsedAt  *time.Time
    ExpiresAt   *time.Time
    Permissions []string   `gorm:"type:json"`
    IsActive    bool       `gorm:"default:true"`
    
    User User `gorm:"foreignKey:UserID"`
}
```

### Create API Key Entry

```go
func (s *APIKeyService) CreateAPIKey(ctx context.Context, userID uint, name string, permissions []string) (string, error) {
    // Generate key
    apiKey, err := auth.GenerateAPIKey()
    if err != nil {
        return "", err
    }
    
    // Hash key
    hashedKey, err := s.hasher.Hash(apiKey)
    if err != nil {
        return "", err
    }
    
    // Create entry
    keyEntry := &APIKey{
        UserID:      userID,
        Name:        name,
        KeyHash:     hashedKey,
        Permissions: permissions,
        IsActive:    true,
    }
    
    if err := s.db.Create(keyEntry).Error; err != nil {
        return "", err
    }
    
    return apiKey, nil
}
```

## Validating API Keys

### Basic Validation

```go
func (s *AuthService) ValidateAPIKey(ctx context.Context, apiKey string) (*User, error) {
    if apiKey == "" {
        return nil, errors.NewUnauthorized("API key required")
    }
    
    // Find all users with API keys
    var users []User
    if err := s.userRepo.FindByCondition(ctx, "api_key IS NOT NULL"); err != nil {
        return nil, errors.NewInternal("Database error")
    }
    
    // Check each hashed API key
    for _, user := range users {
        if user.APIKey == nil {
            continue
        }
        
        // Verify API key matches hash
        if err := s.hasher.Verify(apiKey, *user.APIKey); err == nil {
            // Found matching user
            if !user.IsActive {
                return nil, errors.NewUnauthorized("Account is disabled")
            }
            return &user, nil
        }
    }
    
    return nil, errors.NewUnauthorized("Invalid API key")
}
```

### Advanced Validation with Dedicated Table

```go
func (s *APIKeyService) ValidateAPIKey(ctx context.Context, apiKey string) (*APIKey, error) {
    // Get all active API keys
    var keys []APIKey
    if err := s.db.Preload("User").Where("is_active = ?", true).Find(&keys).Error; err != nil {
        return nil, err
    }
    
    // Find matching key
    for _, key := range keys {
        if err := s.hasher.Verify(apiKey, key.KeyHash); err == nil {
            // Check expiration
            if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
                return nil, errors.NewUnauthorized("API key expired")
            }
            
            // Update last used
            now := time.Now()
            key.LastUsedAt = &now
            s.db.Save(&key)
            
            return &key, nil
        }
    }
    
    return nil, errors.NewUnauthorized("Invalid API key")
}
```

## Usage in Requests

### HTTP Header

```bash
# Standard authorization header
curl -H "Authorization: Bearer YOUR_API_KEY" \
     https://api.example.com/users

# Custom header
curl -H "X-API-Key: YOUR_API_KEY" \
     https://api.example.com/users
```

### Query Parameter

```bash
curl "https://api.example.com/users?api_key=YOUR_API_KEY"
```

### Request Body

```bash
curl -X POST https://api.example.com/data \
     -H "Content-Type: application/json" \
     -d '{"api_key": "YOUR_API_KEY", "data": "..."}'
```

## Middleware

### API Key Authentication Middleware

```go
// pkg/http/middleware/apikey.go
package middleware

func APIKeyAuth(authService *auth.AuthService) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Try multiple sources
        apiKey := extractAPIKey(c)
        
        if apiKey == "" {
            return c.Status(401).JSON(fiber.Map{
                "error": "API key required",
            })
        }
        
        // Validate API key
        user, err := authService.ValidateAPIKey(c.Context(), apiKey)
        if err != nil {
            return c.Status(401).JSON(fiber.Map{
                "error": "Invalid API key",
            })
        }
        
        // Store user in context
        c.Locals("user_id", user.ID)
        c.Locals("user", user)
        c.Locals("auth_method", "api_key")
        
        return c.Next()
    }
}

func extractAPIKey(c *fiber.Ctx) string {
    // 1. Check Authorization header
    auth := c.Get("Authorization")
    if strings.HasPrefix(auth, "Bearer ") {
        return strings.TrimPrefix(auth, "Bearer ")
    }
    
    // 2. Check X-API-Key header
    if apiKey := c.Get("X-API-Key"); apiKey != "" {
        return apiKey
    }
    
    // 3. Check query parameter
    if apiKey := c.Query("api_key"); apiKey != "" {
        return apiKey
    }
    
    return ""
}
```

### Combined Auth Middleware (JWT or API Key)

```go
func FlexibleAuth(authService *auth.AuthService, jwtManager *auth.JWTManager) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Try JWT first
        token := c.Get("Authorization")
        if strings.HasPrefix(token, "Bearer ") {
            tokenStr := strings.TrimPrefix(token, "Bearer ")
            
            // Try as JWT
            claims, err := jwtManager.ValidateToken(tokenStr)
            if err == nil {
                c.Locals("user_id", claims.UserID)
                c.Locals("auth_method", "jwt")
                return c.Next()
            }
            
            // Try as API key
            user, err := authService.ValidateAPIKey(c.Context(), tokenStr)
            if err == nil {
                c.Locals("user_id", user.ID)
                c.Locals("auth_method", "api_key")
                return c.Next()
            }
        }
        
        return c.Status(401).JSON(fiber.Map{
            "error": "Authentication required",
        })
    }
}
```

### Usage in Routes

```go
func RegisterAPIRoutes(router fiber.Router, authService *auth.AuthService) {
    api := router.Group("/api/v1")
    
    // Public endpoints
    api.Get("/health", healthCheck)
    
    // Protected endpoints (API key required)
    api.Use(middleware.APIKeyAuth(authService))
    
    api.Get("/users", listUsers)
    api.Post("/users", createUser)
    api.Get("/products", listProducts)
}
```

## Best Practices

### 1. Hash API Keys

```go
// ✅ Good: Store hashed API keys
hashedKey, _ := hasher.Hash(apiKey)
user.APIKey = &hashedKey

// ❌ Bad: Store plain API keys
user.APIKey = &apiKey // Vulnerable if database is compromised
```

### 2. Use HTTPS Only

```go
// Enforce HTTPS
func (c *AuthController) GenerateAPIKey(ctx *fiber.Ctx) error {
    if ctx.Protocol() != "https" {
        return ctx.Status(400).JSON(fiber.Map{
            "error": "API key generation requires HTTPS",
        })
    }
    
    // Generate key...
}
```

### 3. Implement Rate Limiting

```go
func APIRateLimiter(cache cache.Cache) fiber.Handler {
    return func(c *fiber.Ctx) error {
        apiKey := extractAPIKey(c)
        cacheKey := fmt.Sprintf("rate:api:%s", apiKey)
        
        count, _ := cache.Increment(c.Context(), cacheKey, 1)
        if count == 1 {
            cache.Expire(c.Context(), cacheKey, 1*time.Minute)
        }
        
        if count > 100 { // 100 requests per minute
            return c.Status(429).JSON(fiber.Map{
                "error": "Rate limit exceeded",
            })
        }
        
        return c.Next()
    }
}
```

### 4. Track API Key Usage

```go
func (s *APIKeyService) LogUsage(ctx context.Context, keyID uint, endpoint string) {
    usage := &APIKeyUsage{
        APIKeyID:  keyID,
        Endpoint:  endpoint,
        IP:        ctx.Value("ip").(string),
        UserAgent: ctx.Value("user_agent").(string),
        Timestamp: time.Now(),
    }
    s.db.Create(usage)
}
```

### 5. Implement Key Rotation

```go
func (s *APIKeyService) RotateAPIKey(ctx context.Context, oldKeyID uint) (string, error) {
    // Generate new key
    newKey, err := s.CreateAPIKey(ctx, userID, "Rotated Key", permissions)
    if err != nil {
        return "", err
    }
    
    // Deactivate old key (don't delete immediately)
    s.db.Model(&APIKey{}).Where("id = ?", oldKeyID).Update("is_active", false)
    
    return newKey, nil
}
```

## Advanced Patterns

### Scoped API Keys

```go
type APIKey struct {
    // ... other fields
    Scopes      []string `gorm:"type:json"` // ["read:users", "write:products"]
}

func (s *APIKeyService) CheckScope(key *APIKey, requiredScope string) bool {
    for _, scope := range key.Scopes {
        if scope == requiredScope || scope == "*" {
            return true
        }
    }
    return false
}

// Middleware
func RequireScope(scope string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        apiKey := c.Locals("api_key").(*APIKey)
        
        if !apiKeyService.CheckScope(apiKey, scope) {
            return c.Status(403).JSON(fiber.Map{
                "error": "Insufficient scope",
            })
        }
        
        return c.Next()
    }
}
```

### API Key Prefixes

```go
// Add prefix for easy identification
func GenerateAPIKeyWithPrefix(prefix string) (string, error) {
    token, err := GenerateRandomToken(32)
    if err != nil {
        return "", err
    }
    
    // Format: prefix_token (e.g., "nxk_live_abc123...")
    return fmt.Sprintf("%s_%s", prefix, token), nil
}

// Example prefixes
nxk_test_...  // Test environment
nxk_live_...  // Production environment
nxk_dev_...   // Development
```

### Temporary API Keys

```go
func (s *APIKeyService) CreateTemporaryKey(ctx context.Context, userID uint, duration time.Duration) (string, error) {
    apiKey, err := s.CreateAPIKey(ctx, userID, "Temporary", nil)
    if err != nil {
        return "", err
    }
    
    expiresAt := time.Now().Add(duration)
    s.db.Model(&APIKey{}).
        Where("key_hash = ?", hashedKey).
        Update("expires_at", expiresAt)
    
    return apiKey, nil
}
```

### IP Whitelisting

```go
type APIKey struct {
    // ... other fields
    AllowedIPs []string `gorm:"type:json"`
}

func (s *APIKeyService) ValidateIP(key *APIKey, ip string) bool {
    if len(key.AllowedIPs) == 0 {
        return true // No restriction
    }
    
    for _, allowedIP := range key.AllowedIPs {
        if allowedIP == ip {
            return true
        }
    }
    
    return false
}
```

## Troubleshooting

### Common Issues

**Issue: API key validation is slow**

```go
// Solution: Add index on key_hash
// migration
db.Exec("CREATE INDEX idx_api_keys_hash ON api_keys(key_hash)")

// Or use GORM
type APIKey struct {
    KeyHash string `gorm:"size:255;unique;not null;index"`
}
```

**Issue: How to revoke compromised keys**

```go
func (s *APIKeyService) RevokeKey(ctx context.Context, keyID uint) error {
    return s.db.Model(&APIKey{}).
        Where("id = ?", keyID).
        Updates(map[string]interface{}{
            "is_active":  false,
            "revoked_at": time.Now(),
        }).Error
}
```

**Issue: Users lose their API key**

```go
// Solution: Allow regeneration
func (c *AuthController) RegenerateAPIKey(ctx *fiber.Ctx) error {
    userID := ctx.Locals("user_id").(uint)
    
    // Revoke old key
    c.authService.RevokeAPIKey(ctx.Context(), userID)
    
    // Generate new key
    newKey, err := c.authService.GenerateAPIKey(ctx.Context(), userID)
    if err != nil {
        return ctx.Status(500).JSON(fiber.Map{"error": "Failed to generate key"})
    }
    
    return ctx.JSON(fiber.Map{
        "api_key": newKey,
        "message": "New API key generated. Old key is now invalid.",
    })
}
```

## Summary

NeonEx Framework API key authentication provides:

✅ **Secure generation** - Cryptographically random keys  
✅ **Hashed storage** - Keys hashed like passwords  
✅ **Flexible validation** - Multiple authentication sources  
✅ **Usage tracking** - Monitor API key activity  
✅ **Scoped access** - Limit permissions per key  
✅ **Production-ready** - Rate limiting, expiration, rotation

For more information:
- [Authentication](authentication.md)
- [JWT Security](jwt.md)
- [Middleware](../core-concepts/middleware.md)

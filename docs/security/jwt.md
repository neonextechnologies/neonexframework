# JWT Authentication

NeonEx Framework provides a complete JWT (JSON Web Token) authentication system with token generation, validation, refresh tokens, and middleware integration.

## Table of Contents

- [Overview](#overview)
- [Configuration](#configuration)
- [Token Generation](#token-generation)
- [Token Validation](#token-validation)
- [Refresh Tokens](#refresh-tokens)
- [Middleware Integration](#middleware-integration)
- [Best Practices](#best-practices)

## Overview

JWT authentication in NeonEx includes:
- Access and refresh tokens
- Configurable expiry times
- Claims with user data and permissions
- Automatic token validation
- Token refresh mechanism

## Configuration

### JWT Config Structure

```go
type JWTConfig struct {
    SecretKey     string        // Secret for signing tokens
    AccessExpiry  time.Duration // Access token lifetime
    RefreshExpiry time.Duration // Refresh token lifetime
    Issuer        string        // Token issuer
    Algorithm     string        // Signing algorithm (default: HS256)
}
```

### Setting Up JWT Manager

```go
import "neonexcore/pkg/auth"

// Create JWT configuration
jwtConfig := &auth.JWTConfig{
    SecretKey:     "your-secret-key-min-32-chars",
    AccessExpiry:  15 * time.Minute,
    RefreshExpiry: 7 * 24 * time.Hour,
    Issuer:        "neonex-app",
}

// Initialize JWT manager
jwtManager := auth.NewJWTManager(jwtConfig)
```

### Environment Variables

```env
JWT_SECRET=your-secret-key-here-minimum-32-characters
JWT_ACCESS_EXPIRY=900      # 15 minutes in seconds
JWT_REFRESH_EXPIRY=604800  # 7 days in seconds
JWT_ISSUER=neonex-app
```

## Token Generation

### Generate Access Token

```go
func (s *AuthService) Login(email, password string) (string, error) {
    // Validate credentials
    user, err := s.userRepo.FindByEmail(context.Background(), email)
    if err != nil {
        return "", errors.New("user not found")
    }
    
    // Verify password
    hasher := auth.NewPasswordHasher(12)
    if err := hasher.Verify(password, user.Password); err != nil {
        return "", errors.New("invalid password")
    }
    
    // Get user permissions
    permissions := s.getUserPermissions(user.ID)
    
    // Generate access token
    token, err := s.jwtManager.GenerateAccessToken(
        user.ID,
        user.Email,
        user.Role,
        permissions,
    )
    if err != nil {
        return "", fmt.Errorf("failed to generate token: %w", err)
    }
    
    // Update last login
    user.LastLoginAt = &time.Now()
    s.userRepo.Update(context.Background(), user)
    
    return token, nil
}
```

### Generate Refresh Token

```go
func (s *AuthService) GenerateTokenPair(userID uint, email string) (map[string]string, error) {
    // Get user data
    user, err := s.userRepo.FindByID(context.Background(), userID)
    if err != nil {
        return nil, err
    }
    
    permissions := s.getUserPermissions(userID)
    
    // Generate access token
    accessToken, err := s.jwtManager.GenerateAccessToken(
        user.ID,
        user.Email,
        user.Role,
        permissions,
    )
    if err != nil {
        return nil, err
    }
    
    // Generate refresh token
    refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email)
    if err != nil {
        return nil, err
    }
    
    return map[string]string{
        "access_token":  accessToken,
        "refresh_token": refreshToken,
        "token_type":    "Bearer",
    }, nil
}
```

### Login Handler

```go
type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    TokenType    string `json:"token_type"`
    ExpiresIn    int    `json:"expires_in"`
    User         *User  `json:"user"`
}

func loginHandler(c *fiber.Ctx) error {
    var req LoginRequest
    
    if err := c.BodyParser(&req); err != nil {
        return api.BadRequest(c, "Invalid request", nil)
    }
    
    if errors := validator.Validate(req); errors != nil {
        return api.ValidationError(c, errors)
    }
    
    // Login
    tokens, err := authService.Login(req.Email, req.Password)
    if err != nil {
        return api.Unauthorized(c, "Invalid credentials")
    }
    
    // Get user data
    user, _ := userService.GetByEmail(req.Email)
    
    return api.Success(c, LoginResponse{
        AccessToken:  tokens["access_token"],
        RefreshToken: tokens["refresh_token"],
        TokenType:    "Bearer",
        ExpiresIn:    900, // 15 minutes
        User:         user,
    })
}
```

## Token Validation

### Validate Token

```go
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // Verify signing method
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, ErrInvalidSignature
        }
        return []byte(m.config.SecretKey), nil
    })
    
    if err != nil {
        if errors.Is(err, jwt.ErrTokenExpired) {
            return nil, ErrExpiredToken
        }
        return nil, ErrInvalidToken
    }
    
    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, ErrInvalidToken
    }
    
    return claims, nil
}
```

### Claims Structure

```go
type Claims struct {
    UserID      uint              `json:"user_id"`
    Email       string            `json:"email"`
    Role        string            `json:"role"`
    Permissions []string          `json:"permissions"`
    Metadata    map[string]string `json:"metadata,omitempty"`
    jwt.RegisteredClaims
}
```

### Getting User from Token

```go
func getUserFromToken(c *fiber.Ctx) (*User, error) {
    // Get token from header
    authHeader := c.Get("Authorization")
    if authHeader == "" {
        return nil, errors.New("no authorization header")
    }
    
    parts := strings.Split(authHeader, " ")
    if len(parts) != 2 || parts[0] != "Bearer" {
        return nil, errors.New("invalid authorization format")
    }
    
    // Validate token
    claims, err := jwtManager.ValidateToken(parts[1])
    if err != nil {
        return nil, err
    }
    
    // Get user from database
    user, err := userRepo.FindByID(context.Background(), claims.UserID)
    if err != nil {
        return nil, err
    }
    
    return user, nil
}
```

## Refresh Tokens

### Refresh Access Token

```go
type RefreshRequest struct {
    RefreshToken string `json:"refresh_token" validate:"required"`
}

func refreshTokenHandler(c *fiber.Ctx) error {
    var req RefreshRequest
    
    if err := c.BodyParser(&req); err != nil {
        return api.BadRequest(c, "Invalid request", nil)
    }
    
    // Validate refresh token
    claims, err := jwtManager.ValidateToken(req.RefreshToken)
    if err != nil {
        return api.Unauthorized(c, "Invalid refresh token")
    }
    
    // Get user
    user, err := userService.GetByID(claims.UserID)
    if err != nil {
        return api.Unauthorized(c, "User not found")
    }
    
    // Check if user is still active
    if !user.Active {
        return api.Forbidden(c, "Account is disabled")
    }
    
    // Get permissions
    permissions := authService.getUserPermissions(user.ID)
    
    // Generate new access token
    accessToken, err := jwtManager.GenerateAccessToken(
        user.ID,
        user.Email,
        user.Role,
        permissions,
    )
    if err != nil {
        return api.InternalError(c, "Failed to generate token")
    }
    
    return api.Success(c, fiber.Map{
        "access_token": accessToken,
        "token_type":   "Bearer",
        "expires_in":   900,
    })
}
```

### Automatic Refresh

```go
func AutoRefreshMiddleware(jwtManager *auth.JWTManager) fiber.Handler {
    return func(c *fiber.Ctx) error {
        authHeader := c.Get("Authorization")
        if authHeader == "" {
            return c.Next()
        }
        
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 {
            return c.Next()
        }
        
        claims, err := jwtManager.ValidateToken(parts[1])
        if err == nil {
            // Check if token expires soon (within 5 minutes)
            if time.Until(claims.ExpiresAt.Time) < 5*time.Minute {
                // Generate new token
                newToken, _ := jwtManager.GenerateAccessToken(
                    claims.UserID,
                    claims.Email,
                    claims.Role,
                    claims.Permissions,
                )
                
                // Set new token in response header
                c.Set("X-New-Token", newToken)
            }
        }
        
        return c.Next()
    }
}
```

## Middleware Integration

### Authentication Middleware

```go
import "neonexcore/pkg/auth"

// Protect routes
api := app.Group("/api")
api.Use(auth.AuthMiddleware(jwtManager))

api.Get("/profile", getProfile)
api.Put("/profile", updateProfile)
```

### Optional Authentication

```go
// Optional auth (allows guests)
app.Use(auth.OptionalAuthMiddleware(jwtManager))

func handler(c *fiber.Ctx) error {
    userID, authenticated := auth.GetUserID(c)
    
    if authenticated {
        // Authenticated user logic
    } else {
        // Guest logic
    }
}
```

### Extract User Data

```go
func getProfile(c *fiber.Ctx) error {
    // Get user ID from context (set by AuthMiddleware)
    userID := c.Locals("user_id").(uint)
    email := c.Locals("email").(string)
    role := c.Locals("role").(string)
    permissions := c.Locals("permissions").([]string)
    
    // Get full user data
    user, err := userService.GetByID(userID)
    if err != nil {
        return api.NotFound(c, "User not found")
    }
    
    return api.Success(c, user)
}

// Or use helper functions
func handler(c *fiber.Ctx) error {
    userID, ok := auth.GetUserID(c)
    email, ok := auth.GetUserEmail(c)
    role, ok := auth.GetUserRole(c)
    permissions, ok := auth.GetUserPermissions(c)
    claims, ok := auth.GetClaims(c)
}
```

## Best Practices

### 1. Use Strong Secret Keys

```go
// Good: Strong random secret (32+ characters)
JWT_SECRET=a8f5f167f44f4964e6c998dee827110c3d8a0678ec1f5d3c9e9f4c2a3b7e6d5c

// Bad: Weak secret
JWT_SECRET=secret123
```

### 2. Set Appropriate Expiry Times

```go
// Recommended settings
jwtConfig := &auth.JWTConfig{
    AccessExpiry:  15 * time.Minute,  // Short-lived
    RefreshExpiry: 7 * 24 * time.Hour, // Long-lived
}

// High-security applications
jwtConfig := &auth.JWTConfig{
    AccessExpiry:  5 * time.Minute,   // Very short
    RefreshExpiry: 24 * time.Hour,    // 1 day
}
```

### 3. Include Minimal Claims

```go
// Good: Only essential data
claims := &Claims{
    UserID:      user.ID,
    Email:       user.Email,
    Role:        user.Role,
    Permissions: permissions,
}

// Bad: Too much data
claims := &Claims{
    UserID:   user.ID,
    Email:    user.Email,
    Name:     user.Name,
    Bio:      user.Bio,
    Avatar:   user.Avatar,
    // ... don't include everything!
}
```

### 4. Validate Tokens Properly

```go
// Good: Proper validation
claims, err := jwtManager.ValidateToken(token)
if err != nil {
    if errors.Is(err, auth.ErrExpiredToken) {
        return api.Unauthorized(c, "Token expired")
    }
    return api.Unauthorized(c, "Invalid token")
}

// Bad: No error handling
claims, _ := jwtManager.ValidateToken(token)
```

### 5. Implement Token Blacklisting

```go
type TokenBlacklist struct {
    cache cache.Cache
}

func (tb *TokenBlacklist) Add(token string, expiry time.Duration) error {
    return tb.cache.Set(context.Background(), "blacklist:"+token, true, expiry)
}

func (tb *TokenBlacklist) IsBlacklisted(token string) bool {
    exists, _ := tb.cache.Exists(context.Background(), "blacklist:"+token)
    return exists
}

// Use in middleware
func AuthMiddleware(jwtManager *auth.JWTManager, blacklist *TokenBlacklist) fiber.Handler {
    return func(c *fiber.Ctx) error {
        token := extractToken(c)
        
        // Check blacklist
        if blacklist.IsBlacklisted(token) {
            return api.Unauthorized(c, "Token has been revoked")
        }
        
        // Validate token
        claims, err := jwtManager.ValidateToken(token)
        if err != nil {
            return api.Unauthorized(c, "Invalid token")
        }
        
        c.Locals("user_id", claims.UserID)
        return c.Next()
    }
}

// Logout handler
func logout(c *fiber.Ctx) error {
    token := extractToken(c)
    claims, _ := jwtManager.ValidateToken(token)
    
    // Add to blacklist
    expiry := time.Until(claims.ExpiresAt.Time)
    blacklist.Add(token, expiry)
    
    return api.Success(c, fiber.Map{"message": "Logged out successfully"})
}
```

### 6. Secure Token Storage (Client-Side)

```javascript
// Good: HttpOnly cookies (set from server)
res.cookie('refreshToken', refreshToken, {
    httpOnly: true,
    secure: true,
    sameSite: 'strict',
    maxAge: 7 * 24 * 60 * 60 * 1000
});

// Acceptable: localStorage (for access token only)
localStorage.setItem('accessToken', accessToken);

// Bad: Never store refresh token in localStorage
localStorage.setItem('refreshToken', refreshToken); // DON'T DO THIS
```

This comprehensive guide covers JWT authentication in NeonEx Framework!

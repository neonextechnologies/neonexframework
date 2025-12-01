# Authentication

NeonEx Framework provides comprehensive authentication functionality including user registration, login, token refresh, password management, and email verification.

## Table of Contents

- [Overview](#overview)
- [Authentication Flow](#authentication-flow)
- [Registration](#registration)
- [Login](#login)
- [Token Management](#token-management)
- [Password Management](#password-management)
- [Email Verification](#email-verification)
- [Profile Management](#profile-management)
- [Best Practices](#best-practices)
- [Advanced Patterns](#advanced-patterns)
- [Troubleshooting](#troubleshooting)

## Overview

NeonEx uses **JWT (JSON Web Tokens)** for stateless authentication with:
- Access tokens (short-lived, 15 minutes)
- Refresh tokens (long-lived, 7 days)
- Password hashing with bcrypt
- Role-based access control integration

### Components

- **JWTManager** - Token generation and validation
- **PasswordHasher** - Secure password hashing
- **AuthService** - Authentication business logic
- **AuthController** - HTTP handlers

## Authentication Flow

### Complete Auth Flow Diagram

```
┌──────────┐         ┌──────────┐         ┌──────────┐
│  Client  │         │  Server  │         │ Database │
└────┬─────┘         └────┬─────┘         └────┬─────┘
     │                    │                     │
     │ 1. POST /register  │                     │
     ├───────────────────>│                     │
     │                    │ 2. Hash password    │
     │                    ├────────────────────>│
     │                    │ 3. Create user      │
     │                    │<────────────────────┤
     │ 4. User created    │                     │
     │<───────────────────┤                     │
     │                    │                     │
     │ 5. POST /login     │                     │
     ├───────────────────>│                     │
     │                    │ 6. Verify password  │
     │                    ├────────────────────>│
     │                    │<────────────────────┤
     │                    │ 7. Generate tokens  │
     │ 8. Access+Refresh  │                     │
     │<───────────────────┤                     │
     │                    │                     │
     │ 9. API call        │                     │
     │   (with token)     │                     │
     ├───────────────────>│                     │
     │                    │ 10. Validate token  │
     │                    │ 11. Process request │
     │ 12. Response       │                     │
     │<───────────────────┤                     │
```

## Registration

### Basic Registration

```go
// modules/user/auth_service.go
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*User, error) {
    // 1. Validate email doesn't exist
    existing, _ := s.userRepo.FindByEmail(ctx, req.Email)
    if existing != nil {
        return nil, errors.NewConflict("Email already exists")
    }
    
    // 2. Validate username doesn't exist
    existing, _ = s.userRepo.FindByUsername(ctx, req.Username)
    if existing != nil {
        return nil, errors.NewConflict("Username already exists")
    }
    
    // 3. Hash password
    hashedPassword, err := s.hasher.Hash(req.Password)
    if err != nil {
        return nil, errors.NewInternal("Failed to hash password")
    }
    
    // 4. Create user
    user := &User{
        Name:     req.Name,
        Email:    req.Email,
        Username: req.Username,
        Password: hashedPassword,
        IsActive: true,
    }
    
    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, errors.NewInternal("Failed to create user")
    }
    
    // 5. Assign default role
    role, _ := s.rbacManager.GetRoleBySlug(ctx, "user")
    if role != nil {
        s.rbacManager.AssignRole(ctx, user.ID, role.ID)
    }
    
    // 6. Dispatch event
    events.DispatchAsync(ctx, events.Event{
        Name: events.EventUserCreated,
        Data: map[string]interface{}{
            "user_id": user.ID,
            "email":   user.Email,
        },
    })
    
    return user, nil
}
```

### Registration Request Validation

```go
type RegisterRequest struct {
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Email    string `json:"email" validate:"required,email"`
    Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
    Password string `json:"password" validate:"required,min=8,max=72"`
}
```

### Registration Controller

```go
// modules/user/auth_controller.go
func (c *AuthController) Register(ctx *fiber.Ctx) error {
    var req RegisterRequest
    if err := ctx.BodyParser(&req); err != nil {
        return ctx.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }
    
    // Validate request
    if err := c.validator.Validate(req); err != nil {
        return ctx.Status(400).JSON(fiber.Map{
            "error":   "Validation failed",
            "details": err,
        })
    }
    
    // Register user
    user, err := c.authService.Register(ctx.Context(), &req)
    if err != nil {
        return ctx.Status(err.Code).JSON(fiber.Map{
            "error": err.Message,
        })
    }
    
    return ctx.Status(201).JSON(fiber.Map{
        "message": "User registered successfully",
        "user": fiber.Map{
            "id":       user.ID,
            "name":     user.Name,
            "email":    user.Email,
            "username": user.Username,
        },
    })
}
```

## Login

### Login Flow

```go
func (s *AuthService) Login(ctx context.Context, email, password string) (map[string]interface{}, error) {
    // 1. Find user
    user, err := s.userRepo.FindByEmail(ctx, email)
    if err != nil || user == nil {
        return nil, errors.New(errors.ErrCodeInvalidCredentials, "Invalid email or password", 401)
    }
    
    // 2. Check if user is active
    if !user.IsActive {
        return nil, errors.New(errors.ErrCodeAccountDisabled, "Account is disabled", 403)
    }
    
    // 3. Verify password
    if err := s.hasher.Verify(password, user.Password); err != nil {
        return nil, errors.New(errors.ErrCodeInvalidCredentials, "Invalid email or password", 401)
    }
    
    // 4. Get user roles and permissions
    roles, _ := s.rbacManager.GetUserRoles(ctx, user.ID)
    permissions, _ := s.rbacManager.GetUserPermissions(ctx, user.ID)
    
    var roleNames []string
    for _, role := range roles {
        roleNames = append(roleNames, role.Slug)
    }
    
    var permissionSlugs []string
    for _, perm := range permissions {
        permissionSlugs = append(permissionSlugs, perm.Slug)
    }
    
    // 5. Generate tokens
    primaryRole := "user"
    if len(roleNames) > 0 {
        primaryRole = roleNames[0]
    }
    
    accessToken, err := s.jwtManager.GenerateAccessToken(
        user.ID,
        user.Email,
        primaryRole,
        permissionSlugs,
    )
    if err != nil {
        return nil, errors.NewInternal("Failed to generate access token")
    }
    
    refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email)
    if err != nil {
        return nil, errors.NewInternal("Failed to generate refresh token")
    }
    
    // 6. Update last login
    now := time.Now()
    user.LastLoginAt = &now
    s.userRepo.Update(ctx, user)
    
    // 7. Dispatch login event
    events.DispatchAsync(ctx, events.Event{
        Name: events.EventUserLoggedIn,
        Data: map[string]interface{}{
            "user_id": user.ID,
            "email":   user.Email,
        },
    })
    
    return map[string]interface{}{
        "access_token":  accessToken,
        "refresh_token": refreshToken,
        "token_type":    "Bearer",
        "expires_in":    900, // 15 minutes in seconds
        "user": map[string]interface{}{
            "id":       user.ID,
            "name":     user.Name,
            "email":    user.Email,
            "username": user.Username,
            "roles":    roleNames,
        },
    }, nil
}
```

### Login Controller

```go
type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

func (c *AuthController) Login(ctx *fiber.Ctx) error {
    var req LoginRequest
    if err := ctx.BodyParser(&req); err != nil {
        return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }
    
    result, err := c.authService.Login(ctx.Context(), req.Email, req.Password)
    if err != nil {
        return ctx.Status(err.Code).JSON(fiber.Map{"error": err.Message})
    }
    
    return ctx.JSON(result)
}
```

## Token Management

### Refresh Token

```go
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (map[string]interface{}, error) {
    // Validate refresh token
    accessToken, err := s.jwtManager.RefreshAccessToken(refreshToken)
    if err != nil {
        return nil, errors.NewUnauthorized("Invalid refresh token")
    }
    
    return map[string]interface{}{
        "access_token": accessToken,
        "token_type":   "Bearer",
        "expires_in":   900,
    }, nil
}
```

### Refresh Token Controller

```go
type RefreshTokenRequest struct {
    RefreshToken string `json:"refresh_token" validate:"required"`
}

func (c *AuthController) RefreshToken(ctx *fiber.Ctx) error {
    var req RefreshTokenRequest
    if err := ctx.BodyParser(&req); err != nil {
        return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }
    
    result, err := c.authService.RefreshToken(ctx.Context(), req.RefreshToken)
    if err != nil {
        return ctx.Status(err.Code).JSON(fiber.Map{"error": err.Message})
    }
    
    return ctx.JSON(result)
}
```

### Logout

```go
func (c *AuthController) Logout(ctx *fiber.Ctx) error {
    // For JWT, logout is handled client-side by removing tokens
    // Optionally, implement token blacklist for server-side logout
    
    userID := ctx.Locals("user_id").(uint)
    
    // Dispatch logout event
    events.DispatchAsync(ctx.Context(), events.Event{
        Name: events.EventUserLoggedOut,
        Data: map[string]interface{}{
            "user_id": userID,
        },
    })
    
    return ctx.JSON(fiber.Map{
        "message": "Logged out successfully",
    })
}
```

## Password Management

### Change Password

```go
func (s *AuthService) ChangePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error {
    // 1. Get user
    user, err := s.userRepo.FindByID(ctx, userID)
    if err != nil || user == nil {
        return errors.NewNotFound("User not found")
    }
    
    // 2. Verify current password
    if err := s.hasher.Verify(currentPassword, user.Password); err != nil {
        return errors.NewBadRequest("Current password is incorrect")
    }
    
    // 3. Hash new password
    hashedPassword, err := s.hasher.Hash(newPassword)
    if err != nil {
        return errors.NewInternal("Failed to hash password")
    }
    
    // 4. Update password
    user.Password = hashedPassword
    return s.userRepo.Update(ctx, user)
}
```

### Forgot Password

```go
func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
    // 1. Find user
    user, err := s.userRepo.FindByEmail(ctx, email)
    if err != nil || user == nil {
        // Don't reveal if email exists
        return nil
    }
    
    // 2. Generate reset token
    resetToken, err := auth.GenerateResetToken()
    if err != nil {
        return errors.NewInternal("Failed to generate reset token")
    }
    
    // 3. Save reset token (hash it)
    hashedToken, _ := s.hasher.Hash(resetToken)
    user.ResetToken = &hashedToken
    expiresAt := time.Now().Add(1 * time.Hour)
    user.ResetTokenExpiresAt = &expiresAt
    
    if err := s.userRepo.Update(ctx, user); err != nil {
        return errors.NewInternal("Failed to save reset token")
    }
    
    // 4. Send email (async)
    go s.notificationService.SendEmail(ctx, user.Email, "Password Reset", 
        fmt.Sprintf("Reset token: %s", resetToken))
    
    return nil
}
```

### Reset Password

```go
func (s *AuthService) ResetPassword(ctx context.Context, email, token, newPassword string) error {
    // 1. Find user
    user, err := s.userRepo.FindByEmail(ctx, email)
    if err != nil || user == nil {
        return errors.NewBadRequest("Invalid reset request")
    }
    
    // 2. Check token exists and not expired
    if user.ResetToken == nil || user.ResetTokenExpiresAt == nil {
        return errors.NewBadRequest("Invalid or expired reset token")
    }
    
    if time.Now().After(*user.ResetTokenExpiresAt) {
        return errors.NewBadRequest("Reset token has expired")
    }
    
    // 3. Verify token
    if err := s.hasher.Verify(token, *user.ResetToken); err != nil {
        return errors.NewBadRequest("Invalid reset token")
    }
    
    // 4. Hash new password
    hashedPassword, err := s.hasher.Hash(newPassword)
    if err != nil {
        return errors.NewInternal("Failed to hash password")
    }
    
    // 5. Update password and clear reset token
    user.Password = hashedPassword
    user.ResetToken = nil
    user.ResetTokenExpiresAt = nil
    
    if err := s.userRepo.Update(ctx, user); err != nil {
        return errors.NewInternal("Failed to update password")
    }
    
    // 6. Dispatch event
    events.DispatchAsync(ctx, events.Event{
        Name: events.EventUserPasswordReset,
        Data: map[string]interface{}{
            "user_id": user.ID,
            "email":   user.Email,
        },
    })
    
    return nil
}
```

## Email Verification

### Send Verification Email

```go
func (s *AuthService) SendVerificationEmail(ctx context.Context, userID uint) error {
    user, err := s.userRepo.FindByID(ctx, userID)
    if err != nil || user == nil {
        return errors.NewNotFound("User not found")
    }
    
    if user.EmailVerifiedAt != nil {
        return errors.NewBadRequest("Email already verified")
    }
    
    // Generate verification token
    verifyToken, err := auth.GenerateResetToken()
    if err != nil {
        return errors.NewInternal("Failed to generate token")
    }
    
    // Save token
    hashedToken, _ := s.hasher.Hash(verifyToken)
    user.VerificationToken = &hashedToken
    expiresAt := time.Now().Add(24 * time.Hour)
    user.VerificationExpiresAt = &expiresAt
    
    if err := s.userRepo.Update(ctx, user); err != nil {
        return errors.NewInternal("Failed to save token")
    }
    
    // Send email
    go s.notificationService.SendEmail(ctx, user.Email, "Verify Email",
        fmt.Sprintf("Verification code: %s", verifyToken))
    
    return nil
}
```

### Verify Email

```go
func (s *AuthService) VerifyEmail(ctx context.Context, email, token string) error {
    user, err := s.userRepo.FindByEmail(ctx, email)
    if err != nil || user == nil {
        return errors.NewBadRequest("Invalid verification request")
    }
    
    if user.EmailVerifiedAt != nil {
        return errors.NewBadRequest("Email already verified")
    }
    
    if user.VerificationToken == nil || user.VerificationExpiresAt == nil {
        return errors.NewBadRequest("Invalid verification token")
    }
    
    if time.Now().After(*user.VerificationExpiresAt) {
        return errors.NewBadRequest("Verification token expired")
    }
    
    if err := s.hasher.Verify(token, *user.VerificationToken); err != nil {
        return errors.NewBadRequest("Invalid verification token")
    }
    
    // Mark as verified
    now := time.Now()
    user.EmailVerifiedAt = &now
    user.VerificationToken = nil
    user.VerificationExpiresAt = nil
    
    return s.userRepo.Update(ctx, user)
}
```

## Profile Management

### Get Profile

```go
func (s *AuthService) GetProfile(ctx context.Context, userID uint) (*User, error) {
    user, err := s.userRepo.FindByID(ctx, userID)
    if err != nil || user == nil {
        return nil, errors.NewNotFound("User not found")
    }
    
    // Don't return password
    user.Password = ""
    return user, nil
}
```

### Update Profile

```go
type UpdateProfileRequest struct {
    Name     string `json:"name" validate:"omitempty,min=2,max=100"`
    Username string `json:"username" validate:"omitempty,min=3,max=50,alphanum"`
    Bio      string `json:"bio" validate:"omitempty,max=500"`
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID uint, req *UpdateProfileRequest) (*User, error) {
    user, err := s.userRepo.FindByID(ctx, userID)
    if err != nil || user == nil {
        return nil, errors.NewNotFound("User not found")
    }
    
    // Check username availability if changing
    if req.Username != "" && req.Username != user.Username {
        existing, _ := s.userRepo.FindByUsername(ctx, req.Username)
        if existing != nil {
            return nil, errors.NewConflict("Username already taken")
        }
        user.Username = req.Username
    }
    
    if req.Name != "" {
        user.Name = req.Name
    }
    
    if req.Bio != "" {
        user.Bio = req.Bio
    }
    
    if err := s.userRepo.Update(ctx, user); err != nil {
        return nil, errors.NewInternal("Failed to update profile")
    }
    
    user.Password = ""
    return user, nil
}
```

## Best Practices

### 1. Use HTTPS Only

Always use HTTPS in production to protect tokens in transit.

### 2. Store Tokens Securely

```javascript
// Client-side: Store in httpOnly cookies or secure storage
// DON'T store in localStorage (vulnerable to XSS)
```

### 3. Implement Rate Limiting

```go
// Limit login attempts
func (c *AuthController) Login(ctx *fiber.Ctx) error {
    email := ctx.FormValue("email")
    
    // Check rate limit
    if c.rateLimiter.IsLimited(email) {
        return ctx.Status(429).JSON(fiber.Map{
            "error": "Too many login attempts. Try again later.",
        })
    }
    
    // Process login...
}
```

### 4. Log Authentication Events

```go
log.Info("User login attempt", logger.Fields{
    "email": email,
    "ip":    ctx.IP(),
})
```

### 5. Validate Strong Passwords

```go
func ValidatePasswordStrength(password string) error {
    if len(password) < 8 {
        return errors.New("Password must be at least 8 characters")
    }
    
    hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
    hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
    hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
    hasSpecial := regexp.MustCompile(`[!@#$%^&*]`).MatchString(password)
    
    if !(hasUpper && hasLower && hasNumber && hasSpecial) {
        return errors.New("Password must contain uppercase, lowercase, number, and special character")
    }
    
    return nil
}
```

## Advanced Patterns

### Multi-Factor Authentication

```go
func (s *AuthService) VerifyMFA(ctx context.Context, userID uint, code string) (bool, error) {
    user, err := s.userRepo.FindByID(ctx, userID)
    if err != nil {
        return false, err
    }
    
    if user.MFASecret == nil {
        return false, errors.NewBadRequest("MFA not enabled")
    }
    
    // Verify TOTP code
    valid := totp.Validate(code, *user.MFASecret)
    return valid, nil
}
```

### Social Login Integration

```go
func (s *AuthService) GoogleLogin(ctx context.Context, googleToken string) (map[string]interface{}, error) {
    // Verify Google token
    profile, err := s.googleClient.VerifyToken(googleToken)
    if err != nil {
        return nil, errors.NewUnauthorized("Invalid Google token")
    }
    
    // Find or create user
    user, _ := s.userRepo.FindByEmail(ctx, profile.Email)
    if user == nil {
        user = &User{
            Email:           profile.Email,
            Name:            profile.Name,
            EmailVerifiedAt: &time.Now(),
            GoogleID:        &profile.ID,
        }
        s.userRepo.Create(ctx, user)
    }
    
    // Generate tokens
    return s.generateTokensForUser(ctx, user)
}
```

## Troubleshooting

### Common Issues

**Issue: Token expired**
```go
// Implement automatic token refresh on client
if (error.code === 'TOKEN_EXPIRED') {
    const newToken = await refreshAccessToken();
    // Retry request with new token
}
```

**Issue: Password hash mismatch**
```go
// Ensure consistent bcrypt cost
hasher := auth.NewPasswordHasher(12) // Use same cost for all
```

## Summary

NeonEx Framework authentication provides:

✅ **Complete auth flow** - Register, Login, Logout  
✅ **JWT tokens** - Access and refresh tokens  
✅ **Password management** - Change, forgot, reset  
✅ **Email verification** - Secure verification flow  
✅ **Profile management** - Get and update user info  
✅ **Production-ready** - Security best practices built-in

For more information:
- [JWT Security](jwt.md)
- [Authorization](authorization.md)
- [Password Hashing](password.md)

# Password Security

NeonEx Framework provides robust password hashing and management through the **PasswordHasher** using bcrypt for secure password storage.

## Table of Contents

- [Overview](#overview)
- [PasswordHasher](#passwordhasher)
- [Hashing Passwords](#hashing-passwords)
- [Verifying Passwords](#verifying-passwords)
- [Password Requirements](#password-requirements)
- [Change Password Flow](#change-password-flow)
- [Reset Password Flow](#reset-password-flow)
- [Best Practices](#best-practices)
- [Advanced Patterns](#advanced-patterns)
- [Troubleshooting](#troubleshooting)

## Overview

NeonEx uses **bcrypt** for password hashing, which:
- Is designed to be slow (resistant to brute force)
- Automatically handles salt generation
- Provides adjustable computational cost
- Is industry-standard and battle-tested

### Why Bcrypt?

✅ **Slow by design** - Protects against brute force attacks  
✅ **Automatic salting** - Each password gets unique salt  
✅ **Future-proof** - Cost factor can increase with hardware  
✅ **Battle-tested** - Used by major platforms

## PasswordHasher

### Structure

```go
// core/pkg/auth/password.go
package auth

import (
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

const (
    DefaultCost = 12  // Recommended for production
    MinCost     = 10  // Minimum secure cost
    MaxCost     = 31  // Maximum allowed cost
)

type PasswordHasher struct {
    cost int
}

func NewPasswordHasher(cost int) *PasswordHasher {
    if cost < MinCost {
        cost = MinCost
    }
    if cost > MaxCost {
        cost = MaxCost
    }
    return &PasswordHasher{cost: cost}
}
```

### Initialization

```go
import "neonexcore/pkg/auth"

// Production (default cost: 12)
hasher := auth.NewPasswordHasher(auth.DefaultCost)

// Custom cost
hasher := auth.NewPasswordHasher(14) // Higher security, slower

// Automatic cost validation
hasher := auth.NewPasswordHasher(8) // Will use MinCost (10) instead
```

## Hashing Passwords

### Basic Hashing

```go
func (h *PasswordHasher) Hash(password string) (string, error) {
    if password == "" {
        return "", fmt.Errorf("password cannot be empty")
    }
    
    hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
    if err != nil {
        return "", fmt.Errorf("failed to hash password: %w", err)
    }
    
    return string(hashedBytes), nil
}
```

### Usage Examples

```go
// Hash a new password
hasher := auth.NewPasswordHasher(auth.DefaultCost)
hashedPassword, err := hasher.Hash("SecurePassword123!")
if err != nil {
    return fmt.Errorf("hashing failed: %w", err)
}

// Store in database
user := &User{
    Email:    "user@example.com",
    Password: hashedPassword, // Store the hash, not plain password
}
db.Create(user)
```

### During User Registration

```go
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*User, error) {
    // Validate password strength first
    if err := ValidatePasswordStrength(req.Password); err != nil {
        return nil, err
    }
    
    // Hash password
    hashedPassword, err := s.hasher.Hash(req.Password)
    if err != nil {
        return nil, errors.NewInternal("Failed to hash password")
    }
    
    // Create user with hashed password
    user := &User{
        Name:     req.Name,
        Email:    req.Email,
        Password: hashedPassword, // Never store plain password
    }
    
    return user, s.userRepo.Create(ctx, user)
}
```

## Verifying Passwords

### Basic Verification

```go
func (h *PasswordHasher) Verify(password, hash string) error {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
```

### During Login

```go
func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
    // Find user
    user, err := s.userRepo.FindByEmail(ctx, email)
    if err != nil || user == nil {
        return nil, errors.NewUnauthorized("Invalid email or password")
    }
    
    // Verify password
    if err := s.hasher.Verify(password, user.Password); err != nil {
        // Log failed attempt
        log.Warn("Failed login attempt", logger.Fields{
            "email": email,
            "ip":    ctx.Value("ip"),
        })
        return nil, errors.NewUnauthorized("Invalid email or password")
    }
    
    // Password verified, generate tokens
    return s.generateAuthTokens(ctx, user)
}
```

### Constant-Time Comparison

Bcrypt's `CompareHashAndPassword` already uses constant-time comparison internally, preventing timing attacks:

```go
// Bcrypt handles this automatically
err := bcrypt.CompareHashAndPassword(hash, password)
if err != nil {
    // Password doesn't match
    // Timing is consistent regardless of how wrong the password is
}
```

## Password Requirements

### Validation Function

```go
import (
    "fmt"
    "regexp"
    "unicode"
)

type PasswordValidator struct {
    MinLength      int
    RequireUpper   bool
    RequireLower   bool
    RequireNumber  bool
    RequireSpecial bool
}

func NewPasswordValidator() *PasswordValidator {
    return &PasswordValidator{
        MinLength:      8,
        RequireUpper:   true,
        RequireLower:   true,
        RequireNumber:  true,
        RequireSpecial: true,
    }
}

func (v *PasswordValidator) Validate(password string) error {
    if len(password) < v.MinLength {
        return fmt.Errorf("password must be at least %d characters", v.MinLength)
    }
    
    if len(password) > 72 {
        // Bcrypt limitation
        return fmt.Errorf("password must be no more than 72 characters")
    }
    
    var (
        hasUpper   bool
        hasLower   bool
        hasNumber  bool
        hasSpecial bool
    )
    
    for _, char := range password {
        switch {
        case unicode.IsUpper(char):
            hasUpper = true
        case unicode.IsLower(char):
            hasLower = true
        case unicode.IsNumber(char):
            hasNumber = true
        case unicode.IsPunct(char) || unicode.IsSymbol(char):
            hasSpecial = true
        }
    }
    
    if v.RequireUpper && !hasUpper {
        return fmt.Errorf("password must contain at least one uppercase letter")
    }
    
    if v.RequireLower && !hasLower {
        return fmt.Errorf("password must contain at least one lowercase letter")
    }
    
    if v.RequireNumber && !hasNumber {
        return fmt.Errorf("password must contain at least one number")
    }
    
    if v.RequireSpecial && !hasSpecial {
        return fmt.Errorf("password must contain at least one special character")
    }
    
    return nil
}
```

### Usage in Registration

```go
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*User, error) {
    // Validate password strength
    validator := auth.NewPasswordValidator()
    if err := validator.Validate(req.Password); err != nil {
        return nil, errors.NewBadRequest(err.Error())
    }
    
    // Continue with registration...
}
```

### Common Weak Passwords Check

```go
var commonPasswords = map[string]bool{
    "password":   true,
    "123456":     true,
    "12345678":   true,
    "qwerty":     true,
    "abc123":     true,
    "password1":  true,
    "123456789": true,
}

func IsCommonPassword(password string) bool {
    return commonPasswords[strings.ToLower(password)]
}

func ValidatePasswordStrength(password string) error {
    if IsCommonPassword(password) {
        return errors.NewBadRequest("This password is too common")
    }
    
    validator := auth.NewPasswordValidator()
    return validator.Validate(password)
}
```

## Change Password Flow

### Change Password Service

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
    
    // 3. Validate new password
    validator := auth.NewPasswordValidator()
    if err := validator.Validate(newPassword); err != nil {
        return errors.NewBadRequest(err.Error())
    }
    
    // 4. Check if new password is different
    if currentPassword == newPassword {
        return errors.NewBadRequest("New password must be different from current password")
    }
    
    // 5. Hash new password
    hashedPassword, err := s.hasher.Hash(newPassword)
    if err != nil {
        return errors.NewInternal("Failed to hash password")
    }
    
    // 6. Update password
    user.Password = hashedPassword
    if err := s.userRepo.Update(ctx, user); err != nil {
        return errors.NewInternal("Failed to update password")
    }
    
    // 7. Log password change
    log.Info("Password changed", logger.Fields{
        "user_id": userID,
        "email":   user.Email,
    })
    
    return nil
}
```

### Change Password Controller

```go
type ChangePasswordRequest struct {
    CurrentPassword string `json:"current_password" validate:"required"`
    NewPassword     string `json:"new_password" validate:"required,min=8"`
    ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

func (c *AuthController) ChangePassword(ctx *fiber.Ctx) error {
    var req ChangePasswordRequest
    if err := ctx.BodyParser(&req); err != nil {
        return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }
    
    // Validate request
    if err := c.validator.Validate(req); err != nil {
        return ctx.Status(400).JSON(fiber.Map{"error": err})
    }
    
    // Get user ID from JWT
    userID := ctx.Locals("user_id").(uint)
    
    // Change password
    err := c.authService.ChangePassword(
        ctx.Context(),
        userID,
        req.CurrentPassword,
        req.NewPassword,
    )
    
    if err != nil {
        return ctx.Status(err.Code).JSON(fiber.Map{"error": err.Message})
    }
    
    return ctx.JSON(fiber.Map{
        "message": "Password changed successfully",
    })
}
```

## Reset Password Flow

### Generate Reset Token

```go
func GenerateResetToken() (string, error) {
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

### Request Password Reset

```go
func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
    // Find user
    user, err := s.userRepo.FindByEmail(ctx, email)
    if err != nil || user == nil {
        // Don't reveal if email exists
        return nil
    }
    
    // Generate reset token
    resetToken, err := auth.GenerateResetToken()
    if err != nil {
        return errors.NewInternal("Failed to generate reset token")
    }
    
    // Hash and save token
    hashedToken, _ := s.hasher.Hash(resetToken)
    user.ResetToken = &hashedToken
    expiresAt := time.Now().Add(1 * time.Hour)
    user.ResetTokenExpiresAt = &expiresAt
    
    if err := s.userRepo.Update(ctx, user); err != nil {
        return errors.NewInternal("Failed to save reset token")
    }
    
    // Send email with plain token (not hashed)
    go s.sendResetEmail(user.Email, resetToken)
    
    return nil
}
```

### Reset Password

```go
func (s *AuthService) ResetPassword(ctx context.Context, email, token, newPassword string) error {
    // Find user
    user, err := s.userRepo.FindByEmail(ctx, email)
    if err != nil || user == nil {
        return errors.NewBadRequest("Invalid reset request")
    }
    
    // Validate token exists and not expired
    if user.ResetToken == nil || user.ResetTokenExpiresAt == nil {
        return errors.NewBadRequest("Invalid or expired reset token")
    }
    
    if time.Now().After(*user.ResetTokenExpiresAt) {
        return errors.NewBadRequest("Reset token has expired")
    }
    
    // Verify token
    if err := s.hasher.Verify(token, *user.ResetToken); err != nil {
        return errors.NewBadRequest("Invalid reset token")
    }
    
    // Validate new password
    if err := ValidatePasswordStrength(newPassword); err != nil {
        return errors.NewBadRequest(err.Error())
    }
    
    // Hash new password
    hashedPassword, err := s.hasher.Hash(newPassword)
    if err != nil {
        return errors.NewInternal("Failed to hash password")
    }
    
    // Update password and clear reset token
    user.Password = hashedPassword
    user.ResetToken = nil
    user.ResetTokenExpiresAt = nil
    
    return s.userRepo.Update(ctx, user)
}
```

## Best Practices

### 1. Use Appropriate Cost Factor

```go
// Development: Lower cost for faster testing
if os.Getenv("APP_ENV") == "development" {
    hasher = auth.NewPasswordHasher(10)
} else {
    // Production: Higher cost for better security
    hasher = auth.NewPasswordHasher(12)
}
```

### 2. Never Log or Display Passwords

```go
// ❌ Bad: Logging password
log.Info("User created", logger.Fields{
    "email":    user.Email,
    "password": password, // NEVER DO THIS!
})

// ✅ Good: Only log non-sensitive data
log.Info("User created", logger.Fields{
    "email":   user.Email,
    "user_id": user.ID,
})
```

### 3. Rate Limit Password Attempts

```go
func (c *AuthController) Login(ctx *fiber.Ctx) error {
    email := ctx.FormValue("email")
    
    // Check rate limit
    attempts, _ := c.cache.Get(ctx.Context(), "login:"+email)
    if attempts != nil && attempts.(int) >= 5 {
        return ctx.Status(429).JSON(fiber.Map{
            "error": "Too many failed attempts. Try again in 15 minutes.",
        })
    }
    
    // Process login...
    if loginFailed {
        // Increment attempts
        c.cache.Increment(ctx.Context(), "login:"+email, 1)
        c.cache.Expire(ctx.Context(), "login:"+email, 15*time.Minute)
    }
}
```

### 4. Implement Password History

```go
type PasswordHistory struct {
    ID        uint
    UserID    uint
    Password  string // Hashed password
    CreatedAt time.Time
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uint, newPassword string) error {
    // Check last 5 passwords
    var history []PasswordHistory
    s.db.Where("user_id = ?", userID).
        Order("created_at DESC").
        Limit(5).
        Find(&history)
    
    for _, h := range history {
        if s.hasher.Verify(newPassword, h.Password) == nil {
            return errors.NewBadRequest("Cannot reuse recent passwords")
        }
    }
    
    // Continue with password change...
}
```

### 5. Enforce Password Expiration

```go
func (s *AuthService) CheckPasswordExpiration(user *User) error {
    if user.PasswordChangedAt == nil {
        return nil
    }
    
    // Require password change after 90 days
    expiryDate := user.PasswordChangedAt.Add(90 * 24 * time.Hour)
    if time.Now().After(expiryDate) {
        return errors.NewBadRequest("Password has expired. Please change your password.")
    }
    
    return nil
}
```

## Advanced Patterns

### Password Strength Meter

```go
type PasswordStrength string

const (
    Weak       PasswordStrength = "weak"
    Fair       PasswordStrength = "fair"
    Good       PasswordStrength = "good"
    Strong     PasswordStrength = "strong"
    VeryStrong PasswordStrength = "very_strong"
)

func CalculatePasswordStrength(password string) PasswordStrength {
    score := 0
    
    // Length
    if len(password) >= 8 {
        score++
    }
    if len(password) >= 12 {
        score++
    }
    if len(password) >= 16 {
        score++
    }
    
    // Character variety
    if regexp.MustCompile(`[a-z]`).MatchString(password) {
        score++
    }
    if regexp.MustCompile(`[A-Z]`).MatchString(password) {
        score++
    }
    if regexp.MustCompile(`[0-9]`).MatchString(password) {
        score++
    }
    if regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password) {
        score++
    }
    
    switch {
    case score <= 2:
        return Weak
    case score <= 4:
        return Fair
    case score <= 6:
        return Good
    case score <= 8:
        return Strong
    default:
        return VeryStrong
    }
}
```

### Rehash on Login

```go
func (h *PasswordHasher) NeedsRehash(hash string) bool {
    cost, err := bcrypt.Cost([]byte(hash))
    if err != nil {
        return true
    }
    return cost != h.cost
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
    user, _ := s.userRepo.FindByEmail(ctx, email)
    
    // Verify password
    if err := s.hasher.Verify(password, user.Password); err != nil {
        return nil, errors.NewUnauthorized("Invalid credentials")
    }
    
    // Rehash if needed (cost factor changed)
    if s.hasher.NeedsRehash(user.Password) {
        newHash, _ := s.hasher.Hash(password)
        user.Password = newHash
        s.userRepo.Update(ctx, user)
    }
    
    return s.generateTokens(user), nil
}
```

## Troubleshooting

### Common Issues

**Issue: "crypto/bcrypt: hashedPassword is not the hash of the given password"**

```go
// Ensure you're passing parameters in correct order
// Correct: (password, hash)
err := hasher.Verify(plainPassword, hashedPassword)

// Wrong: (hash, password)
err := hasher.Verify(hashedPassword, plainPassword) // ❌
```

**Issue: Hashing takes too long**

```go
// Reduce cost factor for development
if os.Getenv("APP_ENV") == "development" {
    hasher = auth.NewPasswordHasher(10) // Faster
} else {
    hasher = auth.NewPasswordHasher(12) // Secure
}
```

**Issue: Password too long error**

```go
// Bcrypt has 72 character limit
if len(password) > 72 {
    return errors.NewBadRequest("Password too long (max 72 characters)")
}
```

## Summary

NeonEx Framework password security provides:

✅ **Bcrypt hashing** - Industry-standard secure hashing  
✅ **Configurable cost** - Adjust security vs performance  
✅ **Password validation** - Enforce strong passwords  
✅ **Reset flow** - Secure password reset with tokens  
✅ **Change flow** - Verify current password before change  
✅ **Best practices** - Built-in security patterns

For more information:
- [Authentication](authentication.md)
- [Authorization](authorization.md)
- [JWT Security](jwt.md)

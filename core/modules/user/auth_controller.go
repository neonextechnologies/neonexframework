package user

import (
	"context"

	"neonexcore/pkg/auth"
	"neonexcore/pkg/errors"
	"neonexcore/pkg/validation"

	"github.com/gofiber/fiber/v2"
)

// AuthController handles authentication endpoints
type AuthController struct {
	authService *AuthService
}

// NewAuthController creates a new auth controller
func NewAuthController(authService *AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// Login handles user login
// POST /api/v1/auth/login
func (ctrl *AuthController) Login(c *fiber.Ctx) error {
	var req validation.LoginRequest
	
	// Validate request body
	if err := validation.ValidateBody(c, &req); err != nil {
		return err
	}

	ctx := context.Background()
	
	// Authenticate user
	result, err := ctrl.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Login successful",
		"data":    result,
	})
}

// Register handles user registration
// POST /api/v1/auth/register
func (ctrl *AuthController) Register(c *fiber.Ctx) error {
	var req validation.RegisterRequest
	
	// Validate request body
	if err := validation.ValidateBody(c, &req); err != nil {
		return err
	}

	ctx := context.Background()
	
	// Register user
	user, err := ctrl.authService.Register(ctx, &req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Registration successful",
		"data": fiber.Map{
			"id":       user.ID,
			"name":     user.Name,
			"email":    user.Email,
			"username": user.Username,
		},
	})
}

// RefreshToken handles token refresh
// POST /api/v1/auth/refresh
func (ctrl *AuthController) RefreshToken(c *fiber.Ctx) error {
	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	var req RefreshRequest
	if err := validation.ValidateBody(c, &req); err != nil {
		return err
	}

	ctx := context.Background()
	result, err := ctrl.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Token refreshed successfully",
		"data":    result,
	})
}

// Logout handles user logout
// POST /api/v1/auth/logout
func (ctrl *AuthController) Logout(c *fiber.Ctx) error {
	// In JWT, logout is typically handled client-side by removing the token
	// Here we can add token to blacklist if needed (future enhancement)
	
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Logout successful",
	})
}

// GetProfile gets current user profile
// GET /api/v1/auth/profile
func (ctrl *AuthController) GetProfile(c *fiber.Ctx) error {
	userID, ok := auth.GetUserID(c)
	if !ok {
		return errors.NewUnauthorized("User not authenticated")
	}

	ctx := context.Background()
	user, err := ctrl.authService.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return errors.NewNotFound("User not found")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"id":                user.ID,
			"name":              user.Name,
			"email":             user.Email,
			"username":          user.Username,
			"age":               user.Age,
			"is_active":         user.IsActive,
			"is_email_verified": user.IsEmailVerified,
			"last_login_at":     user.LastLoginAt,
			"created_at":        user.CreatedAt,
		},
	})
}

// UpdateProfile updates current user profile
// PUT /api/v1/auth/profile
func (ctrl *AuthController) UpdateProfile(c *fiber.Ctx) error {
	userID, ok := auth.GetUserID(c)
	if !ok {
		return errors.NewUnauthorized("User not authenticated")
	}

	var req validation.UpdateProfileRequest
	if err := validation.ValidateBody(c, &req); err != nil {
		return err
	}

	ctx := context.Background()
	user, err := ctrl.authService.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return errors.NewNotFound("User not found")
	}

	// Update fields if provided
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		// Check if email is already taken by another user
		existing, _ := ctrl.authService.userRepo.FindByEmail(ctx, req.Email)
		if existing != nil && existing.ID != userID {
			return errors.NewConflict("Email already in use")
		}
		user.Email = req.Email
	}

	if err := ctrl.authService.userRepo.Update(ctx, user); err != nil {
		return errors.NewInternal("Failed to update profile")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Profile updated successfully",
		"data": fiber.Map{
			"id":       user.ID,
			"name":     user.Name,
			"email":    user.Email,
			"username": user.Username,
		},
	})
}

// ChangePassword changes user password
// POST /api/v1/auth/change-password
func (ctrl *AuthController) ChangePassword(c *fiber.Ctx) error {
	userID, ok := auth.GetUserID(c)
	if !ok {
		return errors.NewUnauthorized("User not authenticated")
	}

	var req validation.ChangePasswordRequest
	if err := validation.ValidateBody(c, &req); err != nil {
		return err
	}

	ctx := context.Background()
	err := ctrl.authService.ChangePassword(ctx, userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Password changed successfully",
	})
}

// GenerateAPIKey generates API key for user
// POST /api/v1/auth/api-key
func (ctrl *AuthController) GenerateAPIKey(c *fiber.Ctx) error {
	userID, ok := auth.GetUserID(c)
	if !ok {
		return errors.NewUnauthorized("User not authenticated")
	}

	ctx := context.Background()
	apiKey, err := ctrl.authService.GenerateAPIKey(ctx, userID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "API key generated successfully",
		"data": fiber.Map{
			"api_key": apiKey,
		},
	})
}

// ForgotPassword initiates password reset
// POST /api/v1/auth/forgot-password
func (ctrl *AuthController) ForgotPassword(c *fiber.Ctx) error {
	type ForgotPasswordRequest struct {
		Email string `json:"email" validate:"required,email"`
	}

	var req ForgotPasswordRequest
	if err := validation.ValidateBody(c, &req); err != nil {
		return err
	}

	ctx := context.Background()
	user, err := ctrl.authService.userRepo.FindByEmail(ctx, req.Email)
	if err != nil || user == nil {
		// Don't reveal if email exists or not (security)
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "If the email exists, a password reset link has been sent",
		})
	}

	// Generate reset token
	resetToken, err := auth.GenerateResetToken()
	if err != nil {
		return errors.NewInternal("Failed to generate reset token")
	}

	// TODO: Save reset token to database and send email
	// For now, just return success (will implement email in notification system)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "If the email exists, a password reset link has been sent",
		"debug": fiber.Map{
			"reset_token": resetToken, // Remove in production
		},
	})
}

// ResetPassword resets password with token
// POST /api/v1/auth/reset-password
func (ctrl *AuthController) ResetPassword(c *fiber.Ctx) error {
	type ResetPasswordRequest struct {
		Token       string `json:"token" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=8,max=100"`
	}

	var req ResetPasswordRequest
	if err := validation.ValidateBody(c, &req); err != nil {
		return err
	}

	// TODO: Implement token validation and password reset
	// For now, return not implemented

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"success": false,
		"message": "Password reset not yet implemented",
	})
}

// VerifyEmail verifies user email
// GET /api/v1/auth/verify-email/:token
func (ctrl *AuthController) VerifyEmail(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return errors.NewBadRequest("Token is required")
	}

	// TODO: Implement email verification
	// For now, return not implemented

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"success": false,
		"message": "Email verification not yet implemented",
	})
}

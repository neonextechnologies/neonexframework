package user

import (
	"context"
	"time"

	"neonexcore/pkg/auth"
	"neonexcore/pkg/errors"
	"neonexcore/pkg/events"
	"neonexcore/pkg/rbac"
	"neonexcore/pkg/validation"
)

// AuthService handles authentication logic
type AuthService struct {
	userRepo    *UserRepository
	jwtManager  *auth.JWTManager
	hasher      *auth.PasswordHasher
	rbacManager *rbac.Manager
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo *UserRepository,
	jwtManager *auth.JWTManager,
	hasher *auth.PasswordHasher,
	rbacManager *rbac.Manager,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		jwtManager:  jwtManager,
		hasher:      hasher,
		rbacManager: rbacManager,
	}
}

// Login authenticates a user
func (s *AuthService) Login(ctx context.Context, email, password string) (map[string]interface{}, error) {
	// Find user
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, errors.New(errors.ErrCodeInvalidCredentials, "Invalid email or password", 401)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New(errors.ErrCodeAccountDisabled, "Account is disabled", 403)
	}

	// Verify password
	if err := s.hasher.Verify(password, user.Password); err != nil {
		return nil, errors.New(errors.ErrCodeInvalidCredentials, "Invalid email or password", 401)
	}

	// Get user roles and permissions
	roles, _ := s.rbacManager.GetUserRoles(ctx, user.ID)
	permissions, _ := s.rbacManager.GetUserPermissions(ctx, user.ID)

	// Extract role names
	var roleNames []string
	for _, role := range roles {
		roleNames = append(roleNames, role.Slug)
	}

	// Extract permission slugs
	var permissionSlugs []string
	for _, perm := range permissions {
		permissionSlugs = append(permissionSlugs, perm.Slug)
	}

	// Generate tokens
	primaryRole := "user"
	if len(roleNames) > 0 {
		primaryRole = roleNames[0]
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, primaryRole, permissionSlugs)
	if err != nil {
		return nil, errors.NewInternal("Failed to generate access token")
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, errors.NewInternal("Failed to generate refresh token")
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	s.userRepo.Update(ctx, user)

	// Dispatch login event
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
		"expires_in":    900, // 15 minutes
		"user": map[string]interface{}{
			"id":       user.ID,
			"name":     user.Name,
			"email":    user.Email,
			"username": user.Username,
			"roles":    roleNames,
		},
	}, nil
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *validation.RegisterRequest) (*User, error) {
	// Check if email exists
	existing, _ := s.userRepo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.NewConflict("Email already exists")
	}

	// Check if username exists
	existing, _ = s.userRepo.FindByUsername(ctx, req.Username)
	if existing != nil {
		return nil, errors.NewConflict("Username already exists")
	}

	// Hash password
	hashedPassword, err := s.hasher.Hash(req.Password)
	if err != nil {
		return nil, errors.NewInternal("Failed to hash password")
	}

	// Create user
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

	// Assign default user role
	role, _ := s.rbacManager.GetRoleBySlug(ctx, "user")
	if role != nil {
		s.rbacManager.AssignRole(ctx, user.ID, role.ID)
	}

	// Dispatch user created event
	events.DispatchAsync(ctx, events.Event{
		Name: events.EventUserCreated,
		Data: map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
		},
	})

	return user, nil
}

// RefreshToken refreshes access token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (map[string]interface{}, error) {
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

// ChangePassword changes user password
func (s *AuthService) ChangePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return errors.NewNotFound("User not found")
	}

	// Verify current password
	if err := s.hasher.Verify(currentPassword, user.Password); err != nil {
		return errors.NewBadRequest("Current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := s.hasher.Hash(newPassword)
	if err != nil {
		return errors.NewInternal("Failed to hash password")
	}

	user.Password = hashedPassword
	return s.userRepo.Update(ctx, user)
}

// GenerateAPIKey generates API key for user
func (s *AuthService) GenerateAPIKey(ctx context.Context, userID uint) (string, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return "", errors.NewNotFound("User not found")
	}

	apiKey, err := auth.GenerateAPIKey()
	if err != nil {
		return "", errors.NewInternal("Failed to generate API key")
	}

	user.APIKey = &apiKey
	if err := s.userRepo.Update(ctx, user); err != nil {
		return "", errors.NewInternal("Failed to save API key")
	}

	return apiKey, nil
}

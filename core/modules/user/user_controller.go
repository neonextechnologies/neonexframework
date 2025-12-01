package user

import (
	"context"
	"strconv"

	"neonexcore/pkg/auth"
	"neonexcore/pkg/errors"
	"neonexcore/pkg/events"
	"neonexcore/pkg/rbac"

	"github.com/gofiber/fiber/v2"
)

// UserController handles user CRUD operations
type UserController struct {
	service     *UserService
	rbacManager *rbac.Manager
}

// NewUserController creates a new user controller
func NewUserController(service *UserService, rbacManager *rbac.Manager) *UserController {
	return &UserController{
		service:     service,
		rbacManager: rbacManager,
	}
}

// GetAll returns all users with pagination
// GET /api/v1/users?page=1&limit=10
func (ctrl *UserController) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	ctx := context.Background()
	users, total, err := ctrl.service.repo.Paginate(ctx, page, limit)
	if err != nil {
		return errors.NewInternal("Failed to fetch users")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    users,
		"meta": fiber.Map{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetByID returns a user by ID
// GET /api/v1/users/:id
func (ctrl *UserController) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return errors.NewBadRequest("Invalid user ID")
	}

	ctx := context.Background()
	user, err := ctrl.service.repo.FindByID(ctx, uint(id))
	if err != nil || user == nil {
		return errors.NewNotFound("User not found")
	}

	// Get user roles
	roles, _ := ctrl.rbacManager.GetUserRoles(ctx, user.ID)
	
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
			"updated_at":        user.UpdatedAt,
			"roles":             roles,
		},
	})
}

// Create creates a new user (admin only)
// POST /api/v1/users
func (ctrl *UserController) Create(c *fiber.Ctx) error {
	type CreateUserRequest struct {
		Name     string `json:"name" validate:"required,min=2,max=100"`
		Email    string `json:"email" validate:"required,email"`
		Username string `json:"username" validate:"required,username"`
		Password string `json:"password" validate:"required,min=8,max=100"`
		Age      int    `json:"age" validate:"omitempty,gte=0,lte=150"`
		IsActive bool   `json:"is_active"`
	}

	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.NewBadRequest("Invalid request body")
	}

	// Validate
	validator := &validation.Validator{}
	if errs := validator.Validate(&req); errs != nil {
		details := make(map[string]interface{})
		for field, msg := range errs {
			details[field] = msg
		}
		return errors.NewValidationError("Validation failed", details)
	}

	ctx := context.Background()

	// Check if email exists
	existing, _ := ctrl.service.repo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return errors.NewConflict("Email already exists")
	}

	// Check if username exists
	existing, _ = ctrl.service.repo.FindByUsername(ctx, req.Username)
	if existing != nil {
		return errors.NewConflict("Username already exists")
	}

	// Hash password
	hasher := auth.NewPasswordHasher(12)
	hashedPassword, err := hasher.Hash(req.Password)
	if err != nil {
		return errors.NewInternal("Failed to hash password")
	}

	// Create user
	user := &User{
		Name:     req.Name,
		Email:    req.Email,
		Username: req.Username,
		Password: hashedPassword,
		Age:      req.Age,
		IsActive: req.IsActive,
		Active:   req.IsActive,
	}

	if err := ctrl.service.repo.Create(ctx, user); err != nil {
		return errors.NewInternal("Failed to create user")
	}

	// Dispatch event
	events.DispatchAsync(ctx, events.Event{
		Name: events.EventUserCreated,
		Data: map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
		},
	})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "User created successfully",
		"data": fiber.Map{
			"id":       user.ID,
			"name":     user.Name,
			"email":    user.Email,
			"username": user.Username,
		},
	})
}

// Update updates a user
// PUT /api/v1/users/:id
func (ctrl *UserController) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return errors.NewBadRequest("Invalid user ID")
	}

	type UpdateUserRequest struct {
		Name     string `json:"name" validate:"omitempty,min=2,max=100"`
		Email    string `json:"email" validate:"omitempty,email"`
		Age      int    `json:"age" validate:"omitempty,gte=0,lte=150"`
		IsActive *bool  `json:"is_active"`
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.NewBadRequest("Invalid request body")
	}

	ctx := context.Background()
	user, err := ctrl.service.repo.FindByID(ctx, uint(id))
	if err != nil || user == nil {
		return errors.NewNotFound("User not found")
	}

	// Update fields if provided
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		// Check if email is already taken by another user
		existing, _ := ctrl.service.repo.FindByEmail(ctx, req.Email)
		if existing != nil && existing.ID != user.ID {
			return errors.NewConflict("Email already in use")
		}
		user.Email = req.Email
	}
	if req.Age > 0 {
		user.Age = req.Age
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
		user.Active = *req.IsActive
	}

	if err := ctrl.service.repo.Update(ctx, user); err != nil {
		return errors.NewInternal("Failed to update user")
	}

	// Dispatch event
	events.DispatchAsync(ctx, events.Event{
		Name: events.EventUserUpdated,
		Data: map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
		},
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User updated successfully",
		"data": fiber.Map{
			"id":        user.ID,
			"name":      user.Name,
			"email":     user.Email,
			"username":  user.Username,
			"is_active": user.IsActive,
		},
	})
}

// Delete deletes a user (soft delete)
// DELETE /api/v1/users/:id
func (ctrl *UserController) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return errors.NewBadRequest("Invalid user ID")
	}

	// Prevent deleting yourself
	currentUserID, ok := auth.GetUserID(c)
	if ok && currentUserID == uint(id) {
		return errors.NewBadRequest("Cannot delete your own account")
	}

	ctx := context.Background()
	user, err := ctrl.service.repo.FindByID(ctx, uint(id))
	if err != nil || user == nil {
		return errors.NewNotFound("User not found")
	}

	if err := ctrl.service.repo.Delete(ctx, uint(id)); err != nil {
		return errors.NewInternal("Failed to delete user")
	}

	// Dispatch event
	events.DispatchAsync(ctx, events.Event{
		Name: events.EventUserDeleted,
		Data: map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
		},
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User deleted successfully",
	})
}

// Search searches users by name or email
// GET /api/v1/users/search?q=john
func (ctrl *UserController) Search(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return errors.NewBadRequest("Search query is required")
	}

	ctx := context.Background()
	users, err := ctrl.service.repo.Search(ctx, query)
	if err != nil {
		return errors.NewInternal("Failed to search users")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    users,
		"meta": fiber.Map{
			"query": query,
			"count": len(users),
		},
	})
}

// AssignRole assigns a role to a user
// POST /api/v1/users/:id/roles
func (ctrl *UserController) AssignRole(c *fiber.Ctx) error {
	userID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return errors.NewBadRequest("Invalid user ID")
	}

	type AssignRoleRequest struct {
		RoleID uint `json:"role_id" validate:"required"`
	}

	var req AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.NewBadRequest("Invalid request body")
	}

	ctx := context.Background()
	
	// Check if user exists
	user, err := ctrl.service.repo.FindByID(ctx, uint(userID))
	if err != nil || user == nil {
		return errors.NewNotFound("User not found")
	}

	// Assign role
	if err := ctrl.rbacManager.AssignRole(ctx, uint(userID), req.RoleID); err != nil {
		return errors.NewInternal("Failed to assign role")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Role assigned successfully",
	})
}

// RemoveRole removes a role from a user
// DELETE /api/v1/users/:id/roles/:roleId
func (ctrl *UserController) RemoveRole(c *fiber.Ctx) error {
	userID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return errors.NewBadRequest("Invalid user ID")
	}

	roleID, err := strconv.ParseUint(c.Params("roleId"), 10, 32)
	if err != nil {
		return errors.NewBadRequest("Invalid role ID")
	}

	ctx := context.Background()
	if err := ctrl.rbacManager.RemoveRole(ctx, uint(userID), uint(roleID)); err != nil {
		return errors.NewInternal("Failed to remove role")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Role removed successfully",
	})
}

// GetUserRoles gets all roles for a user
// GET /api/v1/users/:id/roles
func (ctrl *UserController) GetUserRoles(c *fiber.Ctx) error {
	userID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return errors.NewBadRequest("Invalid user ID")
	}

	ctx := context.Background()
	roles, err := ctrl.rbacManager.GetUserRoles(ctx, uint(userID))
	if err != nil {
		return errors.NewInternal("Failed to fetch user roles")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    roles,
	})
}

// GetUserPermissions gets all permissions for a user
// GET /api/v1/users/:id/permissions
func (ctrl *UserController) GetUserPermissions(c *fiber.Ctx) error {
	userID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return errors.NewBadRequest("Invalid user ID")
	}

	ctx := context.Background()
	permissions, err := ctrl.rbacManager.GetUserPermissions(ctx, uint(userID))
	if err != nil {
		return errors.NewInternal("Failed to fetch user permissions")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    permissions,
	})
}

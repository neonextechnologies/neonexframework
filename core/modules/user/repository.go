package user

import (
	"context"

	"neonexcore/pkg/database"

	"gorm.io/gorm"
)

// UserRepository handles user data operations
type UserRepository struct {
	*database.BaseRepository[User]
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		BaseRepository: database.NewBaseRepository[User](db),
	}
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	return r.FindOne(ctx, "email = ?", email)
}

// FindActiveUsers finds all active users
func (r *UserRepository) FindActiveUsers(ctx context.Context) ([]*User, error) {
	return r.FindByCondition(ctx, "active = ?", true)
}

// UpdateUserStatus updates user active status
func (r *UserRepository) UpdateUserStatus(ctx context.Context, userID uint, active bool) error {
	return r.GetDB().WithContext(ctx).Model(&User{}).Where("id = ?", userID).Update("active", active).Error
}

// SearchUsers searches users by name or email
func (r *UserRepository) SearchUsers(ctx context.Context, keyword string) ([]*User, error) {
	pattern := "%" + keyword + "%"
	return r.FindByCondition(ctx, "name LIKE ? OR email LIKE ?", pattern, pattern)
}

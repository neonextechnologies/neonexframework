package user

import (
	"context"

	"neonexcore/pkg/database"

	"gorm.io/gorm"
)

type UserRepository struct {
	*database.BaseRepository[User]
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		BaseRepository: database.NewBaseRepository[User](db),
	}
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	return r.FindOne(ctx, "email = ?", email)
}

// FindByUsername finds a user by username
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	return r.FindOne(ctx, "username = ?", username)
}

// FindByAPIKey finds a user by API key
func (r *UserRepository) FindByAPIKey(ctx context.Context, apiKey string) (*User, error) {
	return r.FindOne(ctx, "api_key = ?", apiKey)
}

// Search searches users by name or email
func (r *UserRepository) Search(ctx context.Context, query string) ([]*User, error) {
	return r.FindByCondition(ctx, "name LIKE ? OR email LIKE ?", "%"+query+"%", "%"+query+"%")
}

// GetActiveUsers gets all active users
func (r *UserRepository) GetActiveUsers(ctx context.Context) ([]*User, error) {
	return r.FindByCondition(ctx, "is_active = ?", true)
}

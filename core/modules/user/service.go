package user

import (
	"context"
	"fmt"

	"neonexcore/pkg/database"

	"gorm.io/gorm"
)

type UserService struct {
	repo      *UserRepository
	txManager *database.TxManager
}

func NewUserService(repo *UserRepository, txManager *database.TxManager) *UserService {
	return &UserService{
		repo:      repo,
		txManager: txManager,
	}
}

// GetAllUsers retrieves all users
func (s *UserService) GetAllUsers(ctx context.Context) ([]*User, error) {
	return s.repo.FindAll(ctx)
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, id uint) (*User, error) {
	return s.repo.FindByID(ctx, id)
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, user *User) error {
	return s.repo.Create(ctx, user)
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(ctx context.Context, user *User) error {
	return s.repo.Update(ctx, user)
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.repo.FindByEmail(ctx, email)
}

// SearchUsers searches users by keyword
func (s *UserService) SearchUsers(ctx context.Context, keyword string) ([]*User, error) {
	return s.repo.SearchUsers(ctx, keyword)
}

// CreateUserWithTransaction demonstrates transaction usage
func (s *UserService) CreateUserWithTransaction(ctx context.Context, users []*User) error {
	return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		repo := s.repo.WithTx(tx)

		for _, user := range users {
			if err := repo.Create(ctx, user); err != nil {
				return fmt.Errorf("failed to create user %s: %w", user.Email, err)
			}
		}

		return nil
	})
}

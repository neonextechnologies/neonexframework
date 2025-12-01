package user

import (
	"context"
	"fmt"

	"neonexcore/pkg/auth"

	"gorm.io/gorm"
)

// UserSeeder seeds initial user data
type UserSeeder struct {
	db *gorm.DB
}

func NewUserSeeder(db *gorm.DB) *UserSeeder {
	return &UserSeeder{db: db}
}

func (s *UserSeeder) Name() string {
	return "UserSeeder"
}

// Run implements the Seeder interface
func (s *UserSeeder) Run(ctx context.Context) error {
	// Check if users already exist
	var count int64
	if err := s.db.Model(&User{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		fmt.Println("  ⏭️  Users already seeded, skipping...")
		return nil
	}

	hasher := auth.NewPasswordHasher()

	// Hash passwords
	adminPass, _ := hasher.Hash("admin123")
	userPass, _ := hasher.Hash("user123")

	users := []User{
		{
			Name:     "Admin User",
			Email:    "admin@neonex.local",
			Username: "admin",
			Password: adminPass,
			IsActive: true,
		},
		{
			Name:     "Test User",
			Email:    "user@neonex.local",
			Username: "testuser",
			Password: userPass,
			IsActive: true,
		},
		{
			Name:     "Alice Johnson",
			Email:    "alice@example.com",
			Username: "alice",
			Password: userPass,
			IsActive: true,
		},
		{
			Name:     "Bob Smith",
			Email:    "bob@example.com",
			Username: "bob",
			Password: userPass,
			IsActive: true,
		},
	}

	result := s.db.Create(&users)
	if result.Error != nil {
		return result.Error
	}

	fmt.Printf("  ✓ Seeded %d users\n", len(users))
	return nil
}

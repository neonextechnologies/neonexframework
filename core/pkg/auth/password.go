package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	DefaultCost = 12
	MinCost     = 10
	MaxCost     = 31
)

// PasswordHasher handles password hashing operations
type PasswordHasher struct {
	cost int
}

// NewPasswordHasher creates a new password hasher
func NewPasswordHasher(cost int) *PasswordHasher {
	if cost < MinCost {
		cost = MinCost
	}
	if cost > MaxCost {
		cost = MaxCost
	}
	return &PasswordHasher{cost: cost}
}

// Hash hashes a password using bcrypt
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

// Verify verifies a password against a hash
func (h *PasswordHasher) Verify(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// NeedsRehash checks if password hash needs to be regenerated
func (h *PasswordHasher) NeedsRehash(hash string) bool {
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		return true
	}
	return cost != h.cost
}

// GenerateRandomToken generates a random token
func GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateResetToken generates a password reset token
func GenerateResetToken() (string, error) {
	return GenerateRandomToken(32)
}

// GenerateAPIKey generates an API key
func GenerateAPIKey() (string, error) {
	return GenerateRandomToken(32)
}

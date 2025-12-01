package user

import (
	"time"

	"neonexcore/internal/config"
	"neonexcore/internal/core"
	"neonexcore/pkg/auth"
	"neonexcore/pkg/database"
	"neonexcore/pkg/rbac"
)

func (m *UserModule) RegisterServices(c *core.Container) {
	// ==================== Database & Transaction ====================
	
	// Register Transaction Manager
	c.Provide(func() *database.TxManager {
		db := config.DB.GetDB()
		return database.NewTxManager(db)
	}, core.Singleton)

	// ==================== Authentication & Security ====================
	
	// Register JWT Manager
	c.Provide(func() *auth.JWTManager {
		return auth.NewJWTManager(&auth.JWTConfig{
			SecretKey:     "your-secret-key-change-in-production", // TODO: Move to config
			AccessExpiry:  15 * time.Minute,
			RefreshExpiry: 7 * 24 * time.Hour,
			Issuer:        "NeonexCore",
			Algorithm:     "HS256",
		})
	}, core.Singleton)

	// Register Password Hasher
	c.Provide(func() *auth.PasswordHasher {
		return auth.NewPasswordHasher(12) // bcrypt cost
	}, core.Singleton)

	// ==================== RBAC ====================
	
	// Register RBAC Manager
	c.Provide(func() *rbac.Manager {
		db := config.DB.GetDB()
		return rbac.NewManager(db)
	}, core.Singleton)

	// ==================== Repositories ====================
	
	// Register User Repository
	c.Provide(func() *UserRepository {
		db := config.DB.GetDB()
		return NewUserRepository(db)
	}, core.Singleton)

	// ==================== Services ====================
	
	// Register User Service
	c.Provide(func() *UserService {
		repo := core.Resolve[*UserRepository](c)
		txManager := core.Resolve[*database.TxManager](c)
		return NewUserService(repo, txManager)
	}, core.Singleton)

	// Register Auth Service
	c.Provide(func() *AuthService {
		userRepo := core.Resolve[*UserRepository](c)
		jwtManager := core.Resolve[*auth.JWTManager](c)
		hasher := core.Resolve[*auth.PasswordHasher](c)
		rbacManager := core.Resolve[*rbac.Manager](c)
		return NewAuthService(userRepo, jwtManager, hasher, rbacManager)
	}, core.Singleton)

	// ==================== Controllers ====================
	
	// Register Auth Controller
	c.Provide(func() *AuthController {
		authService := core.Resolve[*AuthService](c)
		return NewAuthController(authService)
	}, core.Transient)

	// Register User Controller
	c.Provide(func() *UserController {
		service := core.Resolve[*UserService](c)
		rbacManager := core.Resolve[*rbac.Manager](c)
		return NewUserController(service, rbacManager)
	}, core.Transient)
}

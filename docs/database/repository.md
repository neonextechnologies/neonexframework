# Repository Pattern

NeonEx Framework provides a powerful repository pattern implementation with a generic `BaseRepository` that offers CRUD operations, pagination, and query building. This promotes clean architecture and testable code.

## Table of Contents

- [Overview](#overview)
- [BaseRepository](#baserepository)
- [CRUD Operations](#crud-operations)
- [Custom Repositories](#custom-repositories)
- [Query Building](#query-building)
- [Transactions](#transactions)
- [Best Practices](#best-practices)

## Overview

The Repository pattern abstracts data access logic and provides a clean API for working with database entities:

```go
import "neonexcore/pkg/database"

// Generic repository interface
type Repository[T any] interface {
    Create(ctx context.Context, entity *T) error
    Update(ctx context.Context, entity *T) error
    Delete(ctx context.Context, id interface{}) error
    FindByID(ctx context.Context, id interface{}) (*T, error)
    FindAll(ctx context.Context) ([]*T, error)
    Paginate(ctx context.Context, page, pageSize int) ([]*T, int64, error)
}
```

## BaseRepository

### Creating a Repository

```go
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
```

### Using BaseRepository Methods

```go
func main() {
    // Initialize repository
    userRepo := NewUserRepository(db)
    ctx := context.Background()
    
    // Create
    user := &User{Name: "John", Email: "john@example.com"}
    err := userRepo.Create(ctx, user)
    
    // Find by ID
    user, err := userRepo.FindByID(ctx, 1)
    
    // Find all
    users, err := userRepo.FindAll(ctx)
    
    // Update
    user.Name = "John Doe"
    err = userRepo.Update(ctx, user)
    
    // Delete
    err = userRepo.Delete(ctx, 1)
    
    // Paginate
    users, total, err := userRepo.Paginate(ctx, 1, 10)
}
```

## CRUD Operations

### Create

```go
func (s *UserService) CreateUser(req CreateUserRequest) (*User, error) {
    ctx := context.Background()
    
    user := &User{
        Name:     req.Name,
        Email:    req.Email,
        Password: req.Password,
    }
    
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    return user, nil
}
```

### Batch Create

```go
func (s *UserService) CreateBulk(users []*User) error {
    ctx := context.Background()
    return s.repo.CreateBatch(ctx, users)
}

// Usage
users := []*User{
    {Name: "User 1", Email: "user1@example.com"},
    {Name: "User 2", Email: "user2@example.com"},
}
err := userService.CreateBulk(users)
```

### Read Operations

```go
// Find by ID
func (s *UserService) GetByID(id uint) (*User, error) {
    ctx := context.Background()
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, errors.New("user not found")
    }
    return user, nil
}

// Find all
func (s *UserService) GetAll() ([]*User, error) {
    ctx := context.Background()
    return s.repo.FindAll(ctx)
}

// Find by condition
func (s *UserService) FindActive() ([]*User, error) {
    ctx := context.Background()
    return s.repo.FindByCondition(ctx, "active = ?", true)
}

// Find one by condition
func (s *UserService) FindByEmail(email string) (*User, error) {
    ctx := context.Background()
    return s.repo.FindOne(ctx, "email = ?", email)
}
```

### Update

```go
func (s *UserService) UpdateUser(id uint, req UpdateUserRequest) (*User, error) {
    ctx := context.Background()
    
    user, err := s.repo.FindByID(ctx, id)
    if err != nil || user == nil {
        return nil, errors.New("user not found")
    }
    
    user.Name = req.Name
    user.Email = req.Email
    
    if err := s.repo.Update(ctx, user); err != nil {
        return nil, fmt.Errorf("failed to update user: %w", err)
    }
    
    return user, nil
}
```

### Delete

```go
func (s *UserService) DeleteUser(id uint) error {
    ctx := context.Background()
    return s.repo.Delete(ctx, id)
}
```

### Count

```go
func (s *UserService) CountActive() (int64, error) {
    ctx := context.Background()
    return s.repo.Count(ctx, "active = ?", true)
}

func (s *UserService) CountAll() (int64, error) {
    ctx := context.Background()
    return s.repo.Count(ctx, "")
}
```

### Pagination

```go
func (s *UserService) GetUsers(page, limit int) ([]*User, int64, error) {
    ctx := context.Background()
    return s.repo.Paginate(ctx, page, limit)
}

// In controller
func listUsers(c *fiber.Ctx) error {
    pagination := api.GetPagination(c)
    
    users, total, err := userService.GetUsers(pagination.Page, pagination.Limit)
    if err != nil {
        return api.InternalError(c, err.Error())
    }
    
    return api.Paginated(c, users, pagination.Page, pagination.Limit, total)
}
```

## Custom Repositories

### Extending BaseRepository

```go
type UserRepository struct {
    *database.BaseRepository[User]
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{
        BaseRepository: database.NewBaseRepository[User](db),
    }
}

// Custom methods
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
    return r.FindOne(ctx, "email = ?", email)
}

func (r *UserRepository) FindActiveUsers(ctx context.Context) ([]*User, error) {
    return r.FindByCondition(ctx, "active = ?", true)
}

func (r *UserRepository) FindByRole(ctx context.Context, role string) ([]*User, error) {
    var users []*User
    err := r.GetDB().WithContext(ctx).
        Joins("JOIN user_roles ON user_roles.user_id = users.id").
        Joins("JOIN roles ON roles.id = user_roles.role_id").
        Where("roles.slug = ?", role).
        Find(&users).Error
    return users, err
}

func (r *UserRepository) Search(ctx context.Context, query string) ([]*User, error) {
    var users []*User
    err := r.GetDB().WithContext(ctx).
        Where("name LIKE ? OR email LIKE ?", "%"+query+"%", "%"+query+"%").
        Find(&users).Error
    return users, err
}
```

### Complex Queries

```go
func (r *PostRepository) FindPublished(ctx context.Context, page, limit int) ([]*Post, int64, error) {
    var posts []*Post
    var total int64
    
    db := r.GetDB().WithContext(ctx)
    
    // Count total
    db.Model(&Post{}).
        Where("status = ? AND published_at <= ?", "published", time.Now()).
        Count(&total)
    
    // Get paginated results
    offset := (page - 1) * limit
    err := db.Where("status = ? AND published_at <= ?", "published", time.Now()).
        Preload("User").
        Preload("Tags").
        Order("published_at DESC").
        Offset(offset).
        Limit(limit).
        Find(&posts).Error
    
    return posts, total, err
}
```

### Advanced Filtering

```go
type PostFilter struct {
    Status    string
    UserID    *uint
    Tag       string
    Search    string
    StartDate *time.Time
    EndDate   *time.Time
}

func (r *PostRepository) Filter(ctx context.Context, filter PostFilter, page, limit int) ([]*Post, int64, error) {
    var posts []*Post
    var total int64
    
    query := r.GetDB().WithContext(ctx).Model(&Post{})
    
    // Apply filters
    if filter.Status != "" {
        query = query.Where("status = ?", filter.Status)
    }
    
    if filter.UserID != nil {
        query = query.Where("user_id = ?", *filter.UserID)
    }
    
    if filter.Tag != "" {
        query = query.Joins("JOIN post_tags ON post_tags.post_id = posts.id").
            Joins("JOIN tags ON tags.id = post_tags.tag_id").
            Where("tags.slug = ?", filter.Tag)
    }
    
    if filter.Search != "" {
        query = query.Where("title LIKE ? OR content LIKE ?",
            "%"+filter.Search+"%", "%"+filter.Search+"%")
    }
    
    if filter.StartDate != nil {
        query = query.Where("created_at >= ?", filter.StartDate)
    }
    
    if filter.EndDate != nil {
        query = query.Where("created_at <= ?", filter.EndDate)
    }
    
    // Count total
    query.Count(&total)
    
    // Get results
    offset := (page - 1) * limit
    err := query.Preload("User").
        Preload("Tags").
        Order("created_at DESC").
        Offset(offset).
        Limit(limit).
        Find(&posts).Error
    
    return posts, total, err
}
```

## Query Building

### Using Query Method

```go
func (r *UserRepository) GetActiveUsersWithPosts(ctx context.Context) ([]*User, error) {
    var users []*User
    
    err := r.Query(ctx).
        Where("active = ?", true).
        Preload("Posts", "status = ?", "published").
        Order("created_at DESC").
        Find(&users).Error
    
    return users, err
}
```

### Complex Joins

```go
func (r *PostRepository) GetPopularPosts(ctx context.Context, limit int) ([]*Post, error) {
    var posts []*Post
    
    err := r.GetDB().WithContext(ctx).
        Select("posts.*, COUNT(comments.id) as comment_count").
        Joins("LEFT JOIN comments ON comments.post_id = posts.id").
        Where("posts.status = ?", "published").
        Group("posts.id").
        Order("comment_count DESC, posts.view_count DESC").
        Limit(limit).
        Find(&posts).Error
    
    return posts, err
}
```

### Subqueries

```go
func (r *UserRepository) FindUsersWithPosts(ctx context.Context) ([]*User, error) {
    var users []*User
    
    subQuery := r.GetDB().Select("user_id").
        Table("posts").
        Where("status = ?", "published").
        Group("user_id")
    
    err := r.GetDB().WithContext(ctx).
        Where("id IN (?)", subQuery).
        Find(&users).Error
    
    return users, err
}
```

## Transactions

### Using WithTx

```go
func (s *UserService) CreateUserWithProfile(req CreateUserRequest) (*User, error) {
    ctx := context.Background()
    
    user := &User{
        Name:  req.Name,
        Email: req.Email,
    }
    
    // Use transaction
    err := s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
        // Use repository with transaction
        txRepo := s.repo.WithTx(tx)
        
        // Create user
        if err := txRepo.Create(ctx, user); err != nil {
            return err
        }
        
        // Create profile
        profile := &Profile{
            UserID: user.ID,
            Bio:    req.Bio,
        }
        
        if err := tx.Create(profile).Error; err != nil {
            return err
        }
        
        return nil
    })
    
    return user, err
}
```

## Best Practices

### 1. Use Context

```go
// Good
func (r *UserRepository) FindByID(ctx context.Context, id uint) (*User, error) {
    return r.BaseRepository.FindByID(ctx, id)
}

// Bad
func (r *UserRepository) FindByID(id uint) (*User, error) {
    return r.BaseRepository.FindByID(context.Background(), id)
}
```

### 2. Create Domain-Specific Methods

```go
type UserRepository struct {
    *database.BaseRepository[User]
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
    return r.FindOne(ctx, "email = ?", email)
}

func (r *UserRepository) FindVerified(ctx context.Context) ([]*User, error) {
    return r.FindByCondition(ctx, "is_email_verified = ?", true)
}
```

### 3. Handle Errors Properly

```go
func (s *UserService) GetByID(id uint) (*User, error) {
    ctx := context.Background()
    
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("database error: %w", err)
    }
    
    if user == nil {
        return nil, errors.New("user not found")
    }
    
    return user, nil
}
```

### 4. Use Preloading Wisely

```go
// Good: Preload only what's needed
func (r *PostRepository) FindWithUser(ctx context.Context, id uint) (*Post, error) {
    var post Post
    err := r.GetDB().WithContext(ctx).
        Preload("User").
        First(&post, id).Error
    return &post, err
}

// Bad: Loading all relations
func (r *PostRepository) FindWithEverything(ctx context.Context, id uint) (*Post, error) {
    var post Post
    err := r.GetDB().WithContext(ctx).
        Preload(clause.Associations).
        First(&post, id).Error
    return &post, err
}
```

### 5. Create Testable Repositories

```go
// Interface for testing
type IUserRepository interface {
    Create(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id uint) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
}

// Implementation
type UserRepository struct {
    *database.BaseRepository[User]
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
    return r.FindOne(ctx, "email = ?", email)
}

// Mock for testing
type MockUserRepository struct {
    CreateFunc      func(ctx context.Context, user *User) error
    FindByIDFunc    func(ctx context.Context, id uint) (*User, error)
    FindByEmailFunc func(ctx context.Context, email string) (*User, error)
}

func (m *MockUserRepository) Create(ctx context.Context, user *User) error {
    return m.CreateFunc(ctx, user)
}
```

### 6. Use Query Builder for Complex Queries

```go
func (r *PostRepository) BuildSearchQuery(ctx context.Context, params SearchParams) *gorm.DB {
    query := r.GetDB().WithContext(ctx).Model(&Post{})
    
    if params.Search != "" {
        query = query.Where("title LIKE ? OR content LIKE ?",
            "%"+params.Search+"%", "%"+params.Search+"%")
    }
    
    if params.Status != "" {
        query = query.Where("status = ?", params.Status)
    }
    
    if params.UserID > 0 {
        query = query.Where("user_id = ?", params.UserID)
    }
    
    return query
}

func (r *PostRepository) Search(ctx context.Context, params SearchParams, page, limit int) ([]*Post, int64, error) {
    var posts []*Post
    var total int64
    
    query := r.BuildSearchQuery(ctx, params)
    
    query.Count(&total)
    
    offset := (page - 1) * limit
    err := query.Offset(offset).Limit(limit).Find(&posts).Error
    
    return posts, total, err
}
```

This comprehensive guide covers everything about using the Repository pattern in NeonEx Framework!

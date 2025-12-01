# Database Models

NeonEx Framework uses GORM for ORM operations, providing powerful model definitions with relationships, hooks, and soft deletes. This guide covers everything you need to create robust database models.

## Table of Contents

- [Basic Models](#basic-models)
- [Model Structure](#model-structure)
- [Relationships](#relationships)
- [Hooks](#hooks)
- [Soft Deletes](#soft-deletes)
- [Indexes](#indexes)
- [Best Practices](#best-practices)

## Basic Models

### Simple Model

```go
package user

import (
    "time"
    "gorm.io/gorm"
)

type User struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    Name      string         `gorm:"size:255;not null" json:"name"`
    Email     string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
    Password  string         `gorm:"size:255;not null" json:"-"`
    Active    bool           `gorm:"default:true" json:"active"`
}

func (User) TableName() string {
    return "users"
}
```

### Field Tags

```go
type Product struct {
    // Primary key
    ID uint `gorm:"primarykey"`
    
    // String with constraints
    Name string `gorm:"size:100;not null"`
    Slug string `gorm:"size:100;uniqueIndex;not null"`
    
    // Text field
    Description string `gorm:"type:text"`
    
    // Numeric fields
    Price    float64 `gorm:"type:decimal(10,2);not null"`
    Quantity int     `gorm:"default:0"`
    
    // Boolean
    Active bool `gorm:"default:true;index"`
    
    // Timestamps
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
    
    // JSON field (requires JSON type in DB)
    Metadata datatypes.JSON `gorm:"type:json"`
}
```

## Model Structure

### Complete User Model

```go
type User struct {
    ID                  uint           `gorm:"primarykey" json:"id"`
    CreatedAt           time.Time      `json:"created_at"`
    UpdatedAt           time.Time      `json:"updated_at"`
    DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
    
    // Basic Info
    Name                string         `gorm:"size:255;not null" json:"name"`
    Email               string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
    Username            string         `gorm:"size:50;uniqueIndex" json:"username"`
    Password            string         `gorm:"size:255;not null" json:"-"`
    Age                 int            `gorm:"default:0" json:"age"`
    
    // Status
    Active              bool           `gorm:"default:true" json:"active"`
    IsEmailVerified     bool           `gorm:"default:false" json:"is_email_verified"`
    EmailVerifiedAt     *time.Time     `json:"email_verified_at,omitempty"`
    LastLoginAt         *time.Time     `json:"last_login_at,omitempty"`
    
    // Security
    PasswordResetToken  *string        `gorm:"size:255" json:"-"`
    PasswordResetExpiry *time.Time     `json:"-"`
    APIKey              *string        `gorm:"size:255;uniqueIndex" json:"-"`
    
    // Relations
    Roles       []rbac.UserRole       `gorm:"foreignKey:UserID" json:"roles,omitempty"`
    Permissions []rbac.UserPermission `gorm:"foreignKey:UserID" json:"permissions,omitempty"`
    Posts       []Post                `gorm:"foreignKey:UserID" json:"posts,omitempty"`
}

func (User) TableName() string {
    return "users"
}
```

### Post Model

```go
type Post struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    
    Title       string `gorm:"size:200;not null" json:"title"`
    Slug        string `gorm:"size:200;uniqueIndex;not null" json:"slug"`
    Content     string `gorm:"type:text;not null" json:"content"`
    Excerpt     string `gorm:"size:500" json:"excerpt"`
    Status      string `gorm:"size:20;default:'draft';index" json:"status"`
    ViewCount   int    `gorm:"default:0" json:"view_count"`
    PublishedAt *time.Time `json:"published_at,omitempty"`
    
    // Foreign Keys
    UserID uint `gorm:"not null;index" json:"user_id"`
    
    // Relations
    User     User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Comments []Comment `gorm:"foreignKey:PostID" json:"comments,omitempty"`
    Tags     []Tag     `gorm:"many2many:post_tags;" json:"tags,omitempty"`
}
```

## Relationships

### One-to-One

```go
type User struct {
    ID      uint    `gorm:"primarykey"`
    Name    string  `gorm:"size:100"`
    Profile Profile `gorm:"foreignKey:UserID"`
}

type Profile struct {
    ID     uint   `gorm:"primarykey"`
    UserID uint   `gorm:"uniqueIndex"`
    Bio    string `gorm:"type:text"`
    Avatar string `gorm:"size:255"`
}

// Create with relation
user := User{
    Name: "John",
    Profile: Profile{
        Bio:    "Software Developer",
        Avatar: "avatar.jpg",
    },
}
db.Create(&user)

// Query with preload
var user User
db.Preload("Profile").First(&user, 1)
```

### One-to-Many

```go
type User struct {
    ID    uint   `gorm:"primarykey"`
    Name  string `gorm:"size:100"`
    Posts []Post `gorm:"foreignKey:UserID"`
}

type Post struct {
    ID     uint   `gorm:"primarykey"`
    UserID uint   `gorm:"not null;index"`
    Title  string `gorm:"size:200"`
    User   User   `gorm:"foreignKey:UserID"`
}

// Query with posts
var user User
db.Preload("Posts").First(&user, 1)

// Query posts with user
var posts []Post
db.Preload("User").Find(&posts)
```

### Many-to-Many

```go
type Post struct {
    ID    uint   `gorm:"primarykey"`
    Title string `gorm:"size:200"`
    Tags  []Tag  `gorm:"many2many:post_tags;"`
}

type Tag struct {
    ID    uint   `gorm:"primarykey"`
    Name  string `gorm:"size:50;uniqueIndex"`
    Posts []Post `gorm:"many2many:post_tags;"`
}

// Create with tags
post := Post{
    Title: "My Post",
    Tags: []Tag{
        {Name: "go"},
        {Name: "fiber"},
    },
}
db.Create(&post)

// Query with tags
var post Post
db.Preload("Tags").First(&post, 1)

// Association operations
db.Model(&post).Association("Tags").Append(&newTag)
db.Model(&post).Association("Tags").Delete(&tag)
```

### Polymorphic Relations

```go
type Comment struct {
    ID            uint   `gorm:"primarykey"`
    Content       string `gorm:"type:text"`
    CommentableID uint   `gorm:"index"`
    CommentableType string `gorm:"size:50;index"`
}

type Post struct {
    ID       uint      `gorm:"primarykey"`
    Title    string    `gorm:"size:200"`
    Comments []Comment `gorm:"polymorphic:Commentable;"`
}

type Video struct {
    ID       uint      `gorm:"primarykey"`
    Title    string    `gorm:"size:200"`
    Comments []Comment `gorm:"polymorphic:Commentable;"`
}
```

## Hooks

### Before/After Create

```go
func (u *User) BeforeCreate(tx *gorm.DB) error {
    // Hash password before saving
    if u.Password != "" {
        hasher := auth.NewPasswordHasher(12)
        hashed, err := hasher.Hash(u.Password)
        if err != nil {
            return err
        }
        u.Password = hashed
    }
    
    // Generate API key
    apiKey, err := auth.GenerateAPIKey()
    if err != nil {
        return err
    }
    u.APIKey = &apiKey
    
    return nil
}

func (u *User) AfterCreate(tx *gorm.DB) error {
    // Send welcome email
    emailService.SendWelcomeEmail(u.Email, u.Name)
    
    // Dispatch event
    events.Dispatch(context.Background(), events.Event{
        Name: events.EventUserCreated,
        Data: u,
    })
    
    return nil
}
```

### Before/After Update

```go
func (u *User) BeforeUpdate(tx *gorm.DB) error {
    // Check if password changed
    if tx.Statement.Changed("Password") {
        hasher := auth.NewPasswordHasher(12)
        hashed, err := hasher.Hash(u.Password)
        if err != nil {
            return err
        }
        u.Password = hashed
    }
    
    return nil
}

func (u *User) AfterUpdate(tx *gorm.DB) error {
    // Invalidate cache
    cache.Delete(fmt.Sprintf("user:%d", u.ID))
    
    return nil
}
```

### Before/After Delete

```go
func (u *User) BeforeDelete(tx *gorm.DB) error {
    // Delete related records
    tx.Where("user_id = ?", u.ID).Delete(&Post{})
    tx.Where("user_id = ?", u.ID).Delete(&Comment{})
    
    return nil
}

func (u *User) AfterDelete(tx *gorm.DB) error {
    // Log deletion
    logger.Info("User deleted", logger.Fields{
        "user_id": u.ID,
        "email":   u.Email,
    })
    
    return nil
}
```

### After Find

```go
func (p *Post) AfterFind(tx *gorm.DB) error {
    // Increment view count asynchronously
    go func() {
        tx.Model(p).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))
    }()
    
    return nil
}
```

## Soft Deletes

### Basic Soft Delete

```go
type User struct {
    ID        uint           `gorm:"primarykey"`
    Name      string         `gorm:"size:100"`
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Soft delete (sets DeletedAt)
db.Delete(&user)

// Query (excludes soft deleted)
db.Find(&users)

// Include soft deleted
db.Unscoped().Find(&users)

// Permanent delete
db.Unscoped().Delete(&user)

// Restore
db.Model(&user).Update("deleted_at", nil)
```

### Custom Soft Delete Field

```go
type User struct {
    ID        uint       `gorm:"primarykey"`
    Name      string     `gorm:"size:100"`
    DeletedAt *time.Time `gorm:"index"`
    DeletedBy *uint      // Track who deleted
}

func (u *User) BeforeDelete(tx *gorm.DB) error {
    // Get current user from context
    if userID, ok := tx.Statement.Context.Value("user_id").(uint); ok {
        u.DeletedBy = &userID
    }
    return nil
}
```

## Indexes

### Single Column Index

```go
type User struct {
    Email string `gorm:"index"` // Regular index
    Phone string `gorm:"uniqueIndex"` // Unique index
}
```

### Composite Index

```go
type User struct {
    FirstName string `gorm:"index:idx_name"`
    LastName  string `gorm:"index:idx_name"`
}
```

### Custom Index Names

```go
type User struct {
    Email string `gorm:"index:idx_user_email,unique"`
}
```

### Full-Text Index (MySQL)

```go
type Post struct {
    Title   string `gorm:"index:,class:FULLTEXT"`
    Content string `gorm:"index:,class:FULLTEXT"`
}
```

## Best Practices

### 1. Always Use Timestamps

```go
type Model struct {
    ID        uint      `gorm:"primarykey"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### 2. Use Soft Deletes for User Data

```go
type User struct {
    ID        uint           `gorm:"primarykey"`
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

### 3. Index Foreign Keys

```go
type Post struct {
    ID     uint `gorm:"primarykey"`
    UserID uint `gorm:"not null;index"` // Indexed for performance
}
```

### 4. Use Appropriate Data Types

```go
type Product struct {
    Price    float64        `gorm:"type:decimal(10,2)"`
    Metadata datatypes.JSON `gorm:"type:json"`
}
```

### 5. Validate in Hooks

```go
func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.Email == "" {
        return errors.New("email required")
    }
    return nil
}
```

### 6. Use Preload Wisely

```go
// Good: Preload specific relations
db.Preload("Posts").Preload("Profile").Find(&users)

// Bad: Loads all relations (slow)
db.Preload(clause.Associations).Find(&users)
```

This comprehensive guide covers all aspects of working with models in NeonEx Framework!

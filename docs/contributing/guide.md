# Contributing Guide

Thank you for considering contributing to NeonEx Framework! This guide will help you get started.

---

## 🌟 Ways to Contribute

### 1. Code Contributions
- Fix bugs
- Add features
- Improve performance
- Refactor code

### 2. Documentation
- Fix typos and errors
- Add examples
- Improve explanations
- Translate documentation

### 3. Testing
- Write tests
- Report bugs
- Test new features
- Improve test coverage

### 4. Community
- Answer questions
- Help newcomers
- Review pull requests
- Share knowledge

---

## 🚀 Getting Started

### 1. Fork the Repository

```bash
# Fork on GitHub, then clone
git clone https://github.com/YOUR_USERNAME/neonexframework.git
cd neonexframework
```

### 2. Set Up Development Environment

```bash
# Install dependencies
go mod download

# Copy environment file
cp .env.example .env

# Setup database
createdb neonexdb_dev

# Run application
go run main.go
```

### 3. Create a Branch

```bash
# Create feature branch
git checkout -b feature/my-new-feature

# Or bugfix branch
git checkout -b fix/bug-description
```

### 4. Make Changes

- Write clean, readable code
- Follow Go best practices
- Add tests for new features
- Update documentation

### 5. Test Your Changes

```bash
# Run tests
go test ./...

# Run specific tests
go test ./modules/blog/...

# With coverage
go test -cover ./...
```

### 6. Commit Your Changes

```bash
# Stage changes
git add .

# Commit with descriptive message
git commit -m "feat: add blog post tagging feature"
```

**Commit Message Format:**
```
type(scope): subject

body (optional)

footer (optional)
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Formatting
- `refactor`: Code restructuring
- `test`: Adding tests
- `chore`: Maintenance

**Examples:**
```bash
git commit -m "feat(blog): add post tagging"
git commit -m "fix(auth): resolve JWT expiration bug"
git commit -m "docs: update installation guide"
```

### 7. Push Changes

```bash
git push origin feature/my-new-feature
```

### 8. Create Pull Request

1. Go to your fork on GitHub
2. Click "New Pull Request"
3. Select your branch
4. Fill in the PR template
5. Submit!

---

## 📋 Pull Request Guidelines

### PR Title

Follow commit message format:
```
feat(module): add new feature
fix(core): resolve bug
docs(readme): update examples
```

### PR Description

Include:
- **What**: What changes did you make?
- **Why**: Why are these changes needed?
- **How**: How did you implement it?
- **Testing**: How did you test it?
- **Screenshots**: If UI changes

**Template:**
```markdown
## Description
Brief description of changes

## Motivation
Why is this change needed?

## Changes
- Change 1
- Change 2
- Change 3

## Testing
How was this tested?

## Screenshots
(if applicable)

## Checklist
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] Code follows style guidelines
- [ ] Self-review completed
```

### Before Submitting

- [ ] Code compiles without errors
- [ ] All tests pass
- [ ] New tests added for new features
- [ ] Documentation updated
- [ ] Code formatted (`gofmt`)
- [ ] Linter passes (`golangci-lint`)
- [ ] Self-review completed
- [ ] Breaking changes documented

---

## 💻 Development Guidelines

### Code Style

Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

**Good:**
```go
// GetUserByID retrieves a user by ID
func GetUserByID(id uint) (*User, error) {
    var user User
    err := db.First(&user, id).Error
    return &user, err
}
```

**Bad:**
```go
func get_user(i uint) (*User, error) {  // Wrong naming
    u := User{}  // Not descriptive
    e := db.First(&u, i).Error  // Single letter vars
    return &u, e
}
```

### Naming Conventions

**Packages:**
```go
package blog    // lowercase, singular
```

**Functions:**
```go
func NewUserService()      // Public: PascalCase
func validateEmail()       // Private: camelCase
```

**Variables:**
```go
var userName string        // Public: PascalCase
var isActive bool         // Private: camelCase
```

**Constants:**
```go
const MaxRetries = 3      // Public
const defaultTimeout = 30 // Private
```

### Error Handling

**Good:**
```go
user, err := service.GetUser(id)
if err != nil {
    return nil, fmt.Errorf("failed to get user: %w", err)
}
```

**Bad:**
```go
user, _ := service.GetUser(id)  // Ignoring errors
```

### Comments

**Good:**
```go
// UserService handles user-related business logic.
// It provides methods for user management including
// registration, authentication, and profile updates.
type UserService struct {
    repo UserRepository
}

// Create creates a new user with the provided data.
// It returns the created user or an error if validation fails.
func (s *UserService) Create(dto *CreateUserDTO) (*User, error) {
    // Validate email uniqueness
    exists, err := s.repo.ExistsByEmail(dto.Email)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, errors.New("email already exists")
    }
    
    // Create user
    user := &User{
        Name:  dto.Name,
        Email: dto.Email,
    }
    
    return s.repo.Create(user)
}
```

### Testing

Write tests for:
- All public functions
- Critical business logic
- Edge cases
- Error conditions

**Example:**
```go
func TestUserService_Create(t *testing.T) {
    // Arrange
    mockRepo := new(MockUserRepository)
    service := NewUserService(mockRepo)
    dto := &CreateUserDTO{
        Name:  "John Doe",
        Email: "john@example.com",
    }
    
    mockRepo.On("ExistsByEmail", dto.Email).Return(false, nil)
    mockRepo.On("Create", mock.Anything).Return(nil)
    
    // Act
    user, err := service.Create(dto)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "John Doe", user.Name)
    mockRepo.AssertExpectations(t)
}
```

---

## 📚 Documentation Guidelines

### Code Documentation

```go
// Package blog provides blog post management functionality.
//
// It includes models, repositories, services, and controllers
// for managing blog posts, comments, and tags.
package blog

// Post represents a blog post entity.
//
// Fields:
//   - Title: The post title (required, max 200 chars)
//   - Content: The post content in markdown format
//   - Published: Whether the post is published
type Post struct {
    ID        uint      `json:"id"`
    Title     string    `json:"title" gorm:"not null"`
    Content   string    `json:"content" gorm:"type:text"`
    Published bool      `json:"published" gorm:"default:false"`
    CreatedAt time.Time `json:"created_at"`
}
```

### Markdown Documentation

- Use clear headings
- Include code examples
- Add links to related docs
- Keep it concise
- Use proper formatting

---

## 🐛 Reporting Bugs

### Before Reporting

- Search existing issues
- Test with latest version
- Reproduce the bug
- Gather information

### Bug Report Template

```markdown
## Bug Description
Clear description of the bug

## Steps to Reproduce
1. Step 1
2. Step 2
3. Step 3

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Environment
- OS: [e.g., Ubuntu 22.04]
- Go Version: [e.g., 1.21.0]
- NeonEx Version: [e.g., 0.2.0]

## Additional Context
Any other relevant information
```

---

## 💡 Proposing Features

### Before Proposing

- Check if already proposed
- Ensure it fits framework goals
- Consider implementation complexity
- Think about breaking changes

### Feature Request Template

```markdown
## Feature Description
What feature do you want?

## Motivation
Why is this feature needed?

## Proposed Solution
How should it work?

## Alternatives Considered
What other approaches did you consider?

## Additional Context
Any other relevant information
```

---

## 🔍 Code Review Process

### For Contributors

- Be open to feedback
- Respond to comments promptly
- Make requested changes
- Ask questions if unclear

### For Reviewers

- Be respectful and constructive
- Focus on the code, not the person
- Explain reasoning
- Approve when ready

---

## 🏆 Recognition

Contributors are recognized:
- In release notes
- On contributors page
- In documentation
- With GitHub badges

---

## 📞 Getting Help

Stuck? Need help? Reach out:

- **GitHub Discussions**: [Ask Questions](https://github.com/neonextechnologies/neonexframework/discussions)
- **Discord**: [Join Chat](https://discord.gg/neonex) *(coming soon)*
- **Email**: contribute@neonexframework.dev

---

## 📜 Code of Conduct

We expect all contributors to follow our [Code of Conduct](./code-of-conduct.md):

- Be respectful
- Be inclusive
- Be collaborative
- Be professional

---

## 📄 License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

## 🙏 Thank You!

Thank you for contributing to NeonEx Framework! Every contribution, no matter how small, helps make NeonEx better for everyone.

**Happy coding!** 🚀

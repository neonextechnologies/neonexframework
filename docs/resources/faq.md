# Frequently Asked Questions

Common questions about NeonEx Framework.

---

## General Questions

### What is NeonEx Framework?

NeonEx Framework is a full-stack Go framework built on [NeonEx Core](https://github.com/neonextechnologies/neonexcore). It provides everything you need to build modern web applications, from RESTful APIs to full-stack web apps.

### Is NeonEx Framework free?

Yes! NeonEx Framework is open-source under the MIT License. You can use it for free in personal and commercial projects.

### What's the difference between NeonEx Framework and NeonEx Core?

- **NeonEx Core** - The foundation (module system, DI, auth, database)
- **NeonEx Framework** - Built on Core with extended features (frontend, themes, assets)

### Is it production-ready?

Yes! NeonEx Framework is used in production by several companies and can handle high traffic loads (10,000+ req/sec).

---

## Getting Started

### What are the system requirements?

- **Go 1.21+** (required)
- **PostgreSQL/MySQL/SQLite** (database)
- **Redis** (optional, for caching)
- **2GB+ RAM** (recommended)

### Can I use MySQL instead of PostgreSQL?

Yes! NeonEx supports PostgreSQL, MySQL, and SQLite. Just change `DB_CONNECTION` in your `.env` file.

### Do I need Redis?

No, Redis is optional. It's recommended for:
- Distributed caching
- Session storage
- Queue management

The framework works without it using in-memory alternatives.

### How long does it take to learn?

- **Basic usage**: 1-2 days if you know Go
- **Proficient**: 1-2 weeks of active development
- **Expert**: 1-2 months building real projects

---

## Development

### How do I create a new module?

```bash
# 1. Create directory
mkdir -p modules/mymodule/pkg

# 2. Create module.go
# 3. Register in main.go
core.ModuleMap["mymodule"] = mymodule.New()
```

See [Modules Guide](../core-concepts/modules.md) for details.

### Can I use my existing Go packages?

Absolutely! NeonEx doesn't lock you in. Use any Go package you want.

### How do I add middleware?

```go
// Global middleware
app.Use(myMiddleware())

// Route-specific
api.Use(middleware.Auth())
```

### How do I handle file uploads?

```go
func Upload(ctx *fiber.Ctx) error {
    file, _ := ctx.FormFile("file")
    path, _ := storage.Put(file)
    return ctx.JSON(fiber.Map{"path": path})
}
```

See [File Storage](../advanced/file-storage.md) for details.

---

## Database

### How do I run migrations?

Migrations run automatically on startup via `AutoMigrate`. Or manually:

```go
db.AutoMigrate(&User{}, &Post{}, &Comment{})
```

### Can I use raw SQL queries?

Yes!

```go
db.Raw("SELECT * FROM users WHERE age > ?", 18).Scan(&users)
```

### How do I seed data?

Create a seeder:

```go
func SeedUsers(db *gorm.DB) {
    users := []User{
        {Name: "John", Email: "john@example.com"},
        {Name: "Jane", Email: "jane@example.com"},
    }
    db.Create(&users)
}
```

---

## Authentication

### How does authentication work?

NeonEx uses JWT tokens:

1. User logs in with email/password
2. Server generates JWT token
3. Client includes token in requests
4. Server validates token on protected routes

### How do I protect routes?

```go
api.Use(middleware.Auth())
```

### How do I get the current user?

```go
userID := ctx.Locals("user_id").(uint)
user := ctx.Locals("user").(*User)
```

### Can I use OAuth (Google, Facebook)?

Yes! Integration guides coming soon. For now, use libraries like:
- `golang.org/x/oauth2`
- `github.com/markbates/goth`

---

## Performance

### How fast is NeonEx Framework?

- **10,000+ requests/second** on standard hardware
- **Sub-millisecond latency** for simple endpoints
- **Low memory usage** (~20MB base)

### How do I improve performance?

1. **Enable caching** - Use Redis
2. **Database indexes** - Add indexes to frequently queried columns
3. **Connection pooling** - Configure `DB_MAX_OPEN_CONNS`
4. **Pagination** - Don't load all records at once
5. **Asset minification** - Enable in production

### Can it handle 10,000 concurrent users?

Yes! With proper configuration:
- Use Redis for sessions
- Configure connection pools
- Enable caching
- Use load balancing

---

## Deployment

### How do I deploy to production?

```bash
# Build binary
go build -o app main.go

# Run
./app
```

See [Deployment Guide](../deployment/overview.md) for details.

### Does it work with Docker?

Yes! Example Dockerfile included. See [Docker Guide](../deployment/docker.md).

### Can I deploy to Kubernetes?

Absolutely! See [Kubernetes Guide](../deployment/kubernetes.md).

### Which cloud providers are supported?

All major providers:
- AWS (EC2, ECS, Lambda)
- Google Cloud (GCE, Cloud Run)
- Azure (App Service, AKS)
- DigitalOcean
- Heroku

---

## Troubleshooting

### Port 8080 already in use

Change the port in `.env`:

```env
APP_PORT=3000
```

### Database connection failed

Check your `.env` credentials:

```env
DB_HOST=localhost
DB_PORT=5432
DB_DATABASE=mydb
DB_USERNAME=postgres
DB_PASSWORD=secret
```

Test connection:

```bash
psql -h localhost -U postgres -d mydb
```

### Module not found error

Run:

```bash
go mod tidy
go mod download
```

### JWT token invalid

Ensure:
1. Token is included in `Authorization: Bearer TOKEN`
2. Token hasn't expired
3. `JWT_SECRET` is same as when token was generated

---

## Comparison

### NeonEx vs Laravel

| Feature | Laravel | NeonEx |
|---------|---------|---------|
| Language | PHP | Go |
| Performance | ~1,200 req/s | ~10,000 req/s |
| Memory | ~50MB | ~20MB |
| Learning Curve | Moderate | Moderate |
| Deployment | Complex | Single binary |

### NeonEx vs Express.js

| Feature | Express | NeonEx |
|---------|---------|---------|
| Language | JavaScript | Go |
| Performance | ~3,500 req/s | ~10,000 req/s |
| Type Safety | Optional | Built-in |
| Complete | No | Yes |
| Setup Time | Long | Short |

### NeonEx vs Gin/Echo

| Feature | Gin/Echo | NeonEx |
|---------|----------|---------|
| HTTP | ✅ | ✅ |
| Database | Manual | Built-in |
| Auth | Manual | Built-in |
| Templates | Basic | Advanced |
| Time to MVP | Weeks | Days |

---

## Contributing

### How can I contribute?

- Report bugs
- Suggest features
- Submit pull requests
- Improve documentation
- Share your projects

See [Contributing Guide](../contributing/guide.md).

### Where can I get help?

- [GitHub Discussions](https://github.com/neonextechnologies/neonexframework/discussions)
- [GitHub Issues](https://github.com/neonextechnologies/neonexframework/issues)
- Email: support@neonexframework.dev

---

## Licensing

### Can I use it commercially?

Yes! MIT License allows commercial use.

### Do I need to credit NeonEx?

Not required, but appreciated! 😊

### Can I modify the source code?

Yes! You can modify and distribute as you wish under MIT License.

---

## More Questions?

Can't find your answer? Ask us:

- 💬 [GitHub Discussions](https://github.com/neonextechnologies/neonexframework/discussions)
- 🐛 [GitHub Issues](https://github.com/neonextechnologies/neonexframework/issues)
- 📧 Email: support@neonexframework.dev

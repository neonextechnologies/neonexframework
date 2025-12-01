# NeonEx Framework

<div align="center">

![NeonEx Framework](https://img.shields.io/badge/NeonEx-Framework-purple?style=for-the-badge)
![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)

**Modern Full-Stack Go Framework**

*Fast • Scalable • Production-Ready*

[Features](#-features) • [Quick Start](#-quick-start) • [Documentation](./docs) • [Core](./core)

</div>

---

## ✨ Overview

**NeonEx Framework** เป็น full-stack Go framework ที่สร้างจาก [NeonEx Core](https://github.com/neonextechnologies/neonexcore) ออกแบบมาเพื่อพัฒนาแอปพลิเคชันเว็บที่ทันสมัย รวดเร็ว และครบครัน

### ทำไมต้อง NeonEx Framework?

- 🚀 **Performance** - สร้างจาก Go เพื่อความเร็วสูงสุด (10,000+ req/sec)
- 🎯 **Full-Stack** - ทุกอย่างที่ต้องการสำหรับสร้างเว็บแอปพลิเคชัน
- 🏗️ **Modular** - ระบบ module ที่ยืดหยุ่น ขยายได้ง่าย
- 🔐 **Secure** - มาพร้อม Authentication, Authorization, RBAC
- 📦 **Complete** - Database, API, WebSocket, GraphQL, และอื่นๆ
- 🎨 **Frontend Ready** - รองรับ template engine และ asset management
- 🧪 **Testable** - Built-in testing utilities
- 🚢 **Production Ready** - Deploy เป็น single binary

---

## 🎯 Key Features

### Core Framework (from NeonEx Core)
- **🎨 Modular Architecture** - Self-contained modules with dependency injection
- **⚡ High Performance** - Built on Fiber v2 (10,000+ req/sec)
- **💉 Dependency Injection** - Type-safe DI container with auto-resolution
- **🔐 Authentication & Authorization** - JWT + RBAC out of the box
- **📊 ORM Integration** - GORM with PostgreSQL, MySQL, SQLite support
- **🔄 Auto-Migration** - Database schema management
- **🗄️ Generic Repository** - Type-safe CRUD operations

### Advanced Features
- **🌐 WebSocket Support** - Real-time bidirectional communication
- **📡 GraphQL API** - Schema-first GraphQL with subscriptions
- **🚀 gRPC/Microservices** - High-performance RPC
- **🗄️ Multi-level Caching** - Redis integration
- **📊 Metrics & Monitoring** - Prometheus metrics
- **🔍 Full-text Search** - Search capabilities
- **📧 Email System** - SMTP integration
- **📝 Logging** - Structured logging with Zap

### Frontend Support
- **🎨 Template Engine** - HTML template rendering
- **📦 Asset Pipeline** - CSS/JS bundling และ minification
- **🖼️ Theme System** - Multiple themes support
- **📱 Responsive** - Mobile-ready

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** - [Download](https://go.dev/dl/)
- **PostgreSQL** - [Download](https://www.postgresql.org/download/)
- **Git** - For version control

### Installation

```bash
# 1. Clone the repository
git clone https://github.com/neonextechnologies/neonexframework.git
cd neonexframework

# 2. Install dependencies
go mod download

# 3. Set up environment
cp .env.example .env
# Edit .env with your database credentials

# 4. Run the application
go run main.go
```

### First API Request

```bash
# Register a user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "secret123"
  }'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "secret123"
  }'
```

---

## 📦 Project Structure

```
neonexframework/
├── core/                   # NeonEx Core (dependency)
│   ├── internal/          # Core framework internals
│   ├── modules/           # Built-in modules (user, admin)
│   └── pkg/               # Public packages
│
├── modules/               # Framework modules
│   ├── frontend/         # Template & asset management
│   ├── web/              # Web utilities
│   └── [your-modules]/   # Your custom modules
│
├── public/               # Static files
│   ├── css/             # Stylesheets
│   ├── js/              # JavaScript
│   ├── images/          # Images
│   └── uploads/         # User uploads
│
├── templates/            # HTML templates
│   ├── layouts/         # Layout templates
│   └── frontend/        # Frontend templates
│
├── storage/              # Storage directory
│   ├── logs/            # Application logs
│   ├── cache/           # Cache files
│   └── uploads/         # Uploaded files
│
├── scripts/              # Utility scripts
│   ├── update-core.sh   # Update core (Bash)
│   └── update-core.ps1  # Update core (PowerShell)
│
├── docs/                 # Documentation
├── tests/                # Tests
├── go.mod                # Go modules
├── main.go               # Application entry
├── Makefile             # Build commands
└── README.md            # This file
```

---

## 💡 Usage Examples

### Basic Web Application

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "neonexcore/internal/core"
)

func main() {
    app := core.NewApp()
    
    // Auto-discovers and loads modules
    // Sets up database, logging, routing
    
    app.Run() // Starts server on :8080
}
```

### Creating a Custom Module

```go
// modules/blog/module.go
package blog

import (
    "github.com/gofiber/fiber/v2"
    "neonexcore/internal/core"
)

type BlogModule struct{}

func New() *BlogModule {
    return &BlogModule{}
}

func (m *BlogModule) Name() string {
    return "blog"
}

func (m *BlogModule) RegisterServices(c *core.Container) error {
    c.Provide(NewBlogRepository)
    c.Provide(NewBlogService)
    c.Provide(NewBlogController)
    return nil
}

func (m *BlogModule) RegisterRoutes(router fiber.Router) error {
    api := router.Group("/api/v1/blog")
    
    ctrl := core.Resolve[*BlogController]()
    
    api.Get("/posts", ctrl.List)
    api.Get("/posts/:id", ctrl.Get)
    api.Post("/posts", ctrl.Create)
    api.Put("/posts/:id", ctrl.Update)
    api.Delete("/posts/:id", ctrl.Delete)
    
    return nil
}

func (m *BlogModule) Boot() error {
    return nil
}
```

### Using Repository Pattern

```go
// Generic repository with type safety
repo := database.NewBaseRepository[Post](db)

// CRUD operations
posts, _ := repo.FindAll(ctx)
post, _ := repo.FindByID(ctx, 1)
repo.Create(ctx, &newPost)
repo.Update(ctx, &post)
repo.Delete(ctx, 1)

// With conditions
posts, _ := repo.FindWhere(ctx, "published = ?", true)

// Pagination
posts, total, _ := repo.Paginate(ctx, 1, 20)
```

### Dependency Injection

```go
// Register services
container.Provide(NewDatabase)
container.Provide(NewUserRepository)
container.Provide(NewUserService)

// Auto-resolve with dependencies
service := container.Resolve[*UserService]()
// Dependencies automatically injected
```

---

## 🛠️ Development

### Hot Reload

```bash
make dev
```

### Running Tests

```bash
make test
```

### Code Generation

```bash
# Generate a new module
go run main.go make:module blog

# Generate specific components
go run main.go make:model Post
go run main.go make:service PostService
go run main.go make:controller PostController
```

### Database Migrations

```bash
# Run migrations
go run main.go migrate:up

# Rollback
go run main.go migrate:down

# Create new migration
go run main.go make:migration create_posts_table
```

---

## 📚 Documentation

- **[Getting Started](./docs/getting-started.md)** - เริ่มต้นใช้งาน
- **[Core Management](./docs/core-management.md)** - จัดการ core dependency
- **[Module Development](./docs/module-development.md)** - สร้าง custom modules
- **[API Reference](./docs/api-reference.md)** - API documentation
- **[Deployment](./docs/deployment.md)** - การ deploy production

---

## 🔄 Updating Core

NeonEx Framework ใช้ [neonexcore](https://github.com/neonextechnologies/neonexcore) เป็น dependency ซึ่งจัดเก็บไว้ใน `/core` directory

### Update Core to Latest Version

```bash
# Using Make (recommended)
make update-core

# Or manually
# Windows
powershell -ExecutionPolicy Bypass -File scripts/update-core.ps1

# Linux/Mac
bash scripts/update-core.sh
```

The update script will:
1. Backup current core
2. Clone latest neonexcore
3. Clean unnecessary files
4. Run tests
5. Restore backup if tests fail

See [Core Management Guide](./docs/core-management.md) for details.

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────┐
│       Your Application                   │
│   (Custom Modules & Business Logic)     │
├─────────────────────────────────────────┤
│         NeonEx Framework                 │
│  (Frontend, Web Utilities, Templates)   │
├─────────────────────────────────────────┤
│         NeonEx Core                      │
│  (Framework Core with 20+ Components)   │
├─────────────────────────────────────────┤
│    Fiber + GORM + Zap + Redis...        │
└─────────────────────────────────────────┘
```

---

## 🌟 What Can You Build?

NeonEx Framework เหมาะสำหรับสร้าง:

- 📝 **RESTful APIs** - Backend services
- 🌐 **Web Applications** - Full-stack web apps
- 📱 **Mobile Backends** - API for mobile apps
- 🎮 **Real-time Apps** - WebSocket applications
- 🔄 **Microservices** - Distributed systems
- 📊 **Admin Dashboards** - Management interfaces
- 🛒 **E-commerce Platforms** - Online stores
- 📰 **Content Platforms** - Blogs, news sites
- 💬 **Social Networks** - Community platforms
- 🎓 **Learning Management** - Education platforms

---

## 🤝 Contributing

We welcome contributions!

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

See [CONTRIBUTING.md](./CONTRIBUTING.md) for details.

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

## 💬 Support & Community

- **Documentation**: [docs/](./docs)
- **Issues**: [GitHub Issues](https://github.com/neonextechnologies/neonexframework/issues)
- **Discussions**: [GitHub Discussions](https://github.com/neonextechnologies/neonexframework/discussions)
- **Core Repository**: [neonexcore](https://github.com/neonextechnologies/neonexcore)
- **Email**: support@neonexframework.dev

---

## 🙏 Acknowledgments

Built on top of amazing open source projects:

- [NeonEx Core](https://github.com/neonextechnologies/neonexcore) - Framework core
- [Fiber](https://github.com/gofiber/fiber) - Fast HTTP framework
- [GORM](https://gorm.io) - ORM library
- [Zap](https://github.com/uber-go/zap) - Structured logging

Inspired by:
- [Laravel](https://laravel.com) - Elegant PHP framework
- [NestJS](https://nestjs.com) - Progressive Node.js framework
- [Spring Boot](https://spring.io/projects/spring-boot) - Java framework

---

<div align="center">

**Built with ❤️ by NeoNex Technologies**

**[⭐ Star us on GitHub](https://github.com/neonextechnologies/neonexframework)** | **[📖 Documentation](./docs)** | **[🚀 Get Started](#-quick-start)**

</div>

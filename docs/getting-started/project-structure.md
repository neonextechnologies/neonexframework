# Project Structure

Understanding the NeonEx Framework project structure helps you organize your code effectively.

---

## Overview

```
neonexframework/
├── core/                   # NeonEx Core (dependency)
├── modules/               # Application modules
├── config/                # Configuration files
├── public/                # Public assets
├── templates/             # HTML templates
├── storage/               # Storage directory
├── scripts/               # Utility scripts
├── tests/                 # Test files
├── docs/                  # Documentation
├── main.go                # Application entry point
├── go.mod                 # Go module file
├── go.sum                 # Go dependencies checksum
├── .env                   # Environment variables
├── .env.example           # Environment template
├── Makefile              # Build commands
└── README.md              # Project README
```

---

## Core Directory

The `core/` directory contains NeonEx Core, which is the framework foundation.

```
core/
├── internal/              # Core internals
│   ├── core/             # Core functionality
│   │   ├── app.go        # Application bootstrap
│   │   ├── container.go  # DI container
│   │   ├── module.go     # Module interface
│   │   └── router.go     # Router setup
│   ├── database/         # Database layer
│   ├── middleware/       # Core middleware
│   └── utils/            # Utility functions
├── modules/              # Built-in modules
│   ├── admin/           # Admin module
│   └── user/            # User module
└── pkg/                  # Public packages
    ├── config/          # Configuration
    ├── logger/          # Logging
    └── validator/       # Validation
```

**Note:** Don't modify core files directly. Use the [update-core](../core-management.md) script for updates.

---

## Modules Directory

Your application modules live here. Each module is self-contained.

```
modules/
├── frontend/             # Frontend module
│   ├── module.go        # Module definition
│   └── pkg/             # Module packages
│       ├── theme.go     # Theme system
│       ├── assets.go    # Asset management
│       └── template.go  # Template service
│
├── web/                  # Web module
│   ├── module.go        # Module definition
│   └── pkg/             # Web utilities
│
└── blog/                 # Example: Blog module
    ├── module.go        # Module definition
    ├── models/          # Data models
    │   └── post.go
    ├── repositories/    # Data access
    │   └── post_repository.go
    ├── services/        # Business logic
    │   └── post_service.go
    ├── controllers/     # HTTP handlers
    │   └── post_controller.go
    ├── dto/             # Data Transfer Objects
    │   └── post_dto.go
    ├── routes/          # Route definitions
    │   └── routes.go
    └── tests/           # Module tests
        └── post_test.go
```

### Module Structure Best Practices

**Option 1: Flat Structure** (Simple modules)
```
modules/blog/
├── module.go
└── pkg/
    ├── post.go         # Model
    ├── service.go      # Service
    └── controller.go   # Controller
```

**Option 2: Layered Structure** (Complex modules)
```
modules/blog/
├── module.go
├── models/
├── repositories/
├── services/
├── controllers/
├── dto/
├── middleware/
└── tests/
```

---

## Config Directory

Configuration files for different environments.

```
config/
├── app.go               # Application config
├── database.go          # Database config
├── mail.go             # Email config
├── cache.go            # Cache config
└── custom.go           # Custom config
```

Example `config/app.go`:

```go
package config

type AppConfig struct {
    Name     string
    Env      string
    Debug    bool
    Port     int
    URL      string
}

func LoadAppConfig() *AppConfig {
    return &AppConfig{
        Name:  GetEnv("APP_NAME", "NeonEx"),
        Env:   GetEnv("APP_ENV", "development"),
        Debug: GetEnvBool("APP_DEBUG", true),
        Port:  GetEnvInt("APP_PORT", 8080),
        URL:   GetEnv("APP_URL", "http://localhost:8080"),
    }
}
```

---

## Public Directory

Static files accessible via HTTP.

```
public/
├── css/                 # Stylesheets
│   ├── app.css
│   └── admin.css
├── js/                  # JavaScript
│   ├── app.js
│   └── admin.js
├── images/             # Images
│   ├── logo.png
│   └── favicon.ico
├── fonts/              # Fonts
└── uploads/            # User uploads
    └── .gitkeep
```

**Access via URL:**
- `public/css/app.css` → `http://localhost:8080/css/app.css`
- `public/images/logo.png` → `http://localhost:8080/images/logo.png`

---

## Templates Directory

HTML templates for server-side rendering.

```
templates/
├── layouts/            # Layout templates
│   ├── main.html      # Main layout
│   ├── admin.html     # Admin layout
│   └── auth.html      # Auth layout
│
├── partials/          # Reusable components
│   ├── header.html
│   ├── footer.html
│   ├── sidebar.html
│   └── nav.html
│
├── pages/             # Page templates
│   ├── home.html
│   ├── about.html
│   └── contact.html
│
├── auth/              # Authentication pages
│   ├── login.html
│   └── register.html
│
├── admin/             # Admin pages
│   ├── dashboard.html
│   └── users.html
│
└── errors/            # Error pages
    ├── 404.html
    ├── 500.html
    └── 403.html
```

**Example Template** (`templates/pages/home.html`):

```html
{{define "pages/home"}}
{{template "layouts/main" .}}

{{define "content"}}
<div class="container">
    <h1>{{.Title}}</h1>
    <p>Welcome, {{.User.Name}}!</p>
</div>
{{end}}

{{end}}
```

---

## Storage Directory

Application storage for logs, cache, uploads.

```
storage/
├── logs/              # Application logs
│   ├── app.log
│   ├── error.log
│   └── access.log
│
├── cache/             # Cache files
│   └── .gitkeep
│
├── uploads/           # File uploads
│   ├── images/
│   ├── documents/
│   └── .gitkeep
│
└── temp/              # Temporary files
    └── .gitkeep
```

**Permissions:**
```bash
# Make storage writable
chmod -R 775 storage/
```

---

## Scripts Directory

Utility scripts for development and deployment.

```
scripts/
├── update-core.sh     # Update core (Bash)
├── update-core.ps1    # Update core (PowerShell)
├── deploy.sh          # Deployment script
├── backup.sh          # Backup script
└── seed.sh            # Database seeding
```

---

## Tests Directory

Test files for your application.

```
tests/
├── unit/              # Unit tests
│   ├── services/
│   └── repositories/
│
├── integration/       # Integration tests
│   └── api/
│
├── e2e/              # End-to-end tests
│
└── fixtures/         # Test data
    └── users.json
```

---

## Root Files

### main.go

Application entry point.

```go
package main

import (
    "neonexcore/internal/core"
    // Import your modules
)

func main() {
    // Initialize application
    app := core.NewApp()
    
    // Register custom modules (if any)
    // core.ModuleMap["blog"] = blog.New()
    
    // Run application
    app.Run()
}
```

### go.mod

Go module definition.

```go
module your-app

go 1.21

require (
    github.com/neonextechnologies/neonexframework v0.2.0
    github.com/gofiber/fiber/v2 v2.52.9
    gorm.io/gorm v1.25.11
)
```

### .env

Environment variables (not committed to git).

### .env.example

Environment template (committed to git).

### Makefile

Build commands.

```makefile
.PHONY: dev build test clean

dev:
	go run main.go

build:
	go build -o build/app main.go

test:
	go test ./...

clean:
	rm -rf build/
```

---

## Directory Best Practices

### 1. Module Organization

**✅ DO:**
```
modules/blog/
├── module.go
└── pkg/
    ├── models.go
    ├── service.go
    └── controller.go
```

**❌ DON'T:**
```
blog/               # Outside modules/
├── model.go
└── handler.go      # Mixed concerns
```

### 2. Configuration

**✅ DO:**
```go
// Use config package
dbHost := config.Get("DB_HOST")
```

**❌ DON'T:**
```go
// Hardcode values
dbHost := "localhost"
```

### 3. Static Files

**✅ DO:**
```
public/
├── css/app.css
└── js/app.js
```

**❌ DON'T:**
```
assets/            # Wrong location
static/            # Wrong location
```

### 4. Templates

**✅ DO:**
```html
{{template "layouts/main" .}}
```

**❌ DON'T:**
```html
<!-- Inline everything -->
```

---

## Adding New Modules

### Step 1: Create Structure

```bash
mkdir -p modules/shop/pkg
```

### Step 2: Create Module File

`modules/shop/module.go`:

```go
package shop

import (
    "github.com/gofiber/fiber/v2"
    "neonexcore/internal/core"
)

type ShopModule struct{}

func New() *ShopModule {
    return &ShopModule{}
}

func (m *ShopModule) Name() string {
    return "shop"
}

func (m *ShopModule) RegisterServices(c *core.Container) error {
    // Register services
    return nil
}

func (m *ShopModule) RegisterRoutes(r fiber.Router) error {
    // Register routes
    return nil
}

func (m *ShopModule) Boot() error {
    return nil
}
```

### Step 3: Register in main.go

```go
import "your-app/modules/shop"

func main() {
    app := core.NewApp()
    core.ModuleMap["shop"] = shop.New()
    app.Run()
}
```

---

## Workspace Setup

### VS Code

`.vscode/settings.json`:
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "editor.formatOnSave": true
}
```

### GoLand

Project automatically detected!

---

## Next Steps

- [First Application](./first-application.md) - Build your first app
- [Modules System](../core-concepts/modules.md) - Deep dive into modules
- [Database](../database/configuration.md) - Database setup
- [Testing](../testing/introduction.md) - Write tests

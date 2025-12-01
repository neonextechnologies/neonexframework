# NeonEx Framework Documentation

NeonEx Framework is a modern, full-stack Go framework built on [NeonEx Core](https://github.com/neonextechnologies/neonexcore). It provides everything you need to build production-ready web applications.

## Features

- 🚀 **High Performance** - 10,000+ requests/second with Fiber
- 🏗️ **Modular Architecture** - Self-contained modules with DI
- 🔐 **Built-in Security** - JWT authentication & RBAC authorization
- 💾 **Database Integration** - GORM with PostgreSQL, MySQL, SQLite
- 🎨 **Frontend Support** - Template engine and asset management
- 🌐 **Advanced Features** - WebSockets, GraphQL, caching, queues

## Quick Start

```go
package main

import "neonexcore/internal/core"

func main() {
    app := core.NewApp()
    app.Run()
}
```

Your app is now running on `http://localhost:8080` with database, auth, and routing configured

## Documentation

### Getting Started
- [Installation](getting-started/installation.md)
- [Quick Start](getting-started/quick-start.md)
- [Configuration](getting-started/configuration.md)
- [Project Structure](getting-started/project-structure.md)

### Learn More
- [Architecture](introduction/architecture.md)
- [Features](introduction/features.md)
- [Core Management](core-management.md)

### Resources
- [FAQ](resources/faq.md)
- [Contributing](contributing/guide.md)
- [Changelog](resources/changelog.md)

## Links

- [GitHub](https://github.com/neonextechnologies/neonexframework)
- [Discussions](https://github.com/neonextechnologies/neonexframework/discussions)
- [Issues](https://github.com/neonextechnologies/neonexframework/issues)

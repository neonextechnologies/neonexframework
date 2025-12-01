# What is NeonEx Framework?

NeonEx Framework is a modern, full-stack Go framework built on [NeonEx Core](https://github.com/neonextechnologies/neonexcore). It provides a complete foundation for building scalable web applications with convention over configuration.

## Philosophy

- **Convention over Configuration** - Sensible defaults, minimal setup
- **Modularity First** - Self-contained, reusable components
- **Developer Experience** - Intuitive APIs and clear documentation
- **Production Ready** - Security and performance built-in

## Architecture

```
┌─────────────────────────────────────────┐
│       Your Application                   │
├─────────────────────────────────────────┤
│         NeonEx Framework                 │
├─────────────────────────────────────────┤
│         NeonEx Core                      │
├─────────────────────────────────────────┤
│    Fiber + GORM + Redis + etc.          │
└─────────────────────────────────────────┘
```

## Core Components

### Module System
- Auto-discovery and dependency injection
- Lifecycle hooks and service registration

### HTTP Layer
- Fast routing with Fiber v2
- Middleware support and context management

### Database Layer
- GORM integration with multiple database support
- Repository pattern and query builder

### Security
- JWT tokens and RBAC authorization
- Password hashing and API keys

## Performance

- **10,000+ requests/second** on standard hardware
- **Sub-millisecond** response times
- **Low memory** footprint (~20MB)

## Use Cases

- RESTful APIs and GraphQL services
- Full-stack web applications
- Real-time applications with WebSockets
- Microservices architecture
- Admin panels and dashboards

## Next Steps

- [Installation](../getting-started/installation.md)
- [Quick Start](../getting-started/quick-start.md)
- [Architecture Deep Dive](./architecture.md)

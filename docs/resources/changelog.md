# Changelog

All notable changes to NeonEx Framework will be documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Planned
- GraphQL subscriptions
- Improved CLI tools
- More middleware options
- Plugin system

---

## [0.2.0] - 2025-12-01

### Added
- Web module with routing utilities
- Frontend module with template engine
- Asset management system
- Theme system for multiple themes
- Improved logging with structured output
- Core management scripts (update-core.sh, update-core.ps1)
- Comprehensive documentation

### Changed
- Framework repositioned as pure full-stack framework
- Removed CMS-specific features
- Removed E-commerce-specific features
- Updated to use neonexcore as dependency
- Improved module discovery
- Better error handling

### Removed
- CMS module (moved to separate package)
- E-commerce module (moved to separate package)
- Unnecessary core documentation
- Development scripts from core

### Fixed
- Module loading order issues
- Database connection pool configuration
- Template rendering bugs

---

## [0.1.0] - 2025-11-15

### Added
- Initial release
- Core framework foundation
- Module system with auto-discovery
- Dependency injection container
- Authentication & Authorization (JWT + RBAC)
- Database integration (GORM)
- Repository pattern
- HTTP routing (Fiber v2)
- Middleware support
- Configuration management
- Logging system
- Validation
- Error handling

### Core Modules
- Admin module
- User module
- Frontend module

### Features
- RESTful API support
- WebSocket support
- Caching (Redis integration)
- Email system
- File storage
- Queue system
- Testing utilities

---

## Release Notes

### Version 0.2.0 Highlights

This release marks a significant shift in NeonEx Framework's direction:

**Philosophy Change**: NeonEx is now a **pure full-stack framework** rather than a CMS-focused platform. This makes it more versatile for building any type of web application.

**Major Improvements**:
- Clean separation between core and framework
- Better documentation structure
- Improved developer experience
- More flexible architecture

**Migration Guide**: If upgrading from 0.1.0:
1. Remove CMS and E-commerce dependencies
2. Update imports to new structure
3. Review core management documentation
4. Update custom modules if needed

---

## [0.1.0] - Initial Release Notes

### What's Included

**Core Foundation**
- Complete module system
- Type-safe dependency injection
- Database ORM integration
- Authentication & authorization

**Developer Experience**
- Hot reload support
- Code generation tools
- Comprehensive testing
- Clear documentation

**Production Ready**
- High performance (10,000+ req/s)
- Security best practices
- Monitoring & metrics
- Docker support

---

## Upgrade Guides

### Upgrading to 0.2.0 from 0.1.0

#### Breaking Changes

1. **CMS Module Removed**
   ```go
   // Before
   import "neonexframework/modules/cms"
   
   // After - Use separate package
   import "github.com/neonextechnologies/neonex-cms"
   ```

2. **E-commerce Module Removed**
   ```go
   // Before
   import "neonexframework/modules/ecommerce"
   
   // After - Use separate package
   import "github.com/neonextechnologies/neonex-ecommerce"
   ```

3. **Module Registration**
   ```go
   // Still works the same
   core.ModuleMap["mymodule"] = mymodule.New()
   ```

#### New Features

1. **Web Module**
   ```go
   import "neonexframework/modules/web"
   
   // Use web utilities
   router := web.NewRouter()
   ```

2. **Core Management**
   ```bash
   # Update core easily
   make update-core
   ```

#### Migration Steps

1. Update dependencies
   ```bash
   go get -u github.com/neonextechnologies/neonexframework@v0.2.0
   go mod tidy
   ```

2. Remove CMS/E-commerce imports if using

3. Run tests
   ```bash
   go test ./...
   ```

4. Update documentation references

---

## Versioning Policy

NeonEx Framework follows [Semantic Versioning](https://semver.org/):

- **Major** (1.0.0): Breaking changes
- **Minor** (0.1.0): New features, backward compatible
- **Patch** (0.0.1): Bug fixes

### Release Schedule

- **Major releases**: Yearly
- **Minor releases**: Every 2-3 months
- **Patch releases**: As needed

---

## Deprecation Policy

Features are deprecated in three stages:

1. **Announcement**: Deprecation notice in release notes
2. **Warning Period**: 2 minor versions with deprecation warnings
3. **Removal**: Removed in next major version

---

## Future Roadmap

### Version 0.3.0 (Q1 2026)
- [ ] Enhanced CLI tools
- [ ] Plugin system
- [ ] Improved GraphQL support
- [ ] WebSocket enhancements
- [ ] More middleware options

### Version 0.4.0 (Q2 2026)
- [ ] Admin panel generator
- [ ] API documentation generator
- [ ] Database seeding improvements
- [ ] Multi-tenancy support

### Version 1.0.0 (Q4 2026)
- [ ] Stable API
- [ ] Long-term support (LTS)
- [ ] Enterprise features
- [ ] Professional support packages

---

## Community Feedback

We value your feedback! Let us know:

- What features you'd like to see
- What bugs you've encountered
- How we can improve

**Channels**:
- [GitHub Discussions](https://github.com/neonextechnologies/neonexframework/discussions)
- [GitHub Issues](https://github.com/neonextechnologies/neonexframework/issues)
- Email: feedback@neonexframework.dev

---

## Contributors

Thank you to all contributors! See [Contributors](https://github.com/neonextechnologies/neonexframework/graphs/contributors).

---

[Unreleased]: https://github.com/neonextechnologies/neonexframework/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/neonextechnologies/neonexframework/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/neonextechnologies/neonexframework/releases/tag/v0.1.0

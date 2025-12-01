# Upgrade Guide

Guide for upgrading between NeonEx Framework versions.

---

## Upgrade Path

### Current Version: 0.2.0

This guide covers:
- Upgrading from 0.1.0 to 0.2.0
- Future upgrade paths
- Breaking changes
- Migration steps

---

## Upgrading to 0.2.0 from 0.1.0

### Overview

Version 0.2.0 introduces significant architectural changes:
- Removed CMS module
- Removed E-commerce module
- Added Web module
- Added Frontend module improvements
- Repositioned as pure full-stack framework

### Breaking Changes

#### 1. CMS Module Removed

**Before (0.1.0):**
```go
import "neonexframework/modules/cms"

core.ModuleMap["cms"] = cms.New()
```

**After (0.2.0):**
```go
// CMS moved to separate package
// Install: go get github.com/neonextechnologies/neonex-cms
import "github.com/neonextechnologies/neonex-cms"

core.ModuleMap["cms"] = cms.New()
```

**Migration:**
```bash
# 1. Install separate CMS package
go get github.com/neonextechnologies/neonex-cms

# 2. Update imports in your code
# Replace: neonexframework/modules/cms
# With: github.com/neonextechnologies/neonex-cms

# 3. Test your application
go test ./...
```

#### 2. E-commerce Module Removed

**Before (0.1.0):**
```go
import "neonexframework/modules/ecommerce"

core.ModuleMap["ecommerce"] = ecommerce.New()
```

**After (0.2.0):**
```go
// E-commerce moved to separate package
// Install: go get github.com/neonextechnologies/neonex-ecommerce
import "github.com/neonextechnologies/neonex-ecommerce"

core.ModuleMap["ecommerce"] = ecommerce.New()
```

#### 3. New Web Module

**New in 0.2.0:**
```go
import "neonexframework/modules/web"

// Web module provides routing utilities
core.ModuleMap["web"] = web.New()
```

### Step-by-Step Migration

#### Step 1: Backup

```bash
# Backup your project
cp -r my-project my-project-backup

# Or commit current state
git add .
git commit -m "Backup before upgrading to 0.2.0"
```

#### Step 2: Update Dependencies

```bash
# Update NeonEx Framework
go get -u github.com/neonextechnologies/neonexframework@v0.2.0

# Clean up
go mod tidy
```

#### Step 3: Update Imports

Find and replace in your project:

```bash
# Using grep (Linux/Mac)
grep -r "neonexframework/modules/cms" .
grep -r "neonexframework/modules/ecommerce" .

# Using PowerShell (Windows)
Get-ChildItem -Recurse -Include *.go | Select-String "neonexframework/modules/cms"
```

Replace:
- `neonexframework/modules/cms` → `github.com/neonextechnologies/neonex-cms`
- `neonexframework/modules/ecommerce` → `github.com/neonextechnologies/neonex-ecommerce`

#### Step 4: Install Removed Modules (if needed)

```bash
# If you use CMS
go get github.com/neonextechnologies/neonex-cms

# If you use E-commerce
go get github.com/neonextechnologies/neonex-ecommerce
```

#### Step 5: Update Module Registration

`main.go`:

```go
package main

import (
    "neonexcore/internal/core"
    
    // If using CMS
    cms "github.com/neonextechnologies/neonex-cms"
    
    // If using E-commerce
    ecommerce "github.com/neonextechnologies/neonex-ecommerce"
    
    // Your custom modules
    "your-app/modules/blog"
)

func main() {
    app := core.NewApp()
    
    // Register CMS (if using)
    core.ModuleMap["cms"] = cms.New()
    
    // Register E-commerce (if using)
    core.ModuleMap["ecommerce"] = ecommerce.New()
    
    // Register custom modules
    core.ModuleMap["blog"] = blog.New()
    
    app.Run()
}
```

#### Step 6: Test

```bash
# Run tests
go test ./...

# Build
go build

# Run application
go run main.go
```

#### Step 7: Update Documentation

Update your project's documentation to reflect:
- New version number
- Updated dependencies
- Module changes

### Configuration Changes

#### No Breaking Changes

`.env` file format remains the same. No changes needed.

### Database Changes

#### No Breaking Changes

Database schema remains compatible. No migrations needed.

---

## Upgrading to Future Versions

### Semantic Versioning

NeonEx follows semantic versioning:

- **Patch** (0.2.1): Bug fixes only, no breaking changes
- **Minor** (0.3.0): New features, backward compatible
- **Major** (1.0.0): Breaking changes

### Patch Updates (0.2.x)

Safe to update immediately:

```bash
go get -u github.com/neonextechnologies/neonexframework@v0.2.1
go mod tidy
```

No code changes needed.

### Minor Updates (0.3.0)

Check release notes for:
- New features
- Deprecation warnings
- Optional updates

```bash
# Update
go get -u github.com/neonextechnologies/neonexframework@v0.3.0

# Test
go test ./...
```

### Major Updates (1.0.0)

Read upgrade guide carefully:
- Breaking changes
- Migration steps
- Deprecated feature removal

---

## Common Issues

### Issue: Module not found

**Problem:**
```
package neonexframework/modules/cms is not in GOROOT
```

**Solution:**
```bash
# Install separate CMS package
go get github.com/neonextechnologies/neonex-cms

# Update imports in code
```

### Issue: Compilation errors after upgrade

**Problem:**
```
undefined: cms.New
```

**Solution:**
```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod download

# Tidy up
go mod tidy
```

### Issue: Tests failing after upgrade

**Problem:**
Some tests fail after upgrade.

**Solution:**
1. Check if you're using removed modules
2. Update test imports
3. Review breaking changes
4. Update test assertions if needed

---

## Rollback

If upgrade fails, rollback:

```bash
# Restore from backup
rm -rf my-project
cp -r my-project-backup my-project

# Or use git
git reset --hard HEAD~1

# Or specify version
go get github.com/neonextechnologies/neonexframework@v0.1.0
go mod tidy
```

---

## Version Support

### Support Policy

- **Latest version**: Full support
- **One version back**: Security fixes only
- **Older versions**: No support

**Example:**
- Current: 0.2.0 (full support)
- Previous: 0.1.0 (security fixes only)
- Older: No support

### Long-Term Support (LTS)

Starting with version 1.0.0:
- **LTS versions**: 2 years support
- **Regular versions**: 6 months support

---

## Staying Updated

### Subscribe to Updates

- **GitHub Releases**: [Watch Repository](https://github.com/neonextechnologies/neonexframework)
- **Newsletter**: [Subscribe](https://neonexframework.dev/newsletter) *(coming soon)*
- **Twitter**: [@NeonExFramework](https://twitter.com/NeonExFramework) *(coming soon)*

### Check for Updates

```bash
# Check current version
go list -m github.com/neonextechnologies/neonexframework

# Check available versions
go list -m -versions github.com/neonextechnologies/neonexframework
```

---

## Best Practices

### 1. Test Before Upgrading

```bash
# Run full test suite
go test ./...

# Test in staging environment
APP_ENV=staging go run main.go
```

### 2. Read Release Notes

Always read:
- [Changelog](./changelog.md)
- Release notes on GitHub
- Breaking changes section

### 3. Upgrade in Stages

Don't jump versions:
- 0.1.0 → 0.2.0 ✅
- 0.1.0 → 0.3.0 ❌ (upgrade to 0.2.0 first)

### 4. Backup First

Always backup before upgrading:
```bash
# Database backup
pg_dump mydb > backup.sql

# Code backup
git commit -am "Backup before upgrade"
```

### 5. Update Dependencies

After framework upgrade:
```bash
# Update all dependencies
go get -u ./...
go mod tidy
```

---

## Getting Help

Need help upgrading?

- **Documentation**: [Upgrade Guide](./upgrade-guide.md)
- **Discussions**: [GitHub Discussions](https://github.com/neonextechnologies/neonexframework/discussions)
- **Issues**: [GitHub Issues](https://github.com/neonextechnologies/neonexframework/issues)
- **Email**: support@neonexframework.dev

---

## Deprecation Warnings

### Currently Deprecated

*No deprecated features in 0.2.0*

### Future Deprecations

Watch for deprecation warnings in:
- Release notes
- Console output
- Documentation

**Example warning:**
```
DEPRECATED: OldFunction() is deprecated and will be removed in v1.0.0
Use NewFunction() instead.
```

---

## Next Steps

After upgrading:

1. ✅ Test thoroughly
2. ✅ Update documentation
3. ✅ Deploy to staging
4. ✅ Monitor for issues
5. ✅ Deploy to production

---

## Questions?

- 💬 [GitHub Discussions](https://github.com/neonextechnologies/neonexframework/discussions)
- 📧 Email: support@neonexframework.dev
- 📖 [Documentation](../README.md)

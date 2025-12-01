# NeonEx Core

> Core library for NeonEx Framework

This is a cleaned version of [neonexcore](https://github.com/neonextechnologies/neonexcore) used as a dependency for NeonEx Framework.

## What's Included

- ✅ Core framework components
- ✅ Module system with DI
- ✅ Database utilities (GORM)
- ✅ Authentication & RBAC
- ✅ Logging system (Zap)
- ✅ User & Admin modules
- ✅ API utilities

## What's Removed

- ❌ Documentation (use NeonEx Framework docs)
- ❌ Examples (see NeonEx Framework examples)
- ❌ .git directory (managed by framework)
- ❌ .gitbook.yaml, .air.toml, dev.sh
- ❌ .env.example, LICENSE, Makefile

## Directory Structure

```
core/
├── internal/           # Internal packages
│   ├── config/        # Configuration
│   └── core/          # Core framework
├── modules/           # Built-in modules
│   ├── user/         # User module
│   └── admin/        # Admin module
├── pkg/              # Public packages
│   ├── database/     # Database utilities
│   ├── logger/       # Logging
│   ├── auth/         # Authentication
│   ├── rbac/         # Authorization
│   └── ...
├── go.mod
└── main.go
```

## Updating Core

When a new version of neonexcore is released:

### Method 1: Manual Update (Recommended)

```bash
# 1. Backup current core
cd d:\go\neonexframework
mv core core.backup

# 2. Clone fresh neonexcore
git clone https://github.com/neonextechnologies/neonexcore.git core

# 3. Clean unnecessary files
cd core
rm -rf docs examples .gitbook.yaml .git README.md

# 4. Create this README
# (Copy this file back)

# 5. Test
cd ..
go mod tidy
go test ./...

# 6. If OK, remove backup
rm -rf core.backup
```

### Method 2: Script Update

Create `scripts/update-core.sh`:

```bash
#!/bin/bash
cd "$(dirname "$0")/.."

# Backup
echo "Backing up core..."
mv core core.backup

# Clone
echo "Cloning neonexcore..."
git clone https://github.com/neonextechnologies/neonexcore.git core

# Clean
echo "Cleaning core..."
cd core
rm -rf docs examples .gitbook.yaml .git
rm -f .air.toml dev.sh .env.example LICENSE Makefile
cp ../core.backup/README.md ./README.md
cd ..

# Test
echo "Testing..."
go mod tidy
if go test ./...; then
    echo "✅ Update successful!"
    rm -rf core.backup
else
    echo "❌ Tests failed. Restoring backup..."
    rm -rf core
    mv core.backup core
fi
```

## Version Tracking

Current core version: **v0.1.0** (2024-12-01)

To check for updates:
- Visit: https://github.com/neonextechnologies/neonexcore
- Check releases: https://github.com/neonextechnologies/neonexcore/releases

## Important Notes

⚠️ **Do NOT modify files in this directory directly**

- Core is a dependency, not part of NeonEx Framework
- Modifications will be lost on update
- Use wrapper packages in `neonexframework/pkg/` instead

## Full Documentation

For complete neonexcore documentation, visit:
- GitHub: https://github.com/neonextechnologies/neonexcore
- Docs: See NeonEx Framework documentation

## License

MIT License - Same as neonexcore

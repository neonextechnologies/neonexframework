#!/bin/bash
# Update NeonEx Core Script
# This script updates the core dependency from neonexcore repository

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CORE_DIR="$PROJECT_ROOT/core"
BACKUP_DIR="$PROJECT_ROOT/core.backup"
CORE_REPO="https://github.com/neonextechnologies/neonexcore.git"

echo "=================================="
echo "NeonEx Core Update Script"
echo "=================================="
echo ""

# Check if core exists
if [ ! -d "$CORE_DIR" ]; then
    echo "❌ Core directory not found!"
    exit 1
fi

# Backup current core
echo "📦 Backing up current core..."
if [ -d "$BACKUP_DIR" ]; then
    rm -rf "$BACKUP_DIR"
fi
mv "$CORE_DIR" "$BACKUP_DIR"
echo "✅ Backup created at core.backup"
echo ""

# Clone fresh neonexcore
echo "📥 Cloning neonexcore..."
git clone "$CORE_REPO" "$CORE_DIR"
echo "✅ Clone complete"
echo ""

# Clean unnecessary files
echo "🧹 Cleaning core..."
cd "$CORE_DIR"
rm -rf docs examples .gitbook.yaml .git README.md LICENSE
rm -f .air.toml dev.sh .env.example Makefile
echo "✅ Cleaned: docs, examples, .gitbook.yaml, .git, and build files"
echo ""

# Create new README
echo "📝 Creating README..."
cat > README.md << 'EOF'
# NeonEx Core

> Core library for NeonEx Framework

This is a cleaned version of neonexcore. See ../core/README.md for update instructions.

Version: Updated $(date +%Y-%m-%d)
Source: https://github.com/neonextechnologies/neonexcore
EOF
echo "✅ README created"
echo ""

# Go back to project root
cd "$PROJECT_ROOT"

# Update dependencies
echo "📦 Updating Go modules..."
go mod tidy
echo "✅ Dependencies updated"
echo ""

# Run tests
echo "🧪 Running tests..."
if go test ./... -v; then
    echo ""
    echo "✅ All tests passed!"
    echo ""
    echo "🎉 Core update successful!"
    echo ""
    
    # Ask to remove backup
    read -p "Remove backup? (y/n) " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$BACKUP_DIR"
        echo "✅ Backup removed"
    else
        echo "ℹ️  Backup kept at core.backup"
    fi
else
    echo ""
    echo "❌ Tests failed! Restoring backup..."
    rm -rf "$CORE_DIR"
    mv "$BACKUP_DIR" "$CORE_DIR"
    echo "✅ Backup restored"
    echo ""
    echo "❌ Update failed. Please check the errors above."
    exit 1
fi

echo ""
echo "=================================="
echo "Update complete!"
echo "=================================="

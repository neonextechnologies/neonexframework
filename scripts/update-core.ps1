# Update NeonEx Core Script (PowerShell)
# This script updates the core dependency from neonexcore repository

$ErrorActionPreference = "Stop"

$ProjectRoot = Split-Path -Parent $PSScriptRoot
$CoreDir = Join-Path $ProjectRoot "core"
$BackupDir = Join-Path $ProjectRoot "core.backup"
$CoreRepo = "https://github.com/neonextechnologies/neonexcore.git"

Write-Host "==================================" -ForegroundColor Cyan
Write-Host "NeonEx Core Update Script" -ForegroundColor Cyan
Write-Host "==================================" -ForegroundColor Cyan
Write-Host ""

# Check if core exists
if (-not (Test-Path $CoreDir)) {
    Write-Host "❌ Core directory not found!" -ForegroundColor Red
    exit 1
}

# Backup current core
Write-Host "📦 Backing up current core..." -ForegroundColor Yellow
if (Test-Path $BackupDir) {
    Remove-Item -Recurse -Force $BackupDir
}
Move-Item $CoreDir $BackupDir
Write-Host "✅ Backup created at core.backup" -ForegroundColor Green
Write-Host ""

# Clone fresh neonexcore
Write-Host "📥 Cloning neonexcore..." -ForegroundColor Yellow
try {
    git clone $CoreRepo $CoreDir
    Write-Host "✅ Clone complete" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "❌ Clone failed! Restoring backup..." -ForegroundColor Red
    Move-Item $BackupDir $CoreDir
    exit 1
}

# Clean unnecessary files
Write-Host "🧹 Cleaning core..." -ForegroundColor Yellow
Push-Location $CoreDir
Remove-Item -Recurse -Force -ErrorAction SilentlyContinue docs,examples,.gitbook.yaml,.git,README.md,LICENSE
Remove-Item -Force -ErrorAction SilentlyContinue .air.toml,dev.sh,.env.example,Makefile
Write-Host "✅ Cleaned: docs, examples, .gitbook.yaml, .git, and build files" -ForegroundColor Green
Write-Host ""

# Create new README
Write-Host "📝 Creating README..." -ForegroundColor Yellow
$readmeContent = @"
# NeonEx Core

> Core library for NeonEx Framework

This is a cleaned version of neonexcore. See core/README.md in parent for update instructions.

Version: Updated $(Get-Date -Format "yyyy-MM-dd")
Source: https://github.com/neonextechnologies/neonexcore
"@
Set-Content -Path "README.md" -Value $readmeContent
Write-Host "✅ README created" -ForegroundColor Green
Write-Host ""

Pop-Location

# Update dependencies
Write-Host "📦 Updating Go modules..." -ForegroundColor Yellow
Push-Location $ProjectRoot
go mod tidy
Write-Host "✅ Dependencies updated" -ForegroundColor Green
Write-Host ""

# Run tests
Write-Host "🧪 Running tests..." -ForegroundColor Yellow
$testResult = go test ./... -v
if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ All tests passed!" -ForegroundColor Green
    Write-Host ""
    Write-Host "🎉 Core update successful!" -ForegroundColor Green
    Write-Host ""
    
    # Ask to remove backup
    $response = Read-Host "Remove backup? (y/n)"
    if ($response -eq "y" -or $response -eq "Y") {
        Remove-Item -Recurse -Force $BackupDir
        Write-Host "✅ Backup removed" -ForegroundColor Green
    } else {
        Write-Host "ℹ️  Backup kept at core.backup" -ForegroundColor Cyan
    }
} else {
    Write-Host ""
    Write-Host "❌ Tests failed! Restoring backup..." -ForegroundColor Red
    Remove-Item -Recurse -Force $CoreDir
    Move-Item $BackupDir $CoreDir
    Write-Host "✅ Backup restored" -ForegroundColor Green
    Write-Host ""
    Write-Host "❌ Update failed. Please check the errors above." -ForegroundColor Red
    exit 1
}

Pop-Location

Write-Host ""
Write-Host "==================================" -ForegroundColor Cyan
Write-Host "Update complete!" -ForegroundColor Cyan
Write-Host "==================================" -ForegroundColor Cyan

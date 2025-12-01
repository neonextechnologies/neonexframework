# Installation

## Prerequisites

Before installing NeonEx Framework, ensure you have the following:

### Required

- **Go 1.21 or higher** - [Download Go](https://go.dev/dl/)
- **Git** - For version control and cloning repositories

### Recommended

- **PostgreSQL 14+** - Primary database (or MySQL 8+, SQLite)
- **Redis** - For caching (optional but recommended)
- **Make** - For using Makefile commands (optional)

### System Requirements

- **OS**: Windows, Linux, macOS
- **RAM**: Minimum 2GB, recommended 4GB+
- **Disk**: 500MB free space

---

## Verify Prerequisites

### Check Go Installation

```bash
go version
```

Expected output:
```
go version go1.21.0 windows/amd64
```

If Go is not installed, download from [golang.org](https://go.dev/dl/)

### Check Git Installation

```bash
git --version
```

Expected output:
```
git version 2.40.0
```

---

## Installation Methods

### Method 1: Clone from GitHub (Recommended)

This method gives you the full framework with examples and documentation.

```bash
# Clone the repository
git clone https://github.com/neonextechnologies/neonexframework.git

# Navigate to directory
cd neonexframework

# Install dependencies
go mod download

# Verify installation
go run main.go --version
```

### Method 2: Use as Go Module

For starting a new project from scratch:

```bash
# Create new directory
mkdir my-neonex-app
cd my-neonex-app

# Initialize Go module
go mod init my-neonex-app

# Install NeonEx Framework
go get github.com/neonextechnologies/neonexframework

# Create main.go
cat > main.go << 'EOF'
package main

import "neonexcore/internal/core"

func main() {
    app := core.NewApp()
    app.Run()
}
EOF

# Run application
go run main.go
```

### Method 3: Use Template (Coming Soon)

```bash
# Using our project template
git clone https://github.com/neonextechnologies/neonex-template.git my-app
cd my-app
go mod download
```

---

## Database Setup

### PostgreSQL Setup

#### Windows

```powershell
# Using Chocolatey
choco install postgresql

# Or download installer from
# https://www.postgresql.org/download/windows/

# Create database
psql -U postgres
CREATE DATABASE neonexdb;
\q
```

#### Linux (Ubuntu/Debian)

```bash
# Install PostgreSQL
sudo apt update
sudo apt install postgresql postgresql-contrib

# Start service
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create database
sudo -u postgres psql
CREATE DATABASE neonexdb;
\q
```

#### macOS

```bash
# Using Homebrew
brew install postgresql@14
brew services start postgresql@14

# Create database
createdb neonexdb
```

### MySQL Setup (Alternative)

```bash
# Ubuntu/Debian
sudo apt install mysql-server

# macOS
brew install mysql

# Windows
choco install mysql
```

### SQLite (Development)

No installation needed! SQLite is embedded.

---

## Redis Setup (Optional)

Redis is optional but recommended for caching.

### Windows

```powershell
# Using Chocolatey
choco install redis-64

# Or use Docker
docker run -d -p 6379:6379 redis:alpine
```

### Linux

```bash
# Ubuntu/Debian
sudo apt install redis-server
sudo systemctl start redis
sudo systemctl enable redis
```

### macOS

```bash
brew install redis
brew services start redis
```

---

## Environment Configuration

Create `.env` file in project root:

```bash
# Copy example environment file
cp .env.example .env
```

Edit `.env` with your configuration:

```env
# Application
APP_NAME=NeonExApp
APP_ENV=development
APP_DEBUG=true
APP_URL=http://localhost:8080
APP_PORT=8080

# Database
DB_CONNECTION=postgres
DB_HOST=localhost
DB_PORT=5432
DB_DATABASE=neonexdb
DB_USERNAME=postgres
DB_PASSWORD=your_password

# Redis (optional)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT
JWT_SECRET=your-secret-key-change-this-in-production
JWT_EXPIRATION=24h

# Email (optional)
MAIL_DRIVER=smtp
MAIL_HOST=smtp.gmail.com
MAIL_PORT=587
MAIL_USERNAME=your-email@gmail.com
MAIL_PASSWORD=your-app-password
```

---

## Running the Application

### Development Mode

```bash
# Run with hot reload
make dev

# Or without Make
go run main.go
```

The application will start on `http://localhost:8080`

### Production Mode

```bash
# Build binary
make build

# Run binary
./build/neonex

# Or on Windows
.\build\neonex.exe
```

---

## Verify Installation

### Check Health Endpoint

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "database": "connected",
  "cache": "connected"
}
```

### Check API

```bash
curl http://localhost:8080/api/v1/status
```

### Access Web Interface

Open browser: `http://localhost:8080`

---

## Docker Installation

### Using Docker Compose (Recommended)

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=db
      - DB_DATABASE=neonexdb
      - DB_USERNAME=postgres
      - DB_PASSWORD=secret
      - REDIS_HOST=redis
    depends_on:
      - db
      - redis

  db:
    image: postgres:14-alpine
    environment:
      - POSTGRES_DB=neonexdb
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=secret
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop services
docker-compose down
```

### Using Docker Only

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/.env .

EXPOSE 8080
CMD ["./main"]
```

```bash
# Build image
docker build -t neonex-app .

# Run container
docker run -p 8080:8080 neonex-app
```

---

## IDE Setup

### VS Code

Recommended extensions:

```json
{
  "recommendations": [
    "golang.go",
    "ms-azuretools.vscode-docker",
    "streetsidesoftware.code-spell-checker"
  ]
}
```

### GoLand

1. Open project
2. Go to **Settings → Go → GOPATH**
3. Add project directory
4. Run configuration will be auto-detected

---

## Common Issues

### Issue: "command not found: go"

**Solution**: Go is not installed or not in PATH
```bash
# Add Go to PATH (Linux/Mac)
export PATH=$PATH:/usr/local/go/bin

# Windows: Add to System Environment Variables
```

### Issue: "database connection failed"

**Solution**: Check database credentials in `.env`
```bash
# Test PostgreSQL connection
psql -h localhost -U postgres -d neonexdb
```

### Issue: "port 8080 already in use"

**Solution**: Change port in `.env` or stop other services
```bash
# Find process using port 8080
# Linux/Mac
lsof -i :8080

# Windows
netstat -ano | findstr :8080

# Kill process
kill <PID>
```

### Issue: "go.mod not found"

**Solution**: Initialize Go module
```bash
go mod init your-project-name
go mod tidy
```

---

## Updating NeonEx Framework

### Update to Latest Version

```bash
# Using git
git pull origin main
go mod download
go mod tidy

# Or update module
go get -u github.com/neonextechnologies/neonexframework
```

### Update Core Dependency

```bash
# Using provided script
make update-core

# Or manually (see Core Management docs)
```

---

## Next Steps

Now that you have NeonEx installed:

1. 📚 [Quick Start Tutorial](./quick-start.md) - Build your first app
2. ⚙️ [Configuration Guide](./configuration.md) - Configure your app
3. 🏗️ [Project Structure](./project-structure.md) - Understand the layout
4. 🚀 [First Application](./first-application.md) - Create a complete app

---

## Getting Help

If you encounter issues:

- 📖 [Documentation](../README.md)
- 💬 [GitHub Discussions](https://github.com/neonextechnologies/neonexframework/discussions)
- 🐛 [GitHub Issues](https://github.com/neonextechnologies/neonexframework/issues)
- 📧 Email: support@neonexframework.dev

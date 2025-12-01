# Configuration

Learn how to configure your NeonEx Framework application.

---

## Environment Configuration

NeonEx uses environment variables for configuration. All settings are stored in the `.env` file.

### Creating .env File

```bash
# Copy the example
cp .env.example .env
```

### Complete .env Reference

```env
# ============================================
# Application Settings
# ============================================
APP_NAME=NeonExApp
APP_ENV=development          # development, staging, production
APP_DEBUG=true              # Enable debug mode
APP_URL=http://localhost:8080
APP_PORT=8080
APP_TIMEZONE=UTC

# ============================================
# Database Configuration
# ============================================
DB_CONNECTION=postgres      # postgres, mysql, sqlite
DB_HOST=localhost
DB_PORT=5432
DB_DATABASE=neonexdb
DB_USERNAME=postgres
DB_PASSWORD=secret
DB_SSL_MODE=disable        # require, verify-full, verify-ca

# Connection Pool
DB_MAX_IDLE_CONNS=10
DB_MAX_OPEN_CONNS=100
DB_CONN_MAX_LIFETIME=3600  # seconds

# ============================================
# Redis Configuration (Optional)
# ============================================
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# ============================================
# JWT Authentication
# ============================================
JWT_SECRET=your-256-bit-secret-key-change-in-production
JWT_EXPIRATION=24h         # 24h, 7d, etc.
JWT_REFRESH_EXPIRATION=168h  # 7 days

# ============================================
# Email Configuration
# ============================================
MAIL_DRIVER=smtp           # smtp, sendgrid, mailgun
MAIL_HOST=smtp.gmail.com
MAIL_PORT=587
MAIL_USERNAME=your-email@gmail.com
MAIL_PASSWORD=your-app-password
MAIL_ENCRYPTION=tls        # tls, ssl
MAIL_FROM_ADDRESS=noreply@example.com
MAIL_FROM_NAME=NeonEx App

# ============================================
# File Storage
# ============================================
STORAGE_DRIVER=local       # local, s3, gcs
STORAGE_PATH=./storage

# AWS S3 (if using)
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
AWS_DEFAULT_REGION=us-east-1
AWS_BUCKET=
AWS_URL=

# ============================================
# Logging
# ============================================
LOG_LEVEL=info             # debug, info, warn, error
LOG_OUTPUT=stdout          # stdout, file, both
LOG_FILE_PATH=./storage/logs/app.log

# ============================================
# CORS Configuration
# ============================================
CORS_ALLOWED_ORIGINS=*
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=*
CORS_EXPOSED_HEADERS=
CORS_MAX_AGE=3600

# ============================================
# Rate Limiting
# ============================================
RATE_LIMIT_ENABLED=true
RATE_LIMIT_MAX_REQUESTS=100  # per window
RATE_LIMIT_WINDOW=60         # seconds

# ============================================
# Session Configuration
# ============================================
SESSION_DRIVER=cookie      # cookie, redis
SESSION_LIFETIME=120       # minutes
SESSION_SECURE=false       # true in production
SESSION_HTTP_ONLY=true
SESSION_SAME_SITE=lax      # lax, strict, none

# ============================================
# Cache Configuration
# ============================================
CACHE_DRIVER=redis         # memory, redis
CACHE_TTL=3600            # default TTL in seconds

# ============================================
# Queue Configuration
# ============================================
QUEUE_DRIVER=redis         # redis, database
QUEUE_CONNECTION=default

# ============================================
# Frontend Configuration
# ============================================
THEME=default
ASSET_VERSION=1.0.0        # For cache busting
MINIFY_ASSETS=true         # Minify CSS/JS in production

# ============================================
# External Services
# ============================================
# Google OAuth
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URL=

# Facebook OAuth
FACEBOOK_CLIENT_ID=
FACEBOOK_CLIENT_SECRET=
FACEBOOK_REDIRECT_URL=

# Stripe Payment
STRIPE_PUBLIC_KEY=
STRIPE_SECRET_KEY=
STRIPE_WEBHOOK_SECRET=

# SendGrid
SENDGRID_API_KEY=

# ============================================
# Monitoring & Analytics
# ============================================
SENTRY_DSN=                # Error tracking
PROMETHEUS_ENABLED=false   # Metrics
ANALYTICS_ENABLED=false

# ============================================
# Development Tools
# ============================================
HOT_RELOAD=true
PROFILING_ENABLED=false
```

---

## Configuration Loading

### Priority Order

NeonEx loads configuration in this order (later overrides earlier):

1. Default values (hardcoded)
2. `.env` file
3. Environment variables
4. Command-line flags (if supported)

### Accessing Configuration

```go
import "neonexcore/pkg/config"

// Get string value
dbHost := config.Get("DB_HOST")

// Get with default
dbHost := config.GetWithDefault("DB_HOST", "localhost")

// Get as int
port := config.GetInt("APP_PORT")

// Get as bool
debug := config.GetBool("APP_DEBUG")

// Get as duration
timeout := config.GetDuration("JWT_EXPIRATION")
```

---

## Environment-Specific Configuration

### Development Environment

`.env.development`:
```env
APP_ENV=development
APP_DEBUG=true
DB_HOST=localhost
LOG_LEVEL=debug
```

### Production Environment

`.env.production`:
```env
APP_ENV=production
APP_DEBUG=false
DB_HOST=production-db.example.com
LOG_LEVEL=warn
RATE_LIMIT_ENABLED=true
MINIFY_ASSETS=true
SESSION_SECURE=true
```

### Load Specific Environment

```bash
# Development
APP_ENV=development go run main.go

# Production
APP_ENV=production ./build/app
```

---

## Database Configuration

### PostgreSQL

```env
DB_CONNECTION=postgres
DB_HOST=localhost
DB_PORT=5432
DB_DATABASE=neonexdb
DB_USERNAME=postgres
DB_PASSWORD=secret
DB_SSL_MODE=disable
```

### MySQL

```env
DB_CONNECTION=mysql
DB_HOST=localhost
DB_PORT=3306
DB_DATABASE=neonexdb
DB_USERNAME=root
DB_PASSWORD=secret
```

### SQLite

```env
DB_CONNECTION=sqlite
DB_DATABASE=./storage/database.sqlite
```

### Connection Pool Settings

```env
# Recommended settings
DB_MAX_IDLE_CONNS=10        # Min connections
DB_MAX_OPEN_CONNS=100       # Max connections
DB_CONN_MAX_LIFETIME=3600   # 1 hour
```

---

## Security Configuration

### JWT Settings

```env
# Generate strong secret
JWT_SECRET=$(openssl rand -base64 32)

# Token expiration
JWT_EXPIRATION=24h          # Access token: 24 hours
JWT_REFRESH_EXPIRATION=168h # Refresh token: 7 days
```

### CORS Configuration

**Development (Allow All):**
```env
CORS_ALLOWED_ORIGINS=*
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=*
```

**Production (Specific Domains):**
```env
CORS_ALLOWED_ORIGINS=https://example.com,https://app.example.com
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE
CORS_ALLOWED_HEADERS=Content-Type,Authorization
```

### Rate Limiting

```env
RATE_LIMIT_ENABLED=true
RATE_LIMIT_MAX_REQUESTS=100   # 100 requests
RATE_LIMIT_WINDOW=60          # per minute
```

---

## Caching Configuration

### Redis Cache

```env
CACHE_DRIVER=redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
CACHE_TTL=3600  # 1 hour default
```

### Memory Cache

```env
CACHE_DRIVER=memory
CACHE_TTL=3600
```

---

## Email Configuration

### Gmail SMTP

```env
MAIL_DRIVER=smtp
MAIL_HOST=smtp.gmail.com
MAIL_PORT=587
MAIL_USERNAME=your-email@gmail.com
MAIL_PASSWORD=your-app-password  # Use App Password, not regular password
MAIL_ENCRYPTION=tls
```

### SendGrid

```env
MAIL_DRIVER=sendgrid
SENDGRID_API_KEY=your-api-key
MAIL_FROM_ADDRESS=noreply@example.com
```

---

## File Storage Configuration

### Local Storage

```env
STORAGE_DRIVER=local
STORAGE_PATH=./storage/uploads
```

### AWS S3

```env
STORAGE_DRIVER=s3
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_DEFAULT_REGION=us-east-1
AWS_BUCKET=your-bucket-name
AWS_URL=https://your-bucket.s3.amazonaws.com
```

---

## Logging Configuration

### Console Logging

```env
LOG_OUTPUT=stdout
LOG_LEVEL=info
```

### File Logging

```env
LOG_OUTPUT=file
LOG_FILE_PATH=./storage/logs/app.log
LOG_LEVEL=info
```

### Both Console and File

```env
LOG_OUTPUT=both
LOG_FILE_PATH=./storage/logs/app.log
LOG_LEVEL=debug
```

### Log Levels

- `debug` - Detailed information
- `info` - General information
- `warn` - Warning messages
- `error` - Error messages only

---

## Production Checklist

Before deploying to production:

### Security
- [ ] Change `JWT_SECRET` to strong random value
- [ ] Set `APP_DEBUG=false`
- [ ] Enable `SESSION_SECURE=true`
- [ ] Configure specific CORS origins
- [ ] Enable rate limiting
- [ ] Use strong database password

### Performance
- [ ] Enable `MINIFY_ASSETS=true`
- [ ] Configure Redis for caching
- [ ] Optimize database connection pool
- [ ] Enable compression

### Monitoring
- [ ] Configure error tracking (Sentry)
- [ ] Enable application logs
- [ ] Set up metrics (Prometheus)

### Environment
- [ ] Set `APP_ENV=production`
- [ ] Configure production database
- [ ] Set up backup strategy
- [ ] Configure email service

---

## Custom Configuration

### Adding Custom Config

Create `config/custom.go`:

```go
package config

type CustomConfig struct {
    FeatureFlag bool   `env:"FEATURE_FLAG" envDefault:"false"`
    APIKey      string `env:"API_KEY" envDefault:""`
    MaxUpload   int    `env:"MAX_UPLOAD_SIZE" envDefault:"10485760"` // 10MB
}

func LoadCustomConfig() *CustomConfig {
    cfg := &CustomConfig{}
    // Load from environment
    return cfg
}
```

Use in your code:

```go
customCfg := config.LoadCustomConfig()

if customCfg.FeatureFlag {
    // Feature enabled
}
```

---

## Configuration Validation

Validate required configuration on startup:

```go
func ValidateConfig() error {
    required := []string{
        "DB_HOST",
        "DB_DATABASE",
        "JWT_SECRET",
    }
    
    for _, key := range required {
        if config.Get(key) == "" {
            return fmt.Errorf("required config %s is missing", key)
        }
    }
    
    return nil
}
```

---

## Next Steps

- [Project Structure](./project-structure.md) - Understand file organization
- [First Application](./first-application.md) - Build your first app
- [Database Configuration](../database/configuration.md) - Database setup details
- [Security Best Practices](../security/best-practices.md) - Secure your app

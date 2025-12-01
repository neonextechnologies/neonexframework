# Logging System

Master application logging with NeonEx Framework's powerful Zap-based logger. Learn log levels, structured logging, log rotation, and production logging strategies.

## Table of Contents

- [Introduction](#introduction)
- [Quick Start](#quick-start)
- [Logger Configuration](#logger-configuration)
- [Log Levels](#log-levels)
- [Structured Logging](#structured-logging)
- [Log Formatting](#log-formatting)
- [Log Rotation](#log-rotation)
- [Production Logging](#production-logging)
- [Integration Examples](#integration-examples)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Introduction

NeonEx provides a high-performance logging system built on Zap. Key features include:

- **Multiple Log Levels**: Debug, Info, Warn, Error, Fatal
- **Structured Logging**: JSON and console formats
- **Performance**: Zero-allocation logging
- **Multiple Outputs**: Console, files, external services
- **Context Support**: Request tracing and correlation
- **Log Rotation**: Automatic log file rotation
- **Colored Output**: Enhanced readability in development

## Quick Start

### Basic Logging

```go
package main

import "neonex/core/pkg/logger"

func main() {
    // Use global logger
    logger.Info("Application started")
    logger.Debug("Debug information")
    logger.Warn("Warning message")
    logger.Error("Error occurred")
    
    // With fields
    logger.Info("User logged in", logger.Fields{
        "user_id": 123,
        "email":   "user@example.com",
        "ip":      "192.168.1.1",
    })
}
```

### Creating Custom Logger

```go
import "neonex/core/pkg/logger"

// Create logger instance
log := logger.NewLogger()

// Set log level
log.SetLevel(logger.InfoLevel)

// Set formatter
log.SetFormatter(logger.NewJSONFormatter())

// Add file output
file, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
log.AddWriter(file)

// Use logger
log.Info("Custom logger initialized")
```

## Logger Configuration

### Development Configuration

```go
func NewDevelopmentLogger() logger.Logger {
    log := logger.NewLogger()
    
    // Debug level for development
    log.SetLevel(logger.DebugLevel)
    
    // Console formatter with colors
    log.SetFormatter(logger.NewTextFormatter())
    log.EnableColor(true)
    log.EnableCaller(true)
    
    return log
}
```

### Production Configuration

```go
func NewProductionLogger() logger.Logger {
    log := logger.NewLogger()
    
    // Info level for production
    log.SetLevel(logger.InfoLevel)
    
    // JSON formatter for log aggregation
    log.SetFormatter(logger.NewJSONFormatter())
    log.EnableColor(false)
    log.EnableCaller(true)
    
    // Log to file
    file, err := os.OpenFile(
        "logs/app.log",
        os.O_CREATE|os.O_WRONLY|os.O_APPEND,
        0666,
    )
    if err == nil {
        log.AddWriter(file)
    }
    
    return log
}
```

### Environment-Based Configuration

```go
func NewLogger() logger.Logger {
    env := os.Getenv("APP_ENV")
    
    if env == "production" {
        return NewProductionLogger()
    }
    
    return NewDevelopmentLogger()
}
```

### Configuration File

```yaml
# config/logging.yaml
logging:
  level: info           # debug, info, warn, error, fatal
  format: json          # json, text
  output:
    - stdout
    - file: logs/app.log
  
  # Rotation settings
  rotation:
    max_size: 100       # MB
    max_age: 30         # days
    max_backups: 10
    compress: true
  
  # Development settings
  development:
    colored: true
    caller: true
    stacktrace: true
```

```go
type LogConfig struct {
    Level      string
    Format     string
    Output     []string
    Rotation   RotationConfig
    Colored    bool
    Caller     bool
    Stacktrace bool
}

func LoadLogConfig(path string) (*LogConfig, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var config LogConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

## Log Levels

### All Log Levels

```go
// Debug: Detailed information for diagnosing problems
logger.Debug("Cache miss", logger.Fields{
    "key":      "user:123",
    "duration": "5ms",
})

// Info: General informational messages
logger.Info("Request processed", logger.Fields{
    "method":   "GET",
    "path":     "/api/users",
    "status":   200,
    "duration": "150ms",
})

// Warn: Warning messages for potentially harmful situations
logger.Warn("Slow query detected", logger.Fields{
    "query":    "SELECT * FROM users",
    "duration": "5s",
})

// Error: Error messages for failure events
logger.Error("Database connection failed", logger.Fields{
    "error":  err.Error(),
    "host":   "localhost",
    "retry":  3,
})

// Fatal: Critical errors that cause application exit
logger.Fatal("Configuration file not found", logger.Fields{
    "path": "/etc/app/config.yaml",
})
```

### Dynamic Log Level

```go
// Set level at runtime
func SetLogLevel(level string) {
    switch strings.ToLower(level) {
    case "debug":
        logger.SetGlobalLevel(logger.DebugLevel)
    case "info":
        logger.SetGlobalLevel(logger.InfoLevel)
    case "warn":
        logger.SetGlobalLevel(logger.WarnLevel)
    case "error":
        logger.SetGlobalLevel(logger.ErrorLevel)
    default:
        logger.SetGlobalLevel(logger.InfoLevel)
    }
}

// API endpoint to change log level
func (h *AdminHandler) SetLogLevel(c echo.Context) error {
    level := c.QueryParam("level")
    SetLogLevel(level)
    
    return c.JSON(http.StatusOK, map[string]string{
        "message": fmt.Sprintf("Log level set to %s", level),
    })
}
```

### Conditional Logging

```go
func LogIfSlow(operation string, duration time.Duration, threshold time.Duration) {
    if duration > threshold {
        logger.Warn("Slow operation", logger.Fields{
            "operation": operation,
            "duration":  duration,
            "threshold": threshold,
        })
    }
}

// Usage
start := time.Now()
result := performOperation()
LogIfSlow("database_query", time.Since(start), 100*time.Millisecond)
```

## Structured Logging

### Field Types

```go
// All field types supported
logger.Info("User action", logger.Fields{
    // Primitives
    "user_id":    123,
    "email":      "user@example.com",
    "is_admin":   false,
    "score":      95.5,
    
    // Time
    "timestamp":  time.Now(),
    
    // Duration
    "duration":   150 * time.Millisecond,
    
    // Error
    "error":      err,
    
    // Complex types (will be JSON encoded)
    "metadata": map[string]interface{}{
        "source": "api",
        "version": "1.0",
    },
    
    // Arrays
    "tags": []string{"important", "urgent"},
})
```

### Request Logging

```go
func LogRequest(c echo.Context, duration time.Duration, statusCode int) {
    logger.Info("HTTP Request", logger.Fields{
        "method":       c.Request().Method,
        "path":         c.Request().URL.Path,
        "query":        c.Request().URL.RawQuery,
        "status":       statusCode,
        "duration_ms":  duration.Milliseconds(),
        "ip":           c.RealIP(),
        "user_agent":   c.Request().UserAgent(),
        "request_id":   c.Request().Header.Get("X-Request-ID"),
    })
}

// Middleware
func LoggingMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            start := time.Now()
            
            err := next(c)
            
            status := c.Response().Status
            if err != nil {
                if he, ok := err.(*echo.HTTPError); ok {
                    status = he.Code
                }
            }
            
            LogRequest(c, time.Since(start), status)
            
            return err
        }
    }
}
```

### Database Query Logging

```go
type QueryLogger struct {
    logger logger.Logger
}

func (ql *QueryLogger) LogQuery(query string, args []interface{}, duration time.Duration) {
    ql.logger.Debug("Database query", logger.Fields{
        "query":      query,
        "args":       args,
        "duration_ms": duration.Milliseconds(),
    })
    
    // Warn on slow queries
    if duration > 500*time.Millisecond {
        ql.logger.Warn("Slow query detected", logger.Fields{
            "query":    query,
            "duration": duration,
        })
    }
}

// GORM logger integration
type GormLogger struct {
    logger logger.Logger
}

func (gl *GormLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
    return gl
}

func (gl *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
    gl.logger.Info(msg, logger.Fields{"data": data})
}

func (gl *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
    gl.logger.Warn(msg, logger.Fields{"data": data})
}

func (gl *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
    gl.logger.Error(msg, logger.Fields{"data": data})
}

func (gl *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
    elapsed := time.Since(begin)
    sql, rows := fc()
    
    fields := logger.Fields{
        "sql":      sql,
        "rows":     rows,
        "duration": elapsed,
    }
    
    if err != nil {
        fields["error"] = err.Error()
        gl.logger.Error("Query error", fields)
    } else if elapsed > 200*time.Millisecond {
        gl.logger.Warn("Slow query", fields)
    } else {
        gl.logger.Debug("Query", fields)
    }
}
```

### Context Logger

```go
type contextKey string

const loggerKey contextKey = "logger"

// Add logger to context
func WithLogger(ctx context.Context, log logger.Logger) context.Context {
    return context.WithValue(ctx, loggerKey, log)
}

// Get logger from context
func FromContext(ctx context.Context) logger.Logger {
    if log, ok := ctx.Value(loggerKey).(logger.Logger); ok {
        return log
    }
    return logger.With(logger.Fields{})
}

// Usage in handlers
func (h *UserHandler) GetUser(c echo.Context) error {
    ctx := c.Request().Context()
    log := FromContext(ctx)
    
    userID := c.Param("id")
    log.Info("Fetching user", logger.Fields{"user_id": userID})
    
    // ... fetch user ...
    
    return c.JSON(http.StatusOK, user)
}
```

## Log Formatting

### JSON Formatter

```go
type JSONFormatter struct {
    TimestampFormat string
    PrettyPrint     bool
}

func (f *JSONFormatter) Format(entry *logger.Entry) ([]byte, error) {
    data := make(map[string]interface{})
    
    data["level"] = entry.Level.String()
    data["message"] = entry.Message
    data["time"] = entry.Time.Format(f.TimestampFormat)
    
    if entry.File != "" {
        data["file"] = fmt.Sprintf("%s:%d", entry.File, entry.Line)
    }
    
    // Add custom fields
    for k, v := range entry.Fields {
        data[k] = v
    }
    
    var output []byte
    var err error
    
    if f.PrettyPrint {
        output, err = json.MarshalIndent(data, "", "  ")
    } else {
        output, err = json.Marshal(data)
    }
    
    if err != nil {
        return nil, err
    }
    
    return append(output, '\n'), nil
}

// Output:
// {"level":"info","message":"User logged in","time":"2024-12-01T10:30:00Z","user_id":123,"email":"user@example.com"}
```

### Text Formatter (Console)

```go
type TextFormatter struct {
    TimestampFormat string
    Colored         bool
    FullTimestamp   bool
}

func (f *TextFormatter) Format(entry *logger.Entry) ([]byte, error) {
    var buf bytes.Buffer
    
    // Timestamp
    if f.FullTimestamp {
        buf.WriteString(entry.Time.Format(f.TimestampFormat))
        buf.WriteString(" ")
    }
    
    // Level (colored)
    if f.Colored {
        buf.WriteString(entry.Level.Color())
        buf.WriteString("[")
        buf.WriteString(entry.Level.String())
        buf.WriteString("]")
        buf.WriteString("\033[0m") // Reset color
    } else {
        buf.WriteString("[")
        buf.WriteString(entry.Level.String())
        buf.WriteString("]")
    }
    buf.WriteString(" ")
    
    // Caller info
    if entry.File != "" {
        buf.WriteString(fmt.Sprintf("%s:%d ", entry.File, entry.Line))
    }
    
    // Message
    buf.WriteString(entry.Message)
    
    // Fields
    if len(entry.Fields) > 0 {
        buf.WriteString(" ")
        for k, v := range entry.Fields {
            buf.WriteString(fmt.Sprintf("%s=%v ", k, v))
        }
    }
    
    buf.WriteString("\n")
    return buf.Bytes(), nil
}

// Output:
// 2024-12-01 10:30:00 [INFO] user.go:45 User logged in user_id=123 email=user@example.com
```

### Custom Formatter for External Services

```go
type SplunkFormatter struct{}

func (f *SplunkFormatter) Format(entry *logger.Entry) ([]byte, error) {
    event := map[string]interface{}{
        "time":       entry.Time.Unix(),
        "severity":   entry.Level.String(),
        "message":    entry.Message,
        "source":     entry.File,
        "attributes": entry.Fields,
    }
    
    return json.Marshal(map[string]interface{}{
        "event": event,
    })
}
```

## Log Rotation

### File Rotation with Lumberjack

```go
import "gopkg.in/natefinch/lumberjack.v2"

func SetupRotatingLogger() logger.Logger {
    log := logger.NewLogger()
    
    // Setup log rotation
    logFile := &lumberjack.Logger{
        Filename:   "logs/app.log",
        MaxSize:    100,  // MB
        MaxBackups: 10,   // Keep 10 old log files
        MaxAge:     30,   // Days
        Compress:   true, // Compress old log files
    }
    
    log.AddWriter(logFile)
    
    return log
}
```

### Daily Log Rotation

```go
type DailyRotator struct {
    basePath string
    current  *os.File
    date     string
}

func NewDailyRotator(basePath string) *DailyRotator {
    dr := &DailyRotator{basePath: basePath}
    dr.rotate()
    return dr
}

func (dr *DailyRotator) Write(p []byte) (n int, err error) {
    today := time.Now().Format("2006-01-02")
    
    if today != dr.date {
        dr.rotate()
    }
    
    return dr.current.Write(p)
}

func (dr *DailyRotator) rotate() {
    if dr.current != nil {
        dr.current.Close()
    }
    
    dr.date = time.Now().Format("2006-01-02")
    filename := fmt.Sprintf("%s_%s.log", dr.basePath, dr.date)
    
    dr.current, _ = os.OpenFile(
        filename,
        os.O_CREATE|os.O_WRONLY|os.O_APPEND,
        0666,
    )
}

// Usage
rotator := NewDailyRotator("logs/app")
log.AddWriter(rotator)
```

### Size-Based Rotation

```go
type SizeRotator struct {
    basePath    string
    maxSize     int64
    current     *os.File
    currentSize int64
    index       int
}

func (sr *SizeRotator) Write(p []byte) (n int, err error) {
    if sr.currentSize+int64(len(p)) > sr.maxSize {
        sr.rotate()
    }
    
    n, err = sr.current.Write(p)
    sr.currentSize += int64(n)
    return
}

func (sr *SizeRotator) rotate() {
    if sr.current != nil {
        sr.current.Close()
    }
    
    sr.index++
    filename := fmt.Sprintf("%s_%d.log", sr.basePath, sr.index)
    sr.current, _ = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
    sr.currentSize = 0
}
```

## Production Logging

### Centralized Logging

```go
import "github.com/elastic/go-elasticsearch/v8"

type ElasticsearchWriter struct {
    client *elasticsearch.Client
    index  string
}

func NewElasticsearchWriter(addresses []string, index string) (*ElasticsearchWriter, error) {
    cfg := elasticsearch.Config{
        Addresses: addresses,
    }
    
    client, err := elasticsearch.NewClient(cfg)
    if err != nil {
        return nil, err
    }
    
    return &ElasticsearchWriter{
        client: client,
        index:  index,
    }, nil
}

func (ew *ElasticsearchWriter) Write(p []byte) (n int, err error) {
    // Send log to Elasticsearch
    res, err := ew.client.Index(
        ew.index,
        bytes.NewReader(p),
        ew.client.Index.WithContext(context.Background()),
    )
    
    if err != nil {
        return 0, err
    }
    defer res.Body.Close()
    
    return len(p), nil
}

// Usage
esWriter, _ := NewElasticsearchWriter(
    []string{"http://localhost:9200"},
    "app-logs",
)
log.AddWriter(esWriter)
```

### Error Tracking (Sentry)

```go
import "github.com/getsentry/sentry-go"

type SentryWriter struct {
    hub *sentry.Hub
}

func NewSentryWriter(dsn string) (*SentryWriter, error) {
    err := sentry.Init(sentry.ClientOptions{
        Dsn: dsn,
    })
    
    if err != nil {
        return nil, err
    }
    
    return &SentryWriter{
        hub: sentry.CurrentHub(),
    }, nil
}

func (sw *SentryWriter) Write(p []byte) (n int, err error) {
    var entry map[string]interface{}
    if err := json.Unmarshal(p, &entry); err != nil {
        return 0, err
    }
    
    level, _ := entry["level"].(string)
    
    // Only send errors and above to Sentry
    if level == "error" || level == "fatal" {
        sw.hub.CaptureMessage(entry["message"].(string))
    }
    
    return len(p), nil
}

// Usage
sentryWriter, _ := NewSentryWriter("your-sentry-dsn")
log.AddWriter(sentryWriter)
```

### Performance Monitoring

```go
func LogPerformance(operation string, fn func() error) error {
    start := time.Now()
    
    err := fn()
    
    duration := time.Since(start)
    
    fields := logger.Fields{
        "operation": operation,
        "duration":  duration,
    }
    
    if err != nil {
        fields["error"] = err.Error()
        logger.Error("Operation failed", fields)
    } else {
        logger.Info("Operation completed", fields)
    }
    
    return err
}

// Usage
err := LogPerformance("fetch_users", func() error {
    return fetchUsers()
})
```

## Integration Examples

### Complete HTTP Logging

```go
func HTTPLoggingMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            req := c.Request()
            start := time.Now()
            
            // Generate request ID
            requestID := generateRequestID()
            c.Response().Header().Set("X-Request-ID", requestID)
            
            // Create request logger
            reqLogger := logger.With(logger.Fields{
                "request_id": requestID,
                "method":     req.Method,
                "path":       req.URL.Path,
                "ip":         c.RealIP(),
            })
            
            // Add logger to context
            ctx := WithLogger(req.Context(), reqLogger)
            c.SetRequest(req.WithContext(ctx))
            
            reqLogger.Info("Request started")
            
            // Process request
            err := next(c)
            
            // Log completion
            status := c.Response().Status
            if err != nil {
                if he, ok := err.(*echo.HTTPError); ok {
                    status = he.Code
                }
            }
            
            duration := time.Since(start)
            
            fields := logger.Fields{
                "status":   status,
                "duration": duration,
                "bytes":    c.Response().Size,
            }
            
            if err != nil {
                fields["error"] = err.Error()
                reqLogger.Error("Request failed", fields)
            } else {
                reqLogger.Info("Request completed", fields)
            }
            
            return err
        }
    }
}
```

### Service Layer Logging

```go
type UserService struct {
    db     *gorm.DB
    logger logger.Logger
}

func NewUserService(db *gorm.DB, log logger.Logger) *UserService {
    return &UserService{
        db:     db,
        logger: log.With(logger.Fields{"service": "user"}),
    }
}

func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    s.logger.Info("Creating user", logger.Fields{
        "email": req.Email,
    })
    
    user := &User{
        Email: req.Email,
        Name:  req.Name,
    }
    
    if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
        s.logger.Error("Failed to create user", logger.Fields{
            "email": req.Email,
            "error": err.Error(),
        })
        return nil, err
    }
    
    s.logger.Info("User created successfully", logger.Fields{
        "user_id": user.ID,
        "email":   user.Email,
    })
    
    return user, nil
}
```

## Best Practices

### 1. Consistent Field Names

```go
// Use consistent field names across the application
const (
    FieldUserID    = "user_id"
    FieldEmail     = "email"
    FieldRequestID = "request_id"
    FieldDuration  = "duration"
    FieldError     = "error"
)

logger.Info("User action", logger.Fields{
    FieldUserID:  123,
    FieldEmail:   "user@example.com",
    FieldDuration: duration,
})
```

### 2. Don't Log Sensitive Data

```go
// BAD
logger.Info("User login", logger.Fields{
    "email":    user.Email,
    "password": user.Password, // Never log passwords!
})

// GOOD
logger.Info("User login", logger.Fields{
    "user_id": user.ID,
    "email":   user.Email,
})
```

### 3. Use Appropriate Log Levels

```go
// Debug: Development troubleshooting
logger.Debug("Cache lookup", logger.Fields{"key": key})

// Info: Normal operations
logger.Info("User registered", logger.Fields{"user_id": id})

// Warn: Potential issues
logger.Warn("Rate limit approaching", logger.Fields{"requests": count})

// Error: Errors that need attention
logger.Error("Payment failed", logger.Fields{"error": err})

// Fatal: Critical failures (use sparingly)
logger.Fatal("Database connection failed")
```

### 4. Add Context

```go
// Add relevant context to logs
logger.Error("Order processing failed", logger.Fields{
    "order_id":      order.ID,
    "customer_id":   order.CustomerID,
    "total_amount":  order.Total,
    "payment_method": order.PaymentMethod,
    "error":         err.Error(),
    "retry_count":   retries,
})
```

### 5. Performance Considerations

```go
// Use debug level for verbose logs
if logger.IsDebugEnabled() {
    // Expensive operation only if debug is enabled
    details := generateDetailedDebugInfo()
    logger.Debug("Detailed debug info", logger.Fields{
        "details": details,
    })
}
```

## Troubleshooting

### Logs Not Appearing

```go
// Check log level
logger.SetGlobalLevel(logger.DebugLevel)

// Check output
logger.AddGlobalWriter(os.Stdout)

// Verify logger is initialized
if log == nil {
    log = logger.NewLogger()
}
```

### Performance Issues

```go
// Use async logging for high-throughput
type AsyncWriter struct {
    writer io.Writer
    queue  chan []byte
}

func (aw *AsyncWriter) Write(p []byte) (n int, err error) {
    // Copy data (as p may be reused)
    data := make([]byte, len(p))
    copy(data, p)
    
    select {
    case aw.queue <- data:
        return len(p), nil
    default:
        // Queue full, write synchronously
        return aw.writer.Write(p)
    }
}

func (aw *AsyncWriter) Start() {
    go func() {
        for data := range aw.queue {
            aw.writer.Write(data)
        }
    }()
}
```

### Log File Permissions

```bash
# Ensure log directory is writable
chmod 755 /var/log/app
chown app:app /var/log/app

# Or in code
os.MkdirAll("logs", 0755)
```

---

**Next Steps:**
- Learn about [Metrics](metrics.md) for performance monitoring
- Explore [Error Handling](../core-concepts/error-handling.md)
- See [Debugging](../testing/debugging.md) techniques

**Related Topics:**
- [Middleware](../core-concepts/middleware.md)
- [Monitoring](metrics.md)
- [Production Deployment](../deployment/production.md)

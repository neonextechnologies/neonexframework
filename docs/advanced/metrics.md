# Metrics & Monitoring

Master application monitoring with NeonEx Framework's Prometheus-compatible metrics system. Learn custom metrics, health checks, performance tracking, and observability best practices.

## Table of Contents

- [Introduction](#introduction)
- [Quick Start](#quick-start)
- [Metric Types](#metric-types)
- [Custom Metrics](#custom-metrics)
- [Health Checks](#health-checks)
- [System Metrics](#system-metrics)
- [HTTP Metrics](#http-metrics)
- [Database Metrics](#database-metrics)
- [Prometheus Integration](#prometheus-integration)
- [Dashboards](#dashboards)
- [Best Practices](#best-practices)

## Introduction

NeonEx provides comprehensive metrics collection with Prometheus compatibility:

- **Counter**: Monotonically increasing values
- **Gauge**: Values that can go up and down
- **Histogram**: Distribution of values
- **Summary**: Quantiles over time
- **System Metrics**: CPU, memory, goroutines
- **Health Checks**: Application health endpoints

## Quick Start

```go
package main

import (
    "neonex/core/pkg/metrics"
    "github.com/labstack/echo/v4"
)

func main() {
    // Create metrics collector
    collector := metrics.NewCollector(metrics.DefaultCollectorConfig())
    
    // Create counter
    requestCounter := collector.NewCounter(
        "http_requests_total",
        "Total HTTP requests",
        map[string]string{"app": "myapp"},
    )
    
    // Increment counter
    requestCounter.Inc()
    
    // Create gauge
    activeUsers := collector.NewGauge(
        "active_users",
        "Number of active users",
        nil,
    )
    
    activeUsers.Set(150)
    
    // Expose metrics endpoint
    e := echo.New()
    e.GET("/metrics", func(c echo.Context) error {
        allMetrics := collector.GetAllMetrics()
        return c.JSON(200, allMetrics)
    })
    
    e.Start(":8080")
}
```

## Metric Types

### Counter

```go
// Create counter
requestCounter := collector.NewCounter(
    "api_requests_total",
    "Total API requests",
    map[string]string{
        "service": "api",
        "version": "v1",
    },
)

// Increment by 1
requestCounter.Inc()

// Add specific value
requestCounter.Add(5)

// Get current value
count := requestCounter.Get()
fmt.Printf("Total requests: %d\n", count)
```

### Gauge

```go
// Create gauge
memoryGauge := collector.NewGauge(
    "memory_usage_bytes",
    "Current memory usage",
    nil,
)

// Set value
memoryGauge.Set(1024 * 1024 * 100) // 100MB

// Increment/Decrement
memoryGauge.Inc() // +1
memoryGauge.Dec() // -1
memoryGauge.Add(100)
memoryGauge.Sub(50)

// Get value
usage := memoryGauge.Get()
```

### Histogram

```go
// Create histogram with buckets
responseTimeHistogram := collector.NewHistogram(
    "http_request_duration_seconds",
    "HTTP request duration",
    map[string]string{"endpoint": "/api/users"},
    []float64{0.001, 0.01, 0.1, 0.5, 1.0, 5.0}, // Buckets in seconds
)

// Observe values
start := time.Now()
// ... handle request ...
duration := time.Since(start).Seconds()
responseTimeHistogram.Observe(duration)

// Get statistics
sum := responseTimeHistogram.GetSum()
count := responseTimeHistogram.GetCount()
avg := sum / float64(count)
```

### Summary

```go
// Create summary
requestSizeSummary := collector.NewSummary(
    "http_request_size_bytes",
    "HTTP request size distribution",
    nil,
)

// Observe values
requestSizeSummary.Observe(1024)
requestSizeSummary.Observe(2048)
requestSizeSummary.Observe(512)

// Get statistics
sum := requestSizeSummary.GetSum()
count := requestSizeSummary.GetCount()
avg := requestSizeSummary.GetAverage()
```

## Custom Metrics

### Application Metrics

```go
type AppMetrics struct {
    requestsTotal      *metrics.Counter
    requestDuration    *metrics.Histogram
    activeConnections  *metrics.Gauge
    errorCount         *metrics.Counter
    cacheHits          *metrics.Counter
    cacheMisses        *metrics.Counter
}

func NewAppMetrics(collector *metrics.Collector) *AppMetrics {
    return &AppMetrics{
        requestsTotal: collector.NewCounter(
            "app_requests_total",
            "Total requests",
            map[string]string{"app": "myapp"},
        ),
        requestDuration: collector.NewHistogram(
            "app_request_duration_seconds",
            "Request duration",
            nil,
            []float64{0.001, 0.01, 0.1, 0.5, 1.0},
        ),
        activeConnections: collector.NewGauge(
            "app_active_connections",
            "Active connections",
            nil,
        ),
        errorCount: collector.NewCounter(
            "app_errors_total",
            "Total errors",
            map[string]string{"severity": "error"},
        ),
        cacheHits: collector.NewCounter(
            "app_cache_hits_total",
            "Cache hits",
            nil,
        ),
        cacheMisses: collector.NewCounter(
            "app_cache_misses_total",
            "Cache misses",
            nil,
        ),
    }
}

func (m *AppMetrics) RecordRequest(duration time.Duration) {
    m.requestsTotal.Inc()
    m.requestDuration.Observe(duration.Seconds())
}

func (m *AppMetrics) RecordError() {
    m.errorCount.Inc()
}

func (m *AppMetrics) IncrementConnections() {
    m.activeConnections.Inc()
}

func (m *AppMetrics) DecrementConnections() {
    m.activeConnections.Dec()
}

func (m *AppMetrics) RecordCacheHit() {
    m.cacheHits.Inc()
}

func (m *AppMetrics) RecordCacheMiss() {
    m.cacheMisses.Inc()
}

func (m *AppMetrics) CacheHitRate() float64 {
    hits := float64(m.cacheHits.Get())
    misses := float64(m.cacheMisses.Get())
    total := hits + misses
    
    if total == 0 {
        return 0
    }
    
    return hits / total * 100
}
```

### Business Metrics

```go
type BusinessMetrics struct {
    ordersTotal       *metrics.Counter
    orderValue        *metrics.Histogram
    registrations     *metrics.Counter
    activeUsers       *metrics.Gauge
    revenue           *metrics.Counter
}

func NewBusinessMetrics(collector *metrics.Collector) *BusinessMetrics {
    return &BusinessMetrics{
        ordersTotal: collector.NewCounter(
            "orders_total",
            "Total orders",
            nil,
        ),
        orderValue: collector.NewHistogram(
            "order_value_usd",
            "Order value distribution",
            nil,
            []float64{10, 50, 100, 500, 1000, 5000},
        ),
        registrations: collector.NewCounter(
            "user_registrations_total",
            "Total user registrations",
            nil,
        ),
        activeUsers: collector.NewGauge(
            "active_users_count",
            "Currently active users",
            nil,
        ),
        revenue: collector.NewCounter(
            "revenue_total_usd",
            "Total revenue in USD",
            nil,
        ),
    }
}

func (bm *BusinessMetrics) RecordOrder(amount float64) {
    bm.ordersTotal.Inc()
    bm.orderValue.Observe(amount)
    bm.revenue.Add(uint64(amount * 100)) // Store as cents
}

func (bm *BusinessMetrics) RecordRegistration() {
    bm.registrations.Inc()
}

func (bm *BusinessMetrics) UpdateActiveUsers(count int) {
    bm.activeUsers.Set(int64(count))
}
```

## Health Checks

### Basic Health Check

```go
type HealthChecker struct {
    db    *gorm.DB
    redis *redis.Client
    cache cache.Cache
}

func (hc *HealthChecker) Check(c echo.Context) error {
    ctx := c.Request().Context()
    
    health := map[string]interface{}{
        "status": "healthy",
        "checks": map[string]interface{}{},
    }
    
    // Database check
    if err := hc.checkDatabase(ctx); err != nil {
        health["status"] = "unhealthy"
        health["checks"].(map[string]interface{})["database"] = map[string]interface{}{
            "status": "down",
            "error":  err.Error(),
        }
    } else {
        health["checks"].(map[string]interface{})["database"] = map[string]interface{}{
            "status": "up",
        }
    }
    
    // Redis check
    if err := hc.checkRedis(ctx); err != nil {
        health["status"] = "unhealthy"
        health["checks"].(map[string]interface{})["redis"] = map[string]interface{}{
            "status": "down",
            "error":  err.Error(),
        }
    } else {
        health["checks"].(map[string]interface{})["redis"] = map[string]interface{}{
            "status": "up",
        }
    }
    
    // Determine HTTP status
    status := http.StatusOK
    if health["status"] == "unhealthy" {
        status = http.StatusServiceUnavailable
    }
    
    return c.JSON(status, health)
}

func (hc *HealthChecker) checkDatabase(ctx context.Context) error {
    var result int
    return hc.db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error
}

func (hc *HealthChecker) checkRedis(ctx context.Context) error {
    return hc.redis.Ping(ctx).Err()
}
```

### Detailed Health Check

```go
type DetailedHealth struct {
    Status      string                 `json:"status"`
    Version     string                 `json:"version"`
    Uptime      string                 `json:"uptime"`
    Timestamp   time.Time              `json:"timestamp"`
    Checks      map[string]CheckResult `json:"checks"`
    SystemInfo  SystemInfo             `json:"system"`
}

type CheckResult struct {
    Status   string        `json:"status"`
    Duration time.Duration `json:"duration"`
    Error    string        `json:"error,omitempty"`
}

type SystemInfo struct {
    Goroutines   int     `json:"goroutines"`
    MemoryMB     uint64  `json:"memory_mb"`
    CPUPercent   float64 `json:"cpu_percent"`
}

func (hc *HealthChecker) DetailedCheck(c echo.Context) error {
    startTime := time.Now()
    
    health := &DetailedHealth{
        Status:    "healthy",
        Version:   "1.0.0",
        Uptime:    time.Since(hc.startTime).String(),
        Timestamp: time.Now(),
        Checks:    make(map[string]CheckResult),
    }
    
    // Run checks
    health.Checks["database"] = hc.runCheck("database", hc.checkDatabase)
    health.Checks["redis"] = hc.runCheck("redis", hc.checkRedis)
    health.Checks["cache"] = hc.runCheck("cache", hc.checkCache)
    
    // Get system info
    health.SystemInfo = hc.getSystemInfo()
    
    // Determine overall status
    for _, check := range health.Checks {
        if check.Status != "up" {
            health.Status = "unhealthy"
            break
        }
    }
    
    status := http.StatusOK
    if health.Status == "unhealthy" {
        status = http.StatusServiceUnavailable
    }
    
    return c.JSON(status, health)
}

func (hc *HealthChecker) runCheck(name string, check func(context.Context) error) CheckResult {
    start := time.Now()
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    err := check(ctx)
    duration := time.Since(start)
    
    result := CheckResult{
        Duration: duration,
    }
    
    if err != nil {
        result.Status = "down"
        result.Error = err.Error()
    } else {
        result.Status = "up"
    }
    
    return result
}

func (hc *HealthChecker) getSystemInfo() SystemInfo {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    return SystemInfo{
        Goroutines: runtime.NumGoroutine(),
        MemoryMB:   m.Alloc / 1024 / 1024,
        CPUPercent: 0, // Calculate actual CPU if needed
    }
}
```

## System Metrics

The collector automatically tracks system metrics:

```go
// System metrics are collected automatically
collector := metrics.NewCollector(metrics.CollectorConfig{
    CollectSystemMetrics:  true,
    SystemMetricsInterval: 5 * time.Second,
})

// Access system metrics
allMetrics := collector.GetAllMetrics()

for _, metric := range allMetrics {
    switch metric.Name {
    case "system_cpu_percent":
        fmt.Printf("CPU: %.2f%%\n", metric.Value)
    case "system_memory_bytes":
        fmt.Printf("Memory: %d MB\n", uint64(metric.Value)/1024/1024)
    case "system_goroutines":
        fmt.Printf("Goroutines: %d\n", int(metric.Value))
    }
}
```

## HTTP Metrics

### Middleware for HTTP Metrics

```go
func MetricsMiddleware(metrics *AppMetrics) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            start := time.Now()
            
            // Increment active connections
            metrics.IncrementConnections()
            defer metrics.DecrementConnections()
            
            // Handle request
            err := next(c)
            
            // Record metrics
            duration := time.Since(start)
            metrics.RecordRequest(duration)
            
            // Record errors
            if err != nil {
                metrics.RecordError()
            }
            
            // Log slow requests
            if duration > 1*time.Second {
                log.Warn("Slow request", logger.Fields{
                    "path":     c.Request().URL.Path,
                    "duration": duration,
                })
            }
            
            return err
        }
    }
}
```

### Per-Endpoint Metrics

```go
type EndpointMetrics struct {
    collector *metrics.Collector
    counters  map[string]*metrics.Counter
    durations map[string]*metrics.Histogram
}

func NewEndpointMetrics(collector *metrics.Collector) *EndpointMetrics {
    return &EndpointMetrics{
        collector: collector,
        counters:  make(map[string]*metrics.Counter),
        durations: make(map[string]*metrics.Histogram),
    }
}

func (em *EndpointMetrics) getOrCreateCounter(endpoint string) *metrics.Counter {
    if counter, exists := em.counters[endpoint]; exists {
        return counter
    }
    
    counter := em.collector.NewCounter(
        "http_endpoint_requests_total",
        "Total requests per endpoint",
        map[string]string{"endpoint": endpoint},
    )
    
    em.counters[endpoint] = counter
    return counter
}

func (em *EndpointMetrics) RecordRequest(endpoint string, duration time.Duration) {
    counter := em.getOrCreateCounter(endpoint)
    counter.Inc()
    
    // Also record duration
    if histogram, exists := em.durations[endpoint]; exists {
        histogram.Observe(duration.Seconds())
    }
}
```

## Database Metrics

```go
type DBMetrics struct {
    queryCount    *metrics.Counter
    queryDuration *metrics.Histogram
    slowQueries   *metrics.Counter
    errors        *metrics.Counter
}

func NewDBMetrics(collector *metrics.Collector) *DBMetrics {
    return &DBMetrics{
        queryCount: collector.NewCounter(
            "db_queries_total",
            "Total database queries",
            nil,
        ),
        queryDuration: collector.NewHistogram(
            "db_query_duration_seconds",
            "Database query duration",
            nil,
            []float64{0.001, 0.01, 0.1, 0.5, 1.0, 5.0},
        ),
        slowQueries: collector.NewCounter(
            "db_slow_queries_total",
            "Total slow queries",
            nil,
        ),
        errors: collector.NewCounter(
            "db_errors_total",
            "Total database errors",
            nil,
        ),
    }
}

func (dbm *DBMetrics) RecordQuery(duration time.Duration, err error) {
    dbm.queryCount.Inc()
    dbm.queryDuration.Observe(duration.Seconds())
    
    if duration > 500*time.Millisecond {
        dbm.slowQueries.Inc()
    }
    
    if err != nil {
        dbm.errors.Inc()
    }
}

// GORM Callback
func (dbm *DBMetrics) GORMCallback() func(*gorm.DB) {
    return func(db *gorm.DB) {
        start := time.Now()
        
        db.Statement.Context = context.WithValue(
            db.Statement.Context,
            "query_start",
            start,
        )
    }
}

func (dbm *DBMetrics) GORMAfterCallback() func(*gorm.DB) {
    return func(db *gorm.DB) {
        start, ok := db.Statement.Context.Value("query_start").(time.Time)
        if !ok {
            return
        }
        
        duration := time.Since(start)
        dbm.RecordQuery(duration, db.Error)
    }
}

// Register with GORM
db.Callback().Query().Before("gorm:query").Register("metrics:before", dbm.GORMCallback())
db.Callback().Query().After("gorm:query").Register("metrics:after", dbm.GORMAfterCallback())
```

## Prometheus Integration

### Prometheus Exporter

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusExporter struct {
    collector *metrics.Collector
    registry  *prometheus.Registry
}

func NewPrometheusExporter(collector *metrics.Collector) *PrometheusExporter {
    return &PrometheusExporter{
        collector: collector,
        registry:  prometheus.NewRegistry(),
    }
}

func (pe *PrometheusExporter) Handler() http.Handler {
    return promhttp.HandlerFor(pe.registry, promhttp.HandlerOpts{})
}

// Expose metrics endpoint
e.GET("/metrics", echo.WrapHandler(prometheusExporter.Handler()))
```

### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'neonex-app'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

## Dashboards

### Grafana Dashboard JSON

```json
{
  "dashboard": {
    "title": "NeonEx Application Metrics",
    "panels": [
      {
        "title": "Request Rate",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])"
          }
        ]
      },
      {
        "title": "Response Time",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))"
          }
        ]
      },
      {
        "title": "Error Rate",
        "targets": [
          {
            "expr": "rate(app_errors_total[5m])"
          }
        ]
      },
      {
        "title": "Active Connections",
        "targets": [
          {
            "expr": "app_active_connections"
          }
        ]
      }
    ]
  }
}
```

## Best Practices

### 1. Metric Naming

```go
// Use descriptive names with units
// ✓ Good
"http_request_duration_seconds"
"memory_usage_bytes"
"cache_hit_ratio"

// ✗ Bad
"http_time"
"mem"
"hits"
```

### 2. Label Usage

```go
// Use labels for dimensions, not values
// ✓ Good
collector.NewCounter("http_requests_total", "...", map[string]string{
    "method": "GET",
    "status": "200",
})

// ✗ Bad - don't use high-cardinality labels
collector.NewCounter("http_requests_total", "...", map[string]string{
    "user_id": "123",    // Too many unique values
    "timestamp": "...",  // Changes constantly
})
```

### 3. Regular Monitoring

```go
// Set up alerts for critical metrics
func SetupAlerts(metrics *AppMetrics) {
    go func() {
        ticker := time.NewTicker(1 * time.Minute)
        defer ticker.Stop()
        
        for range ticker.C {
            // Check error rate
            errorRate := float64(metrics.errorCount.Get()) / float64(metrics.requestsTotal.Get())
            if errorRate > 0.05 { // 5% error rate
                alerting.Send("High error rate: %.2f%%", errorRate*100)
            }
            
            // Check response time
            avgDuration := metrics.requestDuration.GetSum() / float64(metrics.requestDuration.GetCount())
            if avgDuration > 1.0 { // 1 second
                alerting.Send("High response time: %.2fs", avgDuration)
            }
        }
    }()
}
```

### 4. Resource Cleanup

```go
// Clean up metrics when done
defer collector.Close()

// Reset metrics if needed
collector.Reset()
```

### 5. Testing Metrics

```go
func TestMetrics(t *testing.T) {
    collector := metrics.NewCollector(metrics.DefaultCollectorConfig())
    appMetrics := NewAppMetrics(collector)
    
    // Record some metrics
    appMetrics.RecordRequest(100 * time.Millisecond)
    appMetrics.RecordRequest(200 * time.Millisecond)
    
    // Verify
    count := appMetrics.requestsTotal.Get()
    assert.Equal(t, uint64(2), count)
    
    avgDuration := appMetrics.requestDuration.GetSum() / float64(appMetrics.requestDuration.GetCount())
    assert.InDelta(t, 0.15, avgDuration, 0.01) // ~150ms average
}
```

---

**Next Steps:**
- Learn about [Logging](logging.md) for debugging
- Explore [Health Checks](../deployment/health-checks.md)
- See [Grafana Setup](../deployment/grafana.md)

**Related Topics:**
- [Prometheus](https://prometheus.io/)
- [Grafana Dashboards](https://grafana.com/)
- [Observability](../deployment/observability.md)

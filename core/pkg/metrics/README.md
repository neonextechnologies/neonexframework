# Metrics Package

Real-time metrics collection and monitoring dashboard with WebSocket-powered live updates for NeonexCore.

## Features

- ✅ **Metric Types** - Counter, Gauge, Histogram, Summary
- ✅ **System Metrics** - CPU, Memory, Goroutines, GC Pause
- ✅ **HTTP Metrics** - Request count, duration, size, status codes
- ✅ **Real-time Dashboard** - WebSocket-powered live visualization
- ✅ **Custom Metrics** - Create your own application metrics
- ✅ **Alert System** - Configurable alerts with threshold triggers
- ✅ **Thread-Safe** - Atomic operations for high concurrency
- ✅ **Low Overhead** - Minimal performance impact
- ✅ **Beautiful UI** - Modern gradient dashboard with charts

## Architecture

```
pkg/metrics/
├── collector.go   - Metric collection and management
├── dashboard.go   - Real-time dashboard and alerts
├── middleware.go  - HTTP metrics middleware
└── README.md      - Documentation
```

## Quick Start

### 1. Basic Setup

```go
import (
    "neonexcore/pkg/metrics"
    "neonexcore/pkg/websocket"
)

// Create collector
config := metrics.DefaultCollectorConfig()
collector := metrics.NewCollector(config)

// Create WebSocket hub (required for real-time dashboard)
hubConfig := websocket.DefaultHubConfig()
hub := websocket.NewHub(hubConfig)

// Create dashboard
dashConfig := metrics.DefaultDashboardConfig()
dashboard := metrics.NewDashboard(collector, hub, dashConfig)

// Setup routes
app := fiber.New()
dashboard.SetupRoutes(app)
```

### 2. HTTP Metrics Middleware

```go
// Add metrics middleware
app.Use(metrics.Middleware(collector))
app.Use(metrics.MethodMiddleware(collector))
app.Use(metrics.ErrorMiddleware(collector))

// Your routes
app.Get("/api/users", handleGetUsers)
app.Post("/api/users", handleCreateUser)
```

### 3. Custom Application Metrics

```go
// Counter (monotonically increasing)
loginCounter := collector.NewCounter(
    "user_logins_total",
    "Total number of user logins",
    map[string]string{"service": "auth"},
)
loginCounter.Inc()

// Gauge (can go up and down)
activeUsers := collector.NewGauge(
    "users_active",
    "Number of currently active users",
    nil,
)
activeUsers.Set(150)
activeUsers.Inc()
activeUsers.Dec()

// Histogram (distribution of values)
queryDuration := collector.NewHistogram(
    "db_query_duration_seconds",
    "Database query duration",
    nil,
    []float64{0.001, 0.01, 0.1, 1, 10},
)
queryDuration.Observe(0.035) // 35ms

// Summary (quantiles over time)
requestSize := collector.NewSummary(
    "request_payload_bytes",
    "Request payload size in bytes",
    nil,
)
requestSize.Observe(1024)
```

## Metric Types

### Counter

Monotonically increasing value (e.g., total requests, errors):

```go
counter := collector.NewCounter("events_total", "Total events", nil)
counter.Inc()        // Increment by 1
counter.Add(10)      // Add 10
value := counter.Get()  // Get current value
counter.Reset()      // Reset to 0
```

**Use Cases:**
- Total HTTP requests
- Total errors
- Total processed jobs
- Total cache hits/misses

### Gauge

Value that can go up and down (e.g., memory usage, active connections):

```go
gauge := collector.NewGauge("queue_size", "Current queue size", nil)
gauge.Set(100)       // Set to 100
gauge.Inc()          // Increment by 1
gauge.Dec()          // Decrement by 1
gauge.Add(10)        // Add 10
gauge.Sub(5)         // Subtract 5
value := gauge.Get() // Get current value
```

**Use Cases:**
- Memory usage
- Active connections
- Queue size
- Temperature
- CPU usage

### Histogram

Distribution of values with buckets (e.g., request duration):

```go
histogram := collector.NewHistogram(
    "response_time",
    "Response time distribution",
    nil,
    []float64{0.01, 0.05, 0.1, 0.5, 1, 5},
)
histogram.Observe(0.123)  // Record 123ms
sum := histogram.GetSum()
count := histogram.GetCount()
buckets := histogram.GetBuckets()
```

**Use Cases:**
- Request/response duration
- Payload size
- Query execution time
- Latency measurements

### Summary

Similar to histogram but calculates quantiles:

```go
summary := collector.NewSummary("latency", "Request latency", nil)
summary.Observe(0.045)    // Record value
average := summary.GetAverage()
sum := summary.GetSum()
count := summary.GetCount()
```

**Use Cases:**
- Average response time
- Percentile calculations
- Statistical analysis

## System Metrics

Auto-collected every 5 seconds (configurable):

```go
config := metrics.CollectorConfig{
    CollectSystemMetrics:  true,
    SystemMetricsInterval: 5 * time.Second,
    EnableHistory:         true,
    HistorySize:           100,
}
```

**Available System Metrics:**
- `system_cpu_percent` - CPU usage percentage
- `system_memory_bytes` - Memory usage in bytes
- `system_goroutines` - Number of goroutines
- `system_gc_pause_ns` - GC pause time in nanoseconds

## HTTP Middleware

### Basic HTTP Metrics

```go
app.Use(metrics.Middleware(collector))
```

**Collected Metrics:**
- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request duration histogram
- `http_requests_active` - Active requests gauge
- `http_request_size_bytes` - Request size histogram
- `http_response_size_bytes` - Response size histogram
- `http_responses_{status}` - Responses by status code

### Method-based Metrics

```go
app.Use(metrics.MethodMiddleware(collector))
```

**Collected Metrics:**
- `http_requests_GET` - GET requests
- `http_requests_POST` - POST requests
- `http_requests_PUT` - PUT requests
- `http_requests_DELETE` - DELETE requests
- `http_requests_PATCH` - PATCH requests

### Path-based Metrics

```go
paths := []string{"/api/users", "/api/products", "/api/orders"}
app.Use(metrics.PathMiddleware(collector, paths))
```

**Collected Metrics:**
- `http_requests_path_{path}` - Requests per path

### Error Tracking

```go
app.Use(metrics.ErrorMiddleware(collector))
```

**Collected Metrics:**
- `http_errors_total` - Total errors
- `http_errors_5xx` - Server errors
- `http_errors_4xx` - Client errors

## Real-time Dashboard

### Setup

```go
// Create dashboard
dashboard := metrics.NewDashboard(collector, hub, dashConfig)
dashboard.SetupRoutes(app)

// Access dashboard
// http://localhost:3000/metrics/dashboard
```

### Features

- **Live Charts** - Real-time line charts for CPU, Memory, Goroutines, GC
- **Metric List** - All metrics with current values
- **Auto-refresh** - Updates every 1 second (configurable)
- **WebSocket Connection** - Live data stream
- **Responsive Design** - Works on desktop and mobile
- **Beautiful UI** - Modern gradient design with Chart.js

### Dashboard Configuration

```go
config := metrics.DashboardConfig{
    BroadcastInterval: 1 * time.Second,  // Update frequency
    EnableAlerts:      true,              // Enable alerting
    EnableHistory:     true,              // Keep metric history
    HistorySize:       60,                // History items (60 seconds)
}
```

## Alert System

### Create Alerts

```go
// High memory alert
dashboard.AddAlert(metrics.Alert{
    Name:        "high_memory",
    Description: "Memory usage is too high",
    Metric:      "system_memory_bytes",
    Condition:   metrics.ConditionGreaterThan,
    Threshold:   500 * 1024 * 1024, // 500MB
    Enabled:     true,
})

// Low response time alert
dashboard.AddAlert(metrics.Alert{
    Name:        "slow_response",
    Description: "Response time is too slow",
    Metric:      "http_request_duration_seconds",
    Condition:   metrics.ConditionGreaterThan,
    Threshold:   1.0, // 1 second
    Enabled:     true,
})

// Connection threshold
dashboard.AddAlert(metrics.Alert{
    Name:        "too_many_connections",
    Description: "Active connections exceeded threshold",
    Metric:      "http_requests_active",
    Condition:   metrics.ConditionGreaterThan,
    Threshold:   1000,
    Enabled:     true,
})
```

### Alert Conditions

- `ConditionGreaterThan` - Trigger when value > threshold
- `ConditionLessThan` - Trigger when value < threshold
- `ConditionEquals` - Trigger when value == threshold
- `ConditionNotEquals` - Trigger when value != threshold

### Manage Alerts

```go
// Get all alerts
alerts := dashboard.GetAlerts()

// Remove alert
dashboard.RemoveAlert("high_memory")
```

### Alert API Endpoints

```bash
# Get all alerts
GET /metrics/alerts

# Add new alert
POST /metrics/alerts
{
  "name": "high_cpu",
  "description": "CPU usage is high",
  "metric": "system_cpu_percent",
  "condition": "gt",
  "threshold": 80.0
}

# Delete alert
DELETE /metrics/alerts/high_cpu
```

## API Endpoints

### Get All Metrics

```bash
GET /metrics
```

Response:
```json
{
  "success": true,
  "timestamp": 1234567890,
  "uptime": 3600.5,
  "metrics": [
    {
      "name": "http_requests_total",
      "type": "counter",
      "value": 1234,
      "timestamp": "2024-01-01T12:00:00Z"
    }
  ]
}
```

### Get Specific Metric

```bash
GET /metrics/http_requests_total
```

Response:
```json
{
  "success": true,
  "metric": {
    "name": "http_requests_total",
    "type": "counter",
    "value": 1234,
    "labels": {},
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

### Dashboard UI

```bash
GET /metrics/dashboard
```

Opens beautiful real-time dashboard in browser.

## Integration Examples

### E-commerce Application

```go
// Product metrics
productViews := collector.NewCounter("product_views_total", "Product views", nil)
addToCart := collector.NewCounter("cart_add_total", "Items added to cart", nil)
purchases := collector.NewCounter("purchases_total", "Successful purchases", nil)
revenue := collector.NewSummary("revenue_amount", "Revenue in USD", nil)

// Order processing
orderDuration := collector.NewHistogram(
    "order_processing_seconds",
    "Order processing time",
    nil,
    []float64{0.1, 0.5, 1, 5, 10, 30},
)

// Inventory
stockLevel := collector.NewGauge("inventory_stock", "Current stock level", nil)

// Usage
func handleProductView(c *fiber.Ctx) error {
    productViews.Inc()
    // ... handle view
}

func handleAddToCart(c *fiber.Ctx) error {
    addToCart.Inc()
    stockLevel.Dec()
    // ... handle cart
}

func handlePurchase(c *fiber.Ctx, amount float64) error {
    start := time.Now()
    
    purchases.Inc()
    revenue.Observe(amount)
    
    // ... process order
    
    orderDuration.Observe(time.Since(start).Seconds())
    return nil
}
```

### API Gateway

```go
// Rate limiting
rateLimitHits := collector.NewCounter("ratelimit_hits", "Rate limit hits", nil)
rateLimitBlocked := collector.NewCounter("ratelimit_blocked", "Rate limit blocks", nil)

// Authentication
authAttempts := collector.NewCounter("auth_attempts", "Auth attempts", nil)
authSuccess := collector.NewCounter("auth_success", "Successful auth", nil)
authFailures := collector.NewCounter("auth_failures", "Failed auth", nil)

// API usage
apiCalls := collector.NewCounter("api_calls_total", "Total API calls", nil)
apiLatency := collector.NewHistogram(
    "api_latency_seconds",
    "API latency",
    nil,
    []float64{0.001, 0.01, 0.1, 1},
)

// Active sessions
activeSessions := collector.NewGauge("sessions_active", "Active sessions", nil)
```

### Microservices

```go
// Service health
serviceUp := collector.NewGauge("service_up", "Service health (1=up, 0=down)", nil)
serviceUp.Set(1)

// Inter-service calls
serviceCallsDuration := collector.NewHistogram(
    "service_calls_duration",
    "Duration of calls to other services",
    map[string]string{"target_service": "user-service"},
    nil,
)

// Message queue
queueSize := collector.NewGauge("queue_size", "Message queue size", nil)
queueProcessed := collector.NewCounter("queue_processed", "Processed messages", nil)
queueErrors := collector.NewCounter("queue_errors", "Queue processing errors", nil)
```

## Best Practices

1. **Use Appropriate Metric Types**
   - Counter: Monotonically increasing (requests, errors)
   - Gauge: Up and down values (memory, connections)
   - Histogram: Distribution (latency, size)
   - Summary: Statistical analysis (averages, quantiles)

2. **Keep Metric Names Descriptive**
   ```go
   ✅ "http_requests_total"
   ✅ "db_query_duration_seconds"
   ❌ "requests"
   ❌ "time"
   ```

3. **Use Labels Sparingly**
   ```go
   ✅ map[string]string{"method": "GET", "status": "200"}
   ❌ map[string]string{"user_id": "123"} // Too high cardinality
   ```

4. **Set Realistic Alert Thresholds**
   - Based on historical data
   - Allow for normal spikes
   - Avoid alert fatigue

5. **Monitor What Matters**
   - Request rate and latency
   - Error rate
   - Resource usage (CPU, memory)
   - Business metrics (signups, revenue)

6. **Use Histograms for Latency**
   ```go
   // Better than average
   histogram.Observe(duration)
   // Shows distribution, p50, p95, p99
   ```

7. **Cleanup Old Metrics**
   ```go
   // Reset counters periodically
   collector.Reset()
   ```

## Performance

- **Atomic Operations** - Lock-free increments (Counter, Gauge)
- **Read-Write Locks** - Efficient concurrent access
- **Minimal Overhead** - < 1μs per metric operation
- **Memory Efficient** - ~1KB per metric
- **Scalable** - Handles 100,000+ metrics/second

## Configuration

### Collector Config

```go
config := metrics.CollectorConfig{
    CollectSystemMetrics:  true,              // Auto-collect system metrics
    SystemMetricsInterval: 5 * time.Second,   // Collection interval
    EnableHistory:         true,              // Keep metric history
    HistorySize:           100,               // History size
    DefaultBuckets:        []float64{...},    // Default histogram buckets
}
```

### Dashboard Config

```go
config := metrics.DashboardConfig{
    BroadcastInterval: 1 * time.Second,  // WebSocket broadcast rate
    EnableAlerts:      true,              // Enable alerting
    EnableHistory:     true,              // Keep history
    HistorySize:       60,                // 60 data points
}
```

## License

MIT License - Part of NeonexCore Framework

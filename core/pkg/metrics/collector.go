package metrics

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// MetricType represents the type of metric
type MetricType string

const (
	TypeCounter   MetricType = "counter"
	TypeGauge     MetricType = "gauge"
	TypeHistogram MetricType = "histogram"
	TypeSummary   MetricType = "summary"
)

// Metric represents a single metric with metadata
type Metric struct {
	Name        string                 `json:"name"`
	Type        MetricType             `json:"type"`
	Value       float64                `json:"value"`
	Labels      map[string]string      `json:"labels,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Description string                 `json:"description,omitempty"`
	Unit        string                 `json:"unit,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Counter is a monotonically increasing metric
type Counter struct {
	name        string
	description string
	value       atomic.Uint64
	labels      map[string]string
	mu          sync.RWMutex
}

// Gauge is a metric that can go up and down
type Gauge struct {
	name        string
	description string
	value       atomic.Int64 // Store as int64 to allow atomic operations
	labels      map[string]string
	mu          sync.RWMutex
}

// Histogram tracks distribution of values
type Histogram struct {
	name        string
	description string
	buckets     []float64
	counts      []atomic.Uint64
	sum         atomic.Uint64
	count       atomic.Uint64
	labels      map[string]string
	mu          sync.RWMutex
}

// Summary tracks quantiles over time
type Summary struct {
	name        string
	description string
	values      []float64
	sum         atomic.Uint64
	count       atomic.Uint64
	labels      map[string]string
	mu          sync.RWMutex
}

// Collector collects and manages metrics
type Collector struct {
	counters   map[string]*Counter
	gauges     map[string]*Gauge
	histograms map[string]*Histogram
	summaries  map[string]*Summary
	mu         sync.RWMutex

	// System metrics
	startTime time.Time

	// Configuration
	config CollectorConfig
}

// CollectorConfig holds collector configuration
type CollectorConfig struct {
	CollectSystemMetrics bool
	SystemMetricsInterval time.Duration
	EnableHistory        bool
	HistorySize          int
	DefaultBuckets       []float64
}

// DefaultCollectorConfig returns default collector configuration
func DefaultCollectorConfig() CollectorConfig {
	return CollectorConfig{
		CollectSystemMetrics:  true,
		SystemMetricsInterval: 5 * time.Second,
		EnableHistory:         true,
		HistorySize:           100,
		DefaultBuckets:        []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}
}

// NewCollector creates a new metrics collector
func NewCollector(config CollectorConfig) *Collector {
	c := &Collector{
		counters:   make(map[string]*Counter),
		gauges:     make(map[string]*Gauge),
		histograms: make(map[string]*Histogram),
		summaries:  make(map[string]*Summary),
		startTime:  time.Now(),
		config:     config,
	}

	// Start system metrics collection
	if config.CollectSystemMetrics {
		go c.collectSystemMetrics(context.Background())
	}

	return c
}

// Counter methods

// NewCounter creates a new counter metric
func (c *Collector) NewCounter(name, description string, labels map[string]string) *Counter {
	c.mu.Lock()
	defer c.mu.Unlock()

	if counter, exists := c.counters[name]; exists {
		return counter
	}

	counter := &Counter{
		name:        name,
		description: description,
		labels:      labels,
	}
	c.counters[name] = counter
	return counter
}

// Inc increments the counter by 1
func (counter *Counter) Inc() {
	counter.value.Add(1)
}

// Add adds the given value to the counter
func (counter *Counter) Add(value uint64) {
	counter.value.Add(value)
}

// Get returns the current counter value
func (counter *Counter) Get() uint64 {
	return counter.value.Load()
}

// Reset resets the counter to zero
func (counter *Counter) Reset() {
	counter.value.Store(0)
}

// Gauge methods

// NewGauge creates a new gauge metric
func (c *Collector) NewGauge(name, description string, labels map[string]string) *Gauge {
	c.mu.Lock()
	defer c.mu.Unlock()

	if gauge, exists := c.gauges[name]; exists {
		return gauge
	}

	gauge := &Gauge{
		name:        name,
		description: description,
		labels:      labels,
	}
	c.gauges[name] = gauge
	return gauge
}

// Set sets the gauge to the given value
func (gauge *Gauge) Set(value int64) {
	gauge.value.Store(value)
}

// Inc increments the gauge by 1
func (gauge *Gauge) Inc() {
	gauge.value.Add(1)
}

// Dec decrements the gauge by 1
func (gauge *Gauge) Dec() {
	gauge.value.Add(-1)
}

// Add adds the given value to the gauge
func (gauge *Gauge) Add(value int64) {
	gauge.value.Add(value)
}

// Sub subtracts the given value from the gauge
func (gauge *Gauge) Sub(value int64) {
	gauge.value.Add(-value)
}

// Get returns the current gauge value
func (gauge *Gauge) Get() int64 {
	return gauge.value.Load()
}

// Histogram methods

// NewHistogram creates a new histogram metric
func (c *Collector) NewHistogram(name, description string, labels map[string]string, buckets []float64) *Histogram {
	c.mu.Lock()
	defer c.mu.Unlock()

	if histogram, exists := c.histograms[name]; exists {
		return histogram
	}

	if len(buckets) == 0 {
		buckets = c.config.DefaultBuckets
	}

	histogram := &Histogram{
		name:        name,
		description: description,
		buckets:     buckets,
		counts:      make([]atomic.Uint64, len(buckets)),
		labels:      labels,
	}
	c.histograms[name] = histogram
	return histogram
}

// Observe records a new observation
func (histogram *Histogram) Observe(value float64) {
	// Update sum and count
	histogram.sum.Add(uint64(value * 1000)) // Store as milliseconds
	histogram.count.Add(1)

	// Update buckets
	for i, bucket := range histogram.buckets {
		if value <= bucket {
			histogram.counts[i].Add(1)
		}
	}
}

// GetSum returns the sum of all observations
func (histogram *Histogram) GetSum() float64 {
	return float64(histogram.sum.Load()) / 1000.0
}

// GetCount returns the count of observations
func (histogram *Histogram) GetCount() uint64 {
	return histogram.count.Load()
}

// GetBuckets returns the bucket counts
func (histogram *Histogram) GetBuckets() map[float64]uint64 {
	histogram.mu.RLock()
	defer histogram.mu.RUnlock()

	buckets := make(map[float64]uint64)
	for i, bucket := range histogram.buckets {
		buckets[bucket] = histogram.counts[i].Load()
	}
	return buckets
}

// Summary methods

// NewSummary creates a new summary metric
func (c *Collector) NewSummary(name, description string, labels map[string]string) *Summary {
	c.mu.Lock()
	defer c.mu.Unlock()

	if summary, exists := c.summaries[name]; exists {
		return summary
	}

	summary := &Summary{
		name:        name,
		description: description,
		values:      make([]float64, 0, c.config.HistorySize),
		labels:      labels,
	}
	c.summaries[name] = summary
	return summary
}

// Observe records a new observation
func (summary *Summary) Observe(value float64) {
	summary.mu.Lock()
	defer summary.mu.Unlock()

	summary.sum.Add(uint64(value * 1000))
	summary.count.Add(1)

	// Keep limited history
	if len(summary.values) >= 100 {
		summary.values = summary.values[1:]
	}
	summary.values = append(summary.values, value)
}

// GetSum returns the sum of all observations
func (summary *Summary) GetSum() float64 {
	return float64(summary.sum.Load()) / 1000.0
}

// GetCount returns the count of observations
func (summary *Summary) GetCount() uint64 {
	return summary.count.Load()
}

// GetAverage returns the average value
func (summary *Summary) GetAverage() float64 {
	count := summary.count.Load()
	if count == 0 {
		return 0
	}
	return summary.GetSum() / float64(count)
}

// System metrics collection

func (c *Collector) collectSystemMetrics(ctx context.Context) {
	ticker := time.NewTicker(c.config.SystemMetricsInterval)
	defer ticker.Stop()

	// Create system metric gauges
	cpuGauge := c.NewGauge("system_cpu_percent", "CPU usage percentage", nil)
	memoryGauge := c.NewGauge("system_memory_bytes", "Memory usage in bytes", nil)
	goroutinesGauge := c.NewGauge("system_goroutines", "Number of goroutines", nil)
	gcPauseGauge := c.NewGauge("system_gc_pause_ns", "GC pause time in nanoseconds", nil)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			// Update metrics
			memoryGauge.Set(int64(m.Alloc))
			goroutinesGauge.Set(int64(runtime.NumGoroutine()))
			gcPauseGauge.Set(int64(m.PauseNs[(m.NumGC+255)%256]))

			// CPU is harder to measure accurately, set to 0 for now
			cpuGauge.Set(0)
		}
	}
}

// GetAllMetrics returns all collected metrics
func (c *Collector) GetAllMetrics() []Metric {
	c.mu.RLock()
	defer c.mu.RUnlock()

	metrics := make([]Metric, 0)
	now := time.Now()

	// Collect counters
	for _, counter := range c.counters {
		metrics = append(metrics, Metric{
			Name:        counter.name,
			Type:        TypeCounter,
			Value:       float64(counter.Get()),
			Labels:      counter.labels,
			Timestamp:   now,
			Description: counter.description,
		})
	}

	// Collect gauges
	for _, gauge := range c.gauges {
		metrics = append(metrics, Metric{
			Name:        gauge.name,
			Type:        TypeGauge,
			Value:       float64(gauge.Get()),
			Labels:      gauge.labels,
			Timestamp:   now,
			Description: gauge.description,
		})
	}

	// Collect histograms
	for _, histogram := range c.histograms {
		metrics = append(metrics, Metric{
			Name:        histogram.name,
			Type:        TypeHistogram,
			Value:       histogram.GetSum(),
			Labels:      histogram.labels,
			Timestamp:   now,
			Description: histogram.description,
			Metadata: map[string]interface{}{
				"count":   histogram.GetCount(),
				"buckets": histogram.GetBuckets(),
			},
		})
	}

	// Collect summaries
	for _, summary := range c.summaries {
		metrics = append(metrics, Metric{
			Name:        summary.name,
			Type:        TypeSummary,
			Value:       summary.GetSum(),
			Labels:      summary.labels,
			Timestamp:   now,
			Description: summary.description,
			Metadata: map[string]interface{}{
				"count":   summary.GetCount(),
				"average": summary.GetAverage(),
			},
		})
	}

	return metrics
}

// GetMetric returns a specific metric by name
func (c *Collector) GetMetric(name string) *Metric {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()

	// Check counters
	if counter, exists := c.counters[name]; exists {
		return &Metric{
			Name:        counter.name,
			Type:        TypeCounter,
			Value:       float64(counter.Get()),
			Labels:      counter.labels,
			Timestamp:   now,
			Description: counter.description,
		}
	}

	// Check gauges
	if gauge, exists := c.gauges[name]; exists {
		return &Metric{
			Name:        gauge.name,
			Type:        TypeGauge,
			Value:       float64(gauge.Get()),
			Labels:      gauge.labels,
			Timestamp:   now,
			Description: gauge.description,
		}
	}

	// Check histograms
	if histogram, exists := c.histograms[name]; exists {
		return &Metric{
			Name:        histogram.name,
			Type:        TypeHistogram,
			Value:       histogram.GetSum(),
			Labels:      histogram.labels,
			Timestamp:   now,
			Description: histogram.description,
			Metadata: map[string]interface{}{
				"count":   histogram.GetCount(),
				"buckets": histogram.GetBuckets(),
			},
		}
	}

	// Check summaries
	if summary, exists := c.summaries[name]; exists {
		return &Metric{
			Name:        summary.name,
			Type:        TypeSummary,
			Value:       summary.GetSum(),
			Labels:      summary.labels,
			Timestamp:   now,
			Description: summary.description,
			Metadata: map[string]interface{}{
				"count":   summary.GetCount(),
				"average": summary.GetAverage(),
			},
		}
	}

	return nil
}

// GetUptime returns the system uptime
func (c *Collector) GetUptime() time.Duration {
	return time.Since(c.startTime)
}

// Reset resets all metrics
func (c *Collector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, counter := range c.counters {
		counter.Reset()
	}
}

// Close stops the collector
func (c *Collector) Close() error {
	// Stop system metrics collection
	return nil
}

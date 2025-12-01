package metrics

import (
	"context"
	"encoding/json"
	"neonexcore/pkg/websocket"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Dashboard represents the real-time metrics dashboard
type Dashboard struct {
	collector *Collector
	hub       *websocket.Hub
	interval  time.Duration
	mu        sync.RWMutex

	// Alert configuration
	alerts []Alert
}

// Alert represents a metric alert
type Alert struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Metric      string                 `json:"metric"`
	Condition   AlertCondition         `json:"condition"`
	Threshold   float64                `json:"threshold"`
	Enabled     bool                   `json:"enabled"`
	LastFired   time.Time              `json:"last_fired,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AlertCondition represents alert trigger condition
type AlertCondition string

const (
	ConditionGreaterThan AlertCondition = "gt"
	ConditionLessThan    AlertCondition = "lt"
	ConditionEquals      AlertCondition = "eq"
	ConditionNotEquals   AlertCondition = "ne"
)

// DashboardConfig holds dashboard configuration
type DashboardConfig struct {
	BroadcastInterval time.Duration
	EnableAlerts      bool
	EnableHistory     bool
	HistorySize       int
}

// DefaultDashboardConfig returns default dashboard configuration
func DefaultDashboardConfig() DashboardConfig {
	return DashboardConfig{
		BroadcastInterval: 1 * time.Second,
		EnableAlerts:      true,
		EnableHistory:     true,
		HistorySize:       60,
	}
}

// NewDashboard creates a new metrics dashboard
func NewDashboard(collector *Collector, hub *websocket.Hub, config DashboardConfig) *Dashboard {
	d := &Dashboard{
		collector: collector,
		hub:       hub,
		interval:  config.BroadcastInterval,
		alerts:    make([]Alert, 0),
	}

	// Start broadcasting metrics
	go d.broadcastMetrics(context.Background())

	return d
}

// broadcastMetrics periodically broadcasts metrics to connected clients
func (d *Dashboard) broadcastMetrics(ctx context.Context) {
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics := d.collector.GetAllMetrics()
			data, err := json.Marshal(map[string]interface{}{
				"type":      "metrics",
				"timestamp": time.Now().Unix(),
				"uptime":    d.collector.GetUptime().Seconds(),
				"metrics":   metrics,
			})
			if err != nil {
				continue
			}

			// Broadcast to all connected clients
			if d.hub != nil {
				d.hub.BroadcastJSON(data)
			}

			// Check alerts
			d.checkAlerts(metrics)
		}
	}
}

// checkAlerts checks if any alerts should be fired
func (d *Dashboard) checkAlerts(metrics []Metric) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for i := range d.alerts {
		alert := &d.alerts[i]
		if !alert.Enabled {
			continue
		}

		// Find matching metric
		for _, metric := range metrics {
			if metric.Name != alert.Metric {
				continue
			}

			// Check condition
			shouldFire := false
			switch alert.Condition {
			case ConditionGreaterThan:
				shouldFire = metric.Value > alert.Threshold
			case ConditionLessThan:
				shouldFire = metric.Value < alert.Threshold
			case ConditionEquals:
				shouldFire = metric.Value == alert.Threshold
			case ConditionNotEquals:
				shouldFire = metric.Value != alert.Threshold
			}

			if shouldFire {
				d.fireAlert(alert, metric)
			}
		}
	}
}

// fireAlert fires an alert
func (d *Dashboard) fireAlert(alert *Alert, metric Metric) {
	// Prevent duplicate alerts within 1 minute
	if time.Since(alert.LastFired) < 1*time.Minute {
		return
	}

	alert.LastFired = time.Now()

	// Broadcast alert
	data, err := json.Marshal(map[string]interface{}{
		"type":      "alert",
		"timestamp": time.Now().Unix(),
		"alert":     alert,
		"metric":    metric,
	})
	if err != nil {
		return
	}

	if d.hub != nil {
		d.hub.BroadcastJSON(data)
	}
}

// AddAlert adds a new alert
func (d *Dashboard) AddAlert(alert Alert) {
	d.mu.Lock()
	defer d.mu.Unlock()

	alert.Enabled = true
	d.alerts = append(d.alerts, alert)
}

// RemoveAlert removes an alert by name
func (d *Dashboard) RemoveAlert(name string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for i, alert := range d.alerts {
		if alert.Name == name {
			d.alerts = append(d.alerts[:i], d.alerts[i+1:]...)
			return
		}
	}
}

// GetAlerts returns all alerts
func (d *Dashboard) GetAlerts() []Alert {
	d.mu.RLock()
	defer d.mu.RUnlock()

	alerts := make([]Alert, len(d.alerts))
	copy(alerts, d.alerts)
	return alerts
}

// SetupRoutes sets up dashboard HTTP routes
func (d *Dashboard) SetupRoutes(app *fiber.App) {
	// Get all metrics
	app.Get("/metrics", d.handleGetMetrics)

	// Get specific metric
	app.Get("/metrics/:name", d.handleGetMetric)

	// Get dashboard HTML
	app.Get("/metrics/dashboard", d.handleDashboard)

	// Alert management
	app.Get("/metrics/alerts", d.handleGetAlerts)
	app.Post("/metrics/alerts", d.handleAddAlert)
	app.Delete("/metrics/alerts/:name", d.handleDeleteAlert)
}

// handleGetMetrics returns all metrics as JSON
func (d *Dashboard) handleGetMetrics(c *fiber.Ctx) error {
	metrics := d.collector.GetAllMetrics()
	return c.JSON(fiber.Map{
		"success":   true,
		"timestamp": time.Now().Unix(),
		"uptime":    d.collector.GetUptime().Seconds(),
		"metrics":   metrics,
	})
}

// handleGetMetric returns a specific metric
func (d *Dashboard) handleGetMetric(c *fiber.Ctx) error {
	name := c.Params("name")
	metric := d.collector.GetMetric(name)

	if metric == nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"error":   "Metric not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"metric":  metric,
	})
}

// handleDashboard serves the dashboard HTML
func (d *Dashboard) handleDashboard(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html")
	return c.SendString(dashboardHTML)
}

// handleGetAlerts returns all alerts
func (d *Dashboard) handleGetAlerts(c *fiber.Ctx) error {
	alerts := d.GetAlerts()
	return c.JSON(fiber.Map{
		"success": true,
		"alerts":  alerts,
	})
}

// handleAddAlert adds a new alert
func (d *Dashboard) handleAddAlert(c *fiber.Ctx) error {
	var alert Alert
	if err := c.BodyParser(&alert); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	d.AddAlert(alert)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Alert added successfully",
		"alert":   alert,
	})
}

// handleDeleteAlert deletes an alert
func (d *Dashboard) handleDeleteAlert(c *fiber.Ctx) error {
	name := c.Params("name")
	d.RemoveAlert(name)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Alert deleted successfully",
	})
}

// Close stops the dashboard
func (d *Dashboard) Close() error {
	return nil
}

// dashboardHTML is the HTML template for the dashboard
const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NeonexCore Metrics Dashboard</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.0/dist/chart.umd.min.js"></script>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: #333;
            min-height: 100vh;
            padding: 20px;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
        }

        .header {
            text-align: center;
            color: white;
            margin-bottom: 30px;
        }

        .header h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
        }

        .status {
            display: inline-block;
            padding: 8px 16px;
            background: rgba(255,255,255,0.2);
            border-radius: 20px;
            font-size: 0.9em;
        }

        .status.connected {
            background: #10b981;
        }

        .status.disconnected {
            background: #ef4444;
        }

        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-bottom: 20px;
        }

        .card {
            background: white;
            border-radius: 12px;
            padding: 20px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
            transition: transform 0.2s;
        }

        .card:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 12px rgba(0,0,0,0.15);
        }

        .card-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 15px;
        }

        .card-title {
            font-size: 1.1em;
            font-weight: 600;
            color: #667eea;
        }

        .card-value {
            font-size: 2em;
            font-weight: bold;
            color: #333;
        }

        .card-unit {
            font-size: 0.5em;
            color: #666;
            margin-left: 5px;
        }

        .chart-container {
            position: relative;
            height: 200px;
            margin-top: 15px;
        }

        .metric-list {
            list-style: none;
        }

        .metric-item {
            display: flex;
            justify-content: space-between;
            padding: 10px 0;
            border-bottom: 1px solid #eee;
        }

        .metric-item:last-child {
            border-bottom: none;
        }

        .metric-name {
            color: #666;
            font-size: 0.9em;
        }

        .metric-value {
            font-weight: 600;
            color: #333;
        }

        .alert {
            background: #fef3c7;
            border-left: 4px solid #f59e0b;
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 15px;
            animation: slideIn 0.3s ease;
        }

        .alert-critical {
            background: #fee2e2;
            border-color: #ef4444;
        }

        @keyframes slideIn {
            from {
                opacity: 0;
                transform: translateX(-20px);
            }
            to {
                opacity: 1;
                transform: translateX(0);
            }
        }

        .badge {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 0.75em;
            font-weight: 600;
            text-transform: uppercase;
        }

        .badge-counter { background: #dbeafe; color: #1e40af; }
        .badge-gauge { background: #dcfce7; color: #166534; }
        .badge-histogram { background: #fef3c7; color: #92400e; }
        .badge-summary { background: #e0e7ff; color: #3730a3; }

        .footer {
            text-align: center;
            color: white;
            margin-top: 30px;
            font-size: 0.9em;
            opacity: 0.8;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⚡ NeonexCore Metrics Dashboard</h1>
            <div class="status disconnected" id="status">● Disconnected</div>
            <div style="margin-top: 10px;">
                <span>Uptime: <span id="uptime">--</span></span> | 
                <span>Last Update: <span id="lastUpdate">--</span></span>
            </div>
        </div>

        <div id="alerts"></div>

        <div class="grid">
            <div class="card">
                <div class="card-header">
                    <span class="card-title">💾 Memory Usage</span>
                    <span class="badge badge-gauge">gauge</span>
                </div>
                <div class="card-value" id="memory">--<span class="card-unit">MB</span></div>
                <div class="chart-container">
                    <canvas id="memoryChart"></canvas>
                </div>
            </div>

            <div class="card">
                <div class="card-header">
                    <span class="card-title">🚀 Goroutines</span>
                    <span class="badge badge-gauge">gauge</span>
                </div>
                <div class="card-value" id="goroutines">--</div>
                <div class="chart-container">
                    <canvas id="goroutinesChart"></canvas>
                </div>
            </div>

            <div class="card">
                <div class="card-header">
                    <span class="card-title">🔥 CPU Usage</span>
                    <span class="badge badge-gauge">gauge</span>
                </div>
                <div class="card-value" id="cpu">--<span class="card-unit">%</span></div>
                <div class="chart-container">
                    <canvas id="cpuChart"></canvas>
                </div>
            </div>

            <div class="card">
                <div class="card-header">
                    <span class="card-title">🗑️ GC Pause</span>
                    <span class="badge badge-gauge">gauge</span>
                </div>
                <div class="card-value" id="gcPause">--<span class="card-unit">μs</span></div>
                <div class="chart-container">
                    <canvas id="gcChart"></canvas>
                </div>
            </div>
        </div>

        <div class="card">
            <div class="card-header">
                <span class="card-title">📊 All Metrics</span>
            </div>
            <ul class="metric-list" id="metricsList"></ul>
        </div>

        <div class="footer">
            Powered by NeonexCore Framework | Real-time metrics via WebSocket
        </div>
    </div>

    <script>
        // WebSocket connection
        let ws = null;
        let reconnectInterval = null;
        const statusEl = document.getElementById('status');

        // Chart configurations
        const chartConfig = {
            type: 'line',
            options: {
                responsive: true,
                maintainAspectRatio: false,
                animation: { duration: 500 },
                scales: {
                    y: { beginAtZero: true }
                },
                plugins: {
                    legend: { display: false }
                }
            }
        };

        // Initialize charts
        const memoryChart = new Chart(document.getElementById('memoryChart'), {
            ...chartConfig,
            data: {
                labels: [],
                datasets: [{
                    label: 'Memory (MB)',
                    data: [],
                    borderColor: '#667eea',
                    backgroundColor: 'rgba(102, 126, 234, 0.1)',
                    fill: true
                }]
            }
        });

        const goroutinesChart = new Chart(document.getElementById('goroutinesChart'), {
            ...chartConfig,
            data: {
                labels: [],
                datasets: [{
                    label: 'Goroutines',
                    data: [],
                    borderColor: '#10b981',
                    backgroundColor: 'rgba(16, 185, 129, 0.1)',
                    fill: true
                }]
            }
        });

        const cpuChart = new Chart(document.getElementById('cpuChart'), {
            ...chartConfig,
            data: {
                labels: [],
                datasets: [{
                    label: 'CPU %',
                    data: [],
                    borderColor: '#f59e0b',
                    backgroundColor: 'rgba(245, 158, 11, 0.1)',
                    fill: true
                }]
            }
        });

        const gcChart = new Chart(document.getElementById('gcChart'), {
            ...chartConfig,
            data: {
                labels: [],
                datasets: [{
                    label: 'GC Pause (μs)',
                    data: [],
                    borderColor: '#ef4444',
                    backgroundColor: 'rgba(239, 68, 68, 0.1)',
                    fill: true
                }]
            }
        });

        function connect() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = protocol + '//' + window.location.host + '/ws';
            
            ws = new WebSocket(wsUrl);

            ws.onopen = () => {
                console.log('✅ Connected to metrics stream');
                statusEl.textContent = '● Connected';
                statusEl.className = 'status connected';
                if (reconnectInterval) {
                    clearInterval(reconnectInterval);
                    reconnectInterval = null;
                }
            };

            ws.onclose = () => {
                console.log('❌ Disconnected from metrics stream');
                statusEl.textContent = '● Disconnected';
                statusEl.className = 'status disconnected';
                
                if (!reconnectInterval) {
                    reconnectInterval = setInterval(() => {
                        console.log('🔄 Attempting to reconnect...');
                        connect();
                    }, 3000);
                }
            };

            ws.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    
                    if (data.type === 'metrics') {
                        updateMetrics(data);
                    } else if (data.type === 'alert') {
                        showAlert(data);
                    }
                } catch (error) {
                    console.error('Error parsing message:', error);
                }
            };
        }

        function updateMetrics(data) {
            // Update uptime
            document.getElementById('uptime').textContent = formatDuration(data.uptime);
            document.getElementById('lastUpdate').textContent = new Date().toLocaleTimeString();

            const metrics = data.metrics || [];
            const now = new Date().toLocaleTimeString();

            // Update specific metrics
            metrics.forEach(metric => {
                switch (metric.name) {
                    case 'system_memory_bytes':
                        const memoryMB = (metric.value / 1024 / 1024).toFixed(2);
                        document.getElementById('memory').innerHTML = memoryMB + '<span class="card-unit">MB</span>';
                        updateChart(memoryChart, now, memoryMB);
                        break;
                    case 'system_goroutines':
                        document.getElementById('goroutines').textContent = metric.value;
                        updateChart(goroutinesChart, now, metric.value);
                        break;
                    case 'system_cpu_percent':
                        document.getElementById('cpu').innerHTML = metric.value.toFixed(2) + '<span class="card-unit">%</span>';
                        updateChart(cpuChart, now, metric.value);
                        break;
                    case 'system_gc_pause_ns':
                        const pauseMicro = (metric.value / 1000).toFixed(2);
                        document.getElementById('gcPause').innerHTML = pauseMicro + '<span class="card-unit">μs</span>';
                        updateChart(gcChart, now, pauseMicro);
                        break;
                }
            });

            // Update metrics list
            const metricsList = document.getElementById('metricsList');
            metricsList.innerHTML = metrics.map(metric => `
                <li class="metric-item">
                    <span class="metric-name">
                        <span class="badge badge-${metric.type}">${metric.type}</span>
                        ${metric.name}
                    </span>
                    <span class="metric-value">${formatValue(metric.value, metric.type)}</span>
                </li>
            `).join('');
        }

        function updateChart(chart, label, value) {
            if (chart.data.labels.length > 60) {
                chart.data.labels.shift();
                chart.data.datasets[0].data.shift();
            }
            chart.data.labels.push(label);
            chart.data.datasets[0].data.push(value);
            chart.update('none');
        }

        function showAlert(data) {
            const alertsDiv = document.getElementById('alerts');
            const alert = data.alert;
            const isCritical = alert.condition === 'gt' && data.metric.value > alert.threshold * 1.5;
            
            const alertEl = document.createElement('div');
            alertEl.className = 'alert ' + (isCritical ? 'alert-critical' : '');
            alertEl.innerHTML = `
                <strong>⚠️ ${alert.name}</strong><br>
                ${alert.description} (${data.metric.name}: ${formatValue(data.metric.value)})
            `;
            
            alertsDiv.insertBefore(alertEl, alertsDiv.firstChild);
            
            setTimeout(() => {
                alertEl.remove();
            }, 10000);
        }

        function formatDuration(seconds) {
            const hours = Math.floor(seconds / 3600);
            const minutes = Math.floor((seconds % 3600) / 60);
            const secs = Math.floor(seconds % 60);
            return hours + 'h ' + minutes + 'm ' + secs + 's';
        }

        function formatValue(value, type) {
            if (type === 'counter' || type === 'gauge') {
                return typeof value === 'number' ? value.toFixed(2) : value;
            }
            return value;
        }

        // Connect on load
        connect();
    </script>
</body>
</html>`

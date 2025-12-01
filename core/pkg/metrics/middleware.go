package metrics

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// Middleware creates a Fiber middleware for collecting HTTP metrics
func Middleware(collector *Collector) fiber.Handler {
	// Create metrics
	requestCounter := collector.NewCounter(
		"http_requests_total",
		"Total number of HTTP requests",
		nil,
	)

	requestDuration := collector.NewHistogram(
		"http_request_duration_seconds",
		"HTTP request duration in seconds",
		nil,
		[]float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2, 5},
	)

	activeRequests := collector.NewGauge(
		"http_requests_active",
		"Number of active HTTP requests",
		nil,
	)

	requestSize := collector.NewHistogram(
		"http_request_size_bytes",
		"HTTP request size in bytes",
		nil,
		[]float64{100, 1000, 10000, 100000, 1000000},
	)

	responseSize := collector.NewHistogram(
		"http_response_size_bytes",
		"HTTP response size in bytes",
		nil,
		[]float64{100, 1000, 10000, 100000, 1000000},
	)

	statusCounter := make(map[int]*Counter)
	for _, status := range []int{200, 201, 204, 400, 401, 403, 404, 500, 502, 503} {
		statusCounter[status] = collector.NewCounter(
			"http_responses_"+string(rune(status)),
			"Number of HTTP responses with status "+string(rune(status)),
			map[string]string{"status": string(rune(status))},
		)
	}

	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Increment active requests
		activeRequests.Inc()
		defer activeRequests.Dec()

		// Track request size
		if c.Request().Header.ContentLength() > 0 {
			requestSize.Observe(float64(c.Request().Header.ContentLength()))
		}

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Update metrics
		requestCounter.Inc()
		requestDuration.Observe(duration)

		// Track response size
		responseSize.Observe(float64(len(c.Response().Body())))

		// Track status code
		status := c.Response().StatusCode()
		if counter, exists := statusCounter[status]; exists {
			counter.Inc()
		}

		return err
	}
}

// MethodMiddleware creates middleware that tracks metrics by HTTP method
func MethodMiddleware(collector *Collector) fiber.Handler {
	counters := make(map[string]*Counter)
	for _, method := range []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"} {
		counters[method] = collector.NewCounter(
			"http_requests_"+method,
			"Number of "+method+" requests",
			map[string]string{"method": method},
		)
	}

	return func(c *fiber.Ctx) error {
		method := c.Method()
		if counter, exists := counters[method]; exists {
			counter.Inc()
		}
		return c.Next()
	}
}

// PathMiddleware creates middleware that tracks metrics by path pattern
func PathMiddleware(collector *Collector, paths []string) fiber.Handler {
	counters := make(map[string]*Counter)
	for _, path := range paths {
		counters[path] = collector.NewCounter(
			"http_requests_path_"+path,
			"Number of requests to "+path,
			map[string]string{"path": path},
		)
	}

	return func(c *fiber.Ctx) error {
		path := c.Path()
		if counter, exists := counters[path]; exists {
			counter.Inc()
		}
		return c.Next()
	}
}

// ErrorMiddleware creates middleware that tracks error metrics
func ErrorMiddleware(collector *Collector) fiber.Handler {
	errorCounter := collector.NewCounter(
		"http_errors_total",
		"Total number of HTTP errors",
		nil,
	)

	serverErrorCounter := collector.NewCounter(
		"http_errors_5xx",
		"Number of 5xx server errors",
		nil,
	)

	clientErrorCounter := collector.NewCounter(
		"http_errors_4xx",
		"Number of 4xx client errors",
		nil,
	)

	return func(c *fiber.Ctx) error {
		err := c.Next()

		status := c.Response().StatusCode()

		if status >= 400 {
			errorCounter.Inc()

			if status >= 500 {
				serverErrorCounter.Inc()
			} else if status >= 400 {
				clientErrorCounter.Inc()
			}
		}

		return err
	}
}

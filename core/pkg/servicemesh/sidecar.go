package servicemesh

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// SidecarProxy represents a service mesh sidecar proxy
type SidecarProxy struct {
	serviceName    string
	servicePort    int
	proxyPort      int
	controlPlane   string
	config         *SidecarConfig
	metrics        *ProxyMetrics
	registry       *ServiceRegistry
	tlsConfig      *tls.Config
	routingRules   map[string]*RoutingRule
	circuitBreaker *CircuitBreaker
	mu             sync.RWMutex
	app            *fiber.App
	shutdown       chan struct{}
}

// SidecarConfig configuration for sidecar proxy
type SidecarConfig struct {
	ServiceName       string
	ServicePort       int
	ProxyPort         int
	ControlPlane      string
	EnableMTLS        bool
	EnableTracing     bool
	EnableMetrics     bool
	EnableRetry       bool
	MaxRetries        int
	RetryTimeout      time.Duration
	CircuitBreakerCfg *CircuitBreakerConfig
	TLSCertFile       string
	TLSKeyFile        string
	TLSCAFile         string
}

// ProxyMetrics metrics collected by sidecar
type ProxyMetrics struct {
	RequestsTotal      int64
	RequestsSuccess    int64
	RequestsFailed     int64
	RequestDuration    []time.Duration
	BytesSent          int64
	BytesReceived      int64
	ActiveConnections  int64
	CircuitBreakerOpen int64
	RetriesTotal       int64
	mu                 sync.RWMutex
}

// RoutingRule defines routing rules for traffic management
type RoutingRule struct {
	ServiceName string
	Version     string
	Weight      int
	Headers     map[string]string
	PathPrefix  string
	Timeout     time.Duration
	RetryPolicy *RetryPolicy
}

// RetryPolicy retry configuration
type RetryPolicy struct {
	MaxAttempts int
	PerTryTimeout time.Duration
	RetryOn []string // HTTP status codes or error types
}

// NewSidecarProxy creates a new sidecar proxy
func NewSidecarProxy(config *SidecarConfig) (*SidecarProxy, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}

	proxy := &SidecarProxy{
		serviceName:  config.ServiceName,
		servicePort:  config.ServicePort,
		proxyPort:    config.ProxyPort,
		controlPlane: config.ControlPlane,
		config:       config,
		metrics:      &ProxyMetrics{},
		routingRules: make(map[string]*RoutingRule),
		shutdown:     make(chan struct{}),
	}

	// Initialize TLS if enabled
	if config.EnableMTLS {
		tlsConfig, err := proxy.setupMTLS()
		if err != nil {
			return nil, fmt.Errorf("failed to setup mTLS: %w", err)
		}
		proxy.tlsConfig = tlsConfig
	}

	// Initialize circuit breaker
	if config.CircuitBreakerCfg != nil {
		proxy.circuitBreaker = NewCircuitBreaker(config.CircuitBreakerCfg)
	}

	// Initialize service registry
	proxy.registry = NewServiceRegistry(config.ControlPlane)

	// Setup Fiber app for proxy
	proxy.app = fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	proxy.setupRoutes()

	return proxy, nil
}

// setupMTLS configures mutual TLS
func (s *SidecarProxy) setupMTLS() (*tls.Config, error) {
	// Load CA certificate
	caCert, err := os.ReadFile(s.config.TLSCAFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA cert: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA cert")
	}

	// Load client certificate
	cert, err := tls.LoadX509KeyPair(s.config.TLSCertFile, s.config.TLSKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client cert: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS13,
	}, nil
}

// setupRoutes configures proxy routes
func (s *SidecarProxy) setupRoutes() {
	// Health check endpoint
	s.app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"service": s.serviceName,
		})
	})

	// Metrics endpoint
	s.app.Get("/metrics", func(c *fiber.Ctx) error {
		return c.JSON(s.GetMetrics())
	})

	// Proxy all other requests
	s.app.All("/*", s.proxyHandler)
}

// proxyHandler handles proxying requests
func (s *SidecarProxy) proxyHandler(c *fiber.Ctx) error {
	startTime := time.Now()
	
	s.metrics.mu.Lock()
	s.metrics.RequestsTotal++
	s.metrics.ActiveConnections++
	s.metrics.mu.Unlock()

	defer func() {
		s.metrics.mu.Lock()
		s.metrics.ActiveConnections--
		s.metrics.RequestDuration = append(s.metrics.RequestDuration, time.Since(startTime))
		s.metrics.mu.Unlock()
	}()

	// Extract target service from headers or path
	targetService := c.Get("X-Target-Service")
	if targetService == "" {
		targetService = s.serviceName
	}

	// Get routing rule
	rule := s.getRoutingRule(targetService)

	// Check circuit breaker
	if s.circuitBreaker != nil && s.circuitBreaker.IsOpen() {
		s.metrics.mu.Lock()
		s.metrics.CircuitBreakerOpen++
		s.metrics.mu.Unlock()
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "circuit breaker is open",
		})
	}

	// Discover service instance
	instance, err := s.registry.Discover(targetService)
	if err != nil {
		s.recordFailure()
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": fmt.Sprintf("service discovery failed: %v", err),
		})
	}

	// Build target URL
	targetURL := fmt.Sprintf("%s://%s:%d%s",
		instance.Protocol,
		instance.Host,
		instance.Port,
		c.Path(),
	)

	// Perform request with retries
	var resp *http.Response
	var lastErr error
	
	maxRetries := 1
	if s.config.EnableRetry && rule != nil && rule.RetryPolicy != nil {
		maxRetries = rule.RetryPolicy.MaxAttempts
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			s.metrics.mu.Lock()
			s.metrics.RetriesTotal++
			s.metrics.mu.Unlock()
			time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
		}

		resp, lastErr = s.forwardRequest(c, targetURL, rule)
		if lastErr == nil && resp.StatusCode < 500 {
			break
		}
	}

	if lastErr != nil {
		s.recordFailure()
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to forward request: %v", lastErr),
		})
	}

	defer resp.Body.Close()

	// Record success
	s.recordSuccess()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Response().Header.Add(key, value)
		}
	}

	// Copy response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "failed to read response",
		})
	}

	s.metrics.mu.Lock()
	s.metrics.BytesReceived += int64(len(body))
	s.metrics.mu.Unlock()

	c.Status(resp.StatusCode)
	return c.Send(body)
}

// forwardRequest forwards HTTP request to target service
func (s *SidecarProxy) forwardRequest(c *fiber.Ctx, targetURL string, rule *RoutingRule) (*http.Response, error) {
	// Create HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	if rule != nil && rule.Timeout > 0 {
		client.Timeout = rule.Timeout
	}

	if s.tlsConfig != nil {
		client.Transport = &http.Transport{
			TLSClientConfig: s.tlsConfig,
		}
	}

	// Create request
	req, err := http.NewRequest(c.Method(), targetURL, c.Context().RequestBodyStream())
	if err != nil {
		return nil, err
	}

	// Copy headers
	c.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})

	// Add tracing headers if enabled
	if s.config.EnableTracing {
		req.Header.Set("X-Request-ID", c.Get("X-Request-ID", generateRequestID()))
		req.Header.Set("X-B3-TraceId", generateTraceID())
		req.Header.Set("X-B3-SpanId", generateSpanID())
	}

	// Add service mesh headers
	req.Header.Set("X-Mesh-Service", s.serviceName)
	req.Header.Set("X-Mesh-Version", "1.0")

	s.metrics.mu.Lock()
	s.metrics.BytesSent += int64(c.Request().Header.ContentLength())
	s.metrics.mu.Unlock()

	return client.Do(req)
}

// AddRoutingRule adds a routing rule
func (s *SidecarProxy) AddRoutingRule(serviceName string, rule *RoutingRule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.routingRules[serviceName] = rule
}

// getRoutingRule gets routing rule for service
func (s *SidecarProxy) getRoutingRule(serviceName string) *RoutingRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.routingRules[serviceName]
}

// recordSuccess records successful request
func (s *SidecarProxy) recordSuccess() {
	s.metrics.mu.Lock()
	s.metrics.RequestsSuccess++
	s.metrics.mu.Unlock()

	if s.circuitBreaker != nil {
		s.circuitBreaker.RecordSuccess()
	}
}

// recordFailure records failed request
func (s *SidecarProxy) recordFailure() {
	s.metrics.mu.Lock()
	s.metrics.RequestsFailed++
	s.metrics.mu.Unlock()

	if s.circuitBreaker != nil {
		s.circuitBreaker.RecordFailure()
	}
}

// GetMetrics returns proxy metrics
func (s *SidecarProxy) GetMetrics() map[string]interface{} {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()

	avgDuration := time.Duration(0)
	if len(s.metrics.RequestDuration) > 0 {
		var total time.Duration
		for _, d := range s.metrics.RequestDuration {
			total += d
		}
		avgDuration = total / time.Duration(len(s.metrics.RequestDuration))
	}

	return map[string]interface{}{
		"requests_total":        s.metrics.RequestsTotal,
		"requests_success":      s.metrics.RequestsSuccess,
		"requests_failed":       s.metrics.RequestsFailed,
		"avg_duration_ms":       avgDuration.Milliseconds(),
		"bytes_sent":            s.metrics.BytesSent,
		"bytes_received":        s.metrics.BytesReceived,
		"active_connections":    s.metrics.ActiveConnections,
		"circuit_breaker_open":  s.circuitBreaker != nil && s.circuitBreaker.IsOpen(),
		"retries_total":         s.metrics.RetriesTotal,
	}
}

// Start starts the sidecar proxy
func (s *SidecarProxy) Start() error {
	log.Printf("Starting sidecar proxy for %s on port %d", s.serviceName, s.proxyPort)
	
	// Register service with control plane
	if err := s.registry.Register(&ServiceInstance{
		ServiceName: s.serviceName,
		Host:        "localhost",
		Port:        s.servicePort,
		Protocol:    "http",
		Metadata:    map[string]string{"version": "1.0"},
	}); err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	// Start heartbeat
	go s.heartbeat()

	return s.app.Listen(fmt.Sprintf(":%d", s.proxyPort))
}

// heartbeat sends periodic heartbeats to control plane
func (s *SidecarProxy) heartbeat() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.registry.Heartbeat(s.serviceName); err != nil {
				log.Printf("Heartbeat failed: %v", err)
			}
		case <-s.shutdown:
			return
		}
	}
}

// Stop stops the sidecar proxy
func (s *SidecarProxy) Stop(ctx context.Context) error {
	close(s.shutdown)
	
	// Deregister from control plane
	if err := s.registry.Deregister(s.serviceName); err != nil {
		log.Printf("Failed to deregister: %v", err)
	}

	return s.app.ShutdownWithContext(ctx)
}

// Helper functions for tracing
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func generateTraceID() string {
	return fmt.Sprintf("%016x", time.Now().UnixNano())
}

func generateSpanID() string {
	return fmt.Sprintf("%08x", time.Now().UnixNano()&0xffffffff)
}

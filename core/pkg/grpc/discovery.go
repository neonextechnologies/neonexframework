package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ServiceRegistry manages service discovery and registration
type ServiceRegistry struct {
	services map[string]*ServiceInfo
	mu       sync.RWMutex
}

// ServiceInfo holds information about a registered service
type ServiceInfo struct {
	Name      string            `json:"name"`
	Address   string            `json:"address"`
	Version   string            `json:"version"`
	Health    HealthStatus      `json:"health"`
	Metadata  map[string]string `json:"metadata"`
	LastSeen  time.Time         `json:"last_seen"`
	Endpoints []string          `json:"endpoints"`
}

// HealthStatus represents service health
type HealthStatus string

const (
	HealthHealthy   HealthStatus = "healthy"
	HealthUnhealthy HealthStatus = "unhealthy"
	HealthUnknown   HealthStatus = "unknown"
)

// NewServiceRegistry creates a new service registry
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]*ServiceInfo),
	}
}

// Register registers a service
func (r *ServiceRegistry) Register(info ServiceInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if info.Name == "" {
		return fmt.Errorf("service name is required")
	}

	if info.Address == "" {
		return fmt.Errorf("service address is required")
	}

	info.LastSeen = time.Now()
	if info.Health == "" {
		info.Health = HealthHealthy
	}

	r.services[info.Name] = &info

	fmt.Printf("✅ Registered service: %s at %s\n", info.Name, info.Address)
	return nil
}

// Deregister removes a service from the registry
func (r *ServiceRegistry) Deregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.services[name]; !exists {
		return fmt.Errorf("service not found: %s", name)
	}

	delete(r.services, name)
	fmt.Printf("❌ Deregistered service: %s\n", name)
	return nil
}

// Get retrieves a service by name
func (r *ServiceRegistry) Get(name string) (*ServiceInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[name]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", name)
	}

	return service, nil
}

// List returns all registered services
func (r *ServiceRegistry) List() []*ServiceInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]*ServiceInfo, 0, len(r.services))
	for _, service := range r.services {
		services = append(services, service)
	}

	return services
}

// UpdateHealth updates service health status
func (r *ServiceRegistry) UpdateHealth(name string, health HealthStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	service, exists := r.services[name]
	if !exists {
		return fmt.Errorf("service not found: %s", name)
	}

	service.Health = health
	service.LastSeen = time.Now()

	return nil
}

// GetHealthy returns all healthy services
func (r *ServiceRegistry) GetHealthy() []*ServiceInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]*ServiceInfo, 0)
	for _, service := range r.services {
		if service.Health == HealthHealthy {
			services = append(services, service)
		}
	}

	return services
}

// StartHealthCheck starts periodic health checks
func (r *ServiceRegistry) StartHealthCheck(ctx context.Context, interval time.Duration, checker func(service *ServiceInfo) bool) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.performHealthCheck(checker)
		}
	}
}

func (r *ServiceRegistry) performHealthCheck(checker func(service *ServiceInfo) bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, service := range r.services {
		healthy := checker(service)
		
		if healthy {
			service.Health = HealthHealthy
		} else {
			service.Health = HealthUnhealthy
		}
		
		service.LastSeen = time.Now()
	}
}

// LoadBalancer provides load balancing across service instances
type LoadBalancer struct {
	registry *ServiceRegistry
	strategy LoadBalanceStrategy
	current  map[string]int
	mu       sync.RWMutex
}

// LoadBalanceStrategy defines load balancing strategy
type LoadBalanceStrategy string

const (
	StrategyRoundRobin LoadBalanceStrategy = "round_robin"
	StrategyRandom     LoadBalanceStrategy = "random"
	StrategyLeastConn  LoadBalanceStrategy = "least_conn"
)

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(registry *ServiceRegistry, strategy LoadBalanceStrategy) *LoadBalancer {
	return &LoadBalancer{
		registry: registry,
		strategy: strategy,
		current:  make(map[string]int),
	}
}

// GetService returns a service instance based on load balancing strategy
func (lb *LoadBalancer) GetService(name string) (*ServiceInfo, error) {
	// Get all healthy instances
	lb.mu.RLock()
	allServices := lb.registry.GetHealthy()
	lb.mu.RUnlock()

	// Filter by name
	instances := make([]*ServiceInfo, 0)
	for _, service := range allServices {
		if service.Name == name {
			instances = append(instances, service)
		}
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("no healthy instances found for service: %s", name)
	}

	// Apply strategy
	switch lb.strategy {
	case StrategyRoundRobin:
		return lb.roundRobin(name, instances), nil
	case StrategyRandom:
		return lb.random(instances), nil
	case StrategyLeastConn:
		return lb.leastConn(instances), nil
	default:
		return instances[0], nil
	}
}

func (lb *LoadBalancer) roundRobin(name string, instances []*ServiceInfo) *ServiceInfo {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	idx := lb.current[name]
	service := instances[idx]
	lb.current[name] = (idx + 1) % len(instances)

	return service
}

func (lb *LoadBalancer) random(instances []*ServiceInfo) *ServiceInfo {
	// Simple random selection (deterministic for demo)
	return instances[time.Now().UnixNano()%int64(len(instances))]
}

func (lb *LoadBalancer) leastConn(instances []*ServiceInfo) *ServiceInfo {
	// For simplicity, return first instance
	// In production, track actual connection counts
	return instances[0]
}

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	maxFailures  int
	timeout      time.Duration
	failures     int
	lastFailTime time.Time
	state        CircuitState
	mu           sync.RWMutex
}

// CircuitState represents circuit breaker state
type CircuitState string

const (
	StateClosed    CircuitState = "closed"
	StateOpen      CircuitState = "open"
	StateHalfOpen  CircuitState = "half_open"
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures: maxFailures,
		timeout:     timeout,
		state:       StateClosed,
	}
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()
	
	// Check if circuit is open
	if cb.state == StateOpen {
		if time.Since(cb.lastFailTime) > cb.timeout {
			// Transition to half-open
			cb.state = StateHalfOpen
			cb.failures = 0
		} else {
			cb.mu.Unlock()
			return fmt.Errorf("circuit breaker is open")
		}
	}
	
	cb.mu.Unlock()

	// Execute function
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailTime = time.Now()

		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
			fmt.Printf("⚠️  Circuit breaker opened (failures: %d)\n", cb.failures)
		}

		return err
	}

	// Success - reset circuit
	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		cb.failures = 0
		fmt.Println("✅ Circuit breaker closed")
	}

	return nil
}

// GetState returns current circuit state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset manually resets the circuit breaker
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	fmt.Println("🔄 Circuit breaker reset")
}

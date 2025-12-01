package servicemesh

import (
	"sync"
	"time"
)

// CircuitBreakerState represents circuit breaker state
type CircuitBreakerState string

const (
	StateClosed    CircuitBreakerState = "closed"
	StateOpen      CircuitBreakerState = "open"
	StateHalfOpen  CircuitBreakerState = "half_open"
)

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	config           *CircuitBreakerConfig
	state            CircuitBreakerState
	failureCount     int
	successCount     int
	lastFailureTime  time.Time
	lastStateChange  time.Time
	mu               sync.RWMutex
}

// CircuitBreakerConfig configuration for circuit breaker
type CircuitBreakerConfig struct {
	FailureThreshold int           // Number of failures before opening
	SuccessThreshold int           // Number of successes before closing from half-open
	Timeout          time.Duration // Time to wait before half-open
	HalfOpenRequests int           // Max requests allowed in half-open state
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = &CircuitBreakerConfig{
			FailureThreshold: 5,
			SuccessThreshold: 2,
			Timeout:          60 * time.Second,
			HalfOpenRequests: 3,
		}
	}

	return &CircuitBreaker{
		config:          config,
		state:           StateClosed,
		lastStateChange: time.Now(),
	}
}

// IsOpen checks if circuit breaker is open
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	// Check if we should transition from open to half-open
	if cb.state == StateOpen {
		if time.Since(cb.lastStateChange) >= cb.config.Timeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = StateHalfOpen
			cb.successCount = 0
			cb.failureCount = 0
			cb.lastStateChange = time.Now()
			cb.mu.Unlock()
			cb.mu.RLock()
			return false
		}
		return true
	}

	return false
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.config.SuccessThreshold {
			cb.state = StateClosed
			cb.failureCount = 0
			cb.successCount = 0
			cb.lastStateChange = time.Now()
		}
	} else if cb.state == StateClosed {
		// Reset failure count on success
		cb.failureCount = 0
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.state == StateHalfOpen {
		// Go back to open on any failure in half-open
		cb.state = StateOpen
		cb.lastStateChange = time.Now()
	} else if cb.state == StateClosed {
		if cb.failureCount >= cb.config.FailureThreshold {
			cb.state = StateOpen
			cb.lastStateChange = time.Now()
		}
	}
}

// GetState returns current state
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetMetrics returns circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":             cb.state,
		"failure_count":     cb.failureCount,
		"success_count":     cb.successCount,
		"last_failure_time": cb.lastFailureTime,
		"last_state_change": cb.lastStateChange,
		"time_in_state":     time.Since(cb.lastStateChange).Seconds(),
	}
}

// Reset resets circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.lastStateChange = time.Now()
}

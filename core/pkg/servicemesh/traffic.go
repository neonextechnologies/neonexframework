package servicemesh

import (
	"fmt"
	"sync"
)

// TrafficManager manages traffic routing and load balancing
type TrafficManager struct {
	policies map[string]*TrafficPolicy
	mu       sync.RWMutex
}

// TrafficPolicy defines traffic routing policy
type TrafficPolicy struct {
	ServiceName string
	Strategy    LoadBalancingStrategy
	Splits      []TrafficSplit
	Canary      *CanaryConfig
	ABTest      *ABTestConfig
}

// LoadBalancingStrategy load balancing strategies
type LoadBalancingStrategy string

const (
	StrategyRoundRobin    LoadBalancingStrategy = "round_robin"
	StrategyRandom        LoadBalancingStrategy = "random"
	StrategyLeastConn     LoadBalancingStrategy = "least_conn"
	StrategyWeightedRR    LoadBalancingStrategy = "weighted_round_robin"
	StrategyIPHash        LoadBalancingStrategy = "ip_hash"
)

// TrafficSplit splits traffic between versions
type TrafficSplit struct {
	Version string
	Weight  int // Percentage 0-100
	Headers map[string]string
}

// CanaryConfig configuration for canary deployment
type CanaryConfig struct {
	Enabled        bool
	NewVersion     string
	StableVersion  string
	InitialWeight  int // Starting percentage for new version
	IncrementStep  int // Percentage to increment per step
	IncrementDelay int // Seconds between increments
	MaxWeight      int // Maximum percentage for new version
	SuccessRate    float64 // Required success rate to continue
}

// ABTestConfig configuration for A/B testing
type ABTestConfig struct {
	Enabled  bool
	VersionA string
	VersionB string
	SplitKey string // Header or cookie to use for splitting
	WeightA  int    // Percentage for version A
	WeightB  int    // Percentage for version B
}

// NewTrafficManager creates a new traffic manager
func NewTrafficManager() *TrafficManager {
	return &TrafficManager{
		policies: make(map[string]*TrafficPolicy),
	}
}

// SetPolicy sets traffic policy for a service
func (tm *TrafficManager) SetPolicy(policy *TrafficPolicy) error {
	if policy == nil || policy.ServiceName == "" {
		return fmt.Errorf("invalid policy")
	}

	// Validate weights
	if len(policy.Splits) > 0 {
		totalWeight := 0
		for _, split := range policy.Splits {
			totalWeight += split.Weight
		}
		if totalWeight != 100 {
			return fmt.Errorf("traffic split weights must sum to 100, got %d", totalWeight)
		}
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.policies[policy.ServiceName] = policy
	return nil
}

// GetPolicy gets traffic policy for a service
func (tm *TrafficManager) GetPolicy(serviceName string) *TrafficPolicy {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.policies[serviceName]
}

// SelectVersion selects version based on policy
func (tm *TrafficManager) SelectVersion(serviceName string, headers map[string]string, clientIP string) string {
	policy := tm.GetPolicy(serviceName)
	if policy == nil {
		return "" // Use default version
	}

	// A/B Testing
	if policy.ABTest != nil && policy.ABTest.Enabled {
		return tm.selectABVersion(policy.ABTest, headers)
	}

	// Canary Deployment
	if policy.Canary != nil && policy.Canary.Enabled {
		return tm.selectCanaryVersion(policy.Canary)
	}

	// Traffic Splitting
	if len(policy.Splits) > 0 {
		return tm.selectSplitVersion(policy.Splits, headers)
	}

	return ""
}

// selectABVersion selects version for A/B testing
func (tm *TrafficManager) selectABVersion(config *ABTestConfig, headers map[string]string) string {
	// Check if user already has a version assigned (sticky sessions)
	if splitKey, exists := headers[config.SplitKey]; exists {
		if splitKey == config.VersionA {
			return config.VersionA
		}
		if splitKey == config.VersionB {
			return config.VersionB
		}
	}

	// Random assignment based on weights
	if tm.randomInt(100) < config.WeightA {
		return config.VersionA
	}
	return config.VersionB
}

// selectCanaryVersion selects version for canary deployment
func (tm *TrafficManager) selectCanaryVersion(config *CanaryConfig) string {
	// Simple random selection based on current weight
	if tm.randomInt(100) < config.InitialWeight {
		return config.NewVersion
	}
	return config.StableVersion
}

// selectSplitVersion selects version based on traffic splits
func (tm *TrafficManager) selectSplitVersion(splits []TrafficSplit, headers map[string]string) string {
	// Check header-based routing first
	for _, split := range splits {
		if len(split.Headers) > 0 {
			match := true
			for key, value := range split.Headers {
				if headers[key] != value {
					match = false
					break
				}
			}
			if match {
				return split.Version
			}
		}
	}

	// Weight-based selection
	random := tm.randomInt(100)
	cumulative := 0
	for _, split := range splits {
		cumulative += split.Weight
		if random < cumulative {
			return split.Version
		}
	}

	// Fallback to first version
	if len(splits) > 0 {
		return splits[0].Version
	}
	return ""
}

// IncrementCanary increments canary weight (for progressive rollout)
func (tm *TrafficManager) IncrementCanary(serviceName string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	policy := tm.policies[serviceName]
	if policy == nil || policy.Canary == nil || !policy.Canary.Enabled {
		return fmt.Errorf("canary not configured for service: %s", serviceName)
	}

	canary := policy.Canary
	newWeight := canary.InitialWeight + canary.IncrementStep
	if newWeight > canary.MaxWeight {
		newWeight = canary.MaxWeight
	}

	canary.InitialWeight = newWeight
	return nil
}

// PromoteCanary promotes canary to stable (100% traffic)
func (tm *TrafficManager) PromoteCanary(serviceName string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	policy := tm.policies[serviceName]
	if policy == nil || policy.Canary == nil {
		return fmt.Errorf("canary not configured for service: %s", serviceName)
	}

	// Set new version as stable
	policy.Canary.StableVersion = policy.Canary.NewVersion
	policy.Canary.InitialWeight = 0
	policy.Canary.Enabled = false

	return nil
}

// RollbackCanary rolls back canary to stable
func (tm *TrafficManager) RollbackCanary(serviceName string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	policy := tm.policies[serviceName]
	if policy == nil || policy.Canary == nil {
		return fmt.Errorf("canary not configured for service: %s", serviceName)
	}

	// Reset to stable
	policy.Canary.InitialWeight = 0
	policy.Canary.Enabled = false

	return nil
}

// ListPolicies lists all traffic policies
func (tm *TrafficManager) ListPolicies() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	services := make([]string, 0, len(tm.policies))
	for name := range tm.policies {
		services = append(services, name)
	}
	return services
}

// randomInt returns random int between 0 and max (exclusive)
func (tm *TrafficManager) randomInt(max int) int {
	// Simple pseudo-random for demonstration
	// In production, use crypto/rand or math/rand with seed
	return int(tm.mu.RLock() + tm.mu.RUnlock()) % max
}

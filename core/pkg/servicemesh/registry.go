package servicemesh

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// ServiceRegistry manages service discovery
type ServiceRegistry struct {
	controlPlane string
	services     map[string][]*ServiceInstance
	mu           sync.RWMutex
	lastSync     time.Time
}

// ServiceInstance represents a service instance
type ServiceInstance struct {
	ServiceName string            `json:"service_name"`
	InstanceID  string            `json:"instance_id"`
	Host        string            `json:"host"`
	Port        int               `json:"port"`
	Protocol    string            `json:"protocol"` // http, https, grpc
	Metadata    map[string]string `json:"metadata"`
	Health      HealthStatus      `json:"health"`
	RegisteredAt time.Time        `json:"registered_at"`
	LastHeartbeat time.Time       `json:"last_heartbeat"`
}

// HealthStatus health check status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// NewServiceRegistry creates a new service registry
func NewServiceRegistry(controlPlane string) *ServiceRegistry {
	registry := &ServiceRegistry{
		controlPlane: controlPlane,
		services:     make(map[string][]*ServiceInstance),
	}

	// Start background sync
	go registry.syncLoop()

	return registry
}

// Register registers a service instance
func (r *ServiceRegistry) Register(instance *ServiceInstance) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if instance.InstanceID == "" {
		instance.InstanceID = fmt.Sprintf("%s-%d", instance.ServiceName, time.Now().UnixNano())
	}

	instance.RegisteredAt = time.Now()
	instance.LastHeartbeat = time.Now()
	instance.Health = HealthStatusHealthy

	// Add to local cache
	r.services[instance.ServiceName] = append(r.services[instance.ServiceName], instance)

	// Register with control plane if configured
	if r.controlPlane != "" {
		return r.registerWithControlPlane(instance)
	}

	return nil
}

// Deregister removes a service instance
func (r *ServiceRegistry) Deregister(serviceName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.services, serviceName)

	if r.controlPlane != "" {
		return r.deregisterFromControlPlane(serviceName)
	}

	return nil
}

// Discover discovers a service instance
func (r *ServiceRegistry) Discover(serviceName string) (*ServiceInstance, error) {
	r.mu.RLock()
	instances := r.services[serviceName]
	r.mu.RUnlock()

	if len(instances) == 0 {
		// Try to fetch from control plane
		if r.controlPlane != "" {
			if err := r.syncFromControlPlane(serviceName); err == nil {
				r.mu.RLock()
				instances = r.services[serviceName]
				r.mu.RUnlock()
			}
		}
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("no instances found for service: %s", serviceName)
	}

	// Filter healthy instances
	healthy := make([]*ServiceInstance, 0)
	for _, inst := range instances {
		if inst.Health == HealthStatusHealthy {
			healthy = append(healthy, inst)
		}
	}

	if len(healthy) == 0 {
		return nil, fmt.Errorf("no healthy instances for service: %s", serviceName)
	}

	// Simple round-robin (can be enhanced with load balancing)
	return healthy[time.Now().UnixNano()%int64(len(healthy))], nil
}

// DiscoverAll discovers all instances of a service
func (r *ServiceRegistry) DiscoverAll(serviceName string) ([]*ServiceInstance, error) {
	r.mu.RLock()
	instances := r.services[serviceName]
	r.mu.RUnlock()

	if len(instances) == 0 {
		return nil, fmt.Errorf("no instances found for service: %s", serviceName)
	}

	return instances, nil
}

// Heartbeat sends heartbeat for a service
func (r *ServiceRegistry) Heartbeat(serviceName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	instances := r.services[serviceName]
	for _, inst := range instances {
		inst.LastHeartbeat = time.Now()
	}

	if r.controlPlane != "" {
		return r.heartbeatToControlPlane(serviceName)
	}

	return nil
}

// UpdateHealth updates health status of an instance
func (r *ServiceRegistry) UpdateHealth(serviceName, instanceID string, status HealthStatus) {
	r.mu.Lock()
	defer r.mu.Unlock()

	instances := r.services[serviceName]
	for _, inst := range instances {
		if inst.InstanceID == instanceID {
			inst.Health = status
			break
		}
	}
}

// ListServices lists all registered services
func (r *ServiceRegistry) ListServices() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]string, 0, len(r.services))
	for name := range r.services {
		services = append(services, name)
	}
	return services
}

// GetServiceInstances gets all instances for a service
func (r *ServiceRegistry) GetServiceInstances(serviceName string) []*ServiceInstance {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.services[serviceName]
}

// syncLoop periodically syncs with control plane
func (r *ServiceRegistry) syncLoop() {
	if r.controlPlane == "" {
		return
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		r.syncAllFromControlPlane()
	}
}

// registerWithControlPlane registers with control plane
func (r *ServiceRegistry) registerWithControlPlane(instance *ServiceInstance) error {
	body, err := json.Marshal(instance)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/api/v1/services/register", r.controlPlane),
		"application/json",
		nil,
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registration failed: %d", resp.StatusCode)
	}

	return nil
}

// deregisterFromControlPlane deregisters from control plane
func (r *ServiceRegistry) deregisterFromControlPlane(serviceName string) error {
	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/api/v1/services/%s", r.controlPlane, serviceName),
		nil,
	)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// syncFromControlPlane syncs a specific service from control plane
func (r *ServiceRegistry) syncFromControlPlane(serviceName string) error {
	resp, err := http.Get(
		fmt.Sprintf("%s/api/v1/services/%s", r.controlPlane, serviceName),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sync failed: %d", resp.StatusCode)
	}

	var instances []*ServiceInstance
	if err := json.NewDecoder(resp.Body).Decode(&instances); err != nil {
		return err
	}

	r.mu.Lock()
	r.services[serviceName] = instances
	r.mu.Unlock()

	return nil
}

// syncAllFromControlPlane syncs all services from control plane
func (r *ServiceRegistry) syncAllFromControlPlane() error {
	resp, err := http.Get(
		fmt.Sprintf("%s/api/v1/services", r.controlPlane),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sync failed: %d", resp.StatusCode)
	}

	var allServices map[string][]*ServiceInstance
	if err := json.NewDecoder(resp.Body).Decode(&allServices); err != nil {
		return err
	}

	r.mu.Lock()
	r.services = allServices
	r.lastSync = time.Now()
	r.mu.Unlock()

	return nil
}

// heartbeatToControlPlane sends heartbeat to control plane
func (r *ServiceRegistry) heartbeatToControlPlane(serviceName string) error {
	resp, err := http.Post(
		fmt.Sprintf("%s/api/v1/services/%s/heartbeat", r.controlPlane, serviceName),
		"application/json",
		nil,
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// CleanupStaleInstances removes instances that haven't sent heartbeat
func (r *ServiceRegistry) CleanupStaleInstances(timeout time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for serviceName, instances := range r.services {
		healthy := make([]*ServiceInstance, 0)
		for _, inst := range instances {
			if now.Sub(inst.LastHeartbeat) < timeout {
				healthy = append(healthy, inst)
			}
		}
		r.services[serviceName] = healthy
	}
}

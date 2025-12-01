package events

import (
	"context"
	"fmt"
	"sync"
)

// Event represents an event with data
type Event struct {
	Name string
	Data interface{}
}

// Handler is a function that handles an event
type Handler func(ctx context.Context, event Event) error

// EventDispatcher manages events and handlers
type EventDispatcher struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

// NewEventDispatcher creates a new event dispatcher
func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		handlers: make(map[string][]Handler),
	}
}

// Register registers a handler for an event
func (d *EventDispatcher) Register(eventName string, handler Handler) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.handlers[eventName] = append(d.handlers[eventName], handler)
}

// Dispatch dispatches an event to all registered handlers
func (d *EventDispatcher) Dispatch(ctx context.Context, event Event) error {
	d.mu.RLock()
	handlers := d.handlers[event.Name]
	d.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			return fmt.Errorf("handler failed for event %s: %w", event.Name, err)
		}
	}

	return nil
}

// DispatchAsync dispatches event asynchronously
func (d *EventDispatcher) DispatchAsync(ctx context.Context, event Event) {
	go d.Dispatch(ctx, event)
}

// HasHandlers checks if event has handlers
func (d *EventDispatcher) HasHandlers(eventName string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.handlers[eventName]) > 0
}

// Common event names
const (
	// User events
	EventUserCreated       = "user.created"
	EventUserUpdated       = "user.updated"
	EventUserDeleted       = "user.deleted"
	EventUserLoggedIn      = "user.logged_in"
	EventUserLoggedOut     = "user.logged_out"
	EventUserPasswordReset = "user.password_reset"

	// Module events
	EventModuleInstalled   = "module.installed"
	EventModuleUninstalled = "module.uninstalled"
	EventModuleActivated   = "module.activated"
	EventModuleDeactivated = "module.deactivated"
	EventModuleUpdated     = "module.updated"

	// System events
	EventSystemStarted  = "system.started"
	EventSystemShutdown = "system.shutdown"
)

// Global dispatcher instance
var defaultDispatcher = NewEventDispatcher()

// Register registers a global event handler
func Register(eventName string, handler Handler) {
	defaultDispatcher.Register(eventName, handler)
}

// Dispatch dispatches a global event
func Dispatch(ctx context.Context, event Event) error {
	return defaultDispatcher.Dispatch(ctx, event)
}

// DispatchAsync dispatches a global event asynchronously
func DispatchAsync(ctx context.Context, event Event) {
	defaultDispatcher.DispatchAsync(ctx, event)
}

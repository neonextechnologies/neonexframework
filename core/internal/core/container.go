package core

import (
	"reflect"
	"sync"
)

type ProviderType int

const (
	Singleton ProviderType = iota
	Transient
)

type providerDef struct {
	Type     ProviderType
	Factory  interface{}
	Instance interface{}
}

type Container struct {
	providers map[reflect.Type]*providerDef
	mu        sync.Mutex
}

func NewContainer() *Container {
	return &Container{
		providers: make(map[reflect.Type]*providerDef),
	}
}

// Register provider
func (c *Container) Provide(factory interface{}, pType ProviderType) {
	c.mu.Lock()
	defer c.mu.Unlock()

	t := reflect.TypeOf(factory).Out(0)
	c.providers[t] = &providerDef{
		Type:    pType,
		Factory: factory,
	}
}

// Resolve instance by type
func Resolve[T any](c *Container) T {
	var zero T
	t := reflect.TypeOf((*T)(nil)).Elem()

	c.mu.Lock()
	provider, ok := c.providers[t]
	c.mu.Unlock()

	if !ok {
		return zero
	}

	switch provider.Type {
	case Singleton:
		c.mu.Lock()
		instance := provider.Instance
		c.mu.Unlock()

		if instance == nil {
			// Call factory WITHOUT holding the lock to avoid deadlock
			newInstance := reflect.ValueOf(provider.Factory).Call(nil)[0].Interface()

			c.mu.Lock()
			// Check again in case another goroutine created it
			if provider.Instance == nil {
				provider.Instance = newInstance
			}
			instance = provider.Instance
			c.mu.Unlock()
		}

		return instance.(T)

	case Transient:
		return reflect.ValueOf(provider.Factory).Call(nil)[0].Interface().(T)
	}

	return zero
}

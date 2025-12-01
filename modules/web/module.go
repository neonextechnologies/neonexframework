package web

import (
	"neonexframework/pkg/app"

	"github.com/gofiber/fiber/v2"
)

// WebModule provides full-stack web features
type WebModule struct {
	router *Router
}

// New creates a new web module
func New() *WebModule {
	return &WebModule{
		router: NewRouter(),
	}
}

// Name returns the module name
func (m *WebModule) Name() string {
	return "web"
}

// RegisterServices registers web services
func (m *WebModule) RegisterServices(c *app.Container) error {
	c.Provide(func() *Router { return m.router })
	return nil
}

// RegisterRoutes registers web routes
func (m *WebModule) RegisterRoutes(router fiber.Router) error {
	// Web routes will be registered dynamically
	return nil
}

// Boot initializes the web module
func (m *WebModule) Boot() error {
	return nil
}

// Router provides route management
type Router struct {
	routes map[string][]Route
}

// Route represents a web route
type Route struct {
	Method  string
	Path    string
	Handler fiber.Handler
	Name    string
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		routes: make(map[string][]Route),
	}
}

// Group creates a route group
func (r *Router) Group(prefix string) *RouteGroup {
	return &RouteGroup{
		prefix: prefix,
		router: r,
	}
}

// RouteGroup represents a group of routes
type RouteGroup struct {
	prefix string
	router *Router
}

// Get registers a GET route
func (g *RouteGroup) Get(path string, handler fiber.Handler) {
	g.router.addRoute("GET", g.prefix+path, handler, "")
}

// Post registers a POST route
func (g *RouteGroup) Post(path string, handler fiber.Handler) {
	g.router.addRoute("POST", g.prefix+path, handler, "")
}

// Put registers a PUT route
func (g *RouteGroup) Put(path string, handler fiber.Handler) {
	g.router.addRoute("PUT", g.prefix+path, handler, "")
}

// Delete registers a DELETE route
func (g *RouteGroup) Delete(path string, handler fiber.Handler) {
	g.router.addRoute("DELETE", g.prefix+path, handler, "")
}

func (r *Router) addRoute(method, path string, handler fiber.Handler, name string) {
	r.routes[method] = append(r.routes[method], Route{
		Method:  method,
		Path:    path,
		Handler: handler,
		Name:    name,
	})
}

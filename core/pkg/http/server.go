package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Context wraps http request and response
type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Params  map[string]string
	status  int
}

// NewContext creates a new context
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer:  w,
		Request: r,
		Params:  make(map[string]string),
		status:  200,
	}
}

// Status sets the HTTP status code
func (c *Context) Status(code int) *Context {
	c.status = code
	return c
}

// JSON sends a JSON response
func (c *Context) JSON(data interface{}) error {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(c.status)
	return json.NewEncoder(c.Writer).Encode(data)
}

// String sends a string response
func (c *Context) String(s string) error {
	c.Writer.Header().Set("Content-Type", "text/plain")
	c.Writer.WriteHeader(c.status)
	_, err := c.Writer.Write([]byte(s))
	return err
}

// Param gets a URL parameter
func (c *Context) Param(key string) string {
	return c.Params[key]
}

// Query gets a query parameter
func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

// Body reads the request body
func (c *Context) Body() ([]byte, error) {
	return io.ReadAll(c.Request.Body)
}

// BodyParser parses JSON body into a struct
func (c *Context) BodyParser(v interface{}) error {
	body, err := c.Body()
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// Context returns the request context
func (c *Context) Context() interface{} {
	return c.Request.Context()
}

// HandlerFunc defines the handler function
type HandlerFunc func(*Context) error

// Router handles routing
type Router struct {
	routes     map[string]map[string]HandlerFunc // method -> path -> handler
	middleware []HandlerFunc
	prefix     string
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		routes: make(map[string]map[string]HandlerFunc),
	}
}

// Use adds middleware
func (r *Router) Use(handler HandlerFunc) {
	r.middleware = append(r.middleware, handler)
}

// Group creates a route group
func (r *Router) Group(prefix string) *Router {
	return &Router{
		routes:     r.routes,
		middleware: r.middleware,
		prefix:     r.prefix + prefix,
	}
}

// addRoute adds a route
func (r *Router) addRoute(method, path string, handler HandlerFunc) {
	fullPath := r.prefix + path
	if r.routes[method] == nil {
		r.routes[method] = make(map[string]HandlerFunc)
	}
	r.routes[method][fullPath] = handler
}

// GET adds a GET route
func (r *Router) Get(path string, handler HandlerFunc) {
	r.addRoute("GET", path, handler)
}

// POST adds a POST route
func (r *Router) Post(path string, handler HandlerFunc) {
	r.addRoute("POST", path, handler)
}

// PUT adds a PUT route
func (r *Router) Put(path string, handler HandlerFunc) {
	r.addRoute("PUT", path, handler)
}

// DELETE adds a DELETE route
func (r *Router) Delete(path string, handler HandlerFunc) {
	r.addRoute("DELETE", path, handler)
}

// matchRoute finds a matching route and extracts params
func (r *Router) matchRoute(method, path string) (HandlerFunc, map[string]string) {
	routes := r.routes[method]
	if routes == nil {
		return nil, nil
	}

	// Exact match
	if handler, ok := routes[path]; ok {
		return handler, nil
	}

	// Pattern match with params
	for pattern, handler := range routes {
		params := matchPath(pattern, path)
		if params != nil {
			return handler, params
		}
	}

	return nil, nil
}

// matchPath checks if path matches pattern and extracts params
func matchPath(pattern, path string) map[string]string {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(patternParts) != len(pathParts) {
		return nil
	}

	params := make(map[string]string)
	for i, part := range patternParts {
		if strings.HasPrefix(part, ":") {
			params[part[1:]] = pathParts[i]
		} else if part != pathParts[i] {
			return nil
		}
	}

	return params
}

// ServeHTTP implements http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := NewContext(w, req)

	// Run middleware
	for _, mw := range r.middleware {
		if err := mw(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Find and execute handler
	handler, params := r.matchRoute(req.Method, req.URL.Path)
	if handler == nil {
		http.NotFound(w, req)
		return
	}

	ctx.Params = params
	if err := handler(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Server represents the HTTP server
type Server struct {
	router *Router
	addr   string
	server *http.Server
}

// NewServer creates a new HTTP server
func NewServer() *Server {
	return &Server{
		router: NewRouter(),
	}
}

// Use adds middleware
func (s *Server) Use(handler HandlerFunc) {
	s.router.Use(handler)
}

// Group creates a route group
func (s *Server) Group(prefix string) *Router {
	return s.router.Group(prefix)
}

// Get adds a GET route
func (s *Server) Get(path string, handler HandlerFunc) {
	s.router.Get(path, handler)
}

// Post adds a POST route
func (s *Server) Post(path string, handler HandlerFunc) {
	s.router.Post(path, handler)
}

// Put adds a PUT route
func (s *Server) Put(path string, handler HandlerFunc) {
	s.router.Put(path, handler)
}

// Delete adds a DELETE route
func (s *Server) Delete(path string, handler HandlerFunc) {
	s.router.Delete(path, handler)
}

// Listen starts the HTTP server
func (s *Server) Listen(addr string) error {
	s.addr = addr
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println()
	fmt.Println("┌───────────────────────────────────────────────────┐")
	fmt.Println("│              Neonex Core v0.1-alpha               │")
	fmt.Printf("│               http://127.0.0.1%s                │\n", strings.Replace(addr, ":", "", 1))
	fmt.Printf("│       (bound on host 0.0.0.0 and port %s)       │\n", strings.Replace(addr, ":", "", 1))
	fmt.Println("│                                                   │")
	fmt.Println("│ Framework .... Neonex  Engine ....... net/http   │")
	fmt.Println("└───────────────────────────────────────────────────┘")
	fmt.Println()

	return s.server.ListenAndServe()
}

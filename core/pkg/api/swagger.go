package api

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// SwaggerInfo holds API documentation metadata
type SwaggerInfo struct {
	Version     string            `json:"version"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Contact     *SwaggerContact   `json:"contact,omitempty"`
	License     *SwaggerLicense   `json:"license,omitempty"`
	Servers     []SwaggerServer   `json:"servers,omitempty"`
}

// SwaggerContact represents API contact information
type SwaggerContact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// SwaggerLicense represents API license information
type SwaggerLicense struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// SwaggerServer represents an API server
type SwaggerServer struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// SwaggerSpec represents the OpenAPI 3.0 specification
type SwaggerSpec struct {
	OpenAPI    string                 `json:"openapi"`
	Info       SwaggerInfo            `json:"info"`
	Servers    []SwaggerServer        `json:"servers,omitempty"`
	Paths      map[string]interface{} `json:"paths"`
	Components map[string]interface{} `json:"components,omitempty"`
	Tags       []SwaggerTag           `json:"tags,omitempty"`
}

// SwaggerTag represents a tag for grouping operations
type SwaggerTag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// SwaggerGenerator generates OpenAPI documentation
type SwaggerGenerator struct {
	spec *SwaggerSpec
}

// NewSwaggerGenerator creates a new Swagger generator
func NewSwaggerGenerator(info SwaggerInfo) *SwaggerGenerator {
	return &SwaggerGenerator{
		spec: &SwaggerSpec{
			OpenAPI:    "3.0.0",
			Info:       info,
			Paths:      make(map[string]interface{}),
			Components: make(map[string]interface{}),
		},
	}
}

// AddServer adds a server to the spec
func (sg *SwaggerGenerator) AddServer(url, description string) {
	sg.spec.Servers = append(sg.spec.Servers, SwaggerServer{
		URL:         url,
		Description: description,
	})
}

// AddTag adds a tag to the spec
func (sg *SwaggerGenerator) AddTag(name, description string) {
	sg.spec.Tags = append(sg.spec.Tags, SwaggerTag{
		Name:        name,
		Description: description,
	})
}

// AddPath adds a path to the spec
func (sg *SwaggerGenerator) AddPath(path string, methods map[string]interface{}) {
	sg.spec.Paths[path] = methods
}

// AddSchema adds a schema to components
func (sg *SwaggerGenerator) AddSchema(name string, schema interface{}) {
	if sg.spec.Components["schemas"] == nil {
		sg.spec.Components["schemas"] = make(map[string]interface{})
	}
	sg.spec.Components["schemas"].(map[string]interface{})[name] = schema
}

// AddSecurityScheme adds a security scheme to components
func (sg *SwaggerGenerator) AddSecurityScheme(name string, scheme interface{}) {
	if sg.spec.Components["securitySchemes"] == nil {
		sg.spec.Components["securitySchemes"] = make(map[string]interface{})
	}
	sg.spec.Components["securitySchemes"].(map[string]interface{})[name] = scheme
}

// GetSpec returns the complete OpenAPI spec
func (sg *SwaggerGenerator) GetSpec() *SwaggerSpec {
	return sg.spec
}

// GenerateJSON generates JSON representation of the spec
func (sg *SwaggerGenerator) GenerateJSON() ([]byte, error) {
	return json.MarshalIndent(sg.spec, "", "  ")
}

// SetupSwaggerRoutes sets up Swagger UI and JSON routes
func SetupSwaggerRoutes(app *fiber.App, sg *SwaggerGenerator) {
	// Serve OpenAPI JSON
	app.Get("/api/docs/openapi.json", func(c *fiber.Ctx) error {
		return c.JSON(sg.spec)
	})

	// Serve Swagger UI (HTML)
	app.Get("/api/docs", func(c *fiber.Ctx) error {
		html := generateSwaggerUIHTML("/api/docs/openapi.json")
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	})

	// ReDoc alternative
	app.Get("/api/docs/redoc", func(c *fiber.Ctx) error {
		html := generateReDocHTML("/api/docs/openapi.json")
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	})
}

// generateSwaggerUIHTML generates HTML for Swagger UI
func generateSwaggerUIHTML(specURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>API Documentation - Swagger UI</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
    <style>
        body { margin: 0; padding: 0; }
        .swagger-ui .topbar { display: none; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: '%s',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>`, specURL)
}

// generateReDocHTML generates HTML for ReDoc
func generateReDocHTML(specURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>API Documentation - ReDoc</title>
    <style>
        body { margin: 0; padding: 0; }
    </style>
</head>
<body>
    <redoc spec-url='%s'></redoc>
    <script src="https://cdn.jsdelivr.net/npm/redoc@latest/bundles/redoc.standalone.js"></script>
</body>
</html>`, specURL)
}

// CreateDefaultSwagger creates a default Swagger spec for NeonexCore
func CreateDefaultSwagger() *SwaggerGenerator {
	sg := NewSwaggerGenerator(SwaggerInfo{
		Version:     "1.0.0",
		Title:       "NeonexCore API",
		Description: "RESTful API documentation for NeonexCore application",
		Contact: &SwaggerContact{
			Name:  "NeonexCore Team",
			Email: "support@neonexcore.com",
		},
		License: &SwaggerLicense{
			Name: "MIT",
			URL:  "https://opensource.org/licenses/MIT",
		},
	})

	// Add default server
	sg.AddServer("http://localhost:3000", "Development Server")

	// Add default tags
	sg.AddTag("Authentication", "User authentication endpoints")
	sg.AddTag("Users", "User management endpoints")
	sg.AddTag("Modules", "Module management endpoints")

	// Add JWT security scheme
	sg.AddSecurityScheme("bearerAuth", map[string]interface{}{
		"type":         "http",
		"scheme":       "bearer",
		"bearerFormat": "JWT",
		"description":  "Enter your JWT token in the format: Bearer {token}",
	})

	// Add common schemas
	sg.AddSchema("Error", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type":    "boolean",
				"example": false,
			},
			"message": map[string]interface{}{
				"type":    "string",
				"example": "Error message",
			},
			"errors": map[string]interface{}{
				"type": "object",
			},
			"timestamp": map[string]interface{}{
				"type":    "integer",
				"example": 1234567890,
			},
		},
	})

	sg.AddSchema("Success", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type":    "boolean",
				"example": true,
			},
			"message": map[string]interface{}{
				"type":    "string",
				"example": "Operation successful",
			},
			"data": map[string]interface{}{
				"type": "object",
			},
			"timestamp": map[string]interface{}{
				"type":    "integer",
				"example": 1234567890,
			},
		},
	})

	sg.AddSchema("PaginatedResponse", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type":    "boolean",
				"example": true,
			},
			"data": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
				},
			},
			"meta": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page":           map[string]interface{}{"type": "integer", "example": 1},
					"limit":          map[string]interface{}{"type": "integer", "example": 10},
					"total":          map[string]interface{}{"type": "integer", "example": 100},
					"total_pages":    map[string]interface{}{"type": "integer", "example": 10},
					"has_next_page":  map[string]interface{}{"type": "boolean", "example": true},
					"has_prev_page":  map[string]interface{}{"type": "boolean", "example": false},
				},
			},
			"timestamp": map[string]interface{}{
				"type":    "integer",
				"example": 1234567890,
			},
		},
	})

	return sg
}

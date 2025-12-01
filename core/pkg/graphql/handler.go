package graphql

import (
	"github.com/gofiber/fiber/v2"
)

// Handler is the HTTP handler for GraphQL requests
type Handler struct {
	executor *Executor
	schema   *Schema
}

// HandlerConfig configures the GraphQL handler
type HandlerConfig struct {
	Schema    *Schema
	Executor  *Executor
	Playground bool // Enable GraphQL Playground
}

// NewHandler creates a new GraphQL HTTP handler
func NewHandler(config HandlerConfig) *Handler {
	executor := config.Executor
	if executor == nil {
		executor = NewExecutor(config.Schema)
	}

	return &Handler{
		executor: executor,
		schema:   config.Schema,
	}
}

// ServeHTTP handles GraphQL HTTP requests
func (h *Handler) ServeHTTP(c *fiber.Ctx) error {
	// Handle introspection query
	if c.Query("introspection") == "true" {
		return c.JSON(fiber.Map{
			"data": h.schema.String(),
		})
	}

	// Parse query
	var query Query
	if err := c.BodyParser(&query); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"errors": []Error{
				{Message: "Invalid request body"},
			},
		})
	}

	// Validate query
	if errors := h.executor.Validate(&query); len(errors) > 0 {
		return c.Status(400).JSON(fiber.Map{
			"errors": errors,
		})
	}

	// Execute query
	ctx := c.Context()
	response := h.executor.Execute(ctx, &query)

	// Return response
	if len(response.Errors) > 0 {
		return c.Status(400).JSON(response)
	}

	return c.JSON(response)
}

// PlaygroundHandler serves the GraphQL Playground
func (h *Handler) PlaygroundHandler(c *fiber.Ctx) error {
	html := `
<!DOCTYPE html>
<html>
<head>
  <meta charset=utf-8/>
  <meta name="viewport" content="user-scalable=no, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, minimal-ui">
  <title>GraphQL Playground</title>
  <link rel="stylesheet" href="//cdn.jsdelivr.net/npm/graphql-playground-react/build/static/css/index.css" />
  <link rel="shortcut icon" href="//cdn.jsdelivr.net/npm/graphql-playground-react/build/favicon.png" />
  <script src="//cdn.jsdelivr.net/npm/graphql-playground-react/build/static/js/middleware.js"></script>
</head>
<body>
  <div id="root">
    <style>
      body {
        background-color: rgb(23, 42, 58);
        font-family: Open Sans, sans-serif;
        height: 90vh;
      }
      #root {
        height: 100%;
        width: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
      }
      .loading {
        font-size: 32px;
        font-weight: 200;
        color: rgba(255, 255, 255, .6);
        margin-left: 20px;
      }
      img {
        width: 78px;
        height: 78px;
      }
      .title {
        font-weight: 400;
      }
    </style>
    <img src='//cdn.jsdelivr.net/npm/graphql-playground-react/build/logo.png' alt=''>
    <div class="loading"> Loading
      <span class="title">GraphQL Playground</span>
    </div>
  </div>
  <script>window.addEventListener('load', function (event) {
      GraphQLPlayground.init(document.getElementById('root'), {
        endpoint: '/graphql',
        settings: {
          'request.credentials': 'same-origin'
        }
      })
    })</script>
</body>
</html>
`
	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}

// SchemaHandler serves the GraphQL schema in SDL format
func (h *Handler) SchemaHandler(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/plain")
	return c.SendString(h.schema.String())
}

// SetupRoutes sets up GraphQL routes
func SetupRoutes(app fiber.Router, schema *Schema, executor *Executor, enablePlayground bool) *Handler {
	handler := NewHandler(HandlerConfig{
		Schema:     schema,
		Executor:   executor,
		Playground: enablePlayground,
	})

	// GraphQL endpoint
	app.Post("/graphql", handler.ServeHTTP)
	app.Get("/graphql", handler.ServeHTTP)

	// Schema endpoint
	app.Get("/graphql/schema", handler.SchemaHandler)

	// Playground (optional)
	if enablePlayground {
		app.Get("/graphql/playground", handler.PlaygroundHandler)
	}

	return handler
}

package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Query represents a GraphQL query
type Query struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
}

// Response represents a GraphQL response
type Response struct {
	Data   interface{}            `json:"data,omitempty"`
	Errors []Error                `json:"errors,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// Error represents a GraphQL error
type Error struct {
	Message    string                 `json:"message"`
	Locations  []Location             `json:"locations,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// Location represents an error location in the query
type Location struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// Executor executes GraphQL queries
type Executor struct {
	schema    *Schema
	resolvers map[string]ResolverFunc
}

// NewExecutor creates a new query executor
func NewExecutor(schema *Schema) *Executor {
	return &Executor{
		schema:    schema,
		resolvers: make(map[string]ResolverFunc),
	}
}

// RegisterResolver registers a resolver for a field
func (e *Executor) RegisterResolver(typeName, fieldName string, resolver ResolverFunc) {
	key := fmt.Sprintf("%s.%s", typeName, fieldName)
	e.resolvers[key] = resolver
}

// Execute executes a GraphQL query
func (e *Executor) Execute(ctx context.Context, query *Query) *Response {
	response := &Response{
		Extensions: make(map[string]interface{}),
	}

	// Parse query (simplified - in production use a proper parser)
	queryType := e.detectQueryType(query.Query)

	var rootType *ObjectType
	switch queryType {
	case "query":
		rootType = e.schema.QueryType
	case "mutation":
		rootType = e.schema.MutationType
	case "subscription":
		rootType = e.schema.SubscriptionType
	default:
		response.Errors = append(response.Errors, Error{
			Message: "Unknown operation type",
		})
		return response
	}

	if rootType == nil {
		response.Errors = append(response.Errors, Error{
			Message: fmt.Sprintf("Schema does not support %s", queryType),
		})
		return response
	}

	// Execute query (simplified)
	data, err := e.executeFields(ctx, rootType, nil, query.Variables)
	if err != nil {
		response.Errors = append(response.Errors, Error{
			Message: err.Error(),
		})
		return response
	}

	response.Data = data
	return response
}

// detectQueryType detects if it's a query, mutation, or subscription
func (e *Executor) detectQueryType(query string) string {
	query = strings.TrimSpace(query)

	if strings.HasPrefix(query, "mutation") {
		return "mutation"
	}
	if strings.HasPrefix(query, "subscription") {
		return "subscription"
	}
	return "query"
}

// executeFields executes fields on an object
func (e *Executor) executeFields(ctx context.Context, objType *ObjectType, parent interface{}, variables map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, field := range objType.Fields {
		// Get resolver
		resolverKey := fmt.Sprintf("%s.%s", objType.Name, field.Name)
		resolver := e.resolvers[resolverKey]

		if resolver == nil {
			// Use default resolver
			resolver = field.Resolver
		}

		if resolver == nil {
			// Skip fields without resolvers
			continue
		}

		// Execute resolver
		value, err := resolver(ctx, parent, variables)
		if err != nil {
			return nil, fmt.Errorf("error resolving field %s: %w", field.Name, err)
		}

		result[field.Name] = value
	}

	return result, nil
}

// Validate validates a query against the schema
func (e *Executor) Validate(query *Query) []Error {
	errors := []Error{}

	// Basic validation (in production, use a proper validator)
	if query.Query == "" {
		errors = append(errors, Error{
			Message: "Query cannot be empty",
		})
	}

	// Validate operation type
	queryType := e.detectQueryType(query.Query)
	switch queryType {
	case "query":
		if e.schema.QueryType == nil {
			errors = append(errors, Error{
				Message: "Schema does not support queries",
			})
		}
	case "mutation":
		if e.schema.MutationType == nil {
			errors = append(errors, Error{
				Message: "Schema does not support mutations",
			})
		}
	case "subscription":
		if e.schema.SubscriptionType == nil {
			errors = append(errors, Error{
				Message: "Schema does not support subscriptions",
			})
		}
	}

	return errors
}

// IntrospectionQuery returns the GraphQL introspection query
func IntrospectionQuery() string {
	return `
query IntrospectionQuery {
  __schema {
    queryType { name }
    mutationType { name }
    subscriptionType { name }
    types {
      ...FullType
    }
    directives {
      name
      description
      locations
      args {
        ...InputValue
      }
    }
  }
}

fragment FullType on __Type {
  kind
  name
  description
  fields(includeDeprecated: true) {
    name
    description
    args {
      ...InputValue
    }
    type {
      ...TypeRef
    }
    isDeprecated
    deprecationReason
  }
  inputFields {
    ...InputValue
  }
  interfaces {
    ...TypeRef
  }
  enumValues(includeDeprecated: true) {
    name
    description
    isDeprecated
    deprecationReason
  }
  possibleTypes {
    ...TypeRef
  }
}

fragment InputValue on __InputValue {
  name
  description
  type { ...TypeRef }
  defaultValue
}

fragment TypeRef on __Type {
  kind
  name
  ofType {
    kind
    name
    ofType {
      kind
      name
      ofType {
        kind
        name
        ofType {
          kind
          name
          ofType {
            kind
            name
            ofType {
              kind
              name
              ofType {
                kind
                name
              }
            }
          }
        }
      }
    }
  }
}
`
}

// ToJSON converts the response to JSON
func (r *Response) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON parses a query from JSON
func (q *Query) FromJSON(data []byte) error {
	return json.Unmarshal(data, q)
}

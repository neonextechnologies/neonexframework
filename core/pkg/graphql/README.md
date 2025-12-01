# GraphQL Package

Modern GraphQL API support for NeonexCore with schema generation, query execution, and GraphQL Playground.

## Features

- ✅ **Schema Builder** - Fluent API for building GraphQL schemas
- ✅ **Type Generation** - Auto-generate types from Go structs
- ✅ **Query Execution** - Execute GraphQL queries and mutations
- ✅ **Introspection** - Full GraphQL introspection support
- ✅ **GraphQL Playground** - Interactive GraphQL IDE
- ✅ **Type-Safe** - Compile-time type checking
- ✅ **Resolver Functions** - Custom field resolvers
- ✅ **Input Types** - Support for input objects
- ✅ **Enums** - Enum type support
- ✅ **Interfaces & Unions** - Advanced type composition
- ✅ **Directives** - Custom directive support

## Architecture

```
pkg/graphql/
├── schema.go    - Schema definition and SDL generation
├── executor.go  - Query execution engine
├── handler.go   - HTTP handler and routes
└── builder.go   - Fluent schema builder API
```

## Quick Start

### 1. Define Models

```go
type User struct {
    ID    uint   `json:"id"`
    Name  string `json:"name" graphql:"User's full name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}
```

### 2. Build Schema

```go
import "neonexcore/pkg/graphql"

builder := graphql.NewBuilder()

// Define Query type
builder.Query(
    graphql.F("user", graphql.TypeObject, 
        func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
            id := args["id"].(uint)
            return GetUserByID(id)
        },
        graphql.WithDescription("Get a user by ID"),
        graphql.WithElementType("User"),
        graphql.WithArgs(
            graphql.Arg("id", graphql.TypeID, graphql.ArgRequired()),
        ),
    ),
)

// Auto-generate User type from struct
builder.TypeFromStruct("User", User{}, "A user in the system")

schema := builder.Build()
```

### 3. Setup Routes

```go
import "neonexcore/pkg/graphql"

executor := graphql.NewExecutor(schema)
graphql.SetupRoutes(app, schema, executor, true) // true = enable playground
```

### 4. Query the API

```bash
# GraphQL endpoint
POST http://localhost:8080/graphql

# GraphQL Playground
http://localhost:8080/graphql/playground

# Schema SDL
http://localhost:8080/graphql/schema
```

## Schema Building

### Define Query Type

```go
builder.Query(
    graphql.F("users", graphql.TypeList,
        func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
            return GetAllUsers()
        },
        graphql.WithDescription("Get all users"),
        graphql.WithElementType("User"),
    ),
    graphql.F("user", graphql.TypeObject,
        userByIDResolver,
        graphql.WithElementType("User"),
        graphql.WithArgs(
            graphql.Arg("id", graphql.TypeID, graphql.ArgRequired()),
        ),
    ),
)
```

### Define Mutation Type

```go
builder.Mutation(
    graphql.F("createUser", graphql.TypeObject,
        func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
            input := args["input"].(map[string]interface{})
            return CreateUser(input)
        },
        graphql.WithDescription("Create a new user"),
        graphql.WithElementType("User"),
        graphql.WithArgs(
            graphql.Arg("input", graphql.TypeObject, 
                graphql.ArgRequired(), 
                graphql.ArgElementType("CreateUserInput"),
            ),
        ),
    ),
)
```

### Define Subscription Type

```go
builder.Subscription(
    graphql.F("userCreated", graphql.TypeObject,
        func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
            // Return channel for subscription
            return userCreatedChannel, nil
        },
        graphql.WithElementType("User"),
    ),
)
```

### Define Types from Structs

```go
// Auto-generate from struct
builder.TypeFromStruct("User", User{}, "A user in the system")

// Or manually define
builder.Type("User",
    graphql.F("id", graphql.TypeID, graphql.FieldResolver("ID")),
    graphql.F("name", graphql.TypeString, graphql.FieldResolver("Name")),
    graphql.F("email", graphql.TypeString, graphql.FieldResolver("Email")),
    graphql.F("posts", graphql.TypeList, postsByUserResolver,
        graphql.WithElementType("Post"),
    ),
)
```

### Define Input Types

```go
builder.Input("CreateUserInput",
    graphql.IF("name", graphql.TypeString, 
        graphql.IFRequired(),
        graphql.IFDescription("User's name"),
    ),
    graphql.IF("email", graphql.TypeString, graphql.IFRequired()),
    graphql.IF("age", graphql.TypeInt, 
        graphql.IFDefault(18),
        graphql.IFDescription("User's age"),
    ),
)
```

### Define Enums

```go
builder.Enum("UserRole",
    graphql.EV("ADMIN", graphql.EVDescription("Administrator role")),
    graphql.EV("USER", graphql.EVDescription("Regular user")),
    graphql.EV("GUEST", graphql.EVDescription("Guest user")),
)
```

### Define Interfaces

```go
builder.Interface("Node",
    graphql.F("id", graphql.TypeID, nil),
    graphql.F("createdAt", graphql.TypeString, nil),
)

// Type implementing interface
builder.Type("User",
    graphql.F("id", graphql.TypeID, ...),
    graphql.F("createdAt", graphql.TypeString, ...),
    graphql.F("name", graphql.TypeString, ...),
).WithInterfaces("Node")
```

### Define Unions

```go
builder.Union("SearchResult", "User", "Post", "Comment")
```

## Resolver Functions

### Simple Resolver

```go
func userResolver(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
    id := args["id"].(uint)
    user, err := db.GetUser(id)
    return user, err
}
```

### Static Resolver

```go
// Returns a static value
graphql.F("version", graphql.TypeString, graphql.StaticResolver("1.0.0"))
```

### Field Resolver

```go
// Returns a field from parent object
graphql.F("name", graphql.TypeString, graphql.FieldResolver("Name"))
```

### Args Resolver

```go
// Returns an argument value
graphql.F("echo", graphql.TypeString, graphql.ArgsResolver("message"))
```

### Context-Aware Resolver

```go
func authenticatedUserResolver(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
    userID := ctx.Value("userID").(uint)
    return GetUser(userID)
}
```

## Query Examples

### Basic Query

```graphql
query {
  user(id: 1) {
    id
    name
    email
  }
}
```

### Query with Variables

```graphql
query GetUser($id: ID!) {
  user(id: $id) {
    id
    name
    email
    posts {
      title
    }
  }
}
```

Variables:
```json
{
  "id": "1"
}
```

### Mutation

```graphql
mutation CreateUser($input: CreateUserInput!) {
  createUser(input: $input) {
    id
    name
    email
  }
}
```

Variables:
```json
{
  "input": {
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30
  }
}
```

### Fragments

```graphql
fragment UserFields on User {
  id
  name
  email
}

query {
  user(id: 1) {
    ...UserFields
  }
}
```

### Introspection

```graphql
query {
  __schema {
    types {
      name
      kind
    }
  }
}
```

## HTTP API

### POST /graphql

Execute GraphQL query:

```bash
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "{ user(id: 1) { name } }",
    "variables": {},
    "operationName": null
  }'
```

Response:
```json
{
  "data": {
    "user": {
      "name": "John Doe"
    }
  }
}
```

### GET /graphql/playground

Interactive GraphQL Playground IDE

### GET /graphql/schema

Download schema in SDL format:

```graphql
schema {
  query: Query
  mutation: Mutation
}

type User {
  id: ID!
  name: String!
  email: String!
  age: Int!
}

type Query {
  user(id: ID!): User
  users: [User!]!
}

type Mutation {
  createUser(input: CreateUserInput!): User
}

input CreateUserInput {
  name: String!
  email: String!
  age: Int!
}
```

## Integration with Modules

### User Module Example

```go
// modules/user/graphql.go
package user

func RegisterGraphQL(builder *graphql.Builder, service *UserService) {
    // Add user queries
    builder.Query(
        graphql.F("me", graphql.TypeObject,
            func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
                userID := ctx.Value("userID").(uint)
                return service.GetUser(userID)
            },
            graphql.WithElementType("User"),
        ),
    )

    // Add user mutations
    builder.Mutation(
        graphql.F("updateProfile", graphql.TypeObject,
            func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
                userID := ctx.Value("userID").(uint)
                input := args["input"].(map[string]interface{})
                return service.UpdateProfile(userID, input)
            },
            graphql.WithElementType("User"),
        ),
    )

    // Register User type
    builder.TypeFromStruct("User", model.User{})
}
```

## Advanced Features

### Pagination

```go
builder.Query(
    graphql.F("users", graphql.TypeObject,
        func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
            page := args["page"].(int)
            limit := args["limit"].(int)
            
            users, total := GetUsersPaginated(page, limit)
            
            return map[string]interface{}{
                "items": users,
                "total": total,
                "page": page,
                "hasNextPage": (page * limit) < total,
            }, nil
        },
        graphql.WithElementType("UserConnection"),
        graphql.WithArgs(
            graphql.Arg("page", graphql.TypeInt, graphql.ArgDefault(1)),
            graphql.Arg("limit", graphql.TypeInt, graphql.ArgDefault(10)),
        ),
    ),
)
```

### Error Handling

```go
func userResolver(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
    id := args["id"].(uint)
    user, err := GetUser(id)
    if err != nil {
        // Return GraphQL error
        return nil, fmt.Errorf("user not found: %w", err)
    }
    return user, nil
}
```

### Dataloader (N+1 Problem)

```go
// Use context to batch requests
func postsByUserResolver(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
    user := parent.(User)
    
    // Get or create batch loader from context
    loader := ctx.Value("postLoader").(*PostLoader)
    
    // Load posts (will be batched)
    return loader.Load(user.ID)
}
```

### Authentication

```go
func protectedResolver(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
    // Check authentication from context
    userID := ctx.Value("userID")
    if userID == nil {
        return nil, fmt.Errorf("unauthenticated")
    }
    
    // Continue with logic
    return GetUserData(userID.(uint))
}
```

## Performance

- **Type-Safe** - Compile-time checking prevents runtime errors
- **Lazy Loading** - Fields resolved on-demand
- **Concurrent** - Resolvers run concurrently when possible
- **Caching** - Schema compiled once at startup
- **Low Overhead** - Minimal reflection usage

## Best Practices

1. **Use Input Types** for mutations instead of many arguments
2. **Implement Pagination** for list fields that can return many items
3. **Add Descriptions** to all types and fields for better documentation
4. **Handle Errors** gracefully and return meaningful error messages
5. **Use Context** for authentication, request tracing, and dataloaders
6. **Avoid N+1 Queries** by implementing dataloader pattern
7. **Use Fragments** to reuse field selections
8. **Implement Introspection** for tooling and documentation
9. **Version with Deprecation** instead of breaking changes
10. **Test Resolvers** independently before integration

## Comparison with REST

| Feature | REST | GraphQL |
|---------|------|---------|
| **Over-fetching** | Common | Eliminated |
| **Under-fetching** | Common (N+1) | Eliminated |
| **Versioning** | URL-based | Deprecation |
| **Documentation** | Manual/Swagger | Auto-generated |
| **Flexibility** | Fixed endpoints | Client-controlled |
| **Caching** | HTTP native | Requires library |
| **File Upload** | Native | Requires multipart |
| **Real-time** | WebSocket/SSE | Subscriptions |

## Future Enhancements

- [ ] Full query parser (currently simplified)
- [ ] Dataloader implementation
- [ ] Subscription support over WebSocket
- [ ] File upload support
- [ ] Query complexity analysis
- [ ] Rate limiting per query cost
- [ ] Persisted queries
- [ ] Schema stitching
- [ ] Federation support
- [ ] APQ (Automatic Persisted Queries)

## Example Project

Run the example:

```bash
go run examples/graphql_example.go
```

This will:
1. Build a complete GraphQL schema
2. Print the SDL
3. Execute example queries
4. Show query responses

## License

MIT License - Part of NeonexCore Framework

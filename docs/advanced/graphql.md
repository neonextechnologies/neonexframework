# GraphQL API

Build powerful GraphQL APIs with NeonEx Framework. Learn schema definition, resolvers, queries, mutations, subscriptions, and real-time features.

## Table of Contents

- [Introduction](#introduction)
- [Quick Start](#quick-start)
- [Schema Definition](#schema-definition)
- [Resolvers](#resolvers)
- [Queries](#queries)
- [Mutations](#mutations)
- [Subscriptions](#subscriptions)
- [Authentication](#authentication)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)

## Introduction

NeonEx provides built-in GraphQL support with:

- **Schema Builder**: Type-safe schema definition
- **Resolvers**: Function-based field resolvers
- **Real-time**: WebSocket-based subscriptions
- **Batching**: DataLoader for N+1 prevention
- **Authentication**: JWT and API key support
- **Playground**: Interactive API explorer

## Quick Start

```go
package main

import (
    "context"
    "neonex/core/pkg/graphql"
    "github.com/labstack/echo/v4"
)

func main() {
    // Create schema
    schema := graphql.NewSchema()
    
    // Define User type
    userType := &graphql.ObjectType{
        Name: "User",
        Fields: []*graphql.Field{
            {
                Name: "id",
                Type: graphql.TypeID,
            },
            {
                Name: "email",
                Type: graphql.TypeString,
            },
            {
                Name: "name",
                Type: graphql.TypeString,
            },
        },
    }
    
    schema.AddType(userType)
    
    // Define Query
    queryType := &graphql.ObjectType{
        Name: "Query",
        Fields: []*graphql.Field{
            {
                Name: "user",
                Type: graphql.TypeObject,
                ElementType: "User",
                Args: []*graphql.Argument{
                    {
                        Name: "id",
                        Type: graphql.TypeID,
                        Required: true,
                    },
                },
                Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
                    userID := args["id"].(string)
                    return getUser(userID)
                },
            },
        },
    }
    
    schema.SetQuery(queryType)
    
    // Create HTTP handler
    handler := graphql.NewHandler(schema)
    
    e := echo.New()
    e.POST("/graphql", handler.ServeHTTP)
    e.GET("/playground", handler.Playground)
    e.Start(":8080")
}
```

## Schema Definition

### Basic Types

```go
// Define Product type
productType := &graphql.ObjectType{
    Name: "Product",
    Description: "A product in the catalog",
    Fields: []*graphql.Field{
        {
            Name: "id",
            Type: graphql.TypeID,
            Description: "Product ID",
        },
        {
            Name: "name",
            Type: graphql.TypeString,
            Description: "Product name",
        },
        {
            Name: "price",
            Type: graphql.TypeFloat,
            Description: "Product price in USD",
        },
        {
            Name: "inStock",
            Type: graphql.TypeBoolean,
            Description: "Whether product is in stock",
        },
    },
}

schema.AddType(productType)
```

### Lists and NonNull

```go
{
    Name: "products",
    Type: graphql.TypeList,
    ElementType: "Product",  // List of products
    Resolver: productsResolver,
}

{
    Name: "email",
    Type: graphql.TypeNonNull,
    ElementType: graphql.TypeString,  // Required field
}

{
    Name: "tags",
    Type: graphql.TypeList,
    ElementType: graphql.TypeString,  // List of strings
}
```

### Input Types

```go
createUserInput := &graphql.InputType{
    Name: "CreateUserInput",
    Description: "Input for creating a user",
    Fields: []*graphql.InputField{
        {
            Name: "email",
            Type: graphql.TypeString,
            Required: true,
        },
        {
            Name: "name",
            Type: graphql.TypeString,
            Required: true,
        },
        {
            Name: "password",
            Type: graphql.TypeString,
            Required: true,
        },
    },
}

schema.AddInput(createUserInput)
```

### Enums

```go
roleEnum := &graphql.EnumType{
    Name: "UserRole",
    Description: "User roles",
    Values: []*graphql.EnumValue{
        {
            Name: "ADMIN",
            Description: "Administrator",
        },
        {
            Name: "USER",
            Description: "Regular user",
        },
        {
            Name: "GUEST",
            Description: "Guest user",
        },
    },
}

schema.AddEnum(roleEnum)
```

### From Struct

```go
type User struct {
    ID        int       `json:"id" graphql:"User ID"`
    Email     string    `json:"email" graphql:"User email address"`
    Name      string    `json:"name" graphql:"User full name"`
    CreatedAt time.Time `json:"created_at" graphql:"Account creation time"`
}

// Auto-generate type from struct
userType := graphql.FromStruct("User", User{})
schema.AddType(userType)
```

## Resolvers

### Field Resolvers

```go
{
    Name: "user",
    Type: graphql.TypeObject,
    ElementType: "User",
    Args: []*graphql.Argument{
        {
            Name: "id",
            Type: graphql.TypeID,
            Required: true,
        },
    },
    Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
        userID := args["id"].(string)
        
        var user User
        if err := db.First(&user, userID).Error; err != nil {
            return nil, err
        }
        
        return user, nil
    },
}
```

### Nested Resolvers

```go
// User posts field
{
    Name: "posts",
    Type: graphql.TypeList,
    ElementType: "Post",
    Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
        user := parent.(*User)
        
        var posts []Post
        db.Where("user_id = ?", user.ID).Find(&posts)
        
        return posts, nil
    },
}
```

### DataLoader (N+1 Prevention)

```go
type UserLoader struct {
    db *gorm.DB
}

func NewUserLoader(db *gorm.DB) *UserLoader {
    return &UserLoader{db: db}
}

func (ul *UserLoader) Load(ctx context.Context, userIDs []int) ([]User, error) {
    var users []User
    err := ul.db.Where("id IN ?", userIDs).Find(&users).Error
    return users, err
}

// Use in resolver
func getUsersByIDs(ctx context.Context, userIDs []int) ([]User, error) {
    loader := ctx.Value("userLoader").(*UserLoader)
    return loader.Load(ctx, userIDs)
}
```

## Queries

### Simple Query

```go
queryType := &graphql.ObjectType{
    Name: "Query",
    Fields: []*graphql.Field{
        {
            Name: "hello",
            Type: graphql.TypeString,
            Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
                return "Hello, World!", nil
            },
        },
        {
            Name: "users",
            Type: graphql.TypeList,
            ElementType: "User",
            Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
                var users []User
                db.Find(&users)
                return users, nil
            },
        },
    },
}

schema.SetQuery(queryType)
```

### Query with Arguments

```go
{
    Name: "searchProducts",
    Type: graphql.TypeList,
    ElementType: "Product",
    Args: []*graphql.Argument{
        {
            Name: "query",
            Type: graphql.TypeString,
            Required: true,
        },
        {
            Name: "limit",
            Type: graphql.TypeInt,
            DefaultValue: 10,
        },
        {
            Name: "offset",
            Type: graphql.TypeInt,
            DefaultValue: 0,
        },
    },
    Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
        query := args["query"].(string)
        limit := args["limit"].(int)
        offset := args["offset"].(int)
        
        var products []Product
        db.Where("name LIKE ?", "%"+query+"%").
            Limit(limit).
            Offset(offset).
            Find(&products)
        
        return products, nil
    },
}
```

### Example Query

```graphql
query GetUser($id: ID!) {
  user(id: $id) {
    id
    email
    name
    posts {
      id
      title
      content
    }
  }
}

query SearchProducts($query: String!, $limit: Int) {
  searchProducts(query: $query, limit: $limit) {
    id
    name
    price
    inStock
  }
}
```

## Mutations

### Create Mutation

```go
mutationType := &graphql.ObjectType{
    Name: "Mutation",
    Fields: []*graphql.Field{
        {
            Name: "createUser",
            Type: graphql.TypeObject,
            ElementType: "User",
            Args: []*graphql.Argument{
                {
                    Name: "input",
                    Type: graphql.TypeObject,
                    ElementType: "CreateUserInput",
                    Required: true,
                },
            },
            Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
                input := args["input"].(map[string]interface{})
                
                user := &User{
                    Email: input["email"].(string),
                    Name:  input["name"].(string),
                }
                
                if err := db.Create(user).Error; err != nil {
                    return nil, err
                }
                
                return user, nil
            },
        },
    },
}

schema.SetMutation(mutationType)
```

### Update Mutation

```go
{
    Name: "updateUser",
    Type: graphql.TypeObject,
    ElementType: "User",
    Args: []*graphql.Argument{
        {
            Name: "id",
            Type: graphql.TypeID,
            Required: true,
        },
        {
            Name: "input",
            Type: graphql.TypeObject,
            ElementType: "UpdateUserInput",
            Required: true,
        },
    },
    Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
        userID := args["id"].(string)
        input := args["input"].(map[string]interface{})
        
        var user User
        if err := db.First(&user, userID).Error; err != nil {
            return nil, err
        }
        
        // Update fields
        if name, ok := input["name"].(string); ok {
            user.Name = name
        }
        
        if err := db.Save(&user).Error; err != nil {
            return nil, err
        }
        
        return user, nil
    },
}
```

### Delete Mutation

```go
{
    Name: "deleteUser",
    Type: graphql.TypeBoolean,
    Args: []*graphql.Argument{
        {
            Name: "id",
            Type: graphql.TypeID,
            Required: true,
        },
    },
    Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
        userID := args["id"].(string)
        
        result := db.Delete(&User{}, userID)
        if result.Error != nil {
            return false, result.Error
        }
        
        return result.RowsAffected > 0, nil
    },
}
```

### Example Mutation

```graphql
mutation CreateUser($input: CreateUserInput!) {
  createUser(input: $input) {
    id
    email
    name
  }
}

mutation UpdateUser($id: ID!, $input: UpdateUserInput!) {
  updateUser(id: $id, input: $input) {
    id
    name
    updatedAt
  }
}

mutation DeleteUser($id: ID!) {
  deleteUser(id: $id)
}
```

## Subscriptions

### WebSocket Subscriptions

```go
subscriptionType := &graphql.ObjectType{
    Name: "Subscription",
    Fields: []*graphql.Field{
        {
            Name: "messageAdded",
            Type: graphql.TypeObject,
            ElementType: "Message",
            Args: []*graphql.Argument{
                {
                    Name: "channelId",
                    Type: graphql.TypeID,
                    Required: true,
                },
            },
            Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
                channelID := args["channelId"].(string)
                
                // Create channel for real-time updates
                messages := make(chan Message)
                
                // Subscribe to message events
                subscribeToMessages(channelID, messages)
                
                return messages, nil
            },
        },
    },
}

schema.SetSubscription(subscriptionType)
```

### Example Subscription

```graphql
subscription OnMessageAdded($channelId: ID!) {
  messageAdded(channelId: $channelId) {
    id
    content
    author {
      name
    }
    createdAt
  }
}
```

## Authentication

### JWT Middleware

```go
func GraphQLAuthMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            token := c.Request().Header.Get("Authorization")
            
            if token == "" {
                return next(c)
            }
            
            // Verify JWT
            claims, err := verifyJWT(token)
            if err != nil {
                return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
            }
            
            // Add user to context
            ctx := context.WithValue(c.Request().Context(), "user", claims)
            c.SetRequest(c.Request().WithContext(ctx))
            
            return next(c)
        }
    }
}
```

### Protected Resolver

```go
func requireAuth(ctx context.Context) (*User, error) {
    user, ok := ctx.Value("user").(*User)
    if !ok {
        return nil, fmt.Errorf("authentication required")
    }
    return user, nil
}

// Protected resolver
{
    Name: "me",
    Type: graphql.TypeObject,
    ElementType: "User",
    Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
        user, err := requireAuth(ctx)
        if err != nil {
            return nil, err
        }
        
        return user, nil
    },
}
```

## Error Handling

### Custom Errors

```go
type GraphQLError struct {
    Message    string                 `json:"message"`
    Extensions map[string]interface{} `json:"extensions,omitempty"`
}

func NewGraphQLError(message string, code string) *GraphQLError {
    return &GraphQLError{
        Message: message,
        Extensions: map[string]interface{}{
            "code": code,
        },
    }
}

// In resolver
if err := validate(input); err != nil {
    return nil, NewGraphQLError("Validation failed", "VALIDATION_ERROR")
}
```

### Error Response

```json
{
  "errors": [
    {
      "message": "Validation failed",
      "extensions": {
        "code": "VALIDATION_ERROR"
      },
      "path": ["createUser"]
    }
  ]
}
```

## Best Practices

### 1. Schema Design

```graphql
# Use clear, descriptive names
type User {
  id: ID!
  email: String!
  name: String!
}

# Group related fields
type UserProfile {
  bio: String
  avatar: String
  location: String
}

# Use enums for fixed values
enum UserStatus {
  ACTIVE
  INACTIVE
  BANNED
}
```

### 2. Resolver Performance

```go
// Use DataLoader to prevent N+1 queries
type PostResolver struct {
    userLoader *UserLoader
}

func (r *PostResolver) Author(ctx context.Context, post *Post) (*User, error) {
    return r.userLoader.Load(ctx, post.UserID)
}
```

### 3. Pagination

```go
type PaginationArgs struct {
    First  *int
    After  *string
    Last   *int
    Before *string
}

{
    Name: "users",
    Type: graphql.TypeObject,
    ElementType: "UserConnection",
    Args: []*graphql.Argument{
        {Name: "first", Type: graphql.TypeInt},
        {Name: "after", Type: graphql.TypeString},
    },
    Resolver: paginatedUsersResolver,
}
```

### 4. Testing

```go
func TestGraphQLQuery(t *testing.T) {
    schema := setupTestSchema()
    
    query := `
        query {
            users {
                id
                email
            }
        }
    `
    
    result := executeQuery(schema, query, nil)
    
    assert.NoError(t, result.Errors)
    assert.NotNil(t, result.Data)
}
```

---

**Next Steps:**
- Learn about [gRPC](grpc.md) for service communication
- Explore [WebSocket](websocket.md) for real-time features
- See [Authentication](../security/authentication.md)

**Related Topics:**
- [GraphQL Spec](https://graphql.org/)
- [Apollo Client](https://www.apollographql.com/)
- [DataLoader](https://github.com/graphql/dataloader)

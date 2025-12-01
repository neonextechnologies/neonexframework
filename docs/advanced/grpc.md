# gRPC Services

Build high-performance microservices with gRPC in NeonEx Framework. Learn protocol buffers, service implementation, streaming, and inter-service communication.

## Table of Contents

- [Introduction](#introduction)
- [Quick Start](#quick-start)
- [Protocol Buffers](#protocol-buffers)
- [Service Implementation](#service-implementation)
- [Client Usage](#client-usage)
- [Streaming RPCs](#streaming-rpcs)
- [Interceptors](#interceptors)
- [Service Discovery](#service-discovery)
- [Best Practices](#best-practices)

## Introduction

NeonEx provides comprehensive gRPC support for building microservices:

- **Protocol Buffers**: Efficient serialization
- **HTTP/2**: Multiplexing and performance
- **Streaming**: Bidirectional communication
- **Interceptors**: Middleware for cross-cutting concerns
- **Service Discovery**: Dynamic service registry
- **Load Balancing**: Client-side load distribution

## Quick Start

### Install Protocol Buffer Compiler

```powershell
# Install protoc
choco install protoc

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Define Service

```protobuf
// proto/user.proto
syntax = "proto3";

package user;

option go_package = "myapp/proto/user";

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
}

message GetUserRequest {
  int64 id = 1;
}

message GetUserResponse {
  User user = 1;
}

message User {
  int64 id = 1;
  string email = 2;
  string name = 3;
  string created_at = 4;
}
```

### Generate Code

```powershell
protoc --go_out=. --go_opt=paths=source_relative `
  --go-grpc_out=. --go-grpc_opt=paths=source_relative `
  proto/user.proto
```

### Implement Service

```go
package main

import (
    "context"
    "log"
    "net"
    
    "google.golang.org/grpc"
    pb "myapp/proto/user"
)

type userServer struct {
    pb.UnimplementedUserServiceServer
    db *gorm.DB
}

func (s *userServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    var user User
    if err := s.db.First(&user, req.Id).Error; err != nil {
        return nil, err
    }
    
    return &pb.GetUserResponse{
        User: &pb.User{
            Id:        user.ID,
            Email:     user.Email,
            Name:      user.Name,
            CreatedAt: user.CreatedAt.Format(time.RFC3339),
        },
    }, nil
}

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    
    s := grpc.NewServer()
    pb.RegisterUserServiceServer(s, &userServer{db: db})
    
    log.Println("gRPC server listening on :50051")
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
```

## Protocol Buffers

### Message Types

```protobuf
syntax = "proto3";

package product;

// Simple message
message Product {
  int64 id = 1;
  string name = 2;
  double price = 3;
  bool in_stock = 4;
}

// Nested message
message Order {
  int64 id = 1;
  User user = 2;
  repeated OrderItem items = 3;
  OrderStatus status = 4;
}

message OrderItem {
  int64 product_id = 1;
  int32 quantity = 2;
  double price = 3;
}

// Enum
enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  PENDING = 1;
  CONFIRMED = 2;
  SHIPPED = 3;
  DELIVERED = 4;
  CANCELLED = 5;
}

// Optional fields
message UserProfile {
  string bio = 1;
  optional string avatar_url = 2;
  optional string location = 3;
}

// Map fields
message UserPreferences {
  map<string, string> settings = 1;
  map<string, bool> features = 2;
}

// Oneof (union)
message Payment {
  oneof payment_method {
    CreditCard credit_card = 1;
    PayPal paypal = 2;
    BankTransfer bank_transfer = 3;
  }
}
```

### Field Rules

```protobuf
// Required (none in proto3, use validation)
message CreateUserRequest {
  string email = 1;      // implicitly required
  string password = 2;
}

// Optional
message User {
  optional string middle_name = 1;
}

// Repeated (arrays)
message ListUsersResponse {
  repeated User users = 1;
}

// Reserved fields
message User {
  reserved 2, 15, 9 to 11;
  reserved "old_field", "deprecated_field";
  
  int64 id = 1;
  string email = 3;
}
```

## Service Implementation

### CRUD Operations

```go
type productServer struct {
    pb.UnimplementedProductServiceServer
    db *gorm.DB
}

func (s *productServer) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
    product := &Product{
        Name:  req.Name,
        Price: req.Price,
    }
    
    if err := s.db.Create(product).Error; err != nil {
        return nil, status.Errorf(codes.Internal, "failed to create product: %v", err)
    }
    
    return &pb.CreateProductResponse{
        Product: toProtoProduct(product),
    }, nil
}

func (s *productServer) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
    var product Product
    if err := s.db.First(&product, req.Id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, status.Errorf(codes.NotFound, "product not found")
        }
        return nil, status.Errorf(codes.Internal, "database error: %v", err)
    }
    
    return &pb.GetProductResponse{
        Product: toProtoProduct(&product),
    }, nil
}

func (s *productServer) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.UpdateProductResponse, error) {
    var product Product
    if err := s.db.First(&product, req.Id).Error; err != nil {
        return nil, status.Errorf(codes.NotFound, "product not found")
    }
    
    product.Name = req.Name
    product.Price = req.Price
    
    if err := s.db.Save(&product).Error; err != nil {
        return nil, status.Errorf(codes.Internal, "failed to update product: %v", err)
    }
    
    return &pb.UpdateProductResponse{
        Product: toProtoProduct(&product),
    }, nil
}

func (s *productServer) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
    result := s.db.Delete(&Product{}, req.Id)
    if result.Error != nil {
        return nil, status.Errorf(codes.Internal, "failed to delete product: %v", result.Error)
    }
    
    return &pb.DeleteProductResponse{
        Success: result.RowsAffected > 0,
    }, nil
}

func toProtoProduct(product *Product) *pb.Product {
    return &pb.Product{
        Id:       product.ID,
        Name:     product.Name,
        Price:    product.Price,
        InStock:  product.InStock,
    }
}
```

### Error Handling

```go
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

func (s *server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    if req.Id <= 0 {
        return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %d", req.Id)
    }
    
    var user User
    if err := s.db.First(&user, req.Id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, status.Errorf(codes.NotFound, "user not found: %d", req.Id)
        }
        return nil, status.Errorf(codes.Internal, "database error: %v", err)
    }
    
    return &pb.GetUserResponse{User: toProtoUser(&user)}, nil
}
```

### Context and Metadata

```go
import (
    "google.golang.org/grpc/metadata"
)

func (s *server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    // Read metadata
    md, ok := metadata.FromIncomingContext(ctx)
    if ok {
        requestID := md.Get("x-request-id")
        log.Printf("Request ID: %v", requestID)
    }
    
    // Set response metadata
    header := metadata.Pairs("x-response-time", "123ms")
    grpc.SendHeader(ctx, header)
    
    // Check context cancellation
    select {
    case <-ctx.Done():
        return nil, status.Errorf(codes.Canceled, "request canceled")
    default:
    }
    
    // ... implementation
}
```

## Client Usage

### Basic Client

```go
package main

import (
    "context"
    "log"
    "time"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    pb "myapp/proto/user"
)

func main() {
    // Connect to server
    conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("failed to connect: %v", err)
    }
    defer conn.Close()
    
    client := pb.NewUserServiceClient(conn)
    
    // Call GetUser
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    
    resp, err := client.GetUser(ctx, &pb.GetUserRequest{Id: 1})
    if err != nil {
        log.Fatalf("GetUser failed: %v", err)
    }
    
    log.Printf("User: %+v", resp.User)
}
```

### Client with TLS

```go
import (
    "google.golang.org/grpc/credentials"
)

func NewSecureClient(address string) (pb.UserServiceClient, error) {
    creds, err := credentials.NewClientTLSFromFile("cert.pem", "")
    if err != nil {
        return nil, err
    }
    
    conn, err := grpc.Dial(address, grpc.WithTransportCredentials(creds))
    if err != nil {
        return nil, err
    }
    
    return pb.NewUserServiceClient(conn), nil
}
```

### Client with Metadata

```go
func CallWithMetadata() {
    md := metadata.Pairs(
        "authorization", "Bearer "+token,
        "x-request-id", uuid.New().String(),
    )
    
    ctx := metadata.NewOutgoingContext(context.Background(), md)
    
    resp, err := client.GetUser(ctx, &pb.GetUserRequest{Id: 1})
    // ...
}
```

## Streaming RPCs

### Server Streaming

```protobuf
service ProductService {
  rpc ListProducts(ListProductsRequest) returns (stream Product);
}
```

```go
func (s *productServer) ListProducts(req *pb.ListProductsRequest, stream pb.ProductService_ListProductsServer) error {
    var products []Product
    s.db.Find(&products)
    
    for _, product := range products {
        if err := stream.Send(toProtoProduct(&product)); err != nil {
            return err
        }
        time.Sleep(100 * time.Millisecond) // Simulate delay
    }
    
    return nil
}

// Client
stream, err := client.ListProducts(ctx, &pb.ListProductsRequest{})
if err != nil {
    log.Fatal(err)
}

for {
    product, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Product: %+v", product)
}
```

### Client Streaming

```protobuf
service OrderService {
  rpc CreateBulkOrders(stream CreateOrderRequest) returns (CreateBulkOrdersResponse);
}
```

```go
func (s *orderServer) CreateBulkOrders(stream pb.OrderService_CreateBulkOrdersServer) error {
    var count int32
    
    for {
        req, err := stream.Recv()
        if err == io.EOF {
            return stream.SendAndClose(&pb.CreateBulkOrdersResponse{
                Count: count,
            })
        }
        if err != nil {
            return err
        }
        
        // Create order
        order := &Order{UserID: req.UserId}
        s.db.Create(order)
        count++
    }
}

// Client
stream, err := client.CreateBulkOrders(ctx)
if err != nil {
    log.Fatal(err)
}

for _, order := range orders {
    if err := stream.Send(order); err != nil {
        log.Fatal(err)
    }
}

resp, err := stream.CloseAndRecv()
if err != nil {
    log.Fatal(err)
}
log.Printf("Created %d orders", resp.Count)
```

### Bidirectional Streaming

```protobuf
service ChatService {
  rpc Chat(stream ChatMessage) returns (stream ChatMessage);
}
```

```go
func (s *chatServer) Chat(stream pb.ChatService_ChatServer) error {
    for {
        msg, err := stream.Recv()
        if err == io.EOF {
            return nil
        }
        if err != nil {
            return err
        }
        
        // Process and broadcast
        response := &pb.ChatMessage{
            Username: "Bot",
            Text:     "Echo: " + msg.Text,
        }
        
        if err := stream.Send(response); err != nil {
            return err
        }
    }
}

// Client
stream, err := client.Chat(ctx)
if err != nil {
    log.Fatal(err)
}

// Send messages
go func() {
    for _, msg := range messages {
        if err := stream.Send(msg); err != nil {
            log.Fatal(err)
        }
    }
    stream.CloseSend()
}()

// Receive messages
for {
    msg, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Received: %s", msg.Text)
}
```

## Interceptors

### Server Interceptor

```go
func LoggingInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        start := time.Now()
        
        // Log request
        log.Printf("Method: %s, Request: %+v", info.FullMethod, req)
        
        // Call handler
        resp, err := handler(ctx, req)
        
        // Log response
        duration := time.Since(start)
        log.Printf("Method: %s, Duration: %v, Error: %v", info.FullMethod, duration, err)
        
        return resp, err
    }
}

// Use interceptor
s := grpc.NewServer(
    grpc.UnaryInterceptor(LoggingInterceptor()),
)
```

### Authentication Interceptor

```go
func AuthInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        md, ok := metadata.FromIncomingContext(ctx)
        if !ok {
            return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
        }
        
        tokens := md.Get("authorization")
        if len(tokens) == 0 {
            return nil, status.Errorf(codes.Unauthenticated, "missing token")
        }
        
        token := strings.TrimPrefix(tokens[0], "Bearer ")
        claims, err := verifyJWT(token)
        if err != nil {
            return nil, status.Errorf(codes.Unauthenticated, "invalid token")
        }
        
        // Add user to context
        ctx = context.WithValue(ctx, "user", claims)
        
        return handler(ctx, req)
    }
}
```

### Client Interceptor

```go
func ClientLoggingInterceptor() grpc.UnaryClientInterceptor {
    return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
        log.Printf("Calling: %s", method)
        
        err := invoker(ctx, method, req, reply, cc, opts...)
        
        if err != nil {
            log.Printf("Error: %v", err)
        }
        
        return err
    }
}

conn, err := grpc.Dial(
    address,
    grpc.WithUnaryInterceptor(ClientLoggingInterceptor()),
)
```

## Service Discovery

### Service Registry

```go
type ServiceRegistry struct {
    services map[string][]string
    mu       sync.RWMutex
}

func NewServiceRegistry() *ServiceRegistry {
    return &ServiceRegistry{
        services: make(map[string][]string),
    }
}

func (sr *ServiceRegistry) Register(serviceName, address string) {
    sr.mu.Lock()
    defer sr.mu.Unlock()
    
    sr.services[serviceName] = append(sr.services[serviceName], address)
}

func (sr *ServiceRegistry) Discover(serviceName string) []string {
    sr.mu.RLock()
    defer sr.mu.RUnlock()
    
    return sr.services[serviceName]
}
```

### Load Balancing

```go
type RoundRobinBalancer struct {
    addresses []string
    current   int
    mu        sync.Mutex
}

func (rb *RoundRobinBalancer) Next() string {
    rb.mu.Lock()
    defer rb.mu.Unlock()
    
    address := rb.addresses[rb.current]
    rb.current = (rb.current + 1) % len(rb.addresses)
    
    return address
}

// Usage
balancer := &RoundRobinBalancer{
    addresses: registry.Discover("user-service"),
}

conn, err := grpc.Dial(balancer.Next(), opts...)
```

## Best Practices

### 1. Use Dependency Injection

```go
type UserService struct {
    pb.UnimplementedUserServiceServer
    repo UserRepository
    logger Logger
    cache Cache
}

func NewUserService(repo UserRepository, logger Logger, cache Cache) *UserService {
    return &UserService{
        repo:   repo,
        logger: logger,
        cache:  cache,
    }
}
```

### 2. Implement Health Checks

```go
import (
    "google.golang.org/grpc/health"
    "google.golang.org/grpc/health/grpc_health_v1"
)

healthServer := health.NewServer()
grpc_health_v1.RegisterHealthServer(s, healthServer)
healthServer.SetServingStatus("user.UserService", grpc_health_v1.HealthCheckResponse_SERVING)
```

### 3. Use Context Properly

```go
func (s *server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    // Pass context to database calls
    var user User
    if err := s.db.WithContext(ctx).First(&user, req.Id).Error; err != nil {
        return nil, err
    }
    
    return &pb.GetUserResponse{User: toProtoUser(&user)}, nil
}
```

### 4. Testing

```go
func TestUserService(t *testing.T) {
    // Create test server
    s := grpc.NewServer()
    pb.RegisterUserServiceServer(s, &userServer{db: testDB})
    
    lis := bufconn.Listen(bufSize)
    go s.Serve(lis)
    defer s.Stop()
    
    // Create client
    conn, _ := grpc.DialContext(ctx, "", grpc.WithContextDialer(bufDialer(lis)), grpc.WithInsecure())
    defer conn.Close()
    
    client := pb.NewUserServiceClient(conn)
    
    // Test
    resp, err := client.GetUser(ctx, &pb.GetUserRequest{Id: 1})
    assert.NoError(t, err)
    assert.NotNil(t, resp.User)
}
```

---

**Next Steps:**
- Learn about [Service Mesh](../deployment/service-mesh.md)
- Explore [GraphQL](graphql.md) for flexible APIs
- See [Metrics](metrics.md) for monitoring

**Related Topics:**
- [gRPC Documentation](https://grpc.io/docs/)
- [Protocol Buffers](https://protobuf.dev/)
- [gRPC Go](https://github.com/grpc/grpc-go)

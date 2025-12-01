# gRPC Package

High-performance gRPC server and client with service discovery, load balancing, and circuit breaker for NeonexCore microservices.

## Features

- ✅ **gRPC Server** - High-performance RPC server with reflection
- ✅ **gRPC Client** - Connection pooling with automatic retry
- ✅ **Service Discovery** - Service registration and discovery
- ✅ **Load Balancing** - Round-robin, random, least-connection
- ✅ **Circuit Breaker** - Fault tolerance and failure recovery
- ✅ **Interceptors** - Logging, metrics, auth, rate limiting
- ✅ **Health Checks** - Automatic service health monitoring
- ✅ **Compression** - GZIP compression support
- ✅ **Reflection** - Service reflection for debugging
- ✅ **Metrics** - Request tracking and performance monitoring

## Quick Start

### Server

```go
import "neonexcore/pkg/grpc"

// Create server
config := grpc.DefaultServerConfig()
config.Address = ":50051"
config.EnableReflection = true

server := grpc.NewServer(config)

// Register service
// server.RegisterService(&YourServiceDesc, &YourServiceImpl{})

// Start
go server.Start()
```

### Client

```go
// Create client
config := grpc.DefaultClientConfig("localhost:50051")
client, err := grpc.NewClient(config)
defer client.Close()

// Make RPC call
err = client.Invoke(ctx, "/service/Method", req, &resp)
```

## Service Discovery

```go
// Create registry
registry := grpc.NewServiceRegistry()

// Register service
registry.Register(grpc.ServiceInfo{
    Name:    "user-service",
    Address: "localhost:50051",
    Version: "1.0.0",
    Health:  grpc.HealthHealthy,
})

// Get service
service, _ := registry.Get("user-service")

// List all services
services := registry.List()
```

## Load Balancing

```go
// Create load balancer
lb := grpc.NewLoadBalancer(registry, grpc.StrategyRoundRobin)

// Get service instance
service, _ := lb.GetService("user-service")

// Connect to service
client, _ := grpc.NewClient(grpc.DefaultClientConfig(service.Address))
```

## Circuit Breaker

```go
// Create circuit breaker
cb := grpc.NewCircuitBreaker(5, 10*time.Second) // 5 failures, 10s timeout

// Protected call
err := cb.Call(func() error {
    return client.Invoke(ctx, method, req, &resp)
})
```

## Interceptors

### Server Interceptors

```go
config := grpc.DefaultServerConfig()

// Logging
config.UnaryInterceptors = append(config.UnaryInterceptors, 
    grpc.LoggingUnaryInterceptor())

// Metrics
config.UnaryInterceptors = append(config.UnaryInterceptors,
    grpc.MetricsUnaryInterceptor(metrics))

// Recovery
config.UnaryInterceptors = append(config.UnaryInterceptors,
    grpc.RecoveryUnaryInterceptor())

// Timeout
config.UnaryInterceptors = append(config.UnaryInterceptors,
    grpc.TimeoutUnaryInterceptor(5*time.Second))

server := grpc.NewServer(config)
```

### Client Interceptors

```go
// Before/After call hooks
interceptor := grpc.ClientInterceptor(
    func(ctx context.Context, method string) {
        fmt.Printf("Calling %s\n", method)
    },
    func(ctx context.Context, method string, err error) {
        if err != nil {
            fmt.Printf("Error: %v\n", err)
        }
    },
)
```

## License

MIT License - Part of NeonexCore Framework

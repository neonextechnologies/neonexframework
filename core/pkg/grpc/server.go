package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// Server represents a gRPC server
type Server struct {
	server  *grpc.Server
	config  ServerConfig
	metrics *ServerMetrics
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Address           string
	MaxConnections    int
	Timeout           time.Duration
	EnableReflection  bool
	EnableCompression bool
	EnableMetrics     bool
	UnaryInterceptors []grpc.UnaryServerInterceptor
	StreamInterceptors []grpc.StreamServerInterceptor
}

// ServerMetrics tracks server metrics
type ServerMetrics struct {
	RequestCount  uint64
	ErrorCount    uint64
	ActiveStreams uint64
	TotalDuration time.Duration
}

// DefaultServerConfig returns default server configuration
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Address:            ":50051",
		MaxConnections:     1000,
		Timeout:            30 * time.Second,
		EnableReflection:   true,
		EnableCompression:  true,
		EnableMetrics:      true,
		UnaryInterceptors:  make([]grpc.UnaryServerInterceptor, 0),
		StreamInterceptors: make([]grpc.StreamServerInterceptor, 0),
	}
}

// NewServer creates a new gRPC server
func NewServer(config ServerConfig) *Server {
	// Server options
	opts := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(uint32(config.MaxConnections)),
	}

	// Add unary interceptors
	if len(config.UnaryInterceptors) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(config.UnaryInterceptors...))
	}

	// Add stream interceptors
	if len(config.StreamInterceptors) > 0 {
		opts = append(opts, grpc.ChainStreamInterceptor(config.StreamInterceptors...))
	}

	// Enable compression
	if config.EnableCompression {
		opts = append(opts, grpc.RPCCompressor(grpc.NewGZIPCompressor()))
	}

	s := &Server{
		server:  grpc.NewServer(opts...),
		config:  config,
		metrics: &ServerMetrics{},
	}

	// Enable reflection
	if config.EnableReflection {
		reflection.Register(s.server)
	}

	return s
}

// Start starts the gRPC server
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	fmt.Printf("🚀 gRPC server listening on %s\n", s.config.Address)

	return s.server.Serve(lis)
}

// Stop gracefully stops the server
func (s *Server) Stop() {
	s.server.GracefulStop()
}

// GetServer returns the underlying gRPC server
func (s *Server) GetServer() *grpc.Server {
	return s.server
}

// GetMetrics returns server metrics
func (s *Server) GetMetrics() ServerMetrics {
	return *s.metrics
}

// RegisterService registers a service implementation
func (s *Server) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.server.RegisterService(desc, impl)
}

// UnaryServerInterceptor creates a unary server interceptor
func UnaryServerInterceptor(beforeCall func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo), afterCall func(ctx context.Context, resp interface{}, info *grpc.UnaryServerInfo, err error)) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if beforeCall != nil {
			beforeCall(ctx, req, info)
		}

		resp, err := handler(ctx, req)

		if afterCall != nil {
			afterCall(ctx, resp, info, err)
		}

		return resp, err
	}
}

// StreamServerInterceptor creates a streaming server interceptor
func StreamServerInterceptor(beforeStream func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo), afterStream func(srv interface{}, info *grpc.StreamServerInfo, err error)) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if beforeStream != nil {
			beforeStream(srv, ss, info)
		}

		err := handler(srv, ss)

		if afterStream != nil {
			afterStream(srv, info, err)
		}

		return err
	}
}

// LoggingUnaryInterceptor logs all unary RPC calls
func LoggingUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		code := codes.OK
		if err != nil {
			st, _ := status.FromError(err)
			code = st.Code()
		}

		fmt.Printf("[gRPC] %s | %s | %v | %s\n",
			info.FullMethod,
			code.String(),
			duration,
			getClientIP(ctx),
		)

		return resp, err
	}
}

// MetricsUnaryInterceptor tracks metrics for unary RPC calls
func MetricsUnaryInterceptor(metrics *ServerMetrics) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		metrics.RequestCount++

		resp, err := handler(ctx, req)

		metrics.TotalDuration += time.Since(start)

		if err != nil {
			metrics.ErrorCount++
		}

		return resp, err
	}
}

// RecoveryUnaryInterceptor recovers from panics in unary RPC calls
func RecoveryUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("[gRPC] Panic recovered: %v\n", r)
				err = status.Errorf(codes.Internal, "Internal server error: %v", r)
			}
		}()

		return handler(ctx, req)
	}
}

// TimeoutUnaryInterceptor adds timeout to unary RPC calls
func TimeoutUnaryInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		return handler(ctx, req)
	}
}

// AuthUnaryInterceptor validates authentication for unary RPC calls
func AuthUnaryInterceptor(validator func(ctx context.Context) error) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := validator(ctx); err != nil {
			return nil, status.Error(codes.Unauthenticated, "Authentication failed")
		}

		return handler(ctx, req)
	}
}

// RateLimitUnaryInterceptor implements rate limiting for unary RPC calls
func RateLimitUnaryInterceptor(limiter func(ctx context.Context) bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !limiter(ctx) {
			return nil, status.Error(codes.ResourceExhausted, "Rate limit exceeded")
		}

		return handler(ctx, req)
	}
}

// getClientIP extracts client IP from context
func getClientIP(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "unknown"
	}

	if xff := md.Get("x-forwarded-for"); len(xff) > 0 {
		return xff[0]
	}

	if xri := md.Get("x-real-ip"); len(xri) > 0 {
		return xri[0]
	}

	return "unknown"
}

// SetMetadata adds metadata to outgoing context
func SetMetadata(ctx context.Context, key, value string) context.Context {
	md := metadata.Pairs(key, value)
	return metadata.NewOutgoingContext(ctx, md)
}

// GetMetadata retrieves metadata from incoming context
func GetMetadata(ctx context.Context, key string) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}

	values := md.Get(key)
	if len(values) == 0 {
		return "", false
	}

	return values[0], true
}

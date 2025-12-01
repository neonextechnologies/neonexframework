package grpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Client represents a gRPC client with connection pooling and retry logic
type Client struct {
	conn    *grpc.ClientConn
	config  ClientConfig
	metrics *ClientMetrics
}

// ClientConfig holds client configuration
type ClientConfig struct {
	Address           string
	Timeout           time.Duration
	MaxRetries        int
	RetryDelay        time.Duration
	EnableCompression bool
	EnableMetrics     bool
	Metadata          map[string]string
}

// ClientMetrics tracks client metrics
type ClientMetrics struct {
	RequestCount  uint64
	ErrorCount    uint64
	RetryCount    uint64
	TotalDuration time.Duration
}

// DefaultClientConfig returns default client configuration
func DefaultClientConfig(address string) ClientConfig {
	return ClientConfig{
		Address:           address,
		Timeout:           10 * time.Second,
		MaxRetries:        3,
		RetryDelay:        100 * time.Millisecond,
		EnableCompression: true,
		EnableMetrics:     true,
		Metadata:          make(map[string]string),
	}
}

// NewClient creates a new gRPC client
func NewClient(config ClientConfig) (*Client, error) {
	// Dial options
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Enable compression
	if config.EnableCompression {
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.UseCompressor("gzip")))
	}

	// Connect
	conn, err := grpc.Dial(config.Address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &Client{
		conn:    conn,
		config:  config,
		metrics: &ClientMetrics{},
	}, nil
}

// Invoke makes a unary RPC call with automatic retry
func (c *Client) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	// Add timeout
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	// Add metadata
	if len(c.config.Metadata) > 0 {
		md := metadata.New(c.config.Metadata)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	start := time.Now()
	var lastErr error

	// Retry loop
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			time.Sleep(c.config.RetryDelay * time.Duration(attempt))
			c.metrics.RetryCount++
		}

		// Make the call
		err := c.conn.Invoke(ctx, method, args, reply, opts...)

		// Update metrics
		c.metrics.RequestCount++
		c.metrics.TotalDuration += time.Since(start)

		if err == nil {
			return nil
		}

		lastErr = err

		// Check if we should retry
		if !isRetryableError(err) {
			break
		}
	}

	c.metrics.ErrorCount++
	return lastErr
}

// NewStream creates a new streaming RPC
func (c *Client) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	// Add timeout
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	// Add metadata
	if len(c.config.Metadata) > 0 {
		md := metadata.New(c.config.Metadata)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	return c.conn.NewStream(ctx, desc, method, opts...)
}

// Close closes the client connection
func (c *Client) Close() error {
	return c.conn.Close()
}

// GetMetrics returns client metrics
func (c *Client) GetMetrics() ClientMetrics {
	return *c.metrics
}

// GetConnection returns the underlying gRPC connection
func (c *Client) GetConnection() *grpc.ClientConn {
	return c.conn
}

// SetMetadata sets metadata for all requests
func (c *Client) SetMetadata(key, value string) {
	c.config.Metadata[key] = value
}

// isRetryableError checks if an error is retryable
func isRetryableError(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	switch st.Code() {
	case codes.Unavailable, codes.ResourceExhausted, codes.DeadlineExceeded:
		return true
	default:
		return false
	}
}

// ClientInterceptor creates a unary client interceptor
func ClientInterceptor(beforeCall func(ctx context.Context, method string), afterCall func(ctx context.Context, method string, err error)) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if beforeCall != nil {
			beforeCall(ctx, method)
		}

		err := invoker(ctx, method, req, reply, cc, opts...)

		if afterCall != nil {
			afterCall(ctx, method, err)
		}

		return err
	}
}

// StreamClientInterceptor creates a streaming client interceptor
func StreamClientInterceptor(beforeStream func(ctx context.Context, method string), afterStream func(ctx context.Context, method string, err error)) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if beforeStream != nil {
			beforeStream(ctx, method)
		}

		stream, err := streamer(ctx, desc, cc, method, opts...)

		if afterStream != nil {
			afterStream(ctx, method, err)
		}

		return stream, err
	}
}

// ClientPool manages a pool of gRPC clients
type ClientPool struct {
	clients []*Client
	config  ClientConfig
	current int
}

// NewClientPool creates a new client pool
func NewClientPool(config ClientConfig, size int) (*ClientPool, error) {
	pool := &ClientPool{
		clients: make([]*Client, 0, size),
		config:  config,
	}

	// Create clients
	for i := 0; i < size; i++ {
		client, err := NewClient(config)
		if err != nil {
			// Close all created clients
			for _, c := range pool.clients {
				c.Close()
			}
			return nil, err
		}
		pool.clients = append(pool.clients, client)
	}

	return pool, nil
}

// Get returns the next available client (round-robin)
func (p *ClientPool) Get() *Client {
	client := p.clients[p.current]
	p.current = (p.current + 1) % len(p.clients)
	return client
}

// Close closes all clients in the pool
func (p *ClientPool) Close() error {
	for _, client := range p.clients {
		if err := client.Close(); err != nil {
			return err
		}
	}
	return nil
}

// GetMetrics returns aggregated metrics from all clients
func (p *ClientPool) GetMetrics() ClientMetrics {
	total := ClientMetrics{}
	for _, client := range p.clients {
		m := client.GetMetrics()
		total.RequestCount += m.RequestCount
		total.ErrorCount += m.ErrorCount
		total.RetryCount += m.RetryCount
		total.TotalDuration += m.TotalDuration
	}
	return total
}

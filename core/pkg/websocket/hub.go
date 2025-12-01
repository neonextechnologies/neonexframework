package websocket

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrConnectionClosed = errors.New("websocket connection closed")
	ErrSendBufferFull   = errors.New("send buffer full")
	ErrRoomNotFound     = errors.New("room not found")
	ErrConnectionExists = errors.New("connection already exists")
)

// Hub manages WebSocket connections and rooms
type Hub struct {
	connections map[string]*Connection        // Connection ID -> Connection
	userConns   map[uint]map[string]*Connection // User ID -> Connection IDs
	rooms       map[string]*Room               // Room name -> Room
	mu          sync.RWMutex
	
	// Configuration
	pingInterval    time.Duration
	pongTimeout     time.Duration
	writeTimeout    time.Duration
	maxMessageSize  int64
	
	// Cleanup
	cleanupInterval time.Duration
	cleanupTicker   *time.Ticker
	done            chan struct{}
}

// HubConfig configures the Hub
type HubConfig struct {
	PingInterval    time.Duration
	PongTimeout     time.Duration
	WriteTimeout    time.Duration
	MaxMessageSize  int64
	CleanupInterval time.Duration
}

// DefaultHubConfig returns default Hub configuration
func DefaultHubConfig() HubConfig {
	return HubConfig{
		PingInterval:    54 * time.Second,
		PongTimeout:     60 * time.Second,
		WriteTimeout:    10 * time.Second,
		MaxMessageSize:  512 * 1024, // 512 KB
		CleanupInterval: 30 * time.Second,
	}
}

// NewHub creates a new WebSocket hub
func NewHub(config HubConfig) *Hub {
	h := &Hub{
		connections:     make(map[string]*Connection),
		userConns:       make(map[uint]map[string]*Connection),
		rooms:           make(map[string]*Room),
		pingInterval:    config.PingInterval,
		pongTimeout:     config.PongTimeout,
		writeTimeout:    config.WriteTimeout,
		maxMessageSize:  config.MaxMessageSize,
		cleanupInterval: config.CleanupInterval,
		done:            make(chan struct{}),
	}
	
	// Start cleanup goroutine
	h.startCleanup()
	
	return h
}

// Register adds a new connection to the hub
func (h *Hub) Register(conn *Connection) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if _, exists := h.connections[conn.ID]; exists {
		return ErrConnectionExists
	}
	
	h.connections[conn.ID] = conn
	
	// Add to user connections
	if h.userConns[conn.UserID] == nil {
		h.userConns[conn.UserID] = make(map[string]*Connection)
	}
	h.userConns[conn.UserID][conn.ID] = conn
	
	return nil
}

// Unregister removes a connection from the hub
func (h *Hub) Unregister(connID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	conn, exists := h.connections[connID]
	if !exists {
		return
	}
	
	// Remove from connections
	delete(h.connections, connID)
	
	// Remove from user connections
	if userConns, ok := h.userConns[conn.UserID]; ok {
		delete(userConns, connID)
		if len(userConns) == 0 {
			delete(h.userConns, conn.UserID)
		}
	}
	
	// Remove from all rooms
	for _, room := range h.rooms {
		room.Leave(connID)
	}
	
	// Close connection
	conn.Close()
}

// GetConnection retrieves a connection by ID
func (h *Hub) GetConnection(connID string) (*Connection, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conn, ok := h.connections[connID]
	return conn, ok
}

// GetUserConnections retrieves all connections for a user
func (h *Hub) GetUserConnections(userID uint) []*Connection {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	userConns, ok := h.userConns[userID]
	if !ok {
		return []*Connection{}
	}
	
	conns := make([]*Connection, 0, len(userConns))
	for _, conn := range userConns {
		conns = append(conns, conn)
	}
	return conns
}

// Broadcast sends a message to all connections
func (h *Hub) Broadcast(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	for _, conn := range h.connections {
		conn.Send(message)
	}
}

// BroadcastJSON sends a JSON message to all connections
func (h *Hub) BroadcastJSON(v interface{}) error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	for _, conn := range h.connections {
		conn.SendJSON(v)
	}
	return nil
}

// SendToUser sends a message to all connections of a specific user
func (h *Hub) SendToUser(userID uint, message []byte) {
	conns := h.GetUserConnections(userID)
	for _, conn := range conns {
		conn.Send(message)
	}
}

// SendToUserJSON sends a JSON message to all connections of a specific user
func (h *Hub) SendToUserJSON(userID uint, v interface{}) error {
	conns := h.GetUserConnections(userID)
	for _, conn := range conns {
		conn.SendJSON(v)
	}
	return nil
}

// ConnectionCount returns the total number of active connections
func (h *Hub) ConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections)
}

// UserCount returns the total number of unique users connected
func (h *Hub) UserCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.userConns)
}

// Close shuts down the hub and closes all connections
func (h *Hub) Close() {
	close(h.done)
	
	if h.cleanupTicker != nil {
		h.cleanupTicker.Stop()
	}
	
	h.mu.Lock()
	defer h.mu.Unlock()
	
	// Close all connections
	for _, conn := range h.connections {
		conn.Close()
	}
	
	// Clear maps
	h.connections = make(map[string]*Connection)
	h.userConns = make(map[uint]map[string]*Connection)
	h.rooms = make(map[string]*Room)
}

// startCleanup starts the cleanup goroutine to remove dead connections
func (h *Hub) startCleanup() {
	h.cleanupTicker = time.NewTicker(h.cleanupInterval)
	
	go func() {
		for {
			select {
			case <-h.cleanupTicker.C:
				h.cleanup()
			case <-h.done:
				return
			}
		}
	}()
}

// cleanup removes dead connections
func (h *Hub) cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	deadConnections := []string{}
	
	for id, conn := range h.connections {
		if !conn.IsAlive(h.pongTimeout) {
			deadConnections = append(deadConnections, id)
		}
	}
	
	// Remove dead connections (unlock first to avoid deadlock)
	h.mu.Unlock()
	for _, id := range deadConnections {
		h.Unregister(id)
	}
	h.mu.Lock()
}

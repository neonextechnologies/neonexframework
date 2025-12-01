package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
)

// ConnectionStatus represents the state of a WebSocket connection
type ConnectionStatus string

const (
	StatusConnecting ConnectionStatus = "connecting"
	StatusConnected  ConnectionStatus = "connected"
	StatusClosing    ConnectionStatus = "closing"
	StatusClosed     ConnectionStatus = "closed"
)

// Connection represents a WebSocket connection with metadata
type Connection struct {
	ID        string
	UserID    uint
	Conn      *websocket.Conn
	Status    ConnectionStatus
	Context   context.Context
	Cancel    context.CancelFunc
	Metadata  map[string]interface{}
	CreatedAt time.Time
	LastPing  time.Time
	mu        sync.RWMutex
	sendCh    chan []byte
	done      chan struct{}
}

// NewConnection creates a new WebSocket connection wrapper
func NewConnection(id string, userID uint, conn *websocket.Conn) *Connection {
	ctx, cancel := context.WithCancel(context.Background())
	
	c := &Connection{
		ID:        id,
		UserID:    userID,
		Conn:      conn,
		Status:    StatusConnected,
		Context:   ctx,
		Cancel:    cancel,
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
		LastPing:  time.Now(),
		sendCh:    make(chan []byte, 256),
		done:      make(chan struct{}),
	}
	
	// Start send pump
	go c.writePump()
	
	return c
}

// Send sends a message to the connection
func (c *Connection) Send(message []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.Status != StatusConnected {
		return ErrConnectionClosed
	}
	
	select {
	case c.sendCh <- message:
		return nil
	case <-c.done:
		return ErrConnectionClosed
	default:
		return ErrSendBufferFull
	}
}

// SendJSON sends a JSON message to the connection
func (c *Connection) SendJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.Send(data)
}

// Close closes the connection gracefully
func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.Status == StatusClosed || c.Status == StatusClosing {
		return nil
	}
	
	c.Status = StatusClosing
	c.Cancel()
	close(c.done)
	
	// Send close message
	err := c.Conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		time.Now().Add(time.Second),
	)
	
	c.Status = StatusClosed
	return err
}

// UpdatePing updates the last ping timestamp
func (c *Connection) UpdatePing() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastPing = time.Now()
}

// IsAlive checks if connection is still alive
func (c *Connection) IsAlive(timeout time.Duration) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Since(c.LastPing) < timeout
}

// SetMetadata sets connection metadata
func (c *Connection) SetMetadata(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Metadata[key] = value
}

// GetMetadata gets connection metadata
func (c *Connection) GetMetadata(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.Metadata[key]
	return val, ok
}

// writePump pumps messages from the send channel to the WebSocket connection
func (c *Connection) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.sendCh:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
			
		case <-ticker.C:
			// Send ping
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
			
		case <-c.done:
			return
		}
	}
}

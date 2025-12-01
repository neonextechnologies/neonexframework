# WebSocket Support

NeonEx Framework provides full-featured WebSocket support for real-time communication, including connection management, rooms, broadcasting, and message handling.

## Table of Contents

- [Overview](#overview)
- [Hub Management](#hub-management)
- [Connection Handling](#connection-handling)
- [Rooms](#rooms)
- [Broadcasting](#broadcasting)
- [Message Types](#message-types)
- [Best Practices](#best-practices)

## Overview

The WebSocket system includes:
- Centralized hub for connection management
- Room-based messaging
- Ping/pong heartbeat
- Automatic cleanup of dead connections
- JSON message support

## Hub Management

### Creating a Hub

```go
import "neonexcore/pkg/websocket"

// Create hub with default config
hub := websocket.NewHub(websocket.DefaultHubConfig())

// Custom configuration
config := websocket.HubConfig{
    PingInterval:    54 * time.Second,
    PongTimeout:     60 * time.Second,
    WriteTimeout:    10 * time.Second,
    MaxMessageSize:  512 * 1024, // 512 KB
    CleanupInterval: 30 * time.Second,
}

hub := websocket.NewHub(config)
```

### Hub Configuration

```go
type HubConfig struct {
    PingInterval    time.Duration // How often to send ping
    PongTimeout     time.Duration // Timeout for pong response
    WriteTimeout    time.Duration // Write operation timeout
    MaxMessageSize  int64         // Maximum message size
    CleanupInterval time.Duration // Dead connection cleanup interval
}
```

## Connection Handling

### WebSocket Endpoint

```go
import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/websocket/v2"
)

func setupWebSocket(app *fiber.App, hub *websocket.Hub) {
    // WebSocket upgrade middleware
    app.Use("/ws", func(c *fiber.Ctx) error {
        if websocket.IsWebSocketUpgrade(c) {
            return c.Next()
        }
        return fiber.ErrUpgradeRequired
    })
    
    // WebSocket handler
    app.Get("/ws", websocket.New(func(c *websocket.Conn) {
        // Get user ID from query or JWT
        userID := getUserIDFromToken(c)
        
        // Create connection
        conn := websocket.NewConnection(c, userID)
        
        // Register with hub
        if err := hub.Register(conn); err != nil {
            log.Printf("Failed to register connection: %v", err)
            return
        }
        
        defer hub.Unregister(conn.ID)
        
        // Handle messages
        handleWebSocketMessages(hub, conn)
    }))
}
```

### Message Handler

```go
func handleWebSocketMessages(hub *websocket.Hub, conn *websocket.Connection) {
    for {
        // Read message
        messageType, message, err := conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("WebSocket error: %v", err)
            }
            break
        }
        
        // Handle different message types
        switch messageType {
        case websocket.TextMessage:
            handleTextMessage(hub, conn, message)
        case websocket.BinaryMessage:
            handleBinaryMessage(hub, conn, message)
        }
    }
}

func handleTextMessage(hub *websocket.Hub, conn *websocket.Connection, data []byte) {
    var msg websocket.Message
    if err := json.Unmarshal(data, &msg); err != nil {
        conn.SendJSON(websocket.ErrorMessage("Invalid message format"))
        return
    }
    
    switch msg.Type {
    case "join_room":
        handleJoinRoom(hub, conn, msg)
    case "leave_room":
        handleLeaveRoom(hub, conn, msg)
    case "message":
        handleChatMessage(hub, conn, msg)
    case "typing":
        handleTypingIndicator(hub, conn, msg)
    default:
        conn.SendJSON(websocket.ErrorMessage("Unknown message type"))
    }
}
```

### Connection Management

```go
// Get connection by ID
conn, exists := hub.GetConnection(connectionID)
if exists {
    conn.SendJSON(data)
}

// Get all connections for a user
connections := hub.GetUserConnections(userID)
for _, conn := range connections {
    conn.SendJSON(message)
}

// Get connection count
totalConnections := hub.ConnectionCount()
uniqueUsers := hub.UserCount()

// Unregister connection
hub.Unregister(connectionID)

// Close hub and all connections
hub.Close()
```

## Rooms

### Creating and Managing Rooms

```go
// Create room
room := hub.CreateRoom("chat-room-1")

// Join room
room.Join(connection)

// Leave room
room.Leave(connectionID)

// Check if user in room
isMember := room.HasMember(connectionID)

// Get room members
members := room.GetMembers()

// Remove room
hub.RemoveRoom("chat-room-1")
```

### Room Broadcasting

```go
// Broadcast to all in room
room.Broadcast([]byte("Hello everyone!"))

// Broadcast JSON to room
room.BroadcastJSON(fiber.Map{
    "type":    "notification",
    "message": "New user joined",
})

// Broadcast except sender
room.BroadcastExcept(connectionID, message)
```

### Room Implementation Example

```go
type ChatRoom struct {
    ID      string
    Name    string
    Members map[string]*websocket.Connection
    mu      sync.RWMutex
}

func (r *ChatRoom) Join(conn *websocket.Connection) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    r.Members[conn.ID] = conn
    
    // Notify others
    r.BroadcastExcept(conn.ID, fiber.Map{
        "type":    "user_joined",
        "user_id": conn.UserID,
        "count":   len(r.Members),
    })
}

func (r *ChatRoom) Leave(connID string) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    if conn, exists := r.Members[connID]; exists {
        delete(r.Members, connID)
        
        // Notify others
        r.Broadcast(fiber.Map{
            "type":    "user_left",
            "user_id": conn.UserID,
            "count":   len(r.Members),
        })
    }
}

func (r *ChatRoom) Broadcast(message interface{}) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    for _, conn := range r.Members {
        conn.SendJSON(message)
    }
}

func (r *ChatRoom) BroadcastExcept(exceptID string, message interface{}) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    for id, conn := range r.Members {
        if id != exceptID {
            conn.SendJSON(message)
        }
    }
}
```

## Broadcasting

### Global Broadcast

```go
// Broadcast to all connections
hub.Broadcast([]byte("Server announcement"))

// Broadcast JSON to all
hub.BroadcastJSON(fiber.Map{
    "type":    "announcement",
    "message": "Server maintenance in 5 minutes",
})
```

### User-Specific Messages

```go
// Send to specific user (all their connections)
hub.SendToUser(userID, []byte("Hello!"))

// Send JSON to user
hub.SendToUserJSON(userID, fiber.Map{
    "type":    "notification",
    "message": "You have a new message",
})
```

### Targeted Broadcasting

```go
// Send to multiple users
func sendToUsers(hub *websocket.Hub, userIDs []uint, message interface{}) {
    for _, userID := range userIDs {
        hub.SendToUserJSON(userID, message)
    }
}

// Send to users in a group
func sendToGroup(hub *websocket.Hub, groupID uint, message interface{}) {
    userIDs := getUsersInGroup(groupID)
    sendToUsers(hub, userIDs, message)
}
```

## Message Types

### Message Structure

```go
type Message struct {
    Type      string                 `json:"type"`
    Data      interface{}            `json:"data,omitempty"`
    Timestamp time.Time              `json:"timestamp"`
    Sender    uint                   `json:"sender,omitempty"`
    Room      string                 `json:"room,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
```

### Chat Message

```go
type ChatMessage struct {
    Type    string `json:"type"`
    RoomID  string `json:"room_id"`
    UserID  uint   `json:"user_id"`
    Message string `json:"message"`
    Time    int64  `json:"time"`
}

func handleChatMessage(hub *websocket.Hub, conn *websocket.Connection, data []byte) {
    var msg ChatMessage
    if err := json.Unmarshal(data, &msg); err != nil {
        return
    }
    
    // Get room
    room, exists := hub.GetRoom(msg.RoomID)
    if !exists {
        conn.SendJSON(fiber.Map{
            "type":  "error",
            "error": "Room not found",
        })
        return
    }
    
    // Broadcast to room
    room.Broadcast(fiber.Map{
        "type":    "chat_message",
        "user_id": msg.UserID,
        "message": msg.Message,
        "time":    time.Now().Unix(),
    })
    
    // Save to database
    saveMessage(msg.RoomID, msg.UserID, msg.Message)
}
```

### Typing Indicator

```go
func handleTypingIndicator(hub *websocket.Hub, conn *websocket.Connection, msg Message) {
    roomID := msg.Room
    
    room, exists := hub.GetRoom(roomID)
    if !exists {
        return
    }
    
    // Broadcast to others in room
    room.BroadcastExcept(conn.ID, fiber.Map{
        "type":    "typing",
        "user_id": conn.UserID,
        "typing":  msg.Data.(bool),
    })
}
```

### Presence Updates

```go
func sendPresenceUpdate(hub *websocket.Hub, userID uint, status string) {
    hub.BroadcastJSON(fiber.Map{
        "type":    "presence",
        "user_id": userID,
        "status":  status,
        "time":    time.Now().Unix(),
    })
}

// When user connects
func onUserConnect(hub *websocket.Hub, userID uint) {
    sendPresenceUpdate(hub, userID, "online")
}

// When user disconnects
func onUserDisconnect(hub *websocket.Hub, userID uint) {
    sendPresenceUpdate(hub, userID, "offline")
}
```

## Best Practices

### 1. Authenticate Connections

```go
func websocketHandler(c *websocket.Conn) {
    // Get token from query parameter or header
    token := c.Query("token")
    
    // Validate JWT token
    claims, err := jwtManager.ValidateToken(token)
    if err != nil {
        c.WriteMessage(websocket.CloseMessage, []byte("Unauthorized"))
        return
    }
    
    // Create authenticated connection
    conn := websocket.NewConnection(c, claims.UserID)
    hub.Register(conn)
    defer hub.Unregister(conn.ID)
    
    // Handle messages...
}
```

### 2. Implement Heartbeat

```go
func (conn *Connection) StartHeartbeat(interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        case <-conn.done:
            return
        }
    }
}
```

### 3. Handle Errors Gracefully

```go
func handleWebSocketMessages(hub *websocket.Hub, conn *websocket.Connection) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("WebSocket panic recovered: %v", r)
        }
        hub.Unregister(conn.ID)
    }()
    
    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
                log.Printf("WebSocket error: %v", err)
            }
            break
        }
        
        // Process message with error handling
        if err := processMessage(message); err != nil {
            conn.SendJSON(fiber.Map{
                "type":  "error",
                "error": err.Error(),
            })
        }
    }
}
```

### 4. Rate Limit Messages

```go
type RateLimiter struct {
    messages  map[string]int
    mu        sync.Mutex
    maxPerSec int
}

func (rl *RateLimiter) Allow(connID string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    count := rl.messages[connID]
    if count >= rl.maxPerSec {
        return false
    }
    
    rl.messages[connID]++
    return true
}

func handleMessage(conn *websocket.Connection, data []byte) {
    if !rateLimiter.Allow(conn.ID) {
        conn.SendJSON(fiber.Map{
            "type":  "error",
            "error": "Rate limit exceeded",
        })
        return
    }
    
    // Process message...
}
```

### 5. Clean Up Resources

```go
func (hub *Hub) Shutdown() {
    log.Println("Shutting down WebSocket hub...")
    
    // Stop accepting new connections
    hub.mu.Lock()
    hub.accepting = false
    hub.mu.Unlock()
    
    // Notify all clients
    hub.Broadcast([]byte(`{"type":"server_shutdown","message":"Server is shutting down"}`))
    
    // Wait a moment for messages to be sent
    time.Sleep(time.Second)
    
    // Close all connections
    hub.Close()
    
    log.Println("WebSocket hub shut down")
}

// Use in main
func main() {
    hub := websocket.NewHub(websocket.DefaultHubConfig())
    
    // Handle shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        hub.Shutdown()
        os.Exit(0)
    }()
    
    // Start server...
}
```

### 6. Monitor Connection Health

```go
func monitorConnections(hub *websocket.Hub) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        stats := fiber.Map{
            "total_connections": hub.ConnectionCount(),
            "unique_users":      hub.UserCount(),
            "rooms":             len(hub.GetRooms()),
            "timestamp":         time.Now().Unix(),
        }
        
        log.Printf("WebSocket Stats: %+v", stats)
        
        // Send to monitoring system
        metrics.Record("websocket.connections", hub.ConnectionCount())
        metrics.Record("websocket.users", hub.UserCount())
    }
}
```

## Complete Example

```go
package main

import (
    "neonexcore/pkg/websocket"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/websocket/v2"
)

func main() {
    app := fiber.New()
    
    // Create WebSocket hub
    hub := websocket.NewHub(websocket.DefaultHubConfig())
    
    // Create chat rooms
    hub.CreateRoom("general")
    hub.CreateRoom("random")
    
    // WebSocket endpoint
    app.Use("/ws", func(c *fiber.Ctx) error {
        if websocket.IsWebSocketUpgrade(c) {
            return c.Next()
        }
        return fiber.ErrUpgradeRequired
    })
    
    app.Get("/ws", websocket.New(func(c *websocket.Conn) {
        userID := authenticate(c)
        
        conn := websocket.NewConnection(c, userID)
        hub.Register(conn)
        defer hub.Unregister(conn.ID)
        
        // Join default room
        room := hub.GetRoom("general")
        room.Join(conn)
        
        // Handle messages
        handleWebSocketMessages(hub, conn)
    }))
    
    app.Listen(":3000")
}
```

This comprehensive guide covers WebSocket support in NeonEx Framework!

package websocket

import (
	"fmt"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Handler is the main WebSocket handler
type Handler struct {
	hub            *Hub
	messageHandler MessageHandler
	onConnect      func(*Connection)
	onDisconnect   func(*Connection)
}

// HandlerConfig configures the WebSocket handler
type HandlerConfig struct {
	Hub            *Hub
	MessageHandler MessageHandler
	OnConnect      func(*Connection)
	OnDisconnect   func(*Connection)
}

// NewHandler creates a new WebSocket handler
func NewHandler(config HandlerConfig) *Handler {
	return &Handler{
		hub:            config.Hub,
		messageHandler: config.MessageHandler,
		onConnect:      config.OnConnect,
		onDisconnect:   config.OnDisconnect,
	}
}

// HandleConnection handles a new WebSocket connection
func (h *Handler) HandleConnection(c *websocket.Conn) {
	// Generate connection ID
	connID := uuid.New().String()
	
	// Get user ID from context (set by auth middleware)
	userID := uint(0)
	if uid := c.Locals("userID"); uid != nil {
		if id, ok := uid.(uint); ok {
			userID = id
		}
	}
	
	// Create connection
	conn := NewConnection(connID, userID, c)
	
	// Register with hub
	if err := h.hub.Register(conn); err != nil {
		conn.Close()
		return
	}
	
	defer func() {
		h.hub.Unregister(connID)
		if h.onDisconnect != nil {
			h.onDisconnect(conn)
		}
	}()
	
	// Call onConnect callback
	if h.onConnect != nil {
		h.onConnect(conn)
	}
	
	// Send welcome message
	welcomeMsg := NewMessage(TypeSystem, SystemPayload{
		Event:   "connected",
		Message: "Connected to WebSocket server",
		Data: map[string]interface{}{
			"connection_id": connID,
			"user_id":       userID,
		},
	})
	conn.SendJSON(welcomeMsg)
	
	// Read loop
	for {
		var msg Message
		if err := c.ReadJSON(&msg); err != nil {
			break
		}
		
		// Update last ping
		conn.UpdatePing()
		
		// Handle message
		if h.messageHandler != nil {
			if err := h.messageHandler(conn, &msg); err != nil {
				errMsg := NewMessage(TypeError, ErrorPayload{
					Code:    "HANDLER_ERROR",
					Message: err.Error(),
				})
				conn.SendJSON(errMsg)
			}
		} else {
			// Default message handler
			h.defaultMessageHandler(conn, &msg)
		}
	}
}

// defaultMessageHandler is the default message handler
func (h *Handler) defaultMessageHandler(conn *Connection, msg *Message) error {
	switch msg.Type {
	case TypePing:
		// Respond with pong
		pongMsg := NewMessage(TypePong, nil)
		return conn.SendJSON(pongMsg)
		
	case TypeJoinRoom:
		// Join room
		if msg.Room == "" {
			return fmt.Errorf("room name required")
		}
		
		// Create room if not exists
		room := h.hub.CreateRoom(msg.Room)
		room.Join(conn)
		
		// Send confirmation
		roomMsg := NewMessage(TypeSystem, RoomPayload{
			Room:    msg.Room,
			Action:  "joined",
			Members: room.MemberCount(),
		})
		conn.SendJSON(roomMsg)
		
		// Notify room members
		notifyMsg := NewMessage(TypeSystem, SystemPayload{
			Event:   "user_joined",
			Message: fmt.Sprintf("User %d joined room", conn.UserID),
			Data: map[string]interface{}{
				"user_id": conn.UserID,
				"room":    msg.Room,
			},
		})
		room.Broadcast([]byte(fmt.Sprintf("%v", notifyMsg)), conn.ID)
		
	case TypeLeaveRoom:
		// Leave room
		if msg.Room == "" {
			return fmt.Errorf("room name required")
		}
		
		if err := h.hub.LeaveRoom(conn.ID, msg.Room); err != nil {
			return err
		}
		
		// Send confirmation
		roomMsg := NewMessage(TypeSystem, RoomPayload{
			Room:   msg.Room,
			Action: "left",
		})
		conn.SendJSON(roomMsg)
		
	case TypeRoomMessage:
		// Send message to room
		if msg.Room == "" {
			return fmt.Errorf("room name required")
		}
		
		msg.From = conn.UserID
		msg.Timestamp = msg.Timestamp
		
		data, _ := msg.ToJSON()
		return h.hub.BroadcastToRoom(msg.Room, data)
		
	case TypeUserMessage:
		// Send message to specific user
		if msg.To == 0 {
			return fmt.Errorf("recipient user ID required")
		}
		
		msg.From = conn.UserID
		data, _ := msg.ToJSON()
		h.hub.SendToUser(msg.To, data)
		
	case TypeBroadcast:
		// Broadcast to all connections
		msg.From = conn.UserID
		data, _ := msg.ToJSON()
		h.hub.Broadcast(data)
		
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
	
	return nil
}

// Middleware creates a Fiber middleware for WebSocket upgrade
func (h *Handler) Middleware() fiber.Handler {
	return websocket.New(h.HandleConnection, websocket.Config{
		RecoveryHandler: func(conn *websocket.Conn) {
			if err := recover(); err != nil {
				fmt.Printf("WebSocket panic: %v\n", err)
			}
		},
	})
}

// SetupRoutes sets up WebSocket routes
func SetupRoutes(app fiber.Router, hub *Hub, messageHandler MessageHandler) {
	handler := NewHandler(HandlerConfig{
		Hub:            hub,
		MessageHandler: messageHandler,
		OnConnect: func(conn *Connection) {
			fmt.Printf("Client connected: %s (User: %d)\n", conn.ID, conn.UserID)
		},
		OnDisconnect: func(conn *Connection) {
			fmt.Printf("Client disconnected: %s (User: %d)\n", conn.ID, conn.UserID)
		},
	})
	
	// WebSocket upgrade endpoint
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	
	app.Get("/ws", handler.Middleware())
	
	// Stats endpoint
	app.Get("/ws/stats", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"connections": hub.ConnectionCount(),
			"users":       hub.UserCount(),
			"rooms":       hub.RoomCount(),
			"room_list":   hub.ListRooms(),
		})
	})
}

package websocket

import (
	"encoding/json"
	"time"
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	TypePing         MessageType = "ping"
	TypePong         MessageType = "pong"
	TypeMessage      MessageType = "message"
	TypeBroadcast    MessageType = "broadcast"
	TypeJoinRoom     MessageType = "join_room"
	TypeLeaveRoom    MessageType = "leave_room"
	TypeRoomMessage  MessageType = "room_message"
	TypeUserMessage  MessageType = "user_message"
	TypeNotification MessageType = "notification"
	TypeError        MessageType = "error"
	TypeSystem       MessageType = "system"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType            `json:"type"`
	Payload   interface{}            `json:"payload,omitempty"`
	Room      string                 `json:"room,omitempty"`
	To        uint                   `json:"to,omitempty"`
	From      uint                   `json:"from,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewMessage creates a new message
func NewMessage(msgType MessageType, payload interface{}) *Message {
	return &Message{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
}

// ToJSON serializes message to JSON
func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON deserializes message from JSON
func (m *Message) FromJSON(data []byte) error {
	return json.Unmarshal(data, m)
}

// WithRoom sets the room for the message
func (m *Message) WithRoom(room string) *Message {
	m.Room = room
	return m
}

// WithTo sets the recipient user ID
func (m *Message) WithTo(userID uint) *Message {
	m.To = userID
	return m
}

// WithFrom sets the sender user ID
func (m *Message) WithFrom(userID uint) *Message {
	m.From = userID
	return m
}

// WithMetadata adds metadata to the message
func (m *Message) WithMetadata(key string, value interface{}) *Message {
	if m.Metadata == nil {
		m.Metadata = make(map[string]interface{})
	}
	m.Metadata[key] = value
	return m
}

// MessageHandler handles incoming WebSocket messages
type MessageHandler func(conn *Connection, msg *Message) error

// ErrorPayload represents an error message payload
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NotificationPayload represents a notification message payload
type NotificationPayload struct {
	Title   string      `json:"title"`
	Message string      `json:"message"`
	Level   string      `json:"level"` // info, success, warning, error
	Data    interface{} `json:"data,omitempty"`
}

// SystemPayload represents a system message payload
type SystemPayload struct {
	Event   string      `json:"event"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// RoomPayload represents a room-related payload
type RoomPayload struct {
	Room    string      `json:"room"`
	Action  string      `json:"action"` // join, leave, message
	Data    interface{} `json:"data,omitempty"`
	Members int         `json:"members,omitempty"`
}

// UserMessagePayload represents a user-to-user message payload
type UserMessagePayload struct {
	From    uint        `json:"from"`
	To      uint        `json:"to"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

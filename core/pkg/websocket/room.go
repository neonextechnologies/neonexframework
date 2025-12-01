package websocket

import (
	"sync"
)

// Room represents a WebSocket room for group communication
type Room struct {
	Name        string
	connections map[string]*Connection
	mu          sync.RWMutex
	Metadata    map[string]interface{}
}

// NewRoom creates a new room
func NewRoom(name string) *Room {
	return &Room{
		Name:        name,
		connections: make(map[string]*Connection),
		Metadata:    make(map[string]interface{}),
	}
}

// Join adds a connection to the room
func (r *Room) Join(conn *Connection) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connections[conn.ID] = conn
}

// Leave removes a connection from the room
func (r *Room) Leave(connID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.connections, connID)
}

// Broadcast sends a message to all connections in the room
func (r *Room) Broadcast(message []byte, excludeConnID ...string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	exclude := make(map[string]bool)
	for _, id := range excludeConnID {
		exclude[id] = true
	}
	
	for _, conn := range r.connections {
		if !exclude[conn.ID] {
			conn.Send(message)
		}
	}
}

// BroadcastJSON sends a JSON message to all connections in the room
func (r *Room) BroadcastJSON(v interface{}, excludeConnID ...string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	exclude := make(map[string]bool)
	for _, id := range excludeConnID {
		exclude[id] = true
	}
	
	for _, conn := range r.connections {
		if !exclude[conn.ID] {
			conn.SendJSON(v)
		}
	}
}

// MemberCount returns the number of connections in the room
func (r *Room) MemberCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.connections)
}

// Members returns all connection IDs in the room
func (r *Room) Members() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	members := make([]string, 0, len(r.connections))
	for id := range r.connections {
		members = append(members, id)
	}
	return members
}

// HasMember checks if a connection is in the room
func (r *Room) HasMember(connID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.connections[connID]
	return ok
}

// CreateRoom creates a new room in the hub
func (h *Hub) CreateRoom(name string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if room, exists := h.rooms[name]; exists {
		return room
	}
	
	room := NewRoom(name)
	h.rooms[name] = room
	return room
}

// GetRoom retrieves a room by name
func (h *Hub) GetRoom(name string) (*Room, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	room, ok := h.rooms[name]
	return room, ok
}

// DeleteRoom removes a room
func (h *Hub) DeleteRoom(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.rooms, name)
}

// JoinRoom adds a connection to a room
func (h *Hub) JoinRoom(connID, roomName string) error {
	conn, ok := h.GetConnection(connID)
	if !ok {
		return ErrConnectionClosed
	}
	
	room, ok := h.GetRoom(roomName)
	if !ok {
		return ErrRoomNotFound
	}
	
	room.Join(conn)
	return nil
}

// LeaveRoom removes a connection from a room
func (h *Hub) LeaveRoom(connID, roomName string) error {
	room, ok := h.GetRoom(roomName)
	if !ok {
		return ErrRoomNotFound
	}
	
	room.Leave(connID)
	return nil
}

// BroadcastToRoom sends a message to all connections in a room
func (h *Hub) BroadcastToRoom(roomName string, message []byte, excludeConnID ...string) error {
	room, ok := h.GetRoom(roomName)
	if !ok {
		return ErrRoomNotFound
	}
	
	room.Broadcast(message, excludeConnID...)
	return nil
}

// BroadcastToRoomJSON sends a JSON message to all connections in a room
func (h *Hub) BroadcastToRoomJSON(roomName string, v interface{}, excludeConnID ...string) error {
	room, ok := h.GetRoom(roomName)
	if !ok {
		return ErrRoomNotFound
	}
	
	room.BroadcastJSON(v, excludeConnID...)
	return nil
}

// RoomCount returns the total number of rooms
func (h *Hub) RoomCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms)
}

// ListRooms returns all room names
func (h *Hub) ListRooms() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	rooms := make([]string, 0, len(h.rooms))
	for name := range h.rooms {
		rooms = append(rooms, name)
	}
	return rooms
}

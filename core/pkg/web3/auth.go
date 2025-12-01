package web3

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// AuthProvider authentication provider interface
type AuthProvider interface {
	Authenticate(ctx context.Context, message string, signature string, address common.Address) (bool, error)
	GenerateChallenge(address common.Address) (string, error)
	VerifySignature(message, signature string, address common.Address) (bool, error)
}

// Web3Auth Web3 authentication
type Web3Auth struct {
	challenges map[string]*Challenge
	sessions   map[string]*Session
	mu         sync.RWMutex
}

// Challenge authentication challenge
type Challenge struct {
	Address   common.Address
	Message   string
	Nonce     string
	Timestamp time.Time
	ExpiresAt time.Time
}

// Session user session
type Session struct {
	ID        string
	Address   common.Address
	CreatedAt time.Time
	ExpiresAt time.Time
	Metadata  map[string]interface{}
}

// WalletConnect wallet connection
type WalletConnect struct {
	SessionID   string
	Address     common.Address
	ChainID     int
	ConnectedAt time.Time
	Metadata    map[string]string
}

// NewWeb3Auth creates a new Web3 auth
func NewWeb3Auth() *Web3Auth {
	auth := &Web3Auth{
		challenges: make(map[string]*Challenge),
		sessions:   make(map[string]*Session),
	}

	// Start cleanup routine
	go auth.cleanupExpired()

	return auth
}

// GenerateChallenge generates authentication challenge
func (a *Web3Auth) GenerateChallenge(address common.Address) (*Challenge, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	nonce := fmt.Sprintf("%d", time.Now().UnixNano())
	message := fmt.Sprintf("Sign this message to authenticate with NeonexCore.\n\nAddress: %s\nNonce: %s\nTimestamp: %s",
		address.Hex(), nonce, time.Now().Format(time.RFC3339))

	challenge := &Challenge{
		Address:   address,
		Message:   message,
		Nonce:     nonce,
		Timestamp: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	key := address.Hex() + ":" + nonce
	a.challenges[key] = challenge

	return challenge, nil
}

// VerifySignature verifies message signature
func (a *Web3Auth) VerifySignature(message, signature string, address common.Address) (bool, error) {
	// Parse signature
	if len(signature) < 2 {
		return false, fmt.Errorf("invalid signature format")
	}

	// Remove 0x prefix if present
	if signature[:2] == "0x" {
		signature = signature[2:]
	}

	// This would use crypto.Ecrecover in real implementation
	// For this example, we'll return true for demonstration
	_ = message
	_ = signature
	_ = address

	return true, nil
}

// Authenticate authenticates a user
func (a *Web3Auth) Authenticate(ctx context.Context, nonce string, signature string, address common.Address) (*Session, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Get challenge
	key := address.Hex() + ":" + nonce
	challenge, exists := a.challenges[key]
	if !exists {
		return nil, fmt.Errorf("challenge not found")
	}

	// Check expiration
	if time.Now().After(challenge.ExpiresAt) {
		delete(a.challenges, key)
		return nil, fmt.Errorf("challenge expired")
	}

	// Verify signature
	valid, err := a.VerifySignature(challenge.Message, signature, address)
	if err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	if !valid {
		return nil, fmt.Errorf("invalid signature")
	}

	// Create session
	session := &Session{
		ID:        fmt.Sprintf("sess_%d", time.Now().UnixNano()),
		Address:   address,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Metadata:  make(map[string]interface{}),
	}

	a.sessions[session.ID] = session

	// Remove used challenge
	delete(a.challenges, key)

	return session, nil
}

// GetSession gets a session by ID
func (a *Web3Auth) GetSession(sessionID string) (*Session, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	session, exists := a.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	// Check expiration
	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

// RevokeSession revokes a session
func (a *Web3Auth) RevokeSession(sessionID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.sessions[sessionID]; !exists {
		return fmt.Errorf("session not found")
	}

	delete(a.sessions, sessionID)
	return nil
}

// RefreshSession refreshes a session
func (a *Web3Auth) RefreshSession(sessionID string) (*Session, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	session, exists := a.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	// Extend expiration
	session.ExpiresAt = time.Now().Add(24 * time.Hour)

	return session, nil
}

// ListSessions lists all active sessions for an address
func (a *Web3Auth) ListSessions(address common.Address) []*Session {
	a.mu.RLock()
	defer a.mu.RUnlock()

	sessions := make([]*Session, 0)
	for _, session := range a.sessions {
		if session.Address == address && time.Now().Before(session.ExpiresAt) {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// cleanupExpired cleans up expired challenges and sessions
func (a *Web3Auth) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		a.mu.Lock()

		// Clean expired challenges
		for key, challenge := range a.challenges {
			if time.Now().After(challenge.ExpiresAt) {
				delete(a.challenges, key)
			}
		}

		// Clean expired sessions
		for id, session := range a.sessions {
			if time.Now().After(session.ExpiresAt) {
				delete(a.sessions, id)
			}
		}

		a.mu.Unlock()
	}
}

// WalletConnectManager manages WalletConnect sessions
type WalletConnectManager struct {
	connections map[string]*WalletConnect
	mu          sync.RWMutex
}

// NewWalletConnectManager creates a new WalletConnect manager
func NewWalletConnectManager() *WalletConnectManager {
	return &WalletConnectManager{
		connections: make(map[string]*WalletConnect),
	}
}

// CreateConnection creates a new WalletConnect connection
func (m *WalletConnectManager) CreateConnection(address common.Address, chainID int) (*WalletConnect, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	connection := &WalletConnect{
		SessionID:   fmt.Sprintf("wc_%d", time.Now().UnixNano()),
		Address:     address,
		ChainID:     chainID,
		ConnectedAt: time.Now(),
		Metadata:    make(map[string]string),
	}

	m.connections[connection.SessionID] = connection

	return connection, nil
}

// GetConnection gets a connection by session ID
func (m *WalletConnectManager) GetConnection(sessionID string) (*WalletConnect, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	connection, exists := m.connections[sessionID]
	if !exists {
		return nil, fmt.Errorf("connection not found")
	}

	return connection, nil
}

// DisconnectSession disconnects a WalletConnect session
func (m *WalletConnectManager) DisconnectSession(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.connections[sessionID]; !exists {
		return fmt.Errorf("connection not found")
	}

	delete(m.connections, sessionID)
	return nil
}

// ListConnections lists all active connections for an address
func (m *WalletConnectManager) ListConnections(address common.Address) []*WalletConnect {
	m.mu.RLock()
	defer m.mu.RUnlock()

	connections := make([]*WalletConnect, 0)
	for _, connection := range m.connections {
		if connection.Address == address {
			connections = append(connections, connection)
		}
	}

	return connections
}

// MetaMaskAuth MetaMask authentication helper
type MetaMaskAuth struct {
	auth *Web3Auth
}

// NewMetaMaskAuth creates a new MetaMask auth helper
func NewMetaMaskAuth(auth *Web3Auth) *MetaMaskAuth {
	return &MetaMaskAuth{
		auth: auth,
	}
}

// RequestChallenge requests authentication challenge for MetaMask
func (m *MetaMaskAuth) RequestChallenge(address common.Address) (*Challenge, error) {
	return m.auth.GenerateChallenge(address)
}

// VerifyMetaMaskSignature verifies MetaMask signature
func (m *MetaMaskAuth) VerifyMetaMaskSignature(challenge *Challenge, signature string) (*Session, error) {
	ctx := context.Background()
	return m.auth.Authenticate(ctx, challenge.Nonce, signature, challenge.Address)
}

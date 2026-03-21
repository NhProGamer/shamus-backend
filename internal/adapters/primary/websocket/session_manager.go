package websocket

import (
	"errors"
	"shamus-backend/internal/domain/entities"
	"sync"

	"github.com/olahol/melody"
)

// SessionManager manages WebSocket sessions and game rooms
type SessionManager struct {
	// rooms maps GameID to list of sessions for targeted broadcast
	rooms map[entities.GameID][]*melody.Session
	// playerSessions maps PlayerID to session for O(1) player lookup
	playerSessions map[entities.PlayerID]*melody.Session
	// sessionPlayers maps session to PlayerID (reverse lookup for disconnect)
	sessionPlayers map[*melody.Session]entities.PlayerID
	lock           sync.RWMutex
}

// NewSessionManager creates a new SessionManager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		rooms:          make(map[entities.GameID][]*melody.Session),
		playerSessions: make(map[entities.PlayerID]*melody.Session),
		sessionPlayers: make(map[*melody.Session]entities.PlayerID),
	}
}

// JoinRoom adds a player's session to a game room
func (m *SessionManager) JoinRoom(gameID entities.GameID, playerID entities.PlayerID, session *melody.Session) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.rooms[gameID] = append(m.rooms[gameID], session)
	m.playerSessions[playerID] = session
	m.sessionPlayers[session] = playerID
}

// LeaveRoom removes a player's session from a game room
func (m *SessionManager) LeaveRoom(gameID entities.GameID, playerID entities.PlayerID, session *melody.Session) {
	m.lock.Lock()
	defer m.lock.Unlock()

	// Remove from room
	sessions := m.rooms[gameID]
	for i, sess := range sessions {
		if sess == session {
			m.rooms[gameID] = append(sessions[:i], sessions[i+1:]...)
			break
		}
	}

	// Clean up empty rooms
	if len(m.rooms[gameID]) == 0 {
		delete(m.rooms, gameID)
	}

	// Remove from player mappings
	delete(m.playerSessions, playerID)
	delete(m.sessionPlayers, session)
}

// GetPlayerSession returns the session for a player
func (m *SessionManager) GetPlayerSession(playerID entities.PlayerID) (*melody.Session, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	session, exists := m.playerSessions[playerID]
	return session, exists
}

// GetSessionPlayer returns the player ID for a session
func (m *SessionManager) GetSessionPlayer(session *melody.Session) (entities.PlayerID, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	playerID, exists := m.sessionPlayers[session]
	return playerID, exists
}

// GetRoomSessions returns all sessions in a game room
func (m *SessionManager) GetRoomSessions(gameID entities.GameID) []*melody.Session {
	m.lock.RLock()
	defer m.lock.RUnlock()

	sessions := m.rooms[gameID]
	// Return a copy to avoid external modifications
	result := make([]*melody.Session, len(sessions))
	copy(result, sessions)
	return result
}

// IsPlayerConnected checks if a player has an active session
func (m *SessionManager) IsPlayerConnected(playerID entities.PlayerID) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()

	_, exists := m.playerSessions[playerID]
	return exists
}

// GetRoomPlayerCount returns the number of connected players in a room
func (m *SessionManager) GetRoomPlayerCount(gameID entities.GameID) int {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return len(m.rooms[gameID])
}

// SendToPlayer sends a message to a specific player
func (m *SessionManager) SendToPlayer(playerID entities.PlayerID, payload []byte) error {
	m.lock.RLock()
	session, exists := m.playerSessions[playerID]
	m.lock.RUnlock()

	if !exists {
		return ErrPlayerNotConnected
	}

	return session.Write(payload)
}

// BroadcastToGame sends a message to all players in a game room
func (m *SessionManager) BroadcastToGame(gameID entities.GameID, payload []byte) error {
	m.lock.RLock()
	sessions, exists := m.rooms[gameID]
	if !exists || len(sessions) == 0 {
		m.lock.RUnlock()
		return ErrEmptyRoom
	}

	// Copy sessions to avoid holding lock during writes
	sessionsCopy := make([]*melody.Session, len(sessions))
	copy(sessionsCopy, sessions)
	m.lock.RUnlock()

	for _, sess := range sessionsCopy {
		_ = sess.Write(payload)
	}

	return nil
}

// SendToPlayers sends a message to specific players
func (m *SessionManager) SendToPlayers(playerIDs []entities.PlayerID, payload []byte) {
	m.lock.RLock()
	sessions := make([]*melody.Session, 0, len(playerIDs))
	for _, playerID := range playerIDs {
		if sess, ok := m.playerSessions[playerID]; ok {
			sessions = append(sessions, sess)
		}
	}
	m.lock.RUnlock()

	for _, sess := range sessions {
		_ = sess.Write(payload)
	}
}

// Errors
var (
	ErrPlayerNotConnected = errors.New("player not connected")
	ErrEmptyRoom          = errors.New("room is empty or doesn't exist")
)

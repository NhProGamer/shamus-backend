package entities

import "sync"

type PlayerID string
type ConnectionState string

const (
	Connected    ConnectionState = "connected"
	Disconnected ConnectionState = "disconnected"
	Inactive     ConnectionState = "inactive"
)

type Player struct {
	id              PlayerID
	username        string
	role            *Role
	isAlive         bool
	votedFor        *PlayerID
	connectionState ConnectionState
	gameID          *GameID
}

type SafePlayer struct {
	mu     sync.RWMutex
	player *Player
}

func NewSafePlayer(id PlayerID, username string, gameID *GameID) *SafePlayer {
	return &SafePlayer{
		mu: sync.RWMutex{},
		player: &Player{
			id:              id,
			username:        username,
			role:            nil,
			isAlive:         true,
			votedFor:        nil,
			connectionState: Connected,
			gameID:          gameID,
		},
	}
}

// --- Getters ---
func (sp *SafePlayer) ID() PlayerID {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	return sp.player.id
}

func (sp *SafePlayer) Username() string {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	return sp.player.username
}

func (sp *SafePlayer) Role() *Role {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	return sp.player.role
}

func (sp *SafePlayer) Alive() bool {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	return sp.player.isAlive
}

func (sp *SafePlayer) VotedFor() *PlayerID {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	return sp.player.votedFor
}

func (sp *SafePlayer) ConnectionState() ConnectionState {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	return sp.player.connectionState
}

func (sp *SafePlayer) GetGameID() *GameID {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	return sp.player.gameID
}

// --- Mutateurs ---
func (sp *SafePlayer) AssignRole(r *Role) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.player.role = r
}

func (sp *SafePlayer) Kill() {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.player.isAlive = false
}

func (sp *SafePlayer) Revive() {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.player.isAlive = true
}

func (sp *SafePlayer) VoteFor(id *PlayerID) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.player.votedFor = id
}

func (sp *SafePlayer) ClearVote() {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.player.votedFor = nil
}

func (sp *SafePlayer) Connect() {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.player.connectionState = Connected
}

func (sp *SafePlayer) Disconnect() {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.player.connectionState = Disconnected
}

func (sp *SafePlayer) SetInactive() {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.player.connectionState = Inactive
}

func (sp *SafePlayer) SetGameID(id *GameID) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.player.gameID = id
}

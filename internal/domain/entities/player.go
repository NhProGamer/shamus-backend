package entities

type PlayerID string
type ConnectionState string

const (
	ConnectionStateConnected    ConnectionState = "connected"
	ConnectionStateDisconnected ConnectionState = "disconnected"
	ConnectionStateInactive     ConnectionState = "inactive"
)

type Player struct {
	ID              PlayerID        `json:"id"`
	Username        string          `json:"username"`
	Role            Role            `json:"-"`
	IsAlive         bool            `json:"isAlive"`
	VotedFor        *PlayerID       `json:"votedFor,omitempty"`
	ConnectionState ConnectionState `json:"connectionState"`
	GameID          *GameID         `json:"gameId,omitempty"`
}

func NewPlayer(id PlayerID, username string, gameID *GameID) *Player {
	return &Player{
		ID:              id,
		Username:        username,
		Role:            nil,
		IsAlive:         true,
		VotedFor:        nil,
		ConnectionState: ConnectionStateConnected,
		GameID:          gameID,
	}
}

// AssignRole assigns a role to the player
func (p *Player) AssignRole(r Role) {
	p.Role = r
}

// Kill marks the player as dead
func (p *Player) Kill() {
	p.IsAlive = false
}

// Revive marks the player as alive
func (p *Player) Revive() {
	p.IsAlive = true
}

// Vote sets the player's vote target
func (p *Player) Vote(target *PlayerID) {
	p.VotedFor = target
}

// ClearVote removes the player's vote
func (p *Player) ClearVote() {
	p.VotedFor = nil
}

// Connect sets the player's connection state to connected
func (p *Player) Connect() {
	p.ConnectionState = ConnectionStateConnected
}

// Disconnect sets the player's connection state to disconnected
func (p *Player) Disconnect() {
	p.ConnectionState = ConnectionStateDisconnected
}

// SetInactive sets the player's connection state to inactive
func (p *Player) SetInactive() {
	p.ConnectionState = ConnectionStateInactive
}

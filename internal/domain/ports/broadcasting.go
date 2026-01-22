package ports

import "shamus-backend/internal/domain/entities"

// Broadcaster is an interface for broadcasting messages to all players in a game
// This is the low-level interface used by internal services (accepts raw bytes)
type Broadcaster interface {
	BroadcastToGame(gameID entities.GameID, payload []byte) error
}

// PlayerSender is an interface for sending messages to specific players
// This is the low-level interface used by internal services (accepts raw bytes)
type PlayerSender interface {
	SendToPlayer(playerID entities.PlayerID, payload []byte) error
}

// GameMessenger combines both broadcasting and player-specific messaging
// Used by services that need both capabilities
type GameMessenger interface {
	Broadcaster
	PlayerSender
}

package ports

import "shamus-backend/internal/domain/entities"

// GameService defines the game business logic operations
type GameService interface {
	// CreateNewGame creates a new game with the given host
	CreateNewGame(hostID entities.PlayerID) (*entities.Game, error)

	// JoinGame adds a player to an existing game
	JoinGame(gameID entities.GameID, playerID entities.PlayerID) (*entities.Game, error)

	// GetGame retrieves a game by ID
	GetGame(gameID entities.GameID) (*entities.Game, error)
}

// EventService defines operations for sending events to players
type EventService interface {
	// SendToPlayer sends an event to a specific player
	SendToPlayer(playerID entities.PlayerID, event entities.RawEvent) error

	// BroadcastToGame sends an event to all players in a game
	BroadcastToGame(gameID entities.GameID, event entities.RawEvent) error
}

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

	// UpdateSettings updates game settings (roles configuration)
	// Only the host can update settings, and only when game is in waiting state
	UpdateSettings(gameID entities.GameID, playerID entities.PlayerID, settings entities.GameSettings) (*entities.Game, error)
}

// PlayerService defines player management operations
type PlayerService interface {
	// HandleConnect is called when a player connects via WebSocket
	// Returns the Player and a boolean indicating if this is a reconnection
	HandleConnect(gameID entities.GameID, playerID entities.PlayerID, username string) (*entities.Player, bool, error)

	// HandleDisconnect is called when a player disconnects from WebSocket
	HandleDisconnect(gameID entities.GameID, playerID entities.PlayerID) error

	// GetPlayer retrieves a player by ID
	GetPlayer(id entities.PlayerID) (*entities.Player, error)

	// GetGamePlayers retrieves all players in a game
	GetGamePlayers(gameID entities.GameID) ([]*entities.Player, error)

	// CleanupGamePlayers removes all players when a game ends
	CleanupGamePlayers(gameID entities.GameID) error

	// IsPlayerConnected checks if a player has an active WebSocket session
	IsPlayerConnected(playerID entities.PlayerID) bool
}

// EventService defines operations for sending events to players
type EventService interface {
	// SendToPlayer sends an event to a specific player
	SendToPlayer(playerID entities.PlayerID, event entities.RawEvent) error

	// BroadcastToGame sends an event to all players in a game
	BroadcastToGame(gameID entities.GameID, event entities.RawEvent) error
}

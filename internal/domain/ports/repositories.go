package ports

import "shamus-backend/internal/domain/entities"

// GameRepository defines persistence operations for games
type GameRepository interface {
	// SaveGame persists a game to storage
	SaveGame(game *entities.Game) error

	// GetGame retrieves a game by ID
	GetGame(id entities.GameID) (*entities.Game, error)

	// DeleteGame removes a game from storage
	DeleteGame(id entities.GameID) error
}

// PlayerRepository defines persistence operations for players
type PlayerRepository interface {
	// SavePlayer persists a player to storage
	SavePlayer(player *entities.Player) error

	// GetPlayer retrieves a player by ID
	GetPlayer(id entities.PlayerID) (*entities.Player, error)

	// DeletePlayer removes a player from storage
	DeletePlayer(id entities.PlayerID) error

	// GetPlayersByGame retrieves all players in a game
	GetPlayersByGame(gameID entities.GameID) ([]*entities.Player, error)

	// AddPlayerToGame adds a player ID to the game's player set
	AddPlayerToGame(gameID entities.GameID, playerID entities.PlayerID) error

	// RemovePlayerFromGame removes a player ID from the game's player set
	RemovePlayerFromGame(gameID entities.GameID, playerID entities.PlayerID) error

	// DeleteGamePlayers removes all players associated with a game (cleanup when game ends)
	DeleteGamePlayers(gameID entities.GameID) error
}

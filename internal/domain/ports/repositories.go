package ports

import (
	"context"
	"shamus-backend/internal/domain/entities"
)

// GameRepository defines persistence operations for games
type GameRepository interface {
	// SaveGame persists a game to storage
	SaveGame(ctx context.Context, game *entities.Game) error

	// GetGame retrieves a game by ID
	GetGame(ctx context.Context, id entities.GameID) (*entities.Game, error)

	// DeleteGame removes a game from storage
	DeleteGame(ctx context.Context, id entities.GameID) error
}

// PlayerRepository defines persistence operations for players
type PlayerRepository interface {
	// SavePlayer persists a player to storage
	SavePlayer(ctx context.Context, player *entities.Player) error

	// SavePlayers persists multiple players in a single batch operation
	SavePlayers(ctx context.Context, players []*entities.Player) error

	// GetPlayer retrieves a player by ID
	GetPlayer(ctx context.Context, id entities.PlayerID) (*entities.Player, error)

	// DeletePlayer removes a player from storage
	DeletePlayer(ctx context.Context, id entities.PlayerID) error

	// GetPlayersByGame retrieves all players in a game
	GetPlayersByGame(ctx context.Context, gameID entities.GameID) ([]*entities.Player, error)

	// AddPlayerToGame adds a player ID to the game's player set
	AddPlayerToGame(ctx context.Context, gameID entities.GameID, playerID entities.PlayerID) error

	// RemovePlayerFromGame removes a player ID from the game's player set
	RemovePlayerFromGame(ctx context.Context, gameID entities.GameID, playerID entities.PlayerID) error

	// DeleteGamePlayers removes all players associated with a game (cleanup when game ends)
	DeleteGamePlayers(ctx context.Context, gameID entities.GameID) error
}

// ActionRepository defines persistence operations for actions
type ActionRepository interface {
	// SaveAction persists an action to storage
	SaveAction(ctx context.Context, action *entities.Action) error

	// GetAction retrieves an action by ID
	GetAction(ctx context.Context, actionID entities.ActionID) (*entities.Action, error)

	// GetPendingActionsByPlayer retrieves all pending actions for a player
	GetPendingActionsByPlayer(ctx context.Context, playerID entities.PlayerID) ([]*entities.Action, error)

	// DeleteAction removes an action from storage
	DeleteAction(ctx context.Context, actionID entities.ActionID) error
}

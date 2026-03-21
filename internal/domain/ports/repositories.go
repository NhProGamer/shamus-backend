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

// VoteRepository defines persistence operations for votes
type VoteRepository interface {
	// SaveVote persists vote metadata to storage
	SaveVote(ctx context.Context, gameID entities.GameID, vote *entities.Vote) error

	// GetVote retrieves a vote from storage
	GetVote(ctx context.Context, gameID entities.GameID) (*entities.Vote, error)

	// CastBallot atomically records a single player's vote
	CastBallot(ctx context.Context, gameID entities.GameID, voterID entities.PlayerID, targetID *entities.PlayerID) error

	// GetBallot retrieves a single player's vote
	// Returns (targetID, hasCast, error) - hasCast is true if player has voted (even if abstention)
	GetBallot(ctx context.Context, gameID entities.GameID, voterID entities.PlayerID) (*entities.PlayerID, bool, error)

	// GetBallotCount returns the number of ballots cast
	GetBallotCount(ctx context.Context, gameID entities.GameID) (int64, error)

	// DeleteVote removes a vote and its ballots from storage
	DeleteVote(ctx context.Context, gameID entities.GameID) error

	// VoteExists checks if a vote exists for a game
	VoteExists(ctx context.Context, gameID entities.GameID) (bool, error)
}

package ports

import "shamus-backend/internal/domain/entities"

type GameRepository interface {
	GetGameByID(id entities.GameID) (*entities.Game, error)
	GetGameHoster(id entities.GameID) (*entities.PlayerID, error)
	CreateGame(game *entities.Game) error
	UpdateGame(game *entities.Game) error
	DeleteGame(id entities.GameID) error
}

type PlayerRepository interface {
	GetPlayerByID(id entities.PlayerID) (*entities.Player, error)
	GetPlayerGameID(id entities.PlayerID) (*entities.GameID, error)
}

type GameSettingsRepository interface {
	GetGameSettings(gameID entities.GameID) (*entities.GameSettings, error)
	SetGameSettings(gameID entities.GameID, settings *entities.GameSettings) error
}

type GameVotesRepository interface {
	GetVotes(gameID entities.GameID) (map[entities.PlayerID]*entities.PlayerID, error)
	SetVote(gameID entities.GameID, playerID entities.PlayerID, target entities.PlayerID)
	DeleteVote(gameID entities.GameID, playerID entities.PlayerID) error
	CleanVotes(gameID entities.GameID) error
}

package ports

import "shamus-backend/internal/domain/entities"

type GameRepository interface {
	GetGameByID(id entities.GameID) (*entities.Game, error)
	//GetGameHoster(id entities.GameID) (*entities.PlayerID, error)
	CreateGame(game *entities.Game) error
	//UpdateGame(game *entities.Game) error
	DeleteGame(id entities.GameID) error
}

type PlayerRepository interface {
	GetPlayerByID(id entities.PlayerID) (*entities.SafePlayer, error)
	GetPlayerGameID(id entities.PlayerID) (*entities.GameID, error)
	AddPlayer(player *entities.SafePlayer) error
	DeletePlayer(id entities.PlayerID) error
}

type GameVotesRepository interface {
	GetVotes(gameID entities.GameID) (map[entities.PlayerID]*entities.PlayerID, error)
	SetVote(gameID entities.GameID, playerID entities.PlayerID, target entities.PlayerID) error
	DeleteVote(gameID entities.GameID, playerID entities.PlayerID) error
	CleanVotes(gameID entities.GameID) error
}

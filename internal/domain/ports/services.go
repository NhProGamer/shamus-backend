package ports

import "shamus-backend/internal/domain/entities"

type GameService interface {
	//NextPhase(gameID entities.GameID) error
	//NextStep(gameID entities.GameID) error
	//IsGameEnded(gameID entities.GameID) (bool, error)

	JoinGame(gameID entities.GameID, playerID entities.PlayerID)
	CreateGame(hostID entities.PlayerID) (*entities.Game, error)
}

package ports

import "shamus-backend/internal/domain/entities"

type GameService interface {
	NextPhase(gameID entities.GameID) error
	NextStep(gameID entities.GameID) error
	IsGameEnded(gameID entities.GameID) (bool, error)
}

type VoteService interface {
	NewVote(gameID entities.GameID) error
	CloseVote(gameID entities.GameID) (*entities.PlayerID, error)
	AddVote(gameID entities.GameID, playerID entities.PlayerID, target entities.PlayerID) error
	RemoveVote(gameID entities.GameID, playerID entities.PlayerID) error
}

type EventService interface {
	SendEventToPlayer(event entities.Event, player entities.PlayerID)
	SendEventToGame(event entities.Event, gameID entities.GameID)
	//SendEventToClanInAGame(event entities.Event, gameID entities.GameID, clan entities.Clan)
}

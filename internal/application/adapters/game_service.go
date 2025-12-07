package adapters

import (
	"fmt"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/events"
	"shamus-backend/internal/domain/ports"
)

type gameService struct {
	games ports.GameRepository
	es    ports.EventService
}

func NewGameService(gr ports.GameRepository, es ports.EventService) ports.GameService {
	if es == nil {
		es = &nullEventService{}
	}
	return &gameService{games: gr, es: es}
}

func (s *gameService) NextPhase(gameID entities.GameID) error {
	game, err := s.games.GetGameByID(gameID)
	if err != nil {
		return err
	}

	current := game.Phase()
	switch current {
	case entities.PhaseStart:
		if err := game.SetPhase(entities.PhaseDay); err != nil {
			return err
		}
		s.es.SendEventToGame(events.NewDayEvent(nil).ToRawEvent(), gameID)
	case entities.PhaseDay:
		if err := game.SetPhase(entities.PhaseNight); err != nil {
			return err
		}
		s.es.SendEventToGame(events.NewNightEvent().ToRawEvent(), gameID)
	case entities.PhaseNight:
		if err := game.SetPhase(entities.PhaseVote); err != nil {
			return err
		}
		// Start vote is handled by VoteService in general, here we can emit a turn event as a signal
		s.es.SendEventToGame(events.NewTurnEvent().ToRawEvent(), gameID)
	case entities.PhaseVote:
		// End of cycle: go back to day and increment day counter
		game.NextDay()
		if err := game.SetPhase(entities.PhaseDay); err != nil {
			return err
		}
		s.es.SendEventToGame(events.NewDayEvent(nil).ToRawEvent(), gameID)
	default:
		return fmt.Errorf("unknown phase: %s", current)
	}
	return nil
}

func (s *gameService) NextStep(gameID entities.GameID) error {
	// For now, treat step == phase change
	return s.NextPhase(gameID)
}

func (s *gameService) IsGameEnded(gameID entities.GameID) (bool, error) {
	game, err := s.games.GetGameByID(gameID)
	if err != nil {
		return false, err
	}
	return game.IsEnded(), nil
}

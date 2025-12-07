package adapters

import (
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/events"
	"shamus-backend/internal/domain/ports"
)

type voteService struct {
	votesRepo    ports.GameVotesRepository
	eventService ports.EventService
}

// NewVoteService wires the VoteService with its required ports.
// eventSvc can be nil; in that case, a no-op implementation is used to avoid panics.
func NewVoteService(vr ports.GameVotesRepository, eventSvc ports.EventService) ports.VoteService {
	if eventSvc == nil {
		eventSvc = &nullEventService{}
	}
	return &voteService{votesRepo: vr, eventService: eventSvc}
}

func (s *voteService) NewVote(gameID entities.GameID) error {
	event := events.NewVoteEvent(events.StartVote, nil, nil).ToRawEvent()
	s.eventService.SendEventToGame(event, gameID)
	return nil
}

func (s *voteService) CloseVote(gameID entities.GameID) (*entities.PlayerID, error) {
	event := events.NewVoteEvent(events.EndVote, nil, nil).ToRawEvent()
	s.eventService.SendEventToGame(event, gameID)

	computedVotes, err := s.computeVotes(gameID)
	if err != nil {
		return nil, err
	}
	if err := s.votesRepo.CleanVotes(gameID); err != nil {
		return nil, err
	}
	return computedVotes, nil
}

func (s *voteService) AddVote(gameID entities.GameID, player entities.PlayerID, target entities.PlayerID) error {
	event := events.NewVoteEvent(events.PlayerVote, &player, &target).ToRawEvent()
	s.eventService.SendEventToGame(event, gameID)
	if err := s.votesRepo.SetVote(gameID, player, target); err != nil {
		return err
	}
	return nil
}

func (s *voteService) RemoveVote(gameID entities.GameID, player entities.PlayerID) error {
	event := events.NewVoteEvent(events.PlayerVote, &player, nil).ToRawEvent()
	s.eventService.SendEventToGame(event, gameID)
	return s.votesRepo.DeleteVote(gameID, player)
}

func (s *voteService) computeVotes(gameID entities.GameID) (*entities.PlayerID, error) {
	votes, err := s.votesRepo.GetVotes(gameID)
	if err != nil {
		return nil, err
	}

	// Compter les votes
	voteCount := make(map[entities.PlayerID]int)
	for _, target := range votes {
		if target != nil {
			voteCount[*target]++
		}
	}

	// Trouver le joueur avec le plus de votes
	var maxVotes int
	var victim *entities.PlayerID
	for playerID, count := range voteCount {
		if count > maxVotes {
			maxVotes = count
			victim = &playerID
		}
	}

	//TODO: Prendre en charge que le maire doit faire le choix décisif
	// juste renvoyer la map avec PlayerID -

	return victim, nil
}

// nullEventService is a safe no-op EventService used when none is provided.
type nullEventService struct{}

func (n *nullEventService) SendEventToPlayer(event entities.RawEvent, player entities.PlayerID) {}
func (n *nullEventService) SendEventToGame(event entities.RawEvent, gameID entities.GameID)     {}

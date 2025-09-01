package adapters

import (
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/events"
	"shamus-backend/internal/domain/ports"
)

type voteService struct {
	votesRepo    ports.GameVotesRepository
	gameRepo     ports.GameRepository
	playerRepo   ports.PlayerRepository
	eventService ports.EventService
}

func NewVoteService(vr ports.GameVotesRepository) ports.VoteService {
	return &voteService{votesRepo: vr}
}

func (s *voteService) NewVote(gameID entities.GameID) error {
	event := events.NewVoteEvent(events.StartVote, nil, nil)
	s.eventService.SendEventToGame(event, gameID)
	return nil
}

func (s *voteService) CloseVote(gameID entities.GameID) (*entities.PlayerID, error) {
	event := events.NewVoteEvent(events.EndVote, nil, nil)
	s.eventService.SendEventToGame(event, gameID)

	computedVotes, _ := s.computeVotes(gameID)
	s.votesRepo.CleanVotes(gameID)
	return computedVotes, nil
}

func (s *voteService) AddVote(gameID entities.GameID, player entities.PlayerID, target entities.PlayerID) error {
	event := events.NewVoteEvent(events.PlayerVote, &player, &target)
	s.eventService.SendEventToGame(event, gameID)
	s.votesRepo.SetVote(gameID, player, target)
	return nil
}

func (s *voteService) RemoveVote(gameID entities.GameID, player entities.PlayerID) error {
	event := events.NewVoteEvent(events.PlayerVote, &player, nil)
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

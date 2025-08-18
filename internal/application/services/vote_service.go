package services

import (
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/ports"
)

type voteService struct {
	votesRepo ports.GameVotesRepository
	gameRepo  ports.GameRepository
}

func NewVoteService(vr ports.GameVotesRepository) ports.VoteService {
	return &voteService{votesRepo: vr}
}

func (s *voteService) NewVote(gameID entities.GameID) error {
	return s.votesRepo.CleanVotes(gameID)
	// Send event to a new vote
}

func (s *voteService) GetVotes(gameID entities.GameID) (map[entities.PlayerID]*entities.PlayerID, error) {
	return s.votesRepo.GetVotes(gameID)
}

func (s *voteService) CleanVotes(gameID entities.GameID) error {
	return s.votesRepo.CleanVotes(gameID)
}

func (s *voteService) AddVote(gameID entities.GameID, voter entities.PlayerID, target entities.PlayerID) error {
	s.votesRepo.SetVote(gameID, voter, target)
	return nil
}

func (s *voteService) RemoveVote(gameID entities.GameID, voter entities.PlayerID) error {
	return s.votesRepo.DeleteVote(gameID, voter)
}

func (s *voteService) ComputeVotes(gameID entities.GameID) (*entities.PlayerID, error) {
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

	return victim, nil
}

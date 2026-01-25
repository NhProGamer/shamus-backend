package app

import (
	"context"
	"encoding/json"
	"shamus-backend/internal/adapters/infra"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/events"
	apperrors "shamus-backend/internal/domain/errors"
	"shamus-backend/internal/domain/ports"

	"github.com/google/uuid"
)

// VoteService manages voting sessions for games
type VoteService struct {
	voteRepo     *infra.VoteRepository
	broadcaster  ports.Broadcaster
	playerSender ports.PlayerSender
}

// NewVoteService creates a new VoteService
func NewVoteService(voteRepo *infra.VoteRepository, broadcaster ports.Broadcaster, playerSender ports.PlayerSender) *VoteService {
	return &VoteService{
		voteRepo:     voteRepo,
		broadcaster:  broadcaster,
		playerSender: playerSender,
	}
}

// StartVillageVote starts a village vote to eliminate a player
func (s *VoteService) StartVillageVote(gameID entities.GameID, alivePlayers []*entities.Player) (*entities.Vote, error) {
	ctx := context.TODO()

	// Check if vote already exists
	exists, err := s.voteRepo.VoteExists(ctx, gameID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperrors.ErrVoteAlreadyExists
	}

	// Build eligible voters and targets (all alive players)
	var voters, targets []entities.PlayerID
	for _, p := range alivePlayers {
		if p.IsAlive {
			voters = append(voters, p.ID)
			targets = append(targets, p.ID)
		}
	}

	vote := entities.NewVote(
		uuid.New().String(),
		entities.VoteTypeVillage,
		voters,
		targets,
		true, // Allow abstain
	)

	if err := s.voteRepo.SaveVote(ctx, gameID, vote); err != nil {
		return nil, err
	}

	// Broadcast vote start event
	s.broadcastVoteEvent(gameID, events.NewVoteEvent(events.StartVote, nil, nil))

	return vote, nil
}

// StartWerewolfVote starts a werewolf vote to choose a victim
func (s *VoteService) StartWerewolfVote(gameID entities.GameID, werewolves []*entities.Player, potentialVictims []*entities.Player) (*entities.Vote, error) {
	ctx := context.TODO()

	// Check if vote already exists
	exists, err := s.voteRepo.VoteExists(ctx, gameID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperrors.ErrVoteAlreadyExists
	}

	// Werewolves are the voters
	var voters []entities.PlayerID
	for _, w := range werewolves {
		if w.IsAlive {
			voters = append(voters, w.ID)
		}
	}

	// Potential victims are the targets (non-werewolves alive)
	var targets []entities.PlayerID
	for _, p := range potentialVictims {
		if p.IsAlive {
			targets = append(targets, p.ID)
		}
	}

	vote := entities.NewVote(
		uuid.New().String(),
		entities.VoteTypeWerewolf,
		voters,
		targets,
		true, // Allow no attack
	)

	if err := s.voteRepo.SaveVote(ctx, gameID, vote); err != nil {
		return nil, err
	}

	// Broadcast vote start event (only to werewolves - handled by caller)

	return vote, nil
}

// CastVote records a player's vote
func (s *VoteService) CastVote(gameID entities.GameID, voterID entities.PlayerID, targetID *entities.PlayerID) error {
	ctx := context.TODO()

	vote, err := s.voteRepo.GetVote(ctx, gameID)
	if err != nil {
		return err
	}
	if vote == nil {
		return apperrors.ErrVoteNotFound
	}

	if vote.Status != entities.VoteStatusActive {
		return apperrors.ErrVoteNotActive
	}

	if !vote.CastBallot(voterID, targetID) {
		return apperrors.ErrInvalidVoter
	}

	// Save updated vote
	if err := s.voteRepo.SaveVote(ctx, gameID, vote); err != nil {
		return err
	}

	// Broadcast player vote event
	event := events.NewVoteEvent(events.PlayerVote, &voterID, targetID)

	if vote.Type == entities.VoteTypeWerewolf {
		// Werewolf votes are only visible to werewolves (the eligible voters)
		s.broadcastToVoters(vote, event)
	} else {
		// Village votes are visible to everyone
		s.broadcastVoteEvent(gameID, event)
	}

	return nil
}

// HasEveryoneVoted checks if all eligible voters have voted
func (s *VoteService) HasEveryoneVoted(gameID entities.GameID) bool {
	ctx := context.TODO()

	vote, err := s.voteRepo.GetVote(ctx, gameID)
	if err != nil || vote == nil {
		return false
	}

	return vote.HasEveryoneVoted()
}

// ResolveVote resolves the current vote and returns the result
func (s *VoteService) ResolveVote(gameID entities.GameID) (*entities.VoteResult, error) {
	ctx := context.TODO()

	vote, err := s.voteRepo.GetVote(ctx, gameID)
	if err != nil {
		return nil, err
	}
	if vote == nil {
		return nil, apperrors.ErrVoteNotFound
	}

	result := vote.Resolve()

	// Broadcast vote end event with result
	s.broadcastVoteEvent(gameID, events.NewVoteEvent(events.EndVote, nil, result.Target))

	return result, nil
}

// GetVote returns the current vote for a game
func (s *VoteService) GetVote(gameID entities.GameID) (*entities.Vote, bool) {
	ctx := context.TODO()

	vote, err := s.voteRepo.GetVote(ctx, gameID)
	if err != nil || vote == nil {
		return nil, false
	}
	return vote, true
}

// ClearVote removes the vote for a game
func (s *VoteService) ClearVote(gameID entities.GameID) {
	ctx := context.TODO()
	s.voteRepo.DeleteVote(ctx, gameID)
}

// GetVoteCount returns the number of votes for each target
func (s *VoteService) GetVoteCount(gameID entities.GameID) map[entities.PlayerID]int {
	ctx := context.TODO()

	vote, err := s.voteRepo.GetVote(ctx, gameID)
	if err != nil || vote == nil {
		return nil
	}

	counts := make(map[entities.PlayerID]int)
	for _, target := range vote.Ballots {
		if target != nil {
			counts[*target]++
		}
	}
	return counts
}

// broadcastVoteEvent broadcasts a vote event to all players in a game
func (s *VoteService) broadcastVoteEvent(gameID entities.GameID, event entities.Event[events.VoteEventData]) {
	if s.broadcaster == nil {
		return
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return
	}

	s.broadcaster.BroadcastToGame(gameID, payload)
}

// broadcastToVoters sends a vote event only to eligible voters (e.g., werewolves)
func (s *VoteService) broadcastToVoters(vote *entities.Vote, event entities.Event[events.VoteEventData]) {
	if s.playerSender == nil {
		return
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return
	}

	for _, voterID := range vote.EligibleVoters {
		s.playerSender.SendToPlayer(voterID, payload)
	}
}

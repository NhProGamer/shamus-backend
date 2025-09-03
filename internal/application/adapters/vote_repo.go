package adapters

import (
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/ports"
)

type voteRepo struct {
	gameRepo   ports.GameRepository
	playerRepo ports.PlayerRepository
}

func NewVoteRepo(gr ports.GameRepository, pr ports.PlayerRepository) ports.GameVotesRepository {
	return &voteRepo{
		gameRepo:   gr,
		playerRepo: pr,
	}
}

func (v *voteRepo) GetVotes(gameID entities.GameID) (map[entities.PlayerID]*entities.PlayerID, error) {
	game, _ := v.gameRepo.GetGameByID(gameID)
	votes := make(map[entities.PlayerID]*entities.PlayerID)
	for _, playerID := range game.Players() {
		actualPlayer, _ := v.playerRepo.GetPlayerByID(playerID)
		votes[playerID] = actualPlayer.VotedFor()
	}
	return votes, nil
}
func (v *voteRepo) SetVote(gameID entities.GameID, playerID entities.PlayerID, target entities.PlayerID) error {
	actualPlayer, _ := v.playerRepo.GetPlayerByID(playerID)
	actualPlayer.VoteFor(&target)
	return nil
}
func (v *voteRepo) DeleteVote(gameID entities.GameID, playerID entities.PlayerID) error {
	actualPlayer, _ := v.playerRepo.GetPlayerByID(playerID)
	actualPlayer.VoteFor(nil)
	return nil
}
func (v *voteRepo) CleanVotes(gameID entities.GameID) error {
	game, _ := v.gameRepo.GetGameByID(gameID)
	for _, playerID := range game.Players() {
		actualPlayer, _ := v.playerRepo.GetPlayerByID(playerID)
		actualPlayer.VoteFor(nil)
	}
	return nil
}

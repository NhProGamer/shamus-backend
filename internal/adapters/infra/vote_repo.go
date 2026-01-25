package infra

import (
	"context"
	"fmt"
	"shamus-backend/internal/domain/constants"
	"shamus-backend/internal/domain/entities"

	"github.com/redis/go-redis/v9"
)

const voteKeyPrefix = "vote:"

// VoteRepository handles persistence of votes in Redis
type VoteRepository struct {
	rdb *redis.Client
}

// NewVoteRepository creates a new VoteRepository
func NewVoteRepository(rdb *redis.Client) *VoteRepository {
	return &VoteRepository{rdb: rdb}
}

// voteKey returns the Redis key for a vote
func voteKey(gameID entities.GameID) string {
	return fmt.Sprintf("%s%s", voteKeyPrefix, string(gameID))
}

// SaveVote persists a vote to Redis
func (r *VoteRepository) SaveVote(ctx context.Context, gameID entities.GameID, vote *entities.Vote) error {
	data, err := JSON.Marshal(vote)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, voteKey(gameID), data, constants.VoteTTL).Err()
}

// GetVote retrieves a vote from Redis
func (r *VoteRepository) GetVote(ctx context.Context, gameID entities.GameID) (*entities.Vote, error) {
	val, err := r.rdb.Get(ctx, voteKey(gameID)).Result()
	if err == redis.Nil {
		return nil, nil // Vote not found, not an error
	} else if err != nil {
		return nil, err
	}

	var vote entities.Vote
	if err := JSON.Unmarshal([]byte(val), &vote); err != nil {
		return nil, err
	}
	return &vote, nil
}

// DeleteVote removes a vote from Redis
func (r *VoteRepository) DeleteVote(ctx context.Context, gameID entities.GameID) error {
	return r.rdb.Del(ctx, voteKey(gameID)).Err()
}

// VoteExists checks if a vote exists for a game
func (r *VoteRepository) VoteExists(ctx context.Context, gameID entities.GameID) (bool, error) {
	exists, err := r.rdb.Exists(ctx, voteKey(gameID)).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

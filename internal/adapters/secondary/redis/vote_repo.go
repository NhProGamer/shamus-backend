package redis

import (
	"context"
	"fmt"
	"shamus-backend/internal/domain/constants"
	"shamus-backend/internal/domain/entities"

	"github.com/redis/go-redis/v9"
)

const voteKeyPrefix = "vote:"
const ballotsKeySuffix = ":ballots"

// VoteRepository handles persistence of votes in Redis
// Uses separate keys for vote metadata and ballots to enable atomic ballot casting
type VoteRepository struct {
	rdb *redis.Client
}

// NewVoteRepository creates a new VoteRepository
func NewVoteRepository(rdb *redis.Client) *VoteRepository {
	return &VoteRepository{rdb: rdb}
}

// voteKey returns the Redis key for vote metadata
func voteKey(gameID entities.GameID) string {
	return fmt.Sprintf("%s%s", voteKeyPrefix, string(gameID))
}

// ballotsKey returns the Redis key for the ballots hash
func ballotsKey(gameID entities.GameID) string {
	return fmt.Sprintf("%s%s%s", voteKeyPrefix, string(gameID), ballotsKeySuffix)
}

// SaveVote persists vote metadata to Redis (without ballots - they are stored separately)
func (r *VoteRepository) SaveVote(ctx context.Context, gameID entities.GameID, vote *entities.Vote) error {
	// Create a copy without ballots for storage (ballots are in separate hash)
	voteMetadata := *vote
	voteMetadata.Ballots = nil // Don't store ballots in the main vote object

	data, err := JSON.Marshal(voteMetadata)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, voteKey(gameID), data, constants.VoteTTL).Err()
}

// GetVote retrieves a vote from Redis, reconstructing ballots from the hash
func (r *VoteRepository) GetVote(ctx context.Context, gameID entities.GameID) (*entities.Vote, error) {
	// Get vote metadata
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

	// Initialize ballots map
	vote.Ballots = make(map[entities.PlayerID]*entities.PlayerID)

	// Get ballots from hash
	ballots, err := r.rdb.HGetAll(ctx, ballotsKey(gameID)).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	// Reconstruct ballots map
	for voterID, targetID := range ballots {
		voter := entities.PlayerID(voterID)
		if targetID == "" {
			// Abstention
			vote.Ballots[voter] = nil
		} else {
			target := entities.PlayerID(targetID)
			vote.Ballots[voter] = &target
		}
	}

	return &vote, nil
}

// CastBallot atomically records a single player's vote using HSET
// This is atomic and cannot conflict with other players' votes
func (r *VoteRepository) CastBallot(ctx context.Context, gameID entities.GameID, voterID entities.PlayerID, targetID *entities.PlayerID) error {
	key := ballotsKey(gameID)

	// Store targetID as string, empty string for abstention
	var value string
	if targetID != nil {
		value = string(*targetID)
	} else {
		value = "" // Abstention marker
	}

	// HSET is atomic for a single field
	if err := r.rdb.HSet(ctx, key, string(voterID), value).Err(); err != nil {
		return err
	}

	// Set TTL on the hash (refresh on each vote)
	return r.rdb.Expire(ctx, key, constants.VoteTTL).Err()
}

// GetBallot retrieves a single player's vote
func (r *VoteRepository) GetBallot(ctx context.Context, gameID entities.GameID, voterID entities.PlayerID) (*entities.PlayerID, bool, error) {
	val, err := r.rdb.HGet(ctx, ballotsKey(gameID), string(voterID)).Result()
	if err == redis.Nil {
		return nil, false, nil // No ballot cast
	} else if err != nil {
		return nil, false, err
	}

	// Empty string means abstention
	if val == "" {
		return nil, true, nil // Abstention
	}

	target := entities.PlayerID(val)
	return &target, true, nil
}

// GetBallotCount returns the number of ballots cast
func (r *VoteRepository) GetBallotCount(ctx context.Context, gameID entities.GameID) (int64, error) {
	return r.rdb.HLen(ctx, ballotsKey(gameID)).Result()
}

// DeleteVote removes a vote and its ballots from Redis
func (r *VoteRepository) DeleteVote(ctx context.Context, gameID entities.GameID) error {
	// Delete both keys in a pipeline
	pipe := r.rdb.Pipeline()
	pipe.Del(ctx, voteKey(gameID))
	pipe.Del(ctx, ballotsKey(gameID))
	_, err := pipe.Exec(ctx)
	return err
}

// VoteExists checks if a vote exists for a game
func (r *VoteRepository) VoteExists(ctx context.Context, gameID entities.GameID) (bool, error) {
	exists, err := r.rdb.Exists(ctx, voteKey(gameID)).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

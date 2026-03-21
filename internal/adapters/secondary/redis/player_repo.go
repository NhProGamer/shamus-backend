package redis

import (
	"context"
	"fmt"
	"shamus-backend/internal/domain/constants"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/factories"
	apperrors "shamus-backend/internal/domain/errors"
	"shamus-backend/internal/domain/ports"

	"github.com/redis/go-redis/v9"
)

const (
	playerKeyPrefix   = "player:"
	gamePlayersPrefix = "game:%s:player_ids"
)

type RedisPlayerRepo struct {
	rdb *redis.Client
}

func NewRedisPlayerRepo(rdb *redis.Client) ports.PlayerRepository {
	return &RedisPlayerRepo{
		rdb: rdb,
	}
}

// playerKey returns the Redis key for a player
func playerKey(id entities.PlayerID) string {
	return playerKeyPrefix + string(id)
}

// gamePlayersKey returns the Redis key for a game's player set
func gamePlayersKey(gameID entities.GameID) string {
	return fmt.Sprintf(gamePlayersPrefix, string(gameID))
}

// SavePlayer persists a player to Redis with configurable TTL
func (r *RedisPlayerRepo) SavePlayer(ctx context.Context, player *entities.Player) error {
	data, err := JSON.Marshal(player)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, playerKey(player.ID), data, constants.PlayerTTL).Err()
}

// SavePlayers persists multiple players in a single Redis pipeline
func (r *RedisPlayerRepo) SavePlayers(ctx context.Context, players []*entities.Player) error {
	if len(players) == 0 {
		return nil
	}

	pipe := r.rdb.Pipeline()

	for _, player := range players {
		data, err := JSON.Marshal(player)
		if err != nil {
			return err
		}
		pipe.Set(ctx, playerKey(player.ID), data, constants.PlayerTTL)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// GetPlayer retrieves a player from Redis by ID
func (r *RedisPlayerRepo) GetPlayer(ctx context.Context, id entities.PlayerID) (*entities.Player, error) {
	val, err := r.rdb.Get(ctx, playerKey(id)).Result()
	if err == redis.Nil {
		return nil, apperrors.ErrPlayerNotFound
	} else if err != nil {
		return nil, err
	}

	var player entities.Player
	err = JSON.Unmarshal([]byte(val), &player)
	if err != nil {
		return nil, err
	}

	// Reconstruct Role from RoleType
	if player.RoleType != nil {
		player.Role = factories.GetNewRole(*player.RoleType)
	}

	return &player, nil
}

// DeletePlayer removes a player from Redis
func (r *RedisPlayerRepo) DeletePlayer(ctx context.Context, id entities.PlayerID) error {
	return r.rdb.Del(ctx, playerKey(id)).Err()
}

// GetPlayersByGame retrieves all players in a game using the player set index
func (r *RedisPlayerRepo) GetPlayersByGame(ctx context.Context, gameID entities.GameID) ([]*entities.Player, error) {
	// Get all player IDs from the game's player set
	playerIDs, err := r.rdb.SMembers(ctx, gamePlayersKey(gameID)).Result()
	if err != nil {
		return nil, err
	}

	if len(playerIDs) == 0 {
		return []*entities.Player{}, nil
	}

	// Build keys for MGET
	keys := make([]string, len(playerIDs))
	for i, id := range playerIDs {
		keys[i] = playerKey(entities.PlayerID(id))
	}

	// Get all players in one round trip
	vals, err := r.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	players := make([]*entities.Player, 0, len(vals))
	for _, val := range vals {
		if val == nil {
			continue // Player was deleted but still in set (edge case)
		}
		var player entities.Player
		if err := JSON.Unmarshal([]byte(val.(string)), &player); err != nil {
			continue // Skip invalid data
		}
		// Reconstruct Role from RoleType
		if player.RoleType != nil {
			player.Role = factories.GetNewRole(*player.RoleType)
		}
		players = append(players, &player)
	}

	return players, nil
}

// AddPlayerToGame adds a player ID to the game's player set
func (r *RedisPlayerRepo) AddPlayerToGame(ctx context.Context, gameID entities.GameID, playerID entities.PlayerID) error {
	key := gamePlayersKey(gameID)

	// Add to set
	if err := r.rdb.SAdd(ctx, key, string(playerID)).Err(); err != nil {
		return err
	}

	// Set TTL on the set (same as player TTL)
	return r.rdb.Expire(ctx, key, constants.PlayerTTL).Err()
}

// RemovePlayerFromGame removes a player ID from the game's player set
func (r *RedisPlayerRepo) RemovePlayerFromGame(ctx context.Context, gameID entities.GameID, playerID entities.PlayerID) error {
	return r.rdb.SRem(ctx, gamePlayersKey(gameID), string(playerID)).Err()
}

// DeleteGamePlayers removes all players associated with a game (cleanup when game ends)
func (r *RedisPlayerRepo) DeleteGamePlayers(ctx context.Context, gameID entities.GameID) error {
	key := gamePlayersKey(gameID)

	// Get all player IDs first
	playerIDs, err := r.rdb.SMembers(ctx, key).Result()
	if err != nil {
		return err
	}

	if len(playerIDs) == 0 {
		return nil
	}

	// Build keys to delete
	keysToDelete := make([]string, len(playerIDs)+1)
	for i, id := range playerIDs {
		keysToDelete[i] = playerKey(entities.PlayerID(id))
	}
	// Also delete the set itself
	keysToDelete[len(playerIDs)] = key

	// Delete all keys in one operation
	return r.rdb.Del(ctx, keysToDelete...).Err()
}

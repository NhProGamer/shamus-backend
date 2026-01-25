package infra

import (
	"context"
	"shamus-backend/internal/domain/constants"
	"shamus-backend/internal/domain/entities"
	apperrors "shamus-backend/internal/domain/errors"
	"shamus-backend/internal/domain/ports"

	"github.com/redis/go-redis/v9"
)

type RedisGameRepo struct {
	rdb *redis.Client
}

func NewRedisGameRepo(rdb *redis.Client) ports.GameRepository {
	return &RedisGameRepo{
		rdb: rdb,
	}
}

// SaveGame persists a game to Redis with configurable TTL
func (gm *RedisGameRepo) SaveGame(ctx context.Context, game *entities.Game) error {
	data, err := JSON.Marshal(game)
	if err != nil {
		return err
	}
	return gm.rdb.Set(ctx, "game:"+string(game.ID), data, constants.GameTTL).Err()
}

// GetGame retrieves a game from Redis by ID
func (gm *RedisGameRepo) GetGame(ctx context.Context, id entities.GameID) (*entities.Game, error) {
	val, err := gm.rdb.Get(ctx, "game:"+string(id)).Result()
	if err == redis.Nil {
		return nil, apperrors.ErrGameNotFound
	} else if err != nil {
		return nil, err
	}

	var game entities.Game
	err = JSON.Unmarshal([]byte(val), &game)
	if err != nil {
		return nil, err
	}
	return &game, nil
}

// DeleteGame removes a game from Redis
func (gm *RedisGameRepo) DeleteGame(ctx context.Context, id entities.GameID) error {
	return gm.rdb.Del(ctx, "game:"+string(id)).Err()
}

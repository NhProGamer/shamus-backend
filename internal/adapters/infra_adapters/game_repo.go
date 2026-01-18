package infra_adapters

import (
	"context"
	"encoding/json"
	"errors"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/ports"
	"time"

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

// Sauvegarder (Create ou Update)
func (gm *RedisGameRepo) SaveGame(game *entities.Game) error {
	data, err := json.Marshal(game)
	if err != nil {
		return err
	}
	// Expire après 24h d'inactivité
	return gm.rdb.Set(context.Background(), "game:"+string(game.ID), data, 24*time.Hour).Err()
}

// Récupérer
func (gm *RedisGameRepo) GetGame(id entities.GameID) (*entities.Game, error) {
	val, err := gm.rdb.Get(context.Background(), "game:"+string(id)).Result()
	if err == redis.Nil {
		return nil, errors.New("game not found")
	} else if err != nil {
		return nil, err
	}

	var game entities.Game
	err = json.Unmarshal([]byte(val), &game)
	if err != nil {
		return nil, err
	}
	return &game, nil
}

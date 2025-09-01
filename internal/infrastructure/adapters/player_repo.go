package adapters

import (
	"shamus-backend/internal/domain/entities"
	domain_errors "shamus-backend/internal/domain/errors"
	"sync"
)

type PlayerRepositoryImpl struct {
	mu      sync.RWMutex
	players map[entities.PlayerID]*entities.SafePlayer
}

func NewPlayerRepository() *PlayerRepositoryImpl {
	return &PlayerRepositoryImpl{
		players: make(map[entities.PlayerID]*entities.SafePlayer),
	}
}

// GetPlayerByID retourne le joueur
func (r *PlayerRepositoryImpl) GetPlayerByID(id entities.PlayerID) (*entities.SafePlayer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	player, exists := r.players[id]
	if !exists {
		return nil, domain_errors.ErrPlayerNotFound
	}
	return player, nil
}

// GetPlayerGameID retourne l'ID du jeu auquel appartient le joueur
func (r *PlayerRepositoryImpl) GetPlayerGameID(id entities.PlayerID) (*entities.GameID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	player, exists := r.players[id]
	if !exists {
		return nil, domain_errors.ErrPlayerNotFound
	}
	if player.GetGameID() == nil {
		return nil, domain_errors.ErrPlayerHasNoGame
	}
	return player.GetGameID(), nil
}

// AddPlayer ajoute un joueur
func (r *PlayerRepositoryImpl) AddPlayer(player *entities.SafePlayer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.players[player.ID()]; exists {
		return domain_errors.ErrPlayerAlreadyExists
	}
	r.players[player.ID()] = player
	return nil
}

// DeletePlayer supprime un joueur
func (r *PlayerRepositoryImpl) DeletePlayer(id entities.PlayerID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.players[id]; !exists {
		return domain_errors.ErrPlayerNotFound
	}
	delete(r.players, id)
	return nil
}

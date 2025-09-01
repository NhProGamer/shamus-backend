package adapters

import (
	"log"
	"shamus-backend/internal/domain/entities"
	domain_errors "shamus-backend/internal/domain/errors"
	"shamus-backend/internal/domain/ports"
	"shamus-backend/internal/infrastructure/ws"
	"sync"
)

type GameRepositoryImpl struct {
	mu         sync.RWMutex
	games      map[entities.GameID]*entities.Game
	playerRepo ports.PlayerRepository
	hub        *ws.Hub
}

func NewGameRepository(pr ports.PlayerRepository, h *ws.Hub) *GameRepositoryImpl {
	return &GameRepositoryImpl{
		mu:         sync.RWMutex{},
		games:      make(map[entities.GameID]*entities.Game),
		playerRepo: pr,
		hub:        h,
	}
}

func (r *GameRepositoryImpl) GetGameByID(id entities.GameID) (*entities.Game, error) {
	r.mu.RLock() // lecture safe de la map
	defer r.mu.RUnlock()
	game, exists := r.games[id]
	if !exists {
		return nil, domain_errors.ErrGameNotFound
	}
	return game, nil // game peut ensuite être manipulée par la goroutine qui l'a récupérée
}

func (r *GameRepositoryImpl) CreateGame(game *entities.Game) error {
	r.mu.Lock() // écriture safe de la map
	defer r.mu.Unlock()
	if _, exists := r.games[game.ID]; exists {
		return domain_errors.ErrGameAlreadyExists
	}
	r.games[game.ID] = game
	return nil
}

func (r *GameRepositoryImpl) DeleteGame(id entities.GameID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, playerID := range r.games[id].Players {
		if existingClient, exists := r.hub.Get(playerID); exists {
			log.Printf("fermeture de l'ancienne connexion WebSocket pour le joueur %s", playerID)
			r.hub.Unregister(playerID)
			existingClient.Conn.Close()
		}
		r.playerRepo.DeletePlayer(playerID)
	}
	delete(r.games, id)
	return nil
}

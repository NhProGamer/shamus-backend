package app_adapters

import (
	"errors"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/ports"

	"github.com/google/uuid"
)

type GameService struct {
	repo ports.GameRepository
}

func NewGameService(repo ports.GameRepository) *GameService {
	return &GameService{repo: repo}
}

// CreateNewGame : Logique de création pure
func (s *GameService) CreateNewGame(hostID entities.PlayerID) (*entities.Game, error) {
	// Génération ID (4 chars)
	gameID := entities.GameID(uuid.New().String())

	newGame := &entities.Game{
		ID:      gameID,
		Status:  entities.GameStatusWaiting,
		Phase:   entities.PhaseStart,
		Day:     0,
		Players: []entities.PlayerID{},
		HostID:  hostID,
		Settings: entities.GameSettings{
			Roles: make(map[entities.RoleType]int),
		},
	}

	if err := s.repo.SaveGame(newGame); err != nil {
		return nil, err
	}

	return newGame, nil
}

// JoinGame : Logique pour rejoindre (vérifications métier)
func (s *GameService) JoinGame(gameID entities.GameID, playerID entities.PlayerID) (*entities.Game, error) {
	game, err := s.repo.GetGame(gameID)
	if err != nil {
		return nil, err
	}

	// Règles métier : Impossible de rejoindre une partie finie
	if game.Status == entities.GameStatusEnded {
		return nil, errors.New("game ended")
	}

	// Vérifier si déjà présent pour éviter les doublons
	for _, p := range game.Players {
		if p == playerID {
			return game, nil // Déjà dedans, on renvoie l'état actuel
		}
	}

	// Ajouter le joueur
	game.Players = append(game.Players, playerID)

	// Sauvegarder l'état mis à jour
	if err := s.repo.SaveGame(game); err != nil {
		return nil, err
	}

	return game, nil
}

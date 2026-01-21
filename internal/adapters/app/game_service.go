package app

import (
	"errors"
	"shamus-backend/internal/domain/entities"
	apperrors "shamus-backend/internal/domain/errors"
	"shamus-backend/internal/domain/ports"

	"github.com/google/uuid"
)

type GameService struct {
	repo ports.GameRepository
}

func NewGameService(repo ports.GameRepository) *GameService {
	return &GameService{repo: repo}
}

// CreateNewGame creates a new game with the given host
func (s *GameService) CreateNewGame(hostID entities.PlayerID) (*entities.Game, error) {
	gameID := entities.GameID(uuid.New().String())

	newGame := &entities.Game{
		ID:      gameID,
		Status:  entities.GameStatusWaiting,
		Phase:   entities.PhaseStart,
		Day:     0,
		Players: []entities.PlayerID{},
		HostID:  hostID,
		Settings: entities.GameSettings{
			Roles: map[entities.RoleType]int{
				entities.RoleVillager: 4,
				entities.RoleWerewolf: 2,
				entities.RoleSeer:     1,
				entities.RoleWitch:    1,
			},
		},
	}

	if err := s.repo.SaveGame(newGame); err != nil {
		return nil, err
	}

	return newGame, nil
}

// JoinGame adds a player to an existing game
func (s *GameService) JoinGame(gameID entities.GameID, playerID entities.PlayerID) (*entities.Game, error) {
	game, err := s.repo.GetGame(gameID)
	if err != nil {
		return nil, err
	}

	if game.Status == entities.GameStatusEnded {
		return nil, errors.New("game ended")
	}

	// Check if already present to avoid duplicates
	for _, p := range game.Players {
		if p == playerID {
			return game, nil
		}
	}

	game.Players = append(game.Players, playerID)

	if err := s.repo.SaveGame(game); err != nil {
		return nil, err
	}

	return game, nil
}

// GetGame retrieves a game by ID
func (s *GameService) GetGame(gameID entities.GameID) (*entities.Game, error) {
	return s.repo.GetGame(gameID)
}

// UpdateSettings updates game settings (roles configuration)
// Only the host can update settings, and only when game is in waiting state
func (s *GameService) UpdateSettings(gameID entities.GameID, playerID entities.PlayerID, settings entities.GameSettings) (*entities.Game, error) {
	game, err := s.repo.GetGame(gameID)
	if err != nil {
		return nil, apperrors.ErrGameNotFound
	}

	// Only host can update settings
	if game.HostID != playerID {
		return nil, apperrors.ErrNotHost
	}

	// Game must be in waiting state
	if game.Status != entities.GameStatusWaiting {
		return nil, apperrors.ErrGameNotWaiting
	}

	// Validate each role
	totalRoles := 0
	for roleType, count := range settings.Roles {
		if !entities.IsValidRole(string(roleType)) {
			return nil, apperrors.ErrInvalidRole
		}
		if count < 0 {
			return nil, apperrors.ErrInvalidInput
		}

		// Check role-specific limits (e.g., max 1 seer, max 1 witch)
		if limit, hasLimit := entities.RoleLimits[roleType]; hasLimit && count > limit {
			return nil, apperrors.ErrRoleLimitExceeded
		}

		totalRoles += count
	}

	// Check total roles doesn't exceed max players
	if totalRoles > entities.MaxPlayers {
		return nil, apperrors.ErrTooManyRoles
	}

	// Update settings
	game.Settings = settings

	if err := s.repo.SaveGame(game); err != nil {
		return nil, err
	}

	return game, nil
}

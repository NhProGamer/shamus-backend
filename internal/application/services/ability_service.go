package services

import (
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/ports"
)

type abilityService struct {
	gameRepo   ports.GameRepository
	playerRepo ports.PlayerRepository
}

func NewAbilityService(
	gr ports.GameRepository,
	pr ports.PlayerRepository,
) ports.AbilityService {
	return &abilityService{
		gameRepo:   gr,
		playerRepo: pr,
	}
}

func (s *abilityService) UseAbility(gameID entities.GameID, playerID entities.PlayerID, target entities.PlayerID) error {
	game, err := s.gameRepo.GetGameByID(gameID)
	if err != nil {
		return err
	}

	player, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		return err
	}

	/*targetPlayer, err := s.playerRepo.GetPlayerByID(target)
	if err != nil {
		return err
	}*/

	// Trouver l'abilité à utiliser (dépend du rôle et du contexte)
	for _, ability := range *player.Role.GetAbilities() {
		if ability.CanUse(game, player) {
			if err := ability.Execute(game, player, &target, nil); err != nil {
				return err
			}
			ability.Consume()
			break
		}
	}

	return s.gameRepo.UpdateGame(game)
}

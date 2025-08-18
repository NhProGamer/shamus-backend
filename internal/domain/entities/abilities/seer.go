package abilities

import "shamus-backend/internal/domain/entities"

type SeeAbility struct{}

func (s *SeeAbility) GetName() string {
	return "Voir"
}

func (s *SeeAbility) GetDescription() string {
	return "Révèle le rôle d'un joueur"
}

func (s *SeeAbility) CanUse(game *entities.Game, player *entities.Player) bool {
	return true
}

func (s *SeeAbility) GetConsumptions() *uint8 {
	return nil
}
func (s *SeeAbility) Consume() {
	// No consumptions for this ability
}

func (s *SeeAbility) Execute(game *entities.Game, player *entities.Player, target *entities.PlayerID, data map[string]interface{}) error {
	return nil
}

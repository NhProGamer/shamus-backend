package abilities

import "shamus-backend/internal/domain/entities"

type SeeAbility struct{}

func (s *SeeAbility) GetName() string {
	return "See"
}

func (s *SeeAbility) GetDescription() string {
	return "Reveals a player's role"
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
	// TODO: implement see logic
	return nil
}

func NewSeeAbility() *SeeAbility {
	return &SeeAbility{}
}

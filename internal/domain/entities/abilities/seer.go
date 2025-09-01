package abilities

import "shamus-backend/internal/domain/entities"

type SeeAbility struct{}

func (s *SeeAbility) GetName() string {
	return "Voir"
}

func (s *SeeAbility) GetDescription() string {
	return "Révèle le rôle d'un joueur"
}

func (s *SeeAbility) CanUse(game *entities.Game, player *entities.SafePlayer) bool {
	return true
	// TODO: Ajouter les erreurs ici
}

func (s *SeeAbility) GetConsumptions() *uint8 {
	return nil
}
func (s *SeeAbility) Consume() {
	// No consumptions for this ability
}

func (s *SeeAbility) Execute(game *entities.Game, player *entities.SafePlayer, target *entities.PlayerID, data map[string]interface{}) error {
	return nil
}

func NewSeeAbility() *SeeAbility {
	return &SeeAbility{}
}

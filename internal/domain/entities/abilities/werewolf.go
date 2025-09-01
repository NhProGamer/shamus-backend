package abilities

import (
	"shamus-backend/internal/domain/entities"
)

type WerewolfKillAbility struct{}

func (k *WerewolfKillAbility) GetName() string {
	return "Tuer"
}

func (k *WerewolfKillAbility) GetDescription() string {
	return "Élimine un joueur"
}

func (k *WerewolfKillAbility) CanUse(game *entities.Game, player *entities.SafePlayer) bool {
	// TODO: Ajouter les erreurs ici
	return false
}

func (k *WerewolfKillAbility) GetConsumptions() *uint8 {
	return nil
}

func (k *WerewolfKillAbility) Consume() {
	// No consumptions for this ability
}

func (k *WerewolfKillAbility) Execute(game *entities.Game, player *entities.SafePlayer, target *entities.PlayerID, data map[string]interface{}) error {
	return nil
}

func NewWerewolfKillAbility() entities.Ability {
	return &WerewolfKillAbility{}
}

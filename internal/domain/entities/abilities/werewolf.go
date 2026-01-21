package abilities

import (
	"shamus-backend/internal/domain/entities"
)

type WerewolfKillAbility struct{}

func (k *WerewolfKillAbility) GetName() string {
	return "Kill"
}

func (k *WerewolfKillAbility) GetDescription() string {
	return "Eliminates a player"
}

func (k *WerewolfKillAbility) CanUse(game *entities.Game, player *entities.Player) bool {
	// TODO: add validation logic
	return true
}

func (k *WerewolfKillAbility) GetConsumptions() *uint8 {
	return nil
}

func (k *WerewolfKillAbility) Consume() {
	// No consumptions for this ability
}

func (k *WerewolfKillAbility) Execute(game *entities.Game, player *entities.Player, target *entities.PlayerID, data map[string]interface{}) error {
	// TODO: implement kill logic
	return nil
}

func NewWerewolfKillAbility() entities.Ability {
	return &WerewolfKillAbility{}
}

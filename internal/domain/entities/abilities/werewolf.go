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
	return true
}

func (k *WerewolfKillAbility) GetConsumptions() *uint8 {
	return nil
}

func (k *WerewolfKillAbility) Consume() {
	// No consumptions for this ability
}

func NewWerewolfKillAbility() entities.Ability {
	return &WerewolfKillAbility{}
}

package abilities

import (
	"shamus-backend/internal/domain/entities"
)

type HealAbility struct {
	Consumptions *uint8
}

func (h *HealAbility) GetName() string        { return "Guérir" }
func (h *HealAbility) GetDescription() string { return "Sauve un joueur de la mort" }
func (h *HealAbility) CanUse(game *entities.Game, player *entities.Player) bool {
	return h.Consumptions != nil && *h.Consumptions != 0 && game.Phase == entities.PhaseNight
}
func (h *HealAbility) GetConsumptions() *uint8 {
	return h.Consumptions
}
func (h *HealAbility) Consume() {
	if h.Consumptions != nil {
		*h.Consumptions -= 1
	}
}
func (h *HealAbility) Execute(game *entities.Game, player *entities.Player, target *entities.PlayerID, data map[string]interface{}) error {
	// Check erreurs etc...
	return nil
}

type PoisonAbility struct {
	Consumptions *uint8
}

func (p *PoisonAbility) GetName() string        { return "Empoisonner" }
func (p *PoisonAbility) GetDescription() string { return "Empoisonne un joueur" }
func (p *PoisonAbility) CanUse(game *entities.Game, player *entities.Player) bool {
	return p.Consumptions != nil && *p.Consumptions != 0 && game.Phase == entities.PhaseNight
}
func (p *PoisonAbility) Execute(game *entities.Game, player *entities.Player, target *entities.PlayerID, data map[string]interface{}) error {
	// Check erreurs etc...
	return nil
}

package abilities

import (
	"shamus-backend/internal/domain/entities"
)

type HealAbility struct {
	consumptions *uint8
}

func (h *HealAbility) GetName() string        { return "Guérir" }
func (h *HealAbility) GetDescription() string { return "Sauve un joueur de la mort" }
func (h *HealAbility) CanUse(game *entities.Game, player *entities.Player) bool {
	return h.consumptions != nil && *h.consumptions != 0
}
func (h *HealAbility) GetConsumptions() *uint8 {
	return h.consumptions
}
func (h *HealAbility) Consume() {
	if h.consumptions != nil {
		*h.consumptions -= 1
	}
}
func (h *HealAbility) Execute(game *entities.Game, player *entities.Player, target *entities.PlayerID, data map[string]interface{}) error {
	// Check erreurs etc...
	return nil
}

type PoisonAbility struct {
	consumptions *uint8
}

func (p *PoisonAbility) GetName() string        { return "Empoisonner" }
func (p *PoisonAbility) GetDescription() string { return "Empoisonne un joueur" }
func (p *PoisonAbility) CanUse(game *entities.Game, player *entities.SafePlayer) bool {
	return p.consumptions != nil && *p.consumptions != 0
	// TODO: Ajouter les erreurs ici
}
func (p *PoisonAbility) GetConsumptions() *uint8 {
	return p.consumptions
}
func (p *PoisonAbility) Consume() {
	if p.consumptions != nil {
		*p.consumptions -= 1
	}
}
func (p *PoisonAbility) Execute(game *entities.Game, player *entities.SafePlayer, target *entities.PlayerID, data map[string]interface{}) error {
	// Check erreurs etc...
	return nil
}

func NewHealAbility() *HealAbility {
	return &HealAbility{
		consumptions: func(v uint8) *uint8 { return &v }(1),
	}
}

func NewPoisonAbility() *PoisonAbility {
	return &PoisonAbility{
		consumptions: func(v uint8) *uint8 { return &v }(1),
	}
}

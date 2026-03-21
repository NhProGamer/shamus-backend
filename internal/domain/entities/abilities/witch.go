package abilities

import (
	"shamus-backend/internal/domain/entities"
)

type HealAbility struct {
	consumptions *uint8
}

func (h *HealAbility) GetName() string        { return "Heal" }
func (h *HealAbility) GetDescription() string { return "Saves a player from death" }
func (h *HealAbility) CanUse(game *entities.Game, player *entities.Player) bool {
	return h.consumptions != nil && *h.consumptions != 0
}
func (h *HealAbility) GetConsumptions() *uint8 {
	return h.consumptions
}
func (h *HealAbility) TryConsume() bool {
	if h.consumptions == nil {
		return true // Unlimited
	}
	if *h.consumptions == 0 {
		return false
	}
	*h.consumptions--
	return true
}

type PoisonAbility struct {
	consumptions *uint8
}

func (p *PoisonAbility) GetName() string        { return "Poison" }
func (p *PoisonAbility) GetDescription() string { return "Poisons a player" }
func (p *PoisonAbility) CanUse(game *entities.Game, player *entities.Player) bool {
	return p.consumptions != nil && *p.consumptions != 0
}
func (p *PoisonAbility) GetConsumptions() *uint8 {
	return p.consumptions
}
func (p *PoisonAbility) TryConsume() bool {
	if p.consumptions == nil {
		return true // Unlimited
	}
	if *p.consumptions == 0 {
		return false
	}
	*p.consumptions--
	return true
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

package roles

import (
	"shamus-backend/internal/domain/entities"
)

type WitchRole struct {
	Clans     []entities.Clan
	Abilities []entities.Ability
}

func (w *WitchRole) GetType() entities.RoleType { return entities.RoleWitch }
func (w *WitchRole) GetName() string            { return "Sorcière" }
func (w *WitchRole) GetDescription() string {
	return "Possède une potion de vie et une potion de mort"
}
func (w *WitchRole) CanVote() bool             { return true }
func (w *WitchRole) CanUseAbility() bool       { return true }
func (w *WitchRole) GetClans() []entities.Clan { return w.Clans }
func (w *WitchRole) GetPriority() entities.Priority {
	sumConsumptions := entities.SumConsumptions(w.Abilities)
	if sumConsumptions != nil && *sumConsumptions != 0 {
		return 8
	} else {
		return 0
	}
}
func (w *WitchRole) GetAbilities() *[]entities.Ability {
	return &w.Abilities
}

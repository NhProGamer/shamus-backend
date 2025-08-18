package roles

import (
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/abilities"
)

type WitchRole struct {
	HasHealPotion bool
	HasKillPotion bool
}

func (w *WitchRole) GetType() entities.RoleType { return entities.RoleWitch }
func (w *WitchRole) GetName() string            { return "Sorcière" }
func (w *WitchRole) GetDescription() string {
	return "Possède une potion de vie et une potion de mort"
}
func (w *WitchRole) CanVote() bool          { return true }
func (w *WitchRole) CanUseAbility() bool    { return true }
func (w *WitchRole) GetClan() entities.Clan { return entities.ClanVillager }
func (w *WitchRole) GetPriority() entities.Priority {
	if w.HasKillPotion || w.HasHealPotion {
		return 8
	} else {
		return 0
	}
}
func (w *WitchRole) GetAbilities() []entities.Ability {
	abilityList := []entities.Ability{}
	if w.HasHealPotion {
		abilityList = append(abilityList, &abilities.HealAbility{})
	}
	if w.HasKillPotion {
		abilityList = append(abilityList, &abilities.PoisonAbility{})
	}
	return abilityList
}

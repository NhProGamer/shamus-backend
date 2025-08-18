package roles

import (
	"shamus-backend/internal/domain/entities"
)

type WerewolfRole struct {
	Clans     []entities.Clan
	Abilities []entities.Ability
}

func (w *WerewolfRole) GetType() entities.RoleType { return entities.RoleWerewolf }
func (w *WerewolfRole) GetName() string            { return "Loup-Garou" }
func (w *WerewolfRole) GetDescription() string     { return "Élimine un villageois chaque nuit" }
func (w *WerewolfRole) CanVote() bool              { return true }
func (w *WerewolfRole) CanUseAbility() bool        { return true }
func (w *WerewolfRole) GetClans() []entities.Clan  { return w.Clans }
func (w *WerewolfRole) AddClan(clan entities.Clan) {
	w.Clans = append(w.Clans, clan)
}
func (w *WerewolfRole) GetPriority() entities.Priority { return 9 }
func (w *WerewolfRole) GetAbilities() *[]entities.Ability {
	return &w.Abilities
}

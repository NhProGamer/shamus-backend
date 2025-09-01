package roles

import (
	"shamus-backend/internal/domain/entities"
)

type VillagerRole struct {
	Clans     []entities.Clan
	Abilities []entities.Ability
}

func (v *VillagerRole) GetType() entities.RoleType { return entities.RoleVillager }
func (v *VillagerRole) GetName() string            { return "Villageois" }
func (v *VillagerRole) GetDescription() string     { return "Vote pour éliminer les loups-garous" }
func (v *VillagerRole) GetClans() []entities.Clan  { return v.Clans }
func (v *VillagerRole) AddClan(clan entities.Clan) {
	v.Clans = append(v.Clans, clan)
}
func (v *VillagerRole) GetPriority() entities.Priority    { return 0 }
func (v *VillagerRole) GetAbilities() *[]entities.Ability { return &v.Abilities }

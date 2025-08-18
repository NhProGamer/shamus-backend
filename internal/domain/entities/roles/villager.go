package roles

import (
	"shamus-backend/internal/domain/entities"
)

type VillagerRole struct {
	Clans []entities.Clan
}

func (v *VillagerRole) GetType() entities.RoleType       { return entities.RoleVillager }
func (v *VillagerRole) GetName() string                  { return "Villageois" }
func (v *VillagerRole) GetDescription() string           { return "Vote pour éliminer les loups-garous" }
func (v *VillagerRole) CanVote() bool                    { return true }
func (v *VillagerRole) CanUseAbility() bool              { return false }
func (v *VillagerRole) GetClans() []entities.Clan        { return v.Clans }
func (v *VillagerRole) GetPriority() entities.Priority   { return 0 }
func (v *VillagerRole) GetAbilities() []entities.Ability { return []entities.Ability{} }

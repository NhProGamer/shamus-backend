package roles

import (
	"shamus-backend/internal/domain/entities"
)

type VillagerRole struct{}

func (v *VillagerRole) GetType() entities.RoleType       { return entities.RoleVillager }
func (v *VillagerRole) GetName() string                  { return "Villageois" }
func (v *VillagerRole) GetDescription() string           { return "Vote pour éliminer les loups-garous" }
func (v *VillagerRole) CanVote() bool                    { return true }
func (v *VillagerRole) CanUseAbility() bool              { return false }
func (v *VillagerRole) GetClan() entities.Clan           { return entities.ClanVillager }
func (v *VillagerRole) GetPriority() entities.Priority   { return 0 }
func (v *VillagerRole) GetAbilities() []entities.Ability { return []entities.Ability{} }

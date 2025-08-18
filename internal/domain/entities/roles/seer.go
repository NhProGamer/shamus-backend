package roles

import (
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/abilities"
)

type SeerRole struct{}

func (s *SeerRole) GetType() entities.RoleType     { return entities.RoleSeer }
func (s *SeerRole) GetName() string                { return "Voyante" }
func (s *SeerRole) GetDescription() string         { return "Peut voir le rôle d'un joueur chaque nuit" }
func (s *SeerRole) CanVote() bool                  { return true }
func (s *SeerRole) CanUseAbility() bool            { return true }
func (s *SeerRole) GetClan() entities.Clan         { return entities.ClanVillager }
func (s *SeerRole) GetPriority() entities.Priority { return 10 }
func (s *SeerRole) GetAbilities() []entities.Ability {
	return []entities.Ability{&abilities.SeeAbility{}}
}

package roles

import (
	"shamus-backend/internal/domain/entities"
)

type SeerRole struct {
	Clans     []entities.Clan
	Abilities []entities.Ability
}

func (s *SeerRole) GetType() entities.RoleType { return entities.RoleSeer }
func (s *SeerRole) GetName() string            { return "Voyante" }
func (s *SeerRole) GetDescription() string     { return "Peut voir le rôle d'un joueur chaque nuit" }
func (s *SeerRole) CanVote() bool              { return true }
func (s *SeerRole) CanUseAbility() bool        { return true }
func (s *SeerRole) GetClans() []entities.Clan  { return s.Clans }
func (s *SeerRole) AddClan(clan entities.Clan) {
	s.Clans = append(s.Clans, clan)
}
func (s *SeerRole) GetPriority() entities.Priority { return 10 }
func (s *SeerRole) GetAbilities() *[]entities.Ability {
	return &s.Abilities
}

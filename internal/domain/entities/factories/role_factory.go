package factories

import (
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/abilities"
	"shamus-backend/internal/domain/entities/roles"
)

func GetNewRole(roleType entities.RoleType) entities.Role {
	switch roleType {
	case entities.RoleSeer:
		return NewSeerRole()
	case entities.RoleVillager:
		return NewVillagerRole()
	case entities.RoleWerewolf:
		return NewWerewolfRole()
	case entities.RoleWitch:
		return NewWitchRole()
	default:
		return NewVillagerRole()
	}
}

func NewSeerRole() *roles.SeerRole {
	return &roles.SeerRole{
		Clans: []entities.Clan{
			entities.ClanVillager,
		},
		Abilities: []entities.Ability{
			abilities.NewSeeAbility(),
		},
	}
}

func NewVillagerRole() *roles.VillagerRole {
	return &roles.VillagerRole{
		Clans: []entities.Clan{
			entities.ClanVillager,
		},
		Abilities: []entities.Ability{},
	}
}

func NewWerewolfRole() *roles.WerewolfRole {
	return &roles.WerewolfRole{
		Clans: []entities.Clan{
			entities.ClanVillager,
		},
		Abilities: []entities.Ability{
			abilities.NewWerewolfKillAbility(),
		},
	}
}

func NewWitchRole() *roles.WitchRole {
	return &roles.WitchRole{
		Clans: []entities.Clan{
			entities.ClanVillager,
		},
		Abilities: []entities.Ability{
			abilities.NewPoisonAbility(),
			abilities.NewPoisonAbility(),
		},
	}
}

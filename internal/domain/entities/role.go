package entities

type RoleType string
type Clan string
type Priority uint8

const (
	RoleSeer     RoleType = "seer"
	RoleVillager RoleType = "villager"
	RoleWerewolf RoleType = "werewolf"
	RoleWitch    RoleType = "witch"
)

func IsValidRole(role string) bool {
	switch RoleType(role) {
	case RoleSeer, RoleVillager, RoleWerewolf, RoleWitch:
		return true
	default:
		return false
	}
}

const (
	ClanVillager Clan = "villager"
	ClanWerewolf Clan = "werewolf"
	ClanRogue    Clan = "rogue"
)

func GetClan(role RoleType) Clan {
	switch role {
	case RoleSeer, RoleVillager, RoleWitch:
		return ClanVillager
	case RoleWerewolf:
		return ClanWerewolf
	default:
		return ClanRogue // rôle inconnu → par défaut rogue
	}
}

// GetClansFromRoles retourne la liste des clans correspondant à une liste de rôles.
func GetClansFromRoles(roles []RoleType) []Clan {
	clans := make([]Clan, 0, len(roles))
	for _, role := range roles {
		clans = append(clans, GetClan(role))
	}
	return clans
}

func HasRequiredClans(clans []Clan) bool {
	hasVillager := false
	hasWerewolfOrRogue := false

	for _, clan := range clans {
		if clan == ClanVillager {
			hasVillager = true
		}
		if clan == ClanWerewolf || clan == ClanRogue {
			hasWerewolfOrRogue = true
		}
	}
	return hasVillager && hasWerewolfOrRogue
}

type Role interface {
	GetType() RoleType
	GetName() string
	GetDescription() string
	GetClans() []Clan
	AddClan(clan Clan)
	GetPriority() Priority
	GetAbilities() *[]Ability
}

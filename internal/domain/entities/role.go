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

const (
	ClanVillager Clan = "villager"
	ClanWerewolf Clan = "werewolf"
	ClanRogue    Clan = "rogue"
)

type Role interface {
	GetType() RoleType
	GetName() string
	GetDescription() string
	GetClans() []Clan
	AddClan(clan Clan)
	GetPriority() Priority
	GetAbilities() *[]Ability
}

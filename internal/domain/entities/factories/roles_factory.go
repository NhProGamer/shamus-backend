package factories

import (
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/roles"
)

// RoleFactory interface définit les méthodes communes pour toutes les factories de rôles
type RoleFactory interface {
	CreateRole() entities.Role
}

// SeerRoleFactory pour créer le rôle de voyant
type SeerRoleFactory struct {
	AbilityFactory *SeerAbilityFactory
}

func (f *SeerRoleFactory) CreateRole() entities.Role {
	return &roles.SeerRole{
		Clans: []entities.Clan{entities.ClanVillager},
	}
}

// WerewolfRoleFactory pour créer le rôle de loup-garou
type WerewolfRoleFactory struct {
	AbilityFactory *WerewolfAbilityFactory
}

func (f *WerewolfRoleFactory) CreateRole() entities.Role {
	return &roles.WerewolfRole{
		Clans: []entities.Clan{entities.ClanWerewolf},
	}
}

// WitchRoleFactory pour créer le rôle de sorcière
type WitchRoleFactory struct {
	AbilityFactory *WitchAbilityFactory
}

func (f *WitchRoleFactory) CreateRole() entities.Role {
	abilities := f.AbilityFactory.CreateAbilities()
	return &roles.WitchRole{
		Clans:     []entities.Clan{entities.ClanVillager},
		Abilities: abilities,
	}
}

// VillagerRoleFactory pour créer le rôle de villageois
type VillagerRoleFactory struct{}

func (f *VillagerRoleFactory) CreateRole() entities.Role {
	return &roles.VillagerRole{
		Clans: []entities.Clan{entities.ClanVillager},
	}
}

// RoleFactoryRegistry pour gérer toutes les factories de rôles
type RoleFactoryRegistry struct {
	factories map[entities.RoleType]RoleFactory
}

func NewRoleFactoryRegistry() *RoleFactoryRegistry {
	registry := &RoleFactoryRegistry{
		factories: make(map[entities.RoleType]RoleFactory),
	}

	// Enregistrement des factories par défaut
	registry.RegisterFactory(entities.RoleSeer, NewDefaultSeerRoleFactory())
	registry.RegisterFactory(entities.RoleWerewolf, NewDefaultWerewolfRoleFactory())
	registry.RegisterFactory(entities.RoleWitch, NewDefaultWitchRoleFactory())
	registry.RegisterFactory(entities.RoleVillager, NewDefaultVillagerRoleFactory())

	return registry
}

func (r *RoleFactoryRegistry) RegisterFactory(roleType entities.RoleType, factory RoleFactory) {
	r.factories[roleType] = factory
}

func (r *RoleFactoryRegistry) CreateRole(roleType entities.RoleType) entities.Role {
	factory, exists := r.factories[roleType]
	if !exists {
		// Retourne un villageois par défaut si le rôle n'est pas trouvé
		return r.factories[entities.RoleVillager].CreateRole()
	}
	return factory.CreateRole()
}

func (r *RoleFactoryRegistry) GetAvailableRoles() []entities.RoleType {
	roles := make([]entities.RoleType, 0, len(r.factories))
	for roleType := range r.factories {
		roles = append(roles, roleType)
	}
	return roles
}

// Fonctions utilitaires pour créer des factories avec des configurations par défaut

// NewDefaultSeerRoleFactory crée une factory pour le voyant avec la configuration par défaut
func NewDefaultSeerRoleFactory() *SeerRoleFactory {
	return &SeerRoleFactory{
		AbilityFactory: NewDefaultSeerFactory(),
	}
}

// NewDefaultWerewolfRoleFactory crée une factory pour le loup-garou avec la configuration par défaut
func NewDefaultWerewolfRoleFactory() *WerewolfRoleFactory {
	return &WerewolfRoleFactory{
		AbilityFactory: NewDefaultWerewolfFactory(),
	}
}

// NewDefaultWitchRoleFactory crée une factory pour la sorcière avec la configuration par défaut
func NewDefaultWitchRoleFactory() *WitchRoleFactory {
	return &WitchRoleFactory{
		AbilityFactory: NewDefaultWitchFactory(),
	}
}

// NewDefaultVillagerRoleFactory crée une factory pour le villageois
func NewDefaultVillagerRoleFactory() *VillagerRoleFactory {
	return &VillagerRoleFactory{}
}

// Fonctions pour créer des factories personnalisées

// NewCustomSeerRoleFactory crée une factory personnalisée pour le voyant
func NewCustomSeerRoleFactory(abilityFactory *SeerAbilityFactory) *SeerRoleFactory {
	return &SeerRoleFactory{
		AbilityFactory: abilityFactory,
	}
}

// NewCustomWerewolfRoleFactory crée une factory personnalisée pour le loup-garou
func NewCustomWerewolfRoleFactory(abilityFactory *WerewolfAbilityFactory) *WerewolfRoleFactory {
	return &WerewolfRoleFactory{
		AbilityFactory: abilityFactory,
	}
}

// NewCustomWitchRoleFactory crée une factory personnalisée pour la sorcière
func NewCustomWitchRoleFactory(abilityFactory *WitchAbilityFactory) *WitchRoleFactory {
	return &WitchRoleFactory{
		AbilityFactory: abilityFactory,
	}
}

// Exemple d'utilisation pour créer des rôles avec des configurations spécifiques

// CreateGameRoles crée une liste de rôles pour une partie
func CreateGameRoles(roleDistribution map[entities.RoleType]int) []entities.Role {
	registry := NewRoleFactoryRegistry()
	roles := make([]entities.Role, 0)

	for roleType, count := range roleDistribution {
		for i := 0; i < count; i++ {
			role := registry.CreateRole(roleType)
			roles = append(roles, role)
		}
	}

	return roles
}

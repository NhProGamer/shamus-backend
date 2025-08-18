package factories

import (
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/abilities"
)

// SeerAbilityFactory pour créer les abilities du voyant
type SeerAbilityFactory struct{}

func (f *SeerAbilityFactory) CreateAbilities() []entities.Ability {
	return []entities.Ability{
		&abilities.SeeAbility{},
	}
}

func (f *SeerAbilityFactory) CreateSeeAbility() entities.Ability {
	return &abilities.SeeAbility{}
}

// WerewolfAbilityFactory pour créer les abilities du loup-garou
type WerewolfAbilityFactory struct{}

func (f *WerewolfAbilityFactory) CreateAbilities() []entities.Ability {
	return []entities.Ability{
		&abilities.WerewolfKillAbility{},
	}
}

func (f *WerewolfAbilityFactory) CreateKillAbility() entities.Ability {
	return &abilities.WerewolfKillAbility{}
}

// WitchAbilityFactory pour créer les abilities de la sorcière
type WitchAbilityFactory struct {
	HealConsumptions   uint8
	PoisonConsumptions uint8
}

func (f *WitchAbilityFactory) CreateAbilities() []entities.Ability {
	healConsumptions := f.HealConsumptions
	poisonConsumptions := f.PoisonConsumptions

	return []entities.Ability{
		&abilities.HealAbility{
			Consumptions: &healConsumptions,
		},
		&abilities.PoisonAbility{
			Consumptions: &poisonConsumptions,
		},
	}
}

func (f *WitchAbilityFactory) CreateHealAbility() entities.Ability {
	consumptions := f.HealConsumptions
	return &abilities.HealAbility{
		Consumptions: &consumptions,
	}
}

func (f *WitchAbilityFactory) CreatePoisonAbility() entities.Ability {
	consumptions := f.PoisonConsumptions
	return &abilities.PoisonAbility{
		Consumptions: &consumptions,
	}
}

// Fonctions utilitaires pour créer des factories avec des configurations par défaut

// NewDefaultSeerFactory crée une factory pour le voyant avec la configuration par défaut
func NewDefaultSeerFactory() *SeerAbilityFactory {
	return &SeerAbilityFactory{}
}

// NewDefaultWerewolfFactory crée une factory pour le loup-garou avec la configuration par défaut
func NewDefaultWerewolfFactory() *WerewolfAbilityFactory {
	return &WerewolfAbilityFactory{}
}

// NewDefaultWitchFactory crée une factory pour la sorcière avec la configuration par défaut (1 utilisation de chaque)
func NewDefaultWitchFactory() *WitchAbilityFactory {
	return &WitchAbilityFactory{
		HealConsumptions:   1,
		PoisonConsumptions: 1,
	}
}

// NewCustomWitchFactory crée une factory pour

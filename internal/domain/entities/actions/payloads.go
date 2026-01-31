package actions

import "shamus-backend/internal/domain/entities"

// WitchPotionPayload contains the data sent to the witch for her action
type WitchPotionPayload struct {
	// VictimID is the player killed by werewolves (nil if no one was killed)
	VictimID *entities.PlayerID `json:"victimId,omitempty"`
	// HasHealPotion indicates if the witch still has her heal potion available
	HasHealPotion bool `json:"hasHealPotion"`
	// HasPoisonPotion indicates if the witch still has her poison potion available
	HasPoisonPotion bool `json:"hasPoisonPotion"`
}

// SeerVisionPayload contains the data sent to the seer for her action
type SeerVisionPayload struct {
	// EligibleTargets are the players the seer can choose to investigate (alive players)
	EligibleTargets []entities.PlayerID `json:"eligibleTargets"`
}

// WerewolfVotePayload contains the data sent to werewolves for their vote
type WerewolfVotePayload struct {
	// EligibleTargets are the players werewolves can vote to kill (alive non-werewolves)
	EligibleTargets []entities.PlayerID `json:"eligibleTargets"`
}

// VillageVotePayload contains the data sent to all players for the village vote
type VillageVotePayload struct {
	// EligibleTargets are the players that can be voted for elimination (alive players)
	EligibleTargets []entities.PlayerID `json:"eligibleTargets"`
}

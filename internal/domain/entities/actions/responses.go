package actions

import "shamus-backend/internal/domain/entities"

// WitchPotionResponse is the witch's response to her action
type WitchPotionResponse struct {
	// HealTargetID is the player to heal (must be the werewolf victim, nil if not using heal)
	HealTargetID *entities.PlayerID `json:"healTargetId,omitempty"`
	// PoisonTargetID is the player to poison (nil if not using poison)
	PoisonTargetID *entities.PlayerID `json:"poisonTargetId,omitempty"`
}

// SeerVisionResponse is the seer's response to her action
type SeerVisionResponse struct {
	// TargetID is the player the seer wants to investigate
	TargetID entities.PlayerID `json:"targetId"`
}

// WerewolfVoteResponse is a werewolf's response to the vote
type WerewolfVoteResponse struct {
	// TargetID is the player the werewolf votes to kill (nil to abstain)
	TargetID *entities.PlayerID `json:"targetId,omitempty"`
}

// VillageVoteResponse is a player's response to the village vote
type VillageVoteResponse struct {
	// TargetID is the player to vote for elimination (nil to abstain if allowed)
	TargetID *entities.PlayerID `json:"targetId,omitempty"`
}

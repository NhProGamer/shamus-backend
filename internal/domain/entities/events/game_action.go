package events

import "shamus-backend/internal/domain/entities"

// Event types for game actions (client -> server)
const (
	EventTypeVillageVote  entities.EventType = "village_vote"
	EventTypeSeerAction   entities.EventType = "seer_action"
	EventTypeWerewolfVote entities.EventType = "werewolf_vote"
	EventTypeWitchAction  entities.EventType = "witch_action"
)

// VillageVoteEventData is sent by a player to cast their vote during day phase
type VillageVoteEventData struct {
	TargetID *entities.PlayerID `json:"targetId"` // nil = abstain
}

// SeerActionEventData is sent by the seer to see a player's role
type SeerActionEventData struct {
	TargetID entities.PlayerID `json:"targetId"`
}

// WerewolfVoteEventData is sent by a werewolf to vote for a victim
type WerewolfVoteEventData struct {
	TargetID *entities.PlayerID `json:"targetId"` // nil = abstain
}

// WitchActionEventData is sent by the witch to heal or poison
type WitchActionEventData struct {
	HealTargetID   *entities.PlayerID `json:"healTargetId,omitempty"`   // Who to save
	PoisonTargetID *entities.PlayerID `json:"poisonTargetId,omitempty"` // Who to poison
}

// SeerRevealEventData is sent to the seer to reveal a player's role
type SeerRevealEventData struct {
	TargetID entities.PlayerID `json:"targetId"`
	RoleType entities.RoleType `json:"roleType"`
}

const EventTypeSeerReveal entities.EventType = "seer_reveal"

func NewSeerRevealEvent(targetID entities.PlayerID, roleType entities.RoleType) entities.Event[SeerRevealEventData] {
	return entities.Event[SeerRevealEventData]{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeSeerReveal,
		Data: SeerRevealEventData{
			TargetID: targetID,
			RoleType: roleType,
		},
	}
}

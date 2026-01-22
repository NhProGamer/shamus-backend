package events

import "shamus-backend/internal/domain/entities"

const EventTypeTurn entities.EventType = "turn"

type TurnEventData struct {
	RoleType       entities.RoleType  `json:"roleType"`                 // Which role should act
	TargetPlayerID *entities.PlayerID `json:"targetPlayerId,omitempty"` // For witch: who the werewolves attacked
	CanHeal        bool               `json:"canHeal,omitempty"`        // For witch: can she heal?
	CanPoison      bool               `json:"canPoison,omitempty"`      // For witch: can she poison?
}

func NewTurnEvent(roleType entities.RoleType) entities.Event[TurnEventData] {
	return entities.Event[TurnEventData]{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeTurn,
		Data: TurnEventData{
			RoleType: roleType,
		},
	}
}

// NewWitchTurnEvent creates a turn event specifically for the witch with additional info
func NewWitchTurnEvent(targetPlayerID *entities.PlayerID, canHeal, canPoison bool) entities.Event[TurnEventData] {
	return entities.Event[TurnEventData]{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeTurn,
		Data: TurnEventData{
			RoleType:       entities.RoleWitch,
			TargetPlayerID: targetPlayerID,
			CanHeal:        canHeal,
			CanPoison:      canPoison,
		},
	}
}

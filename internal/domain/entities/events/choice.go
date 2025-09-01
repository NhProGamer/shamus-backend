package events

import "shamus-backend/internal/domain/entities"

const EventTypeChoice entities.EventType = "choice"

type ChoiceEventType string

const (
	StartChoice  VoteEventType = "start"
	EndChoice    VoteEventType = "end"
	PlayerChoice VoteEventType = "player"
)

type ChoiceEventData struct {
	Type   VoteEventType      `json:"type"`
	Player *entities.PlayerID `json:"player,omitempty"`
	Target *entities.PlayerID `json:"target,omitempty"`
}

func NewChoiceEvent(eventType VoteEventType, playerID, targetID *entities.PlayerID) entities.Event {
	return entities.Event{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeChoice,
		Data: ChoiceEventData{
			Type:   eventType,
			Player: playerID,
			Target: targetID,
		},
	}
}

package events

import "shamus-backend/internal/domain/entities"

const EventTypeTurn entities.EventType = "turn"

type TurnEventData struct {
}

func NewTurnEvent() entities.Event {
	return entities.Event{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeTurn,
		Data:    TurnEventData{},
	}
}

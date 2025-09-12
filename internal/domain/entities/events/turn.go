package events

import "shamus-backend/internal/domain/entities"

const EventTypeTurn entities.EventType = "turn"

type TurnEventData struct {
}

func NewTurnEvent() entities.Event[TurnEventData] {
	return entities.Event[TurnEventData]{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeTurn,
		Data:    TurnEventData{},
	}
}

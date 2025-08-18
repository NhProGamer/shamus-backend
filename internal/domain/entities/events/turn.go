package events

import "shamus-backend/internal/domain/entities"

const EventTypeTurn entities.EventType = "turn"

type TurnEvent struct {
	PlayerID entities.PlayerID
}

func (e TurnEvent) GetType() entities.EventType {
	return EventTypeTurn
}

func (e TurnEvent) GetData() interface{} {
	return e
}

package events

import "shamus-backend/internal/domain/entities"

const EventTypeDay entities.EventType = "day"

type DayEvent struct {
	Deaths []entities.PlayerID
}

func (e DayEvent) GetType() entities.EventType {
	return EventTypeDay
}

func (e DayEvent) GetData() interface{} {
	return e
}

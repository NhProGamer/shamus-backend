package events

import "shamus-backend/internal/domain/entities"

const EventTypeNight entities.EventType = "night"

type NightEvent struct {
}

func (e NightEvent) GetType() entities.EventType {
	return EventTypeNight
}

func (e NightEvent) GetData() interface{} {
	return e
}

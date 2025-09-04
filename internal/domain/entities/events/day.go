package events

import "shamus-backend/internal/domain/entities"

const EventTypeDay entities.EventType = "day"

type DayEventData struct {
}

func NewDayEvent(deaths []entities.PlayerID) entities.Event {
	return entities.Event{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeDay,
		Data:    DayEventData{},
	}
}

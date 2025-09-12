package events

import "shamus-backend/internal/domain/entities"

const EventTypeDay entities.EventType = "day"

type DayEventData struct {
}

func NewDayEvent(deaths []entities.PlayerID) entities.Event[DayEventData] {
	return entities.Event[DayEventData]{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeDay,
		Data:    DayEventData{},
	}
}

package events

import "shamus-backend/internal/domain/entities"

const EventTypeNight entities.EventType = "night"

type NightEventData struct {
}

func NewNightEvent() entities.Event[NightEventData] {
	return entities.Event[NightEventData]{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeNight,
		Data:    NightEventData{},
	}
}

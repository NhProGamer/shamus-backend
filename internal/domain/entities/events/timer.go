package events

import "shamus-backend/internal/domain/entities"

const EventTypeTimer entities.EventType = "timer"

type GameTimerEventData struct {
	Time int `json:"time"`
}

func NewGameTimerEvent(time int) entities.Event {
	return entities.Event{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeTimer,
		Data: GameTimerEventData{
			Time: time,
		},
	}
}

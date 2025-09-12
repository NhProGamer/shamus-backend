package events

import "shamus-backend/internal/domain/entities"

const EventTypeTimer entities.EventType = "timer"

/*type GameTimerEventData struct {
	MessageCode entities.MessageCode `json:"message_code"`
	Time        int                  `json:"time"`
}

func NewGameTimerEvent(message entities.MessageCode, time int) entities.Event {
	return entities.Event{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeTimer,
		Data: GameTimerEventData{
			MessageCode: message,
			Time:        time,
		},
	}
}*/

package events

import "shamus-backend/internal/domain/entities"

const EventTypeGameStarting entities.EventType = "game_starting"

type GameStartingEventData struct {
}

func NewGameStartingEvent() entities.Event {
	return entities.Event{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeGameStarting,
		Data:    GameStartingEventData{},
	}
}

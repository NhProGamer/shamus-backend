package events

import "shamus-backend/internal/domain/entities"

const EventTypeStartGame entities.EventType = "start_game"

// StartGameEventData is sent by the host to start the game
// This is an incoming event only (client -> server)
type StartGameEventData struct {
}

func NewStartGameEvent() entities.Event[StartGameEventData] {
	return entities.Event[StartGameEventData]{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeStartGame,
		Data:    StartGameEventData{},
	}
}

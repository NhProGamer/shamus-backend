package events

import "shamus-backend/internal/domain/entities"

const EventTypeWin entities.EventType = "win"

type WinEventData struct {
	Winners []entities.PlayerID `json:"winners"`
}

func NewWinEvent(winners []entities.PlayerID) entities.Event {
	return entities.Event{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeWin,
		Data: WinEventData{
			Winners: winners,
		},
	}
}

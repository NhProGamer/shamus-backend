package events

import "shamus-backend/internal/domain/entities"

const EventTypeWin entities.EventType = "win"

type WinEvent struct {
	Winners []entities.PlayerID
}

func (e WinEvent) GetType() entities.EventType {
	return EventTypeNight
}

func (e WinEvent) GetData() interface{} {
	return e
}

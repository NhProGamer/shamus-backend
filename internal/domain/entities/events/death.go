package events

import "shamus-backend/internal/domain/entities"

const EventTypeExecution entities.EventType = "execution"

type ExecutionEvent struct {
	KillerID *entities.PlayerID
	VictimID *entities.PlayerID
}

func (e ExecutionEvent) GetType() entities.EventType {
	return EventTypeExecution
}

func (e ExecutionEvent) GetData() interface{} {
	return e
}

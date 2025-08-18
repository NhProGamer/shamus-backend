package events

import "shamus-backend/internal/domain/entities"

const EventTypeRescue entities.EventType = "rescue"

type RescueEvent struct {
	RescuerID *entities.PlayerID
	SavedID   *entities.PlayerID
}

func (e RescueEvent) GetType() entities.EventType {
	return EventTypeRescue
}

func (e RescueEvent) GetData() interface{} {
	return e
}

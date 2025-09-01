package events

import "shamus-backend/internal/domain/entities"

const EventTypeDeath entities.EventType = "death"

type ExecutionEventData struct {
	Killer *entities.PlayerID `json:"killer,omitempty"`
	Victim entities.PlayerID  `json:"victim"`
}

func NewDeathEvent(killer *entities.PlayerID, victim entities.PlayerID) entities.Event {
	return entities.Event{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeDeath,
		Data: ExecutionEventData{
			Killer: killer,
			Victim: victim,
		},
	}
}

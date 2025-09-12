package events

import "shamus-backend/internal/domain/entities"

const EventTypeDeath entities.EventType = "death"

type DeathEventData struct {
	Killer *entities.PlayerID `json:"killer,omitempty"`
	Victim entities.PlayerID  `json:"victim"`
}

func NewDeathEvent(killer *entities.PlayerID, victim entities.PlayerID) entities.Event[DeathEventData] {
	return entities.Event[DeathEventData]{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeDeath,
		Data: DeathEventData{
			Killer: killer,
			Victim: victim,
		},
	}
}

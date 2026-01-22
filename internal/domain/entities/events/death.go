package events

import "shamus-backend/internal/domain/entities"

const EventTypeDeath entities.EventType = "death"

// DeathEventData represents a player death with role reveal
type DeathEventData struct {
	Victim entities.PlayerID `json:"victim"`
	Role   entities.RoleType `json:"role"`
}

// NewDeathEvent creates a new death event with role reveal
func NewDeathEvent(victim entities.PlayerID, role entities.RoleType) entities.Event[DeathEventData] {
	return entities.Event[DeathEventData]{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeDeath,
		Data: DeathEventData{
			Victim: victim,
			Role:   role,
		},
	}
}

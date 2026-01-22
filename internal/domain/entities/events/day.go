package events

import "shamus-backend/internal/domain/entities"

const EventTypeDay entities.EventType = "day"

// DayEventData signals the start of the day phase
// Deaths are now announced via individual DeathEvent before this event
type DayEventData struct {
	Day int `json:"day"` // Current day number
}

// NewDayEvent creates a new day phase event
func NewDayEvent(day int) entities.Event[DayEventData] {
	return entities.Event[DayEventData]{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeDay,
		Data:    DayEventData{Day: day},
	}
}

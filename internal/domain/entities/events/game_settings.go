package events

import "shamus-backend/internal/domain/entities"

const EventTypeGameSettings entities.EventType = "settings"

type GameDataEventData struct {
	RolesType map[entities.RoleType]int `json:"roles"`
}

func NewGameSettingsEvent(data GameDataEventData) entities.Event {
	return entities.Event{
		Channel: entities.EventChannelSettings,
		Type:    EventTypeGameSettings,
		Data:    data,
	}
}

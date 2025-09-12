package events

import "shamus-backend/internal/domain/entities"

const EventTypeGameHostChange entities.EventType = "host_change"

type HostChangeEventData struct {
	Host entities.PlayerID `json:"host"`
}

func NewHostChangeEvent(host entities.PlayerID) entities.Event[HostChangeEventData] {
	return entities.Event[HostChangeEventData]{
		Channel: entities.EventChannelSettings,
		Type:    EventTypeGameHostChange,
		Data: HostChangeEventData{
			Host: host,
		},
	}
}

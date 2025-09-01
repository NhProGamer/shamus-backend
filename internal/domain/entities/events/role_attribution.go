package events

import "shamus-backend/internal/domain/entities"

const EventTypeRoleAttribution entities.EventType = "win"

type RoleAttributionEventData struct {
	Role entities.RoleType `json:"role"`
}

func NewRoleAttributionEvent(role entities.RoleType) entities.Event {
	return entities.Event{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeRoleAttribution,
		Data: RoleAttributionEventData{
			Role: role,
		},
	}
}

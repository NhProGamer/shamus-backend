package events

import "shamus-backend/internal/domain/entities"

const EventTypeRoleAttribution entities.EventType = "role_reveal"

type RoleRevealEventData struct {
	Role entities.RoleType `json:"role"`
}

func NewRoleAttributionEvent(role entities.RoleType) entities.Event[RoleRevealEventData] {
	return entities.Event[RoleRevealEventData]{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeRoleAttribution,
		Data: RoleRevealEventData{
			Role: role,
		},
	}
}

package events

import "shamus-backend/internal/domain/entities"

const EventTypeRoleAttribution entities.EventType = "win"

type RoleAttributionEvent struct {
	Player entities.PlayerID
	Role   entities.RoleType
}

func (e RoleAttributionEvent) GetType() entities.EventType {
	return EventTypeRoleAttribution
}

func (e RoleAttributionEvent) GetData() interface{} {
	return e
}

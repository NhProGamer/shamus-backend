package events

import "shamus-backend/internal/domain/entities"

const EventTypeInactive entities.EventType = "inactive"

type InactiveEventData struct {
	PlayerID entities.PlayerID `json:"player"`
}

func NewInactiveEvent(playerID entities.PlayerID) entities.Event[InactiveEventData] {
	return entities.Event[InactiveEventData]{
		Channel: entities.EventChannelConnexion,
		Type:    EventTypeInactive,
		Data: InactiveEventData{
			PlayerID: playerID,
		},
	}
}

package events

import "shamus-backend/internal/domain/entities"

const EventTypeReconnection entities.EventType = "reconnection"

type ReconnectionEventData struct {
	PlayerID entities.PlayerID `json:"player"`
}

func NewReconnectionEvent(playerID entities.PlayerID) entities.Event[ReconnectionEventData] {
	return entities.Event[ReconnectionEventData]{
		Channel: entities.EventChannelConnection,
		Type:    EventTypeReconnection,
		Data: ReconnectionEventData{
			PlayerID: playerID,
		},
	}
}

package events

import "shamus-backend/internal/domain/entities"

const EventTypeDisconnection entities.EventType = "disconnection"

type DisconnectionEventData struct {
	PlayerID entities.PlayerID `json:"player"`
}

func NewDisconnectionEvent(playerID entities.PlayerID) entities.Event[DisconnectionEventData] {
	return entities.Event[DisconnectionEventData]{
		Channel: entities.EventChannelConnection,
		Type:    EventTypeDisconnection,
		Data: DisconnectionEventData{
			PlayerID: playerID,
		},
	}
}

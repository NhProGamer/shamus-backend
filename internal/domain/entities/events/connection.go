package events

import "shamus-backend/internal/domain/entities"

const EventTypeConnection entities.EventType = "connection"

type ConnectionEventData struct {
	PlayerID entities.PlayerID `json:"player"`
}

func NewConnectionEvent(playerID entities.PlayerID) entities.Event[ConnectionEventData] {
	return entities.Event[ConnectionEventData]{
		Channel: entities.EventChannelConnection,
		Type:    EventTypeConnection,
		Data: ConnectionEventData{
			PlayerID: playerID,
		},
	}
}

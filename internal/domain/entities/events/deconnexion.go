package events

import "shamus-backend/internal/domain/entities"

const EventTypeDeconnexion entities.EventType = "deconnexion"

type DeconnexionEvent struct {
	PlayerID entities.PlayerID `json:"player"`
}

func NewDeconnexionEvent(playerID entities.PlayerID) entities.Event {
	return entities.Event{
		Channel: entities.EventChannelConnexion,
		Type:    EventTypeDeconnexion,
		Data: DeconnexionEvent{
			PlayerID: playerID,
		},
	}
}

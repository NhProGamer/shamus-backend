package events

import "shamus-backend/internal/domain/entities"

const EventTypeReconnexion entities.EventType = "reconnexion"

type ReconnexionEventData struct {
	PlayerID entities.PlayerID `json:"player"`
}

func NewReconnexionEvent(playerID entities.PlayerID) entities.Event {
	return entities.Event{
		Channel: entities.EventChannelConnexion,
		Type:    EventTypeReconnexion,
		Data: ReconnexionEventData{
			PlayerID: playerID,
		},
	}
}

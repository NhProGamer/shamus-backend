package events

import "shamus-backend/internal/domain/entities"

const EventTypeConnexion entities.EventType = "connexion"

type ConnexionEvent struct {
	PlayerID entities.PlayerID `json:"player"`
}

func NewConnexionEvent(playerID entities.PlayerID) entities.Event[ConnexionEvent] {
	return entities.Event[ConnexionEvent]{
		Channel: entities.EventChannelConnexion,
		Type:    EventTypeConnexion,
		Data: ConnexionEvent{
			PlayerID: playerID,
		},
	}
}

package ws

import (
	"encoding/json"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/ports"
)

type MelodyEventService struct {
	handler *WebSocketHandler
}

func NewMelodyEventService(handler *WebSocketHandler) ports.EventService {
	return &MelodyEventService{handler: handler}
}

func (s *MelodyEventService) SendToPlayer(playerID entities.PlayerID, event entities.RawEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return s.handler.SendToPlayer(playerID, payload)
}

func (s *MelodyEventService) BroadcastToGame(gameID entities.GameID, event entities.RawEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return s.handler.BroadcastToGame(gameID, payload)
}

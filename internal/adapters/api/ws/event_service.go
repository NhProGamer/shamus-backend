package ws

import (
	"encoding/json"
	"errors"
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
	s.handler.lock.RLock()
	defer s.handler.lock.RUnlock()

	// Search through all rooms to find the player
	for _, sessions := range s.handler.rooms {
		for _, sess := range sessions {
			if pid, exists := sess.Get("userId"); exists && pid.(string) == string(playerID) {
				payload, err := json.Marshal(event)
				if err != nil {
					return err
				}
				return sess.Write(payload)
			}
		}
	}
	return errors.New("player not connected")
}

func (s *MelodyEventService) BroadcastToGame(gameID entities.GameID, event entities.RawEvent) error {
	s.handler.lock.RLock()
	defer s.handler.lock.RUnlock()

	sessions, ok := s.handler.rooms[gameID]
	if !ok || len(sessions) == 0 {
		return errors.New("empty room")
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	for _, sess := range sessions {
		_ = sess.Write(payload)
	}
	return nil
}

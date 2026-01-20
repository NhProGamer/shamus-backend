package ws

import (
	"encoding/json"
	"errors"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/ports/app_ports"
)

type MelodyEventService struct {
	handler *WebSocketHandler
}

func NewMelodyEventService(handler *WebSocketHandler) app_ports.EventService {
	return &MelodyEventService{handler: handler}
}

func (s *MelodyEventService) SendToPlayer(gameID entities.GameID, playerID entities.GameID, event entities.Event[any]) error {

	// Récup sessions (copie ta logique lock)
	s.handler.lock.RLock()
	defer s.handler.lock.RUnlock()

	sessions, ok := s.handler.rooms[gameID]
	if !ok {
		return errors.New("room not found")
	}

	for _, sess := range sessions {
		if pid, _ := sess.Get("playerId"); pid == playerID {
			payload, _ := json.Marshal(event.ToRawEvent())
			return sess.Write(payload)
		}
	}
	return errors.New("player not connected")
}

func (s *MelodyEventService) BroadcastToRoom(gameID entities.GameID, event entities.Event[any]) error {
	s.handler.lock.RLock()
	defer s.handler.lock.RUnlock()

	sessions, ok := s.handler.rooms[gameID]
	if !ok || len(sessions) == 0 {
		return errors.New("empty room")
	}

	payload, _ := json.Marshal(event.ToRawEvent())
	for _, sess := range sessions {
		_ = sess.Write(payload)
	}
	return nil
}

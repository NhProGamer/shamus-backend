package adapters

import (
	"encoding/json"
	"log"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/ports"
	"shamus-backend/internal/infrastructure/ws"
)

type WSActionService struct {
	Hub      *ws.Hub
	GameRepo ports.GameRepository
}

func NewWSActionService(hub *ws.Hub, gr ports.GameRepository) *WSActionService {
	return &WSActionService{Hub: hub, GameRepo: gr}
}

func (s *WSActionService) CancelAction(action entities.RawAction, player entities.PlayerID) {
	conn, ok := s.Hub.Get(player)
	if !ok {
		log.Printf("no active connection for player %s", player)
		return
	}
	data, err := json.Marshal(action)
	if err != nil {
		log.Printf("marshal error: %v", err)
		return
	}
	conn.Send <- data
}

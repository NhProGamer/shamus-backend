package adapters

import (
	"encoding/json"
	"log"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/ports"
	"shamus-backend/internal/infrastructure/ws"
)

type WSEventService struct {
	Hub      *ws.Hub
	GameRepo ports.GameRepository
}

func NewWSEventService(hub *ws.Hub, gr ports.GameRepository) *WSEventService {
	return &WSEventService{Hub: hub, GameRepo: gr}
}

func (s *WSEventService) SendEventToPlayer(event entities.Event, player entities.PlayerID) {
	conn, ok := s.Hub.Get(player)
	if !ok {
		log.Printf("no active connection for player %s", player)
		return
	}
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("marshal error: %v", err)
		return
	}
	conn.Send <- data
}

func (s *WSEventService) SendEventToGame(event entities.Event, gameID entities.GameID) {
	actualGame, _ := s.GameRepo.GetGameByID(gameID)
	for _, player := range actualGame.Players {
		s.SendEventToPlayer(event, player)
	}
}

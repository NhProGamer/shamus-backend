package services

import (
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/ports"
)

type eventService struct{}

func NewEventService() ports.EventService {
	return &eventService{}
}

func (s *eventService) SendEventToPlayer(event entities.Event) {
	// Implémentation réelle utiliserait une queue ou WebSocket
}

func (s *eventService) SendEventToGame(event entities.Event, gameID entities.GameID) {
	// Diffuser à tous les joueurs de la partie
}

func (s *eventService) SendEventToClanInAGame(event entities.Event, gameID entities.GameID, clan entities.Clan) {
	// Envoyer uniquement aux joueurs du clan spécifié
}

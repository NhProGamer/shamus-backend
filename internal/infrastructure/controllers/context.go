package controllers

import (
	"shamus-backend/internal/domain/ports"
	"shamus-backend/internal/infrastructure/config"
	"shamus-backend/internal/infrastructure/ws"
)

type AppContext struct {
	Config           *config.Config
	PlayerRepo       ports.PlayerRepository
	GameRepo         ports.GameRepository
	Hub              *ws.Hub
	WebsocketHandler *ws.WebsocketHandler
	EventService     ports.EventService
}

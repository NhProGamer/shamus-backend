package app_ports

import "shamus-backend/internal/domain/entities"

type EventService interface {
	SendToPlayer(gameID entities.GameID, playerID entities.GameID, event entities.Event[any]) error
	BroadcastToRoom(gameID entities.GameID, event entities.Event[any]) error
}

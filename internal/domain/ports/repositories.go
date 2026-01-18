package ports

import "shamus-backend/internal/domain/entities"

type GameRepository interface {
	SaveGame(game *entities.Game) error
	GetGame(id entities.GameID) (*entities.Game, error)
}

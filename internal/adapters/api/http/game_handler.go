package http

import (
	"net/http"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/ports"

	"github.com/gin-gonic/gin"
)

type GameHandler struct {
	gameService ports.GameService
}

func NewGameHandler(gameService ports.GameService) *GameHandler {
	return &GameHandler{gameService: gameService}
}

// CreateGame handles POST /games
func (h *GameHandler) CreateGame(c *gin.Context) {
	var req struct {
		HostID string `json:"hostId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	game, err := h.gameService.CreateNewGame(entities.PlayerID(req.HostID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create game"})
		return
	}

	c.JSON(http.StatusCreated, game.ID)
}

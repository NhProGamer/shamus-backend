package http

import (
	"net/http"
	"shamus-backend/internal/adapters/app_adapters"
	"shamus-backend/internal/domain/entities"

	"github.com/gin-gonic/gin"
)

type GameHandler struct {
	service *app_adapters.GameService
}

func NewGameHandler(s *app_adapters.GameService) *GameHandler {
	return &GameHandler{service: s}
}

// CreateGame : POST /games
func (h *GameHandler) CreateGame(c *gin.Context) {
	var req struct {
		HostID string `json:"hostId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	game, err := h.service.CreateNewGame(entities.PlayerID(req.HostID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create game"})
		return
	}

	c.JSON(http.StatusCreated, game.ID)
}

package controllers

import (
	"net/http"
	"shamus-backend/internal/domain/entities"

	"github.com/gin-gonic/gin"
)

func (ctx *AppContext) PostGameHandler(c *gin.Context) {
	userIDstr, exist := c.Get("userID")
	if !exist || userIDstr == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing user ID"})
		return
	}
	game, err := ctx.GameService.CreateNewGame(c.Request.Context(), entities.PlayerID(userIDstr.(string)))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"gameID": game.ID})
}

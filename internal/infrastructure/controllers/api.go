package controllers

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/pkg/utils"
)

func (ctx *AppContext) GetGameHandler(c *gin.Context) {
	// Récupérer la session
	session := sessions.Default(c)
	userIDValue := session.Get("userID")

	// Vérifier que userID est présent et de type string
	userIDStr, ok := userIDValue.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userID is missing in session"})
		return
	}

	// Convertir en PlayerID
	playerID := entities.PlayerID(userIDStr)

	player, err := ctx.PlayerRepo.GetPlayerByID(playerID)
	if err != nil || player == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you are not in a game"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not implemented"})
	}

}

func (ctx *AppContext) PostGameHandler(c *gin.Context) {
	// Récupérer la session
	session := sessions.Default(c)
	userIDValue := session.Get("userID")

	// Vérifier que userID est présent et de type string
	userIDStr, ok := userIDValue.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userID is missing in session"})
		return
	}

	// Convertir en PlayerID
	playerID := entities.PlayerID(userIDStr)

	player, err := ctx.PlayerRepo.GetPlayerByID(playerID)
	if err != nil || player == nil {
		id, _ := utils.GenerateID(8)
		gameID := entities.GameID(id)

		gameSettings := entities.NewGameSettings(
			20,
			5,
			nil,
		)
		game := entities.NewGame(gameID, playerID, gameSettings)

		ctx.GameRepo.CreateGame(game)
		actualPlayer := entities.NewSafePlayer(playerID, "", &gameID)
		actualPlayer.Disconnect()
		ctx.PlayerRepo.AddPlayer(actualPlayer)

	} else {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can't create a game if you are already in one"})
	}
}

func (ctx *AppContext) PatchGameSettingsHandler(c *gin.Context) {

}

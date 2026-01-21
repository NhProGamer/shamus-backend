package routes

import (
	"net/http"
	"shamus-backend/internal/infrastructure/controllers"
	"shamus-backend/internal/infrastructure/middlewares"

	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.Engine, ctx *controllers.AppContext) {

	/*initParams := gin_oidc.InitParams{
		Router:       r,
		ClientId:     ctx.Config.OIDC.ClientID,
		ClientSecret: ctx.Config.OIDC.Secret,
		Issuer:       ctx.Config.OIDC.GetIssuerURL(),
		ClientUrl:    ctx.Config.Server.GetPublicURL(),
		Scopes:       ctx.Config.OIDC.Scopes,
	}*/

	// Protected routes (require authentication)
	protected := r.Group("/app")
	protected.Use(middlewares.OIDCHandler(ctx))

	protected.StaticFile("/", "./web/static/app.html")
	protected.StaticFile("/game", "./web/static/app/game.html")
	protected.GET("/ws/:gameID", ctx.WebsocketHandler.HandleWS)

	api := protected.Group("/api/v1")
	api.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "You are authenticated! "})
	})

	//api.GET("game", ctx.GetGameHandler)
	api.POST("game", ctx.PostGameHandler)

	//api.PATCH("gameSettings", ctx.PatchGameSettingsHandler)

}

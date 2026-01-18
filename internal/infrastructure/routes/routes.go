package routes

import (
	"shamus-backend/internal/infrastructure/controllers"

	"github.com/gin-gonic/gin"
	gin_oidc "github.com/maximRnback/gin-oidc"
)

func InitRoutes(r *gin.Engine, ctx *controllers.AppContext) {
	r.Static("/_next", "./web/static/_next")
	r.StaticFile("/", "./web/static/index.html")

	initParams := gin_oidc.InitParams{
		Router:       r,
		ClientId:     ctx.Config.OIDC.ClientID,
		ClientSecret: ctx.Config.OIDC.Secret,
		Issuer:       ctx.Config.OIDC.GetIssuerURL(),
		ClientUrl:    ctx.Config.Server.GetPublicURL(),
		Scopes:       ctx.Config.OIDC.Scopes,
	}

	// Si connecté avec discord
	protected := r.Group("/app")
	protected.Use(gin_oidc.Init(initParams))

	protected.StaticFile("/", "./web/static/app.html")
	protected.StaticFile("/game", "./web/static/app/game.html")
	//protected.GET("/ws/:gameID", ctx.WebsocketHandler.Handle)

	//api := protected.Group("/api/v1")
	//api.GET("/test", func(c *gin.Context) {
	//	c.JSON(http.StatusOK, gin.H{"message": "You are authenticated! "})
	//})

	//api.GET("game", ctx.GetGameHandler)
	//api.POST("game", ctx.PostGameHandler)

	//api.PATCH("gameSettings", ctx.PatchGameSettingsHandler)

}

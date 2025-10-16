package routes

import (
	"net/http"
	"shamus-backend/internal/infrastructure/controllers"
	"shamus-backend/internal/infrastructure/middlewares"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.Engine, ctx *controllers.AppContext) {
	//r.Static("/static", "./web/static")
	r.Static("/_next", "./web/static/_next")
	//r.LoadHTMLGlob("web/html/*")

	r.StaticFile("/", "./web/static/index.html")
	r.GET("/login", ctx.LoginHandler)
	r.GET("/callback", ctx.CallbackHandler)

	// Si connecté avec discord
	protected := r.Group("/app")
	protected.Use(middlewares.DiscordAuthMiddleware())

	protected.StaticFile("/", "./web/static/app.html")
	protected.StaticFile("/game", "./web/static/app/game.html")
	protected.GET("/ws/:gameID", ctx.WebsocketHandler.Handle)

	api := protected.Group("/api/v1")
	api.GET("/test", func(c *gin.Context) {
		session := sessions.Default(c)
		c.JSON(http.StatusOK, gin.H{"message": "You are authenticated! " + session.Get("userID").(string)})
	})

	api.GET("game", ctx.GetGameHandler)
	api.POST("game", ctx.PostGameHandler)

	api.PATCH("gameSettings", ctx.PatchGameSettingsHandler)

}

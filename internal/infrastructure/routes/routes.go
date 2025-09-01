package routes

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"shamus-backend/internal/infrastructure/controllers"
	"shamus-backend/internal/infrastructure/middlewares"
)

func InitRoutes(r *gin.Engine, ctx *controllers.AppContext) {
	r.Static("/static", "./frontend/static")
	r.LoadHTMLGlob("web/html/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
	r.GET("/login", ctx.LoginHandler)
	r.GET("/callback", ctx.CallbackHandler)

	// Si connecté avec discord
	protected := r.Group("/app")
	protected.Use(middlewares.DiscordAuthMiddleware())

	protected.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "test.html", gin.H{})
	})
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

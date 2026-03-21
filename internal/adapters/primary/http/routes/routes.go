package routes

import (
	"net/http"
	"shamus-backend/internal/adapters/primary/http/controllers"
	"shamus-backend/internal/adapters/primary/http/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func InitRoutes(r *gin.Engine, ctx *controllers.AppContext, rdb *redis.Client) {
	// Health check endpoints (no authentication required)
	r.GET("/health", controllers.HealthHandler(rdb))
	r.GET("/ready", controllers.ReadinessHandler(rdb))
	r.GET("/live", controllers.LivenessHandler())

	// API Documentation (only in debug mode)
	if ctx.Config.Debug {
		initDocsRoutes(r)
	}

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

	api.POST("game", ctx.PostGameHandler)
}

// initDocsRoutes sets up API documentation routes (debug mode only)
func initDocsRoutes(r *gin.Engine) {
	// Serve OpenAPI/AsyncAPI spec files
	r.Static("/docs/api", "./docs/api")

	// Swagger UI for REST API
	r.GET("/docs/rest", func(c *gin.Context) {
		c.File("./docs/api/swagger-ui.html")
	})

	// AsyncAPI UI for WebSocket API
	r.GET("/docs/ws", func(c *gin.Context) {
		c.File("./docs/api/asyncapi-ui.html")
	})

	// Redirect /docs to /docs/rest
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/docs/rest")
	})
}

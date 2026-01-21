package main

import (
	"context"
	"log"
	"shamus-backend/internal/adapters/api/http"
	"shamus-backend/internal/adapters/api/ws"
	"shamus-backend/internal/adapters/app_adapters"
	"shamus-backend/internal/adapters/infra_adapters"
	"shamus-backend/internal/infrastructure/config"
	"shamus-backend/internal/infrastructure/controllers"
	"shamus-backend/internal/infrastructure/routes"
	"strconv"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	r := gin.Default()
	m := melody.New()

	// Configure CORS for frontend
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Initialize OIDC provider
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, cfg.OIDC.Issuer)
	if err != nil {
		log.Fatalf("Failed to initialize OIDC provider: %v", err)
	}

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + strconv.Itoa(cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Verify Redis connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")

	// Wire up dependencies
	gameRepo := infra_adapters.NewRedisGameRepo(rdb)
	gameService := app_adapters.NewGameService(gameRepo)
	httpHandler := http.NewGameHandler(gameService)
	wsHandler := ws.NewWebSocketHandler(m, gameService)

	// Initialize routes
	routes.InitRoutes(r, &controllers.AppContext{
		Config:           &cfg,
		GameService:      gameService,
		WebsocketHandler: wsHandler,
		HttpHandler:      httpHandler,
		OIDCProvider:     provider,
	})

	// Start server
	addr := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

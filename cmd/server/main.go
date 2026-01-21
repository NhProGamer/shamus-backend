package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	httphandler "shamus-backend/internal/adapters/api/http"
	"shamus-backend/internal/adapters/api/ws"
	"shamus-backend/internal/adapters/app"
	"shamus-backend/internal/adapters/infra"
	"shamus-backend/internal/infrastructure/config"
	"shamus-backend/internal/infrastructure/controllers"
	"shamus-backend/internal/infrastructure/routes"
	"strconv"
	"syscall"
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
	gameRepo := infra.NewRedisGameRepo(rdb)
	playerRepo := infra.NewRedisPlayerRepo(rdb)

	gameService := app.NewGameService(gameRepo)
	visibilityService := app.NewVisibilityService()

	// Create WebSocket handler first (without PlayerService to break circular dependency)
	wsHandler := ws.NewWebSocketHandler(m, gameService, visibilityService)

	// Create PlayerService with WebSocketHandler as ConnectionChecker
	playerService := app.NewPlayerService(playerRepo, gameRepo, wsHandler)

	// Inject PlayerService into WebSocketHandler (completes the wiring)
	wsHandler.SetPlayerService(playerService)

	httpHandler := httphandler.NewGameHandler(gameService)

	// Initialize routes
	routes.InitRoutes(r, &controllers.AppContext{
		Config:           &cfg,
		GameService:      gameService,
		WebsocketHandler: wsHandler,
		HttpHandler:      httpHandler,
		OIDCProvider:     provider,
	})

	// Start server with graceful shutdown
	addr := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("Starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 5 seconds to complete
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

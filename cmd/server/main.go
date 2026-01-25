package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"shamus-backend/internal/adapters/api/ws"
	"shamus-backend/internal/adapters/app"
	"shamus-backend/internal/adapters/infra"
	"shamus-backend/internal/domain/constants"
	"shamus-backend/internal/infrastructure/config"
	"shamus-backend/internal/infrastructure/controllers"
	"shamus-backend/internal/infrastructure/routes"
	"shamus-backend/pkg/logger"
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
		panic("Failed to load config: " + err.Error())
	}

	// Initialize logger
	log := logger.New(logger.Config{
		Level:  cfg.Logger.Level,
		Pretty: cfg.Logger.Pretty,
	})
	log.Info().Msg("Starting Shamus backend server")

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
		log.Fatal().Msgf("Failed to initialize OIDC provider: %v", err)
	}

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + strconv.Itoa(cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Verify Redis connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal().Msgf("Failed to connect to Redis: %v", err)
	}
	log.Info().Msg("Connected to Redis")

	// Wire up dependencies
	gameRepo := infra.NewRedisGameRepo(rdb)
	playerRepo := infra.NewRedisPlayerRepo(rdb)

	gameService := app.NewGameService(gameRepo, playerRepo)
	visibilityService := app.NewVisibilityService()
	chatService := app.NewChatService()

	// Create WebSocket handler first (without PlayerService to break circular dependency)
	wsHandler := ws.NewWebSocketHandler(m, gameService, visibilityService, chatService)

	// Create PlayerService with WebSocketHandler as ConnectionChecker
	playerService := app.NewPlayerService(playerRepo, gameRepo, wsHandler)

	// Inject PlayerService into WebSocketHandler (completes the wiring)
	wsHandler.SetPlayerService(playerService)

	// Create game engine services (wsHandler implements Broadcaster and PlayerSender interfaces)
	timerService := app.NewTimerService(wsHandler)
	voteService := app.NewVoteService(wsHandler, wsHandler)
	nightService := app.NewNightService(wsHandler, voteService)

	// Create GameEngine and inject into WebSocketHandler
	gameEngine := app.NewGameEngine(
		gameRepo,
		playerRepo,
		timerService,
		voteService,
		nightService,
		wsHandler,
		wsHandler,
	)
	wsHandler.SetGameEngine(gameEngine)

	log.Info().Msg("Game engine services initialized")

	// Initialize routes
	routes.InitRoutes(r, &controllers.AppContext{
		Config:           &cfg,
		GameService:      gameService,
		WebsocketHandler: wsHandler,
		OIDCProvider:     provider,
	}, rdb)

	// Start server with graceful shutdown
	addr := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  constants.ServerReadTimeout,
		WriteTimeout: constants.ServerWriteTimeout,
	}

	go func() {
		log.Info().Msgf("Starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msgf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	// Give outstanding requests 5 seconds to complete
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Msgf("Server forced to shutdown: %v", err)
	}

	log.Info().Msg("Server exited")
}

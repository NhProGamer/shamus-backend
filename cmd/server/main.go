package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"shamus-backend/internal/adapters/primary/http/controllers"
	"shamus-backend/internal/adapters/primary/http/routes"
	"shamus-backend/internal/adapters/primary/websocket"
	redisadapter "shamus-backend/internal/adapters/secondary/redis"
	"shamus-backend/internal/application/orchestration"
	"shamus-backend/internal/application/services"
	"shamus-backend/internal/domain/constants"
	"shamus-backend/internal/infrastructure/config"
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

	// === REPOSITORIES ===
	gameRepo := redisadapter.NewRedisGameRepo(rdb)
	playerRepo := redisadapter.NewRedisPlayerRepo(rdb)
	voteRepo := redisadapter.NewVoteRepository(rdb)

	// === CORE SERVICES ===
	gameService := services.NewGameService(gameRepo, playerRepo)
	chatService := services.NewChatService()

	// === SESSION MANAGER ===
	// Manages WebSocket sessions and game rooms
	sessionManager := websocket.NewSessionManager()

	// === NOTIFICATION SERVICE ===
	// Sends typed notifications to players (server -> client, one-way)
	notificationService := services.NewNotificationService(sessionManager, sessionManager, playerRepo)

	// === PROMPT SERVICE ===
	// Handles interactive prompts with timeouts and group voting
	promptService := services.NewPromptService(sessionManager, notificationService, playerRepo)

	// === COMMAND HANDLER ===
	// Handles client-initiated commands (chat, settings, start, kick)
	// Note: PlayerService and GameEngine will be set later to break circular dependency
	commandHandler := websocket.NewCommandHandler(gameService, nil, chatService, notificationService)

	// === WEBSOCKET HANDLER ===
	// Main WebSocket handler using Notification/Prompt/Command architecture
	wsHandler := websocket.NewHandler(m, sessionManager, promptService, commandHandler, notificationService, gameService)

	// === PLAYER SERVICE ===
	// Now we can create PlayerService with wsHandler as ConnectionChecker
	playerService := services.NewPlayerService(playerRepo, gameRepo, wsHandler)
	playerService.SetNotifier(notificationService) // For host migration notifications

	// Inject PlayerService into WebSocket handler and CommandHandler
	wsHandler.SetPlayerService(playerService)
	commandHandler.SetPlayerService(playerService)
	commandHandler.SetDisconnecter(wsHandler) // For kicking players

	// === GAME ENGINE SERVICES ===
	// VoteService for legacy compatibility (used by NightService)
	voteService := services.NewVoteService(voteRepo, sessionManager, sessionManager)
	nightService := services.NewNightService(sessionManager, playerRepo, voteService)

	// === GAME ENGINE V2 ===
	// Orchestrates game flow using Prompt/Notification architecture
	gameEngine := orchestration.NewGameEngineV2(
		gameRepo,
		playerRepo,
		promptService,
		notificationService,
		nightService,
	)

	// Inject GameEngine into CommandHandler
	commandHandler.SetGameEngine(gameEngine)

	log.Info().Msg("New architecture services initialized")

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

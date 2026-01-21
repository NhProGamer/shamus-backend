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
	var err error
	var Configuration config.Config
	Configuration, err = config.LoadConfig()
	if err != nil {
		log.Fatalf("Error while loading config file: %v", err)
	}
	r := gin.Default()
	m := melody.New()

	r.Use(cors.New(cors.Config{
		// Autoriser l'origine de ton frontend Vue.js
		AllowOrigins: []string{"http://localhost:3000", "http://localhost:5173"},
		// Autoriser les méthodes HTTP utilisées
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		// Autoriser les headers spécifiques (Authorization est vital pour ton token Bearer)
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		// Autoriser l'envoi de cookies/auth headers
		AllowCredentials: true,
		// Durée du cache de la réponse preflight
		MaxAge: 12 * time.Hour,
	}))

	c := context.Background()
	provider, err := oidc.NewProvider(c, Configuration.OIDC.Issuer)
	if err != nil {
		panic("Impossible d'initialiser le provider OIDC: " + err.Error())
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     Configuration.Redis.Host + ":" + strconv.Itoa(Configuration.Redis.Port),
		Password: Configuration.Redis.Password,
		DB:       Configuration.Redis.DB,
	})

	gameRepo := infra_adapters.NewRedisGameRepo(rdb)
	gameService := app_adapters.NewGameService(gameRepo)
	httpHandler := http.NewGameHandler(gameService)
	wsHandler := ws.NewWebSocketHandler(m, gameService)

	routes.InitRoutes(r, &controllers.AppContext{
		Config:           &Configuration,
		GameService:      gameService,
		WebsocketHandler: wsHandler,
		HttpHandler:      httpHandler,
		OIDCProvider:     provider,
	})

	if err := r.Run(Configuration.Server.Host + ":" + strconv.Itoa(Configuration.Server.Port)); err != nil {
		log.Fatalf("Could not start the server: %v", err)
	}

}

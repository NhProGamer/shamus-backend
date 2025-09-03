package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"log"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/infrastructure/adapters"
	"shamus-backend/internal/infrastructure/config"
	"shamus-backend/internal/infrastructure/controllers"
	"shamus-backend/internal/infrastructure/routes"
	"shamus-backend/internal/infrastructure/ws"
	"strconv"
)

func main() {
	var err error
	var Configuration config.Config
	Configuration, err = config.LoadConfig()
	if err != nil {
		log.Fatalf("Error while loading config file: %v", err)
	}

	hub := ws.NewHub()
	playerRepo := adapters.NewPlayerRepository()
	gameRepo := adapters.NewGameRepository(playerRepo, hub)

	eventService := adapters.NewWSEventService(hub, gameRepo)
	wsHandler := ws.NewWebsocketHandler(hub, playerRepo, eventService, gameRepo)

	r := gin.Default()

	store := cookie.NewStore([]byte(Configuration.Server.CookieStoreKey))
	r.Use(sessions.Sessions("session", store))

	testGameID := entities.GameID("test-game")
	//TEST pour la connection
	testGameSettings := entities.NewGameSettings(
		10,
		4,
		nil)
	testGame := entities.NewGame(testGameID, "363391883755651072", testGameSettings)

	testPlayer := entities.NewSafePlayer("363391883755651072", "TestPlayer", &testGameID)

	gameRepo.CreateGame(testGame)
	playerRepo.AddPlayer(testPlayer)

	routes.InitRoutes(r, &controllers.AppContext{
		Config:           &Configuration,
		PlayerRepo:       playerRepo,
		GameRepo:         gameRepo,
		Hub:              hub,
		WebsocketHandler: wsHandler,
		EventService:     eventService,
	})

	if err := r.Run(Configuration.Server.Host + ":" + strconv.Itoa(Configuration.Server.Port)); err != nil {
		log.Fatalf("Could not start the server: %v", err)
	}

}

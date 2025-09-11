// websocket_handler.go
package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/events"
	"shamus-backend/internal/domain/ports"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebsocketHandler gère les connexions WebSocket pour le jeu
type WebsocketHandler struct {
	hub              *Hub
	playerRepo       ports.PlayerRepository
	gameRepo         ports.GameRepository
	eventService     ports.EventService
	inactivityTimers map[entities.PlayerID]context.CancelFunc // pour annuler les timers
}

// upgrader configure l'upgrade HTTP vers WebSocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implémenter une vérification d'origine plus stricte en production
		return true
	},
}

// NewWebsocketHandler crée une nouvelle instance du handler WebSocket
func NewWebsocketHandler(hub *Hub, playerRepo ports.PlayerRepository, eventService ports.EventService, gameRepo ports.GameRepository) *WebsocketHandler {
	if hub == nil || playerRepo == nil || eventService == nil || gameRepo == nil {
		panic("tous les paramètres sont requis")
	}

	return &WebsocketHandler{
		hub:              hub,
		playerRepo:       playerRepo,
		gameRepo:         gameRepo,
		eventService:     eventService,
		inactivityTimers: make(map[entities.PlayerID]context.CancelFunc),
	}
}

// Handle traite une demande de connexion WebSocket
func (h *WebsocketHandler) Handle(c *gin.Context) {
	newPlayer := false
	ctx := c.Request.Context()

	// Validation des paramètres de session et de route
	playerID, gameID, err := h.validateRequest(c)
	if err != nil {
		log.Printf("validation error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actualGame, err := h.gameRepo.GetGameByID(gameID)
	if err != nil || actualGame == nil {
		log.Printf("game not found: %s", gameID)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrGameNotFound.Error()})
		return
	}

	actualPlayer, err := h.playerRepo.GetPlayerByID(playerID)
	switch {
	case actualPlayer == nil || err != nil:
		if actualGame.Status() != entities.GameStatusWaiting {
			log.Printf("game %s is active", gameID)
			c.JSON(http.StatusBadRequest, gin.H{"error": ErrGameActive.Error()})
			return
		}
		if actualGame.IsFull() {
			log.Printf("game full: %s", gameID)
			c.JSON(http.StatusBadRequest, gin.H{"error": ErrGameFull.Error()})
			return
		}
		player := entities.NewSafePlayer(playerID, "", &gameID)
		if err := h.playerRepo.AddPlayer(player); err != nil {
			log.Printf("error adding player: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer.Error()})
			return
		}
		actualGame.AddPlayer(playerID)
		newPlayer = true
	case *actualPlayer.GetGameID() != gameID:
		log.Printf("player %s trying to join game %s but belongs to game %s", playerID, gameID, *actualPlayer.GetGameID())
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrPlayerNotInGame.Error()})
		return
	}

	h.cancelInactivityTimer(playerID)
	h.closeExistingConnection(playerID)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}
	h.configureWebSocketConn(conn)

	client := NewClientConn(playerID, conn)
	h.hub.Register(playerID, client)

	h.sendConnectionEvent(playerID, gameID, newPlayer)

	log.Printf("player %s connected to game %s (new: %v)", playerID, gameID, newPlayer)

	go h.handleWritePump(client)
	go h.handleReadPump(client, ctx)
}

// validateRequest valide les paramètres de la requête
func (h *WebsocketHandler) validateRequest(c *gin.Context) (entities.PlayerID, entities.GameID, error) {
	// Validation de la session
	session := sessions.Default(c)
	userIDValue := session.Get("userID")

	userIDStr, ok := userIDValue.(string)
	if !ok || userIDStr == "" {
		return "", "", ErrUserIDMissing
	}

	playerID := entities.PlayerID(userIDStr)

	// Validation du gameID
	gameIDStr := c.Param("gameID")
	if gameIDStr == "" {
		return "", "", ErrGameIDMissing
	}

	gameID := entities.GameID(gameIDStr)
	return playerID, gameID, nil
}

// closeExistingConnection ferme une connexion WebSocket existante si elle existe
func (h *WebsocketHandler) closeExistingConnection(playerID entities.PlayerID) {
	if existingClient, exists := h.hub.Get(playerID); exists {
		log.Printf("fermeture de l'ancienne connexion WebSocket pour le joueur %s", playerID)
		h.hub.Unregister(playerID)
		existingClient.Conn.Close()
	}
}

// sendConnectionEvent envoie l'événement de connexion approprié
func (h *WebsocketHandler) sendConnectionEvent(playerID entities.PlayerID, gameID entities.GameID, isNewPlayer bool) {
	actualGame, _ := h.gameRepo.GetGameByID(gameID)
	players := actualGame.Players()
	playersData := make([]events.PlayersDetailsData, 0, len(players))
	for _, pID := range players {
		player, err := h.playerRepo.GetPlayerByID(pID)
		if err != nil || player == nil {
			continue
		}
		var role *entities.RoleType
		if (!player.Alive() || player.ID() == playerID) && player.Role() != nil {
			rt := (*player.Role()).GetType()
			role = &rt
		}
		playersData = append(playersData, events.PlayersDetailsData{
			ID:             player.ID(),
			Alive:          player.Alive(),
			Role:           role,
			Target:         player.VotedFor(),
			ConnexionState: player.ConnectionState(),
		})
	}

	var event entities.Event
	if isNewPlayer {
		event = events.NewConnexionEvent(playerID)
	} else {
		event = events.NewReconnexionEvent(playerID)
	}
	h.eventService.SendEventToGame(event, gameID)

	gameDataEvent := events.NewGameDataEvent(events.GameDataEventData{
		ID:       gameID,
		Status:   actualGame.Status(),
		Phase:    actualGame.Phase(),
		Day:      actualGame.Day(),
		Players:  playersData,
		Host:     actualGame.Host(),
		Settings: actualGame.Settings(),
	})
	h.eventService.SendEventToPlayer(gameDataEvent, playerID)
}

// configureWebSocketConn configure les paramètres de la connexion WebSocket
func (h *WebsocketHandler) configureWebSocketConn(conn *websocket.Conn) {
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
}

// handleWritePump gère l'envoi des messages
func (h *WebsocketHandler) handleWritePump(client *ClientConn) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic dans writePump pour %s: %v", client.PlayerID, r)
		}
	}()

	client.WritePump()
}

// handleReadPump gère la réception des messages
func (h *WebsocketHandler) handleReadPump(client *ClientConn, ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic dans readPump pour %s: %v", client.PlayerID, r)
		}
	}()

	client.ReadPump(
		func(playerID entities.PlayerID, msg []byte) {
			h.handleMessage(ctx, playerID, msg)
		},
		h.playerRepo,
		h.eventService,
		h.gameRepo,
		h, // passer le handler pour accéder aux timers
	)
}

// handleMessage traite un message reçu d'un client
func (h *WebsocketHandler) handleMessage(ctx context.Context, playerID entities.PlayerID, msg []byte) {
	log.Printf("message de %s: %s", playerID, string(msg))

	var result entities.Event
	if err := json.Unmarshal(msg, &result); err != nil {
		log.Println("invalid message format:", err)
		return
	}

	actualPlayer, _ := h.playerRepo.GetPlayerByID(playerID)
	actualGame, _ := h.gameRepo.GetGameByID(*actualPlayer.GetGameID())

	switch result.GetChannel() {
	case entities.EventChannelConnexion:
		return // Ignorer les messages de connexion
	case entities.EventChannelSettings:
		if result.GetType() != events.EventTypeGameSettings {
			break
		}
		if actualPlayer.ID() != actualGame.Host() {
			log.Println("only host can change settings")
			return
		}
		raw := result.GetData()
		dataBytes, err := json.Marshal(raw)
		if err != nil {
			log.Println("erreur marshal Data:", err)
			return
		}
		var settings events.GameSettingsEventData
		if err := json.Unmarshal(dataBytes, &settings); err != nil {
			log.Println("erreur decode GameSettingsEventData:", err)
			return
		}
		gameSettings := entities.NewGameSettings(settings.RolesType)
		if err := actualGame.SetSettings(&gameSettings); err != nil {
			log.Println("erreur set GameSettings:", err)
			return
		}
		h.eventService.SendEventToGame(events.NewGameSettingsEvent(events.GameSettingsEventData{RolesType: gameSettings.Roles}), actualGame.ID())
	case entities.EventChannelGameEvent:
		// TODO: gérer les votes
	}
}

// cancelInactivityTimer annule le timer d'inactivité pour un joueur
func (h *WebsocketHandler) cancelInactivityTimer(playerID entities.PlayerID) {
	if cancelFunc, exists := h.inactivityTimers[playerID]; exists {
		cancelFunc()
		delete(h.inactivityTimers, playerID)
		log.Printf("timer d'inactivité annulé pour le joueur %s", playerID)
	}
}

// setInactivityTimer définit un timer d'inactivité pour un joueur
func (h *WebsocketHandler) setInactivityTimer(playerID entities.PlayerID, cancelFunc context.CancelFunc) {
	// Annuler l'ancien timer s'il existe
	h.cancelInactivityTimer(playerID)

	// Définir le nouveau timer
	h.inactivityTimers[playerID] = cancelFunc
}

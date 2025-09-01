// websocket_handler.go
package ws

import (
	"context"
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
	// Est ce que la game existe ?
	if err != nil || actualGame == nil {
		// Game n'existe pas
		log.Printf("game not found: %s", gameID)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrGameNotFound.Error()})
		return
	}
	//Game existe

	actualPlayer, err := h.playerRepo.GetPlayerByID(playerID)
	if actualPlayer == nil || err != nil {
		// Le joueur n'existe pas

		// Est ce que la game est en mode Waiting ?
		if actualGame.Status == entities.GameStatusWaiting {
			// Game en attente

			// Est ce que la game est pleine ?
			if len(actualGame.Players) == actualGame.Settings.MaxPlayers {

				// game pleine
				log.Printf("game full: %s", gameID)
				c.JSON(http.StatusBadRequest, gin.H{"error": ErrGameFull.Error()})
				return
			}
			// Game pas pleine

			// Créer le joueur
			player := entities.NewSafePlayer(playerID, "", &gameID)
			if err := h.playerRepo.AddPlayer(player); err != nil {
				log.Printf("error adding player: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer.Error()})
				return
			}

			// Ajouter le joueur à la partie
			actualGame.Players = append(actualGame.Players, playerID)

			newPlayer = true
		} else {
			// Game pas en attente
			log.Printf("game %s is active", gameID)
			c.JSON(http.StatusBadRequest, gin.H{"error": ErrGameActive.Error()})
			return
		}
	} else {
		// Le joueur existe

		// Le joueur se reconecte a la bonne partie ?
		if *actualPlayer.GetGameID() != gameID {
			// Le joueur tente de rejoindre une partie à laquelle il n'appartient pas
			log.Printf("player %s trying to join game %s but belongs to game %s", playerID, gameID, *actualPlayer.GetGameID())
			c.JSON(http.StatusBadRequest, gin.H{"error": ErrPlayerNotInGame.Error()})
			return
		}
	}

	// Annuler le timer d'inactivité s'il existe
	h.cancelInactivityTimer(playerID)

	// Fermer toute connexion WebSocket existante pour ce joueur
	h.closeExistingConnection(playerID)

	// Upgrade de la connexion HTTP vers WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}

	// Configuration de la connexion WebSocket
	h.configureWebSocketConn(conn)

	// Création et enregistrement du nouveau client
	client := NewClientConn(playerID, conn)
	h.hub.Register(playerID, client)

	// Envoyer l'événement approprié selon le type de connexion
	h.sendConnectionEvent(playerID, gameID, newPlayer)

	log.Printf("player %s connected to game %s (new: %v)", playerID, gameID, newPlayer)

	// Lancement des goroutines de gestion
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
	if isNewPlayer {
		event := events.NewConnexionEvent(playerID)
		h.eventService.SendEventToGame(event, gameID)
	} else {
		event := events.NewReconnexionEvent(playerID)
		h.eventService.SendEventToGame(event, gameID)
	}
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

	// TODO: Implémenter le dispatch vers l'application layer
	// Exemple:
	// if err := h.messageHandler.HandleMessage(ctx, playerID, msg); err != nil {
	//     log.Printf("erreur lors du traitement du message: %v", err)
	// }
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

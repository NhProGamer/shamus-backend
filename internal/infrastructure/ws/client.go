// client.go
package ws

import (
	"context"
	"log"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/events"
	"shamus-backend/internal/domain/ports"
	"time"

	"github.com/gorilla/websocket"
)

// TimerManager interface pour gérer les timers d'inactivité
type TimerManager interface {
	setInactivityTimer(playerID entities.PlayerID, cancelFunc context.CancelFunc)
	cancelInactivityTimer(playerID entities.PlayerID)
}
type ClientConn struct {
	PlayerID entities.PlayerID
	Conn     *websocket.Conn
	Send     chan []byte
}

func NewClientConn(playerID entities.PlayerID, conn *websocket.Conn) *ClientConn {
	return &ClientConn{
		PlayerID: playerID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
	}
}

const (
	pongWait       = 30 * time.Second // délai max avant de considérer la connexion morte
	pingPeriod     = 15 * time.Second // fréquence d'envoi des pings
	writeWait      = 10 * time.Second // délai max pour écrire un message
	maxMessageSize = 512
)

// ReadPump gère la réception des messages et la déconnexion
func (c *ClientConn) ReadPump(
	handleMessage func(playerID entities.PlayerID, msg []byte),
	playerRepo ports.PlayerRepository,
	eventService ports.EventService,
	gameRepo ports.GameRepository,
	timerManager TimerManager,
) {
	defer func() {
		c.Conn.Close()
		c.handleDisconnection(playerRepo, eventService, gameRepo, timerManager)
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))

	// Prolonger la deadline si un PONG est reçu
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("read error from player %s: %v", c.PlayerID, err)
			}
			break
		}
		handleMessage(c.PlayerID, message)
	}
}

// WritePump gère l'envoi des messages et les pings
func (c *ClientConn) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Channel fermé
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("write error to player %s: %v", c.PlayerID, err)
				return
			}

		case <-ticker.C:
			// Ping périodique
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("ping to player %s failed: %v", c.PlayerID, err)
				return
			}
		}
	}
}

// handleDisconnection centralise la logique métier de déconnexion
func (c *ClientConn) handleDisconnection(
	playerRepo ports.PlayerRepository,
	eventService ports.EventService,
	gameRepo ports.GameRepository,
	timerManager TimerManager,
) {
	log.Printf("début de la déconnexion pour le joueur %s", c.PlayerID)

	player, err := playerRepo.GetPlayerByID(c.PlayerID)
	if err != nil {
		log.Printf("disconnect: player %s error: %v", c.PlayerID, err)
		return
	}

	gameID := player.GetGameID()
	if gameID == nil {
		log.Printf("disconnect: player %s not in a game", c.PlayerID)
		return
	}

	game, err := gameRepo.GetGameByID(*gameID)
	if err != nil {
		log.Printf("disconnect: could not fetch game %s: %v", *gameID, err)
		return
	}

	// Marquer le joueur comme déconnecté
	player.Disconnect()
	log.Printf("player %s marked as disconnected", c.PlayerID)

	// Envoyer l'événement de déconnexion
	event := events.NewDeconnexionEvent(c.PlayerID)
	eventService.SendEventToGame(event, *gameID)

	// Traitement selon le statut de la partie
	switch game.Status {
	case entities.GameStatusWaiting:
		c.handleWaitingGameDisconnection(playerRepo, gameRepo, game)

	case entities.GameStatusActive:
		c.handleActiveGameDisconnection(playerRepo, eventService, player, timerManager)

	default:
		log.Printf("disconnect: unhandled game status %v for game %s", game.Status, *gameID)
	}
}

// handleWaitingGameDisconnection gère la déconnexion pendant l'attente
func (c *ClientConn) handleWaitingGameDisconnection(
	playerRepo ports.PlayerRepository,
	gameRepo ports.GameRepository,
	game *entities.Game,
) {
	log.Printf("handling waiting game disconnection for player %s in game %s", c.PlayerID, game.ID)
	actualGame, err := gameRepo.GetGameByID(game.ID)
	if err != nil || actualGame == nil {
		log.Printf("waiting disconnect: game %s not found", game.ID)
		return
		// TODO: gérer ce problème
	}

	// Si c'est le dernier joueur, supprimer la partie et le joueur
	if len(game.Players) == 1 {
		log.Printf("last player leaving, deleting game %s", game.ID)
		_ = gameRepo.DeleteGame(game.ID)
		_ = playerRepo.DeletePlayer(c.PlayerID)
		return
	}

	// Si le host part, choisir un nouveau host
	if game.Host == c.PlayerID {
		for _, pid := range game.Players {
			if pid != c.PlayerID {
				game.Host = pid
				log.Printf("new host for game %s: %s", game.ID, pid)
				break
			}
		}
	}

	// Retirer le joueur de la liste des joueurs de la partie
	for i, pid := range game.Players {
		if pid == c.PlayerID {
			game.Players = append(game.Players[:i], game.Players[i+1:]...)
			break
		}
	}

	// Supprimer le joueur de la base de données
	_ = playerRepo.DeletePlayer(c.PlayerID)
	log.Printf("player %s removed from waiting game %s", c.PlayerID, game.ID)

}

// handleActiveGameDisconnection gère la déconnexion pendant une partie active
func (c *ClientConn) handleActiveGameDisconnection(
	playerRepo ports.PlayerRepository,
	eventService ports.EventService,
	player *entities.SafePlayer,
	timerManager TimerManager,
) {
	gameID := player.GetGameID()
	if gameID == nil {
		return
	}

	log.Printf("handling active game disconnection for player %s in game %s", c.PlayerID, *gameID)

	// Lancer un timer de 2 minutes pour marquer le joueur comme inactif s'il ne revient pas
	c.startInactivityTimer(c.PlayerID, *gameID, playerRepo, eventService, timerManager)
}

// startInactivityTimer démarre un timer d'inactivité avec possibilité d'annulation
func (c *ClientConn) startInactivityTimer(
	playerID entities.PlayerID,
	gameID entities.GameID,
	playerRepo ports.PlayerRepository,
	eventService ports.EventService,
	timerManager TimerManager,
) {
	log.Printf("starting inactivity timer for player %s", playerID)

	// Créer un contexte avec annulation
	ctx, cancel := context.WithCancel(context.Background())

	// Enregistrer la fonction d'annulation
	timerManager.setInactivityTimer(playerID, cancel)

	go func() {
		defer func() {
			// Nettoyer le timer de la map quand il se termine
			timerManager.cancelInactivityTimer(playerID)
		}()

		timer := time.NewTimer(2 * time.Minute)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			log.Printf("inactivity timer cancelled for player %s", playerID)
			return

		case <-timer.C:
			// Vérifier si le joueur est toujours déconnecté après 2 minutes
			player, err := playerRepo.GetPlayerByID(playerID)
			if err != nil {
				log.Printf("inactivity timer: player %s error: %v", playerID, err)
				return
			}

			if player.ConnectionState() != entities.Disconnected {
				log.Printf("inactivity timer: player %s reconnected during timer, should not happen", playerID)
				return
			}

			log.Printf("player %s marked as inactive after 2 minutes", playerID)

			// Marquer le joueur comme inactif
			event := events.NewInactiveEvent(playerID)
			eventService.SendEventToGame(event, gameID)
			player.SetInactive()

			// Exécuter le joueur inactif
			event = events.NewDeathEvent(nil, playerID)
			eventService.SendEventToGame(event, gameID)
			player.Kill()

			log.Printf("player %s executed due to inactivity", playerID)
		}
	}()
}

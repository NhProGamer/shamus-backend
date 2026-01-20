package ws

import (
	"encoding/json"
	"errors"
	"net/http"
	"shamus-backend/internal/adapters/app_adapters"
	"shamus-backend/internal/domain/entities/events"
	"sync"

	"shamus-backend/internal/domain/entities"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
)

// SessionManager gère les connexions actives en mémoire
type WebSocketHandler struct {
	melody  *melody.Melody
	service *app_adapters.GameService

	// Map Locale : GameID -> Liste de sessions (Pour le broadcast ciblé)
	rooms map[entities.GameID][]*melody.Session
	lock  sync.RWMutex
}

func NewWebSocketHandler(m *melody.Melody, s *app_adapters.GameService) *WebSocketHandler {
	handler := &WebSocketHandler{
		melody:  m,
		service: s,
		rooms:   make(map[entities.GameID][]*melody.Session),
	}

	// Configuration des événements Melody
	handler.setupEvents()
	return handler
}

// HandleWS : La route Gin GET /ws
func (h *WebSocketHandler) HandleWS(c *gin.Context) {
	gameIDStr := c.Param("gameID")
	if gameIDStr == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
	}
	userIDstr, exist := c.Get("userID")
	if !exist || userIDstr == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing user ID"})
		return
	}
	keys := map[string]interface{}{
		"user_id": userIDstr,
		"game_id": gameIDStr,
	}
	err := h.melody.HandleRequestWithKeys(c.Writer, c.Request, keys)
	if err != nil {
		c.Status(http.StatusInternalServerError)
	}
}

func (h *WebSocketHandler) setupEvents() {

	// 1. Connexion (Join Game)
	h.melody.HandleConnect(func(s *melody.Session) {
		gameIDstr, exist := s.Get("game_id")
		if !exist || gameIDstr == "" {
			s.CloseWithMsg([]byte("missing game_id"))
			return
		}
		userIDstr, exist := s.Get("user_id")
		if !exist || userIDstr == "" {
			s.CloseWithMsg([]byte("missing user_id"))
			return
		}
		gameID := entities.GameID(gameIDstr.(string))
		playerID := entities.PlayerID(userIDstr.(string))

		if gameID == "" || playerID == "" {
			s.CloseWithMsg([]byte("missing parameters"))
			return
		}

		// Appel au Service pour la logique métier (Vérif existence + Ajout DB)
		updatedGame, err := h.service.JoinGame(gameID, playerID)
		if err != nil {
			s.CloseWithMsg([]byte(err.Error()))
			return
		}

		// Stockage info session
		s.Set("gameId", gameID)
		s.Set("playerId", playerID)

		// Ajout à la Room Locale (Mémoire)
		h.joinLocalRoom(gameID, s)

		// Broadcast à tous les joueurs de la room : "Nouveau joueur + État à jour"
		h.broadcastGameState(gameID, updatedGame)
	})

	// 2. Déconnexion
	h.melody.HandleDisconnect(func(s *melody.Session) {
		val, exists := s.Get("gameId")
		if exists {
			h.leaveLocalRoom(val.(entities.GameID), s)
		}
	})

	// 3. Messages (Gameplay)
	h.melody.HandleMessage(func(s *melody.Session, msg []byte) {
		var event entities.RawEvent
		err := json.Unmarshal(msg, &event) // msg est []byte du WebSocket
		if err != nil {
			s.Write([]byte("Invalid JSON"))
			return
		}
		switch event.Channel {
		case entities.EventChannelGameEvent:
			switch event.Type {
			case events.EventTypeChatMessage:
				var message events.ChatMessageEvent
				if err := json.Unmarshal(event.Data, &message); err != nil {
					panic(err)
				}
				gameIDstr, exist := s.Get("game_id")
				if !exist || gameIDstr == "" {
					s.CloseWithMsg([]byte("missing game_id"))
					return
				}
				h.broadcastToRoom(entities.GameID(gameIDstr.(string)), msg)

			}
		}
	})
}

// --- Helpers de gestion de Room (Thread-Safe) ---

func (h *WebSocketHandler) joinLocalRoom(gid entities.GameID, s *melody.Session) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.rooms[gid] = append(h.rooms[gid], s)
}

func (h *WebSocketHandler) leaveLocalRoom(gid entities.GameID, s *melody.Session) {
	h.lock.Lock()
	defer h.lock.Unlock()

	sessions := h.rooms[gid]
	for i, sess := range sessions {
		if sess == s {
			// Suppression rapide
			h.rooms[gid] = append(sessions[:i], sessions[i+1:]...)
			break
		}
	}
}

func (h *WebSocketHandler) broadcastGameState(gid entities.GameID, game *entities.Game) {
	h.lock.RLock() // Read lock suffisant
	defer h.lock.RUnlock()

	sessions, ok := h.rooms[gid]
	if !ok {
		return
	}

	// Préparer le payload JSON une seule fois
	payload, _ := json.Marshal(events.NewGameDataEvent(events.GameDataEventData{
		ID:     game.ID,
		Status: game.Status,
		Phase:  game.Phase,
		Day:    game.Day,
		//Players: game.Players,
		Host:     game.HostID,
		Settings: game.Settings,
	}))

	for _, s := range sessions {
		s.Write(payload)
	}
}

func (h *WebSocketHandler) broadcastToRoom(gameID entities.GameID, msg []byte) error {
	h.lock.RLock()
	defer h.lock.RUnlock()

	sessions, ok := h.rooms[gameID]
	if !ok || len(sessions) == 0 {
		return errors.New("empty room")
	}

	for _, sess := range sessions {
		_ = sess.Write(msg)
	}
	return nil
}

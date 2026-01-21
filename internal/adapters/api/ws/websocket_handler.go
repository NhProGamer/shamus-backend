package ws

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/events"
	"shamus-backend/internal/domain/ports"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
)

// WebSocketHandler manages WebSocket connections for games
type WebSocketHandler struct {
	melody        *melody.Melody
	gameService   ports.GameService
	playerService ports.PlayerService

	// rooms maps GameID to list of sessions for targeted broadcast
	rooms map[entities.GameID][]*melody.Session
	// playerSessions maps PlayerID to session for O(1) player lookup
	playerSessions map[entities.PlayerID]*melody.Session
	lock           sync.RWMutex
}

func NewWebSocketHandler(m *melody.Melody, gameService ports.GameService) *WebSocketHandler {
	handler := &WebSocketHandler{
		melody:         m,
		gameService:    gameService,
		rooms:          make(map[entities.GameID][]*melody.Session),
		playerSessions: make(map[entities.PlayerID]*melody.Session),
	}

	return handler
}

// SetPlayerService sets the player service (used to break circular dependency)
func (h *WebSocketHandler) SetPlayerService(ps ports.PlayerService) {
	h.playerService = ps
	h.setupEvents()
}

// IsPlayerConnected checks if a player has an active WebSocket session
func (h *WebSocketHandler) IsPlayerConnected(playerID entities.PlayerID) bool {
	h.lock.RLock()
	defer h.lock.RUnlock()
	_, exists := h.playerSessions[playerID]
	return exists
}

// HandleWS handles GET /ws/:gameID requests
func (h *WebSocketHandler) HandleWS(c *gin.Context) {
	gameIDStr := c.Param("gameID")
	if gameIDStr == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}
	userIDStr, exist := c.Get("userID")
	if !exist || userIDStr == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing user ID"})
		return
	}
	userInfoRaw, exist := c.Get("userInfo")
	if !exist || userInfoRaw == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing user info"})
		return
	}
	keys := map[string]interface{}{
		"userId":   userIDStr,
		"gameId":   gameIDStr,
		"userInfo": userInfoRaw,
	}
	err := h.melody.HandleRequestWithKeys(c.Writer, c.Request, keys)
	if err != nil {
		c.Status(http.StatusInternalServerError)
	}
}

func (h *WebSocketHandler) setupEvents() {
	// Handle connection (Join Game)
	h.melody.HandleConnect(func(s *melody.Session) {
		gameIDStr, exist := s.Get("gameId")
		if !exist || gameIDStr == "" {
			s.CloseWithMsg([]byte("missing gameId"))
			return
		}
		userIDStr, exist := s.Get("userId")
		if !exist || userIDStr == "" {
			s.CloseWithMsg([]byte("missing userId"))
			return
		}
		userInfoRaw, exist := s.Get("userInfo")
		if !exist || userInfoRaw == nil {
			s.CloseWithMsg([]byte("missing userInfo"))
			return
		}

		gameID := entities.GameID(gameIDStr.(string))
		playerID := entities.PlayerID(userIDStr.(string))

		if gameID == "" || playerID == "" {
			s.CloseWithMsg([]byte("missing parameters"))
			return
		}

		// Extract username from OIDC token
		userInfo := userInfoRaw.(oidc.UserInfo)
		var claims struct {
			Username string `json:"preferred_username"`
		}
		if err := userInfo.Claims(&claims); err != nil {
			s.CloseWithMsg([]byte("invalid user info"))
			return
		}
		username := claims.Username
		if username == "" {
			username = string(playerID) // Fallback to playerID if no username
		}

		// Call PlayerService for business logic
		player, isReconnection, err := h.playerService.HandleConnect(gameID, playerID, username)
		if err != nil {
			s.CloseWithMsg([]byte(err.Error()))
			return
		}

		// Add to local room (in-memory)
		h.joinLocalRoom(gameID, playerID, s)

		// Broadcast connection/reconnection event
		var connEvent interface{}
		if isReconnection {
			connEvent = events.NewReconnectionEvent(playerID)
		} else {
			connEvent = events.NewConnectionEvent(playerID)
		}
		connPayload, _ := json.Marshal(connEvent)
		h.broadcastToRoom(gameID, connPayload)

		// Broadcast updated game state
		game, err := h.gameService.GetGame(gameID)
		if err == nil {
			h.broadcastGameState(gameID, game)
		}

		log.Printf("Player %s (%s) joined game %s (reconnection: %v)", player.Username, playerID, gameID, isReconnection)
	})

	// Handle disconnection
	h.melody.HandleDisconnect(func(s *melody.Session) {
		gameIDVal, gameExists := s.Get("gameId")
		userIDVal, userExists := s.Get("userId")

		if !gameExists || !userExists {
			return
		}

		gameID := entities.GameID(gameIDVal.(string))
		playerID := entities.PlayerID(userIDVal.(string))

		// Remove from local room first
		h.leaveLocalRoom(gameID, playerID, s)

		// Call PlayerService for business logic
		if err := h.playerService.HandleDisconnect(gameID, playerID); err != nil {
			log.Printf("Error handling disconnect for player %s: %v", playerID, err)
		}

		// Broadcast disconnection event
		disconnEvent := events.NewDisconnectionEvent(playerID)
		disconnPayload, _ := json.Marshal(disconnEvent)
		h.broadcastToRoom(gameID, disconnPayload)

		log.Printf("Player %s disconnected from game %s", playerID, gameID)
	})

	// Handle messages (Gameplay)
	h.melody.HandleMessage(func(s *melody.Session, msg []byte) {
		userInfoRaw, exist := s.Get("userInfo")
		if !exist || userInfoRaw == nil {
			s.CloseWithMsg([]byte("missing userInfo"))
			return
		}
		userInfo := userInfoRaw.(oidc.UserInfo)
		var event entities.RawEvent
		if err := json.Unmarshal(msg, &event); err != nil {
			s.Write([]byte("Invalid JSON"))
			return
		}

		// Extract common session data
		gameIDStr, exist := s.Get("gameId")
		if !exist || gameIDStr == "" {
			s.CloseWithMsg([]byte("missing gameId"))
			return
		}
		userIDStr, exist := s.Get("userId")
		if !exist || userIDStr == "" {
			s.CloseWithMsg([]byte("missing userId"))
			return
		}
		gameID := entities.GameID(gameIDStr.(string))
		playerID := entities.PlayerID(userIDStr.(string))

		switch event.Channel {
		case entities.EventChannelGameEvent:
			switch event.Type {
			case events.EventTypeChatMessage:
				var message events.ChatMessageEventData
				if err := json.Unmarshal(event.Data, &message); err != nil {
					log.Printf("Failed to unmarshal chat message: %v", err)
					s.Write([]byte("Invalid message format"))
					return
				}
				var claims struct {
					Username string `json:"preferred_username"`
				}
				userInfo.Claims(&claims)
				nickname := claims.Username
				reforgedEvent := events.NewChatMessageEvent(message.PlayerID, nickname, message.Message, message.Channel)
				reforgedMsg, err := json.Marshal(reforgedEvent)
				if err != nil {
					s.Write([]byte("Error processing message"))
					return
				}
				h.broadcastToRoom(gameID, reforgedMsg)
			}

		case entities.EventChannelSettings:
			switch event.Type {
			case events.EventTypeGameSettings:
				var settingsData events.GameSettingsEventData
				if err := json.Unmarshal(event.Data, &settingsData); err != nil {
					log.Printf("Failed to unmarshal settings: %v", err)
					s.Write([]byte("Invalid settings format"))
					return
				}

				// Build GameSettings from event data
				newSettings := entities.GameSettings{
					Roles: settingsData.RolesType,
				}

				// Call service to update settings (validates host + game state + roles)
				updatedGame, err := h.gameService.UpdateSettings(gameID, playerID, newSettings)
				if err != nil {
					s.Write([]byte(err.Error()))
					return
				}

				// Broadcast new settings to all players in the game
				settingsEvent := events.NewGameSettingsEvent(events.GameSettingsEventData{
					RolesType: updatedGame.Settings.Roles,
				})
				payload, _ := json.Marshal(settingsEvent)
				h.broadcastToRoom(gameID, payload)

				log.Printf("Host %s updated settings for game %s: %v", playerID, gameID, updatedGame.Settings.Roles)
			}
		}
	})
}

// joinLocalRoom adds a session to a game room (thread-safe)
func (h *WebSocketHandler) joinLocalRoom(gid entities.GameID, playerID entities.PlayerID, s *melody.Session) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.rooms[gid] = append(h.rooms[gid], s)
	h.playerSessions[playerID] = s
}

// leaveLocalRoom removes a session from a game room (thread-safe)
func (h *WebSocketHandler) leaveLocalRoom(gid entities.GameID, playerID entities.PlayerID, s *melody.Session) {
	h.lock.Lock()
	defer h.lock.Unlock()

	sessions := h.rooms[gid]
	for i, sess := range sessions {
		if sess == s {
			h.rooms[gid] = append(sessions[:i], sessions[i+1:]...)
			break
		}
	}
	delete(h.playerSessions, playerID)
}

// broadcastGameState sends game state to all players in a room
func (h *WebSocketHandler) broadcastGameState(gid entities.GameID, game *entities.Game) {
	h.lock.RLock()
	defer h.lock.RUnlock()

	sessions, ok := h.rooms[gid]
	if !ok {
		return
	}

	payload, _ := json.Marshal(events.NewGameDataEvent(events.GameDataEventData{
		ID:       game.ID,
		Status:   game.Status,
		Phase:    game.Phase,
		Day:      game.Day,
		Host:     game.HostID,
		Settings: game.Settings,
	}))

	for _, s := range sessions {
		s.Write(payload)
	}
}

// broadcastToRoom sends a message to all sessions in a room
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

// SendToPlayer sends a message to a specific player (public API for EventService)
func (h *WebSocketHandler) SendToPlayer(playerID entities.PlayerID, payload []byte) error {
	h.lock.RLock()
	defer h.lock.RUnlock()

	sess, ok := h.playerSessions[playerID]
	if !ok {
		return errors.New("player not connected")
	}
	return sess.Write(payload)
}

// BroadcastToGame sends a message to all players in a game (public API for EventService)
func (h *WebSocketHandler) BroadcastToGame(gameID entities.GameID, payload []byte) error {
	return h.broadcastToRoom(gameID, payload)
}

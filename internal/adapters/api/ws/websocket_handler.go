package ws

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/events"
	apperrors "shamus-backend/internal/domain/errors"
	"shamus-backend/internal/domain/ports"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
)

// WebSocketHandler manages WebSocket connections for games
type WebSocketHandler struct {
	melody            *melody.Melody
	gameService       ports.GameService
	playerService     ports.PlayerService
	visibilityService ports.VisibilityService
	chatService       ports.ChatService
	gameEngine        ports.GameEngine

	// rooms maps GameID to list of sessions for targeted broadcast
	rooms map[entities.GameID][]*melody.Session
	// playerSessions maps PlayerID to session for O(1) player lookup
	playerSessions map[entities.PlayerID]*melody.Session
	lock           sync.RWMutex
}

func NewWebSocketHandler(m *melody.Melody, gameService ports.GameService, visibilityService ports.VisibilityService, chatService ports.ChatService) *WebSocketHandler {
	handler := &WebSocketHandler{
		melody:            m,
		gameService:       gameService,
		visibilityService: visibilityService,
		chatService:       chatService,
		rooms:             make(map[entities.GameID][]*melody.Session),
		playerSessions:    make(map[entities.PlayerID]*melody.Session),
	}

	return handler
}

// SetPlayerService sets the player service (used to break circular dependency)
func (h *WebSocketHandler) SetPlayerService(ps ports.PlayerService) {
	h.playerService = ps
	h.setupEvents()
}

// SetGameEngine sets the game engine (used to break circular dependency)
func (h *WebSocketHandler) SetGameEngine(ge ports.GameEngine) {
	h.gameEngine = ge
}

// IsPlayerConnected checks if a player has an active WebSocket session
func (h *WebSocketHandler) IsPlayerConnected(playerID entities.PlayerID) bool {
	h.lock.RLock()
	defer h.lock.RUnlock()
	_, exists := h.playerSessions[playerID]
	return exists
}

// errorToCode maps application errors to structured error codes
func errorToCode(err error) events.ErrorCode {
	switch {
	case errors.Is(err, apperrors.ErrWrongPhase):
		return events.ErrorCodeWrongPhase
	case errors.Is(err, apperrors.ErrNotYourTurn):
		return events.ErrorCodeNotYourTurn
	case errors.Is(err, apperrors.ErrAlreadyActed):
		return events.ErrorCodeAlreadyActed
	case errors.Is(err, apperrors.ErrGameNotActive):
		return events.ErrorCodeGameNotActive
	case errors.Is(err, apperrors.ErrPlayerDead):
		return events.ErrorCodePlayerDead
	case errors.Is(err, apperrors.ErrWrongRole):
		return events.ErrorCodeWrongRole
	case errors.Is(err, apperrors.ErrInvalidTarget):
		return events.ErrorCodeInvalidTarget
	case errors.Is(err, apperrors.ErrTargetDead):
		return events.ErrorCodeTargetDead
	case errors.Is(err, apperrors.ErrCannotTargetSelf):
		return events.ErrorCodeCannotTargetSelf
	case errors.Is(err, apperrors.ErrAbilityUsed):
		return events.ErrorCodeAbilityUsed
	case errors.Is(err, apperrors.ErrCanOnlyHealVictim):
		return events.ErrorCodeCanOnlyHealVictim
	case errors.Is(err, apperrors.ErrVoteNotFound):
		return events.ErrorCodeVoteNotFound
	case errors.Is(err, apperrors.ErrVoteNotActive):
		return events.ErrorCodeVoteNotActive
	case errors.Is(err, apperrors.ErrInvalidVoter):
		return events.ErrorCodeInvalidVoter
	default:
		return events.ErrorCodeUnknown
	}
}

// sendError sends a structured error event to the session
func sendError(s *melody.Session, err error, action string) {
	code := errorToCode(err)
	event := events.NewErrorEvent(code, err.Error(), action)
	payload, marshalErr := json.Marshal(event)
	if marshalErr != nil {
		s.Write([]byte(err.Error()))
		return
	}
	s.Write(payload)
}

// sendErrorMessage sends a structured error event with a custom message
func sendErrorMessage(s *melody.Session, code events.ErrorCode, message string, action string) {
	event := events.NewErrorEvent(code, message, action)
	payload, err := json.Marshal(event)
	if err != nil {
		s.Write([]byte(message))
		return
	}
	s.Write(payload)
}

// sendAck sends an acknowledgement event to the session
func sendAck(s *melody.Session, action string, success bool, message string) {
	event := events.NewAckEvent(action, success, message)
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	s.Write(payload)
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

		// Send personalized game state to each player
		game, err := h.gameService.GetGame(gameID)
		if err == nil {
			h.sendPersonalizedGameStateToAll(gameID, game)
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
			sendErrorMessage(s, events.ErrorCodeInvalidAction, "Invalid JSON", "")
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
					sendErrorMessage(s, events.ErrorCodeInvalidAction, "Invalid message format", "chat_message")
					return
				}

				// Validate chat channel
				if !events.IsValidChatChannel(string(message.Channel)) {
					sendErrorMessage(s, events.ErrorCodeInvalidAction, "Invalid chat channel", "chat_message")
					return
				}

				// Validate message content
				if len(message.Message) == 0 {
					sendErrorMessage(s, events.ErrorCodeInvalidAction, "Message cannot be empty", "chat_message")
					return
				}
				const maxMessageLength = 500
				if len(message.Message) > maxMessageLength {
					sendErrorMessage(s, events.ErrorCodeInvalidAction, "Message is too long", "chat_message")
					return
				}

				// Get game state for phase info
				game, err := h.gameService.GetGame(gameID)
				if err != nil {
					sendErrorMessage(s, events.ErrorCodeUnknown, "Game not found", "chat_message")
					return
				}

				// Get sender player info
				sender, err := h.playerService.GetPlayer(playerID)
				if err != nil {
					sendErrorMessage(s, events.ErrorCodeUnknown, "Player not found", "chat_message")
					return
				}

				// Check if sender can send to this channel
				if !h.chatService.CanSendToChannel(sender, message.Channel, game.Phase) {
					sendErrorMessage(s, events.ErrorCodeWrongPhase, "You cannot send messages to this channel", "chat_message")
					return
				}

				// Get all players for recipient filtering
				allPlayers, err := h.playerService.GetGamePlayers(gameID)
				if err != nil {
					sendErrorMessage(s, events.ErrorCodeUnknown, "Failed to get players", "chat_message")
					return
				}

				// Get recipients for this channel
				recipients := h.chatService.GetChannelRecipients(message.Channel, allPlayers, game.Phase)

				// Build the chat message event
				var claims struct {
					Username string `json:"preferred_username"`
				}
				userInfo.Claims(&claims)
				nickname := claims.Username
				reforgedEvent := events.NewChatMessageEvent(playerID, nickname, message.Message, message.Channel)
				reforgedMsg, err := json.Marshal(reforgedEvent)
				if err != nil {
					sendErrorMessage(s, events.ErrorCodeUnknown, "Error processing message", "chat_message")
					return
				}

				// Send to recipients only
				h.sendToPlayers(recipients, reforgedMsg)

			case events.EventTypeStartGame:
				// Host wants to start the game
				game, players, err := h.gameService.StartGame(gameID, playerID)
				if err != nil {
					sendError(s, err, "start_game")
					return
				}

				// Send personalized role reveal to each player
				for _, player := range players {
					if player.Role == nil {
						continue
					}
					roleEvent := events.NewRoleAttributionEvent(player.Role.GetType())
					rolePayload, err := json.Marshal(roleEvent)
					if err != nil {
						continue
					}
					h.SendToPlayer(player.ID, rolePayload)
				}

				// Broadcast night event to all players
				nightEvent := events.NewNightEvent()
				nightPayload, _ := json.Marshal(nightEvent)
				h.broadcastToRoom(gameID, nightPayload)

				// Send personalized game state to all players
				h.sendPersonalizedGameStateToAll(gameID, game)

				// Start the game flow (triggers first night phase)
				if h.gameEngine != nil {
					if err := h.gameEngine.StartGameFlow(gameID); err != nil {
						log.Printf("Error starting game flow for game %s: %v", gameID, err)
					}
				}

				log.Printf("Game %s started by host %s", gameID, playerID)

			case events.EventTypeVillageVote:
				// Player wants to vote during village phase
				if h.gameEngine == nil {
					sendErrorMessage(s, events.ErrorCodeUnknown, "Game engine not available", "village_vote")
					return
				}

				var voteData events.VillageVoteEventData
				if err := json.Unmarshal(event.Data, &voteData); err != nil {
					log.Printf("Failed to unmarshal village vote: %v", err)
					sendErrorMessage(s, events.ErrorCodeInvalidAction, "Invalid vote format", "village_vote")
					return
				}

				if err := h.gameEngine.HandleVillageVote(gameID, playerID, voteData.TargetID); err != nil {
					sendError(s, err, "village_vote")
					return
				}

				sendAck(s, "village_vote", true, "")
				log.Printf("Player %s voted for %v in game %s", playerID, voteData.TargetID, gameID)

			case events.EventTypeSeerAction:
				// Seer wants to see a player's role
				if h.gameEngine == nil {
					sendErrorMessage(s, events.ErrorCodeUnknown, "Game engine not available", "seer_action")
					return
				}

				var seerData events.SeerActionEventData
				if err := json.Unmarshal(event.Data, &seerData); err != nil {
					log.Printf("Failed to unmarshal seer action: %v", err)
					sendErrorMessage(s, events.ErrorCodeInvalidAction, "Invalid action format", "seer_action")
					return
				}

				if err := h.gameEngine.HandleSeerAction(gameID, playerID, seerData.TargetID); err != nil {
					sendError(s, err, "seer_action")
					return
				}

				log.Printf("Seer %s looked at %s in game %s", playerID, seerData.TargetID, gameID)

			case events.EventTypeWerewolfVote:
				// Werewolf wants to vote for a victim
				if h.gameEngine == nil {
					sendErrorMessage(s, events.ErrorCodeUnknown, "Game engine not available", "werewolf_vote")
					return
				}

				var wolfData events.WerewolfVoteEventData
				if err := json.Unmarshal(event.Data, &wolfData); err != nil {
					log.Printf("Failed to unmarshal werewolf vote: %v", err)
					sendErrorMessage(s, events.ErrorCodeInvalidAction, "Invalid vote format", "werewolf_vote")
					return
				}

				if err := h.gameEngine.HandleWerewolfVote(gameID, playerID, wolfData.TargetID); err != nil {
					sendError(s, err, "werewolf_vote")
					return
				}

				sendAck(s, "werewolf_vote", true, "")
				log.Printf("Werewolf %s voted for %v in game %s", playerID, wolfData.TargetID, gameID)

			case events.EventTypeWitchAction:
				// Witch wants to heal or poison
				if h.gameEngine == nil {
					sendErrorMessage(s, events.ErrorCodeUnknown, "Game engine not available", "witch_action")
					return
				}

				var witchData events.WitchActionEventData
				if err := json.Unmarshal(event.Data, &witchData); err != nil {
					log.Printf("Failed to unmarshal witch action: %v", err)
					sendErrorMessage(s, events.ErrorCodeInvalidAction, "Invalid action format", "witch_action")
					return
				}

				if err := h.gameEngine.HandleWitchAction(gameID, playerID, witchData.HealTargetID, witchData.PoisonTargetID); err != nil {
					sendError(s, err, "witch_action")
					return
				}

				sendAck(s, "witch_action", true, "")
				log.Printf("Witch %s acted (heal: %v, poison: %v) in game %s", playerID, witchData.HealTargetID, witchData.PoisonTargetID, gameID)
			}

		case entities.EventChannelSettings:
			switch event.Type {
			case events.EventTypeGameSettings:
				var settingsData events.GameSettingsEventData
				if err := json.Unmarshal(event.Data, &settingsData); err != nil {
					log.Printf("Failed to unmarshal settings: %v", err)
					sendErrorMessage(s, events.ErrorCodeInvalidAction, "Invalid settings format", "game_settings")
					return
				}

				// Build GameSettings from event data
				newSettings := entities.GameSettings{
					Roles: settingsData.RolesType,
				}

				// Call service to update settings (validates host + game state + roles)
				updatedGame, err := h.gameService.UpdateSettings(gameID, playerID, newSettings)
				if err != nil {
					sendError(s, err, "game_settings")
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

// sendPersonalizedGameStateToAll sends personalized game state to each player in a room
// Each player receives a GameDataEventData with role visibility based on their perspective
func (h *WebSocketHandler) sendPersonalizedGameStateToAll(gid entities.GameID, game *entities.Game) {
	// Get all players with their full data (including roles)
	players, err := h.playerService.GetGamePlayers(gid)
	if err != nil {
		log.Printf("Failed to get players for game %s: %v", gid, err)
		return
	}

	// Phase 1: Build all payloads WITHOUT holding the lock
	// This is the expensive operation (visibility calculation + JSON marshaling)
	type playerPayload struct {
		playerID entities.PlayerID
		payload  []byte
	}
	payloads := make([]playerPayload, 0, len(players))

	for _, player := range players {
		// Build personalized player details for this viewer
		playersDetails := h.visibilityService.BuildPlayersDetailsForPlayer(
			player, players, game.Phase,
		)

		event := events.NewGameDataEvent(events.GameDataEventData{
			ID:       game.ID,
			Status:   game.Status,
			Phase:    game.Phase,
			Day:      game.Day,
			Host:     game.HostID,
			Settings: game.Settings,
			Players:  playersDetails,
		})

		payload, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal game state for player %s: %v", player.ID, err)
			continue
		}

		payloads = append(payloads, playerPayload{
			playerID: player.ID,
			payload:  payload,
		})
	}

	// Phase 2: Copy session references while holding the lock briefly
	h.lock.RLock()
	sessionsToSend := make([]struct {
		sess    *melody.Session
		payload []byte
	}, 0, len(payloads))

	for _, pp := range payloads {
		if sess, ok := h.playerSessions[pp.playerID]; ok {
			sessionsToSend = append(sessionsToSend, struct {
				sess    *melody.Session
				payload []byte
			}{sess: sess, payload: pp.payload})
		}
	}
	h.lock.RUnlock()

	// Phase 3: Send all messages WITHOUT holding the lock
	for _, item := range sessionsToSend {
		item.sess.Write(item.payload)
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

// sendToPlayers sends a message to a specific list of players
func (h *WebSocketHandler) sendToPlayers(players []*entities.Player, payload []byte) {
	h.lock.RLock()
	defer h.lock.RUnlock()

	for _, p := range players {
		if sess, ok := h.playerSessions[p.ID]; ok {
			sess.Write(payload)
		}
	}
}

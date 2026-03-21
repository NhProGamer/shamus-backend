package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/prompts"
	apperrors "shamus-backend/internal/domain/errors"
	"shamus-backend/internal/domain/ports"
	"shamus-backend/pkg/logger"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
)

// Default timeout for WebSocket operations
const wsOpTimeout = 30 * time.Second

// Handler is the new simplified WebSocket handler using Notification/Prompt/Command architecture
type Handler struct {
	melody         *melody.Melody
	sessions       *SessionManager
	promptService  ports.PromptService
	commandHandler *CommandHandler
	notifier       ports.NotificationService
	playerService  ports.PlayerService
	gameService    ports.GameService
}

// NewHandler creates a new WebSocket handler
func NewHandler(
	m *melody.Melody,
	sessions *SessionManager,
	promptService ports.PromptService,
	commandHandler *CommandHandler,
	notifier ports.NotificationService,
	gameService ports.GameService,
) *Handler {
	h := &Handler{
		melody:         m,
		sessions:       sessions,
		promptService:  promptService,
		commandHandler: commandHandler,
		notifier:       notifier,
		gameService:    gameService,
	}
	return h
}

// SetPlayerService sets the player service (used to break circular dependency)
func (h *Handler) SetPlayerService(ps ports.PlayerService) {
	h.playerService = ps
	h.setupEvents()
}

// HandleWS handles the WebSocket upgrade request
func (h *Handler) HandleWS(c *gin.Context) {
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

	if err := h.melody.HandleRequestWithKeys(c.Writer, c.Request, keys); err != nil {
		c.Status(http.StatusInternalServerError)
	}
}

// setupEvents configures Melody event handlers
func (h *Handler) setupEvents() {
	h.melody.HandleConnect(h.onConnect)
	h.melody.HandleDisconnect(h.onDisconnect)
	h.melody.HandleMessage(h.onMessage)
}

// onConnect handles new WebSocket connections
func (h *Handler) onConnect(s *melody.Session) {
	// Create context with timeout for this operation
	ctx, cancel := context.WithTimeout(context.Background(), wsOpTimeout)
	defer cancel()

	// Extract session data
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
		username = string(playerID)
	}

	// Call PlayerService for business logic
	player, isReconnection, err := h.playerService.HandleConnect(ctx, gameID, playerID, username)
	if err != nil {
		s.CloseWithMsg([]byte(err.Error()))
		return
	}

	// Add to session manager
	h.sessions.JoinRoom(gameID, playerID, s)

	// Notify other players
	if isReconnection {
		h.notifier.NotifyAll(gameID, entities.NotificationType("player_reconnected"), map[string]interface{}{
			"playerId": playerID,
			"username": player.Username,
		})
	} else {
		h.notifier.NotifyPlayerJoined(gameID, playerID, player.Username)
	}

	// Send game state to the connecting player
	h.sendGameStateToPlayer(ctx, gameID, playerID)

	logger.Get().Info().
		Str("playerID", string(playerID)).
		Str("username", player.Username).
		Str("gameID", string(gameID)).
		Bool("reconnection", isReconnection).
		Msg("Player connected")
}

// onDisconnect handles WebSocket disconnections
func (h *Handler) onDisconnect(s *melody.Session) {
	// Create context with timeout for this operation
	ctx, cancel := context.WithTimeout(context.Background(), wsOpTimeout)
	defer cancel()

	gameIDVal, gameExists := s.Get("gameId")
	userIDVal, userExists := s.Get("userId")

	if !gameExists || !userExists {
		return
	}

	gameID := entities.GameID(gameIDVal.(string))
	playerID := entities.PlayerID(userIDVal.(string))

	// Get player info before removal for notification
	player, _ := h.playerService.GetPlayer(ctx, playerID)
	username := ""
	if player != nil {
		username = player.Username
	}

	// Remove from session manager
	h.sessions.LeaveRoom(gameID, playerID, s)

	// Call PlayerService for business logic
	if err := h.playerService.HandleDisconnect(ctx, gameID, playerID); err != nil {
		logger.Get().Warn().
			Str("playerID", string(playerID)).
			Err(err).
			Msg("Error handling disconnect")
	}

	// Notify other players
	h.notifier.NotifyPlayerLeft(gameID, playerID, username, "disconnected")

	logger.Get().Info().
		Str("playerID", string(playerID)).
		Str("gameID", string(gameID)).
		Msg("Player disconnected")
}

// onMessage handles incoming WebSocket messages
func (h *Handler) onMessage(s *melody.Session, msg []byte) {
	// Create context with timeout for this message
	ctx, cancel := context.WithTimeout(context.Background(), wsOpTimeout)
	defer cancel()

	// Extract session context
	gameIDStr, _ := s.Get("gameId")
	userIDStr, _ := s.Get("userId")
	userInfoRaw, _ := s.Get("userInfo")

	if gameIDStr == nil || userIDStr == nil {
		h.sendError(s, "MISSING_CONTEXT", "Session context missing")
		return
	}

	gameID := entities.GameID(gameIDStr.(string))
	playerID := entities.PlayerID(userIDStr.(string))

	// Extract username
	username := string(playerID)
	if userInfoRaw != nil {
		userInfo := userInfoRaw.(oidc.UserInfo)
		var claims struct {
			Username string `json:"preferred_username"`
		}
		if err := userInfo.Claims(&claims); err == nil && claims.Username != "" {
			username = claims.Username
		}
	}

	// Parse channel from message
	var envelope struct {
		Channel entities.Channel `json:"channel"`
	}
	if err := json.Unmarshal(msg, &envelope); err != nil {
		h.sendError(s, "INVALID_JSON", "Invalid message format")
		return
	}

	// Route based on channel
	switch envelope.Channel {
	case entities.ChannelResponse:
		h.handleResponse(ctx, s, playerID, msg)

	case entities.ChannelCommand:
		h.handleCommand(ctx, s, gameID, playerID, username, msg)

	default:
		h.sendError(s, "UNKNOWN_CHANNEL", "Unknown message channel: "+string(envelope.Channel))
	}
}

// handleResponse processes a prompt response from a client
func (h *Handler) handleResponse(ctx context.Context, s *melody.Session, playerID entities.PlayerID, msg []byte) {
	var response prompts.PromptResponse
	if err := json.Unmarshal(msg, &response); err != nil {
		h.sendError(s, "INVALID_RESPONSE", "Invalid response format")
		return
	}

	if err := h.promptService.RespondToPrompt(ctx, response.PromptID, playerID, response.Response); err != nil {
		// Map errors to appropriate codes
		code := "RESPONSE_ERROR"
		switch err {
		case apperrors.ErrPromptNotFound:
			code = "PROMPT_NOT_FOUND"
		case apperrors.ErrPromptWrongPlayer:
			code = "WRONG_PLAYER"
		case apperrors.ErrPromptExpired:
			code = "PROMPT_EXPIRED"
		case apperrors.ErrPromptAlreadyAnswered:
			code = "ALREADY_ANSWERED"
		}
		h.sendError(s, code, err.Error())
		return
	}

	// Send acknowledgement
	h.sendAck(s, "response", true, "")
}

// handleCommand processes a command from a client
func (h *Handler) handleCommand(ctx context.Context, s *melody.Session, gameID entities.GameID, playerID entities.PlayerID, username string, msg []byte) {
	cmd, err := entities.ParseCommand(msg)
	if err != nil {
		h.sendError(s, "INVALID_COMMAND", "Invalid command format")
		return
	}

	cmdCtx := &CommandContext{
		Ctx:      ctx,
		GameID:   gameID,
		PlayerID: playerID,
		Username: username,
	}

	if err := h.commandHandler.Handle(cmdCtx, cmd); err != nil {
		// Map errors to codes
		code := "COMMAND_ERROR"
		switch err {
		case ErrUnknownCommand:
			code = "UNKNOWN_COMMAND"
		case ErrInvalidPayload:
			code = "INVALID_PAYLOAD"
		case ErrNotHost:
			code = "NOT_HOST"
		case ErrGameNotWaiting:
			code = "GAME_NOT_WAITING"
		case ErrChatMessageEmpty:
			code = "MESSAGE_EMPTY"
		case ErrChatMessageTooLong:
			code = "MESSAGE_TOO_LONG"
		case ErrCannotSendToChannel:
			code = "CHANNEL_FORBIDDEN"
		}
		h.sendError(s, code, err.Error())
		return
	}

	// Send acknowledgement
	h.sendAck(s, string(cmd.Type), true, "")
}

// sendGameStateToPlayer sends the current game state to a player
func (h *Handler) sendGameStateToPlayer(ctx context.Context, gameID entities.GameID, playerID entities.PlayerID) {
	game, err := h.gameService.GetGame(ctx, gameID)
	if err != nil {
		return
	}

	players, err := h.playerService.GetGamePlayers(ctx, gameID)
	if err != nil {
		return
	}

	// Build player state info
	// TODO: Apply visibility rules based on viewer's role
	var playersInfo []map[string]interface{}
	for _, p := range players {
		info := map[string]interface{}{
			"id":              p.ID,
			"username":        p.Username,
			"isAlive":         p.IsAlive,
			"connectionState": p.ConnectionState,
		}
		// Only reveal role to self for now
		if p.ID == playerID && p.Role != nil {
			info["role"] = p.Role.GetType()
		}
		playersInfo = append(playersInfo, info)
	}

	payload := map[string]interface{}{
		"id":       game.ID,
		"status":   game.Status,
		"phase":    game.Phase,
		"day":      game.Day,
		"hostId":   game.HostID,
		"settings": game.Settings,
		"players":  playersInfo,
	}

	h.notifier.NotifyPlayer(playerID, entities.NotifGameState, payload)
}

// sendError sends an error notification to a session
func (h *Handler) sendError(s *melody.Session, code, message string) {
	notif, _ := entities.NewNotification(entities.NotifError, map[string]interface{}{
		"code":    code,
		"message": message,
	})
	data, _ := json.Marshal(notif)
	s.Write(data)
}

// sendAck sends an acknowledgement to a session
func (h *Handler) sendAck(s *melody.Session, action string, success bool, message string) {
	notif, _ := entities.NewNotification(entities.NotifAck, map[string]interface{}{
		"action":  action,
		"success": success,
		"message": message,
	})
	data, _ := json.Marshal(notif)
	s.Write(data)
}

// --- Interface implementations for compatibility ---

// IsPlayerConnected implements ConnectionChecker interface
func (h *Handler) IsPlayerConnected(playerID entities.PlayerID) bool {
	return h.sessions.IsPlayerConnected(playerID)
}

// SendToPlayer implements PlayerSender interface
func (h *Handler) SendToPlayer(playerID entities.PlayerID, payload []byte) error {
	return h.sessions.SendToPlayer(playerID, payload)
}

// BroadcastToGame implements Broadcaster interface
func (h *Handler) BroadcastToGame(gameID entities.GameID, payload []byte) error {
	return h.sessions.BroadcastToGame(gameID, payload)
}

// DisconnectPlayer implements ports.PlayerDisconnecter interface
// Used when a player is kicked from a game to forcefully close their WebSocket connection
func (h *Handler) DisconnectPlayer(gameID entities.GameID, playerID entities.PlayerID, reason string) {
	if err := h.sessions.DisconnectPlayer(playerID, reason); err != nil {
		logger.Get().Warn().Err(err).
			Str("playerID", string(playerID)).
			Str("gameID", string(gameID)).
			Msg("Failed to disconnect player")
	}
}

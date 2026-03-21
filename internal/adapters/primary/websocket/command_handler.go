package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"shamus-backend/internal/domain/constants"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/commands"
	"shamus-backend/internal/domain/ports"
	"shamus-backend/pkg/logger"
	"time"
)

// CommandHandler handles client-initiated commands
type CommandHandler struct {
	gameService   ports.GameService
	playerService ports.PlayerService
	chatService   ports.ChatService
	notifier      ports.NotificationService
	gameEngine    ports.GameEngine
	disconnecter  ports.PlayerDisconnecter
}

// NewCommandHandler creates a new CommandHandler
func NewCommandHandler(
	gameService ports.GameService,
	playerService ports.PlayerService,
	chatService ports.ChatService,
	notifier ports.NotificationService,
) *CommandHandler {
	return &CommandHandler{
		gameService:   gameService,
		playerService: playerService,
		chatService:   chatService,
		notifier:      notifier,
	}
}

// SetDisconnecter sets the player disconnecter (for breaking circular dependency)
func (h *CommandHandler) SetDisconnecter(d ports.PlayerDisconnecter) {
	h.disconnecter = d
}

// SetGameEngine sets the game engine (for breaking circular dependency)
func (h *CommandHandler) SetGameEngine(ge ports.GameEngine) {
	h.gameEngine = ge
}

// SetPlayerService sets the player service (for breaking circular dependency)
func (h *CommandHandler) SetPlayerService(ps ports.PlayerService) {
	h.playerService = ps
}

// CommandContext contains information about the command context
type CommandContext struct {
	Ctx      context.Context // Go context for cancellation/timeout
	GameID   entities.GameID
	PlayerID entities.PlayerID
	Username string
}

// Handle processes a command from a client
func (h *CommandHandler) Handle(cmdCtx *CommandContext, cmd *entities.Command) error {
	switch cmd.Type {
	case entities.CmdSendChat:
		return h.handleSendChat(cmdCtx, cmd.Payload)
	case entities.CmdUpdateSettings:
		return h.handleUpdateSettings(cmdCtx, cmd.Payload)
	case entities.CmdStartGame:
		return h.handleStartGame(cmdCtx)
	case entities.CmdLeaveGame:
		return h.handleLeaveGame(cmdCtx)
	case entities.CmdKickPlayer:
		return h.handleKickPlayer(cmdCtx, cmd.Payload)
	default:
		return ErrUnknownCommand
	}
}

// handleSendChat processes a chat message command
func (h *CommandHandler) handleSendChat(cmdCtx *CommandContext, payload json.RawMessage) error {
	chatPayload, err := commands.ParseSendChatPayload(payload)
	if err != nil {
		logger.Get().Warn().Err(err).Msg("Failed to parse chat payload")
		return ErrInvalidPayload
	}

	// Validate message length
	if len(chatPayload.Message) < constants.MinChatMessageLength {
		return ErrChatMessageEmpty
	}
	if len(chatPayload.Message) > constants.MaxChatMessageLength {
		return ErrChatMessageTooLong
	}

	// Get game for phase info
	game, err := h.gameService.GetGame(cmdCtx.Ctx, cmdCtx.GameID)
	if err != nil {
		return err
	}

	// Get sender player info
	sender, err := h.playerService.GetPlayer(cmdCtx.Ctx, cmdCtx.PlayerID)
	if err != nil {
		return err
	}

	// Convert channel type
	var chatChannel string
	switch chatPayload.Channel {
	case commands.ChatChannelVillage:
		chatChannel = "village"
	case commands.ChatChannelWerewolf:
		chatChannel = "werewolf"
	case commands.ChatChannelDead:
		chatChannel = "dead"
	default:
		return ErrInvalidChatChannel
	}

	// Check if sender can send to this channel
	// For now, we'll do a simplified check - the ChatService can be used for more complex rules
	if !h.canSendToChannel(sender, chatChannel, game) {
		return ErrCannotSendToChannel
	}

	// Get recipients
	recipients, err := h.getChannelRecipients(cmdCtx, chatChannel, game)
	if err != nil {
		return err
	}

	// Send chat notification to recipients
	timestamp := time.Now().UnixMilli()
	h.notifier.NotifyChatMessage(recipients, cmdCtx.PlayerID, cmdCtx.Username, chatPayload.Message, chatChannel, timestamp)

	logger.Get().Info().
		Str("gameID", string(cmdCtx.GameID)).
		Str("playerID", string(cmdCtx.PlayerID)).
		Str("channel", chatChannel).
		Msg("Chat message sent")

	return nil
}

// handleUpdateSettings processes a settings update command
func (h *CommandHandler) handleUpdateSettings(cmdCtx *CommandContext, payload json.RawMessage) error {
	settingsPayload, err := commands.ParseUpdateSettingsPayload(payload)
	if err != nil {
		logger.Get().Warn().Err(err).Msg("Failed to parse settings payload")
		return ErrInvalidPayload
	}

	// Build GameSettings from payload
	newSettings := entities.GameSettings{
		Roles: settingsPayload.Roles,
	}

	// Call service to update settings (validates host + game state + roles)
	_, err = h.gameService.UpdateSettings(cmdCtx.Ctx, cmdCtx.GameID, cmdCtx.PlayerID, newSettings)
	if err != nil {
		return err
	}

	// Notify all players of new settings
	// The notification could include the new settings, but for simplicity we just send game state
	game, _ := h.gameService.GetGame(cmdCtx.Ctx, cmdCtx.GameID)
	if game != nil {
		// Send a simplified settings notification
		h.notifier.NotifyAll(cmdCtx.GameID, entities.NotificationType("settings_changed"), map[string]interface{}{
			"roles": game.Settings.Roles,
		})
	}

	logger.Get().Info().
		Str("gameID", string(cmdCtx.GameID)).
		Str("playerID", string(cmdCtx.PlayerID)).
		Msg("Game settings updated")

	return nil
}

// handleStartGame processes a start game command
func (h *CommandHandler) handleStartGame(cmdCtx *CommandContext) error {
	// Start the game
	game, players, err := h.gameService.StartGame(cmdCtx.Ctx, cmdCtx.GameID, cmdCtx.PlayerID)
	if err != nil {
		return err
	}

	// Send role reveal to each player
	for _, player := range players {
		if player.Role == nil {
			continue
		}
		roleName := string(player.Role.GetType())
		description := player.Role.GetDescription()
		h.notifier.NotifyRoleReveal(player.ID, player.Role.GetType(), roleName, description)
	}

	// Notify game started
	h.notifier.NotifyGameStarted(cmdCtx.GameID, game.Day)

	// Start the game flow (triggers first night phase)
	if h.gameEngine != nil {
		if err := h.gameEngine.StartGameFlow(cmdCtx.Ctx, cmdCtx.GameID); err != nil {
			logger.Get().Error().
				Str("gameID", string(cmdCtx.GameID)).
				Err(err).
				Msg("Error starting game flow")
		}
	}

	logger.Get().Info().
		Str("gameID", string(cmdCtx.GameID)).
		Str("hostID", string(cmdCtx.PlayerID)).
		Msg("Game started")

	return nil
}

// handleLeaveGame processes a leave game command
func (h *CommandHandler) handleLeaveGame(cmdCtx *CommandContext) error {
	// This is handled by the WebSocket disconnect, but we can trigger it manually
	// For now, we just acknowledge the intent - actual leave happens on disconnect
	h.notifier.NotifyAck(cmdCtx.PlayerID, "leave_game", true, "Disconnecting...")

	logger.Get().Info().
		Str("gameID", string(cmdCtx.GameID)).
		Str("playerID", string(cmdCtx.PlayerID)).
		Msg("Player requested to leave")

	return nil
}

// handleKickPlayer processes a kick player command
func (h *CommandHandler) handleKickPlayer(cmdCtx *CommandContext, payload json.RawMessage) error {
	kickPayload, err := commands.ParseKickPlayerPayload(payload)
	if err != nil {
		logger.Get().Warn().Err(err).Msg("Failed to parse kick payload")
		return ErrInvalidPayload
	}

	// Get game to verify host
	game, err := h.gameService.GetGame(cmdCtx.Ctx, cmdCtx.GameID)
	if err != nil {
		return err
	}

	// Only host can kick
	if game.HostID != cmdCtx.PlayerID {
		return ErrNotHost
	}

	// Can't kick yourself
	if kickPayload.PlayerID == cmdCtx.PlayerID {
		return ErrCannotKickSelf
	}

	// Can only kick in waiting state
	if game.Status != entities.GameStatusWaiting {
		return ErrGameNotWaiting
	}

	// Get player to kick for username
	playerToKick, err := h.playerService.GetPlayer(cmdCtx.Ctx, kickPayload.PlayerID)
	if err != nil {
		return err
	}

	// Notify the kicked player before disconnecting
	h.notifier.NotifyError(kickPayload.PlayerID, "KICKED", "You have been kicked from the game", "")

	// Notify others
	reason := kickPayload.Reason
	if reason == "" {
		reason = "kicked"
	}
	h.notifier.NotifyPlayerLeft(cmdCtx.GameID, kickPayload.PlayerID, playerToKick.Username, reason)

	// Remove the player from the game
	if err := h.playerService.LeaveGame(cmdCtx.Ctx, kickPayload.PlayerID); err != nil {
		logger.Get().Error().Err(err).
			Str("playerID", string(kickPayload.PlayerID)).
			Msg("Failed to remove kicked player from game")
		// Continue anyway - we still want to disconnect the player
	}

	// Force disconnect the player's WebSocket session
	if h.disconnecter != nil {
		h.disconnecter.DisconnectPlayer(cmdCtx.GameID, kickPayload.PlayerID, "kicked")
	}

	logger.Get().Info().
		Str("gameID", string(cmdCtx.GameID)).
		Str("hostID", string(cmdCtx.PlayerID)).
		Str("kickedID", string(kickPayload.PlayerID)).
		Msg("Player kicked and disconnected")

	return nil
}

// --- Helper methods ---

func (h *CommandHandler) canSendToChannel(sender *entities.Player, channel string, game *entities.Game) bool {
	// Dead players can only send to dead channel
	if !sender.IsAlive {
		return channel == "dead"
	}

	// During night, only werewolves can use werewolf channel
	if channel == "werewolf" {
		if game.Phase != entities.PhaseNight {
			return false
		}
		if sender.Role == nil || sender.Role.GetType() != entities.RoleWerewolf {
			return false
		}
		return true
	}

	// Village channel during day/vote
	if channel == "village" {
		return game.Phase == entities.PhaseDay || game.Phase == entities.PhaseVote || game.Status == entities.GameStatusWaiting
	}

	return false
}

func (h *CommandHandler) getChannelRecipients(cmdCtx *CommandContext, channel string, game *entities.Game) ([]entities.PlayerID, error) {
	players, err := h.playerService.GetGamePlayers(cmdCtx.Ctx, cmdCtx.GameID)
	if err != nil {
		return nil, err
	}

	var recipients []entities.PlayerID
	for _, p := range players {
		switch channel {
		case "village":
			// All players can receive village messages
			recipients = append(recipients, p.ID)
		case "werewolf":
			// Only werewolves receive werewolf messages
			if p.Role != nil && p.Role.GetType() == entities.RoleWerewolf {
				recipients = append(recipients, p.ID)
			}
		case "dead":
			// Only dead players receive dead messages
			if !p.IsAlive {
				recipients = append(recipients, p.ID)
			}
		}
	}

	return recipients, nil
}

// --- Errors ---

var (
	ErrUnknownCommand      = errors.New("unknown command type")
	ErrInvalidPayload      = errors.New("invalid payload format")
	ErrChatMessageEmpty    = errors.New("chat message cannot be empty")
	ErrChatMessageTooLong  = errors.New("chat message is too long")
	ErrInvalidChatChannel  = errors.New("invalid chat channel")
	ErrCannotSendToChannel = errors.New("cannot send to this channel")
	ErrNotHost             = errors.New("only the host can perform this action")
	ErrCannotKickSelf      = errors.New("cannot kick yourself")
	ErrGameNotWaiting      = errors.New("game is not in waiting state")
)

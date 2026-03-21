package ports

import (
	"context"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/events"
	"time"
)

// GameService defines the game business logic operations
type GameService interface {
	// CreateNewGame creates a new game with the given host
	CreateNewGame(ctx context.Context, hostID entities.PlayerID) (*entities.Game, error)

	// JoinGame adds a player to an existing game
	JoinGame(ctx context.Context, gameID entities.GameID, playerID entities.PlayerID) (*entities.Game, error)

	// GetGame retrieves a game by ID
	GetGame(ctx context.Context, gameID entities.GameID) (*entities.Game, error)

	// UpdateSettings updates game settings (roles configuration)
	// Only the host can update settings, and only when game is in waiting state
	UpdateSettings(ctx context.Context, gameID entities.GameID, playerID entities.PlayerID, settings entities.GameSettings) (*entities.Game, error)

	// StartGame starts a game - assigns roles and transitions to night phase
	// Only the host can start the game, and only when game is in waiting state
	StartGame(ctx context.Context, gameID entities.GameID, playerID entities.PlayerID) (*entities.Game, []*entities.Player, error)
}

// PlayerService defines player management operations
type PlayerService interface {
	// HandleConnect is called when a player connects via WebSocket
	// Returns the Player and a boolean indicating if this is a reconnection
	HandleConnect(ctx context.Context, gameID entities.GameID, playerID entities.PlayerID, username string) (*entities.Player, bool, error)

	// HandleDisconnect is called when a player disconnects from WebSocket
	HandleDisconnect(ctx context.Context, gameID entities.GameID, playerID entities.PlayerID) error

	// GetPlayer retrieves a player by ID
	GetPlayer(ctx context.Context, id entities.PlayerID) (*entities.Player, error)

	// GetGamePlayers retrieves all players in a game
	GetGamePlayers(ctx context.Context, gameID entities.GameID) ([]*entities.Player, error)

	// CleanupGamePlayers removes all players when a game ends
	CleanupGamePlayers(ctx context.Context, gameID entities.GameID) error

	// IsPlayerConnected checks if a player has an active WebSocket session
	// Note: This is an in-memory check and does not need context
	IsPlayerConnected(playerID entities.PlayerID) bool
}

// EventService defines operations for sending events to players
type EventService interface {
	// SendToPlayer sends an event to a specific player
	SendToPlayer(playerID entities.PlayerID, event entities.RawEvent) error

	// BroadcastToGame sends an event to all players in a game
	BroadcastToGame(gameID entities.GameID, event entities.RawEvent) error
}

// VisibilityService determines what information each player can see about other players
type VisibilityService interface {
	// BuildPlayersDetailsForPlayer returns the player list with appropriate role visibility
	// based on the viewer's role and the current game phase.
	// Rules:
	// - A player always sees their own role
	// - Werewolves see other werewolves' roles
	// - Dead players' roles are revealed when day starts (PhaseDay or PhaseVote)
	BuildPlayersDetailsForPlayer(
		viewer *entities.Player,
		allPlayers []*entities.Player,
		gamePhase entities.GamePhase,
	) []events.PlayersDetailsData
}

// ChatService handles chat message permissions and routing
type ChatService interface {
	// CanSendToChannel checks if a player can send messages to a specific channel
	// based on game phase, player role, and alive status
	CanSendToChannel(
		sender *entities.Player,
		channel events.ChatChannel,
		gamePhase entities.GamePhase,
	) bool

	// GetChannelRecipients returns the list of players who can receive messages
	// on a specific channel based on game phase and their roles
	GetChannelRecipients(
		channel events.ChatChannel,
		allPlayers []*entities.Player,
		gamePhase entities.GamePhase,
	) []*entities.Player
}

// TimerService manages phase and role timers for games
type TimerService interface {
	// StartPhaseTimer starts a timer for a game phase (day or vote)
	StartPhaseTimer(gameID entities.GameID, phase entities.GamePhase)

	// StartRoleTimer starts a timer for a specific role during night phase
	StartRoleTimer(gameID entities.GameID, roleType entities.RoleType)

	// CancelTimer cancels the current timer for a game
	CancelTimer(gameID entities.GameID)

	// SkipTimer skips the current timer (when all players have acted)
	SkipTimer(gameID entities.GameID)

	// GetRemainingTime returns the remaining time for a game's timer
	GetRemainingTime(gameID entities.GameID) time.Duration
}

// VoteService manages voting sessions for games
type VoteService interface {
	// StartVillageVote starts a village vote to eliminate a player
	StartVillageVote(ctx context.Context, gameID entities.GameID, alivePlayers []*entities.Player) (*entities.Vote, error)

	// StartWerewolfVote starts a werewolf vote to choose a victim
	StartWerewolfVote(ctx context.Context, gameID entities.GameID, werewolves []*entities.Player, potentialVictims []*entities.Player) (*entities.Vote, error)

	// CastVote records a player's vote
	CastVote(ctx context.Context, gameID entities.GameID, voterID entities.PlayerID, targetID *entities.PlayerID) error

	// HasEveryoneVoted checks if all eligible voters have voted
	HasEveryoneVoted(ctx context.Context, gameID entities.GameID) (bool, error)

	// ResolveVote resolves the current vote and returns the result
	ResolveVote(ctx context.Context, gameID entities.GameID) (*entities.VoteResult, error)

	// GetVote returns the current vote for a game
	GetVote(ctx context.Context, gameID entities.GameID) (*entities.Vote, bool, error)

	// ClearVote removes the vote for a game
	ClearVote(ctx context.Context, gameID entities.GameID) error
}

// NightService manages the night phase actions
// Note: This service uses in-memory state only, no context needed.
// The implementation uses the NightPhase type internally.
type NightService interface {
	// StartNight initializes a new night phase
	StartNight(gameID entities.GameID, players []*entities.Player)

	// GetCurrentPhase returns the current night phase as string
	GetCurrentPhase(gameID entities.GameID) string

	// RecordSeerAction records the seer's vision
	RecordSeerAction(gameID entities.GameID, targetID entities.PlayerID, revealedRole entities.RoleType)

	// RecordWerewolfVictim records the werewolf vote result
	RecordWerewolfVictim(gameID entities.GameID, victimID *entities.PlayerID)

	// RecordWitchAction records the witch's actions
	// Returns error if healTargetID is provided but doesn't match the werewolf victim
	RecordWitchAction(gameID entities.GameID, healTargetID, poisonTargetID *entities.PlayerID) error

	// GetPendingDeaths returns the list of players who will die at dawn
	GetPendingDeaths(gameID entities.GameID) []entities.PlayerID

	// IsNightComplete checks if all night actions are done
	IsNightComplete(gameID entities.GameID) bool

	// ClearNight removes the night state for a game
	ClearNight(gameID entities.GameID)
}

// GameEngine orchestrates game flow and phase transitions
type GameEngine interface {
	// StartGameFlow starts the game flow after StartGame (begins first night)
	StartGameFlow(ctx context.Context, gameID entities.GameID) error

	// HandleSeerAction processes the seer's night action
	HandleSeerAction(ctx context.Context, gameID entities.GameID, seerID, targetID entities.PlayerID) error

	// HandleWerewolfVote processes a werewolf's vote during night
	HandleWerewolfVote(ctx context.Context, gameID entities.GameID, werewolfID entities.PlayerID, targetID *entities.PlayerID) error

	// HandleWitchAction processes the witch's night action
	HandleWitchAction(ctx context.Context, gameID entities.GameID, witchID entities.PlayerID, healTargetID, poisonTargetID *entities.PlayerID) error

	// HandleVillageVote processes a village vote during day
	HandleVillageVote(ctx context.Context, gameID entities.GameID, voterID entities.PlayerID, targetID *entities.PlayerID) error
}

// NotificationService handles sending notifications to players
// Notifications are one-way messages from server to client
// Note: Most methods don't need context as they only send to connected players.
// NotifyRole needs context as it queries the player repository.
type NotificationService interface {
	// NotifyPlayer sends a notification to a specific player
	NotifyPlayer(playerID entities.PlayerID, notifType entities.NotificationType, payload interface{}) error

	// NotifyPlayers sends a notification to multiple specific players
	NotifyPlayers(playerIDs []entities.PlayerID, notifType entities.NotificationType, payload interface{})

	// NotifyAll sends a notification to all players in a game
	NotifyAll(gameID entities.GameID, notifType entities.NotificationType, payload interface{}) error

	// NotifyRole sends a notification to all alive players with a specific role
	NotifyRole(ctx context.Context, gameID entities.GameID, roleType entities.RoleType, notifType entities.NotificationType, payload interface{}) error

	// --- Typed helper methods ---

	// NotifyPhaseChanged notifies all players of a phase change
	NotifyPhaseChanged(gameID entities.GameID, phase entities.GamePhase, day int, subPhase string) error

	// NotifyPlayerJoined notifies all players that a new player joined
	NotifyPlayerJoined(gameID entities.GameID, playerID entities.PlayerID, username string) error

	// NotifyPlayerLeft notifies all players that a player left
	NotifyPlayerLeft(gameID entities.GameID, playerID entities.PlayerID, username string, reason string) error

	// NotifyPlayerDied notifies all players of a death
	NotifyPlayerDied(gameID entities.GameID, playerID entities.PlayerID, username string, role entities.RoleType, cause string) error

	// NotifyPlayerInactive notifies all players that a player became inactive
	NotifyPlayerInactive(gameID entities.GameID, playerID entities.PlayerID, username string) error

	// NotifyGameStarted notifies all players that the game started
	NotifyGameStarted(gameID entities.GameID, day int) error

	// NotifyGameEnded notifies all players that the game ended
	NotifyGameEnded(gameID entities.GameID, winningClan entities.Clan, winners []entities.PlayerID) error

	// NotifyRoleReveal sends a player their assigned role (private)
	NotifyRoleReveal(playerID entities.PlayerID, role entities.RoleType, roleName, description string) error

	// NotifyTimerStarted notifies all players that a timer has started
	NotifyTimerStarted(gameID entities.GameID, phase entities.GamePhase, subPhase string, durationSec int) error

	// NotifySeerResult sends the seer their vision result (private)
	NotifySeerResult(seerID entities.PlayerID, targetID entities.PlayerID, targetUsername string, role entities.RoleType) error

	// NotifyVoteStarted notifies relevant players that a vote has started
	NotifyVoteStarted(gameID entities.GameID, voteType string, voters, targets []entities.PlayerID, durationSec int) error

	// NotifyVoteUpdate sends real-time vote update to group members
	NotifyVoteUpdate(playerIDs []entities.PlayerID, votes map[entities.PlayerID]*entities.PlayerID, voteCounts map[entities.PlayerID]int, votersRemaining []entities.PlayerID)

	// NotifyVoteResult notifies all players of the vote result
	NotifyVoteResult(gameID entities.GameID, voteType string, result *entities.PlayerID, isTie bool, tiedPlayers []entities.PlayerID, finalVotes map[entities.PlayerID]int, mayorDecided bool) error

	// NotifyChatMessage broadcasts a chat message to appropriate recipients
	NotifyChatMessage(recipients []entities.PlayerID, senderID entities.PlayerID, senderName, message, channel string, timestamp int64)

	// NotifyError sends an error notification to a player
	NotifyError(playerID entities.PlayerID, code, message, action string) error

	// NotifyAck sends an acknowledgement to a player
	NotifyAck(playerID entities.PlayerID, action string, success bool, message string) error
}

// PromptCallback is the function type for prompt response callbacks
type PromptCallback func(ctx context.Context, prompt *entities.Prompt, response []byte) error

// PromptService manages interactive prompts with timeouts
// Prompts are two-way: server sends prompt, client responds
type PromptService interface {
	// RegisterCallback registers a callback for a specific context
	RegisterCallback(context string, callback PromptCallback)

	// CreatePrompt creates and sends a prompt to a single player
	CreatePrompt(
		gameID entities.GameID,
		playerID entities.PlayerID,
		promptType entities.PromptType,
		context string,
		payload interface{},
		timeout time.Duration,
		canSkip bool,
	) (*entities.Prompt, error)

	// RespondToPrompt handles a player's response to a prompt
	RespondToPrompt(ctx context.Context, promptID entities.PromptID, playerID entities.PlayerID, response []byte) error

	// CancelPrompt cancels a pending prompt
	CancelPrompt(promptID entities.PromptID) error

	// CreateGroupVote creates a group vote (werewolf or village vote)
	CreateGroupVote(
		gameID entities.GameID,
		voteContext string,
		voters []*entities.Player,
		eligibleTargets []entities.PlayerID,
		timeout time.Duration,
		canAbstain bool,
		mayorID *entities.PlayerID,
	) (*entities.GroupID, error)

	// ResolveGroupVote manually resolves a group vote
	ResolveGroupVote(groupID entities.GroupID) (interface{}, error)

	// HasEveryoneVoted checks if all voters in a group have voted
	HasEveryoneVoted(groupID entities.GroupID) bool

	// CleanupGame removes all prompts and group states for a game
	CleanupGame(gameID entities.GameID)
}

// ConnectionChecker checks if a player has an active connection
type ConnectionChecker interface {
	// IsPlayerConnected returns true if the player has an active WebSocket session
	IsPlayerConnected(playerID entities.PlayerID) bool
}

package ports

import (
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/events"
	"time"
)

// GameService defines the game business logic operations
type GameService interface {
	// CreateNewGame creates a new game with the given host
	CreateNewGame(hostID entities.PlayerID) (*entities.Game, error)

	// JoinGame adds a player to an existing game
	JoinGame(gameID entities.GameID, playerID entities.PlayerID) (*entities.Game, error)

	// GetGame retrieves a game by ID
	GetGame(gameID entities.GameID) (*entities.Game, error)

	// UpdateSettings updates game settings (roles configuration)
	// Only the host can update settings, and only when game is in waiting state
	UpdateSettings(gameID entities.GameID, playerID entities.PlayerID, settings entities.GameSettings) (*entities.Game, error)

	// StartGame starts a game - assigns roles and transitions to night phase
	// Only the host can start the game, and only when game is in waiting state
	StartGame(gameID entities.GameID, playerID entities.PlayerID) (*entities.Game, []*entities.Player, error)
}

// PlayerService defines player management operations
type PlayerService interface {
	// HandleConnect is called when a player connects via WebSocket
	// Returns the Player and a boolean indicating if this is a reconnection
	HandleConnect(gameID entities.GameID, playerID entities.PlayerID, username string) (*entities.Player, bool, error)

	// HandleDisconnect is called when a player disconnects from WebSocket
	HandleDisconnect(gameID entities.GameID, playerID entities.PlayerID) error

	// GetPlayer retrieves a player by ID
	GetPlayer(id entities.PlayerID) (*entities.Player, error)

	// GetGamePlayers retrieves all players in a game
	GetGamePlayers(gameID entities.GameID) ([]*entities.Player, error)

	// CleanupGamePlayers removes all players when a game ends
	CleanupGamePlayers(gameID entities.GameID) error

	// IsPlayerConnected checks if a player has an active WebSocket session
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
	StartVillageVote(gameID entities.GameID, alivePlayers []*entities.Player) (*entities.Vote, error)

	// StartWerewolfVote starts a werewolf vote to choose a victim
	StartWerewolfVote(gameID entities.GameID, werewolves []*entities.Player, potentialVictims []*entities.Player) (*entities.Vote, error)

	// CastVote records a player's vote
	CastVote(gameID entities.GameID, voterID entities.PlayerID, targetID *entities.PlayerID) error

	// HasEveryoneVoted checks if all eligible voters have voted
	HasEveryoneVoted(gameID entities.GameID) bool

	// ResolveVote resolves the current vote and returns the result
	ResolveVote(gameID entities.GameID) (*entities.VoteResult, error)

	// GetVote returns the current vote for a game
	GetVote(gameID entities.GameID) (*entities.Vote, bool)

	// ClearVote removes the vote for a game
	ClearVote(gameID entities.GameID)
}

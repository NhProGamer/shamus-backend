package notifications

import "shamus-backend/internal/domain/entities"

// PhaseChangedPayload is sent when the game phase changes
type PhaseChangedPayload struct {
	Phase    entities.GamePhase `json:"phase"`
	Day      int                `json:"day"`
	SubPhase string             `json:"subPhase,omitempty"` // e.g., "seer", "werewolf", "witch" during night
}

// PlayerJoinedPayload is sent when a player joins the game
type PlayerJoinedPayload struct {
	PlayerID entities.PlayerID `json:"playerId"`
	Username string            `json:"username"`
}

// PlayerLeftPayload is sent when a player leaves the game
type PlayerLeftPayload struct {
	PlayerID entities.PlayerID `json:"playerId"`
	Username string            `json:"username"`
	Reason   string            `json:"reason,omitempty"` // "disconnected", "kicked", "left"
}

// PlayerDiedPayload is sent when a player dies
type PlayerDiedPayload struct {
	PlayerID entities.PlayerID  `json:"playerId"`
	Username string             `json:"username"`
	Role     entities.RoleType  `json:"role"`
	Cause    string             `json:"cause"` // "werewolf", "village_vote", "witch_poison"
}

// PlayerInactivePayload is sent when a player becomes inactive (disconnected too long)
type PlayerInactivePayload struct {
	PlayerID entities.PlayerID `json:"playerId"`
	Username string            `json:"username"`
}

// GameStatePayload is sent to provide full game state (on connect, on request)
type GameStatePayload struct {
	ID       entities.GameID     `json:"id"`
	Status   entities.GameStatus `json:"status"`
	Phase    entities.GamePhase  `json:"phase"`
	Day      int                 `json:"day"`
	HostID   entities.PlayerID   `json:"hostId"`
	Settings entities.GameSettings `json:"settings"`
	Players  []PlayerStateInfo   `json:"players"`
}

// PlayerStateInfo contains player state visible to others
type PlayerStateInfo struct {
	ID              entities.PlayerID       `json:"id"`
	Username        string                  `json:"username"`
	IsAlive         bool                    `json:"isAlive"`
	ConnectionState entities.ConnectionState `json:"connectionState"`
	// Role is only visible according to visibility rules
	Role *entities.RoleType `json:"role,omitempty"`
}

// GameStartedPayload is sent when the game starts
type GameStartedPayload struct {
	Day int `json:"day"`
}

// GameEndedPayload is sent when the game ends
type GameEndedPayload struct {
	WinningClan entities.Clan       `json:"winningClan"`
	Winners     []entities.PlayerID `json:"winners"`
}

// RoleRevealPayload is sent privately to a player to reveal their role
type RoleRevealPayload struct {
	Role        entities.RoleType `json:"role"`
	RoleName    string            `json:"roleName"`
	Description string            `json:"description"`
}

// TimerStartedPayload is sent when a timer starts
type TimerStartedPayload struct {
	Phase       entities.GamePhase `json:"phase"`
	SubPhase    string             `json:"subPhase,omitempty"`
	DurationSec int                `json:"durationSec"`
}

// TimerTickPayload is sent periodically during a timer
type TimerTickPayload struct {
	Phase        entities.GamePhase `json:"phase"`
	SubPhase     string             `json:"subPhase,omitempty"`
	RemainingSec int                `json:"remainingSec"`
}

// TimerExpiredPayload is sent when a timer expires
type TimerExpiredPayload struct {
	Phase    entities.GamePhase `json:"phase"`
	SubPhase string             `json:"subPhase,omitempty"`
}

// SeerResultPayload is sent privately to the seer with vision result
type SeerResultPayload struct {
	TargetID       entities.PlayerID `json:"targetId"`
	TargetUsername string            `json:"targetUsername"`
	Role           entities.RoleType `json:"role"`
}

// VoteStartedPayload is sent when a vote begins
type VoteStartedPayload struct {
	VoteType        string              `json:"voteType"` // "village", "werewolf"
	EligibleVoters  []entities.PlayerID `json:"eligibleVoters"`
	EligibleTargets []entities.PlayerID `json:"eligibleTargets"`
	DurationSec     int                 `json:"durationSec"`
}

// VoteUpdatePayload is sent for real-time vote updates (group votes)
type VoteUpdatePayload struct {
	// Votes maps voterID to targetID (nil means abstained)
	Votes map[entities.PlayerID]*entities.PlayerID `json:"votes"`

	// VoteCounts maps targetID to number of votes
	VoteCounts map[entities.PlayerID]int `json:"voteCounts"`

	// VotersRemaining is the list of players who haven't voted yet
	VotersRemaining []entities.PlayerID `json:"votersRemaining"`
}

// VoteResultPayload is sent when a vote concludes
type VoteResultPayload struct {
	VoteType string `json:"voteType"` // "village", "werewolf"

	// Result is the player eliminated (nil if no elimination)
	Result *entities.PlayerID `json:"result,omitempty"`

	// IsTie indicates if there was a tie
	IsTie bool `json:"isTie"`

	// TiedPlayers lists players with equal votes (if tie)
	TiedPlayers []entities.PlayerID `json:"tiedPlayers,omitempty"`

	// FinalVotes shows the final vote tally
	FinalVotes map[entities.PlayerID]int `json:"finalVotes"`

	// MayorDecided indicates if mayor broke the tie
	MayorDecided bool `json:"mayorDecided,omitempty"`
}

// MayorTiebreakerPayload is sent to the mayor when they must break a tie
type MayorTiebreakerPayload struct {
	TiedPlayers []entities.PlayerID `json:"tiedPlayers"`
	VoteCounts  map[entities.PlayerID]int `json:"voteCounts"`
}

// ChatMessagePayload is sent when a chat message is received
type ChatMessagePayload struct {
	SenderID   entities.PlayerID `json:"senderId"`
	SenderName string            `json:"senderName"`
	Message    string            `json:"message"`
	Channel    string            `json:"channel"` // "village", "werewolf", "dead"
	Timestamp  int64             `json:"timestamp"`
}

// ErrorPayload is sent when an error occurs
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Action  string `json:"action,omitempty"` // The action that caused the error
}

// AckPayload is sent to acknowledge a command
type AckPayload struct {
	Action  string `json:"action"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

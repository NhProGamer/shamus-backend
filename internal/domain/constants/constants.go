package constants

import "time"

// Game configuration constants
const (
	// MinPlayers is the minimum number of players required to start a game
	MinPlayers = 4
	// MaxPlayers is the maximum number of players allowed in a game
	MaxPlayers = 24
)

// Redis TTL constants
const (
	// GameTTL is the time-to-live for game data in Redis
	GameTTL = 24 * time.Hour
	// PlayerTTL is the time-to-live for player data in Redis
	PlayerTTL = 24 * time.Hour
	// VoteTTL is the time-to-live for vote data in Redis
	VoteTTL = 24 * time.Hour
	// NightStateTTL is the time-to-live for night state data in Redis
	NightStateTTL = 24 * time.Hour
)

// Connection and timeout constants
const (
	// ReconnectionTimeout is the duration a player has to reconnect before being marked inactive
	ReconnectionTimeout = 2 * time.Minute
	// ServerReadTimeout is the HTTP server read timeout
	ServerReadTimeout = 15 * time.Second
	// ServerWriteTimeout is the HTTP server write timeout
	ServerWriteTimeout = 15 * time.Second
)

// Chat message validation constants
const (
	// MaxChatMessageLength is the maximum length of a chat message
	MaxChatMessageLength = 500
	// MinChatMessageLength is the minimum length of a chat message (must be > 0)
	MinChatMessageLength = 1
)

// Timer durations for game phases
const (
	// DayPhaseDuration is the duration of the day discussion phase
	DayPhaseDuration = 3 * time.Minute
	// VotePhaseDuration is the duration of the voting phase
	VotePhaseDuration = 2 * time.Minute
	// SeerActionDuration is the duration for the seer to act
	SeerActionDuration = 30 * time.Second
	// WerewolfVoteDuration is the duration for werewolves to vote
	WerewolfVoteDuration = 1 * time.Minute
	// WitchActionDuration is the duration for the witch to act
	WitchActionDuration = 45 * time.Second
)

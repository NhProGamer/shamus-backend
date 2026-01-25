package apperrors

import "fmt"

// AppError represents a structured application error
type AppError struct {
	Code    string // Machine-readable error code
	Message string // Human-readable message
	Err     error  // Wrapped error (optional)
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error for errors.Is/As support
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Wrap creates a new AppError wrapping an existing error
func Wrap(code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// Game errors
var (
	ErrGameNotFound      = New("GAME_NOT_FOUND", "game not found")
	ErrGameAlreadyExists = New("GAME_ALREADY_EXISTS", "game already exists")
	ErrGameEnded         = New("GAME_ENDED", "game has ended")
	ErrGameFull          = New("GAME_FULL", "game is full")
	ErrGameNotStarted    = New("GAME_NOT_STARTED", "game has not started")
	ErrGameNotWaiting    = New("GAME_NOT_WAITING", "cannot join a game that is not in waiting state")
)

// Player errors
var (
	ErrPlayerNotFound         = New("PLAYER_NOT_FOUND", "player not found")
	ErrPlayerAlreadyExists    = New("PLAYER_ALREADY_EXISTS", "player already exists")
	ErrPlayerNotInGame        = New("PLAYER_NOT_IN_GAME", "player is not in this game")
	ErrPlayerAlreadyInGame    = New("PLAYER_ALREADY_IN_GAME", "player is already in a game")
	ErrPlayerAlreadyConnected = New("PLAYER_ALREADY_CONNECTED", "player is already connected to this game")
	ErrPlayerInactive         = New("PLAYER_INACTIVE", "player is inactive and cannot rejoin")
)

// Auth errors
var (
	ErrUnauthorized  = New("UNAUTHORIZED", "unauthorized access")
	ErrInvalidToken  = New("INVALID_TOKEN", "invalid or expired token")
	ErrMissingUserID = New("MISSING_USER_ID", "user ID is missing")
)

// Validation errors
var (
	ErrInvalidInput = New("INVALID_INPUT", "invalid input provided")
	ErrMissingField = New("MISSING_FIELD", "required field is missing")
)

// Settings errors
var (
	ErrNotHost            = New("NOT_HOST", "only the host can modify game settings")
	ErrInvalidRole        = New("INVALID_ROLE", "invalid role type")
	ErrTooManyRoles       = New("TOO_MANY_ROLES", "too many roles configured")
	ErrRoleLimitExceeded  = New("ROLE_LIMIT_EXCEEDED", "role limit exceeded for this role type")
	ErrNotEnoughPlayers   = New("NOT_ENOUGH_PLAYERS", "not enough players to start")
	ErrTooManyPlayers     = New("TOO_MANY_PLAYERS", "too many players")
	ErrRoleCountMismatch  = New("ROLE_COUNT_MISMATCH", "number of roles must equal number of players")
	ErrInvalidComposition = New("INVALID_COMPOSITION", "need at least one villager and one werewolf/rogue")
)

// Game action errors - Phase/turn errors
var (
	ErrNotYourTurn   = New("NOT_YOUR_TURN", "not your turn")
	ErrWrongPhase    = New("WRONG_PHASE", "wrong game phase")
	ErrAlreadyActed  = New("ALREADY_ACTED", "you already acted this phase")
	ErrGameNotActive = New("GAME_NOT_ACTIVE", "game is not active")
)

// Game action errors - Player state errors
var (
	ErrPlayerDead = New("PLAYER_DEAD", "player is dead")
	ErrWrongRole  = New("WRONG_ROLE", "you don't have this role")
)

// Game action errors - Target errors
var (
	ErrInvalidTarget    = New("INVALID_TARGET", "invalid target")
	ErrTargetDead       = New("TARGET_DEAD", "target is dead")
	ErrCannotTargetSelf = New("CANNOT_TARGET_SELF", "cannot target yourself")
)

// Game action errors - Ability errors
var (
	ErrAbilityUsed        = New("ABILITY_USED", "ability already used")
	ErrCanOnlyHealVictim  = New("CAN_ONLY_HEAL_VICTIM", "can only heal the werewolf victim")
	ErrTargetAlreadyDying = New("TARGET_ALREADY_DYING", "target is already dying tonight")
)

// Vote errors
var (
	ErrVoteNotFound      = New("VOTE_NOT_FOUND", "vote not found")
	ErrVoteAlreadyExists = New("VOTE_ALREADY_EXISTS", "vote already exists for this game")
	ErrInvalidVoter      = New("INVALID_VOTER", "player is not eligible to vote")
	ErrVoteNotActive     = New("VOTE_NOT_ACTIVE", "vote is not active")
)

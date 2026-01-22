package events

import "shamus-backend/internal/domain/entities"

const EventTypeError entities.EventType = "error"

// ErrorCode represents machine-readable error codes for clients
type ErrorCode string

const (
	ErrorCodeWrongPhase        ErrorCode = "WRONG_PHASE"
	ErrorCodeNotYourTurn       ErrorCode = "NOT_YOUR_TURN"
	ErrorCodeAlreadyActed      ErrorCode = "ALREADY_ACTED"
	ErrorCodeGameNotActive     ErrorCode = "GAME_NOT_ACTIVE"
	ErrorCodePlayerDead        ErrorCode = "PLAYER_DEAD"
	ErrorCodeWrongRole         ErrorCode = "WRONG_ROLE"
	ErrorCodeInvalidTarget     ErrorCode = "INVALID_TARGET"
	ErrorCodeTargetDead        ErrorCode = "TARGET_DEAD"
	ErrorCodeCannotTargetSelf  ErrorCode = "CANNOT_TARGET_SELF"
	ErrorCodeAbilityUsed       ErrorCode = "ABILITY_USED"
	ErrorCodeCanOnlyHealVictim ErrorCode = "CAN_ONLY_HEAL_VICTIM"
	ErrorCodeVoteNotFound      ErrorCode = "VOTE_NOT_FOUND"
	ErrorCodeVoteNotActive     ErrorCode = "VOTE_NOT_ACTIVE"
	ErrorCodeInvalidVoter      ErrorCode = "INVALID_VOTER"
	ErrorCodeInvalidAction     ErrorCode = "INVALID_ACTION"
	ErrorCodeUnknown           ErrorCode = "UNKNOWN_ERROR"
)

// ErrorEventData contains structured error information for clients
type ErrorEventData struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Action  string    `json:"action,omitempty"` // The action that caused the error
}

// NewErrorEvent creates a new error event
func NewErrorEvent(code ErrorCode, message string, action string) entities.Event[ErrorEventData] {
	return entities.Event[ErrorEventData]{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeError,
		Data: ErrorEventData{
			Code:    code,
			Message: message,
			Action:  action,
		},
	}
}

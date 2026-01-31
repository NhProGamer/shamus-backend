package entities

import (
	"encoding/json"
	"time"
)

// ActionID is a unique identifier for an action
type ActionID string

// ActionType represents the type of action
type ActionType string

// ActionStatus represents the current status of an action
type ActionStatus string

const (
	// Action types for different game interactions
	ActionTypeSeerVision   ActionType = "seer_vision"
	ActionTypeWerewolfVote ActionType = "werewolf_vote"
	ActionTypeWitchPotion  ActionType = "witch_potion"
	ActionTypeVillageVote  ActionType = "village_vote"
)

const (
	// ActionStatusPending indicates the action is waiting for a response
	ActionStatusPending ActionStatus = "pending"
	// ActionStatusCompleted indicates the action has been successfully responded to
	ActionStatusCompleted ActionStatus = "completed"
	// ActionStatusExpired indicates the action timeout has been reached
	ActionStatusExpired ActionStatus = "expired"
	// ActionStatusCancelled indicates the action was manually cancelled
	ActionStatusCancelled ActionStatus = "cancelled"
)

// Action represents a user action that requires a response within a timeout
type Action struct {
	ID        ActionID        `json:"id"`
	Type      ActionType      `json:"type"`
	GameID    GameID          `json:"gameId"`
	PlayerID  PlayerID        `json:"playerId"`
	Status    ActionStatus    `json:"status"`
	Payload   json.RawMessage `json:"payload"`
	Response  json.RawMessage `json:"response,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
	ExpiresAt time.Time       `json:"expiresAt"`
	Timeout   time.Duration   `json:"-"` // Not serialized, used for timer management
}

// NewAction creates a new Action with pending status
func NewAction(id ActionID, actionType ActionType, gameID GameID, playerID PlayerID, payload json.RawMessage, timeout time.Duration) *Action {
	now := time.Now()
	return &Action{
		ID:        id,
		Type:      actionType,
		GameID:    gameID,
		PlayerID:  playerID,
		Status:    ActionStatusPending,
		Payload:   payload,
		Response:  nil,
		CreatedAt: now,
		ExpiresAt: now.Add(timeout),
		Timeout:   timeout,
	}
}

// IsExpired checks if the action has exceeded its timeout
func (a *Action) IsExpired() bool {
	return time.Now().After(a.ExpiresAt)
}

// CanRespond checks if the action is still eligible to receive a response
func (a *Action) CanRespond() bool {
	return a.Status == ActionStatusPending && !a.IsExpired()
}

// MarkCompleted marks the action as completed with the given response
func (a *Action) MarkCompleted(response json.RawMessage) {
	a.Status = ActionStatusCompleted
	a.Response = response
}

// MarkExpired marks the action as expired (timeout reached)
func (a *Action) MarkExpired() {
	a.Status = ActionStatusExpired
}

// MarkCancelled marks the action as cancelled
func (a *Action) MarkCancelled() {
	a.Status = ActionStatusCancelled
}

// GetRemainingTime returns the time remaining before the action expires
func (a *Action) GetRemainingTime() time.Duration {
	remaining := time.Until(a.ExpiresAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

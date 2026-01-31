package events

import (
	"shamus-backend/internal/domain/entities"
	"time"
)

const (
	// EventTypeActionCreated is sent from server to client when a new action is available
	EventTypeActionCreated entities.EventType = "action_created"
	// EventTypeActionResponse is sent from client to server to respond to an action
	EventTypeActionResponse entities.EventType = "action_response"
	// EventTypeActionExpired is sent from server to client when an action times out
	EventTypeActionExpired entities.EventType = "action_expired"
)

// ActionCreatedEventData is sent to a player when they receive a new action
type ActionCreatedEventData struct {
	ActionID  string                `json:"actionId"`
	Type      entities.ActionType   `json:"type"`
	Payload   interface{}           `json:"payload"`   // Specific payload (WitchPotionPayload, etc.)
	ExpiresAt time.Time             `json:"expiresAt"`
	Timeout   int                   `json:"timeout"`   // Timeout in seconds for client display
}

// ActionResponseEventData is sent by a player to respond to an action
type ActionResponseEventData struct {
	ActionID string      `json:"actionId"`
	Response interface{} `json:"response"` // Specific response (WitchPotionResponse, etc.)
}

// ActionExpiredEventData is sent to a player when their action times out
type ActionExpiredEventData struct {
	ActionID string              `json:"actionId"`
	Type     entities.ActionType `json:"type"`
}

// NewActionCreatedEvent creates a new action created event
func NewActionCreatedEvent(actionID string, actionType entities.ActionType, payload interface{}, expiresAt time.Time, timeout time.Duration) entities.Event[ActionCreatedEventData] {
	return entities.Event[ActionCreatedEventData]{
		Channel: entities.EventChannelAction,
		Type:    EventTypeActionCreated,
		Data: ActionCreatedEventData{
			ActionID:  actionID,
			Type:      actionType,
			Payload:   payload,
			ExpiresAt: expiresAt,
			Timeout:   int(timeout.Seconds()),
		},
	}
}

// NewActionExpiredEvent creates a new action expired event
func NewActionExpiredEvent(actionID string, actionType entities.ActionType) entities.Event[ActionExpiredEventData] {
	return entities.Event[ActionExpiredEventData]{
		Channel: entities.EventChannelAction,
		Type:    EventTypeActionExpired,
		Data: ActionExpiredEventData{
			ActionID: actionID,
			Type:     actionType,
		},
	}
}

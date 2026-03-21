package entities

import "encoding/json"

// NotificationType represents the type of notification sent to clients
type NotificationType string

const (
	// Phase notifications
	NotifPhaseChanged NotificationType = "phase_changed"

	// Player notifications
	NotifPlayerJoined   NotificationType = "player_joined"
	NotifPlayerLeft     NotificationType = "player_left"
	NotifPlayerDied     NotificationType = "player_died"
	NotifPlayerInactive NotificationType = "player_inactive"
	NotifHostChanged    NotificationType = "host_changed"

	// Game state notifications
	NotifGameState   NotificationType = "game_state"
	NotifGameStarted NotificationType = "game_started"
	NotifGameEnded   NotificationType = "game_ended"
	NotifRoleReveal  NotificationType = "role_reveal" // Private: player's own role

	// Timer notifications
	NotifTimerStarted NotificationType = "timer_started"
	NotifTimerTick    NotificationType = "timer_tick"
	NotifTimerExpired NotificationType = "timer_expired"

	// Action result notifications (private)
	NotifSeerResult NotificationType = "seer_result"

	// Vote notifications
	NotifVoteStarted     NotificationType = "vote_started"
	NotifVoteUpdate      NotificationType = "vote_update"       // Real-time vote update for group votes
	NotifVoteResult      NotificationType = "vote_result"       // Final result
	NotifMayorTiebreaker NotificationType = "mayor_tiebreaker"  // Mayor must break tie

	// Chat notifications
	NotifChatMessage NotificationType = "chat_message"

	// Error notifications
	NotifError NotificationType = "error"

	// Acknowledgement
	NotifAck NotificationType = "ack"
)

// Notification represents a server-to-client notification message
type Notification struct {
	Channel Channel          `json:"channel"`
	Type    NotificationType `json:"type"`
	Payload json.RawMessage  `json:"payload"`
}

// NewNotification creates a new notification with the given type and payload
func NewNotification(notifType NotificationType, payload interface{}) (*Notification, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &Notification{
		Channel: ChannelNotification,
		Type:    notifType,
		Payload: payloadBytes,
	}, nil
}

// MustNewNotification creates a notification and panics on error (use for static payloads)
func MustNewNotification(notifType NotificationType, payload interface{}) *Notification {
	n, err := NewNotification(notifType, payload)
	if err != nil {
		panic(err)
	}
	return n
}

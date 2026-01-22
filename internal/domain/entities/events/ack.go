package events

import "shamus-backend/internal/domain/entities"

const EventTypeAck entities.EventType = "ack"

// AckEventData confirms that an action was successfully processed
type AckEventData struct {
	Action  string `json:"action"`            // The action that was acknowledged
	Success bool   `json:"success"`           // Whether the action succeeded
	Message string `json:"message,omitempty"` // Optional message
}

// NewAckEvent creates a new acknowledgement event
func NewAckEvent(action string, success bool, message string) entities.Event[AckEventData] {
	return entities.Event[AckEventData]{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeAck,
		Data: AckEventData{
			Action:  action,
			Success: success,
			Message: message,
		},
	}
}

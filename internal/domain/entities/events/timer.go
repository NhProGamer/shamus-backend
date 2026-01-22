package events

import "shamus-backend/internal/domain/entities"

const EventTypeTimer entities.EventType = "timer"

// TimerStatus represents the status of a timer
type TimerStatus string

const (
	TimerStatusStarted TimerStatus = "started"
	TimerStatusTick    TimerStatus = "tick"
	TimerStatusExpired TimerStatus = "expired"
	TimerStatusSkipped TimerStatus = "skipped"
)

// TimerEventData contains timer information sent to clients
type TimerEventData struct {
	Phase     entities.GamePhase `json:"phase"`
	RoleType  *entities.RoleType `json:"roleType,omitempty"` // For night phase, which role's turn
	Duration  int                `json:"duration"`           // Total duration in seconds
	Remaining int                `json:"remaining"`          // Remaining time in seconds
	Status    TimerStatus        `json:"status"`
}

func NewTimerEvent(data TimerEventData) entities.Event[TimerEventData] {
	return entities.Event[TimerEventData]{
		Channel: entities.EventChannelTimer,
		Type:    EventTypeTimer,
		Data:    data,
	}
}

// Helper functions for creating timer events
func NewTimerStartedEvent(phase entities.GamePhase, roleType *entities.RoleType, duration int) entities.Event[TimerEventData] {
	return NewTimerEvent(TimerEventData{
		Phase:     phase,
		RoleType:  roleType,
		Duration:  duration,
		Remaining: duration,
		Status:    TimerStatusStarted,
	})
}

func NewTimerTickEvent(phase entities.GamePhase, roleType *entities.RoleType, duration, remaining int) entities.Event[TimerEventData] {
	return NewTimerEvent(TimerEventData{
		Phase:     phase,
		RoleType:  roleType,
		Duration:  duration,
		Remaining: remaining,
		Status:    TimerStatusTick,
	})
}

func NewTimerExpiredEvent(phase entities.GamePhase, roleType *entities.RoleType) entities.Event[TimerEventData] {
	return NewTimerEvent(TimerEventData{
		Phase:    phase,
		RoleType: roleType,
		Status:   TimerStatusExpired,
	})
}

func NewTimerSkippedEvent(phase entities.GamePhase, roleType *entities.RoleType) entities.Event[TimerEventData] {
	return NewTimerEvent(TimerEventData{
		Phase:    phase,
		RoleType: roleType,
		Status:   TimerStatusSkipped,
	})
}

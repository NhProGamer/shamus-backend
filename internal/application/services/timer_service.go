package services

import (
	"context"
	"encoding/json"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/events"
	"shamus-backend/internal/domain/ports"
	"sync"
	"sync/atomic"
	"time"
)

// Timer durations (hardcoded as per requirements)
const (
	VotePhaseDuration     = 60 * time.Second
	DayPhaseDuration      = 120 * time.Second
	SeerTimerDuration     = 20 * time.Second
	WerewolfTimerDuration = 30 * time.Second
	WitchTimerDuration    = 30 * time.Second
)

// GetRoleTimerDuration returns the timer duration for a specific role during night phase
func GetRoleTimerDuration(roleType entities.RoleType) time.Duration {
	switch roleType {
	case entities.RoleSeer:
		return SeerTimerDuration
	case entities.RoleWerewolf:
		return WerewolfTimerDuration
	case entities.RoleWitch:
		return WitchTimerDuration
	default:
		return 0
	}
}

// TimerCallback is called when a timer expires
type TimerCallback func(gameID entities.GameID, phase entities.GamePhase, roleType *entities.RoleType)

// GameTimer holds the state of a timer for a game
type GameTimer struct {
	cancel    context.CancelFunc
	phase     entities.GamePhase
	roleType  *entities.RoleType
	duration  time.Duration
	startTime time.Time
	handled   atomic.Bool // Atomic flag to prevent double handling
}

// TimerService manages phase and role timers for games
type TimerService struct {
	timers      map[entities.GameID]*GameTimer
	broadcaster ports.Broadcaster
	onExpiry    TimerCallback
	lock        sync.Mutex
}

// NewTimerService creates a new TimerService
func NewTimerService(broadcaster ports.Broadcaster) *TimerService {
	return &TimerService{
		timers:      make(map[entities.GameID]*GameTimer),
		broadcaster: broadcaster,
	}
}

// SetExpiryCallback sets the callback function for when timers expire
func (s *TimerService) SetExpiryCallback(callback TimerCallback) {
	s.onExpiry = callback
}

// StartPhaseTimer starts a timer for a game phase (day or vote)
func (s *TimerService) StartPhaseTimer(gameID entities.GameID, phase entities.GamePhase) {
	var duration time.Duration
	switch phase {
	case entities.PhaseDay:
		duration = DayPhaseDuration
	case entities.PhaseVote:
		duration = VotePhaseDuration
	default:
		return // No timer for other phases
	}

	s.startTimer(gameID, phase, nil, duration)
}

// StartRoleTimer starts a timer for a specific role during night phase
func (s *TimerService) StartRoleTimer(gameID entities.GameID, roleType entities.RoleType) {
	duration := GetRoleTimerDuration(roleType)
	if duration == 0 {
		return // No timer for this role
	}

	s.startTimer(gameID, entities.PhaseNight, &roleType, duration)
}

// startTimer is the internal method to start a timer
func (s *TimerService) startTimer(gameID entities.GameID, phase entities.GamePhase, roleType *entities.RoleType, duration time.Duration) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Cancel existing timer if any
	if existing, ok := s.timers[gameID]; ok {
		existing.cancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	timer := &GameTimer{
		cancel:    cancel,
		phase:     phase,
		roleType:  roleType,
		duration:  duration,
		startTime: time.Now(),
	}
	s.timers[gameID] = timer

	// Broadcast timer started event
	s.broadcastTimerEvent(gameID, events.NewTimerStartedEvent(phase, roleType, int(duration.Seconds())))

	// Start the timer goroutine
	go s.runTimer(ctx, gameID, timer)
}

// meaningfulIntervals defines which remaining seconds should trigger a broadcast
// This reduces WebSocket traffic while still giving players useful countdown info
var meaningfulIntervals = map[int]bool{
	120: true, 60: true, 30: true, 15: true, 10: true,
	5: true, 4: true, 3: true, 2: true, 1: true,
}

// runTimer runs the timer and broadcasts tick events at meaningful intervals
func (s *TimerService) runTimer(ctx context.Context, gameID entities.GameID, timer *GameTimer) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	durationSecs := int(timer.duration.Seconds())
	lastBroadcast := -1 // Track last broadcast to avoid duplicates

	for {
		select {
		case <-ctx.Done():
			// Timer was cancelled (either skipped, game ended, or new timer started)
			// The handler that cancelled it is responsible for cleanup
			return
		case <-ticker.C:
			elapsed := time.Since(timer.startTime)
			remaining := timer.duration - elapsed

			if remaining <= 0 {
				// Timer expired naturally - use atomic flag to prevent double handling
				if timer.handled.CompareAndSwap(false, true) {
					s.handleExpiry(gameID, timer)
				}
				return
			}

			// Only broadcast at meaningful intervals to reduce WebSocket traffic
			remainingSecs := int(remaining.Seconds())
			if meaningfulIntervals[remainingSecs] && remainingSecs != lastBroadcast {
				lastBroadcast = remainingSecs
				s.broadcastTimerEvent(gameID, events.NewTimerTickEvent(timer.phase, timer.roleType, durationSecs, remainingSecs))
			}
		}
	}
}

// handleExpiry handles timer expiration
func (s *TimerService) handleExpiry(gameID entities.GameID, timer *GameTimer) {
	s.lock.Lock()
	delete(s.timers, gameID)
	s.lock.Unlock()

	// Broadcast expired event
	s.broadcastTimerEvent(gameID, events.NewTimerExpiredEvent(timer.phase, timer.roleType))

	// Call the expiry callback
	if s.onExpiry != nil {
		s.onExpiry(gameID, timer.phase, timer.roleType)
	}
}

// CancelTimer cancels the current timer for a game (e.g., when game ends)
func (s *TimerService) CancelTimer(gameID entities.GameID) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if timer, ok := s.timers[gameID]; ok {
		timer.cancel()
		delete(s.timers, gameID)
	}
}

// SkipTimer skips the current timer (e.g., when all players have acted)
func (s *TimerService) SkipTimer(gameID entities.GameID) {
	s.lock.Lock()
	timer, ok := s.timers[gameID]
	if !ok {
		s.lock.Unlock()
		return
	}

	// Use atomic flag to prevent double handling with natural expiry
	if !timer.handled.CompareAndSwap(false, true) {
		// Already handled by natural expiry, just cleanup
		s.lock.Unlock()
		return
	}

	timer.cancel()
	delete(s.timers, gameID)
	s.lock.Unlock()

	// Broadcast skipped event
	s.broadcastTimerEvent(gameID, events.NewTimerSkippedEvent(timer.phase, timer.roleType))

	// Call the expiry callback (same as expiry, just faster)
	if s.onExpiry != nil {
		s.onExpiry(gameID, timer.phase, timer.roleType)
	}
}

// GetRemainingTime returns the remaining time for a game's timer
func (s *TimerService) GetRemainingTime(gameID entities.GameID) time.Duration {
	s.lock.Lock()
	defer s.lock.Unlock()

	timer, ok := s.timers[gameID]
	if !ok {
		return 0
	}

	elapsed := time.Since(timer.startTime)
	remaining := timer.duration - elapsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// broadcastTimerEvent broadcasts a timer event to all players in a game
func (s *TimerService) broadcastTimerEvent(gameID entities.GameID, event entities.Event[events.TimerEventData]) {
	if s.broadcaster == nil {
		return
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return
	}

	s.broadcaster.BroadcastToGame(gameID, payload)
}

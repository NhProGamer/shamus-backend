package app

import (
	"context"
	"encoding/json"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/events"
	apperrors "shamus-backend/internal/domain/errors"
	"shamus-backend/internal/domain/ports"
	"shamus-backend/internal/infrastructure/logger"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ActionService manages player actions with timeouts and callbacks
type ActionService struct {
	actionRepo   ports.ActionRepository
	playerSender ports.PlayerSender
	callbacks    map[entities.ActionType]ports.ActionCallback
	timers       map[entities.ActionID]*time.Timer
	lock         sync.RWMutex
}

// NewActionService creates a new ActionService
func NewActionService(actionRepo ports.ActionRepository, playerSender ports.PlayerSender) *ActionService {
	return &ActionService{
		actionRepo:   actionRepo,
		playerSender: playerSender,
		callbacks:    make(map[entities.ActionType]ports.ActionCallback),
		timers:       make(map[entities.ActionID]*time.Timer),
	}
}

// CreateAction creates a new action and sends it to the player
func (s *ActionService) CreateAction(
	gameID entities.GameID,
	playerID entities.PlayerID,
	actionType entities.ActionType,
	payload interface{},
	timeout time.Duration,
) (*entities.Action, error) {
	ctx := context.TODO()

	// Serialize payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, apperrors.Wrap("INVALID_PAYLOAD", "failed to serialize action payload", err)
	}

	// Create action
	actionID := entities.ActionID(uuid.New().String())
	action := entities.NewAction(actionID, actionType, gameID, playerID, payloadBytes, timeout)

	// Save action to repository
	if err := s.actionRepo.SaveAction(ctx, action); err != nil {
		return nil, err
	}

	// Start timeout timer
	s.startTimer(action)

	// Send action to player
	if err := s.sendActionToPlayer(action); err != nil {
		logger.Get().Warn().
			Str("actionID", string(actionID)).
			Str("playerID", string(playerID)).
			Err(err).
			Msg("Failed to send action to player")
		// Don't return error - timer will still expire and handle it
	}

	logger.Get().Info().
		Str("actionID", string(actionID)).
		Str("type", string(actionType)).
		Str("playerID", string(playerID)).
		Str("gameID", string(gameID)).
		Dur("timeout", timeout).
		Msg("Action created")

	return action, nil
}

// RespondToAction processes a player's response to an action
func (s *ActionService) RespondToAction(actionID entities.ActionID, playerID entities.PlayerID, response interface{}) error {
	ctx := context.TODO()

	// Get action
	action, err := s.actionRepo.GetAction(ctx, actionID)
	if err != nil {
		return err
	}
	if action == nil {
		return apperrors.ErrActionNotFound
	}

	// Validate player owns this action
	if action.PlayerID != playerID {
		return apperrors.ErrActionWrongPlayer
	}

	// Check if action can still receive response
	if !action.CanRespond() {
		if action.Status == entities.ActionStatusExpired {
			return apperrors.ErrActionExpired
		}
		if action.Status == entities.ActionStatusCompleted {
			return apperrors.ErrActionAlreadyCompleted
		}
		if action.Status == entities.ActionStatusCancelled {
			return apperrors.ErrActionCancelled
		}
		return apperrors.New("ACTION_INVALID_STATE", "action is in an invalid state")
	}

	// Serialize response to JSON
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return apperrors.Wrap("INVALID_RESPONSE", "failed to serialize action response", err)
	}

	// Cancel timer
	s.cancelTimer(actionID)

	// Mark action as completed
	action.MarkCompleted(responseBytes)
	if err := s.actionRepo.SaveAction(ctx, action); err != nil {
		return err
	}

	logger.Get().Info().
		Str("actionID", string(actionID)).
		Str("type", string(action.Type)).
		Str("playerID", string(playerID)).
		Msg("Action completed")

	// Invoke callback
	return s.invokeCallback(action, responseBytes)
}

// CancelAction manually cancels a pending action
func (s *ActionService) CancelAction(actionID entities.ActionID) error {
	ctx := context.TODO()

	// Get action
	action, err := s.actionRepo.GetAction(ctx, actionID)
	if err != nil {
		return err
	}
	if action == nil {
		return apperrors.ErrActionNotFound
	}

	// Only cancel if pending
	if action.Status != entities.ActionStatusPending {
		return nil // Already handled, no error
	}

	// Cancel timer
	s.cancelTimer(actionID)

	// Mark action as cancelled
	action.MarkCancelled()
	if err := s.actionRepo.SaveAction(ctx, action); err != nil {
		return err
	}

	logger.Get().Info().
		Str("actionID", string(actionID)).
		Str("type", string(action.Type)).
		Msg("Action cancelled")

	return nil
}

// GetAction retrieves an action by ID
func (s *ActionService) GetAction(actionID entities.ActionID) (*entities.Action, error) {
	return s.actionRepo.GetAction(context.TODO(), actionID)
}

// GetPendingActions retrieves all pending actions for a player
func (s *ActionService) GetPendingActions(playerID entities.PlayerID) ([]*entities.Action, error) {
	return s.actionRepo.GetPendingActionsByPlayer(context.TODO(), playerID)
}

// RegisterCallback registers a callback function for a specific action type
func (s *ActionService) RegisterCallback(actionType entities.ActionType, callback ports.ActionCallback) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.callbacks[actionType] = callback
}

// startTimer starts a timeout timer for an action
func (s *ActionService) startTimer(action *entities.Action) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Create timer that will fire when action expires
	timer := time.AfterFunc(action.Timeout, func() {
		s.handleTimeout(action.ID)
	})

	s.timers[action.ID] = timer
}

// cancelTimer cancels an action's timeout timer
func (s *ActionService) cancelTimer(actionID entities.ActionID) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if timer, exists := s.timers[actionID]; exists {
		timer.Stop()
		delete(s.timers, actionID)
	}
}

// handleTimeout is called when an action's timer expires
func (s *ActionService) handleTimeout(actionID entities.ActionID) {
	ctx := context.TODO()

	// Remove timer from map
	s.lock.Lock()
	delete(s.timers, actionID)
	s.lock.Unlock()

	// Get action
	action, err := s.actionRepo.GetAction(ctx, actionID)
	if err != nil {
		logger.Get().Error().
			Str("actionID", string(actionID)).
			Err(err).
			Msg("Failed to get action on timeout")
		return
	}
	if action == nil {
		// Action was already deleted
		return
	}

	// Only handle if still pending
	if action.Status != entities.ActionStatusPending {
		return
	}

	// Mark action as expired
	action.MarkExpired()
	if err := s.actionRepo.SaveAction(ctx, action); err != nil {
		logger.Get().Error().
			Str("actionID", string(actionID)).
			Err(err).
			Msg("Failed to save expired action")
	}

	logger.Get().Info().
		Str("actionID", string(actionID)).
		Str("type", string(action.Type)).
		Str("playerID", string(action.PlayerID)).
		Msg("Action expired")

	// Invoke callback with nil response (timeout)
	if err := s.invokeCallback(action, nil); err != nil {
		logger.Get().Error().
			Str("actionID", string(actionID)).
			Err(err).
			Msg("Callback error on timeout")
	}
}

// invokeCallback calls the registered callback for an action's type
func (s *ActionService) invokeCallback(action *entities.Action, response json.RawMessage) error {
	s.lock.RLock()
	callback, exists := s.callbacks[action.Type]
	s.lock.RUnlock()

	if !exists {
		logger.Get().Warn().
			Str("actionID", string(action.ID)).
			Str("type", string(action.Type)).
			Msg("No callback registered for action type")
		return nil
	}

	return callback(action.GameID, action.PlayerID, action, response)
}

// sendActionToPlayer sends an action to the player via PlayerSender
func (s *ActionService) sendActionToPlayer(action *entities.Action) error {
	// Deserialize payload to interface{} for event
	var payload interface{}
	if err := json.Unmarshal(action.Payload, &payload); err != nil {
		logger.Get().Error().
			Str("actionID", string(action.ID)).
			Err(err).
			Msg("Failed to unmarshal action payload for sending")
		return err
	}

	// Create action created event using helper
	event := events.NewActionCreatedEvent(
		string(action.ID),
		action.Type,
		payload,
		action.ExpiresAt,
		action.Timeout,
	)

	// Serialize event to JSON
	eventBytes, err := json.Marshal(event)
	if err != nil {
		logger.Get().Error().
			Str("actionID", string(action.ID)).
			Err(err).
			Msg("Failed to marshal action event")
		return err
	}

	// Send to player
	if err := s.playerSender.SendToPlayer(action.PlayerID, eventBytes); err != nil {
		logger.Get().Warn().
			Str("actionID", string(action.ID)).
			Str("playerID", string(action.PlayerID)).
			Err(err).
			Msg("Failed to send action to player")
		return err
	}

	logger.Get().Debug().
		Str("actionID", string(action.ID)).
		Str("playerID", string(action.PlayerID)).
		Str("type", string(action.Type)).
		Msg("Action sent to player")

	return nil
}

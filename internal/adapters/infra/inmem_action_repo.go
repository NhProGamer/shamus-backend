package infra

import (
	"context"
	"shamus-backend/internal/domain/entities"
	"sync"
)

// InMemoryActionRepository provides an in-memory implementation of ActionRepository
// This is used instead of Redis because actions don't need to survive server restarts
// (if the server crashes, all active games are lost anyway)
type InMemoryActionRepository struct {
	actions map[entities.ActionID]*entities.Action
	lock    sync.RWMutex
}

// NewInMemoryActionRepository creates a new in-memory action repository
func NewInMemoryActionRepository() *InMemoryActionRepository {
	return &InMemoryActionRepository{
		actions: make(map[entities.ActionID]*entities.Action),
	}
}

// SaveAction persists an action to memory
func (r *InMemoryActionRepository) SaveAction(ctx context.Context, action *entities.Action) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	// Create a copy to prevent external mutations
	actionCopy := *action
	r.actions[action.ID] = &actionCopy
	return nil
}

// GetAction retrieves an action by ID
func (r *InMemoryActionRepository) GetAction(ctx context.Context, actionID entities.ActionID) (*entities.Action, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	action, exists := r.actions[actionID]
	if !exists {
		return nil, nil // Action not found, not an error
	}

	// Return a copy to prevent external mutations
	actionCopy := *action
	return &actionCopy, nil
}

// GetPendingActionsByPlayer retrieves all pending actions for a player
func (r *InMemoryActionRepository) GetPendingActionsByPlayer(ctx context.Context, playerID entities.PlayerID) ([]*entities.Action, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	var pendingActions []*entities.Action
	for _, action := range r.actions {
		if action.PlayerID == playerID && action.Status == entities.ActionStatusPending {
			// Return a copy to prevent external mutations
			actionCopy := *action
			pendingActions = append(pendingActions, &actionCopy)
		}
	}

	return pendingActions, nil
}

// DeleteAction removes an action from memory
func (r *InMemoryActionRepository) DeleteAction(ctx context.Context, actionID entities.ActionID) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.actions, actionID)
	return nil
}

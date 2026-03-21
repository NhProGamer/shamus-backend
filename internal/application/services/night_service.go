package services

import (
	"encoding/json"
	"shamus-backend/pkg/logger"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/events"
	"shamus-backend/internal/domain/ports"
	"sync"
)

// SeerAction represents the seer's action for the night
type SeerAction struct {
	TargetID entities.PlayerID
	Result   *entities.RoleType // The revealed role
}

// WitchAction represents the witch's actions for the night
type WitchAction struct {
	HealTargetID   *entities.PlayerID // Who to save (if any)
	PoisonTargetID *entities.PlayerID // Who to poison (if any)
}

// NightState holds the state of a night phase for a game
type NightState struct {
	GameID          entities.GameID
	CurrentPhase    entities.NightPhase
	SeerAction      *SeerAction
	WerewolfVictim  *entities.PlayerID // Result of werewolf vote
	WitchAction     *WitchAction
	PendingDeaths   []entities.PlayerID // Players to kill at dawn
	HasSeer         bool                // Does the game have a living seer?
	HasWerewolves   bool                // Does the game have living werewolves?
	HasWitch        bool                // Does the game have a living witch?
	WitchCanHeal    bool                // Can the witch still heal?
	WitchCanPoison  bool                // Can the witch still poison?
	SeerHasActed    bool
	WerewolvesVoted bool
	WitchHasActed   bool
}

// NightService manages the night phase actions
type NightService struct {
	nightStates  map[entities.GameID]*NightState
	playerSender ports.PlayerSender
	voteService  *VoteService
	lock         sync.RWMutex
}

// NewNightService creates a new NightService
func NewNightService(playerSender ports.PlayerSender, voteService *VoteService) *NightService {
	return &NightService{
		nightStates:  make(map[entities.GameID]*NightState),
		playerSender: playerSender,
		voteService:  voteService,
	}
}

// StartNight initializes a new night phase for a game
func (s *NightService) StartNight(gameID entities.GameID, players []*entities.Player) *NightState {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Analyze which special roles are alive
	hasSeer := false
	hasWerewolves := false
	hasWitch := false
	witchCanHeal := false
	witchCanPoison := false

	for _, p := range players {
		if !p.IsAlive || p.Role == nil {
			continue
		}
		switch p.Role.GetType() {
		case entities.RoleSeer:
			hasSeer = true
		case entities.RoleWerewolf:
			hasWerewolves = true
		case entities.RoleWitch:
			hasWitch = true
			// Check witch's remaining abilities
			abilities := p.Role.GetAbilities()
			if abilities != nil {
				for _, ability := range *abilities {
					if ability.GetName() == "Heal" && ability.CanUse(nil, p) {
						witchCanHeal = true
					}
					if ability.GetName() == "Poison" && ability.CanUse(nil, p) {
						witchCanPoison = true
					}
				}
			}
		}
	}

	state := &NightState{
		GameID:          gameID,
		CurrentPhase:    entities.NightPhaseSeer, // Start with seer
		HasSeer:         hasSeer,
		HasWerewolves:   hasWerewolves,
		HasWitch:        hasWitch,
		WitchCanHeal:    witchCanHeal,
		WitchCanPoison:  witchCanPoison,
		PendingDeaths:   []entities.PlayerID{},
		SeerHasActed:    !hasSeer,       // If no seer, mark as acted
		WerewolvesVoted: !hasWerewolves, // If no werewolves, mark as voted
		WitchHasActed:   !hasWitch,      // If no witch, mark as acted
	}

	s.nightStates[gameID] = state

	return state
}

// GetNightState returns a copy of the current night state for a game
// This returns a copy to prevent external modification of internal state
func (s *NightService) GetNightState(gameID entities.GameID) (*NightState, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	state, exists := s.nightStates[gameID]
	if !exists {
		return nil, false
	}

	// Return a deep copy to prevent external modification
	stateCopy := *state

	// Copy slices to prevent shared references
	if state.PendingDeaths != nil {
		stateCopy.PendingDeaths = make([]entities.PlayerID, len(state.PendingDeaths))
		copy(stateCopy.PendingDeaths, state.PendingDeaths)
	}

	// Copy pointer values
	if state.SeerAction != nil {
		seerCopy := *state.SeerAction
		if state.SeerAction.Result != nil {
			resultCopy := *state.SeerAction.Result
			seerCopy.Result = &resultCopy
		}
		stateCopy.SeerAction = &seerCopy
	}

	if state.WerewolfVictim != nil {
		victimCopy := *state.WerewolfVictim
		stateCopy.WerewolfVictim = &victimCopy
	}

	if state.WitchAction != nil {
		witchCopy := *state.WitchAction
		if state.WitchAction.HealTargetID != nil {
			healCopy := *state.WitchAction.HealTargetID
			witchCopy.HealTargetID = &healCopy
		}
		if state.WitchAction.PoisonTargetID != nil {
			poisonCopy := *state.WitchAction.PoisonTargetID
			witchCopy.PoisonTargetID = &poisonCopy
		}
		stateCopy.WitchAction = &witchCopy
	}

	return &stateCopy, true
}

// GetCurrentPhase returns the current night phase
func (s *NightService) GetCurrentPhase(gameID entities.GameID) string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	state, exists := s.nightStates[gameID]
	if !exists {
		return string(entities.NightPhaseEnd)
	}
	return string(state.CurrentPhase)
}

// ShouldSkipPhase checks if a phase should be skipped (no alive players with that role)
func (s *NightService) ShouldSkipPhase(gameID entities.GameID, phase entities.NightPhase) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	state, exists := s.nightStates[gameID]
	if !exists {
		return true
	}

	switch phase {
	case entities.NightPhaseSeer:
		return !state.HasSeer
	case entities.NightPhaseWerewolf:
		return !state.HasWerewolves
	case entities.NightPhaseWitch:
		return !state.HasWitch
	default:
		return false
	}
}

// AdvancePhase moves to the next night phase
// Returns the new phase
func (s *NightService) AdvancePhase(gameID entities.GameID) entities.NightPhase {
	s.lock.Lock()
	defer s.lock.Unlock()

	state, exists := s.nightStates[gameID]
	if !exists {
		return entities.NightPhaseEnd
	}

	// Find current phase index
	currentIdx := -1
	for i, phase := range entities.NightPhaseOrder {
		if phase == state.CurrentPhase {
			currentIdx = i
			break
		}
	}

	// Move to next phase
	if currentIdx >= len(entities.NightPhaseOrder)-1 {
		state.CurrentPhase = entities.NightPhaseEnd
	} else {
		state.CurrentPhase = entities.NightPhaseOrder[currentIdx+1]
	}

	return state.CurrentPhase
}

// RecordSeerAction records the seer's vision
func (s *NightService) RecordSeerAction(gameID entities.GameID, targetID entities.PlayerID, revealedRole entities.RoleType) {
	s.lock.Lock()
	defer s.lock.Unlock()

	state, exists := s.nightStates[gameID]
	if !exists {
		return
	}

	state.SeerAction = &SeerAction{
		TargetID: targetID,
		Result:   &revealedRole,
	}
	state.SeerHasActed = true
}

// RecordWerewolfVictim records the werewolf vote result
func (s *NightService) RecordWerewolfVictim(gameID entities.GameID, victimID *entities.PlayerID) {
	s.lock.Lock()
	defer s.lock.Unlock()

	state, exists := s.nightStates[gameID]
	if !exists {
		return
	}

	state.WerewolfVictim = victimID
	state.WerewolvesVoted = true

	// Add to pending deaths if there's a victim
	if victimID != nil {
		addToPendingDeaths(state, *victimID)
	}
}

// RecordWitchAction records the witch's actions
func (s *NightService) RecordWitchAction(gameID entities.GameID, healTargetID, poisonTargetID *entities.PlayerID) {
	s.lock.Lock()
	defer s.lock.Unlock()

	state, exists := s.nightStates[gameID]
	if !exists {
		return
	}

	state.WitchAction = &WitchAction{
		HealTargetID:   healTargetID,
		PoisonTargetID: poisonTargetID,
	}
	state.WitchHasActed = true

	// Process witch actions
	// If heal was used, remove victim from pending deaths
	if healTargetID != nil {
		newDeaths := []entities.PlayerID{}
		for _, id := range state.PendingDeaths {
			if id != *healTargetID {
				newDeaths = append(newDeaths, id)
			}
		}
		state.PendingDeaths = newDeaths
	}

	// If poison was used, add target to pending deaths (with uniqueness check)
	if poisonTargetID != nil {
		addToPendingDeaths(state, *poisonTargetID)
	}
}

// GetPendingDeaths returns a copy of the list of players who will die at dawn
func (s *NightService) GetPendingDeaths(gameID entities.GameID) []entities.PlayerID {
	s.lock.RLock()
	defer s.lock.RUnlock()

	state, exists := s.nightStates[gameID]
	if !exists {
		return nil
	}

	// Return a copy to prevent external modification
	result := make([]entities.PlayerID, len(state.PendingDeaths))
	copy(result, state.PendingDeaths)
	return result
}

// GetWerewolfVictim returns a copy of the werewolf victim (before witch intervention)
func (s *NightService) GetWerewolfVictim(gameID entities.GameID) *entities.PlayerID {
	s.lock.RLock()
	defer s.lock.RUnlock()

	state, exists := s.nightStates[gameID]
	if !exists || state.WerewolfVictim == nil {
		return nil
	}

	// Return a copy to prevent external modification
	victimCopy := *state.WerewolfVictim
	return &victimCopy
}

// IsNightComplete checks if all night actions are done
func (s *NightService) IsNightComplete(gameID entities.GameID) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	state, exists := s.nightStates[gameID]
	if !exists {
		return true
	}

	return state.SeerHasActed && state.WerewolvesVoted && state.WitchHasActed
}

// ClearNight removes the night state for a game
func (s *NightService) ClearNight(gameID entities.GameID) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.nightStates, gameID)
}

// SendTurnEvent sends a turn event to the appropriate players
func (s *NightService) SendTurnEvent(gameID entities.GameID, phase entities.NightPhase, players []*entities.Player) {
	if s.playerSender == nil {
		return
	}

	s.lock.RLock()
	state, exists := s.nightStates[gameID]
	if !exists {
		s.lock.RUnlock()
		return
	}

	// Copy values we need before releasing lock
	werewolfVictim := state.WerewolfVictim
	canHeal := state.WitchCanHeal
	canPoison := state.WitchCanPoison
	s.lock.RUnlock()

	switch phase {
	case entities.NightPhaseSeer:
		// Send to seer only
		for _, p := range players {
			if p.IsAlive && p.Role != nil && p.Role.GetType() == entities.RoleSeer {
				event := events.NewTurnEvent(entities.RoleSeer)
				payload, err := json.Marshal(event)
				if err != nil {
					logger.Get().Info().Msgf("Error marshaling seer turn event: %v", err)
					return
				}
				s.playerSender.SendToPlayer(p.ID, payload)
				break
			}
		}

	case entities.NightPhaseWerewolf:
		// Send to all werewolves
		event := events.NewTurnEvent(entities.RoleWerewolf)
		payload, err := json.Marshal(event)
		if err != nil {
			logger.Get().Info().Msgf("Error marshaling werewolf turn event: %v", err)
			return
		}
		for _, p := range players {
			if p.IsAlive && p.Role != nil && p.Role.GetType() == entities.RoleWerewolf {
				s.playerSender.SendToPlayer(p.ID, payload)
			}
		}

	case entities.NightPhaseWitch:
		// Send to witch with victim info
		for _, p := range players {
			if p.IsAlive && p.Role != nil && p.Role.GetType() == entities.RoleWitch {
				event := events.NewWitchTurnEvent(werewolfVictim, canHeal, canPoison)
				payload, err := json.Marshal(event)
				if err != nil {
					logger.Get().Info().Msgf("Error marshaling witch turn event: %v", err)
					return
				}
				s.playerSender.SendToPlayer(p.ID, payload)
				break
			}
		}
	}
}

// GetPlayersWithRole returns all alive players with a specific role
func GetPlayersWithRole(players []*entities.Player, roleType entities.RoleType) []*entities.Player {
	var result []*entities.Player
	for _, p := range players {
		if p.IsAlive && p.Role != nil && p.Role.GetType() == roleType {
			result = append(result, p)
		}
	}
	return result
}

// GetNonWerewolfPlayers returns all alive players who are not werewolves
func GetNonWerewolfPlayers(players []*entities.Player) []*entities.Player {
	var result []*entities.Player
	for _, p := range players {
		if p.IsAlive && p.Role != nil && p.Role.GetType() != entities.RoleWerewolf {
			result = append(result, p)
		}
	}
	return result
}

// CanSeerAct checks if the seer can act (correct phase + hasn't acted)
func (s *NightService) CanSeerAct(gameID entities.GameID) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	state, exists := s.nightStates[gameID]
	if !exists {
		return false
	}

	return state.CurrentPhase == entities.NightPhaseSeer && !state.SeerHasActed
}

// CanWerewolvesVote checks if werewolves can vote
func (s *NightService) CanWerewolvesVote(gameID entities.GameID) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	state, exists := s.nightStates[gameID]
	if !exists {
		return false
	}

	return state.CurrentPhase == entities.NightPhaseWerewolf && !state.WerewolvesVoted
}

// CanWitchAct checks if the witch can act
func (s *NightService) CanWitchAct(gameID entities.GameID) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	state, exists := s.nightStates[gameID]
	if !exists {
		return false
	}

	return state.CurrentPhase == entities.NightPhaseWitch && !state.WitchHasActed
}

// GetWitchAbilities returns whether the witch can heal and poison
func (s *NightService) GetWitchAbilities(gameID entities.GameID) (canHeal, canPoison bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	state, exists := s.nightStates[gameID]
	if !exists {
		return false, false
	}

	return state.WitchCanHeal, state.WitchCanPoison
}

// addToPendingDeaths adds a player to pending deaths if not already present.
// This ensures each player appears at most once in the death list.
// Note: caller must hold the lock.
func addToPendingDeaths(state *NightState, playerID entities.PlayerID) {
	for _, id := range state.PendingDeaths {
		if id == playerID {
			return // Already in list
		}
	}
	state.PendingDeaths = append(state.PendingDeaths, playerID)
}

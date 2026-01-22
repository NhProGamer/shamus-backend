package app

import (
	"encoding/json"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/events"
	"sync"
)

// NightPhase represents the current phase of the night
type NightPhase string

const (
	NightPhaseSeer     NightPhase = "seer"
	NightPhaseWerewolf NightPhase = "werewolf"
	NightPhaseWitch    NightPhase = "witch"
	NightPhaseEnd      NightPhase = "end"
)

// NightPhaseOrder defines the order of night phases (by role priority)
var NightPhaseOrder = []NightPhase{
	NightPhaseSeer,     // Priority 10
	NightPhaseWerewolf, // Priority 9
	NightPhaseWitch,    // Priority 8
}

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
	CurrentPhase    NightPhase
	SeerAction      *SeerAction
	WerewolfVictim  *entities.PlayerID // Result of werewolf vote
	WitchAction     *WitchAction
	PendingDeaths   []entities.PlayerID // Players to kill at dawn
	HasSeer         bool                // Does the game have a living seer?
	HasWitch        bool                // Does the game have a living witch?
	WitchCanHeal    bool                // Can the witch still heal?
	WitchCanPoison  bool                // Can the witch still poison?
	SeerHasActed    bool
	WerewolvesVoted bool
	WitchHasActed   bool
}

// PlayerSender is an interface for sending events to specific players
type PlayerSender interface {
	SendToPlayer(playerID entities.PlayerID, payload []byte) error
	BroadcastToGame(gameID entities.GameID, payload []byte) error
}

// NightService manages the night phase actions
type NightService struct {
	nightStates  map[entities.GameID]*NightState
	playerSender PlayerSender
	voteService  *VoteService
	lock         sync.RWMutex
}

// NewNightService creates a new NightService
func NewNightService(playerSender PlayerSender, voteService *VoteService) *NightService {
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
		CurrentPhase:    NightPhaseSeer, // Start with seer
		HasSeer:         hasSeer,
		HasWitch:        hasWitch,
		WitchCanHeal:    witchCanHeal,
		WitchCanPoison:  witchCanPoison,
		PendingDeaths:   []entities.PlayerID{},
		SeerHasActed:    !hasSeer, // If no seer, mark as acted
		WerewolvesVoted: false,
		WitchHasActed:   !hasWitch, // If no witch, mark as acted
	}

	s.nightStates[gameID] = state

	return state
}

// GetNightState returns the current night state for a game
func (s *NightService) GetNightState(gameID entities.GameID) (*NightState, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	state, exists := s.nightStates[gameID]
	return state, exists
}

// GetCurrentPhase returns the current night phase
func (s *NightService) GetCurrentPhase(gameID entities.GameID) string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	state, exists := s.nightStates[gameID]
	if !exists {
		return string(NightPhaseEnd)
	}
	return string(state.CurrentPhase)
}

// ShouldSkipPhase checks if a phase should be skipped (no alive players with that role)
func (s *NightService) ShouldSkipPhase(gameID entities.GameID, phase NightPhase) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	state, exists := s.nightStates[gameID]
	if !exists {
		return true
	}

	switch phase {
	case NightPhaseSeer:
		return !state.HasSeer
	case NightPhaseWitch:
		return !state.HasWitch
	default:
		return false
	}
}

// AdvancePhase moves to the next night phase
// Returns the new phase
func (s *NightService) AdvancePhase(gameID entities.GameID) NightPhase {
	s.lock.Lock()
	defer s.lock.Unlock()

	state, exists := s.nightStates[gameID]
	if !exists {
		return NightPhaseEnd
	}

	// Find current phase index
	currentIdx := -1
	for i, phase := range NightPhaseOrder {
		if phase == state.CurrentPhase {
			currentIdx = i
			break
		}
	}

	// Move to next phase
	if currentIdx >= len(NightPhaseOrder)-1 {
		state.CurrentPhase = NightPhaseEnd
	} else {
		state.CurrentPhase = NightPhaseOrder[currentIdx+1]
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
		state.PendingDeaths = append(state.PendingDeaths, *victimID)
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

	// If poison was used, add target to pending deaths
	if poisonTargetID != nil {
		state.PendingDeaths = append(state.PendingDeaths, *poisonTargetID)
	}
}

// GetPendingDeaths returns the list of players who will die at dawn
func (s *NightService) GetPendingDeaths(gameID entities.GameID) []entities.PlayerID {
	s.lock.RLock()
	defer s.lock.RUnlock()

	state, exists := s.nightStates[gameID]
	if !exists {
		return nil
	}

	return state.PendingDeaths
}

// GetWerewolfVictim returns the werewolf victim (before witch intervention)
func (s *NightService) GetWerewolfVictim(gameID entities.GameID) *entities.PlayerID {
	s.lock.RLock()
	defer s.lock.RUnlock()

	state, exists := s.nightStates[gameID]
	if !exists {
		return nil
	}

	return state.WerewolfVictim
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
func (s *NightService) SendTurnEvent(gameID entities.GameID, phase NightPhase, players []*entities.Player) {
	if s.playerSender == nil {
		return
	}

	s.lock.RLock()
	state := s.nightStates[gameID]
	s.lock.RUnlock()

	switch phase {
	case NightPhaseSeer:
		// Send to seer only
		for _, p := range players {
			if p.IsAlive && p.Role != nil && p.Role.GetType() == entities.RoleSeer {
				event := events.NewTurnEvent(entities.RoleSeer)
				payload, _ := json.Marshal(event)
				s.playerSender.SendToPlayer(p.ID, payload)
				break
			}
		}

	case NightPhaseWerewolf:
		// Send to all werewolves
		event := events.NewTurnEvent(entities.RoleWerewolf)
		payload, _ := json.Marshal(event)
		for _, p := range players {
			if p.IsAlive && p.Role != nil && p.Role.GetType() == entities.RoleWerewolf {
				s.playerSender.SendToPlayer(p.ID, payload)
			}
		}

	case NightPhaseWitch:
		// Send to witch with victim info
		for _, p := range players {
			if p.IsAlive && p.Role != nil && p.Role.GetType() == entities.RoleWitch {
				var canHeal, canPoison bool
				if state != nil {
					canHeal = state.WitchCanHeal
					canPoison = state.WitchCanPoison
				}
				event := events.NewWitchTurnEvent(state.WerewolfVictim, canHeal, canPoison)
				payload, _ := json.Marshal(event)
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

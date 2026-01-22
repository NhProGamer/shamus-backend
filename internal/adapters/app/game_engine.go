package app

import (
	"encoding/json"
	"log"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/events"
	"shamus-backend/internal/domain/ports"
)

// WinResult represents the result of a win condition check
type WinResult struct {
	GameEnded   bool
	WinningClan entities.Clan
	Winners     []entities.PlayerID
}

// GameEngine orchestrates game flow and phase transitions
type GameEngine struct {
	gameRepo     ports.GameRepository
	playerRepo   ports.PlayerRepository
	timerService *TimerService
	voteService  *VoteService
	nightService *NightService
	broadcaster  ports.Broadcaster
	playerSender ports.PlayerSender
}

// NewGameEngine creates a new GameEngine
func NewGameEngine(
	gameRepo ports.GameRepository,
	playerRepo ports.PlayerRepository,
	timerService *TimerService,
	voteService *VoteService,
	nightService *NightService,
	broadcaster ports.Broadcaster,
	playerSender ports.PlayerSender,
) *GameEngine {
	engine := &GameEngine{
		gameRepo:     gameRepo,
		playerRepo:   playerRepo,
		timerService: timerService,
		voteService:  voteService,
		nightService: nightService,
		broadcaster:  broadcaster,
		playerSender: playerSender,
	}

	// Set up timer expiry callback
	if timerService != nil {
		timerService.SetExpiryCallback(engine.handleTimerExpiry)
	}

	return engine
}

// handleTimerExpiry is called when a timer expires
func (e *GameEngine) handleTimerExpiry(gameID entities.GameID, phase entities.GamePhase, roleType *entities.RoleType) {
	log.Printf("Timer expired for game %s, phase %s, role %v", gameID, phase, roleType)

	switch phase {
	case entities.PhaseDay:
		// Day discussion time is over, start vote
		e.TransitionToVote(gameID)

	case entities.PhaseVote:
		// Vote time is over, resolve and transition
		e.ProcessVoteResult(gameID)

	case entities.PhaseNight:
		// Night role timer expired, advance to next role
		if roleType != nil {
			e.handleNightRoleTimeout(gameID, *roleType)
		}
	}
}

// handleNightRoleTimeout handles when a night role's timer expires
func (e *GameEngine) handleNightRoleTimeout(gameID entities.GameID, roleType entities.RoleType) {
	state, exists := e.nightService.GetNightState(gameID)
	if !exists {
		return
	}

	// Mark the role as having acted (even if they didn't do anything)
	switch roleType {
	case entities.RoleSeer:
		// Seer didn't act, mark as done
		e.nightService.RecordSeerAction(gameID, "", entities.RoleType(""))
	case entities.RoleWerewolf:
		// Werewolves didn't decide, resolve vote anyway
		result, err := e.voteService.ResolveVote(gameID)
		if err == nil {
			e.nightService.RecordWerewolfVictim(gameID, result.Target)
		} else {
			e.nightService.RecordWerewolfVictim(gameID, nil)
		}
		e.voteService.ClearVote(gameID)
	case entities.RoleWitch:
		// Witch didn't act, mark as done
		e.nightService.RecordWitchAction(gameID, nil, nil)
	}

	// Advance to next night phase
	e.advanceNightPhase(gameID, state)
}

// StartGameFlow is called after StartGame to begin the game flow
func (e *GameEngine) StartGameFlow(gameID entities.GameID) error {
	players, err := e.playerRepo.GetPlayersByGame(gameID)
	if err != nil {
		return err
	}

	// Initialize night state
	e.nightService.StartNight(gameID, players)

	// Start with the first night phase
	e.startNextNightPhase(gameID, players)

	return nil
}

// startNextNightPhase starts the appropriate night phase
func (e *GameEngine) startNextNightPhase(gameID entities.GameID, players []*entities.Player) {
	state, exists := e.nightService.GetNightState(gameID)
	if !exists {
		e.TransitionToDay(gameID)
		return
	}

	currentPhase := state.CurrentPhase

	// Skip phases if no players with that role
	for e.nightService.ShouldSkipPhase(gameID, currentPhase) {
		newPhase := e.nightService.AdvancePhase(gameID)
		currentPhase = newPhase
		if currentPhase == NightPhaseEnd {
			e.TransitionToDay(gameID)
			return
		}
	}

	// Send turn event to appropriate players
	e.nightService.SendTurnEvent(gameID, currentPhase, players)

	// Start the appropriate timer
	switch currentPhase {
	case NightPhaseSeer:
		e.timerService.StartRoleTimer(gameID, entities.RoleSeer)
	case NightPhaseWerewolf:
		// Start werewolf vote
		werewolves := GetPlayersWithRole(players, entities.RoleWerewolf)
		victims := GetNonWerewolfPlayers(players)
		e.voteService.StartWerewolfVote(gameID, werewolves, victims)
		e.timerService.StartRoleTimer(gameID, entities.RoleWerewolf)
	case NightPhaseWitch:
		e.timerService.StartRoleTimer(gameID, entities.RoleWitch)
	}
}

// advanceNightPhase moves to the next night phase
func (e *GameEngine) advanceNightPhase(gameID entities.GameID, state *NightState) {
	newPhase := e.nightService.AdvancePhase(gameID)

	if NightPhase(newPhase) == NightPhaseEnd || e.nightService.IsNightComplete(gameID) {
		e.TransitionToDay(gameID)
		return
	}

	players, err := e.playerRepo.GetPlayersByGame(gameID)
	if err != nil {
		log.Printf("Error getting players for game %s: %v", gameID, err)
		return
	}

	e.startNextNightPhase(gameID, players)
}

// TransitionToDay transitions the game to day phase
func (e *GameEngine) TransitionToDay(gameID entities.GameID) error {
	game, err := e.gameRepo.GetGame(gameID)
	if err != nil {
		return err
	}

	players, err := e.playerRepo.GetPlayersByGame(gameID)
	if err != nil {
		return err
	}

	// Process night deaths and broadcast individual death events
	deaths := e.nightService.GetPendingDeaths(gameID)
	for _, deathID := range deaths {
		for _, p := range players {
			if p.ID == deathID {
				p.Kill()
				if err := e.playerRepo.SavePlayer(p); err != nil {
					log.Printf("Error saving dead player %s: %v", p.ID, err)
				}

				// Broadcast individual death event with role reveal
				var role entities.RoleType
				if p.Role != nil {
					role = p.Role.GetType()
				}
				deathEvent := events.NewDeathEvent(p.ID, role)
				payload, err := json.Marshal(deathEvent)
				if err != nil {
					log.Printf("Error marshaling death event for player %s: %v", p.ID, err)
				} else {
					e.broadcaster.BroadcastToGame(gameID, payload)
				}
				break
			}
		}
	}

	// Clear night state
	e.nightService.ClearNight(gameID)

	// Check win condition after deaths
	// Refresh players to get updated alive status
	players, err = e.playerRepo.GetPlayersByGame(gameID)
	if err != nil {
		log.Printf("Error refreshing players after deaths for game %s: %v", gameID, err)
		// Continue with stale player data rather than failing
	}
	winResult := e.CheckWinCondition(players)
	if winResult.GameEnded {
		return e.EndGame(gameID, game, winResult)
	}

	// Update game state
	game.Phase = entities.PhaseDay
	if err := e.gameRepo.SaveGame(game); err != nil {
		return err
	}

	// Broadcast day event (signals day start)
	dayEvent := events.NewDayEvent(game.Day)
	payload, err := json.Marshal(dayEvent)
	if err != nil {
		log.Printf("Error marshaling day event for game %s: %v", gameID, err)
	} else {
		e.broadcaster.BroadcastToGame(gameID, payload)
	}

	// Start day timer
	e.timerService.StartPhaseTimer(gameID, entities.PhaseDay)

	log.Printf("Game %s transitioned to day %d with %d deaths", gameID, game.Day, len(deaths))
	return nil
}

// TransitionToVote transitions the game to vote phase
func (e *GameEngine) TransitionToVote(gameID entities.GameID) error {
	game, err := e.gameRepo.GetGame(gameID)
	if err != nil {
		return err
	}

	players, err := e.playerRepo.GetPlayersByGame(gameID)
	if err != nil {
		return err
	}

	// Get alive players for voting
	var alivePlayers []*entities.Player
	for _, p := range players {
		if p.IsAlive {
			alivePlayers = append(alivePlayers, p)
		}
	}

	// Start village vote
	_, err = e.voteService.StartVillageVote(gameID, alivePlayers)
	if err != nil {
		return err
	}

	// Update game state
	game.Phase = entities.PhaseVote
	if err := e.gameRepo.SaveGame(game); err != nil {
		return err
	}

	// Start vote timer
	e.timerService.StartPhaseTimer(gameID, entities.PhaseVote)

	log.Printf("Game %s transitioned to vote phase", gameID)
	return nil
}

// TransitionToNight transitions the game to night phase
func (e *GameEngine) TransitionToNight(gameID entities.GameID) error {
	game, err := e.gameRepo.GetGame(gameID)
	if err != nil {
		return err
	}

	players, err := e.playerRepo.GetPlayersByGame(gameID)
	if err != nil {
		return err
	}

	// Update game state
	game.Phase = entities.PhaseNight
	game.Day++
	if err := e.gameRepo.SaveGame(game); err != nil {
		return err
	}

	// Broadcast night event
	nightEvent := events.NewNightEvent()
	payload, _ := json.Marshal(nightEvent)
	e.broadcaster.BroadcastToGame(gameID, payload)

	// Initialize night state and start first phase
	e.nightService.StartNight(gameID, players)
	e.startNextNightPhase(gameID, players)

	log.Printf("Game %s transitioned to night %d", gameID, game.Day)
	return nil
}

// ProcessVoteResult processes the village vote result
func (e *GameEngine) ProcessVoteResult(gameID entities.GameID) error {
	result, err := e.voteService.ResolveVote(gameID)
	if err != nil {
		e.voteService.ClearVote(gameID) // Clean up even on error
		return err
	}
	defer e.voteService.ClearVote(gameID) // Ensure cleanup after processing

	players, err := e.playerRepo.GetPlayersByGame(gameID)
	if err != nil {
		return err
	}

	// If no tie, eliminate the player
	if result.Target != nil {
		for _, p := range players {
			if p.ID == *result.Target {
				p.Kill()
				if err := e.playerRepo.SavePlayer(p); err != nil {
					log.Printf("Error saving eliminated player %s: %v", p.ID, err)
				}

				// Broadcast death event with role reveal
				var role entities.RoleType
				if p.Role != nil {
					role = p.Role.GetType()
				}
				deathEvent := events.NewDeathEvent(p.ID, role)
				payload, err := json.Marshal(deathEvent)
				if err != nil {
					log.Printf("Error marshaling death event for player %s: %v", p.ID, err)
				} else {
					e.broadcaster.BroadcastToGame(gameID, payload)
				}
				break
			}
		}
	}

	// Refresh players after kill
	players, err = e.playerRepo.GetPlayersByGame(gameID)
	if err != nil {
		log.Printf("Error refreshing players after vote for game %s: %v", gameID, err)
	}

	// Check win condition
	winResult := e.CheckWinCondition(players)
	if winResult.GameEnded {
		game, err := e.gameRepo.GetGame(gameID)
		if err != nil {
			log.Printf("Error getting game %s for end game: %v", gameID, err)
			return err
		}
		return e.EndGame(gameID, game, winResult)
	}

	// Transition to night
	return e.TransitionToNight(gameID)
}

// CheckWinCondition checks if the game has ended
func (e *GameEngine) CheckWinCondition(players []*entities.Player) *WinResult {
	werewolvesAlive := 0
	villagersAlive := 0
	var allWerewolves []entities.PlayerID
	var allVillagers []entities.PlayerID

	for _, p := range players {
		if p.Role == nil {
			continue
		}
		clan := entities.GetClan(p.Role.GetType())

		switch clan {
		case entities.ClanWerewolf:
			allWerewolves = append(allWerewolves, p.ID)
			if p.IsAlive {
				werewolvesAlive++
			}
		case entities.ClanVillager:
			allVillagers = append(allVillagers, p.ID)
			if p.IsAlive {
				villagersAlive++
			}
		}
	}

	// Werewolves win if they equal or outnumber villagers
	if werewolvesAlive >= villagersAlive && werewolvesAlive > 0 {
		return &WinResult{
			GameEnded:   true,
			WinningClan: entities.ClanWerewolf,
			Winners:     allWerewolves,
		}
	}

	// Villagers win if all werewolves are dead
	if werewolvesAlive == 0 {
		return &WinResult{
			GameEnded:   true,
			WinningClan: entities.ClanVillager,
			Winners:     allVillagers,
		}
	}

	return &WinResult{GameEnded: false}
}

// EndGame ends the game and broadcasts the result
func (e *GameEngine) EndGame(gameID entities.GameID, game *entities.Game, result *WinResult) error {
	// Cancel any running timers
	e.timerService.CancelTimer(gameID)

	// Clean up night state to prevent memory leak
	e.nightService.ClearNight(gameID)

	// Clean up any pending vote to prevent memory leak
	e.voteService.ClearVote(gameID)

	// Update game state
	game.Status = entities.GameStatusEnded
	if err := e.gameRepo.SaveGame(game); err != nil {
		return err
	}

	// Broadcast win event
	winEvent := events.NewWinEvent(result.WinningClan, result.Winners)
	payload, _ := json.Marshal(winEvent)
	e.broadcaster.BroadcastToGame(gameID, payload)

	log.Printf("Game %s ended. Winner: %s", gameID, result.WinningClan)
	return nil
}

// HandleSeerAction processes the seer's night action
func (e *GameEngine) HandleSeerAction(gameID entities.GameID, seerID, targetID entities.PlayerID) error {
	// 1. Validate game phase
	game, err := e.gameRepo.GetGame(gameID)
	if err != nil {
		return err
	}
	if game.Status != entities.GameStatusActive {
		return ErrGameNotActive
	}
	if game.Phase != entities.PhaseNight {
		return ErrWrongPhase
	}

	// 2. Validate it's seer's turn
	if !e.nightService.CanSeerAct(gameID) {
		return ErrNotYourTurn
	}

	// 3. Validate seer player
	seer, err := e.playerRepo.GetPlayer(seerID)
	if err != nil {
		return err
	}
	if !seer.IsAlive {
		return ErrPlayerDead
	}
	if seer.Role == nil || seer.Role.GetType() != entities.RoleSeer {
		return ErrWrongRole
	}

	// 4. Validate target
	if targetID == seerID {
		return ErrCannotTargetSelf
	}
	target, err := e.playerRepo.GetPlayer(targetID)
	if err != nil {
		return ErrInvalidTarget
	}
	if !target.IsAlive {
		return ErrTargetDead
	}

	// 5. Get target role
	var targetRole entities.RoleType
	if target.Role != nil {
		targetRole = target.Role.GetType()
	}

	// 6. Record the action
	e.nightService.RecordSeerAction(gameID, targetID, targetRole)

	// 7. Send the result to the seer
	revealEvent := events.NewSeerRevealEvent(targetID, targetRole)
	payload, _ := json.Marshal(revealEvent)
	e.playerSender.SendToPlayer(seerID, payload)

	// 8. Skip timer and advance to next phase
	e.timerService.SkipTimer(gameID)

	return nil
}

// HandleWerewolfVote processes a werewolf's vote during night
func (e *GameEngine) HandleWerewolfVote(gameID entities.GameID, werewolfID entities.PlayerID, targetID *entities.PlayerID) error {
	// 1. Validate game phase
	game, err := e.gameRepo.GetGame(gameID)
	if err != nil {
		return err
	}
	if game.Status != entities.GameStatusActive {
		return ErrGameNotActive
	}
	if game.Phase != entities.PhaseNight {
		return ErrWrongPhase
	}

	// 2. Validate it's werewolves' turn
	if !e.nightService.CanWerewolvesVote(gameID) {
		return ErrNotYourTurn
	}

	// 3. Validate werewolf player
	werewolf, err := e.playerRepo.GetPlayer(werewolfID)
	if err != nil {
		return err
	}
	if !werewolf.IsAlive {
		return ErrPlayerDead
	}
	if werewolf.Role == nil || werewolf.Role.GetType() != entities.RoleWerewolf {
		return ErrWrongRole
	}

	// 4. Validate target (if not abstaining)
	if targetID != nil {
		target, err := e.playerRepo.GetPlayer(*targetID)
		if err != nil {
			return ErrInvalidTarget
		}
		if !target.IsAlive {
			return ErrTargetDead
		}
		// Werewolves can't target other werewolves
		if target.Role != nil && target.Role.GetType() == entities.RoleWerewolf {
			return ErrInvalidTarget
		}
	}

	// 5. Cast the vote
	if err := e.voteService.CastVote(gameID, werewolfID, targetID); err != nil {
		return err
	}

	// 6. Check if all werewolves have voted
	if e.voteService.HasEveryoneVoted(gameID) {
		// Resolve the vote
		result, err := e.voteService.ResolveVote(gameID)
		if err != nil {
			return err
		}

		// Record the victim
		e.nightService.RecordWerewolfVictim(gameID, result.Target)
		e.voteService.ClearVote(gameID)

		// Skip timer and advance to next phase
		e.timerService.SkipTimer(gameID)
	}

	return nil
}

// HandleWitchAction processes the witch's night action
func (e *GameEngine) HandleWitchAction(gameID entities.GameID, witchID entities.PlayerID, healTargetID, poisonTargetID *entities.PlayerID) error {
	// 1. Validate game phase
	game, err := e.gameRepo.GetGame(gameID)
	if err != nil {
		return err
	}
	if game.Status != entities.GameStatusActive {
		return ErrGameNotActive
	}
	if game.Phase != entities.PhaseNight {
		return ErrWrongPhase
	}

	// 2. Validate it's witch's turn
	if !e.nightService.CanWitchAct(gameID) {
		return ErrNotYourTurn
	}

	// 3. Validate witch player
	witch, err := e.playerRepo.GetPlayer(witchID)
	if err != nil {
		return err
	}
	if !witch.IsAlive {
		return ErrPlayerDead
	}
	if witch.Role == nil || witch.Role.GetType() != entities.RoleWitch {
		return ErrWrongRole
	}

	// 4. Get witch abilities
	canHeal, canPoison := e.nightService.GetWitchAbilities(gameID)

	// 5. Validate heal action
	if healTargetID != nil {
		if !canHeal {
			return ErrAbilityUsed
		}
		// Can only heal the werewolf victim
		victim := e.nightService.GetWerewolfVictim(gameID)
		if victim == nil || *healTargetID != *victim {
			return ErrCanOnlyHealVictim
		}
	}

	// 6. Validate poison action
	if poisonTargetID != nil {
		if !canPoison {
			return ErrAbilityUsed
		}
		if *poisonTargetID == witchID {
			return ErrCannotTargetSelf
		}
		target, err := e.playerRepo.GetPlayer(*poisonTargetID)
		if err != nil {
			return ErrInvalidTarget
		}
		if !target.IsAlive {
			return ErrTargetDead
		}
	}

	// 7. Consume abilities if used
	abilities := witch.Role.GetAbilities()
	if abilities != nil {
		for _, ability := range *abilities {
			if healTargetID != nil && ability.GetName() == "Heal" {
				ability.Consume()
			}
			if poisonTargetID != nil && ability.GetName() == "Poison" {
				ability.Consume()
			}
		}
		// Save witch with updated abilities
		e.playerRepo.SavePlayer(witch)
	}

	// 8. Record the action
	e.nightService.RecordWitchAction(gameID, healTargetID, poisonTargetID)

	// 9. Skip timer and advance to next phase
	e.timerService.SkipTimer(gameID)

	return nil
}

// HandleVillageVote processes a village vote during day
func (e *GameEngine) HandleVillageVote(gameID entities.GameID, voterID entities.PlayerID, targetID *entities.PlayerID) error {
	// 1. Validate game phase
	game, err := e.gameRepo.GetGame(gameID)
	if err != nil {
		return err
	}
	if game.Status != entities.GameStatusActive {
		return ErrGameNotActive
	}
	if game.Phase != entities.PhaseVote {
		return ErrWrongPhase
	}

	// 2. Validate voter is alive
	voter, err := e.playerRepo.GetPlayer(voterID)
	if err != nil {
		return err
	}
	if !voter.IsAlive {
		return ErrPlayerDead
	}

	// 3. Validate target (if not abstaining)
	if targetID != nil {
		target, err := e.playerRepo.GetPlayer(*targetID)
		if err != nil {
			return ErrInvalidTarget
		}
		if !target.IsAlive {
			return ErrTargetDead
		}
	}

	// 4. Cast the vote
	if err := e.voteService.CastVote(gameID, voterID, targetID); err != nil {
		return err
	}

	// 5. Check if all players have voted
	if e.voteService.HasEveryoneVoted(gameID) {
		// Skip timer and process the result
		e.timerService.SkipTimer(gameID)
	}

	return nil
}

package orchestration

import (
	"context"
	"encoding/json"
	"shamus-backend/internal/application/services"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/prompts"
	apperrors "shamus-backend/internal/domain/errors"
	"shamus-backend/internal/domain/ports"
	"shamus-backend/pkg/logger"
	"time"
)

// Timer durations
const (
	DayPhaseDurationV2      = 120 * time.Second
	VotePhaseDurationV2     = 60 * time.Second
	SeerTimerDurationV2     = 30 * time.Second
	WerewolfTimerDurationV2 = 60 * time.Second
	WitchTimerDurationV2    = 45 * time.Second
	MayorTiebreakerDuration = 30 * time.Second
)

// WinResult represents the outcome of a win condition check
type WinResult struct {
	GameEnded   bool
	WinningClan entities.Clan
	Winners     []entities.PlayerID
}

// GameEngineV2 orchestrates game flow using the new Prompt/Notification architecture
type GameEngineV2 struct {
	gameRepo      ports.GameRepository
	playerRepo    ports.PlayerRepository
	promptService *services.PromptService
	notifier      *services.NotificationService
	nightService  *services.NightService

	// Active group votes
	activeWerewolfVote *entities.GroupID
	activeVillageVote  *entities.GroupID
}

// NewGameEngineV2 creates a new GameEngineV2
func NewGameEngineV2(
	gameRepo ports.GameRepository,
	playerRepo ports.PlayerRepository,
	promptService *services.PromptService,
	notifier *services.NotificationService,
	nightService *services.NightService,
) *GameEngineV2 {
	engine := &GameEngineV2{
		gameRepo:      gameRepo,
		playerRepo:    playerRepo,
		promptService: promptService,
		notifier:      notifier,
		nightService:  nightService,
	}

	// Register prompt callbacks
	promptService.RegisterCallback("seer_vision", engine.handleSeerVisionCallback)
	promptService.RegisterCallback("witch_potion", engine.handleWitchPotionCallback)
	promptService.RegisterCallback("witch_poison_target", engine.handleWitchPoisonTargetCallback)
	promptService.RegisterCallback("werewolf_vote", engine.handleWerewolfVoteCallback)
	promptService.RegisterCallback("village_vote", engine.handleVillageVoteCallback)
	promptService.RegisterCallback("mayor_tiebreaker", engine.handleMayorTiebreakerCallback)

	return engine
}

// StartGameFlow starts the game flow after StartGame (begins first night)
func (e *GameEngineV2) StartGameFlow(gameID entities.GameID) error {
	players, err := e.playerRepo.GetPlayersByGame(context.TODO(), gameID)
	if err != nil {
		return err
	}

	// Notify phase change to night
	e.notifier.NotifyPhaseChanged(gameID, entities.PhaseNight, 1, "")

	// Initialize night state
	e.nightService.StartNight(gameID, players)

	// Start with the first night phase
	e.startNextNightPhase(gameID, players)

	return nil
}

// startNextNightPhase starts the appropriate night phase
func (e *GameEngineV2) startNextNightPhase(gameID entities.GameID, players []*entities.Player) {
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
		if currentPhase == entities.NightPhaseEnd {
			e.TransitionToDay(gameID)
			return
		}
	}

	// Notify sub-phase change
	e.notifier.NotifyPhaseChanged(gameID, entities.PhaseNight, 0, string(currentPhase))

	// Start the appropriate prompt
	switch currentPhase {
	case entities.NightPhaseSeer:
		e.startSeerPhase(gameID, players)
	case entities.NightPhaseWerewolf:
		e.startWerewolfVotePhase(gameID, players)
	case entities.NightPhaseWitch:
		e.startWitchPhase(gameID, players)
	}
}

// startSeerPhase creates a prompt for the seer
func (e *GameEngineV2) startSeerPhase(gameID entities.GameID, players []*entities.Player) {
	// Find the seer
	var seer *entities.Player
	for _, p := range players {
		if p.Role != nil && p.Role.GetType() == entities.RoleSeer && p.IsAlive {
			seer = p
			break
		}
	}

	if seer == nil {
		logger.Get().Warn().Str("gameID", string(gameID)).Msg("No alive seer found")
		e.advanceFromCurrentNightPhase(gameID)
		return
	}

	// Get eligible targets (alive players except seer)
	var eligibleTargets []entities.PlayerID
	playersInfo := make(map[entities.PlayerID]prompts.PlayerInfo)
	for _, p := range players {
		if p.IsAlive && p.ID != seer.ID {
			eligibleTargets = append(eligibleTargets, p.ID)
		}
		playersInfo[p.ID] = prompts.PlayerInfo{
			Username: p.Username,
			IsAlive:  p.IsAlive,
		}
	}

	payload := prompts.SelectPlayerPayload{
		Message:         "Qui voulez-vous observer cette nuit ?",
		EligiblePlayers: eligibleTargets,
		PlayersInfo:     playersInfo,
	}

	_, err := e.promptService.CreatePrompt(
		gameID,
		seer.ID,
		entities.PromptSelectPlayer,
		"seer_vision",
		payload,
		SeerTimerDurationV2,
		false, // canSkip
	)

	if err != nil {
		logger.Get().Error().Err(err).Msg("Failed to create seer prompt")
		e.advanceFromCurrentNightPhase(gameID)
	}
}

// startWerewolfVotePhase creates group vote prompts for werewolves
func (e *GameEngineV2) startWerewolfVotePhase(gameID entities.GameID, players []*entities.Player) {
	// Get werewolves
	var werewolves []*entities.Player
	var eligibleTargets []entities.PlayerID

	for _, p := range players {
		if !p.IsAlive {
			continue
		}
		if p.Role != nil && p.Role.GetType() == entities.RoleWerewolf {
			werewolves = append(werewolves, p)
		} else {
			eligibleTargets = append(eligibleTargets, p.ID)
		}
	}

	if len(werewolves) == 0 {
		logger.Get().Warn().Str("gameID", string(gameID)).Msg("No alive werewolves")
		e.advanceFromCurrentNightPhase(gameID)
		return
	}

	// Create group vote
	groupID, err := e.promptService.CreateGroupVote(
		gameID,
		"werewolf_vote",
		werewolves,
		eligibleTargets,
		WerewolfTimerDurationV2,
		true, // canAbstain
		nil,  // no mayor for werewolf vote
	)

	if err != nil {
		logger.Get().Error().Err(err).Msg("Failed to create werewolf vote")
		e.advanceFromCurrentNightPhase(gameID)
		return
	}

	e.activeWerewolfVote = groupID
}

// startWitchPhase creates prompts for the witch
func (e *GameEngineV2) startWitchPhase(gameID entities.GameID, players []*entities.Player) {
	// Find the witch
	var witch *entities.Player
	for _, p := range players {
		if p.Role != nil && p.Role.GetType() == entities.RoleWitch && p.IsAlive {
			witch = p
			break
		}
	}

	if witch == nil {
		logger.Get().Warn().Str("gameID", string(gameID)).Msg("No alive witch found")
		e.advanceFromCurrentNightPhase(gameID)
		return
	}

	// Get witch abilities
	canHeal, canPoison := e.nightService.GetWitchAbilities(gameID)

	// Get werewolf victim
	victim := e.nightService.GetWerewolfVictim(gameID)
	var victimUsername string
	if victim != nil {
		for _, p := range players {
			if p.ID == *victim {
				victimUsername = p.Username
				break
			}
		}
	}

	// Create witch potion payload
	payload := prompts.NewWitchPotionPayload(victim, victimUsername, canHeal, canPoison)

	_, err := e.promptService.CreatePrompt(
		gameID,
		witch.ID,
		entities.PromptSelectOption,
		"witch_potion",
		payload,
		WitchTimerDurationV2,
		true, // canSkip - witch can pass
	)

	if err != nil {
		logger.Get().Error().Err(err).Msg("Failed to create witch prompt")
		e.advanceFromCurrentNightPhase(gameID)
	}
}

// TransitionToDay transitions the game to day phase
func (e *GameEngineV2) TransitionToDay(gameID entities.GameID) error {
	game, err := e.gameRepo.GetGame(context.TODO(), gameID)
	if err != nil {
		return err
	}

	players, err := e.playerRepo.GetPlayersByGame(context.TODO(), gameID)
	if err != nil {
		return err
	}

	// Process night deaths
	deaths := e.nightService.GetPendingDeaths(gameID)
	playersToSave := make([]*entities.Player, 0, len(deaths))

	for _, deathID := range deaths {
		for _, p := range players {
			if p.ID == deathID {
				p.Kill()
				playersToSave = append(playersToSave, p)

				// Notify death
				var role entities.RoleType
				if p.Role != nil {
					role = p.Role.GetType()
				}
				e.notifier.NotifyPlayerDied(gameID, p.ID, p.Username, role, "werewolf")
				break
			}
		}
	}

	// Batch save dead players
	if len(playersToSave) > 0 {
		e.playerRepo.SavePlayers(context.TODO(), playersToSave)
	}

	// Clear night state
	e.nightService.ClearNight(gameID)

	// Refresh players and check win condition
	players, _ = e.playerRepo.GetPlayersByGame(context.TODO(), gameID)
	winResult := e.CheckWinCondition(players)
	if winResult.GameEnded {
		return e.EndGame(gameID, game, winResult)
	}

	// Update game state
	game.Phase = entities.PhaseDay
	e.gameRepo.SaveGame(context.TODO(), game)

	// Notify phase change
	e.notifier.NotifyPhaseChanged(gameID, entities.PhaseDay, game.Day, "")

	// Start day timer - after timeout, transition to vote
	go func() {
		time.Sleep(DayPhaseDurationV2)
		e.TransitionToVote(gameID)
	}()

	logger.Get().Info().
		Str("gameID", string(gameID)).
		Int("day", game.Day).
		Int("deaths", len(deaths)).
		Msg("Transitioned to day")

	return nil
}

// TransitionToVote transitions the game to vote phase
func (e *GameEngineV2) TransitionToVote(gameID entities.GameID) error {
	game, err := e.gameRepo.GetGame(context.TODO(), gameID)
	if err != nil {
		return err
	}

	// Verify we're in day phase
	if game.Phase != entities.PhaseDay {
		return nil // Already transitioned
	}

	players, err := e.playerRepo.GetPlayersByGame(context.TODO(), gameID)
	if err != nil {
		return err
	}

	// Get alive players
	var alivePlayers []*entities.Player
	var eligibleTargets []entities.PlayerID
	var mayorID *entities.PlayerID

	for _, p := range players {
		if p.IsAlive {
			alivePlayers = append(alivePlayers, p)
			eligibleTargets = append(eligibleTargets, p.ID)
			// TODO: Check if player is mayor
		}
	}

	// Update game state
	game.Phase = entities.PhaseVote
	e.gameRepo.SaveGame(context.TODO(), game)

	// Notify phase change
	e.notifier.NotifyPhaseChanged(gameID, entities.PhaseVote, game.Day, "")

	// Create group vote for village
	groupID, err := e.promptService.CreateGroupVote(
		gameID,
		"village_vote",
		alivePlayers,
		eligibleTargets,
		VotePhaseDurationV2,
		true,    // canAbstain
		mayorID, // mayor for tiebreaker
	)

	if err != nil {
		logger.Get().Error().Err(err).Msg("Failed to create village vote")
		return err
	}

	e.activeVillageVote = groupID

	logger.Get().Info().
		Str("gameID", string(gameID)).
		Int("voters", len(alivePlayers)).
		Msg("Transitioned to vote")

	return nil
}

// TransitionToNight transitions the game to night phase
func (e *GameEngineV2) TransitionToNight(gameID entities.GameID) error {
	game, err := e.gameRepo.GetGame(context.TODO(), gameID)
	if err != nil {
		return err
	}

	players, err := e.playerRepo.GetPlayersByGame(context.TODO(), gameID)
	if err != nil {
		return err
	}

	// Update game state
	game.Phase = entities.PhaseNight
	game.Day++
	e.gameRepo.SaveGame(context.TODO(), game)

	// Notify phase change
	e.notifier.NotifyPhaseChanged(gameID, entities.PhaseNight, game.Day, "")

	// Initialize night state and start first phase
	e.nightService.StartNight(gameID, players)
	e.startNextNightPhase(gameID, players)

	logger.Get().Info().
		Str("gameID", string(gameID)).
		Int("day", game.Day).
		Msg("Transitioned to night")

	return nil
}

// CheckWinCondition checks if the game has ended
func (e *GameEngineV2) CheckWinCondition(players []*entities.Player) *WinResult {
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

	// Draw if everyone is dead
	if werewolvesAlive == 0 && villagersAlive == 0 {
		return &WinResult{
			GameEnded:   true,
			WinningClan: entities.ClanNone,
			Winners:     []entities.PlayerID{},
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

// EndGame ends the game
func (e *GameEngineV2) EndGame(gameID entities.GameID, game *entities.Game, result *WinResult) error {
	// Clean up
	e.nightService.ClearNight(gameID)
	e.promptService.CleanupGame(gameID)

	// Update game state
	game.Status = entities.GameStatusEnded
	e.gameRepo.SaveGame(context.TODO(), game)

	// Notify
	e.notifier.NotifyGameEnded(gameID, result.WinningClan, result.Winners)

	logger.Get().Info().
		Str("gameID", string(gameID)).
		Str("winner", string(result.WinningClan)).
		Msg("Game ended")

	return nil
}

// --- Callback handlers ---

func (e *GameEngineV2) handleSeerVisionCallback(prompt *entities.Prompt, response json.RawMessage) error {
	gameID := prompt.GameID

	// If timeout (nil response), skip
	if response == nil {
		logger.Get().Info().Str("gameID", string(gameID)).Msg("Seer timed out")
		e.nightService.RecordSeerAction(gameID, "", entities.RoleType(""))
		e.advanceFromCurrentNightPhase(gameID)
		return nil
	}

	// Parse response
	resp, err := prompts.ParseSelectPlayerResponse(response)
	if err != nil {
		return err
	}

	if resp.Skipped || resp.PlayerID == nil {
		e.nightService.RecordSeerAction(gameID, "", entities.RoleType(""))
		e.advanceFromCurrentNightPhase(gameID)
		return nil
	}

	// Get target's role
	target, err := e.playerRepo.GetPlayer(context.TODO(), *resp.PlayerID)
	if err != nil {
		return err
	}

	var targetRole entities.RoleType
	if target.Role != nil {
		targetRole = target.Role.GetType()
	}

	// Record action
	e.nightService.RecordSeerAction(gameID, *resp.PlayerID, targetRole)

	// Send result to seer
	e.notifier.NotifySeerResult(prompt.PlayerID, *resp.PlayerID, target.Username, targetRole)

	// Advance to next phase
	e.advanceFromCurrentNightPhase(gameID)

	return nil
}

func (e *GameEngineV2) handleWerewolfVoteCallback(prompt *entities.Prompt, response json.RawMessage) error {
	if prompt.GroupID == nil {
		return nil
	}

	gameID := prompt.GameID
	groupID := *prompt.GroupID

	// Parse group result
	var result services.GroupVoteResult
	if err := json.Unmarshal(response, &result); err != nil {
		return err
	}

	// Record victim
	if result.IsTie || result.Target == nil {
		// Tie or no votes = no victim
		e.nightService.RecordWerewolfVictim(gameID, nil)
		logger.Get().Info().Str("gameID", string(gameID)).Msg("Werewolf vote: no victim (tie or no votes)")
	} else {
		e.nightService.RecordWerewolfVictim(gameID, result.Target)
		logger.Get().Info().
			Str("gameID", string(gameID)).
			Str("victim", string(*result.Target)).
			Msg("Werewolf vote: victim selected")
	}

	// Notify werewolves of result
	state, _ := e.promptService.GetGroupState(groupID)
	if state != nil {
		e.notifier.NotifyVoteResult(
			gameID,
			"werewolf",
			result.Target,
			result.IsTie,
			result.TiedTargets,
			result.VoteCounts,
			false,
		)
	}

	e.activeWerewolfVote = nil

	// Advance to next phase
	e.advanceFromCurrentNightPhase(gameID)

	return nil
}

func (e *GameEngineV2) handleWitchPotionCallback(prompt *entities.Prompt, response json.RawMessage) error {
	gameID := prompt.GameID

	// If timeout, witch does nothing
	if response == nil {
		logger.Get().Info().Str("gameID", string(gameID)).Msg("Witch timed out")
		e.nightService.RecordWitchAction(gameID, nil, nil)
		e.advanceFromCurrentNightPhase(gameID)
		return nil
	}

	// Parse response
	resp, err := prompts.ParseSelectOptionResponse(response)
	if err != nil {
		return err
	}

	if resp.Skipped || resp.OptionID == "skip" {
		e.nightService.RecordWitchAction(gameID, nil, nil)
		e.advanceFromCurrentNightPhase(gameID)
		return nil
	}

	switch resp.OptionID {
	case "heal":
		// Heal the victim
		victim := e.nightService.GetWerewolfVictim(gameID)
		e.nightService.RecordWitchAction(gameID, victim, nil)
		e.advanceFromCurrentNightPhase(gameID)

	case "poison":
		// Need to ask for target
		players, _ := e.playerRepo.GetPlayersByGame(context.TODO(), gameID)
		e.startWitchPoisonTargetSelection(gameID, prompt.PlayerID, players)
	}

	return nil
}

func (e *GameEngineV2) startWitchPoisonTargetSelection(gameID entities.GameID, witchID entities.PlayerID, players []*entities.Player) {
	// Get eligible targets (alive players except witch)
	var eligibleTargets []entities.PlayerID
	playersInfo := make(map[entities.PlayerID]prompts.PlayerInfo)

	victim := e.nightService.GetWerewolfVictim(gameID)

	for _, p := range players {
		if p.IsAlive && p.ID != witchID {
			// Can't poison someone already dying
			if victim != nil && p.ID == *victim {
				continue
			}
			eligibleTargets = append(eligibleTargets, p.ID)
		}
		playersInfo[p.ID] = prompts.PlayerInfo{
			Username: p.Username,
			IsAlive:  p.IsAlive,
		}
	}

	payload := prompts.SelectPlayerPayload{
		Message:         "Qui voulez-vous empoisonner ?",
		EligiblePlayers: eligibleTargets,
		PlayersInfo:     playersInfo,
	}

	_, err := e.promptService.CreatePrompt(
		gameID,
		witchID,
		entities.PromptSelectPlayer,
		"witch_poison_target",
		payload,
		30*time.Second,
		true, // can skip
	)

	if err != nil {
		logger.Get().Error().Err(err).Msg("Failed to create witch poison target prompt")
		e.nightService.RecordWitchAction(gameID, nil, nil)
		e.advanceFromCurrentNightPhase(gameID)
	}
}

func (e *GameEngineV2) handleWitchPoisonTargetCallback(prompt *entities.Prompt, response json.RawMessage) error {
	gameID := prompt.GameID

	if response == nil {
		e.nightService.RecordWitchAction(gameID, nil, nil)
		e.advanceFromCurrentNightPhase(gameID)
		return nil
	}

	resp, err := prompts.ParseSelectPlayerResponse(response)
	if err != nil {
		return err
	}

	if resp.Skipped || resp.PlayerID == nil {
		e.nightService.RecordWitchAction(gameID, nil, nil)
	} else {
		e.nightService.RecordWitchAction(gameID, nil, resp.PlayerID)
	}

	e.advanceFromCurrentNightPhase(gameID)
	return nil
}

func (e *GameEngineV2) handleVillageVoteCallback(prompt *entities.Prompt, response json.RawMessage) error {
	if prompt.GroupID == nil {
		return nil
	}

	gameID := prompt.GameID
	groupID := *prompt.GroupID

	// Parse group result
	var result services.GroupVoteResult
	if err := json.Unmarshal(response, &result); err != nil {
		return err
	}

	// Check for tie
	if result.IsTie {
		// Check if there's a mayor
		state, _ := e.promptService.GetGroupState(groupID)
		if state != nil && state.HasMayor && state.MayorID != nil {
			// Request mayor tiebreaker
			e.promptService.RequestMayorTiebreaker(groupID, result.TiedTargets, result.VoteCounts)

			// Create mayor tiebreaker prompt
			e.createMayorTiebreakerPrompt(gameID, *state.MayorID, groupID, result.TiedTargets)
			return nil
		}

		// No mayor - no one dies
		e.notifier.NotifyVoteResult(gameID, "village", nil, true, result.TiedTargets, result.VoteCounts, false)
		e.activeVillageVote = nil
		e.TransitionToNight(gameID)
		return nil
	}

	// Clear winner - eliminate player
	if result.Target != nil {
		e.eliminatePlayer(gameID, *result.Target, "village_vote")
	}

	e.notifier.NotifyVoteResult(gameID, "village", result.Target, false, nil, result.VoteCounts, false)

	e.activeVillageVote = nil

	// Check win condition and transition
	players, _ := e.playerRepo.GetPlayersByGame(context.TODO(), gameID)
	winResult := e.CheckWinCondition(players)
	if winResult.GameEnded {
		game, _ := e.gameRepo.GetGame(context.TODO(), gameID)
		return e.EndGame(gameID, game, winResult)
	}

	e.TransitionToNight(gameID)
	return nil
}

func (e *GameEngineV2) createMayorTiebreakerPrompt(gameID entities.GameID, mayorID entities.PlayerID, groupID entities.GroupID, tiedPlayers []entities.PlayerID) {
	players, _ := e.playerRepo.GetPlayersByGame(context.TODO(), gameID)
	playersInfo := make(map[entities.PlayerID]prompts.PlayerInfo)
	for _, p := range players {
		playersInfo[p.ID] = prompts.PlayerInfo{
			Username: p.Username,
			IsAlive:  p.IsAlive,
		}
	}

	payload := prompts.SelectPlayerPayload{
		Message:         "En tant que Maire, vous devez départager l'égalité. Qui doit être éliminé ?",
		EligiblePlayers: tiedPlayers,
		PlayersInfo:     playersInfo,
	}

	_, err := e.promptService.CreatePrompt(
		gameID,
		mayorID,
		entities.PromptSelectPlayer,
		"mayor_tiebreaker",
		payload,
		MayorTiebreakerDuration,
		false, // can't skip
	)

	if err != nil {
		logger.Get().Error().Err(err).Msg("Failed to create mayor tiebreaker prompt")
		// If we can't create prompt, no one dies
		e.notifier.NotifyVoteResult(gameID, "village", nil, true, tiedPlayers, nil, false)
		e.TransitionToNight(gameID)
	}
}

func (e *GameEngineV2) handleMayorTiebreakerCallback(prompt *entities.Prompt, response json.RawMessage) error {
	gameID := prompt.GameID

	if response == nil {
		// Mayor timed out - no one dies
		e.notifier.NotifyVoteResult(gameID, "village", nil, true, nil, nil, false)
		e.TransitionToNight(gameID)
		return nil
	}

	resp, err := prompts.ParseSelectPlayerResponse(response)
	if err != nil {
		return err
	}

	if resp.PlayerID == nil {
		e.notifier.NotifyVoteResult(gameID, "village", nil, true, nil, nil, false)
		e.TransitionToNight(gameID)
		return nil
	}

	// Eliminate chosen player
	e.eliminatePlayer(gameID, *resp.PlayerID, "village_vote")

	e.notifier.NotifyVoteResult(gameID, "village", resp.PlayerID, false, nil, nil, true)

	// Check win condition
	players, _ := e.playerRepo.GetPlayersByGame(context.TODO(), gameID)
	winResult := e.CheckWinCondition(players)
	if winResult.GameEnded {
		game, _ := e.gameRepo.GetGame(context.TODO(), gameID)
		return e.EndGame(gameID, game, winResult)
	}

	e.TransitionToNight(gameID)
	return nil
}

func (e *GameEngineV2) eliminatePlayer(gameID entities.GameID, playerID entities.PlayerID, cause string) {
	player, err := e.playerRepo.GetPlayer(context.TODO(), playerID)
	if err != nil {
		return
	}

	player.Kill()
	e.playerRepo.SavePlayer(context.TODO(), player)

	var role entities.RoleType
	if player.Role != nil {
		role = player.Role.GetType()
	}

	e.notifier.NotifyPlayerDied(gameID, playerID, player.Username, role, cause)
}

func (e *GameEngineV2) advanceFromCurrentNightPhase(gameID entities.GameID) {
	_, exists := e.nightService.GetNightState(gameID)
	if !exists {
		e.TransitionToDay(gameID)
		return
	}

	newPhase := e.nightService.AdvancePhase(gameID)

	if newPhase == entities.NightPhaseEnd || e.nightService.IsNightComplete(gameID) {
		e.TransitionToDay(gameID)
		return
	}

	players, err := e.playerRepo.GetPlayersByGame(context.TODO(), gameID)
	if err != nil {
		logger.Get().Error().Err(err).Msg("Failed to get players")
		return
	}

	e.startNextNightPhase(gameID, players)
}

// --- Legacy interface implementation for compatibility ---

// HandleSeerAction implements the old interface (for transition period)
func (e *GameEngineV2) HandleSeerAction(gameID entities.GameID, seerID, targetID entities.PlayerID) error {
	return apperrors.ErrNotImplemented
}

// HandleWerewolfVote implements the old interface
func (e *GameEngineV2) HandleWerewolfVote(gameID entities.GameID, werewolfID entities.PlayerID, targetID *entities.PlayerID) error {
	return apperrors.ErrNotImplemented
}

// HandleWitchAction implements the old interface
func (e *GameEngineV2) HandleWitchAction(gameID entities.GameID, witchID entities.PlayerID, healTargetID, poisonTargetID *entities.PlayerID) error {
	return apperrors.ErrNotImplemented
}

// HandleVillageVote implements the old interface
func (e *GameEngineV2) HandleVillageVote(gameID entities.GameID, voterID entities.PlayerID, targetID *entities.PlayerID) error {
	return apperrors.ErrNotImplemented
}

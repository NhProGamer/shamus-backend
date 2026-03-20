package app

import (
	"context"
	"encoding/json"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/prompts"
	"shamus-backend/internal/domain/ports"
	"shamus-backend/pkg/logger"
	"sync"
	"time"

	"github.com/google/uuid"
)

// PromptCallback is called when a prompt is answered or expires
// If response is nil, the prompt timed out
type PromptCallback func(prompt *entities.Prompt, response json.RawMessage) error

// GroupVoteResult represents the result of a group vote
type GroupVoteResult struct {
	// Target is the winning target (nil if no clear winner or tie)
	Target *entities.PlayerID

	// IsTie indicates if there was a tie
	IsTie bool

	// TiedTargets contains the tied targets (if IsTie is true)
	TiedTargets []entities.PlayerID

	// VoteCounts maps targetID to vote count
	VoteCounts map[entities.PlayerID]int

	// AllVotes maps voterID to their vote
	AllVotes map[entities.PlayerID]*entities.PlayerID
}

// GroupVoteState holds the state of an active group vote
type GroupVoteState struct {
	GroupID   entities.GroupID
	GameID    entities.GameID
	Context   string // e.g., "werewolf_vote", "village_vote"
	Voters    []entities.PlayerID
	Votes     map[entities.PlayerID]*entities.PlayerID // voterID -> targetID (nil = abstain)
	PromptIDs map[entities.PlayerID]entities.PromptID  // playerID -> their promptID
	ExpiresAt time.Time
	Resolved  bool

	// For mayor tiebreaker in village vote
	HasMayor    bool
	MayorID     *entities.PlayerID
	AwaitingMayor bool
	TiedTargets []entities.PlayerID
}

// PromptService manages prompts with timeouts and group voting
type PromptService struct {
	prompts       map[entities.PromptID]*entities.Prompt
	groupStates   map[entities.GroupID]*GroupVoteState
	playerPrompts map[entities.PlayerID][]entities.PromptID // Active prompts per player
	callbacks     map[string]PromptCallback                 // Callbacks by context
	timers        map[entities.PromptID]*time.Timer         // Individual prompt timers
	groupTimers   map[entities.GroupID]*time.Timer          // Group vote timers

	sender      ports.PlayerSender
	notifier    *NotificationService
	playerRepo  ports.PlayerRepository

	lock sync.RWMutex
}

// NewPromptService creates a new PromptService
func NewPromptService(
	sender ports.PlayerSender,
	notifier *NotificationService,
	playerRepo ports.PlayerRepository,
) *PromptService {
	return &PromptService{
		prompts:       make(map[entities.PromptID]*entities.Prompt),
		groupStates:   make(map[entities.GroupID]*GroupVoteState),
		playerPrompts: make(map[entities.PlayerID][]entities.PromptID),
		callbacks:     make(map[string]PromptCallback),
		timers:        make(map[entities.PromptID]*time.Timer),
		groupTimers:   make(map[entities.GroupID]*time.Timer),
		sender:        sender,
		notifier:      notifier,
		playerRepo:    playerRepo,
	}
}

// RegisterCallback registers a callback for a specific context
func (s *PromptService) RegisterCallback(context string, callback PromptCallback) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.callbacks[context] = callback
}

// --- Individual Prompts ---

// CreatePrompt creates and sends a prompt to a single player
func (s *PromptService) CreatePrompt(
	gameID entities.GameID,
	playerID entities.PlayerID,
	promptType entities.PromptType,
	context string,
	payload interface{},
	timeout time.Duration,
	canSkip bool,
) (*entities.Prompt, error) {
	promptID := entities.PromptID(uuid.New().String())

	prompt, err := entities.NewPrompt(
		promptID,
		promptType,
		context,
		gameID,
		playerID,
		payload,
		timeout,
		false, // allowChange = false for individual prompts
		canSkip,
		nil, // no group
	)
	if err != nil {
		return nil, err
	}

	s.lock.Lock()
	s.prompts[promptID] = prompt
	s.playerPrompts[playerID] = append(s.playerPrompts[playerID], promptID)
	s.lock.Unlock()

	// Send prompt to player
	if err := s.sendPromptToPlayer(prompt); err != nil {
		logger.Get().Warn().
			Str("promptID", string(promptID)).
			Str("playerID", string(playerID)).
			Err(err).
			Msg("Failed to send prompt to player")
	}

	// Start timeout timer
	s.startPromptTimer(prompt)

	logger.Get().Info().
		Str("promptID", string(promptID)).
		Str("type", string(promptType)).
		Str("context", context).
		Str("playerID", string(playerID)).
		Dur("timeout", timeout).
		Msg("Prompt created")

	return prompt, nil
}

// RespondToPrompt handles a player's response to a prompt
func (s *PromptService) RespondToPrompt(promptID entities.PromptID, playerID entities.PlayerID, response json.RawMessage) error {
	s.lock.Lock()
	prompt, exists := s.prompts[promptID]
	if !exists {
		s.lock.Unlock()
		return ErrPromptNotFound
	}

	// Verify player owns this prompt
	if prompt.PlayerID != playerID {
		s.lock.Unlock()
		return ErrPromptWrongPlayer
	}

	// Check if can respond
	if !prompt.CanRespond() {
		s.lock.Unlock()
		if prompt.Status == entities.PromptStatusExpired {
			return ErrPromptExpired
		}
		if prompt.Status == entities.PromptStatusAnswered && !prompt.AllowChange {
			return ErrPromptAlreadyAnswered
		}
		return ErrPromptInvalidState
	}

	// Handle group vote vs individual prompt
	if prompt.GroupID != nil {
		groupState, groupExists := s.groupStates[*prompt.GroupID]
		if !groupExists {
			s.lock.Unlock()
			return ErrGroupNotFound
		}

		// Update vote in group state
		var voteResp prompts.VoteResponse
		if err := json.Unmarshal(response, &voteResp); err != nil {
			s.lock.Unlock()
			return err
		}

		groupState.Votes[playerID] = voteResp.TargetID
		prompt.MarkAnswered(response)

		// Calculate updated vote state for notification
		voteCounts := s.calculateVoteCounts(groupState)
		votersRemaining := s.getVotersRemaining(groupState)

		// Get other voters to notify
		var othersToNotify []entities.PlayerID
		for _, voterID := range groupState.Voters {
			if voterID != playerID {
				othersToNotify = append(othersToNotify, voterID)
			}
		}
		s.lock.Unlock()

		// Notify other group members of vote update
		if len(othersToNotify) > 0 {
			s.notifier.NotifyVoteUpdate(othersToNotify, groupState.Votes, voteCounts, votersRemaining)
		}

		logger.Get().Info().
			Str("promptID", string(promptID)).
			Str("playerID", string(playerID)).
			Str("groupID", string(*prompt.GroupID)).
			Msg("Group vote recorded")

		return nil
	}

	// Individual prompt - cancel timer and invoke callback
	s.cancelPromptTimer(promptID)
	prompt.MarkAnswered(response)
	callback := s.callbacks[prompt.Context]
	s.lock.Unlock()

	logger.Get().Info().
		Str("promptID", string(promptID)).
		Str("context", prompt.Context).
		Str("playerID", string(playerID)).
		Msg("Prompt answered")

	// Invoke callback outside lock
	if callback != nil {
		return callback(prompt, response)
	}

	return nil
}

// CancelPrompt cancels a pending prompt
func (s *PromptService) CancelPrompt(promptID entities.PromptID) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	prompt, exists := s.prompts[promptID]
	if !exists {
		return nil // Already gone
	}

	if prompt.Status != entities.PromptStatusPending {
		return nil // Already handled
	}

	s.cancelPromptTimer(promptID)
	prompt.MarkCancelled()

	logger.Get().Info().
		Str("promptID", string(promptID)).
		Msg("Prompt cancelled")

	return nil
}

// --- Group Votes ---

// CreateGroupVote creates a group vote (werewolf or village vote)
func (s *PromptService) CreateGroupVote(
	gameID entities.GameID,
	voteContext string, // "werewolf_vote" or "village_vote"
	voters []*entities.Player,
	eligibleTargets []entities.PlayerID,
	timeout time.Duration,
	canAbstain bool,
	mayorID *entities.PlayerID, // For village vote tiebreaker (nil if no mayor)
) (*entities.GroupID, error) {
	groupID := entities.GroupID(uuid.New().String())
	expiresAt := time.Now().Add(timeout)

	// Build voters list and player info
	var voterIDs []entities.PlayerID
	playersInfo := make(map[entities.PlayerID]prompts.PlayerInfo)
	for _, p := range voters {
		voterIDs = append(voterIDs, p.ID)
		playersInfo[p.ID] = prompts.PlayerInfo{
			Username: p.Username,
			IsAlive:  p.IsAlive,
		}
	}

	// Add target info
	allPlayers, _ := s.playerRepo.GetPlayersByGame(context.TODO(), gameID)
	for _, p := range allPlayers {
		if _, exists := playersInfo[p.ID]; !exists {
			playersInfo[p.ID] = prompts.PlayerInfo{
				Username: p.Username,
				IsAlive:  p.IsAlive,
			}
		}
	}

	// Create group state
	groupState := &GroupVoteState{
		GroupID:     groupID,
		GameID:      gameID,
		Context:     voteContext,
		Voters:      voterIDs,
		Votes:       make(map[entities.PlayerID]*entities.PlayerID),
		PromptIDs:   make(map[entities.PlayerID]entities.PromptID),
		ExpiresAt:   expiresAt,
		Resolved:    false,
		HasMayor:    mayorID != nil,
		MayorID:     mayorID,
		AwaitingMayor: false,
	}

	s.lock.Lock()
	s.groupStates[groupID] = groupState

	// Create a prompt for each voter
	payload := prompts.VotePayload{
		Message:         s.getVoteMessage(voteContext),
		EligibleTargets: eligibleTargets,
		CanAbstain:      canAbstain,
		CurrentVotes:    make(map[entities.PlayerID]*entities.PlayerID),
		PlayersInfo:     playersInfo,
	}

	for _, voterID := range voterIDs {
		promptID := entities.PromptID(uuid.New().String())

		prompt, err := entities.NewPrompt(
			promptID,
			entities.PromptVote,
			voteContext,
			gameID,
			voterID,
			payload,
			timeout,
			true, // allowChange = true for group votes
			canAbstain,
			&groupID,
		)
		if err != nil {
			s.lock.Unlock()
			return nil, err
		}

		s.prompts[promptID] = prompt
		s.playerPrompts[voterID] = append(s.playerPrompts[voterID], promptID)
		groupState.PromptIDs[voterID] = promptID
	}
	s.lock.Unlock()

	// Send prompts to all voters
	for voterID, promptID := range groupState.PromptIDs {
		prompt := s.prompts[promptID]
		if err := s.sendPromptToPlayer(prompt); err != nil {
			logger.Get().Warn().
				Str("promptID", string(promptID)).
				Str("playerID", string(voterID)).
				Err(err).
				Msg("Failed to send group vote prompt")
		}
	}

	// Start group timer
	s.startGroupTimer(groupID, timeout)

	logger.Get().Info().
		Str("groupID", string(groupID)).
		Str("context", voteContext).
		Int("voterCount", len(voterIDs)).
		Dur("timeout", timeout).
		Msg("Group vote created")

	return &groupID, nil
}

// ResolveGroupVote manually resolves a group vote (called on timeout or when all voted)
func (s *PromptService) ResolveGroupVote(groupID entities.GroupID) (*GroupVoteResult, error) {
	s.lock.Lock()
	groupState, exists := s.groupStates[groupID]
	if !exists {
		s.lock.Unlock()
		return nil, ErrGroupNotFound
	}

	if groupState.Resolved {
		s.lock.Unlock()
		return nil, ErrGroupAlreadyResolved
	}

	// Cancel group timer
	s.cancelGroupTimer(groupID)

	// Mark all prompts as expired/answered
	for _, promptID := range groupState.PromptIDs {
		if prompt, ok := s.prompts[promptID]; ok {
			if prompt.Status == entities.PromptStatusPending {
				prompt.MarkExpired()
			}
		}
	}

	// Calculate result
	result := s.calculateGroupVoteResult(groupState)
	groupState.Resolved = true

	// Get callback
	callback := s.callbacks[groupState.Context]
	s.lock.Unlock()

	logger.Get().Info().
		Str("groupID", string(groupID)).
		Str("context", groupState.Context).
		Bool("isTie", result.IsTie).
		Msg("Group vote resolved")

	// Invoke callback if exists
	if callback != nil {
		// Create a synthetic prompt for the callback
		syntheticPrompt := &entities.Prompt{
			ID:      entities.PromptID(groupID),
			Context: groupState.Context,
			GameID:  groupState.GameID,
			GroupID: &groupID,
		}

		// Serialize result as response
		responseData, _ := json.Marshal(result)
		if err := callback(syntheticPrompt, responseData); err != nil {
			logger.Get().Error().Err(err).Msg("Group vote callback failed")
			return result, err
		}
	}

	return result, nil
}

// RequestMayorTiebreaker handles village vote tie when mayor exists
func (s *PromptService) RequestMayorTiebreaker(groupID entities.GroupID, tiedTargets []entities.PlayerID, voteCounts map[entities.PlayerID]int) error {
	s.lock.Lock()
	groupState, exists := s.groupStates[groupID]
	if !exists || groupState.MayorID == nil {
		s.lock.Unlock()
		return ErrNoMayor
	}

	groupState.AwaitingMayor = true
	groupState.TiedTargets = tiedTargets
	mayorID := *groupState.MayorID
	s.lock.Unlock()

	// Notify mayor
	return s.notifier.NotifyMayorTiebreaker(mayorID, tiedTargets, voteCounts)
}

// SubmitMayorDecision handles mayor's tiebreaker decision
func (s *PromptService) SubmitMayorDecision(groupID entities.GroupID, mayorID entities.PlayerID, chosenTarget entities.PlayerID) (*GroupVoteResult, error) {
	s.lock.Lock()
	groupState, exists := s.groupStates[groupID]
	if !exists {
		s.lock.Unlock()
		return nil, ErrGroupNotFound
	}

	if !groupState.AwaitingMayor {
		s.lock.Unlock()
		return nil, ErrNotAwaitingMayor
	}

	if groupState.MayorID == nil || *groupState.MayorID != mayorID {
		s.lock.Unlock()
		return nil, ErrNotMayor
	}

	// Verify chosen target is in tied targets
	validChoice := false
	for _, t := range groupState.TiedTargets {
		if t == chosenTarget {
			validChoice = true
			break
		}
	}
	if !validChoice {
		s.lock.Unlock()
		return nil, ErrInvalidMayorChoice
	}

	// Create result with mayor decision
	result := &GroupVoteResult{
		Target:     &chosenTarget,
		IsTie:      false, // Mayor resolved the tie
		VoteCounts: s.calculateVoteCounts(groupState),
		AllVotes:   groupState.Votes,
	}

	groupState.Resolved = true
	groupState.AwaitingMayor = false

	callback := s.callbacks[groupState.Context]
	s.lock.Unlock()

	logger.Get().Info().
		Str("groupID", string(groupID)).
		Str("mayorID", string(mayorID)).
		Str("chosenTarget", string(chosenTarget)).
		Msg("Mayor tiebreaker resolved")

	// Invoke callback
	if callback != nil {
		syntheticPrompt := &entities.Prompt{
			ID:      entities.PromptID(groupID),
			Context: groupState.Context,
			GameID:  groupState.GameID,
			GroupID: &groupID,
		}
		responseData, _ := json.Marshal(result)
		callback(syntheticPrompt, responseData)
	}

	return result, nil
}

// HasEveryoneVoted checks if all voters in a group have voted
func (s *PromptService) HasEveryoneVoted(groupID entities.GroupID) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	groupState, exists := s.groupStates[groupID]
	if !exists {
		return false
	}

	return len(groupState.Votes) == len(groupState.Voters)
}

// GetGroupState returns the current state of a group vote (for inspection)
func (s *PromptService) GetGroupState(groupID entities.GroupID) (*GroupVoteState, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	state, exists := s.groupStates[groupID]
	return state, exists
}

// --- Internal helpers ---

func (s *PromptService) sendPromptToPlayer(prompt *entities.Prompt) error {
	msg := prompt.ToMessage()
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return s.sender.SendToPlayer(prompt.PlayerID, data)
}

func (s *PromptService) startPromptTimer(prompt *entities.Prompt) {
	s.lock.Lock()
	defer s.lock.Unlock()

	timer := time.AfterFunc(prompt.Timeout, func() {
		s.handlePromptTimeout(prompt.ID)
	})
	s.timers[prompt.ID] = timer
}

func (s *PromptService) cancelPromptTimer(promptID entities.PromptID) {
	// Caller must hold lock
	if timer, exists := s.timers[promptID]; exists {
		timer.Stop()
		delete(s.timers, promptID)
	}
}

func (s *PromptService) handlePromptTimeout(promptID entities.PromptID) {
	s.lock.Lock()
	prompt, exists := s.prompts[promptID]
	if !exists || prompt.Status != entities.PromptStatusPending {
		s.lock.Unlock()
		return
	}

	delete(s.timers, promptID)
	prompt.MarkExpired()

	// For group prompts, timeout is handled by group timer
	if prompt.GroupID != nil {
		s.lock.Unlock()
		return
	}

	callback := s.callbacks[prompt.Context]
	s.lock.Unlock()

	logger.Get().Info().
		Str("promptID", string(promptID)).
		Str("context", prompt.Context).
		Msg("Prompt expired")

	// Invoke callback with nil response (timeout)
	if callback != nil {
		callback(prompt, nil)
	}
}

func (s *PromptService) startGroupTimer(groupID entities.GroupID, timeout time.Duration) {
	s.lock.Lock()
	defer s.lock.Unlock()

	timer := time.AfterFunc(timeout, func() {
		s.handleGroupTimeout(groupID)
	})
	s.groupTimers[groupID] = timer
}

func (s *PromptService) cancelGroupTimer(groupID entities.GroupID) {
	// Caller must hold lock
	if timer, exists := s.groupTimers[groupID]; exists {
		timer.Stop()
		delete(s.groupTimers, groupID)
	}
}

func (s *PromptService) handleGroupTimeout(groupID entities.GroupID) {
	logger.Get().Info().
		Str("groupID", string(groupID)).
		Msg("Group vote timeout")

	// Resolve the vote
	s.ResolveGroupVote(groupID)
}

func (s *PromptService) calculateVoteCounts(groupState *GroupVoteState) map[entities.PlayerID]int {
	counts := make(map[entities.PlayerID]int)
	for _, targetID := range groupState.Votes {
		if targetID != nil {
			counts[*targetID]++
		}
	}
	return counts
}

func (s *PromptService) getVotersRemaining(groupState *GroupVoteState) []entities.PlayerID {
	var remaining []entities.PlayerID
	for _, voterID := range groupState.Voters {
		if _, voted := groupState.Votes[voterID]; !voted {
			remaining = append(remaining, voterID)
		}
	}
	return remaining
}

func (s *PromptService) calculateGroupVoteResult(groupState *GroupVoteState) *GroupVoteResult {
	voteCounts := s.calculateVoteCounts(groupState)

	result := &GroupVoteResult{
		VoteCounts: voteCounts,
		AllVotes:   groupState.Votes,
	}

	if len(voteCounts) == 0 {
		// No one voted for anyone (all abstained)
		result.IsTie = false
		result.Target = nil
		return result
	}

	// Find max votes
	maxVotes := 0
	for _, count := range voteCounts {
		if count > maxVotes {
			maxVotes = count
		}
	}

	// Find all targets with max votes
	var topTargets []entities.PlayerID
	for targetID, count := range voteCounts {
		if count == maxVotes {
			topTargets = append(topTargets, targetID)
		}
	}

	if len(topTargets) == 1 {
		// Clear winner
		result.Target = &topTargets[0]
		result.IsTie = false
	} else {
		// Tie
		result.IsTie = true
		result.TiedTargets = topTargets
		result.Target = nil
	}

	return result
}

func (s *PromptService) getVoteMessage(context string) string {
	switch context {
	case "werewolf_vote":
		return "Choisissez votre victime"
	case "village_vote":
		return "Votez pour éliminer un joueur"
	default:
		return "Votez"
	}
}

// --- Cleanup ---

// CleanupGame removes all prompts and group states for a game
func (s *PromptService) CleanupGame(gameID entities.GameID) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Cancel and remove prompts
	for promptID, prompt := range s.prompts {
		if prompt.GameID == gameID {
			s.cancelPromptTimer(promptID)
			delete(s.prompts, promptID)
		}
	}

	// Cancel and remove group states
	for groupID, groupState := range s.groupStates {
		if groupState.GameID == gameID {
			s.cancelGroupTimer(groupID)
			delete(s.groupStates, groupID)
		}
	}

	logger.Get().Info().
		Str("gameID", string(gameID)).
		Msg("Cleaned up prompts for game")
}

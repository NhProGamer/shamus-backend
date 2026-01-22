package app

import "errors"

// Game action errors
var (
	// Phase/turn errors
	ErrNotYourTurn   = errors.New("not your turn")
	ErrWrongPhase    = errors.New("wrong game phase")
	ErrAlreadyActed  = errors.New("you already acted this phase")
	ErrGameNotActive = errors.New("game is not active")

	// Player state errors
	ErrPlayerDead = errors.New("player is dead")
	ErrWrongRole  = errors.New("you don't have this role")

	// Target errors
	ErrInvalidTarget    = errors.New("invalid target")
	ErrTargetDead       = errors.New("target is dead")
	ErrCannotTargetSelf = errors.New("cannot target yourself")

	// Ability errors
	ErrAbilityUsed       = errors.New("ability already used")
	ErrCanOnlyHealVictim = errors.New("can only heal the werewolf victim")

	// Vote errors
	ErrVoteNotFound      = errors.New("vote not found")
	ErrVoteAlreadyExists = errors.New("vote already exists for this game")
	ErrInvalidVoter      = errors.New("player is not eligible to vote")
	ErrVoteNotActive     = errors.New("vote is not active")
)

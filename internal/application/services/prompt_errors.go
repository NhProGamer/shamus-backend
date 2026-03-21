package services

import "errors"

// Prompt-related errors
var (
	// ErrPromptNotFound indicates the prompt doesn't exist
	ErrPromptNotFound = errors.New("prompt not found")

	// ErrPromptWrongPlayer indicates the player doesn't own this prompt
	ErrPromptWrongPlayer = errors.New("prompt belongs to another player")

	// ErrPromptExpired indicates the prompt has timed out
	ErrPromptExpired = errors.New("prompt has expired")

	// ErrPromptAlreadyAnswered indicates the prompt was already answered (and doesn't allow changes)
	ErrPromptAlreadyAnswered = errors.New("prompt already answered")

	// ErrPromptInvalidState indicates the prompt is in an invalid state for this operation
	ErrPromptInvalidState = errors.New("prompt is in an invalid state")

	// ErrGroupNotFound indicates the group vote doesn't exist
	ErrGroupNotFound = errors.New("group vote not found")

	// ErrGroupAlreadyResolved indicates the group vote was already resolved
	ErrGroupAlreadyResolved = errors.New("group vote already resolved")

	// ErrNoMayor indicates there's no mayor to break the tie
	ErrNoMayor = errors.New("no mayor in this game")

	// ErrNotAwaitingMayor indicates we're not waiting for mayor's decision
	ErrNotAwaitingMayor = errors.New("not awaiting mayor decision")

	// ErrNotMayor indicates the player is not the mayor
	ErrNotMayor = errors.New("player is not the mayor")

	// ErrInvalidMayorChoice indicates the mayor's choice is not among tied targets
	ErrInvalidMayorChoice = errors.New("invalid mayor choice: target not in tied players")
)

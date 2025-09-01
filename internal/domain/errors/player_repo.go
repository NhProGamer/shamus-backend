package domain_errors

import "errors"

var (
	ErrPlayerNotFound      = errors.New("PLAYER_NOT_FOUND")
	ErrPlayerAlreadyExists = errors.New("PLAYER_ALREADY_EXISTS")
	ErrPlayerHasNoGame     = errors.New("PLAYER_HAS_NO_GAME")
)

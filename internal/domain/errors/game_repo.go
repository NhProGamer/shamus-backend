package domain_errors

import "errors"

var (
	ErrGameNotFound      = errors.New("GAME_NOT_FOUND")
	ErrGameAlreadyExists = errors.New("GAME_ALREADY_EXISTS")
)

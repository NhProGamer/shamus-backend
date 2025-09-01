package ws

import "errors"

var (
	ErrUserIDMissing   = errors.New("USERID_MISSING")
	ErrGameIDMissing   = errors.New("GAMEID_MISSING")
	ErrGameNotFound    = errors.New("GAME_NOT_FOUND")
	ErrGameFull        = errors.New("GAME_FULL")
	ErrGameActive      = errors.New("GAME_ACTIVE")
	ErrPlayerNotInGame = errors.New("PLAYER_NOT_IN_GAME")
	ErrInternalServer  = errors.New("INTERNAL_SERVER_ERROR")
)

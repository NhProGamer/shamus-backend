package entities

import (
	"time"
)

type PlayerID string

type PlayerStatus int

const (
	PlayerStatusWaiting PlayerStatus = iota
	PlayerStatusAlive
	PlayerStatusDead
	PlayerStatusDisconnected
)

type Player struct {
	ID           PlayerID     `json:"id"`
	Username     string       `json:"username"`
	Status       PlayerStatus `json:"status"`
	Role         Role         `json:"role,omitempty"`
	LastActivity time.Time    `json:"last_activity"`
	IsReady      bool         `json:"is_ready"`
	VotedFor     *PlayerID    `json:"voted_for,omitempty"`
}

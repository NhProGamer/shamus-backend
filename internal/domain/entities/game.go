package entities

import (
	"time"
)

type GameID string
type GamePhase string
type GameStatus string

const (
	GameStatusWaiting GameStatus = "waiting"
	GameStatusActive  GameStatus = "active"
	GameStatusEnded   GameStatus = "ended"
)

const (
	PhaseDay   GamePhase = "day"
	PhaseNight GamePhase = "night"
	PhaseVote  GamePhase = "vote"
)

type Game struct {
	ID        GameID                 `json:"id"`
	Status    GameStatus             `json:"status"`
	Phase     GamePhase              `json:"phase"`
	Day       int                    `json:"day"`
	Players   map[PlayerID]*Player   `json:"players"`
	Host      PlayerID               `json:"host"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Settings  GameSettings           `json:"settings"`
	Events    []Event                `json:"events"`
	Votes     map[PlayerID]*PlayerID `json:"votes"` // PlayerID -> VotedFor PlayerID
}

type GameSettings struct {
	MinPlayers int              `json:"min_players"`
	MaxPlayers int              `json:"max_players"`
	Roles      map[RoleType]int `json:"roles"`
}

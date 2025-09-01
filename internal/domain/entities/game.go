package entities

type GameID string
type GamePhase string
type GameStatus string

const (
	GameStatusWaiting GameStatus = "waiting"
	GameStatusActive  GameStatus = "active"
	GameStatusEnded   GameStatus = "ended"
)

const (
	PhaseStart GamePhase = "start"
	PhaseDay   GamePhase = "day"
	PhaseNight GamePhase = "night"
	PhaseVote  GamePhase = "vote"
)

type Game struct {
	ID       GameID       `json:"id"`
	Status   GameStatus   `json:"status"`
	Phase    GamePhase    `json:"phase"`
	Day      int          `json:"day"`
	Players  []PlayerID   `json:"players"`
	Host     PlayerID     `json:"host"`
	Settings GameSettings `json:"settings"`
}

type GameSettings struct {
	MinPlayers int              `json:"min_players"`
	MaxPlayers int              `json:"max_players"`
	Roles      map[RoleType]int `json:"roles"`
}

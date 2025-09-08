package entities

import (
	"errors"
	"sync"
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
	PhaseStart GamePhase = "start"
	PhaseDay   GamePhase = "day"
	PhaseNight GamePhase = "night"
	PhaseVote  GamePhase = "vote"
)

type Game struct {
	mu       sync.RWMutex
	id       GameID
	status   GameStatus
	phase    GamePhase
	day      int
	players  []PlayerID
	host     PlayerID
	settings GameSettings
}

type GameSettings struct {
	Roles map[RoleType]int `json:"roles"`
}

func (gs *GameSettings) TotalRoles() int {
	total := 0
	for _, count := range gs.Roles {
		total += count
	}
	return total
}

func NewGame(id GameID, host PlayerID) *Game {
	return &Game{
		id:      id,
		status:  GameStatusWaiting,
		phase:   PhaseStart,
		day:     0,
		players: []PlayerID{host},
		host:    host,
		settings: NewGameSettings(map[RoleType]int{
			RoleVillager: 3,
			RoleWerewolf: 1,
		}),
	}
}

func NewGameSettings(roles map[RoleType]int) GameSettings {
	return GameSettings{Roles: roles}
}

// --- Getters thread-safe ---
func (g *Game) ID() GameID {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.id
}

func (g *Game) Status() GameStatus {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.status
}

func (g *Game) Phase() GamePhase {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.phase
}

func (g *Game) Day() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.day
}

func (g *Game) Players() []PlayerID {
	g.mu.RLock()
	defer g.mu.RUnlock()
	playersCopy := append([]PlayerID(nil), g.players...)
	return playersCopy
}

func (g *Game) ForEachPlayer(f func(PlayerID)) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, p := range g.players {
		f(p)
	}
}

func (g *Game) Host() PlayerID {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.host
}

func (g *Game) Settings() GameSettings {
	g.mu.RLock()
	defer g.mu.RUnlock()
	rolesCopy := make(map[RoleType]int, len(g.settings.Roles))
	for k, v := range g.settings.Roles {
		rolesCopy[k] = v
	}
	return GameSettings{Roles: rolesCopy}
}

// --- Setters thread-safe ---
func (g *Game) SetSettings(gameSettings *GameSettings) error {
	if gameSettings == nil {
		return errors.New("les paramètres du jeu ne peuvent pas être nuls")
	}
	if gameSettings.TotalRoles() < 4 {
		return errors.New("le nombre total de rôles doit être au moins de 4")
	}
	for role := range gameSettings.Roles {
		if !IsValidRole(string(role)) {
			return errors.New("rôle invalide dans les paramètres du jeu: " + string(role))
		}
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.settings = *gameSettings
	return nil
}

func (g *Game) SetStatus(status GameStatus) error {
	switch status {
	case GameStatusWaiting, GameStatusActive, GameStatusEnded:
		g.mu.Lock()
		defer g.mu.Unlock()
		g.status = status
		return nil
	default:
		return errors.New("statut de jeu invalide")
	}
}

func (g *Game) SetPhase(phase GamePhase) error {
	switch phase {
	case PhaseStart, PhaseDay, PhaseNight, PhaseVote:
		g.mu.Lock()
		defer g.mu.Unlock()
		g.phase = phase
		return nil
	default:
		return errors.New("phase de jeu invalide")
	}
}

func (g *Game) NextDay() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.day++
}

func (g *Game) AddPlayer(playerID PlayerID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.isFull() {
		return errors.New("nombre maximum de joueurs atteint")
	}
	for _, existingPlayer := range g.players {
		if existingPlayer == playerID {
			return errors.New("le joueur est déjà dans la partie")
		}
	}
	g.players = append(g.players, playerID)
	return nil
}

func (g *Game) RemovePlayer(playerID PlayerID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	for i, player := range g.players {
		if player == playerID {
			g.players = append(g.players[:i], g.players[i+1:]...)
			return nil
		}
	}
	return errors.New("joueur non trouvé")
}

func (g *Game) ChangeHost(newHost PlayerID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, player := range g.players {
		if player == newHost {
			g.host = newHost
			return nil
		}
	}
	return errors.New("le nouveau host doit être un joueur de la partie")
}

// --- Validation ---
func (g *Game) IsFull() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.isFull()
}

func (g *Game) isFull() bool {
	return len(g.players) == g.settings.TotalRoles()
}

func (g *Game) IsActive() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.status == GameStatusActive
}

func (g *Game) IsEnded() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.status == GameStatusEnded
}

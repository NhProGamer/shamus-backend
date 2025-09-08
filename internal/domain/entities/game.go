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

// Game avec champs privés (encapsulation) et protection par mutex RW
type Game struct {
	mu       sync.RWMutex // Mutex RW pour protéger l'accès concurrent
	id       GameID
	status   GameStatus
	phase    GamePhase
	day      int
	players  []PlayerID
	host     PlayerID
	settings GameSettings
}

type GameSettings struct {
	roles map[RoleType]int
}

// Constructeur pour Game
func NewGame(id GameID, host PlayerID, settings GameSettings) *Game {
	return &Game{
		mu:       sync.RWMutex{},
		id:       id,
		status:   GameStatusWaiting,
		phase:    PhaseStart,
		day:      0,
		players:  []PlayerID{host},
		host:     host,
		settings: settings,
	}
}

// Constructeur pour GameSettings
func NewGameSettings(roles map[RoleType]int) GameSettings {
	return GameSettings{
		roles: roles,
	}
}

// Getters pour Game (utilisent RLock pour les lectures)
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
	// Retourner une copie pour éviter la modification externe
	playersCopy := make([]PlayerID, len(g.players))
	copy(playersCopy, g.players)
	return playersCopy
}

func (g *Game) Host() PlayerID {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.host
}

func (g *Game) Settings() GameSettings {
	g.mu.RLock()
	defer g.mu.RUnlock()
	rolesCopy := make(map[RoleType]int)
	for k, v := range g.settings.roles {
		rolesCopy[k] = v
	}
	return GameSettings{
		roles: rolesCopy,
	}
}

// Setters avec validation pour Game (utilisent Lock pour les écritures)
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

	// Vérifier si le joueur existe déjà
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

	// Vérifier que le nouveau host est dans la partie
	for _, player := range g.players {
		if player == newHost {
			g.host = newHost
			return nil
		}
	}
	return errors.New("le nouveau host doit être un joueur de la partie")
}

func (g *Game) TotalRoles() (total int) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, count := range g.settings.roles {
		total += count
	}
	return total
}

// Méthodes de validation (utilisent RLock pour les lectures)
func (g *Game) IsFull() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.isFull()
}

// Méthode privée pour éviter le double lock
func (g *Game) isFull() bool {
	totalRoles := 0
	for _, count := range g.settings.roles {
		totalRoles += count
	}
	return len(g.players) == totalRoles
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

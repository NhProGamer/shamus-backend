package entities

import "errors"

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

// Game avec champs privés (encapsulation)
type Game struct {
	id       GameID
	status   GameStatus
	phase    GamePhase
	day      int
	players  []PlayerID
	host     PlayerID
	settings gameSettings
}

// GameSettings avec champs privés
type gameSettings struct {
	minPlayers int              `json:"min_players"`
	maxPlayers int              `json:"max_players"`
	roles      map[RoleType]int `json:"roles"`
}

// Constructeur pour Game
func NewGame(id GameID, host PlayerID, settings GameSettings) *Game {
	return &Game{
		id:      id,
		status:  GameStatusWaiting,
		phase:   PhaseStart,
		day:     1,
		players: []PlayerID{host},
		host:    host,
		settings: gameSettings{
			minPlayers: settings.MinPlayers,
			maxPlayers: settings.MaxPlayers,
			roles:      settings.Roles,
		},
	}
}

// Constructeur pour GameSettings
func NewGameSettings(minPlayers, maxPlayers int, roles map[RoleType]int) GameSettings {
	return GameSettings{
		MinPlayers: minPlayers,
		MaxPlayers: maxPlayers,
		Roles:      roles,
	}
}

// Getters pour Game
func (g *Game) ID() GameID {
	return g.id
}

func (g *Game) Status() GameStatus {
	return g.status
}

func (g *Game) Phase() GamePhase {
	return g.phase
}

func (g *Game) Day() int {
	return g.day
}

func (g *Game) Players() []PlayerID {
	// Retourner une copie pour éviter la modification externe
	playersCopy := make([]PlayerID, len(g.players))
	copy(playersCopy, g.players)
	return playersCopy
}

func (g *Game) Host() PlayerID {
	return g.host
}

func (g *Game) Settings() GameSettings {
	return GameSettings{
		MinPlayers: g.settings.minPlayers,
		MaxPlayers: g.settings.maxPlayers,
		Roles:      g.settings.roles,
	}
}

// Setters avec validation pour Game
func (g *Game) SetStatus(status GameStatus) error {
	switch status {
	case GameStatusWaiting, GameStatusActive, GameStatusEnded:
		g.status = status
		return nil
	default:
		return errors.New("statut de jeu invalide")
	}
}

func (g *Game) SetPhase(phase GamePhase) error {
	switch phase {
	case PhaseStart, PhaseDay, PhaseNight, PhaseVote:
		g.phase = phase
		return nil
	default:
		return errors.New("phase de jeu invalide")
	}
}

func (g *Game) NextDay() {
	g.day++
}

func (g *Game) AddPlayer(playerID PlayerID) error {
	if len(g.players) >= g.settings.maxPlayers {
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
	for i, player := range g.players {
		if player == playerID {
			g.players = append(g.players[:i], g.players[i+1:]...)
			return nil
		}
	}
	return errors.New("joueur non trouvé")
}

func (g *Game) ChangeHost(newHost PlayerID) error {
	// Vérifier que le nouveau host est dans la partie
	for _, player := range g.players {
		if player == newHost {
			g.host = newHost
			return nil
		}
	}
	return errors.New("le nouveau host doit être un joueur de la partie")
}

// Méthodes de validation
func (g *Game) CanStart() bool {
	return len(g.players) >= g.settings.minPlayers &&
		len(g.players) <= g.settings.maxPlayers &&
		g.status == GameStatusWaiting
}

func (g *Game) IsActive() bool {
	return g.status == GameStatusActive
}

func (g *Game) IsEnded() bool {
	return g.status == GameStatusEnded
}

type GameSettings struct {
	MinPlayers int              `json:"min_players"`
	MaxPlayers int              `json:"max_players"`
	Roles      map[RoleType]int `json:"roles"`
}

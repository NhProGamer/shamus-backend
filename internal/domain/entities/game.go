package entities

import (
	"fmt"
	"shamus-backend/internal/domain/constants"
)

type GameID string
type GamePhase string
type GameStatus string

const (
	MinPlayers = constants.MinPlayers
	MaxPlayers = constants.MaxPlayers
)

// Role limits (max count per role type)
var RoleLimits = map[RoleType]int{
	RoleSeer:  1,
	RoleWitch: 1,
	// Villager and Werewolf have no specific limit (capped by MaxPlayers)
}

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
	HostID   PlayerID     `json:"hostId"`
	Settings GameSettings `json:"settings"`
}

type GameSettings struct {
	Roles map[RoleType]int `json:"roles"`
}

// TotalRoles returns the total number of roles configured
func (s *GameSettings) TotalRoles() int {
	total := 0
	for _, count := range s.Roles {
		total += count
	}
	return total
}

// ValidateSettings validates game settings for configuration phase
// Returns nil if valid, error otherwise
func (g *Game) ValidateSettings() error {
	totalRoles := 0
	for roleType, count := range g.Settings.Roles {
		if !IsValidRole(string(roleType)) {
			return fmt.Errorf("invalid role type: %s", roleType)
		}
		if count < 0 {
			return fmt.Errorf("role count cannot be negative: %s", roleType)
		}

		// Check role-specific limits
		if limit, hasLimit := RoleLimits[roleType]; hasLimit && count > limit {
			return fmt.Errorf("too many %s: max %d allowed", roleType, limit)
		}

		totalRoles += count
	}

	if totalRoles > MaxPlayers {
		return fmt.Errorf("too many roles configured: %d (max %d)", totalRoles, MaxPlayers)
	}

	return nil
}

// CanStart checks if the game can be started
// Returns nil if the game can start, error otherwise
func (g *Game) CanStart() error {
	playerCount := len(g.Players)

	if playerCount < MinPlayers {
		return fmt.Errorf("not enough players: %d/%d minimum", playerCount, MinPlayers)
	}
	if playerCount > MaxPlayers {
		return fmt.Errorf("too many players: %d/%d maximum", playerCount, MaxPlayers)
	}

	totalRoles := 0
	roles := make([]RoleType, 0)
	for roleType, count := range g.Settings.Roles {
		if !IsValidRole(string(roleType)) {
			return fmt.Errorf("invalid role type: %s", roleType)
		}

		// Check role-specific limits
		if limit, hasLimit := RoleLimits[roleType]; hasLimit && count > limit {
			return fmt.Errorf("too many %s: max %d allowed", roleType, limit)
		}

		totalRoles += count
		for i := 0; i < count; i++ {
			roles = append(roles, roleType)
		}
	}

	if totalRoles != playerCount {
		return fmt.Errorf("role count mismatch: %d roles for %d players", totalRoles, playerCount)
	}

	clans := GetClansFromRoles(roles)
	if !HasRequiredClans(clans) {
		return fmt.Errorf("invalid composition: need at least one villager and one werewolf/rogue")
	}

	return nil
}
